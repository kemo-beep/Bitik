package providers

import (
	"context"
	"testing"

	shippingstore "github.com/bitik/backend/internal/store/shipping"
)

func TestLocalCourier_CreateLabel(t *testing.T) {
	a := ByCode("local-courier")
	sh := shippingstore.Shipment{}
	p := shippingstore.ShippingProvider{Code: "local-courier", Metadata: []byte(`{"tracking_url_template":"https://t.local/{tracking_number}"}`)}
	url, md, err := a.CreateLabel(context.Background(), sh, p)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if url == "" {
		t.Fatalf("expected label url")
	}
	if _, ok := md["tracking_url_template"]; !ok {
		t.Fatalf("expected tracking_url_template metadata")
	}
}

