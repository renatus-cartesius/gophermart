package loyalty

import "time"

type OrderRecord struct {
	UserID   string    `json:"-"`
	Number   int64     `json:"number"`
	Status   string    `json:"status"`
	Accrual  float64   `json:"accrual"`
	Uploaded time.Time `json:"uploaded"`
}
