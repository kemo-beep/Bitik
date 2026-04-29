//go:build integration

package sellersvc

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func TestConcurrentInventoryAdjustments(t *testing.T) {
	dsn := os.Getenv("BITIK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("BITIK_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	userID := uuid.New()
	sellerID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()
	itemID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, status, email_verified) VALUES ($1, $2, 'x', 'active', true);
		INSERT INTO sellers (id, user_id, shop_name, slug, status) VALUES ($3, $1, 'Concurrent Seller', $4, 'active');
		INSERT INTO products (id, seller_id, name, slug, status, min_price_cents, max_price_cents) VALUES ($5, $3, 'Concurrent Product', $6, 'draft', 100, 100);
		INSERT INTO product_variants (id, product_id, sku, price_cents) VALUES ($7, $5, $8, 100);
		INSERT INTO inventory_items (id, product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold) VALUES ($9, $5, $7, 100, 0, 5);
	`, userID, userID.String()+"@bitik.test", sellerID, "seller-"+sellerID.String(), productID, "product-"+productID.String(), variantID, "sku-"+variantID.String(), itemID)
	if err != nil {
		t.Fatal(err)
	}

	svc := NewService(config.Config{}, zap.NewNop(), pool)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := svc.adjustInventoryTx(ctx, pgxutil.UUID(sellerID), pgxutil.UUID(itemID), pgxutil.UUID(userID), nil, nil, nil, -1, 0, true, "concurrent test", "test", pgxutil.UUID(uuid.New()))
			if err != nil {
				t.Errorf("adjustInventoryTx: %v", err)
			}
		}()
	}
	wg.Wait()

	var available int32
	if err := pool.QueryRow(ctx, `SELECT quantity_available FROM inventory_items WHERE id = $1`, itemID).Scan(&available); err != nil {
		t.Fatal(err)
	}
	if available != 90 {
		t.Fatalf("expected 90 available, got %d", available)
	}
}
