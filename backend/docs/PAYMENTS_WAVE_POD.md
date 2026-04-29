# Payments (Wave Manual + POD)

This document describes how Phase 6 payments work in Bitik, with a focus on **manual Wave approval** and **Pay-on-Delivery (POD)**.

## Provider models

- **`wave_manual`**: Buyer initiates payment and submits evidence; payment remains `pending` until an **admin** or **`ops_payments`** approver accepts/rejects it.
- **`pod`**: Buyer chooses pay-on-delivery; payment remains `pending` until a **seller** or **admin/ops** captures it.

## Idempotency

All payment mutations are protected by the idempotency middleware for routes containing `/payments/`.

- **Create intent** (`POST /api/v1/buyer/payments/create-intent`) requires `Idempotency-Key`.
- If the client retries with the same key, the API returns the originally created payment.

## Wave manual flow

### Buyer

1. Call `POST /api/v1/buyer/payments/create-intent` with provider `wave_manual`.
2. The response includes buyer-facing instructions from `platform_settings.wave_manual_instructions` (JSON string).
3. Buyer can call `POST /api/v1/buyer/payments/confirm` (currently a no-op that returns the payment).

### Admin / Ops approver

- **Approve**: `POST /api/v1/admin/payments/{payment_id}/wave/approve`
  - Creates a decision row in `manual_wave_payment_approvals`
  - Sets `payments.status = paid`
  - Transitions the order to `paid` (which consumes inventory reservations via Phase 5 logic)
- **Reject**: `POST /api/v1/admin/payments/{payment_id}/wave/reject`
  - Creates a decision row in `manual_wave_payment_approvals`
  - Sets `payments.status = failed`

Wave decisions are guarded to avoid double-decision per payment.

## POD flow

1. Buyer creates intent with provider `pod`.
2. Seller/admin captures the payment:
   - Seller: `POST /api/v1/seller/payments/{payment_id}/pod/capture`
   - Admin/Ops: `POST /api/v1/admin/payments/{payment_id}/pod/capture`
3. Capture:
   - Inserts `pod_payment_captures`
   - Sets `payments.status = paid`
   - Transitions the order to `paid` (consumes reservations)

## Webhooks (skeleton)

Incoming provider webhooks are stored for later processing:

- `POST /api/v1/webhooks/{provider}`
  - Requires `X-Event-Id`
  - Stores raw payload in `payment_webhook_events`

