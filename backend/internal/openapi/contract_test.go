package openapi

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOpenAPIMVPContract(t *testing.T) {
	var doc struct {
		Paths map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(YAML, &doc); err != nil {
		t.Fatalf("parse openapi yaml: %v", err)
	}
	required := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/buyer/checkout/sessions/{checkout_session_id}/place-order",
		"/api/v1/admin/payments/{payment_id}/wave/approve",
		"/api/v1/seller/orders/{order_id}/ship",
		"/api/v1/buyer/orders/{order_id}/confirm-received",
	}
	for _, path := range required {
		if _, ok := doc.Paths[path]; !ok {
			t.Fatalf("required MVP path missing in OpenAPI: %s", path)
		}
	}
}
