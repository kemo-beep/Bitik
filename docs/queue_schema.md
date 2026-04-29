# Queue Schema (Phase 10)

## Exchanges

- `bitik.jobs` (topic): primary job exchange.
- `bitik.jobs.retry` (topic): delayed retry exchange.
- `bitik.jobs.dlq` (topic): dead-letter exchange.

## Envelope v1

All jobs use this JSON envelope:

```json
{
  "message_id": "uuid",
  "job_type": "worker.cancel_unpaid_orders",
  "schema_version": "v1",
  "created_at": "2026-04-28T11:00:00Z",
  "trace_id": "optional-trace-id",
  "dedupe_key": "cancel_unpaid:2026-04-28T11:00",
  "attempt": 1,
  "payload": {}
}
```

## Retry / DLQ

- Max attempts: `8`.
- Retry path: consumer publish to `bitik.jobs.retry` with same routing key.
- DLQ path: broker dead-letters to `<queue>.dlq` queue after retries fail.
- Base retry delay: queue-level delayed retry queue (TTL 30s in v1).

## Routing Keys

- `notify.email.send`
- `notify.sms.otp.send`
- `notify.push.send`
- `payments.timeout.confirmation`
- `payments.wave.stale_order_timeout`
- `orders.invoice.generate`
- `media.image.process`
- `search.product.index`
- `search.product.reindex_full`
- `orders.checkout.expire`
- `orders.cancel_unpaid`
- `orders.inventory.release_expired`
- `shipping.tracking.refresh`
- `wallets.settle_seller`
- `wallets.payout.process`
- `reports.generate`
- `notify.fanout`
