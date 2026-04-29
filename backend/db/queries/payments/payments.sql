-- name: ListPaymentMethodsForUser :many
SELECT id, user_id, provider, type, display_name, token_reference, is_default, metadata, deleted_at, created_at, updated_at
FROM payment_methods
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY is_default DESC, created_at DESC;

-- name: GetPaymentMethodForUser :one
SELECT id, user_id, provider, type, display_name, token_reference, is_default, metadata, deleted_at, created_at, updated_at
FROM payment_methods
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ClearDefaultPaymentMethods :exec
UPDATE payment_methods SET is_default = FALSE
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: CreatePaymentMethod :one
INSERT INTO payment_methods (user_id, provider, type, display_name, token_reference, is_default, metadata)
VALUES (sqlc.arg('user_id'), sqlc.arg('provider'), sqlc.arg('type'), sqlc.narg('display_name'), sqlc.narg('token_reference'), sqlc.arg('is_default'), COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb))
RETURNING id, user_id, provider, type, display_name, token_reference, is_default, metadata, deleted_at, created_at, updated_at;

-- name: SetDefaultPaymentMethod :one
UPDATE payment_methods
SET is_default = TRUE
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, user_id, provider, type, display_name, token_reference, is_default, metadata, deleted_at, created_at, updated_at;

-- name: DeletePaymentMethod :exec
UPDATE payment_methods
SET deleted_at = now(), is_default = FALSE
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetPaymentByID :one
SELECT id, order_id, payment_method_id, provider, provider_payment_id, status, amount_cents, currency, idempotency_key, metadata, paid_at, failed_at, created_at, updated_at
FROM payments
WHERE id = $1;

-- name: GetPaymentForUser :one
SELECT p.id, p.order_id, p.payment_method_id, p.provider, p.provider_payment_id, p.status, p.amount_cents, p.currency, p.idempotency_key, p.metadata, p.paid_at, p.failed_at, p.created_at, p.updated_at
FROM payments p
JOIN orders o ON o.id = p.order_id
WHERE p.id = $1 AND o.user_id = $2;

-- name: GetPaymentByOrderAndIdempotencyKey :one
SELECT id, order_id, payment_method_id, provider, provider_payment_id, status, amount_cents, currency, idempotency_key, metadata, paid_at, failed_at, created_at, updated_at
FROM payments
WHERE order_id = $1 AND idempotency_key = $2;

-- name: CreatePayment :one
INSERT INTO payments (order_id, payment_method_id, provider, provider_payment_id, status, amount_cents, currency, idempotency_key, metadata)
VALUES (sqlc.arg('order_id'), sqlc.narg('payment_method_id'), sqlc.arg('provider'), sqlc.narg('provider_payment_id'), COALESCE(sqlc.narg('status')::payment_status, 'pending'), sqlc.arg('amount_cents'), COALESCE(sqlc.narg('currency')::text, 'USD'), sqlc.narg('idempotency_key'), COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb))
RETURNING id, order_id, payment_method_id, provider, provider_payment_id, status, amount_cents, currency, idempotency_key, metadata, paid_at, failed_at, created_at, updated_at;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = sqlc.arg('status'),
    provider_payment_id = COALESCE(sqlc.narg('provider_payment_id')::text, provider_payment_id),
    paid_at = CASE WHEN sqlc.arg('status') = 'paid' THEN COALESCE(paid_at, now()) ELSE paid_at END,
    failed_at = CASE WHEN sqlc.arg('status') = 'failed' THEN COALESCE(failed_at, now()) ELSE failed_at END
WHERE id = sqlc.arg('id')
RETURNING id, order_id, payment_method_id, provider, provider_payment_id, status, amount_cents, currency, idempotency_key, metadata, paid_at, failed_at, created_at, updated_at;

-- name: UpdatePaymentMetadataMerge :one
UPDATE payments
SET metadata = COALESCE(metadata, '{}'::jsonb) || COALESCE(sqlc.narg('metadata_patch')::jsonb, '{}'::jsonb),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, order_id, payment_method_id, provider, provider_payment_id, status, amount_cents, currency, idempotency_key, metadata, paid_at, failed_at, created_at, updated_at;

-- name: CreateManualWaveApproval :one
INSERT INTO manual_wave_payment_approvals (payment_id, reference, sender_phone, amount_cents, currency, approved_by, approved_at, note)
VALUES (sqlc.arg('payment_id'), sqlc.arg('reference'), sqlc.narg('sender_phone'), sqlc.arg('amount_cents'), COALESCE(sqlc.narg('currency')::text, 'USD'), sqlc.arg('approved_by'), now(), sqlc.narg('note'))
RETURNING id, payment_id, reference, sender_phone, amount_cents, currency, approved_by, approved_at, rejected_by, rejected_at, note, created_at;

-- name: CreateManualWaveRejection :one
INSERT INTO manual_wave_payment_approvals (payment_id, reference, sender_phone, amount_cents, currency, rejected_by, rejected_at, note)
VALUES (sqlc.arg('payment_id'), sqlc.arg('reference'), sqlc.narg('sender_phone'), sqlc.arg('amount_cents'), COALESCE(sqlc.narg('currency')::text, 'USD'), sqlc.arg('rejected_by'), now(), sqlc.narg('note'))
RETURNING id, payment_id, reference, sender_phone, amount_cents, currency, approved_by, approved_at, rejected_by, rejected_at, note, created_at;

-- name: GetManualWaveDecisionForPayment :one
SELECT id, payment_id, reference, sender_phone, amount_cents, currency, approved_by, approved_at, rejected_by, rejected_at, note, created_at
FROM manual_wave_payment_approvals
WHERE payment_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListPendingWaveManualPayments :many
SELECT p.id, p.order_id, p.payment_method_id, p.provider, p.provider_payment_id, p.status, p.amount_cents, p.currency, p.idempotency_key, p.metadata, p.paid_at, p.failed_at, p.created_at, p.updated_at
FROM payments p
WHERE p.provider = 'wave_manual'
  AND p.status = 'pending'
ORDER BY p.created_at ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreatePODCapture :one
INSERT INTO pod_payment_captures (payment_id, captured_by, amount_cents, currency, captured_at, note)
VALUES (sqlc.arg('payment_id'), sqlc.arg('captured_by'), sqlc.arg('amount_cents'), COALESCE(sqlc.narg('currency')::text, 'USD'), now(), sqlc.narg('note'))
RETURNING id, payment_id, captured_by, amount_cents, currency, captured_at, note, created_at;

-- name: CountDeliveredShipmentsForOrder :one
SELECT COUNT(1)::bigint
FROM shipments
WHERE order_id = $1 AND status = 'delivered';

-- name: CountDeliveredShipmentsForOrderSeller :one
SELECT COUNT(1)::bigint
FROM shipments
WHERE order_id = $1 AND seller_id = $2 AND status = 'delivered';

-- name: CreateRefund :one
INSERT INTO refunds (order_id, payment_id, requested_by, reason, amount_cents, currency, metadata)
VALUES (sqlc.arg('order_id'), sqlc.narg('payment_id'), sqlc.narg('requested_by'), sqlc.narg('reason'), sqlc.arg('amount_cents'), COALESCE(sqlc.narg('currency')::text, 'USD'), COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb))
RETURNING id, order_id, payment_id, requested_by, status, reason, amount_cents, currency, reviewed_by, reviewed_at, processed_at, metadata, created_at, updated_at;

-- name: GetRefundByID :one
SELECT id, order_id, payment_id, requested_by, status, reason, amount_cents, currency, reviewed_by, reviewed_at, processed_at, metadata, created_at, updated_at
FROM refunds
WHERE id = $1;

-- name: ListRefundsForAdmin :many
SELECT id, order_id, payment_id, requested_by, status, reason, amount_cents, currency, reviewed_by, reviewed_at, processed_at, metadata, created_at, updated_at
FROM refunds
WHERE (sqlc.narg('status')::refund_status IS NULL OR status = sqlc.narg('status')::refund_status)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ReviewRefund :one
UPDATE refunds
SET status = sqlc.arg('status'),
    reviewed_by = sqlc.arg('reviewed_by'),
    reviewed_at = now()
WHERE id = sqlc.arg('id')
  AND status = 'requested'
RETURNING id, order_id, payment_id, requested_by, status, reason, amount_cents, currency, reviewed_by, reviewed_at, processed_at, metadata, created_at, updated_at;

-- name: MarkRefundProcessed :one
UPDATE refunds
SET status = 'refunded',
    processed_at = now()
WHERE id = $1
  AND status = 'approved'
RETURNING id, order_id, payment_id, requested_by, status, reason, amount_cents, currency, reviewed_by, reviewed_at, processed_at, metadata, created_at, updated_at;

-- name: CreateReturnRequest :one
INSERT INTO return_requests (order_id, order_item_id, requested_by, reason, quantity, metadata)
VALUES (sqlc.arg('order_id'), sqlc.narg('order_item_id'), sqlc.narg('requested_by'), sqlc.narg('reason'), COALESCE(sqlc.narg('quantity')::int, 1), COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb))
RETURNING id, order_id, order_item_id, requested_by, status, reason, quantity, reviewed_by, reviewed_at, received_at, metadata, created_at, updated_at;

-- name: InsertPaymentWebhookEvent :one
INSERT INTO payment_webhook_events (provider, event_id, event_type, payload)
VALUES (sqlc.arg('provider'), sqlc.arg('event_id'), sqlc.narg('event_type'), sqlc.arg('payload'))
ON CONFLICT (provider, event_id) DO UPDATE SET payload = EXCLUDED.payload
RETURNING id, provider, event_id, event_type, payload, processed_at, created_at;

-- name: MarkPaymentWebhookProcessed :one
UPDATE payment_webhook_events
SET processed_at = now()
WHERE id = $1 AND processed_at IS NULL
RETURNING id, provider, event_id, event_type, payload, processed_at, created_at;

-- name: GetOrCreateSellerWallet :one
INSERT INTO seller_wallets (seller_id, currency)
VALUES (sqlc.arg('seller_id'), COALESCE(sqlc.narg('currency')::text, 'USD'))
ON CONFLICT (seller_id) DO UPDATE SET updated_at = now()
RETURNING id, seller_id, balance_cents, pending_balance_cents, currency, updated_at;

-- name: LockSellerWallet :one
SELECT id, seller_id, balance_cents, pending_balance_cents, currency, updated_at
FROM seller_wallets
WHERE seller_id = $1
FOR UPDATE;

-- name: CreateWalletTransaction :one
INSERT INTO seller_wallet_transactions (seller_wallet_id, type, amount_cents, balance_before_cents, balance_after_cents, reference_type, reference_id, description)
VALUES (sqlc.arg('seller_wallet_id'), sqlc.arg('type'), sqlc.arg('amount_cents'), sqlc.arg('balance_before_cents'), sqlc.arg('balance_after_cents'), sqlc.narg('reference_type'), sqlc.narg('reference_id'), sqlc.narg('description'))
RETURNING id, seller_wallet_id, type, amount_cents, balance_before_cents, balance_after_cents, reference_type, reference_id, description, created_at;

-- name: ListWalletTransactions :many
SELECT id, seller_wallet_id, type, amount_cents, balance_before_cents, balance_after_cents, reference_type, reference_id, description, created_at
FROM seller_wallet_transactions
WHERE seller_wallet_id = $1
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpdateWalletBalances :one
UPDATE seller_wallets
SET balance_cents = sqlc.arg('balance_cents'),
    pending_balance_cents = sqlc.arg('pending_balance_cents'),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, seller_id, balance_cents, pending_balance_cents, currency, updated_at;

-- name: CreatePayoutRequest :one
INSERT INTO seller_payouts (seller_id, amount_cents, currency, status, provider, provider_payout_id)
VALUES (sqlc.arg('seller_id'), sqlc.arg('amount_cents'), COALESCE(sqlc.narg('currency')::text, 'USD'), COALESCE(sqlc.narg('status')::text, 'pending'), sqlc.narg('provider'), sqlc.narg('provider_payout_id'))
RETURNING id, seller_id, amount_cents, currency, status, provider, provider_payout_id, requested_at, processed_at;

-- name: ListPayoutsForSeller :many
SELECT id, seller_id, amount_cents, currency, status, provider, provider_payout_id, requested_at, processed_at
FROM seller_payouts
WHERE seller_id = $1
ORDER BY requested_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpdatePayoutStatus :one
UPDATE seller_payouts
SET status = sqlc.arg('status'),
    provider = COALESCE(sqlc.narg('provider')::text, provider),
    provider_payout_id = COALESCE(sqlc.narg('provider_payout_id')::text, provider_payout_id),
    processed_at = CASE WHEN sqlc.arg('status') IN ('processed','failed','cancelled') THEN COALESCE(processed_at, now()) ELSE processed_at END
WHERE id = sqlc.arg('id')
RETURNING id, seller_id, amount_cents, currency, status, provider, provider_payout_id, requested_at, processed_at;

-- name: GetPayoutByID :one
SELECT id, seller_id, amount_cents, currency, status, provider, provider_payout_id, requested_at, processed_at
FROM seller_payouts
WHERE id = $1;

-- name: LockPayoutByID :one
SELECT id, seller_id, amount_cents, currency, status, provider, provider_payout_id, requested_at, processed_at
FROM seller_payouts
WHERE id = $1
FOR UPDATE;

-- name: ListUnsettledSellerOrderCredits :many
SELECT s.id AS seller_id, o.id AS order_id, SUM(oi.total_price_cents)::bigint AS amount_cents, o.currency
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
JOIN sellers s ON s.id = oi.seller_id
LEFT JOIN seller_wallets sw ON sw.seller_id = s.id
WHERE o.status = 'completed'
  AND o.completed_at IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM seller_wallet_transactions swt
    WHERE swt.seller_wallet_id = sw.id
      AND swt.reference_type = 'order'
      AND swt.reference_id = o.id
  )
GROUP BY s.id, o.id, o.currency
ORDER BY o.completed_at ASC
LIMIT sqlc.arg('limit');

-- name: ListOrderCreditsPendingHoldRelease :many
SELECT sw.id AS seller_wallet_id, sw.seller_id, sw.currency, swt.reference_id AS order_id, SUM(swt.amount_cents)::bigint AS amount_cents
FROM seller_wallet_transactions swt
JOIN seller_wallets sw ON sw.id = swt.seller_wallet_id
WHERE swt.reference_type = 'order'
  AND swt.reference_id IS NOT NULL
  AND swt.created_at <= now() - (sqlc.arg('older_than_days')::int || ' days')::interval
  AND NOT EXISTS (
    SELECT 1 FROM seller_wallet_transactions swt2
    WHERE swt2.seller_wallet_id = sw.id
      AND swt2.reference_type = 'hold_release'
      AND swt2.reference_id = swt.reference_id
  )
GROUP BY sw.id, sw.seller_id, sw.currency, swt.reference_id
ORDER BY MIN(swt.created_at) ASC
LIMIT sqlc.arg('limit');
