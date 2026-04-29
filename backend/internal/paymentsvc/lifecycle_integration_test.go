//go:build integration

package paymentsvc

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/ordersvc"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/goosemigrate"
	paymentstore "github.com/bitik/backend/internal/store/payments"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func cleanupAllPayments(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, _ = pool.Exec(ctx, `
TRUNCATE TABLE
  payment_webhook_events,
  pod_payment_captures,
  manual_wave_payment_approvals,
  payments,
  seller_wallet_transactions,
  seller_payouts,
  seller_wallets,
  refunds,
  return_requests,
  shipment_tracking_events,
  shipment_labels,
  shipments,
  voucher_redemptions,
  order_invoices,
  order_status_history,
  order_items,
  orders,
  inventory_reservations,
  inventory_movements,
  inventory_items,
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

func seedOrderWithReservation(t *testing.T, ctx context.Context, pool *pgxpool.Pool, qty int32) (buyer uuid.UUID, seller uuid.UUID, orderID uuid.UUID, paymentID uuid.UUID, variantID uuid.UUID) {
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
	var vID uuid.UUID
	if err := pool.QueryRow(ctx, `
INSERT INTO product_variants (product_id, sku, price_cents, currency, is_active)
VALUES ($1, $2, 1000, 'USD', true)
RETURNING id`, productID, "SKU-"+uuid.NewString()[:12]).Scan(&vID); err != nil {
		t.Fatalf("insert variant: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO inventory_items (product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold) VALUES ($1, $2, 10, 0, 1)`, productID, vID); err != nil {
		t.Fatalf("insert inventory: %v", err)
	}

	var invID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM inventory_items WHERE variant_id=$1`, vID).Scan(&invID); err != nil {
		t.Fatalf("load inventory id: %v", err)
	}

	var oID uuid.UUID
	if err := pool.QueryRow(ctx, `
INSERT INTO orders (order_number, user_id, status, subtotal_cents, total_cents, currency, placed_at)
VALUES ($1, $2, 'pending_payment', 1000, 1000, 'USD', now())
RETURNING id`, "ORD-"+uuid.NewString()[:10], userID).Scan(&oID); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO order_items (order_id, seller_id, product_id, variant_id, product_name, quantity, unit_price_cents, total_price_cents, currency, status)
VALUES ($1,$2,$3,$4,'Test Product',$5,1000,$6,'USD','pending_payment')`,
		oID, sellerID, productID, vID, qty, int64(qty)*1000,
	); err != nil {
		t.Fatalf("insert order item: %v", err)
	}

	if _, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity_reserved=$2 WHERE id=$1`, invID, qty); err != nil {
		t.Fatalf("reserve inventory: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO inventory_reservations (checkout_session_id, order_id, inventory_item_id, quantity, reserved_at)
VALUES (NULL, $1, $2, $3, now())`, oID, invID, qty); err != nil {
		t.Fatalf("insert reservation: %v", err)
	}

	var pID uuid.UUID
	if err := pool.QueryRow(ctx, `
INSERT INTO payments (order_id, provider, status, amount_cents, currency, idempotency_key, metadata)
VALUES ($1,'wave_manual','pending',1000,'USD',$2,'{}'::jsonb)
RETURNING id`, oID, "idem-"+uuid.NewString()).Scan(&pID); err != nil {
		t.Fatalf("insert payment: %v", err)
	}

	return userID, sellerID, oID, pID, vID
}

func invByVariant(t *testing.T, ctx context.Context, pool *pgxpool.Pool, variantID uuid.UUID) (avail int32, reserved int32) {
	t.Helper()
	if err := pool.QueryRow(ctx, `SELECT quantity_available, quantity_reserved FROM inventory_items WHERE variant_id=$1`, variantID).Scan(&avail, &reserved); err != nil {
		t.Fatalf("inv: %v", err)
	}
	return
}

func TestLifecycle_WaveApprove_TransitionsOrder_ConsumesReservation(t *testing.T) {
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

	cleanupAllPayments(t, ctx, pool)
	_, _, orderID, paymentID, variantID := seedOrderWithReservation(t, ctx, pool, 2)

	beforeAvail, beforeReserved := invByVariant(t, ctx, pool, variantID)
	if beforeReserved != 2 {
		t.Fatalf("expected reserved=2 got %d", beforeReserved)
	}

	paySvc := NewService(config.Config{}, zap.NewNop(), pool)
	actor := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	// Mimic Wave approval: mark payment paid then transition order to paid (consume reservations).
	_, err = paySvc.pay.CreateManualWaveApproval(ctx, paymentstore.CreateManualWaveApprovalParams{
		PaymentID:   pgxutil.UUID(paymentID),
		Reference:   "WAVE-REF-1",
		SenderPhone: text(""),
		AmountCents: 1000,
		Currency:    text("USD"),
		ApprovedBy:  actor,
		Note:        text("ok"),
	})
	if err != nil {
		t.Fatalf("approval insert: %v", err)
	}
	payment, err := paySvc.pay.UpdatePaymentStatus(ctx, paymentstore.UpdatePaymentStatusParams{ID: pgxutil.UUID(paymentID), Status: "paid"})
	if err != nil {
		t.Fatalf("payment paid: %v", err)
	}
	if payment.Status != "paid" {
		t.Fatalf("expected payment paid got %s", payment.Status)
	}

	orderSvc := ordersvc.NewService(config.Config{}, zap.NewNop(), pool, nil)
	_, err = orderSvc.TransitionOrder(ctx, pgxutil.UUID(orderID), "paid", "wave approved", actor)
	if err != nil {
		t.Fatalf("order transition: %v", err)
	}

	afterAvail, afterReserved := invByVariant(t, ctx, pool, variantID)
	if afterAvail != beforeAvail-2 {
		t.Fatalf("expected available %d got %d", beforeAvail-2, afterAvail)
	}
	if afterReserved != 0 {
		t.Fatalf("expected reserved 0 got %d", afterReserved)
	}
}

func TestLifecycle_WalletSettlement_And_HoldRelease(t *testing.T) {
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

	cleanupAllPayments(t, ctx, pool)

	// Minimal completed order for settlement.
	var userID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, status, email_verified) VALUES ($1, 'active', true) RETURNING id`, "buyer-"+uuid.NewString()+"@example.com").Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var sellerID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO sellers (user_id, shop_name, slug, status) VALUES ($1, $2, $3, 'active') RETURNING id`, userID, "Test Shop", "test-shop-"+uuid.NewString()[:8]).Scan(&sellerID); err != nil {
		t.Fatalf("insert seller: %v", err)
	}
	var orderID uuid.UUID
	if err := pool.QueryRow(ctx, `
INSERT INTO orders (order_number, user_id, status, subtotal_cents, total_cents, currency, placed_at, completed_at)
VALUES ($1, $2, 'completed', 1000, 1000, 'USD', now(), now())
RETURNING id`, "ORD-"+uuid.NewString()[:10], userID).Scan(&orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO order_items (order_id, seller_id, product_id, variant_id, product_name, quantity, unit_price_cents, total_price_cents, currency, status)
VALUES ($1,$2,$3,$4,'X',1,1000,1000,'USD','completed')`, orderID, sellerID, uuid.New(), uuid.New()); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	svc := NewService(config.Config{}, zap.NewNop(), pool)
	if err := svc.applySettlementCredit(ctx, pgxutil.UUID(sellerID), pgxutil.UUID(orderID), 1000); err != nil {
		t.Fatalf("settle: %v", err)
	}

	wallet, err := svc.pay.GetOrCreateSellerWallet(ctx, paymentstore.GetOrCreateSellerWalletParams{SellerID: pgxutil.UUID(sellerID)})
	if err != nil {
		t.Fatalf("wallet: %v", err)
	}
	if wallet.PendingBalanceCents != 1000 {
		t.Fatalf("expected pending 1000 got %d", wallet.PendingBalanceCents)
	}

	// Backdate settlement tx so hold-release can pick it up.
	if _, err := pool.Exec(ctx, `UPDATE seller_wallet_transactions SET created_at=$2 WHERE reference_type='order' AND reference_id=$1`, orderID, time.Now().UTC().Add(-10*24*time.Hour)); err != nil {
		t.Fatalf("backdate tx: %v", err)
	}
	if err := svc.releaseHold(ctx, pgxutil.UUID(sellerID), pgxutil.UUID(orderID), 1000); err != nil {
		t.Fatalf("release: %v", err)
	}
	wallet, err = svc.pay.GetOrCreateSellerWallet(ctx, paymentstore.GetOrCreateSellerWalletParams{SellerID: pgxutil.UUID(sellerID)})
	if err != nil {
		t.Fatalf("wallet: %v", err)
	}
	if wallet.BalanceCents != 1000 || wallet.PendingBalanceCents != 0 {
		t.Fatalf("expected balance=1000 pending=0 got balance=%d pending=%d", wallet.BalanceCents, wallet.PendingBalanceCents)
	}
}

func TestLifecycle_RefundApproval_TransitionsOrder_ReleasesReservation(t *testing.T) {
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

	cleanupAllPayments(t, ctx, pool)
	_, _, orderID, paymentID, variantID := seedOrderWithReservation(t, ctx, pool, 2)

	paySvc := NewService(config.Config{}, zap.NewNop(), pool)
	// Create requested refund linked to payment.
	refund, err := paySvc.pay.CreateRefund(ctx, paymentstore.CreateRefundParams{
		OrderID:     pgxutil.UUID(orderID),
		PaymentID:   pgxutil.UUID(paymentID),
		RequestedBy: pgtype.UUID{},
		Reason:      text("test"),
		AmountCents: 1000,
		Currency:    text("USD"),
		Metadata:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("create refund: %v", err)
	}

	actor := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err = paySvc.pay.ReviewRefund(ctx, paymentstore.ReviewRefundParams{ID: refund.ID, Status: "approved", ReviewedBy: actor})
	if err != nil {
		t.Fatalf("review refund: %v", err)
	}
	_, err = paySvc.pay.MarkRefundProcessed(ctx, refund.ID)
	if err != nil {
		t.Fatalf("mark processed: %v", err)
	}

	orderSvc := ordersvc.NewService(config.Config{}, zap.NewNop(), pool, nil)
	_, err = orderSvc.TransitionOrder(ctx, pgxutil.UUID(orderID), "refunded", "refund approved", actor)
	if err != nil {
		t.Fatalf("order refunded: %v", err)
	}

	_, afterReserved := invByVariant(t, ctx, pool, variantID)
	if afterReserved != 0 {
		t.Fatalf("expected reserved 0 got %d", afterReserved)
	}
}
