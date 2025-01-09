package loyalty

import (
	"context"
	"time"

	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

func (l *Loyalty) Dispatch(ctx context.Context) error {
	dispatchTicker := time.NewTicker(time.Second * 10)
	defer dispatchTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info(
				"closing loyalty dispatcher",
			)
			return nil
		case <-dispatchTicker.C:
			logger.Log.Debug(
				"begin unhandled order processing",
			)
			logger.Log.Debug(
				"order processing ended",
				zap.Error(l.ProcessUnhandledOrders(ctx)),
			)
		}
	}
}

func (l *Loyalty) ProcessUnhandledOrders(ctx context.Context) error {
	orders, err := l.storage.GetUnhandledOrders(ctx)
	if err != nil {
		return err
	}

	for _, order := range orders {
		err := l.UpdateOrderStatus(ctx, order)
		if err != nil {
			return err
		}
	}

	return nil

}

func (l *Loyalty) UpdateOrderStatus(ctx context.Context, orderID string) error {
	orderInfo, err := l.accrual.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	err = l.storage.UpdateOrder(ctx, orderInfo)
	return err
}
