package providers

import (
	"context"

	shippingstore "github.com/bitik/backend/internal/store/shipping"
)

type Adapter interface {
	Code() string
	CreateLabel(ctx context.Context, shipment shippingstore.Shipment, provider shippingstore.ShippingProvider) (labelURL string, metadata map[string]any, err error)
}

func ByCode(code string) Adapter {
	switch code {
	case "local-courier":
		return localCourierAdapter{}
	default:
		return manualAdapter{}
	}
}
