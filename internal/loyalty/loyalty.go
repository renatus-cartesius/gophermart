package loyalty

import (
	"context"
	"database/sql"
	"errors"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
	"github.com/renatus-cartesius/gophermart/pkg/luhn"
)

var (
	ErrOrderUploadedAnotherUser = errors.New("the order has already been uploaded by another user")
	ErrOrderInvalid             = errors.New("order is invalid")
)

// type Loyaltier interface {
// 	ListOrders(user string) error
// }

type LoyaltyStorager interface {
	AddOrder(ctx context.Context, userID string, orderInfo *accrual.OrderInfo) error
	GetOrders(ctx context.Context, userID string) ([]*Order, error)
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
	// row := l.db.QueryRowContext(ctx, "SELECT ")

	// if row.Scan()

	// res, err := l.db.ExecContext(ctx, "INSERT")
	return nil
}

func (l *LoyaltyStoragePg) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	return nil, nil
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
