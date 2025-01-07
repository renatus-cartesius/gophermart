package loyalty

import (
	"time"
)

type Order struct {
	UserID   string    `json:"-"`
	ID       string    `json:"number"`
	Status   string    `json:"status"`
	Accrual  float64   `json:"accrual"`
	Uploaded time.Time `json:"uploaded"`
}
