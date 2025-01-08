package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

const (
	TypeStatusRegistered = "REGISTERED"
	TypeStatusInvalid    = "INVALID"
	TypeStatusProcessing = "PROCESSING"
	TypeStatusProcessed  = "PROCESSED"
)

var (
	ErrOrderNotFound     = errors.New("the order wasn`t registered in accrual")
	ErrOrderNotProcessed = errors.New("the order isn`t processed")
)

type OrderInfo struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Accrualler interface {
	GetOrder(context.Context, string) (*OrderInfo, error)
}

type Accrual struct {
	accrualAddress string
	httpClient     *resty.Client
}

func NewAccrual(aAddress string) *Accrual {
	httpClient := resty.New()
	httpClient.
		SetRetryCount(3).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		})

	return &Accrual{
		accrualAddress: aAddress,
		httpClient:     httpClient,
	}
}

func (a *Accrual) GetOrder(ctx context.Context, orderID string) (*OrderInfo, error) {

	logger.Log.Debug(
		"checking order in accrual",
		zap.String("orderID", orderID),
	)

	req := a.httpClient.R()

	orderInfoRaw, err := req.Get(a.accrualAddress + "/api/orders/" + orderID)
	if err != nil {
		logger.Log.Debug(
			"error on making request to accrual",
			zap.Error(err),
		)
		return nil, err
	}

	orderInfo := &OrderInfo{}

	if orderInfoRaw.StatusCode() == 204 {
		logger.Log.Debug(
			"order wan not found in accrual",
			zap.String("orderID", orderID),
		)
		return nil, ErrOrderNotFound
	}

	if err := json.Unmarshal(orderInfoRaw.Body(), &orderInfo); err != nil {
		logger.Log.Debug(
			"error on reading result from accrual",
			zap.Error(err),
			zap.String("resp", string(orderInfoRaw.Body())),
			zap.String("code", orderInfoRaw.Status()),
		)
	}

	logger.Log.Debug(
		"order was found in accrual",
		zap.String("orderID", orderInfo.Order),
		zap.String("staus", orderInfo.Status),
		zap.Float64("accrual", orderInfo.Accrual),
	)
	return orderInfo, nil
}
