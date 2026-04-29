# Bitik Backend Production Task Plan

## Goal

Build a production-ready Shopee-like marketplace backend for Bitik that supports buyer shopping, seller center operations, admin moderation, checkout, manual Wave payment approval, pay on delivery, media uploads, search, notifications, chat, analytics, and operational tooling.

## Source Inputs

- Architecture baseline: `docs/rough.md`
- API surface: `api_list.md`
- Database baseline: `db_tables.md`

## Core Technical Requirements

- Language/framework: Go with Gin.
- Database: PostgreSQL on Neon for production, local PostgreSQL for development.
- SQL tooling: sqlc for typed queries, Goose for migrations, pgx for database access.
- Cache/session/locks: Redis with go-redis.
- Background jobs: RabbitMQ for v1 workers; keep boundaries ready for Kafka event streaming and NATS real-time fanout later.
- Search: OpenSearch for product, seller, brand, and category search.
- Object storage: Cloudflare R2 in production and MinIO locally.
- CDN: Cloudflare CDN for public media.
- Auth: JWT access tokens, hashed refresh tokens, OAuth, phone OTP, email verification, device/session management.
- Authorization: Casbin/RBAC for admin and staff permissions; seller and buyer ownership checks at handler/service layer.
- Observability: structured logging, metrics, tracing, audit logs, Sentry, Prometheus, Grafana, Loki, Uptime Kuma, OpenTelemetry.
- Deployment: Docker first, Dokploy for early production, Kubernetes-ready service boundaries later.
- Security: secure password hashing, rate limits, request validation, CORS, CSRF-aware web flows, idempotency keys, webhook signing, secret management.

## Production Acceptance Criteria

- All MVP APIs are implemented, documented, tested, and deployed behind `/api/v1`.
- Database migrations can run forward from an empty database and support seeded development data.
- The checkout and order state machines are explicit, transactional, and safe under concurrent requests.
- Payments support pay on delivery and manual Wave approval in v1; future provider webhooks are isolated behind provider adapters.
- Seller, buyer, and admin permissions are enforced with tests for cross-tenant access.
- Background jobs are retryable, idempotent, observable, and dead-lettered.
- Media upload uses presigned upload flows, validates file types/sizes, and stores durable metadata.
- Search results are indexed asynchronously and recoverable through reindex jobs.
- API docs are published through Swagger/OpenAPI and match implemented behavior.
- CI runs formatting, linting, unit tests, integration tests, sqlc generation checks, and migration checks.
- Production runbooks exist for deploys, rollback, database migrations, payment approval, incident response, and worker failures.

## Phase 0: Product And Backend Foundation

### Requirements

- Confirm v1 countries, cities, currencies, languages, tax assumptions, seller verification rules, shipping policy, and payment methods.
- Define exact order statuses, payment statuses, shipment statuses, refund/return statuses, and allowed transitions.
- Decide v1 API compatibility rules, versioning policy, pagination format, error envelope, date/time format, and money format.
- Decide local development stack and environment variable contract.

### Tasks

- Create backend repository structure:
  - `cmd/api`
  - `cmd/worker`
  - `cmd/ws`
  - `internal/config`
  - `internal/http`
  - `internal/middleware`
  - `internal/modules`
  - `internal/platform/db`
  - `internal/platform/cache`
  - `internal/platform/queue`
  - `internal/platform/storage`
  - `internal/platform/search`
  - `internal/platform/observability`
- Add Docker Compose for PostgreSQL, Redis, RabbitMQ, MinIO, OpenSearch, Jaeger/Tempo-compatible tracing, and local mail/SMS stubs.
- Add configuration loading with Viper or equivalent and typed config structs.
- Add standard API response envelope:
  - success payloads
  - validation errors
  - authorization errors
  - domain errors
  - trace/request IDs
- Add global middleware:
  - request ID
  - structured logging
  - panic recovery
  - CORS
  - rate limiting
  - auth parsing
  - idempotency middleware for mutating checkout/payment/order endpoints
- Add health endpoints: `GET /health`, `GET /ready`, `GET /metrics`, `GET /version`.

### Deliverables

- Backend skeleton builds locally and in CI.
- Local stack starts with one command.
- API conventions documented.
- OpenAPI/Swagger baseline available.
- Health, readiness, metrics, and version endpoints working.

## Phase 1: Database, Migrations, And Data Access

### Requirements

- Convert `db_tables.md` into production-grade Goose migrations.
- Keep schema compatible with the marketplace API surface while adding missing production fields.
- Use sqlc for all core SQL access.
- Make migrations repeatable in local, staging, and production.

### Tasks

- Create initial migration for extensions, enums, and base tables from `db_tables.md`.
- Review and add missing v1 schema elements:
  - email verification tokens
  - phone OTP attempts and verification records
  - OAuth identities
  - user devices
  - seller applications and seller documents
  - checkout sessions and checkout session items
  - payment methods
  - manual Wave payment approvals
  - POD payment capture records
  - refunds and return requests
  - moderation reports/cases
  - CMS pages, banners, FAQs, announcements
  - feature flags and platform settings
  - notification preferences
  - review votes and reports
  - search recent/click tracking
  - shipment labels and provider metadata
- Add `created_at`, `updated_at`, soft-delete, status, and audit fields where needed.
- Add indexes for all high-volume filters:
  - user/order/product/seller lookups
  - status dashboards
  - search/index jobs
  - payment status and provider identifiers
  - inventory reservations expiry
  - notifications unread counts
- Add constraints for money, quantities, ratings, uniqueness, and ownership invariants.
- Add SQL query packages per domain.
- Add seed data:
  - admin user and roles
  - default permissions
  - sample categories/brands
  - sample seller and products for local development
  - shipping providers
  - platform settings
- Add migration checks in CI.

### Deliverables

- Goose migration set.
- sqlc query files and generated typed stores.
- Seed command for development/staging.
- Schema diagram or table reference.
- Migration rollback policy for production.

## Phase 2: Authentication, Users, Sessions, And RBAC

### Requirements

- Support account registration, login, logout, refresh, password reset, email verification, phone OTP, OAuth, and session/device management.
- Enforce user status and ban/deletion rules.
- Protect admin APIs with RBAC.

### Tasks

- Implement `POST /auth/register`.
- Implement `POST /auth/login`.
- Implement `POST /auth/logout`.
- Implement `POST /auth/refresh-token`.
- Implement forgot/reset password token flows.
- Implement email verification and resend endpoints.
- Implement phone OTP send/verify with provider adapter and rate limits.
- Implement Google, Apple, and Facebook OAuth initiation/callback flows.
- Store refresh token hashes with rotation and reuse detection.
- Track sessions and devices from user agent, IP, platform, and push token metadata.
- Implement `GET /users/me`, `PATCH /users/me`, `DELETE /users/me`.
- Implement profile endpoints.
- Implement sessions/devices listing and revoke endpoints.
- Implement RBAC tables, seed roles, permission keys, and Casbin policy loading.
- Add admin auth endpoints.
- Add audit logs for sensitive auth/admin actions.

### Deliverables

- Complete auth API module.
- JWT and refresh token lifecycle tests.
- OTP and OAuth provider adapter contracts.
- RBAC seed data and permission matrix.
- Security documentation for auth and session behavior.

### Implementation status (backend, Phase 2)

- **Done:** Register/login/logout/refresh; forgot/reset password; verify/resend email; phone OTP with Redis rate limit + injectable `OTPSender`; Google, Facebook, and Apple OAuth (Redis CSRF state, provider-verified email checks); refresh hash storage, rotation, reuse lockout, and session-bound refresh invalidation; DB-backed active user/role enforcement for protected JWT routes; `user_sessions` + device metadata; `GET/PATCH/DELETE /users/me`, profile, sessions/devices list + revoke; SQL RBAC + Casbin policies from `casbin_rule`; admin/seller stubs behind JWT + active-user + Casbin middleware; `audit_logs` on sensitive actions; `backend/docs/AUTH.md`, `RBAC_MATRIX.md`; JWT/helper tests and optional refresh lifecycle integration test; OTP/OAuth provider adapter contracts in `internal/authsvc/contracts.go`; API runs Goose **up** on startup when `database.auto_migrate` is true (default).
- **Follow-up:** Casbin reload at runtime is not implemented; production email/SMS providers still need concrete `EmailSender` / `OTPSender` implementations wired from infrastructure code; Apple id_token is parsed after token exchange for profile claims and can be extended with full JWKS signature verification.
- **Ops:** `docker compose up --remove-orphans` clears orphan warnings; enum migration is idempotent for DBs missing `goose_db_version` sync.

## Phase 3: Public Catalog, Categories, Brands, And Media

### Requirements

- Buyers can browse categories, brands, sellers, products, variants, images, reviews, related products, banners, and homepage sections.
- Sellers and admins can upload product/shop media through safe storage flows.

### Tasks

- Implement public catalog endpoints:
  - `GET /public/home`
  - `GET /public/home/sections`
  - `GET /public/banners`
  - `GET /public/categories`
  - `GET /public/categories/{category_id}`
  - `GET /public/categories/{category_id}/products`
  - `GET /public/brands`
  - `GET /public/brands/{brand_id}`
  - `GET /public/brands/{brand_id}/products`
  - `GET /public/products`
  - `GET /public/products/{product_id}`
  - `GET /public/products/slug/{slug}`
  - `GET /public/products/{product_id}/variants`
  - `GET /public/products/{product_id}/reviews`
  - `GET /public/products/{product_id}/related`
  - `GET /public/sellers/{seller_id}`
  - `GET /public/sellers/{seller_id}/products`
  - `GET /public/sellers/{seller_id}/reviews`
  - `GET /public/sellers/slug/{slug}`
- Implement product list filters, sorting, and pagination.
- Implement media endpoints:
  - `POST /media/upload`
  - `POST /media/upload/presigned-url`
  - `GET /media/files`
  - `GET /media/files/{file_id}`
  - `DELETE /media/files/{file_id}`
  - seller/admin scoped media uploads
- Add R2/MinIO storage adapter.
- Validate file MIME type, extension, size, dimensions, ownership, and malware-scanning hook.
- Generate public URLs and store object metadata.
- Add background image processing job hooks.

### Deliverables

- Public catalog module.
- Media storage module with local and production providers.
- Catalog fixture data.
- Product list performance tests.
- OpenAPI docs for catalog/media endpoints.

### Implementation status (backend, Phase 3)

- **Done:** Public catalog module under `internal/catalogsvc`; homepage, banners, categories, brands, sellers, products, variants, reviews, and related-products endpoints; product list filters, sort options, pagination metadata, and product search query support; active-product plus active-seller public visibility enforcement; invalid filter rejection; Phase 3 catalog/media support migration for list/search indexes and search-vector maintenance; catalog fixture data for banners, homepage settings, images, products, variants, and reviews; seller/admin media module under `internal/mediasvc` with MinIO/R2-compatible storage, multipart uploads, pending/complete presigned uploads, file listing/detail/delete, byte-sniffed MIME/extension/size/dimension validation, ownership checks, malware scanner hook, and image processor hook; OpenAPI path docs; unit coverage and product list mapping benchmark.
- **Follow-up:** Production malware scanning and image processing workers need concrete implementations wired through `mediasvc` options; abandoned pending presigned reservations should be cleaned by a scheduled reconciliation job.

## Phase 4: Seller Onboarding, Seller Center, Products, Variants, And Inventory

### Requirements

- Sellers can apply, manage shop profile, create products, upload images, manage options/variants, publish/unpublish products, and control inventory.
- Admins can approve/reject/suspend sellers and moderate products.
- Inventory changes must be auditable and safe under concurrent checkout.

### Tasks

- Implement seller onboarding:
  - `POST /seller/apply`
  - `GET /seller/application`
  - `PATCH /seller/application`
  - `POST /seller/documents`
  - `DELETE /seller/documents/{document_id}`
- Implement seller profile:
  - `GET /seller/me`
  - `PATCH /seller/me`
  - `PATCH /seller/me/profile`
  - `PATCH /seller/me/settings`
  - logo and banner upload endpoints
- Implement seller dashboard:
  - stats
  - sales chart
  - top products
  - recent orders
  - low stock
- Implement seller products:
  - list, create, read, update, delete
  - publish, unpublish, duplicate
  - image upload, reorder, delete
  - variants create/update/delete
  - options create/update/delete
- Implement seller inventory:
  - list and detail
  - update and adjust
  - movement history
  - low-stock
  - bulk update
- Add product status transition rules.
- Add inventory movement records for every adjustment.
- Add row-level locking or equivalent transactional checks for reservation and release.
- Add admin product/seller moderation support required by v1.

### Deliverables

- Seller onboarding/product/inventory modules.
- Inventory concurrency tests.
- Seller permission and ownership tests.
- Product moderation workflow tests.
- Seller center dashboard API docs.

### Implementation status (backend, Phase 4)

- **Done:** Seller center module under `internal/sellersvc`; seller application create/read/update/submit with omitted metadata preserved; seller document create/delete; transactional admin application approval with seller creation and role assignment; seller profile/settings/media updates; dashboard stats, sales chart, top products, recent orders, and low-stock data with explicit error handling; seller product list/create/read/update/delete/publish/unpublish/duplicate; product image create/list/reorder/delete; variant create/update/delete with transactional inventory creation and initial stock movement; option/value create/list/delete; inventory list/detail/set/adjust/movement history/low-stock/bulk update; row-level inventory locking, all-or-nothing bulk updates, reserved <= available invariants, and movement records containing actor and before/after quantities; admin seller status and product moderation endpoints; Phase 4 migration for seller settings, product moderation, inventory movement audit fields, and seller-center indexes; seller center OpenAPI paths and docs.
- **Tests:** Unit coverage for seller helper behavior, application metadata preservation, inventory invariants, and movement quantity calculations; optional `integration` build-tag concurrency test for inventory row locking via `BITIK_TEST_DATABASE_URL`.
- **Follow-up:** Admin audit-log writes can be added around moderation actions once the admin activity log API is centralized; product image endpoints currently accept validated media URLs from Phase 3 rather than streaming files directly.

## Phase 5: Cart, Checkout, Orders, And State Machines

### Requirements

- Buyers can manage cart, create checkout sessions, select address/shipping/payment, apply vouchers, validate totals, and place orders.
- Checkout must snapshot prices, product names, variants, shipping address, and seller split details.
- Order state transitions must be explicit and auditable.

### Tasks

- Implement address APIs.
- Implement cart APIs:
  - get cart
  - add item
  - update item
  - delete item
  - clear cart
  - merge guest cart
  - select items
  - apply/remove voucher
- Implement checkout APIs:
  - create session
  - get session
  - update address
  - update shipping
  - update payment method
  - apply/remove voucher
  - validate
  - place order
- Implement buyer order APIs:
  - list, detail, items, status history
  - cancel
  - confirm received
  - request refund/return
  - dispute
  - invoice
  - tracking
- Implement seller order APIs:
  - list, detail, items
  - accept/reject
  - mark processing
  - pack
  - ship
  - cancel
  - refund
  - approve/reject return
- Implement admin order APIs.
- Define order aggregate behavior for multi-seller carts:
  - one parent order with seller-scoped order items, or seller sub-orders if added to schema.
  - status rollup rules from item/seller shipment states.
- Add inventory reservation on checkout/session or place-order, with expiry and release job.
- Add idempotency to place-order and status mutations.
- Add invoice generation job.

### Deliverables

- Cart, checkout, and order modules.
- State-machine documentation.
- Transactional checkout tests.
- Multi-seller order tests.
- Inventory reservation and release worker.
- Invoice generation worker.

### Implementation status (backend, Phase 5)

- **Done:** Buyer address APIs; buyer cart get/add/update/delete/clear/select/merge and voucher apply/remove; checkout session create/get/address/shipping/payment/voucher/validate/place-order; transactional checkout item snapshots; row-locked inventory reservations with reserve/release movements; buyer order list/detail/items/history/cancel/confirm/refund/return/dispute/invoice/tracking; seller order list/detail/items/accept/reject/processing/pack/ship/cancel/refund/return review; admin order list/detail/status/cancel/refund/history/payments/shipments; invoice persistence; internal reservation release and invoice generation worker endpoints; Phase 5 support migration for cart selection, voucher links, and invoices; state-machine docs.
- **Tests:** Unit coverage for order transition rules and voucher discount calculation. Full transactional checkout and multi-seller integration tests should be run against a test database as the next hardening layer.

## Phase 6: Payments, Manual Wave Approval, POD, Refunds, And Wallets

### Requirements

- V1 supports pay on delivery and manual Wave merchant confirmation.
- Payment state must be consistent with order state.
- All payment mutations must be idempotent, audited, and permission-protected.
- Future provider webhooks must be possible without rewriting order/payment state machines.

### Tasks

- Define payment providers:
  - `pod`
  - `wave_manual`
  - future provider adapter interface
- Extend payment status handling for manual Wave pending confirmation if needed.
- Implement buyer payment APIs:
  - create intent
  - confirm
  - get payment
  - retry
  - cancel
  - payment methods list/create/delete/default
- Implement manual Wave flow:
  - buyer sees exact amount, order reference, and Wave business account instructions.
  - order/payment remains pending manual confirmation.
  - admin or authorized ops user can approve/reject with note.
  - approval records approver ID, timestamp, evidence/note, and audit log.
  - stale pending payments expire through worker.
- Implement POD flow:
  - eligibility by city/order value/risk rules.
  - mark payment as payable on delivery.
  - capture/mark paid on delivery by authorized seller/admin/shipping role.
- Implement refunds:
  - buyer request
  - seller/admin approval
  - payment refund record
  - order/refund status transitions
- Implement payment webhook event table and generic ingestion for future providers.
- Implement seller wallet and payout accounting:
  - wallet balance
  - transactions
  - payout requests
  - bank account management
  - settlement job after completed orders
- Keep Stripe/PayPal/ECPay/LinePay/JKOPay endpoints as later provider tasks unless explicitly required for launch.

### Deliverables

- Payment module with Wave manual and POD providers.
- Admin payment approval APIs.
- Refund and payout modules.
- Payment runbook for manual Wave verification.
- Idempotency and payment-state tests.

## Phase 7: Shipping, Tracking, Returns, And Delivery Operations

### Requirements

- Buyers, sellers, and admins can view and manage shipments.
- Seller can create shipment records, print labels when provider supports it, mark shipped, and update tracking.
- Buyers can track shipments and request returns/refunds.

### Tasks

- Implement buyer shipping APIs.
- Implement seller shipping settings, shipments, labels, mark-shipped, tracking APIs.
- Implement admin shipping provider and shipment APIs.
- Add provider adapter interface for manual/local shipping and future logistics providers.
- Add tracking events table integration and update-shipment-tracking worker.
- Add return request workflow tied to order items and shipments.
- Add delivery confirmation and completed-order settlement triggers.

### Deliverables

- Shipping module.
- Shipment provider adapter.
- Tracking event ingestion worker.
- Return/refund workflow tests.
- Seller and buyer tracking API docs.

## Phase 8: Search, Promotions, Reviews, Notifications, And Chat

### Requirements

- Search is fast, filterable, observable, and recoverable.
- Promotions/vouchers are validated consistently in cart and checkout.
- Reviews support verified purchase checks, images, voting, reporting, seller replies, and moderation.
- Notifications and chat support API-first behavior with WebSocket delivery later.

### Tasks

- Search:
  - implement `/public/search`, suggestions, trending, recent, click tracking.
  - add OpenSearch mappings for products, categories, brands, and sellers.
  - add index-product and reindex-products workers.
  - add fallback behavior if OpenSearch is degraded.
- Promotions/vouchers:
  - buyer voucher listing/claim/validate.
  - public promotions/campaigns/flash sales.
  - seller vouchers/promotions.
  - admin vouchers/campaigns/flash sales.
  - checkout integration with usage limits and redemption records.
- Reviews:
  - buyer CRUD.
  - image upload/delete.
  - vote and report.
  - seller reply/report.
  - admin hide/unhide/delete and reported review moderation.
- Notifications:
  - buyer notification list/unread/read/delete.
  - preferences.
  - push token registration/delete.
  - worker fanout for email/SMS/push/in-app.
- Chat:
  - conversations and messages.
  - attachments.
  - read receipts.
  - deletion rules.
  - `/ws/chat`, `/ws/notifications`, `/ws/order-status` service design.
  - NATS-ready internal pub/sub boundary.

### Deliverables

- Search, promotions, reviews, notifications, and chat modules.
- OpenSearch mapping and reindex command.
- Notification provider adapters.
- WebSocket service baseline.
- Moderation tests.

## Phase 9: Admin, CMS, Moderation, Settings, Analytics, And Internal APIs

### Requirements

- Admin panel APIs cover users, sellers, products, orders, payments, shipping, promotions, moderation, CMS, RBAC, settings, audit logs, activity logs, and system health.
- Internal job APIs are private and protected.

### Tasks

- Implement admin dashboard stats and charts.
- Implement admin user management.
- Implement admin seller management and approval workflow.
- Implement admin product management and reported products.
- Implement admin category and brand management.
- Implement admin order/payment/shipping APIs.
- Implement admin reviews and moderation cases.
- Implement CMS pages, banners, FAQs, announcements.
- Implement RBAC management APIs.
- Implement settings and feature flags.
- Implement audit logs, activity logs, and system health.
- Implement analytics/event logging for:
  - product views
  - search clicks
  - cart actions
  - checkout start/place-order
  - payment approval/rejection
  - seller product/order actions
- Implement private internal worker APIs with network and token restrictions.

### Deliverables

- Complete admin API module.
- CMS module.
- Moderation workflow.
- Analytics event pipeline.
- Internal API authentication policy.
- Admin permission tests.

## Phase 10: Background Workers And Eventing

### Requirements

- Workers must be reliable, retryable, idempotent, observable, and safe to restart.
- RabbitMQ is the v1 queue. Kafka and NATS are future integrations behind internal interfaces.

### Tasks

- Add worker process with queue consumers.
- Define queue names, routing keys, retry policy, dead-letter queues, and message schemas.
- Implement workers for:
  - send email
  - send SMS/OTP
  - send push
  - payment confirmation timeout
  - manual Wave stale-order timeout
  - invoice generation
  - image processing
  - product indexing
  - full product reindex
  - expire checkout
  - cancel unpaid orders
  - release expired inventory
  - update shipment tracking
  - settle seller wallets
  - process payouts
  - generate seller/admin reports
  - notification fanout
- Add job deduplication and idempotency.
- Add worker metrics and dashboards.
- Add replay/requeue runbook.

### Deliverables

- Worker service.
- Queue schema documentation.
- DLQ and retry dashboards.
- Job tests and idempotency tests.
- Operational runbook.

## Phase 11: Security, Compliance, And Production Hardening

### Requirements

- Backend should be safe for public internet exposure and marketplace financial workflows.

### Tasks

- Add secure password hashing with Argon2id or bcrypt.
- Add request body limits and upload limits.
- Add rate limits by IP, user, route, and auth risk.
- Add account lockout/risk signals for repeated failures.
- Add ownership checks for every buyer/seller/admin scoped resource.
- Add webhook signature verification for future automated providers.
- Add CORS policy per environment.
- Add secure cookie policy if web uses cookies.
- Add PII handling policy and data deletion workflow.
- Add audit logs for admin, payment, seller, product, order, and auth mutations.
- Add dependency scanning and container scanning in CI.
- Add secret management through Doppler, Infisical, or cloud secret manager.
- Add backup and restore procedures for PostgreSQL and object storage metadata.
- Add database connection pooling and timeout policies.
- Add graceful shutdown for API, worker, and WebSocket services.

### Deliverables

- Security checklist.
- Threat model for auth, checkout, payments, seller/admin workflows.
- Audit log coverage report.
- Backup/restore runbook.
- Rate-limit configuration.

## Phase 12: Testing, Documentation, CI/CD, And Launch Readiness

### Requirements

- Every production domain has automated tests and operational documentation.
- CI/CD prevents broken migrations, generated-code drift, and unsafe deploys.

### Tasks

- Unit tests for validators, services, state machines, pricing, vouchers, inventory, payments, auth, and permissions.
- Integration tests for DB queries, migrations, Redis, RabbitMQ, storage, and OpenSearch.
- API contract tests for MVP endpoints.
- E2E tests for critical flows:
  - register/login
  - seller apply and product publish
  - buyer browse/search/cart/checkout
  - Wave manual payment approval
  - POD order lifecycle
  - seller ship order
  - buyer confirm received
  - review submission
  - admin moderation
- Load tests for:
  - catalog listing
  - product detail
  - search
  - cart updates
  - checkout/place-order
  - notification fanout
- CI pipeline:
  - format
  - lint
  - sqlc generate check
  - migration up/down check
  - unit tests
  - integration tests
  - OpenAPI generation check
  - Docker build
  - vulnerability scan
- CD pipeline:
  - staging deploy
  - migration gate
  - smoke tests
  - production deploy
  - rollback path
- Documentation:
  - API docs
  - architecture overview
  - database schema reference
  - environment variables
  - local setup
  - deployment runbook
  - payment approval runbook
  - incident response runbook

### Deliverables

- Full CI/CD workflow.
- Test suite with coverage report.
- OpenAPI documentation.
- Production runbooks.
- Launch checklist.

## MVP Build Order

1. Foundation, migrations, auth, users, RBAC baseline.
2. Public categories/products/search and media upload.
3. Seller application, seller profile, product, variant, and inventory management.
4. Cart, checkout sessions, place-order, inventory reservation.
5. Buyer/seller/admin order workflows.
6. Wave manual payment and POD payment flows.
7. Notifications, reviews, basic admin dashboard.
8. Search indexing, workers, observability, and production hardening.

## MVP API Completion Checklist

- [ ] `POST /auth/register`
- [ ] `POST /auth/login`
- [ ] `POST /auth/refresh-token`
- [ ] `GET /users/me`
- [ ] `PATCH /users/me`
- [ ] `GET /public/categories`
- [ ] `GET /public/products`
- [ ] `GET /public/products/{product_id}`
- [ ] `GET /public/search`
- [ ] `POST /seller/apply`
- [ ] `GET /seller/me`
- [ ] `PATCH /seller/me`
- [ ] `POST /seller/products`
- [ ] `GET /seller/products`
- [ ] `PATCH /seller/products/{product_id}`
- [ ] `POST /seller/products/{product_id}/variants`
- [ ] `PATCH /seller/inventory/{inventory_item_id}`
- [ ] `GET /buyer/cart`
- [ ] `POST /buyer/cart/items`
- [ ] `PATCH /buyer/cart/items/{cart_item_id}`
- [ ] `DELETE /buyer/cart/items/{cart_item_id}`
- [ ] `POST /buyer/checkout/sessions`
- [ ] `POST /buyer/checkout/sessions/{checkout_session_id}/place-order`
- [ ] `GET /buyer/orders`
- [ ] `GET /buyer/orders/{order_id}`
- [ ] `POST /buyer/payments/create-intent`
- [ ] Wave manual approval endpoints
- [ ] POD payment lifecycle endpoints
- [ ] `GET /seller/orders`
- [ ] `POST /seller/orders/{order_id}/ship`
- [ ] `POST /buyer/reviews`
- [ ] `GET /buyer/notifications`
- [ ] `GET /admin/dashboard`
- [ ] `GET /admin/users`
- [ ] `GET /admin/sellers`
- [ ] `GET /admin/products`
- [ ] `GET /admin/orders`

## Open Decisions

- Final launch country/currency and whether all prices should default to local currency instead of USD.
- Whether multi-seller checkout creates one order with seller-scoped items or separate seller sub-orders.
- Exact seller verification requirements and required document types.
- Exact shipping providers and whether label creation is manual or integrated for v1.
- Whether mobile and web auth use bearer tokens only, secure cookies only, or both.
- Whether Stripe/PayPal-style provider endpoints are deferred fully or kept as stubbed adapters.
