//go:build integration

package searchsvc

import (
	"context"
	"os"
	"testing"

	"github.com/bitik/backend/internal/config"
	platformsearch "github.com/bitik/backend/internal/platform/search"
)

func TestIntegration_OpenSearchHealthCheck(t *testing.T) {
	searchURL := os.Getenv("BITIK_SEARCH_URL")
	if searchURL == "" {
		t.Skip("BITIK_SEARCH_URL is not configured")
	}
	cfg := config.SearchConfig{
		URL:           searchURL,
		ProductsIndex: "products_v1",
	}
	if err := platformsearch.Check(context.Background(), cfg); err != nil {
		t.Fatalf("opensearch check failed: %v", err)
	}
}
