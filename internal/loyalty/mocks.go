package loyalty

import (
	"context"
	"time"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
)

type MockAccrualler struct {
	Orders map[int64]*accrual.OrderInfo
}

func (ma MockAccrualler) GetOrder(ctx context.Context, orderID int64) (*accrual.OrderInfo, error) {
	order, ok := ma.Orders[orderID]
	if !ok {
		return nil, accrual.ErrOrderNotFound
	}
	return order, nil
}

type MockLoyaltyStorager struct {
	Records map[int64]*Order
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

func (mls MockLoyaltyStorager) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	res := make([]*Order, 0)

	for _, record := range mls.Records {
		if record.UserID == userID {
			res = append(res, record)
		}
	}

	return res, nil
}
