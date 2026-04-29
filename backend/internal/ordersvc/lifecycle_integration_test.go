//go:build integration

package ordersvc

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/platform/goosemigrate"
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func TestLifecycle_CheckoutReserve_CancelUnpaid_Releases(t *testing.T) {
	dsn := os.Getenv("BITIK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("BITIK_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	if err := goosemigrate.RunFromDSN(dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	cleanupAll(t, ctx, pool)
	uid, sellerID, variantID := seedUserSellerProduct(t, ctx, pool)
	setInventory(t, ctx, pool, variantID, 10, 0)

	svc := NewService(config.Config{}, zap.NewNop(), pool)
	cart, err := svc.queries.GetOrCreateCart(ctx, orderstore.GetOrCreateCartParams{UserID: toUUID(uid)})
	if err != nil {
		t.Fatalf("cart: %v", err)
	}
	_ = sellerID
	item := mustAddCartItem(t, ctx, pool, cart.ID, variantID, 2)
	_ = item

	session, _, err := svc.createCheckoutSessionTx(ctx, uid)
	if err != nil {
		t.Fatalf("checkout create: %v", err)
	}
	inv := getInventoryByVariant(t, ctx, pool, variantID)
	if inv.quantityReserved != 2 {
		t.Fatalf("expected reserved=2 got %d", inv.quantityReserved)
	}

	orderID := mustPlaceOrder(t, ctx, svc, uid, session.ID)
	mustBackdateOrderPlacedAt(t, ctx, pool, orderID, time.Now().UTC().Add(-2*time.Hour))

	cancelled, released, err := svc.cancelStaleUnpaidOrders(ctx, 10, 30)
	if err != nil {
		t.Fatalf("cancel unpaid: %v", err)
	}
	if cancelled != 1 || released != 1 {
		t.Fatalf("expected cancelled=1 released=1 got cancelled=%d released=%d", cancelled, released)
	}
	inv = getInventoryByVariant(t, ctx, pool, variantID)
	if inv.quantityReserved != 0 {
		t.Fatalf("expected reserved=0 got %d", inv.quantityReserved)
	}
}

func TestLifecycle_OrderPaid_ConsumesReservation(t *testing.T) {
	dsn := os.Getenv("BITIK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("BITIK_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	if err := goosemigrate.RunFromDSN(dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	cleanupAll(t, ctx, pool)
	uid, _, variantID := seedUserSellerProduct(t, ctx, pool)
	setInventory(t, ctx, pool, variantID, 10, 0)

	svc := NewService(config.Config{}, zap.NewNop(), pool)
	cart, err := svc.queries.GetOrCreateCart(ctx, orderstore.GetOrCreateCartParams{UserID: toUUID(uid)})
	if err != nil {
		t.Fatalf("cart: %v", err)
	}
	_ = mustAddCartItem(t, ctx, pool, cart.ID, variantID, 3)

	session, _, err := svc.createCheckoutSessionTx(ctx, uid)
	if err != nil {
		t.Fatalf("checkout create: %v", err)
	}
	orderID := mustPlaceOrder(t, ctx, svc, uid, session.ID)

	order := getOrder(t, ctx, pool, orderID)
	updated, err := svc.transitionOrder(ctx, order, "paid", "payment captured", pgtype.UUID{})
	if err != nil {
		t.Fatalf("transition paid: %v", err)
	}
	if statusString(updated.Status) != "paid" {
		t.Fatalf("expected status paid got %s", statusString(updated.Status))
	}

	inv := getInventoryByVariant(t, ctx, pool, variantID)
	if inv.quantityAvailable != 7 {
		t.Fatalf("expected available=7 got %d", inv.quantityAvailable)
	}
	if inv.quantityReserved != 0 {
		t.Fatalf("expected reserved=0 got %d", inv.quantityReserved)
	}
}
