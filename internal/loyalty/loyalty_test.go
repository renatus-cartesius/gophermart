package loyalty

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
)

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
