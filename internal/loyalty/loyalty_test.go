package loyalty

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
)

type MockAccrualler struct {
	Orders map[int64]*accrual.OrderInfo
}

func (ma MockAccrualler) GetOrder(ctx context.Context, orderID int64) (*accrual.OrderInfo, error) {
	order, ok := ma.Orders[orderID]
	if !ok {
		return nil, accrual.ErrOrderNotFound
	}
	return order, nil
}

type MockLoyaltyStorager struct {
	Records map[int64]*OrderRecord
}

func (mls MockLoyaltyStorager) AddOrder(ctx context.Context, userID string, order *accrual.OrderInfo) error {
	orderRecord, ok := mls.Records[order.Order]
	if ok {
		if orderRecord.UserID != userID {
			return ErrOrderUploadedAnotherUser
		}
		return nil
	}

	mls.Records[order.Order] = &OrderRecord{
		UserID:   userID,
		Number:   order.Order,
		Status:   order.Status,
		Accrual:  order.Accrual,
		Uploaded: time.Now(),
	}
	return nil
}

func (mls MockLoyaltyStorager) GetOrders(ctx context.Context, userID string) ([]*OrderRecord, error) {
	res := make([]*OrderRecord, 0)

	for _, record := range mls.Records {
		if record.UserID == userID {
			res = append(res, record)
		}
	}

	return res, nil
}

func TestLoyalty_UploadOrder(t *testing.T) {

	ctx := context.Background()

	mockAccrualler := MockAccrualler{
		Orders: map[int64]*accrual.OrderInfo{
			79927398713: {
				Order:   79927398713,
				Accrual: 400,
				Status:  accrual.TypeStatusProcessed,
			},
			4929972884676289: {
				Order:   4929972884676289,
				Accrual: 999999,
				Status:  accrual.TypeStatusProcessed,
			},
			1984: {
				Order:  1984,
				Status: accrual.TypeStatusProcessing,
			},
			4532733309529845: {
				Order:  4532733309529845,
				Status: accrual.TypeStatusRegistered,
			},
		},
	}

	mockLoyaltyStorager := MockLoyaltyStorager{
		Records: map[int64]*OrderRecord{
			4929972884676289: {
				UserID:   "e713ebf8-bb4b-11ef-9718-a7e5292ccfb8",
				Number:   4929972884676289,
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
		orderID int64
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
				orderID: 4929972884676289,
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
				orderID: 2,
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
				orderID: 4532733309529845,
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
				orderID: 4532733309529845,
			},
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
			storageJson, _ := json.MarshalIndent(mockLoyaltyStorager, "", " ")
			t.Logf("LoyaltyStorage: %v", string(storageJson))
		})
	}
}
