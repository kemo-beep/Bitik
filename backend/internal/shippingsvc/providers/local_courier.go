package providers

import (
	"context"
	"encoding/json"
	"strings"

	shippingstore "github.com/bitik/backend/internal/store/shipping"
)

type localCourierAdapter struct{}

func (localCourierAdapter) Code() string { return "local-courier" }

func (localCourierAdapter) CreateLabel(ctx context.Context, shipment shippingstore.Shipment, provider shippingstore.ShippingProvider) (string, map[string]any, error) {
	// v1 stub: build a deterministic local label url; store tracking url template if present.
	meta := map[string]any{}
	if len(provider.Metadata) > 0 {
		_ = json.Unmarshal(provider.Metadata, &meta)
	}
	labelURL := "https://labels.local/" + shipment.ID.String() + ".pdf"
	if tpl, ok := meta["tracking_url_template"].(string); ok && strings.TrimSpace(tpl) != "" {
		return labelURL, map[string]any{"tracking_url_template": strings.TrimSpace(tpl)}, nil
	}
	return labelURL, map[string]any{}, nil
}
