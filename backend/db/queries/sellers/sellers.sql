-- name: GetSellerApplicationByUserID :one
SELECT id, user_id, shop_name, slug, business_type, country, currency, status, submitted_at, reviewed_by, reviewed_at, rejection_reason, metadata, created_at, updated_at
FROM seller_applications
WHERE user_id = $1
  AND status <> 'cancelled'
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateSellerApplication :one
INSERT INTO seller_applications (user_id, shop_name, slug, business_type, country, currency, status, metadata)
VALUES ($1, $2, $3, $4, $5, COALESCE($6, 'USD'), 'draft', COALESCE($7, '{}'::jsonb))
RETURNING id, user_id, shop_name, slug, business_type, country, currency, status, submitted_at, reviewed_by, reviewed_at, rejection_reason, metadata, created_at, updated_at;

-- name: UpdateSellerApplication :one
UPDATE seller_applications
SET shop_name = COALESCE(sqlc.narg('shop_name')::text, shop_name),
    slug = COALESCE(sqlc.narg('slug')::text, slug),
    business_type = COALESCE(sqlc.narg('business_type')::text, business_type),
    country = COALESCE(sqlc.narg('country')::text, country),
    currency = COALESCE(sqlc.narg('currency')::text, currency),
    metadata = COALESCE(sqlc.narg('metadata')::jsonb, metadata)
WHERE id = sqlc.arg('id')
  AND user_id = sqlc.arg('user_id')
  AND status IN ('draft', 'rejected')
RETURNING id, user_id, shop_name, slug, business_type, country, currency, status, submitted_at, reviewed_by, reviewed_at, rejection_reason, metadata, created_at, updated_at;

-- name: SubmitSellerApplication :one
UPDATE seller_applications
SET status = 'submitted',
    submitted_at = now()
WHERE id = $1
  AND user_id = $2
  AND status IN ('draft', 'rejected')
RETURNING id, user_id, shop_name, slug, business_type, country, currency, status, submitted_at, reviewed_by, reviewed_at, rejection_reason, metadata, created_at, updated_at;

-- name: ListSellerApplicationsForAdmin :many
SELECT id, user_id, shop_name, slug, business_type, country, currency, status, submitted_at, reviewed_by, reviewed_at, rejection_reason, metadata, created_at, updated_at
FROM seller_applications
WHERE status = COALESCE(sqlc.narg('status')::seller_application_status, status)
ORDER BY submitted_at DESC NULLS LAST, created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ReviewSellerApplication :one
UPDATE seller_applications
SET status = sqlc.arg('status')::seller_application_status,
    reviewed_by = sqlc.arg('reviewed_by'),
    reviewed_at = now(),
    rejection_reason = sqlc.narg('rejection_reason')::text
WHERE id = sqlc.arg('id')
  AND status IN ('submitted', 'in_review')
RETURNING id, user_id, shop_name, slug, business_type, country, currency, status, submitted_at, reviewed_by, reviewed_at, rejection_reason, metadata, created_at, updated_at;

-- name: CreateSellerFromApplication :one
INSERT INTO sellers (user_id, application_id, shop_name, slug, status)
SELECT sa.user_id, sa.id, sa.shop_name, sa.slug, 'active'
FROM seller_applications sa
WHERE sa.id = $1
  AND sa.status = 'approved'
ON CONFLICT (user_id) DO UPDATE
SET application_id = EXCLUDED.application_id,
    shop_name = EXCLUDED.shop_name,
    slug = EXCLUDED.slug,
    status = 'active',
    deleted_at = NULL
RETURNING id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at, settings;

-- name: AssignSellerRole :exec
INSERT INTO user_roles (user_id, role_id)
SELECT $1, id FROM roles WHERE name = 'seller'
ON CONFLICT DO NOTHING;

-- name: CreateSellerDocument :one
INSERT INTO seller_documents (seller_application_id, seller_id, document_type, file_url, metadata)
VALUES (sqlc.arg('seller_application_id'), sqlc.arg('seller_id'), sqlc.arg('document_type'), sqlc.arg('file_url'), COALESCE(sqlc.arg('metadata'), '{}'::jsonb))
RETURNING id, seller_application_id, seller_id, document_type, file_url, status, reviewed_by, reviewed_at, rejection_reason, metadata, created_at;

-- name: ListSellerDocuments :many
SELECT id, seller_application_id, seller_id, document_type, file_url, status, reviewed_by, reviewed_at, rejection_reason, metadata, created_at
FROM seller_documents
WHERE (seller_application_id = sqlc.narg('seller_application_id')::uuid OR seller_id = sqlc.narg('seller_id')::uuid)
ORDER BY created_at DESC;

-- name: DeleteSellerDocument :exec
DELETE FROM seller_documents
WHERE seller_documents.id = $1
  AND (
    seller_documents.seller_application_id IN (SELECT sa.id FROM seller_applications sa WHERE sa.user_id = $2)
    OR seller_documents.seller_id IN (SELECT s.id FROM sellers s WHERE s.user_id = $2)
  );

-- name: GetSellerByUserID :one
SELECT id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at, settings
FROM sellers
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: GetSellerByID :one
SELECT id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at, settings
FROM sellers
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateSellerProfile :one
UPDATE sellers
SET shop_name = COALESCE(sqlc.narg('shop_name')::text, shop_name),
    slug = COALESCE(sqlc.narg('slug')::text, slug),
    description = COALESCE(sqlc.narg('description')::text, description)
WHERE id = sqlc.arg('id')
  AND user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL
RETURNING id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at, settings;

-- name: UpdateSellerSettings :one
UPDATE sellers
SET settings = COALESCE(sqlc.arg('settings')::jsonb, '{}'::jsonb)
WHERE id = sqlc.arg('id')
  AND user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL
RETURNING id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at, settings;

-- name: UpdateSellerMedia :one
UPDATE sellers
SET logo_url = COALESCE(sqlc.narg('logo_url')::text, logo_url),
    banner_url = COALESCE(sqlc.narg('banner_url')::text, banner_url)
WHERE id = sqlc.arg('id')
  AND user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL
RETURNING id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at, settings;

-- name: SuspendSellerForAdmin :one
UPDATE sellers
SET status = sqlc.arg('status')::seller_status
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at, settings;

-- name: CountSellerProducts :one
SELECT count(*)::bigint
FROM products
WHERE seller_id = $1
  AND deleted_at IS NULL;

-- name: ListSellerProducts :many
SELECT id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at, moderation_status, moderation_reason, moderated_by, moderated_at
FROM products
WHERE seller_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetSellerProduct :one
SELECT id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at, moderation_status, moderation_reason, moderated_by, moderated_at
FROM products
WHERE id = $1
  AND seller_id = $2
  AND deleted_at IS NULL;

-- name: CreateSellerProduct :one
INSERT INTO products (seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, moderation_status)
VALUES (sqlc.arg('seller_id'), sqlc.arg('category_id'), sqlc.arg('brand_id'), sqlc.arg('name'), sqlc.arg('slug'), sqlc.arg('description'), COALESCE(sqlc.arg('status'), 'draft'), sqlc.arg('min_price_cents'), sqlc.arg('max_price_cents'), COALESCE(sqlc.arg('currency'), 'USD'), 'approved')
RETURNING id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at, moderation_status, moderation_reason, moderated_by, moderated_at;

-- name: UpdateSellerProduct :one
UPDATE products
SET category_id = COALESCE(sqlc.narg('category_id')::uuid, category_id),
    brand_id = COALESCE(sqlc.narg('brand_id')::uuid, brand_id),
    name = COALESCE(sqlc.narg('name')::text, name),
    slug = COALESCE(sqlc.narg('slug')::text, slug),
    description = COALESCE(sqlc.narg('description')::text, description),
    min_price_cents = COALESCE(sqlc.narg('min_price_cents')::bigint, min_price_cents),
    max_price_cents = COALESCE(sqlc.narg('max_price_cents')::bigint, max_price_cents),
    currency = COALESCE(sqlc.narg('currency')::text, currency),
    moderation_status = 'pending'
WHERE id = sqlc.arg('id')
  AND seller_id = sqlc.arg('seller_id')
  AND deleted_at IS NULL
  AND status <> 'deleted'
RETURNING id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at, moderation_status, moderation_reason, moderated_by, moderated_at;

-- name: SoftDeleteSellerProduct :exec
UPDATE products
SET status = 'deleted',
    deleted_at = now()
WHERE id = $1
  AND seller_id = $2
  AND deleted_at IS NULL;

-- name: PublishSellerProduct :one
UPDATE products
SET status = 'active',
    published_at = COALESCE(published_at, now())
WHERE id = $1
  AND seller_id = $2
  AND deleted_at IS NULL
  AND status IN ('draft', 'inactive')
  AND moderation_status = 'approved'
RETURNING id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at, moderation_status, moderation_reason, moderated_by, moderated_at;

-- name: UnpublishSellerProduct :one
UPDATE products
SET status = 'inactive'
WHERE id = $1
  AND seller_id = $2
  AND deleted_at IS NULL
  AND status = 'active'
RETURNING id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at, moderation_status, moderation_reason, moderated_by, moderated_at;

-- name: DuplicateSellerProduct :one
INSERT INTO products (seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, moderation_status)
SELECT p.seller_id, p.category_id, p.brand_id, sqlc.arg('name'), sqlc.arg('slug'), p.description, 'draft', p.min_price_cents, p.max_price_cents, p.currency, 'approved'
FROM products p
WHERE p.id = sqlc.arg('source_product_id')
  AND p.seller_id = sqlc.arg('seller_id')
  AND p.deleted_at IS NULL
RETURNING id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at, moderation_status, moderation_reason, moderated_by, moderated_at;

-- name: ModerateProductForAdmin :one
UPDATE products
SET moderation_status = sqlc.arg('moderation_status')::text,
    moderation_reason = sqlc.narg('moderation_reason')::text,
    moderated_by = sqlc.arg('moderated_by'),
    moderated_at = now(),
    status = CASE WHEN sqlc.arg('moderation_status')::text = 'rejected' THEN 'inactive'::product_status ELSE status END
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at, deleted_at, created_at, updated_at, moderation_status, moderation_reason, moderated_by, moderated_at;

-- name: CreateProductImage :one
INSERT INTO product_images (product_id, url, alt_text, sort_order, is_primary)
VALUES ($1, $2, $3, $4, COALESCE($5, FALSE))
RETURNING id, product_id, url, alt_text, sort_order, is_primary, created_at;

-- name: ListProductImagesForSeller :many
SELECT pi.id, pi.product_id, pi.url, pi.alt_text, pi.sort_order, pi.is_primary, pi.created_at
FROM product_images pi
JOIN products p ON p.id = pi.product_id
WHERE pi.product_id = $1
  AND p.seller_id = $2
ORDER BY pi.sort_order ASC, pi.created_at ASC;

-- name: UpdateProductImageOrder :one
UPDATE product_images pi
SET sort_order = $3,
    is_primary = $4
FROM products p
WHERE pi.id = $1
  AND pi.product_id = p.id
  AND p.seller_id = $2
RETURNING pi.id, pi.product_id, pi.url, pi.alt_text, pi.sort_order, pi.is_primary, pi.created_at;

-- name: DeleteProductImage :exec
DELETE FROM product_images pi
USING products p
WHERE pi.id = $1
  AND pi.product_id = p.id
  AND p.seller_id = $2;

-- name: CreateProductVariant :one
INSERT INTO product_variants (product_id, sku, name, price_cents, compare_at_price_cents, currency, weight_grams, is_active)
VALUES ($1, $2, $3, $4, $5, COALESCE($6, 'USD'), $7, COALESCE($8, TRUE))
RETURNING id, product_id, sku, name, price_cents, compare_at_price_cents, currency, weight_grams, is_active, deleted_at, created_at, updated_at;

-- name: ListProductVariantsForSeller :many
SELECT pv.id, pv.product_id, pv.sku, pv.name, pv.price_cents, pv.compare_at_price_cents, pv.currency, pv.weight_grams, pv.is_active, pv.deleted_at, pv.created_at, pv.updated_at
FROM product_variants pv
JOIN products p ON p.id = pv.product_id
WHERE pv.product_id = $1
  AND p.seller_id = $2
  AND pv.deleted_at IS NULL
ORDER BY pv.created_at ASC;

-- name: UpdateProductVariant :one
UPDATE product_variants pv
SET sku = COALESCE(sqlc.narg('sku')::text, sku),
    name = COALESCE(sqlc.narg('name')::text, name),
    price_cents = COALESCE(sqlc.narg('price_cents')::bigint, price_cents),
    compare_at_price_cents = sqlc.narg('compare_at_price_cents')::bigint,
    currency = COALESCE(sqlc.narg('currency')::text, currency),
    weight_grams = sqlc.narg('weight_grams')::int,
    is_active = COALESCE(sqlc.narg('is_active')::boolean, is_active)
FROM products p
WHERE pv.id = sqlc.arg('id')
  AND pv.product_id = p.id
  AND p.seller_id = sqlc.arg('seller_id')
  AND pv.deleted_at IS NULL
RETURNING pv.id, pv.product_id, pv.sku, pv.name, pv.price_cents, pv.compare_at_price_cents, pv.currency, pv.weight_grams, pv.is_active, pv.deleted_at, pv.created_at, pv.updated_at;

-- name: DeleteProductVariant :exec
UPDATE product_variants pv
SET deleted_at = now(),
    is_active = FALSE
FROM products p
WHERE pv.id = $1
  AND pv.product_id = p.id
  AND p.seller_id = $2
  AND pv.deleted_at IS NULL;

-- name: CreateVariantOption :one
INSERT INTO variant_options (product_id, name, sort_order)
VALUES ($1, $2, COALESCE($3, 0))
RETURNING id, product_id, name, sort_order;

-- name: CreateVariantOptionValue :one
INSERT INTO variant_option_values (option_id, value, sort_order)
SELECT vo.id, sqlc.arg('value')::text, COALESCE(sqlc.arg('sort_order')::int, 0)
FROM variant_options vo
JOIN products p ON p.id = vo.product_id
WHERE vo.id = sqlc.arg('option_id')
  AND p.seller_id = sqlc.arg('seller_id')
RETURNING id, option_id, value, sort_order;

-- name: ListVariantOptionsForSeller :many
SELECT vo.id, vo.product_id, vo.name, vo.sort_order
FROM variant_options vo
JOIN products p ON p.id = vo.product_id
WHERE vo.product_id = $1
  AND p.seller_id = $2
ORDER BY vo.sort_order ASC, vo.name ASC;

-- name: ListVariantOptionValues :many
SELECT id, option_id, value, sort_order
FROM variant_option_values
WHERE option_id = $1
ORDER BY sort_order ASC, value ASC;

-- name: DeleteVariantOption :exec
DELETE FROM variant_options vo
USING products p
WHERE vo.id = $1
  AND vo.product_id = p.id
  AND p.seller_id = $2;

-- name: UpsertInventoryItem :one
INSERT INTO inventory_items (product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold)
VALUES ($1, $2, $3, 0, COALESCE($4, 5))
ON CONFLICT (variant_id) DO UPDATE
SET quantity_available = EXCLUDED.quantity_available,
    low_stock_threshold = EXCLUDED.low_stock_threshold
RETURNING id, product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold, created_at, updated_at;

-- name: ListInventoryForSeller :many
SELECT ii.id, ii.product_id, ii.variant_id, ii.quantity_available, ii.quantity_reserved, ii.low_stock_threshold, ii.created_at, ii.updated_at
FROM inventory_items ii
JOIN products p ON p.id = ii.product_id
WHERE p.seller_id = $1
ORDER BY ii.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: ListLowStockForSeller :many
SELECT ii.id, ii.product_id, ii.variant_id, ii.quantity_available, ii.quantity_reserved, ii.low_stock_threshold, ii.created_at, ii.updated_at
FROM inventory_items ii
JOIN products p ON p.id = ii.product_id
WHERE p.seller_id = $1
  AND ii.quantity_available <= ii.low_stock_threshold
ORDER BY ii.quantity_available ASC, ii.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: GetInventoryItemForSeller :one
SELECT ii.id, ii.product_id, ii.variant_id, ii.quantity_available, ii.quantity_reserved, ii.low_stock_threshold, ii.created_at, ii.updated_at
FROM inventory_items ii
JOIN products p ON p.id = ii.product_id
WHERE ii.id = $1
  AND p.seller_id = $2;

-- name: LockInventoryItemForSeller :one
SELECT ii.id, ii.product_id, ii.variant_id, ii.quantity_available, ii.quantity_reserved, ii.low_stock_threshold, ii.created_at, ii.updated_at
FROM inventory_items ii
JOIN products p ON p.id = ii.product_id
WHERE ii.id = $1
  AND p.seller_id = $2
FOR UPDATE;

-- name: UpdateInventoryQuantities :one
UPDATE inventory_items
SET quantity_available = $2,
    quantity_reserved = $3,
    low_stock_threshold = COALESCE($4, low_stock_threshold)
WHERE id = $1
RETURNING id, product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold, created_at, updated_at;

-- name: CreateInventoryMovement :one
INSERT INTO inventory_movements (inventory_item_id, movement_type, quantity, reason, reference_type, reference_id, actor_user_id, before_available, after_available, before_reserved, after_reserved)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, inventory_item_id, movement_type, quantity, reason, reference_type, reference_id, created_at, actor_user_id, before_available, after_available, before_reserved, after_reserved;

-- name: ListInventoryMovementsForSeller :many
SELECT im.id, im.inventory_item_id, im.movement_type, im.quantity, im.reason, im.reference_type, im.reference_id, im.created_at, im.actor_user_id, im.before_available, im.after_available, im.before_reserved, im.after_reserved
FROM inventory_movements im
JOIN inventory_items ii ON ii.id = im.inventory_item_id
JOIN products p ON p.id = ii.product_id
WHERE im.inventory_item_id = $1
  AND p.seller_id = $2
ORDER BY im.created_at DESC
LIMIT $3 OFFSET $4;

-- name: SellerDashboardStats :one
SELECT
  (SELECT count(*)::bigint FROM products p WHERE p.seller_id = $1 AND p.deleted_at IS NULL) AS product_count,
  (SELECT count(*)::bigint FROM products p WHERE p.seller_id = $1 AND p.status = 'active' AND p.deleted_at IS NULL) AS active_product_count,
  (SELECT count(*)::bigint FROM inventory_items ii JOIN products p ON p.id = ii.product_id WHERE p.seller_id = $1 AND ii.quantity_available <= ii.low_stock_threshold) AS low_stock_count,
  (SELECT COALESCE(sum(oi.total_price_cents), 0)::bigint FROM order_items oi WHERE oi.seller_id = $1) AS gross_sales_cents,
  (SELECT count(DISTINCT oi.order_id)::bigint FROM order_items oi WHERE oi.seller_id = $1) AS order_count;

-- name: SellerSalesChart :many
SELECT date_trunc('day', o.placed_at)::date AS day,
       COALESCE(sum(oi.total_price_cents), 0)::bigint AS sales_cents,
       count(DISTINCT oi.order_id)::bigint AS order_count
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE oi.seller_id = $1
  AND o.placed_at >= now() - (sqlc.arg('days')::int * interval '1 day')
GROUP BY day
ORDER BY day ASC;

-- name: SellerTopProducts :many
SELECT p.id, p.seller_id, p.category_id, p.brand_id, p.name, p.slug, p.description, p.status, p.min_price_cents, p.max_price_cents, p.currency, p.total_sold, p.rating, p.review_count, p.published_at, p.deleted_at, p.created_at, p.updated_at, p.moderation_status, p.moderation_reason, p.moderated_by, p.moderated_at
FROM products p
WHERE p.seller_id = $1
  AND p.deleted_at IS NULL
ORDER BY p.total_sold DESC, p.rating DESC, p.created_at DESC
LIMIT $2;

-- name: SellerRecentOrders :many
SELECT o.id, o.order_number, o.user_id, o.checkout_session_id, o.status, o.subtotal_cents, o.discount_cents, o.shipping_cents, o.tax_cents, o.total_cents, o.currency, o.shipping_address, o.billing_address, o.placed_at, o.paid_at, o.cancelled_at, o.completed_at, o.created_at, o.updated_at
FROM orders o
WHERE EXISTS (SELECT 1 FROM order_items oi WHERE oi.order_id = o.id AND oi.seller_id = $1)
ORDER BY o.placed_at DESC
LIMIT $2;
