//go:build integration

package mediasvc

import (
	"context"
	"os"
	"testing"

	"github.com/bitik/backend/internal/config"
	platformstorage "github.com/bitik/backend/internal/platform/storage"
)

func TestIntegration_ObjectStorageConnect(t *testing.T) {
	endpoint := os.Getenv("BITIK_STORAGE_ENDPOINT")
	accessKey := os.Getenv("BITIK_STORAGE_ACCESS_KEY_ID")
	secret := os.Getenv("BITIK_STORAGE_SECRET_ACCESS_KEY")
	bucket := os.Getenv("BITIK_STORAGE_BUCKET")
	if endpoint == "" || accessKey == "" || secret == "" || bucket == "" {
		t.Skip("storage env variables not configured")
	}
	cfg := config.StorageConfig{
		Endpoint:        endpoint,
		AccessKeyID:     accessKey,
		SecretAccessKey: secret,
		Bucket:          bucket,
		UseSSL:          false,
	}
	client, err := platformstorage.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("storage connect failed: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil storage client")
	}
}
