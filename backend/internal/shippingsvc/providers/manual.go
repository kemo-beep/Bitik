package providers

import (
	"context"
	"errors"

	shippingstore "github.com/bitik/backend/internal/store/shipping"
)

type manualAdapter struct{}

func (manualAdapter) Code() string { return "manual" }

func (manualAdapter) CreateLabel(ctx context.Context, shipment shippingstore.Shipment, provider shippingstore.ShippingProvider) (string, map[string]any, error) {
	return "", nil, errors.New("labels_not_supported")
}
