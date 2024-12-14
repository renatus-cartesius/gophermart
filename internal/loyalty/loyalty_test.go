package loyalty

import (
	"testing"

	"github.com/renatus-cartesius/gophermart/internal/accrual"
)

func TestLoyalty_Orders(t *testing.T) {
	type fields struct {
		accrual accrual.Accrualler
	}
	type args struct {
		order int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Loyalty{
				accrual: tt.fields.accrual,
			}
			if err := l.Orders(tt.args.order); (err != nil) != tt.wantErr {
				t.Errorf("Loyalty.Orders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
