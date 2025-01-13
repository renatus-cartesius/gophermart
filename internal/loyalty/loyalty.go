package loyalty

import (
	"context"
	"errors"
	"strconv"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
	"github.com/renatus-cartesius/gophermart/pkg/luhn"
)

const (
	TypeStatusNew        = "NEW"
	TypeStatusInvalid    = "INVALID"
	TypeStatusProcessing = "PROCESSING"
	TypeStatusProcessed  = "PROCESSED"
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
	AddOrder(ctx context.Context, userID string, orderID string) error
	GetOrders(ctx context.Context, userID string) ([]*Order, error)
	GetWithdrawals(ctx context.Context, userID string) ([]*Withdraw, error)
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	GetBalance(ctx context.Context, userID string) (*Balance, error)
	AddWithdraw(ctx context.Context, wr *Withdraw) error
	GetUnhandledOrders(ctx context.Context) ([]string, error)
	UpdateOrder(ctx context.Context, orderInfo *accrual.OrderInfo) error
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

func (l *Loyalty) UploadOrder(ctx context.Context, userID string, orderID string) error {
	// Validating order num with Luhn algorithm (409)
	number, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return err
	}

	if !luhn.Valid(number) {
		return ErrOrderInvalid
	}

	return l.storage.AddOrder(ctx, userID, orderID)

}

func (l *Loyalty) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	return l.storage.GetOrders(ctx, userID)
}

func (l *Loyalty) GetWithdrawals(ctx context.Context, userID string) ([]*Withdraw, error) {
	return l.storage.GetWithdrawals(ctx, userID)
}

func (l *Loyalty) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	return l.storage.GetOrder(ctx, orderID)
}

func (l *Loyalty) GetBalance(ctx context.Context, userID string) (*Balance, error) {
	return l.storage.GetBalance(ctx, userID)
}

// func (l *Loyalty) AddWithdraw(ctx context.Context, wr *WithdrawRequest) error {
// 	return l.storage.AddWithdraw(ctx, wr)
// }

func (l *Loyalty) Withdraw(ctx context.Context, wr *Withdraw) error {
	// Check if order processed
	// orderInfo, err := l.accrual.GetOrder(ctx, wr.OrderID)
	// if err != nil {
	// 	return err
	// }

	// if orderInfo.Status != accrual.TypeStatusProcessed {
	// 	return accrual.ErrOrderNotProcessed
	// }

	// Check if orderID is Luhn-valid

	number, err := strconv.ParseInt(wr.OrderID, 10, 64)
	if err != nil {
		return err
	}

	if !luhn.Valid(number) {
		return ErrOrderInvalid
	}

	return l.storage.AddWithdraw(ctx, wr)
}
