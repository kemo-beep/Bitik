# Architecture Overview (Phase 12)

This document is the launch-readiness architecture snapshot for Bitik backend + web.

## Runtime Components

- API service (`backend/cmd/api`): public/admin/internal HTTP endpoints, migrations on startup (configurable), OpenAPI serving.
- Worker service (`backend/cmd/worker`): RabbitMQ consumers, retry/DLQ handling, scheduled jobs, idempotency ledger.
- PostgreSQL: system of record, analytics rollups, audit/admin activity logs, worker job ledgers.
- Redis: rate limit buckets, idempotency state, auth lockout counters.
- RabbitMQ: async jobs for notifications, indexing, payments, reports, media processing.
- OpenSearch: product indexing/search and related relevance tracking.
- MinIO/object storage: media asset persistence and presigned upload flows.

## Deployment Topology

- Staging and production deploy to Docker hosts.
- Images built from `backend/Dockerfile`.
- CD flow executes migration gate -> deploy -> smoke test.
- Production workflow includes rollback step to previous image tag on failure.

## Test Strategy

- Unit: services and parser/validator/state-transition helpers.
- Integration (`-tags=integration`): queue, storage, search, and persistence dependencies.
- Contract: OpenAPI MVP endpoint presence and consistency checks.
- E2E: Playwright critical journeys (`@critical`) plus smoke.
- Load: k6 scripts for catalog/detail/search/cart-checkout/fanout.
