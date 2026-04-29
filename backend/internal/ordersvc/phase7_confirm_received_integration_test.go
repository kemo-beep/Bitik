//go:build integration

package ordersvc

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/goosemigrate"
	orderstore "github.com/bitik/backend/internal/store/orders"
	paymentstore "github.com/bitik/backend/internal/store/payments"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func TestLifecycle_ConfirmReceived_CompletesAndSettles(t *testing.T) {
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
	_ = mustAddCartItem(t, ctx, pool, cart.ID, variantID, 1)

	session, _, err := svc.createCheckoutSessionTx(ctx, uid)
	if err != nil {
		t.Fatalf("checkout create: %v", err)
	}
	orderID := mustPlaceOrder(t, ctx, svc, uid, session.ID)

	// Mark shipment delivered.
	if _, err := svc.queries.UpdateShipmentForSeller(ctx, orderstore.UpdateShipmentForSellerParams{
		OrderID:        pgxutil.UUID(orderID),
		SellerID:       pgxutil.UUID(sellerID),
		Status:         "delivered",
		TrackingNumber: text("T123"),
	}); err != nil {
		t.Fatalf("deliver shipment: %v", err)
	}

	// Call handler directly with a test gin context.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/buyer/orders/"+orderID.String()+"/confirm-received", nil)
	c.Params = gin.Params{{Key: "order_id", Value: orderID.String()}}
	c.Set(middleware.AuthUserIDKey, uid)

	svc.HandleBuyerConfirmReceived(c)
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}

	// Wallet should have pending credit for seller (settlement trigger).
	pq := paymentstore.New(pool)
	wallet, err := pq.GetOrCreateSellerWallet(ctx, paymentstore.GetOrCreateSellerWalletParams{SellerID: pgxutil.UUID(sellerID)})
	if err != nil {
		t.Fatalf("wallet: %v", err)
	}
	if wallet.PendingBalanceCents <= 0 {
		t.Fatalf("expected pending balance credited, got %d", wallet.PendingBalanceCents)
	}
}
