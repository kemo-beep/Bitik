package ordersvc

import (
	"testing"

	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name    string
		current string
		next    string
		want    bool
	}{
		{name: "pending to paid", current: "pending_payment", next: "paid", want: true},
		{name: "paid to processing", current: "paid", next: "processing", want: true},
		{name: "processing to shipped", current: "processing", next: "shipped", want: true},
		{name: "terminal cancelled", current: "cancelled", next: "paid", want: false},
		{name: "skip invalid delivery", current: "paid", next: "delivered", want: false},
		{name: "dispute after shipped", current: "shipped", next: "disputed", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canTransition(tt.current, tt.next); got != tt.want {
				t.Fatalf("canTransition(%q, %q) = %v, want %v", tt.current, tt.next, got, tt.want)
			}
		})
	}
}

func TestCalculateDiscount(t *testing.T) {
	fixed := voucher("fixed", 500, 0, 0)
	if got := calculateDiscount(fixed, 2_000, 300); got != 500 {
		t.Fatalf("fixed discount = %d", got)
	}

	percentage := voucher("percentage", 20, 0, 300)
	if got := calculateDiscount(percentage, 2_000, 0); got != 300 {
		t.Fatalf("capped percentage discount = %d", got)
	}

	minOrder := voucher("fixed", 500, 5_000, 0)
	if got := calculateDiscount(minOrder, 2_000, 0); got != 0 {
		t.Fatalf("min order discount = %d", got)
	}

	freeShipping := voucher("free_shipping", 0, 0, 0)
	if got := calculateDiscount(freeShipping, 2_000, 450); got != 450 {
		t.Fatalf("free shipping discount = %d", got)
	}
}

func voucher(kind string, value, minOrder, max int64) orderstore.Voucher {
	v := orderstore.Voucher{DiscountType: kind, DiscountValue: value, MinOrderCents: minOrder}
	if max > 0 {
		v.MaxDiscountCents = pgtype.Int8{Int64: max, Valid: true}
	}
	return v
}
