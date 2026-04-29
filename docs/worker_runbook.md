# Worker Operations Runbook

## Check Worker Health

- Verify worker process logs contain `worker_ready`.
- Check Prometheus endpoint: `GET /metrics` on `BITIK_HTTP_METRICS_ADDR`.
- Watch queue counters:
  - `bitik_queue_consumed_total`
  - `bitik_queue_retried_total`
  - `bitik_queue_dlq_total`

## Inspect Queues

- Primary queues: `jobs.*.v1`.
- Retry queues: `<queue>.retry`.
- Dead-letter queues: `<queue>.dlq`.
- Use RabbitMQ UI or `rabbitmqctl list_queues name messages messages_ready messages_unacknowledged`.

## Replay / Requeue DLQ Messages

1. Inspect DLQ payload and confirm handler bug/failure is fixed.
2. Re-publish the original envelope to `bitik.jobs` with same routing key.
3. Keep the same `dedupe_key` so worker idempotency still protects duplicate side effects.
4. Monitor `bitik_queue_acked_total` and domain logs until success.

## Idempotency Notes

- Worker ledger table: `worker_job_runs`.
- Unique key: `(job_type, dedupe_key)`.
- If a row is already `done`, duplicate delivery is acknowledged without re-running handler logic.
- Use dedupe keys that represent business identity (`invoice:<order_id>`, etc).

## Scheduler Lock / Failover

- Scheduler owner is selected by Postgres advisory lock (`pg_try_advisory_lock`).
- Only lock holder publishes periodic jobs.
- On worker crash or shutdown, lock is released and another worker instance can take over on next tick.
- If scheduling stalls, verify DB connectivity and lock state in Postgres.
