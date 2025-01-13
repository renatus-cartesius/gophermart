package loyalty

import (
	"errors"
	"time"
)

var (
	ErrWithdrawNotEnoughPoints = errors.New("not enough points for withdraw")
)

type Withdraw struct {
	OrderID string    `json:"order"`
	UserID  string    `json:"-"`
	Sum     float64   `json:"sum"`
	Created time.Time `json:"processed_at"`
}
