package loyalty

import (
	"errors"
	"time"
)

var (
	ErrWithdrawNotEnoughPoints = errors.New("not enough points for withdraw")
)

type Withdraw struct {
	OrderID int64     `json:"order"`
	UserID  string    `json:"-"`
	Sum     float64   `json:"sum"`
	Created time.Time `json:"-"`
}
