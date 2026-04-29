//go:build integration

package ordersvc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/bitik/backend/internal/pgxutil"
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type invRow struct {
	quantityAvailable int32
	quantityReserved  int32
}

func cleanupAll(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, _ = pool.Exec(ctx, `
TRUNCATE TABLE
  shipment_tracking_events,
  shipment_labels,
  shipments,
  payments,
  refunds,
  return_requests,
  voucher_redemptions,
  order_invoices,
  order_status_history,
  order_items,
  orders,
  inventory_reservations,
  checkout_session_items,
  checkout_sessions,
  cart_items,
  carts,
  inventory_movements,
  inventory_items,
  product_variant_option_values,
  variant_option_values,
  variant_options,
  product_images,
  product_variants,
  products,
  seller_documents,
  seller_bank_accounts,
  sellers,
  seller_applications,
  user_addresses,
  user_profiles,
  oauth_identities,
  user_devices,
  refresh_tokens,
  email_verification_tokens,
  phone_otp_attempts,
  users
CASCADE;
`)
}

func seedUserSellerProduct(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	var userID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, status, email_verified) VALUES ($1, 'active', true) RETURNING id`, "buyer-"+uuid.NewString()+"@example.com").Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var sellerID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO sellers (user_id, shop_name, slug, status) VALUES ($1, $2, $3, 'active') RETURNING id`, userID, "Test Shop", "test-shop-"+uuid.NewString()[:8]).Scan(&sellerID); err != nil {
		t.Fatalf("insert seller: %v", err)
	}
	var productID uuid.UUID
	if err := pool.QueryRow(ctx, `
INSERT INTO products (seller_id, name, slug, status, min_price_cents, max_price_cents, currency)
VALUES ($1, $2, $3, 'active', 1000, 1000, 'USD')
RETURNING id`, sellerID, "Test Product", "test-product-"+uuid.NewString()[:8]).Scan(&productID); err != nil {
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
	return userID, sellerID, variantID
}

func setInventory(t *testing.T, ctx context.Context, pool *pgxpool.Pool, variantID uuid.UUID, available, reserved int32) {
	t.Helper()
	if _, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_available=$2, quantity_reserved=$3 WHERE variant_id=$1`, variantID, available, reserved); err != nil {
		t.Fatalf("set inventory: %v", err)
	}
}

func getInventoryByVariant(t *testing.T, ctx context.Context, pool *pgxpool.Pool, variantID uuid.UUID) invRow {
	t.Helper()
	var inv invRow
	if err := pool.QueryRow(ctx, `SELECT quantity_available, quantity_reserved FROM inventory_items WHERE variant_id=$1`, variantID).Scan(&inv.quantityAvailable, &inv.quantityReserved); err != nil {
		t.Fatalf("get inventory: %v", err)
	}
	return inv
}

func mustAddCartItem(t *testing.T, ctx context.Context, pool *pgxpool.Pool, cartID pgtype.UUID, variantID uuid.UUID, qty int32) uuid.UUID {
	t.Helper()
	cid, ok := pgxutil.ToUUID(cartID)
	if !ok {
		t.Fatalf("invalid cart id")
	}
	var productID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT product_id FROM product_variants WHERE id=$1`, variantID).Scan(&productID); err != nil {
		t.Fatalf("variant product: %v", err)
	}
	var id uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO cart_items (cart_id, product_id, variant_id, quantity, selected) VALUES ($1,$2,$3,$4,true) RETURNING id`, cid, productID, variantID, qty).Scan(&id); err != nil {
		t.Fatalf("insert cart item: %v", err)
	}
	return id
}

func mustPlaceOrder(t *testing.T, ctx context.Context, svc *Service, uid uuid.UUID, sessionID pgtype.UUID) uuid.UUID {
	t.Helper()
	addr := map[string]any{"name": "Test", "line1": "Somewhere", "country": "X"}
	addrBytes, _ := json.Marshal(addr)
	if _, err := svc.queries.UpdateCheckoutAddress(ctx, orderstore.UpdateCheckoutAddressParams{ID: sessionID, UserID: toUUID(uid), ShippingAddress: addrBytes}); err != nil {
		t.Fatalf("set shipping address: %v", err)
	}
	order, _, _, err := svc.placeOrderTx(ctx, toUUID(uid), sessionID)
	if err != nil {
		t.Fatalf("place order: %v", err)
	}
	id, ok := pgxutil.ToUUID(order.ID)
	if !ok {
		t.Fatalf("bad order id")
	}
	return id
}

func mustBackdateOrderPlacedAt(t *testing.T, ctx context.Context, pool *pgxpool.Pool, orderID uuid.UUID, ts time.Time) {
	t.Helper()
	if _, err := pool.Exec(ctx, `UPDATE orders SET placed_at=$2 WHERE id=$1`, orderID, ts); err != nil {
		t.Fatalf("backdate order: %v", err)
	}
}

func getOrder(t *testing.T, ctx context.Context, pool *pgxpool.Pool, orderID uuid.UUID) orderstore.Order {
	t.Helper()
	q := orderstore.New(pool)
	order, err := q.GetOrderByID(ctx, pgxutil.UUID(orderID))
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	return order
}

func toUUID(id uuid.UUID) pgtype.UUID { return pgxutil.UUID(id) }
