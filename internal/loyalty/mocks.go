package loyalty

import (
	"context"
	"time"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
)

type MockAccrualler struct {
	Orders map[string]*accrual.OrderInfo
}

func (ma MockAccrualler) GetOrder(ctx context.Context, orderID string) (*accrual.OrderInfo, error) {
	order, ok := ma.Orders[orderID]
	if !ok {
		return nil, accrual.ErrOrderNotFound
	}
	return order, nil
}

type MockLoyaltyStorager struct {
	Records     map[string]*Order
	Withdrawals map[string]*Withdraw
}

func (mls MockLoyaltyStorager) AddOrder(ctx context.Context, userID string, order *accrual.OrderInfo) error {
	orderRecord, ok := mls.Records[order.Order]
	if ok {
		if orderRecord.UserID != userID {
			return ErrOrderUploadedAnotherUser
		}
		return nil
	}

	mls.Records[order.Order] = &Order{
		UserID:   userID,
		ID:       order.Order,
		Status:   order.Status,
		Accrual:  order.Accrual,
		Uploaded: time.Now(),
	}
	return nil
}

func (mls MockLoyaltyStorager) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	orderRecord, ok := mls.Records[orderID]
	if !ok {
		return nil, ErrOrderNotFound
	}
	return orderRecord, nil
}

func (mls MockLoyaltyStorager) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	res := make([]*Order, 0)

	for _, record := range mls.Records {
		if record.UserID == userID {
			res = append(res, record)
		}
	}

	return res, nil
}

func (mls MockLoyaltyStorager) GetWithdrawals(ctx context.Context, userID string) ([]*Withdraw, error) {
	res := make([]*Withdraw, 0)

	for _, withdraw := range mls.Withdrawals {
		if withdraw.UserID == userID {
			res = append(res, withdraw)
		}
	}

	return res, nil
}

func (mls MockLoyaltyStorager) AddWithdraw(ctx context.Context, wr *Withdraw) error {
	mls.Withdrawals[wr.OrderID] = wr
	return nil
}

func (mls MockLoyaltyStorager) GetBalance(ctx context.Context, userID string) (*Balance, error) {
	balance := &Balance{}

	for _, v := range mls.Withdrawals {
		if v.UserID == userID {
			balance.Withdrawn += v.Sum
		}
	}

	for _, v := range mls.Records {
		if v.UserID == userID {
			balance.Current += v.Accrual
		}
	}

	balance.Current = balance.Current - balance.Withdrawn
	return balance, nil
}
