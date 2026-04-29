package shippingsvc

import "testing"

func TestCanTransitionShipment(t *testing.T) {
	cases := []struct {
		old  string
		next string
		want bool
	}{
		{"pending", "packed", true},
		{"pending", "shipped", true},
		{"packed", "shipped", true},
		{"shipped", "in_transit", true},
		{"in_transit", "delivered", true},
		{"delivered", "shipped", false},
		{"pending", "delivered", false},
	}
	for _, tc := range cases {
		if got := canTransitionShipment(tc.old, tc.next); got != tc.want {
			t.Fatalf("canTransitionShipment(%q,%q)=%v want %v", tc.old, tc.next, got, tc.want)
		}
	}
}

