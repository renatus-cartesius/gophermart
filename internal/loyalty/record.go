package loyalty

import (
	"encoding/json"
	"time"
)

type Order struct {
	UserID   string    `json:"-"`
	ID       string    `json:"number"`
	Status   string    `json:"status"`
	Accrual  float64   `json:"accrual"`
	Uploaded time.Time `json:"uploaded"`
}

func (o *Order) MarshalJSON() ([]byte, error) {

	if o.Status != TypeStatusProcessed {
		return json.Marshal(
			struct {
				UserID   string    `json:"-"`
				ID       string    `json:"number"`
				Status   string    `json:"status"`
				Uploaded time.Time `json:"uploaded"`
			}{
				UserID:   o.UserID,
				ID:       o.ID,
				Status:   o.Status,
				Uploaded: o.Uploaded,
			})
	} else {
		return json.Marshal(
			struct {
				UserID   string    `json:"-"`
				ID       string    `json:"number"`
				Status   string    `json:"status"`
				Accrual  float64   `json:"accrual"`
				Uploaded time.Time `json:"uploaded"`
			}{
				UserID:   o.UserID,
				ID:       o.ID,
				Status:   o.Status,
				Accrual:  o.Accrual,
				Uploaded: o.Uploaded,
			})
	}
}
