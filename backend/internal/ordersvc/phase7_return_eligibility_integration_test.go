//go:build integration

package ordersvc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/goosemigrate"
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func TestLifecycle_ReturnRequiresDeliveredShipmentForSameSeller(t *testing.T) {
	dsn := os.Getenv("BITIK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("BITIK_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	// Run migrations for standalone execution.
	if err := goosemigrate.RunFromDSN(dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cleanupAll(t, ctx, pool)

	// Buyer
	buyerID := mustInsertUser(t, ctx, pool, "buyer-"+uuid.NewString()+"@example.com")

	// Seller A + product
	sellerAOwner := mustInsertUser(t, ctx, pool, "sellerA-"+uuid.NewString()+"@example.com")
	sellerA := mustInsertSeller(t, ctx, pool, sellerAOwner, "Shop A")
	variantA := mustInsertProductVariant(t, ctx, pool, sellerA, "Product A")
	setInventory(t, ctx, pool, variantA, 10, 0)

	// Seller B + product
	sellerBOwner := mustInsertUser(t, ctx, pool, "sellerB-"+uuid.NewString()+"@example.com")
	sellerB := mustInsertSeller(t, ctx, pool, sellerBOwner, "Shop B")
	variantB := mustInsertProductVariant(t, ctx, pool, sellerB, "Product B")
	setInventory(t, ctx, pool, variantB, 10, 0)

	svc := NewService(config.Config{}, zap.NewNop(), pool)
	cart, err := svc.queries.GetOrCreateCart(ctx, orderstore.GetOrCreateCartParams{UserID: toUUID(buyerID)})
	if err != nil {
		t.Fatalf("cart: %v", err)
	}
	_ = mustAddCartItem(t, ctx, pool, cart.ID, variantA, 1)
	_ = mustAddCartItem(t, ctx, pool, cart.ID, variantB, 1)

	session, _, err := svc.createCheckoutSessionTx(ctx, buyerID)
	if err != nil {
		t.Fatalf("checkout create: %v", err)
	}
	orderID := mustPlaceOrder(t, ctx, svc, buyerID, session.ID)

	// Deliver only seller A shipment.
	if _, err := svc.queries.UpdateShipmentForSeller(ctx, orderstore.UpdateShipmentForSellerParams{
		OrderID:        pgxutil.UUID(orderID),
		SellerID:       pgxutil.UUID(sellerA),
		Status:         "delivered",
		TrackingNumber: text("T-A"),
	}); err != nil {
		t.Fatalf("deliver shipment A: %v", err)
	}

	// Find an order_item for seller B.
	var itemB uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM order_items WHERE order_id=$1 AND seller_id=$2 LIMIT 1`, orderID, sellerB).Scan(&itemB); err != nil {
		t.Fatalf("order item B: %v", err)
	}

	bodyBytes, _ := json.Marshal(map[string]any{
		"order_item_id": itemB.String(),
		"quantity":      1,
		"reason":        "changed mind",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/buyer/returns", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.AuthUserIDKey, buyerID)

	svc.HandleBuyerRequestReturn(c)
	if w.Code != 400 {
		t.Fatalf("expected 400 got %d body=%s", w.Code, w.Body.String())
	}
}

func mustInsertUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, status, email_verified) VALUES ($1, 'active', true) RETURNING id`, email).Scan(&id); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func mustInsertSeller(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, shopName string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO sellers (user_id, shop_name, slug, status) VALUES ($1, $2, $3, 'active') RETURNING id`,
		userID, shopName, strings.ToLower(strings.ReplaceAll(shopName, " ", "-"))+"-"+uuid.NewString()[:8],
	).Scan(&id); err != nil {
		t.Fatalf("insert seller: %v", err)
	}
	return id
}

func mustInsertProductVariant(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sellerID uuid.UUID, productName string) uuid.UUID {
	t.Helper()
	var productID uuid.UUID
	if err := pool.QueryRow(ctx, `
INSERT INTO products (seller_id, name, slug, status, min_price_cents, max_price_cents, currency)
VALUES ($1, $2, $3, 'active', 1000, 1000, 'USD')
RETURNING id`, sellerID, productName, strings.ToLower(strings.ReplaceAll(productName, " ", "-"))+"-"+uuid.NewString()[:8]).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	var variantID uuid.UUID
	if err := pool.QueryRow(ctx, `
INSERT INTO product_variants (product_id, sku, price_cents, currency, is_active)
VALUES ($1, $2, 1000, 'USD', true)
RETURNING id`, productID, "SKU-"+uuid.NewString()[:12]).Scan(&variantID); err != nil {
		t.Fatalf("insert variant: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO inventory_items (product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold) VALUES ($1, $2, 0, 0, 1) ON CONFLICT (variant_id) DO NOTHING`, productID, variantID); err != nil {
		t.Fatalf("insert inventory: %v", err)
	}
	return variantID
}
