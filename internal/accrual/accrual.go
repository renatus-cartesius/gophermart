package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	Order   int64   `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Accrualler interface {
	GetOrder(context.Context, int64) (*OrderInfo, error)
}

type Accrual struct {
	accrualURL string
	httpClient *resty.Client
}

func NewAccrual(aurl string) *Accrual {
	httpClient := resty.New()
	httpClient.
		SetRetryCount(3).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		})

	return &Accrual{
		accrualURL: aurl,
		httpClient: httpClient,
	}
}

func (a *Accrual) GetOrder(ctx context.Context, orderID int64) (*OrderInfo, error) {
	payload := []byte(fmt.Sprintf("%v", orderID))
	req := a.httpClient.R().SetBody(payload)

	orderInfoRaw, err := req.Get(a.accrualURL)
	if err != nil {
		logger.Log.Debug(
			"error on making request to accrual",
			zap.Error(err),
		)
		return nil, err
	}

	orderInfo := &OrderInfo{}

	if err := json.NewDecoder(orderInfoRaw.RawBody()).Decode(&orderInfo); err != nil {
		logger.Log.Debug(
			"error on reading result from accrual",
			zap.Error(err),
		)
	}

	return orderInfo, nil
}
