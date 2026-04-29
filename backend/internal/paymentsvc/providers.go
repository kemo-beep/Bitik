package paymentsvc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	orderstore "github.com/bitik/backend/internal/store/orders"
	paymentstore "github.com/bitik/backend/internal/store/payments"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	errPODNotEligible     = errors.New("pod not eligible")
	errPODNotDelivered    = errors.New("pod cannot be captured before delivery")
	errInvalidProvider    = errors.New("invalid provider")
	errOrderNotPendingPay = errors.New("order not pending payment")
	errPaymentNotPending  = errors.New("payment not pending")
)

type providerAdapter interface {
	Name() string
	ValidateCreateIntent(order orderstore.Order) error
	ValidateCapture(ctx context.Context, s *Service, paymentOrderID pgtype.UUID, actorSellerID pgtype.UUID, actorIsSeller bool) error
}

func (s *Service) provider(name string) (providerAdapter, error) {
	switch strings.TrimSpace(name) {
	case "wave_manual":
		return waveManualAdapter{}, nil
	case "pod":
		return podAdapter{}, nil
	default:
		return nil, errInvalidProvider
	}
}

type waveManualAdapter struct{}

func (waveManualAdapter) Name() string { return "wave_manual" }
func (waveManualAdapter) ValidateCreateIntent(order orderstore.Order) error {
	if statusString(order.Status) != "pending_payment" {
		return errOrderNotPendingPay
	}
	return nil
}
func (waveManualAdapter) ValidateCapture(ctx context.Context, s *Service, paymentOrderID pgtype.UUID, actorSellerID pgtype.UUID, actorIsSeller bool) error {
	// Wave manual is never captured by seller; it is approved via admin/ops endpoints.
	return nil
}

type podAdapter struct{}

func (podAdapter) Name() string { return "pod" }

func (podAdapter) ValidateCreateIntent(order orderstore.Order) error {
	if statusString(order.Status) != "pending_payment" {
		return errOrderNotPendingPay
	}

	// v1 conservative eligibility rules (can be replaced with city/risk tables later):
	// - requires shipping_address.city and shipping_address.country
	// - caps order total to reduce fraud exposure (default 2000.00 if cents)
	if order.TotalCents > 200000 {
		return errPODNotEligible
	}

	var addr map[string]any
	if len(order.ShippingAddress) == 0 {
		return errPODNotEligible
	}
	if err := json.Unmarshal(order.ShippingAddress, &addr); err != nil {
		return errPODNotEligible
	}
	city, _ := addr["city"].(string)
	country, _ := addr["country"].(string)
	if strings.TrimSpace(city) == "" || strings.TrimSpace(country) == "" {
		return errPODNotEligible
	}
	return nil
}

func (podAdapter) ValidateCapture(ctx context.Context, s *Service, paymentOrderID pgtype.UUID, actorSellerID pgtype.UUID, actorIsSeller bool) error {
	if actorIsSeller {
		n, err := s.pay.CountDeliveredShipmentsForOrderSeller(ctx, paymentstore.CountDeliveredShipmentsForOrderSellerParams{OrderID: paymentOrderID, SellerID: actorSellerID})
		if err != nil {
			return err
		}
		if n == 0 {
			return errPODNotDelivered
		}
		return nil
	}
	n, err := s.pay.CountDeliveredShipmentsForOrder(ctx, paymentOrderID)
	if err != nil {
		return err
	}
	if n == 0 {
		return errPODNotDelivered
	}
	return nil
}
