package sellersvc

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/bitik/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func TestToSlug(t *testing.T) {
	got := toSlug("  Demo Seller Shop! ")
	if got != "demo-seller-shop" {
		t.Fatalf("unexpected slug: %q", got)
	}
}

func TestPaginationCapsPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest("GET", "/?page=2&per_page=999", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	p := pagination(c)
	if p.Page != 2 || p.PerPage != maxPerPage || p.Offset != maxPerPage {
		t.Fatalf("unexpected pagination: %+v", p)
	}
}

func TestHasRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(middleware.AuthRolesKey, []string{"buyer", "seller"})

	if !hasRole(c, "seller") {
		t.Fatal("expected seller role")
	}
	if hasRole(c, "admin") {
		t.Fatal("did not expect admin role")
	}
}

func TestJSONOptionalPreservesWhenOmitted(t *testing.T) {
	if got := jsonOptional(nil); got != nil {
		t.Fatalf("expected nil JSON for omitted value, got %s", got)
	}
	got := jsonOptional(map[string]any{"tier": "gold"})
	if !json.Valid(got) || !bytes.Contains(got, []byte(`"tier"`)) {
		t.Fatalf("expected marshaled JSON, got %s", got)
	}
}

func TestValidateInventoryState(t *testing.T) {
	tests := []struct {
		name      string
		available int32
		reserved  int32
		threshold int32
		wantErr   bool
	}{
		{name: "valid", available: 10, reserved: 4, threshold: 2},
		{name: "negative available", available: -1, reserved: 0, threshold: 2, wantErr: true},
		{name: "reserved above available", available: 3, reserved: 4, threshold: 2, wantErr: true},
		{name: "negative threshold", available: 3, reserved: 1, threshold: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInventoryState(tt.available, tt.reserved, tt.threshold)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
		})
	}
}

func TestMovementQuantityUsesActualDelta(t *testing.T) {
	if got := movementQuantity(10, 10, 2, 2); got != 0 {
		t.Fatalf("expected no movement for unchanged quantities, got %d", got)
	}
	if got := movementQuantity(10, 7, 2, 5); got != 6 {
		t.Fatalf("expected combined absolute delta, got %d", got)
	}
}
