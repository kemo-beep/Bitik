-- name: ListActiveVouchers :many
SELECT id, code, title, description, discount_type, discount_value, min_order_cents, max_discount_cents, usage_limit, usage_count, starts_at, ends_at, is_active, created_at, seller_id
FROM vouchers
WHERE is_active = TRUE
  AND starts_at <= now()
  AND ends_at > now()
  AND (usage_limit IS NULL OR usage_count < usage_limit)
  AND (sqlc.narg('seller_id')::uuid IS NULL OR seller_id = sqlc.narg('seller_id')::uuid)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountActiveVouchers :one
SELECT count(*)::bigint
FROM vouchers
WHERE is_active = TRUE
  AND starts_at <= now()
  AND ends_at > now()
  AND (usage_limit IS NULL OR usage_count < usage_limit)
  AND (sqlc.narg('seller_id')::uuid IS NULL OR seller_id = sqlc.narg('seller_id')::uuid);

-- name: ListSellerVouchers :many
SELECT id, code, title, description, discount_type, discount_value, min_order_cents, max_discount_cents, usage_limit, usage_count, starts_at, ends_at, is_active, created_at, seller_id
FROM vouchers
WHERE seller_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSellerVouchers :one
SELECT count(*)::bigint
FROM vouchers
WHERE seller_id = $1;

-- name: ListAllVouchersAdmin :many
SELECT id, code, title, description, discount_type, discount_value, min_order_cents, max_discount_cents, usage_limit, usage_count, starts_at, ends_at, is_active, created_at, seller_id
FROM vouchers
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllVouchersAdmin :one
SELECT count(*)::bigint FROM vouchers;

-- name: CreateVoucher :one
INSERT INTO vouchers (code, title, description, discount_type, discount_value, min_order_cents, max_discount_cents, usage_limit, starts_at, ends_at, is_active, seller_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, code, title, description, discount_type, discount_value, min_order_cents, max_discount_cents, usage_limit, usage_count, starts_at, ends_at, is_active, created_at, seller_id;

-- name: UpdateVoucher :one
UPDATE vouchers
SET title = COALESCE(sqlc.narg('title')::text, title),
    description = COALESCE(sqlc.narg('description')::text, description),
    discount_type = COALESCE(sqlc.narg('discount_type')::text, discount_type),
    discount_value = COALESCE(sqlc.narg('discount_value')::bigint, discount_value),
    min_order_cents = COALESCE(sqlc.narg('min_order_cents')::bigint, min_order_cents),
    max_discount_cents = COALESCE(sqlc.narg('max_discount_cents')::bigint, max_discount_cents),
    usage_limit = COALESCE(sqlc.narg('usage_limit')::int, usage_limit),
    starts_at = COALESCE(sqlc.narg('starts_at')::timestamptz, starts_at),
    ends_at = COALESCE(sqlc.narg('ends_at')::timestamptz, ends_at),
    is_active = COALESCE(sqlc.narg('is_active')::bool, is_active)
WHERE id = sqlc.arg('id')
RETURNING id, code, title, description, discount_type, discount_value, min_order_cents, max_discount_cents, usage_limit, usage_count, starts_at, ends_at, is_active, created_at, seller_id;

-- name: DeleteVoucher :exec
DELETE FROM vouchers WHERE id = $1;

