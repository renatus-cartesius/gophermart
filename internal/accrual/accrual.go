package accrual

import (
	"context"
	"errors"
)

const (
	TypeStatusRegistered = "REGISTERED"
	TypeStatusInvalid    = "INVALID"
	TypeStatusProcessing = "PROCESSING"
	TypeStatusProcessed  = "PROCESSED"
)

var (
	ErrOrderNotFound = errors.New("the order wasn`t registered in accrual")
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
	accrualUrl string
}
