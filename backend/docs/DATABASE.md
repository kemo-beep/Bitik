# Database

Phase 1 adds the durable PostgreSQL layer for the marketplace backend.

## Layout

- `db/migrations`: modular Goose migrations. Migrations are forward-only in normal production operations and include a rollback section for local/staging validation.
- `db/seeds`: repeatable seed SQL for development and staging.
- `db/queries/<domain>`: sqlc input queries grouped by module boundary.
- `internal/store/<domain>`: generated sqlc stores used by application modules.
- `cmd/migrate`: migration runner using the configured `BITIK_DATABASE_URL`.
- `cmd/api`: when `database.auto_migrate` is true (default), applies pending migrations on startup via the same Goose files (set `BITIK_DATABASE_AUTO_MIGRATE=false` if your platform runs migrations separately).
- `cmd/seed`: repeatable seed runner using the configured `BITIK_DATABASE_URL`.

## Local Commands

```sh
make compose-up
make migrate-up
make seed
make migrate-status
```

Local Docker Compose binds Postgres to `127.0.0.1:55432` to avoid conflicts with an existing local Postgres on `5432`.

## Migration Files

Migration files are split by stable domain boundaries, not one file per table:

- `extensions_and_functions`
- `enums`
- `identity_and_auth`
- `rbac_and_admin`
- `sellers`
- `catalog`
- `inventory`
- `cart_checkout_orders`
- `payments_refunds_returns`
- `shipping`
- `promotions_reviews_wishlist`
- `chat_notifications`
- `wallet_media_cms_settings`
- `moderation_audit_events_search`
- `phase2_auth_rbac_sessions` (user sessions, password reset tokens, `casbin_rule`, refresh token `session_id`, dev bcrypt for seed users)
- `phase3_catalog_media_support` (public product-list indexes, media owner index, banner window index, product search-vector trigger)
- `phase4_seller_center_support` (seller settings, product moderation fields, inventory movement audit columns, seller-center indexes)

Create new migrations with Goose, then edit the generated file:

```sh
make migration-create name=create_example_table
```

Keep related indexes and triggers in the same migration as their table. Put cross-domain prerequisites in an earlier shared migration.

## sqlc

Queries are generated per domain:

- `auth`: refresh tokens, email verification, phone OTP, OAuth identities.
- `users`: users, profiles, addresses, devices.
- `catalog`: categories, brands, products, reviews, banners, media files, search tracking.
- `sellers`: seller onboarding/profile, seller center products, variants/options/images, inventory, dashboard, moderation.
- `orders`: checkout sessions, orders, status history.
- `payments`: payments, Wave approvals, POD captures, refunds, returns.
- `system`: settings, feature flags, audit logs, moderation reports.
- `rbac`: role names per user, Casbin policy rows.

Regenerate stores after query/schema changes:

```sh
make sqlc-generate
```

`sqlc` must be installed locally. The generated Go files are committed so normal builds do not require sqlc.

## Table Reference

Core identity:

- `users`, `user_profiles`, `user_addresses`, `user_devices`, `user_sessions`
- `refresh_tokens`, `email_verification_tokens`, `phone_otp_attempts`, `oauth_identities`, `password_reset_tokens`
- `casbin_rule` (HTTP RBAC policies for Casbin)
- `roles`, `permissions`, `role_permissions`, `user_roles`

Seller and catalog:

- `seller_applications`, `seller_documents`, `sellers`, `seller_bank_accounts`
- `categories`, `brands`, `products`, `product_images`, `product_variants`
- `variant_options`, `variant_option_values`, `product_variant_option_values`
- `inventory_items`, `inventory_movements`, `inventory_reservations`

Commerce:

- `carts`, `cart_items`, `checkout_sessions`, `checkout_session_items`
- `orders`, `order_items`, `order_status_history`
- `payment_methods`, `payments`, `manual_wave_payment_approvals`, `pod_payment_captures`, `payment_webhook_events`
- `refunds`, `return_requests`
- `shipping_providers`, `shipments`, `shipment_tracking_events`, `shipment_labels`

Engagement and operations:

- `vouchers`, `voucher_redemptions`, `product_reviews`, `product_review_images`, `review_votes`, `review_reports`
- `wishlists`, `wishlist_items`, `chat_conversations`, `chat_messages`
- `notification_preferences`, `notifications`, `push_tokens`
- `cms_pages`, `cms_banners`, `cms_faqs`, `cms_announcements`
- `feature_flags`, `platform_settings`, `moderation_reports`, `moderation_cases`
- `seller_wallets`, `seller_wallet_transactions`, `seller_payouts`
- `media_files`, `audit_logs`, `admin_activity_logs`, `idempotency_keys`, `event_logs`
- `search_recent_queries`, `search_click_events`
- `analytics_event_queue`, `admin_metrics_daily`
- `worker_job_runs`, `worker_job_executions`

## Rollback Policy

Production rollback should prefer a new forward migration that restores behavior or data shape. Use `go run ./cmd/migrate down` only before a migration has been promoted beyond a controlled deployment window and only after confirming the down migration is data-safe.

For destructive changes:

1. Add nullable/new columns first.
2. Backfill data in a separate migration or background job.
3. Switch application reads/writes.
4. Drop old columns/tables only after the previous version no longer runs.

Every migration must be validated in CI against a clean PostgreSQL instance.
