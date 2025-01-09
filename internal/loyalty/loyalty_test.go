package loyalty

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
)

type MockAccrualler struct {
	Orders map[string]*accrual.OrderInfo
}

func (ma MockAccrualler) GetOrder(ctx context.Context, orderID string) (*accrual.OrderInfo, error) {
	order, ok := ma.Orders[orderID]
	if !ok {
		return nil, accrual.ErrOrderNotFound
	}
	return order, nil
}

type MockLoyaltyStorager struct {
	Records     map[string]*Order
	Withdrawals map[string]*Withdraw
}

func (mls MockLoyaltyStorager) AddOrder(ctx context.Context, userID string, orderID string) error {
	orderRecord, ok := mls.Records[orderID]
	if ok {
		if orderRecord.UserID != userID {
			return ErrOrderUploadedAnotherUser
		}
		return nil
	}

	mls.Records[orderID] = &Order{
		UserID:   userID,
		ID:       orderID,
		Status:   TypeStatusNew,
		Accrual:  0,
		Uploaded: time.Now(),
	}
	return nil
}

func (mls MockLoyaltyStorager) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	orderRecord, ok := mls.Records[orderID]
	if !ok {
		return nil, ErrOrderNotFound
	}
	return orderRecord, nil
}

func (mls MockLoyaltyStorager) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	res := make([]*Order, 0)

	for _, record := range mls.Records {
		if record.UserID == userID {
			res = append(res, record)
		}
	}

	return res, nil
}

func (mls MockLoyaltyStorager) GetWithdrawals(ctx context.Context, userID string) ([]*Withdraw, error) {
	res := make([]*Withdraw, 0)

	for _, withdraw := range mls.Withdrawals {
		if withdraw.UserID == userID {
			res = append(res, withdraw)
		}
	}

	return res, nil
}

func (mls MockLoyaltyStorager) AddWithdraw(ctx context.Context, wr *Withdraw) error {
	mls.Withdrawals[wr.OrderID] = wr
	return nil
}

func (mls MockLoyaltyStorager) GetBalance(ctx context.Context, userID string) (*Balance, error) {
	balance := &Balance{}

	for _, v := range mls.Withdrawals {
		if v.UserID == userID {
			balance.Withdrawn += v.Sum
		}
	}

	for _, v := range mls.Records {
		if v.UserID == userID {
			balance.Current += v.Accrual
		}
	}

	balance.Current = balance.Current - balance.Withdrawn
	return balance, nil
}

func (mls MockLoyaltyStorager) GetUnhandledOrders(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (mls MockLoyaltyStorager) UpdateOrder(ctx context.Context, orderInfo *accrual.OrderInfo) error {
	return nil
}

func TestLoyalty_UploadOrder(t *testing.T) {

	ctx := context.Background()

	mockAccrualler := MockAccrualler{
		Orders: map[string]*accrual.OrderInfo{
			"79927398713": {
				Order:   "79927398713",
				Accrual: 400,
				Status:  accrual.TypeStatusProcessed,
			},
			"4929972884676289": {
				Order:   "4929972884676289",
				Accrual: 999999,
				Status:  accrual.TypeStatusProcessed,
			},
			"1984": {
				Order:  "1984",
				Status: accrual.TypeStatusProcessing,
			},
			"4532733309529845": {
				Order:  "4532733309529845",
				Status: accrual.TypeStatusRegistered,
			},
		},
	}

	mockLoyaltyStorager := MockLoyaltyStorager{
		Records: map[string]*Order{
			"4929972884676289": {
				UserID:   "e713ebf8-bb4b-11ef-9718-a7e5292ccfb8",
				ID:       "4929972884676289",
				Status:   accrual.TypeStatusProcessed,
				Accrual:  999999,
				Uploaded: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	type fields struct {
		accrual accrual.Accrualler
		storage LoyaltyStorager
	}
	type args struct {
		userID  string
		orderID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "UploadExistingOrder",
			fields: fields{
				accrual: mockAccrualler,
				storage: mockLoyaltyStorager,
			},
			args: args{
				userID:  "e713ebf8-bb4b-11ef-9718-a7e5292ccfb8",
				orderID: "4929972884676289",
			},
		},
		{
			name: "UploadUnknownOrder",
			fields: fields{
				accrual: mockAccrualler,
				storage: mockLoyaltyStorager,
			},
			args: args{
				userID:  "e713ebf8-bb4b-11ef-9718-a7e5292ccfb8",
				orderID: "6014736448",
			},
			wantErr: true,
		},
		{
			name: "UploadNewOrder",
			fields: fields{
				accrual: mockAccrualler,
				storage: mockLoyaltyStorager,
			},
			args: args{
				userID:  "e713ebf8-bb4b-11ef-9718-a7e5292ccfb8",
				orderID: "4532733309529845",
			},
		},
		{
			name: "UploadInvalidOrder",
			fields: fields{
				accrual: mockAccrualler,
				storage: mockLoyaltyStorager,
			},
			args: args{
				userID:  "e713ebf8-bb4b-11ef-9718-a7e5292ccfb8",
				orderID: "2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Loyalty{
				accrual: tt.fields.accrual,
				storage: tt.fields.storage,
			}
			if err := l.UploadOrder(ctx, tt.args.userID, tt.args.orderID); (err != nil) != tt.wantErr {
				t.Errorf("Loyalty.UploadOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			storageJSON, _ := json.MarshalIndent(mockLoyaltyStorager, "", " ")
			t.Logf("LoyaltyStorage: %v", string(storageJSON))
		})
	}
}

func TestLoyalty_GetOrders(t *testing.T) {

	ctx := context.Background()

	mockAccrualler := MockAccrualler{
		Orders: map[string]*accrual.OrderInfo{
			"79927398713": {
				Order:   "79927398713",
				Accrual: 331.3,
				Status:  accrual.TypeStatusProcessing,
			},
			"3938230889": {
				Order:   "3938230889",
				Accrual: 0,
				Status:  accrual.TypeStatusInvalid,
			},
			"4929972884676289": {
				Order:   "4929972884676289",
				Accrual: 999999,
				Status:  accrual.TypeStatusProcessed,
			},
		},
	}

	mockLoyaltyStorager := MockLoyaltyStorager{
		Records: map[string]*Order{
			"4929972884676289": {
				UserID:   "e713ebf8-bb4b-11ef-9718-a7e5292ccfb8",
				ID:       "4929972884676289",
				Status:   accrual.TypeStatusProcessed,
				Accrual:  999999,
				Uploaded: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			"3938230889": {
				UserID:   "5c18f4b8-bbb8-11ef-bd1a-8bd0750e0c51",
				ID:       "3938230889",
				Status:   accrual.TypeStatusProcessed,
				Accrual:  331.3,
				Uploaded: time.Date(2012, 3, 10, 5, 4, 0, 0, time.UTC),
			},
			"79927398713": {
				UserID:   "5c18f4b8-bbb8-11ef-bd1a-8bd0750e0c51",
				ID:       "79927398713",
				Status:   accrual.TypeStatusProcessed,
				Accrual:  331.3,
				Uploaded: time.Date(2012, 3, 10, 5, 4, 0, 0, time.UTC),
			},
		},
	}
	type fields struct {
		accrual accrual.Accrualler
		storage LoyaltyStorager
	}
	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "SimpleGetFirstUserOrders",
			fields: fields{
				accrual: mockAccrualler,
				storage: mockLoyaltyStorager,
			},
			args: args{
				userID: "e713ebf8-bb4b-11ef-9718-a7e5292ccfb8",
			},
			want: "[{\"number\":\"4929972884676289\",\"status\":\"PROCESSED\",\"accrual\":999999,\"uploaded\":\"2000-01-01T00:00:00Z\"}]",
		},
		{
			name: "SimpleGetSecondUserOrders",
			fields: fields{
				accrual: mockAccrualler,
				storage: mockLoyaltyStorager,
			},
			args: args{
				userID: "5c18f4b8-bbb8-11ef-bd1a-8bd0750e0c51",
			},
			want: "[{\"number\":\"3938230889\",\"status\":\"PROCESSED\",\"accrual\":331.3,\"uploaded\":\"2012-03-10T05:04:00Z\"},{\"number\":\"79927398713\",\"status\":\"PROCESSED\",\"accrual\":331.3,\"uploaded\":\"2012-03-10T05:04:00Z\"}]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Loyalty{
				accrual: tt.fields.accrual,
				storage: tt.fields.storage,
			}
			got, err := l.GetOrders(ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Loyalty.GetOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotJSON, err := json.Marshal(got)

			if err != nil {
				t.Errorf("Can`t marshall, error: %v", err)
				return
			}

			if string(gotJSON) != tt.want {
				t.Errorf("Loyalty.GetOrders() = %v, want %v", string(gotJSON), tt.want)
			}
		})
	}
}
