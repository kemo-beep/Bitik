-- name: ListCategories :many
SELECT id, parent_id, name, slug, image_url, sort_order, is_active, deleted_at, created_at, updated_at
FROM categories
WHERE deleted_at IS NULL
  AND is_active = TRUE
ORDER BY sort_order ASC, name ASC;

-- name: GetCategoryByID :one
SELECT id, parent_id, name, slug, image_url, sort_order, is_active, deleted_at, created_at, updated_at
FROM categories
WHERE id = $1
  AND deleted_at IS NULL
  AND is_active = TRUE;

-- name: ListBrands :many
SELECT id, name, slug, logo_url, is_active, deleted_at, created_at, updated_at
FROM brands
WHERE deleted_at IS NULL
  AND is_active = TRUE
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: GetBrandByID :one
SELECT id, name, slug, logo_url, is_active, deleted_at, created_at, updated_at
FROM brands
WHERE id = $1
  AND deleted_at IS NULL
  AND is_active = TRUE;

-- name: CountPublicProducts :one
SELECT count(*)::bigint
FROM products p
JOIN sellers s ON s.id = p.seller_id
WHERE p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active'
  AND (sqlc.narg('category_id')::uuid IS NULL OR p.category_id = sqlc.narg('category_id')::uuid)
  AND (sqlc.narg('brand_id')::uuid IS NULL OR p.brand_id = sqlc.narg('brand_id')::uuid)
  AND (sqlc.narg('seller_id')::uuid IS NULL OR p.seller_id = sqlc.narg('seller_id')::uuid)
  AND (sqlc.narg('min_price_cents')::bigint IS NULL OR p.min_price_cents >= sqlc.narg('min_price_cents')::bigint)
  AND (sqlc.narg('max_price_cents')::bigint IS NULL OR p.max_price_cents <= sqlc.narg('max_price_cents')::bigint)
  AND (
    sqlc.narg('query')::text IS NULL
    OR sqlc.narg('query')::text = ''
    OR p.search_vector @@ plainto_tsquery('simple', sqlc.narg('query')::text)
    OR p.name ILIKE '%' || sqlc.narg('query')::text || '%'
  );

-- name: ListProducts :many
SELECT p.id, p.seller_id, p.category_id, p.brand_id, p.name, p.slug, p.description, p.status, p.min_price_cents, p.max_price_cents, p.currency, p.total_sold, p.rating, p.review_count, p.published_at, p.deleted_at, p.created_at, p.updated_at,
       pi.url AS primary_image_url
FROM products p
JOIN sellers s ON s.id = p.seller_id
LEFT JOIN LATERAL (
  SELECT url
  FROM product_images
  WHERE product_id = p.id
  ORDER BY is_primary DESC, sort_order ASC, created_at ASC
  LIMIT 1
) pi ON TRUE
WHERE p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active'
  AND (sqlc.narg('category_id')::uuid IS NULL OR p.category_id = sqlc.narg('category_id')::uuid)
  AND (sqlc.narg('brand_id')::uuid IS NULL OR p.brand_id = sqlc.narg('brand_id')::uuid)
  AND (sqlc.narg('seller_id')::uuid IS NULL OR p.seller_id = sqlc.narg('seller_id')::uuid)
  AND (sqlc.narg('min_price_cents')::bigint IS NULL OR p.min_price_cents >= sqlc.narg('min_price_cents')::bigint)
  AND (sqlc.narg('max_price_cents')::bigint IS NULL OR p.max_price_cents <= sqlc.narg('max_price_cents')::bigint)
  AND (
    sqlc.narg('query')::text IS NULL
    OR sqlc.narg('query')::text = ''
    OR p.search_vector @@ plainto_tsquery('simple', sqlc.narg('query')::text)
    OR p.name ILIKE '%' || sqlc.narg('query')::text || '%'
  )
ORDER BY
  CASE WHEN sqlc.arg('sort')::text = 'price_asc' THEN p.min_price_cents END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'price_desc' THEN p.min_price_cents END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'popular' THEN p.total_sold END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'rating' THEN p.rating END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'newest' THEN p.published_at END DESC NULLS LAST,
  p.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetProductByID :one
SELECT p.id, p.seller_id, p.category_id, p.brand_id, p.name, p.slug, p.description, p.status, p.min_price_cents, p.max_price_cents, p.currency, p.total_sold, p.rating, p.review_count, p.published_at, p.deleted_at, p.created_at, p.updated_at
FROM products p
JOIN sellers s ON s.id = p.seller_id
WHERE p.id = $1
  AND p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active';

-- name: GetProductBySlug :one
SELECT p.id, p.seller_id, p.category_id, p.brand_id, p.name, p.slug, p.description, p.status, p.min_price_cents, p.max_price_cents, p.currency, p.total_sold, p.rating, p.review_count, p.published_at, p.deleted_at, p.created_at, p.updated_at
FROM products p
JOIN sellers s ON s.id = p.seller_id
WHERE p.slug = $1
  AND p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active';

-- name: ListProductImages :many
SELECT id, product_id, url, alt_text, sort_order, is_primary, created_at
FROM product_images
WHERE product_id = $1
ORDER BY is_primary DESC, sort_order ASC, created_at ASC;

-- name: ListProductVariants :many
SELECT id, product_id, sku, name, price_cents, compare_at_price_cents, currency, weight_grams, is_active, deleted_at, created_at, updated_at
FROM product_variants
WHERE product_id = $1
  AND deleted_at IS NULL
  AND is_active = TRUE
ORDER BY price_cents ASC, created_at ASC;

-- name: ListProductReviews :many
SELECT pr.id, pr.product_id, pr.order_item_id, pr.user_id, pr.rating, pr.title, pr.body, pr.is_verified_purchase, pr.is_hidden, pr.deleted_at, pr.created_at, pr.updated_at,
       pr.seller_reply, pr.seller_reply_at
FROM product_reviews pr
JOIN products p ON p.id = pr.product_id
JOIN sellers s ON s.id = p.seller_id
WHERE pr.product_id = $1
  AND pr.deleted_at IS NULL
  AND pr.is_hidden = FALSE
  AND p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active'
ORDER BY pr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListRelatedProducts :many
SELECT p.id, p.seller_id, p.category_id, p.brand_id, p.name, p.slug, p.description, p.status, p.min_price_cents, p.max_price_cents, p.currency, p.total_sold, p.rating, p.review_count, p.published_at, p.deleted_at, p.created_at, p.updated_at,
       pi.url AS primary_image_url
FROM products p
JOIN products base ON base.id = sqlc.arg('product_id')
JOIN sellers s ON s.id = p.seller_id
JOIN sellers base_seller ON base_seller.id = base.seller_id
LEFT JOIN LATERAL (
  SELECT url
  FROM product_images
  WHERE product_id = p.id
  ORDER BY is_primary DESC, sort_order ASC, created_at ASC
  LIMIT 1
) pi ON TRUE
WHERE p.deleted_at IS NULL
  AND p.status = 'active'
  AND base.deleted_at IS NULL
  AND base.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active'
  AND base_seller.deleted_at IS NULL
  AND base_seller.status = 'active'
  AND p.id <> base.id
  AND (
    (base.category_id IS NOT NULL AND p.category_id = base.category_id)
    OR (base.brand_id IS NOT NULL AND p.brand_id = base.brand_id)
    OR p.seller_id = base.seller_id
  )
ORDER BY p.rating DESC, p.total_sold DESC, p.created_at DESC
LIMIT sqlc.arg('limit');

-- name: GetPublicSellerByID :one
SELECT id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at
FROM sellers
WHERE id = $1
  AND deleted_at IS NULL
  AND status = 'active';

-- name: GetPublicSellerBySlug :one
SELECT id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at
FROM sellers
WHERE slug = $1
  AND deleted_at IS NULL
  AND status = 'active';

-- name: ListSellerReviews :many
SELECT pr.id, pr.product_id, pr.order_item_id, pr.user_id, pr.rating, pr.title, pr.body, pr.is_verified_purchase, pr.is_hidden, pr.deleted_at, pr.created_at, pr.updated_at,
       pr.seller_reply, pr.seller_reply_at
FROM product_reviews pr
JOIN products p ON p.id = pr.product_id
JOIN sellers s ON s.id = p.seller_id
WHERE p.seller_id = $1
  AND pr.deleted_at IS NULL
  AND pr.is_hidden = FALSE
  AND p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active'
ORDER BY pr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveBanners :many
SELECT id, title, image_url, link_url, placement, sort_order, status, starts_at, ends_at, created_at, updated_at
FROM cms_banners
WHERE status = 'published'
  AND placement = COALESCE(sqlc.narg('placement')::text, placement)
  AND (starts_at IS NULL OR starts_at <= now())
  AND (ends_at IS NULL OR ends_at > now())
ORDER BY sort_order ASC, created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListHomeSections :many
SELECT key, value, description, is_public, created_at, updated_at
FROM platform_settings
WHERE is_public = TRUE
  AND key LIKE 'home.%'
ORDER BY key ASC;

-- name: CreateMediaFile :one
INSERT INTO media_files (owner_user_id, url, bucket, object_key, mime_type, size_bytes, metadata)
VALUES (sqlc.arg('owner_user_id'), sqlc.arg('url'), sqlc.arg('bucket'), sqlc.arg('object_key'), sqlc.arg('mime_type'), sqlc.arg('size_bytes'), COALESCE(sqlc.arg('metadata'), '{}'::jsonb))
RETURNING id, owner_user_id, url, bucket, object_key, mime_type, size_bytes, metadata, created_at;

-- name: ListMediaFiles :many
SELECT id, owner_user_id, url, bucket, object_key, mime_type, size_bytes, metadata, created_at
FROM media_files
WHERE owner_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetMediaFileByID :one
SELECT id, owner_user_id, url, bucket, object_key, mime_type, size_bytes, metadata, created_at
FROM media_files
WHERE id = $1;

-- name: DeleteMediaFile :exec
DELETE FROM media_files
WHERE id = $1
  AND owner_user_id = $2;

-- name: UpdateMediaFileMetadata :one
UPDATE media_files
SET url = $2,
    mime_type = $3,
    size_bytes = $4,
    metadata = $5
WHERE id = $1
  AND owner_user_id = $6
RETURNING id, owner_user_id, url, bucket, object_key, mime_type, size_bytes, metadata, created_at;

-- name: SearchProducts :many
SELECT id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at
FROM products
WHERE deleted_at IS NULL
  AND status = 'active'
  AND (
    $1::text = ''
    OR search_vector @@ plainto_tsquery('simple', $1)
    OR name ILIKE '%' || $1 || '%'
  )
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: TrackSearchQuery :one
INSERT INTO search_recent_queries (user_id, session_id, query, filters)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, session_id, query, filters, created_at;

-- name: TrackSearchClick :one
INSERT INTO search_click_events (user_id, session_id, query, product_id, position, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, session_id, query, product_id, position, metadata, created_at;
