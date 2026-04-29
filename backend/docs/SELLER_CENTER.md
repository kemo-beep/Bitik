# Seller Center

Phase 4 adds `internal/sellersvc`, a dedicated seller-center module for onboarding, seller profile management, dashboard data, product management, variants/options/images, inventory, and admin moderation.

## Onboarding

Authenticated users can create and submit seller applications with:

- `POST /api/v1/seller/apply`
- `GET /api/v1/seller/application`
- `PATCH /api/v1/seller/application`
- `POST /api/v1/seller/documents`
- `DELETE /api/v1/seller/documents/{document_id}`

Admins can review applications with `POST /api/v1/admin/seller-applications/{application_id}/review`. Approving an application creates or reactivates the seller account and assigns the `seller` role.

## Seller Profile

Seller/admin scoped endpoints are protected by JWT, active-account checks, role checks, and seller ownership checks:

- `GET/PATCH /api/v1/seller/me`
- `PATCH /api/v1/seller/me/profile`
- `PATCH /api/v1/seller/me/settings`
- `PATCH /api/v1/seller/me/media`

Seller settings are stored as JSONB on `sellers.settings` for flexible future expansion.

## Products

Seller products support list, create, read, update, soft delete, publish, unpublish, duplicate, images, variants, and options.

Product transition rules:

- Created products start as `draft` with `moderation_status = approved`.
- Updating product catalog fields moves `moderation_status` to `pending`.
- Publishing is allowed from `draft` or `inactive` only when moderation is `approved`.
- Rejected products are forced to `inactive` by admin moderation.

## Inventory

Inventory updates use row-level locks through `SELECT ... FOR UPDATE` before writing quantities. Every update creates an `inventory_movements` row with actor, before/after available quantity, before/after reserved quantity, reason, and reference metadata.

Seller inventory endpoints include list, detail, set/adjust, movement history, low-stock, and bulk update.

## Admin Moderation

Admins can:

- List and review seller applications.
- Update seller status.
- Moderate products with `approved`, `pending`, or `rejected`.

The Phase 4 support migration adds seller settings, product moderation fields, inventory movement audit columns, and seller-center indexes.
