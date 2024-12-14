package loyalty

import (
	"github.com/renatus-cartesius/gophermart/internal/accrual"
	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

type Loyalty struct {
	accrual accrual.Accrualler
}

func (l *Loyalty) Orders(order int) error {
	res, err := l.accrual.GetOrder(order)
	if err != nil {
		return err
	}

	logger.Log.Debug(
		"getting order",
		zap.Int64("order", res.Order),
		zap.String("status", res.Status),
		zap.Float64("accrual", res.Accrual),
	)
	return nil
}
