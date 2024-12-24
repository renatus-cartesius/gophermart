package loyalty

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"github.com/renatus-cartesius/gophermart/pkg/luhn"
	"go.uber.org/zap"
)

var (
	ErrOrderUploadedAnotherUser = errors.New("the order has already been uploaded by another user")
	ErrOrderInvalid             = errors.New("order is invalid")
	ErrOrderAlreadyUploaded     = errors.New("order is already uploaded")
	ErrOrderNotFound            = errors.New("order not found in storage")
)

// type Loyaltier interface {
// 	ListOrders(user string) error
// }

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type LoyaltyStorager interface {
	AddOrder(ctx context.Context, userID string, orderInfo *accrual.OrderInfo) error
	GetOrders(ctx context.Context, userID string) ([]*Order, error)
	GetOrder(ctx context.Context, orderID int64) (*Order, error)
	GetBalance(ctx context.Context, userID string) (*Balance, error)
	AddWithdraw(ctx context.Context, wr *WithdrawRequest) error
}

type LoyaltyStoragePg struct {
	db *sql.DB
}

func NewLoyaltyStoragePg(db *sql.DB) LoyaltyStorager {

	return &LoyaltyStoragePg{
		db: db,
	}
}

func (l *LoyaltyStoragePg) AddOrder(ctx context.Context, userID string, orderInfo *accrual.OrderInfo) error {

	// Check if order already in db
	var count int
	if err := l.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE id = $1", orderInfo.Order).Scan(&count); err != nil {
		logger.Log.Debug(
			"error on scanning row",
			zap.Error(err),
		)
		return err
	}

	if count > 0 {
		return ErrOrderAlreadyUploaded
	}

	_, err := l.db.ExecContext(ctx, "INSERT INTO orders (id, userID, status, accrual, uploaded) VALUES ($1, $2, $3, $4, $5)", orderInfo.Order, userID, orderInfo.Status, orderInfo.Accrual, time.Now())
	return err
}

func (l *LoyaltyStoragePg) GetOrders(ctx context.Context, userID string) ([]*Order, error) {

	orders := make([]*Order, 0)

	rows, err := l.db.QueryContext(ctx, "SELECT * FROM orders WHERE userID = $1 ORDER BY uploaded", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		order := &Order{}
		if err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.Uploaded); err != nil {
			logger.Log.Debug(
				"error on scanning row to Order",
				zap.Error(err),
			)
			continue
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (l *LoyaltyStoragePg) GetOrder(ctx context.Context, orderID int64) (*Order, error) {
	return nil, nil
}

func (l *LoyaltyStoragePg) GetBalance(ctx context.Context, userID string) (*Balance, error) {
	row := l.db.QueryRowContext(ctx, `
		select 
			((select SUM(accrual) from orders where userID = $1) -
			(select SUM(sum) from withdrawals where userID = $1)) as balance,
			(select SUM(sum) from withdrawals where userID = $1) as withdrawn;
		;
	`, userID)

	balance := &Balance{}
	err := row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		logger.Log.Debug(
			"error on scanning to row to balance",
		)
		return nil, err
	}

	return balance, row.Err()
}

func (l *LoyaltyStoragePg) AddWithdraw(ctx context.Context, wr *WithdrawRequest) error {
	_, err := l.db.ExecContext(ctx, "INSERT INTO withdrawals (orderID, userID, sum, created) VALUES ($1, $2, $3, $4)", wr.OrderID, wr.UserID, wr.Sum, wr.Created)
	return err
}

type Loyalty struct {
	accrual accrual.Accrualler
	storage LoyaltyStorager
}

func NewLoyalty(accrual accrual.Accrualler, storage LoyaltyStorager) *Loyalty {
	return &Loyalty{
		accrual: accrual,
		storage: storage,
	}
}

func (l *Loyalty) UploadOrder(ctx context.Context, userID string, orderID int64) error {
	// Validating order num with Luhn algorithm (409)
	if !luhn.Valid(orderID) {
		return ErrOrderInvalid
	}

	// Need to check what error is (204, 429, 500)
	orderInfo, err := l.accrual.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	// Checking if order is has already been uploaded to db but that user or another
	return l.storage.AddOrder(ctx, userID, orderInfo)
}

func (l *Loyalty) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	return l.storage.GetOrders(ctx, userID)
}

func (l *Loyalty) GetOrder(ctx context.Context, orderID int64) (*Order, error) {
	return l.storage.GetOrder(ctx, orderID)
}

func (l *Loyalty) GetBalance(ctx context.Context, userID string) (*Balance, error) {
	return l.storage.GetBalance(ctx, userID)
}

func (l *Loyalty) AddWithdraw(ctx context.Context, wr *WithdrawRequest) error {
	return l.storage.AddWithdraw(ctx, wr)
}

func (l *Loyalty) Withdraw(ctx context.Context, wr *WithdrawRequest) error {
	// Check if order processed
	orderInfo, err := l.accrual.GetOrder(ctx, wr.OrderID)
	if err != nil {
		return err
	}

	if orderInfo.Status != accrual.TypeStatusProcessed {
		return accrual.ErrOrderNotProcessed
	}

	// Check if user has enough points for withdraw
	balance, err := l.GetBalance(ctx, wr.UserID)
	if err != nil {
		return err
	}

	if wr.Sum > balance.Current {
		return ErrWithdrawNotEnoughPoints
	}

	return l.storage.AddWithdraw(ctx, wr)
}
