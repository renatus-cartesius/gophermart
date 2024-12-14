package accrual

const (
	TypeStatusRegistered = "REGISTERED"
	TypeStatusInvalid    = "INVALID"
	TypeStatusProcessing = "PROCESSING"
	TypeStatusProcessed  = "PROCESSED"
)

type OrderInfo struct {
	Order   int64   `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Accrualler interface {
	GetOrder(int) (*OrderInfo, error)
}

type Accrual struct {
	accrualUrl string
}
