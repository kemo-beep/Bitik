-- name: ListUserAddresses :many
SELECT id, user_id, full_name, phone, country, state, city, district, postal_code, address_line1, address_line2, latitude, longitude, is_default, deleted_at, created_at, updated_at
FROM user_addresses
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY is_default DESC, created_at DESC;

-- name: GetUserAddress :one
SELECT id, user_id, full_name, phone, country, state, city, district, postal_code, address_line1, address_line2, latitude, longitude, is_default, deleted_at, created_at, updated_at
FROM user_addresses
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetSellerByUserID :one
SELECT id, user_id, application_id, shop_name, slug, description, logo_url, banner_url, status, rating, total_sales, deleted_at, created_at, updated_at, settings
FROM sellers
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: CreateUserAddress :one
INSERT INTO user_addresses (user_id, full_name, phone, country, state, city, district, postal_code, address_line1, address_line2, is_default)
VALUES (sqlc.arg('user_id'), sqlc.arg('full_name'), sqlc.arg('phone'), sqlc.arg('country'), sqlc.narg('state'), sqlc.narg('city'), sqlc.narg('district'), sqlc.narg('postal_code'), sqlc.arg('address_line1'), sqlc.narg('address_line2'), sqlc.arg('is_default'))
RETURNING id, user_id, full_name, phone, country, state, city, district, postal_code, address_line1, address_line2, latitude, longitude, is_default, deleted_at, created_at, updated_at;

-- name: UpdateUserAddress :one
UPDATE user_addresses
SET full_name = COALESCE(sqlc.narg('full_name')::text, full_name),
    phone = COALESCE(sqlc.narg('phone')::text, phone),
    country = COALESCE(sqlc.narg('country')::text, country),
    state = COALESCE(sqlc.narg('state')::text, state),
    city = COALESCE(sqlc.narg('city')::text, city),
    district = COALESCE(sqlc.narg('district')::text, district),
    postal_code = COALESCE(sqlc.narg('postal_code')::text, postal_code),
    address_line1 = COALESCE(sqlc.narg('address_line1')::text, address_line1),
    address_line2 = COALESCE(sqlc.narg('address_line2')::text, address_line2)
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id') AND deleted_at IS NULL
RETURNING id, user_id, full_name, phone, country, state, city, district, postal_code, address_line1, address_line2, latitude, longitude, is_default, deleted_at, created_at, updated_at;

-- name: ClearDefaultAddress :exec
UPDATE user_addresses SET is_default = FALSE
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: SetDefaultAddress :one
UPDATE user_addresses SET is_default = TRUE
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, user_id, full_name, phone, country, state, city, district, postal_code, address_line1, address_line2, latitude, longitude, is_default, deleted_at, created_at, updated_at;

-- name: DeleteUserAddress :exec
UPDATE user_addresses SET deleted_at = now(), is_default = FALSE
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetOrCreateCart :one
INSERT INTO carts (user_id, session_id)
VALUES (sqlc.arg('user_id'), sqlc.narg('session_id'))
ON CONFLICT (user_id) DO UPDATE SET updated_at = now()
RETURNING id, user_id, session_id, created_at, updated_at, voucher_id;

-- name: GetCartForUser :one
SELECT id, user_id, session_id, created_at, updated_at, voucher_id
FROM carts
WHERE user_id = $1;

-- name: GetActiveVariantForCart :one
SELECT p.id AS product_id, p.seller_id, p.name AS product_name, p.slug AS product_slug,
       pv.id AS variant_id, pv.sku, pv.name AS variant_name, pv.price_cents, pv.currency,
       img.url AS image_url,
       COALESCE(inv.quantity_available - inv.quantity_reserved, 0)::int AS purchasable_quantity
FROM product_variants pv
JOIN products p ON p.id = pv.product_id
JOIN sellers s ON s.id = p.seller_id
LEFT JOIN inventory_items inv ON inv.variant_id = pv.id
LEFT JOIN LATERAL (
  SELECT url FROM product_images
  WHERE product_id = p.id
  ORDER BY is_primary DESC, sort_order ASC, created_at ASC
  LIMIT 1
) img ON TRUE
WHERE pv.id = $1
  AND pv.deleted_at IS NULL
  AND pv.is_active = TRUE
  AND p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active';

-- name: GetCartItemForUser :one
SELECT ci.id, ci.cart_id, ci.product_id, ci.variant_id, ci.quantity, ci.created_at, ci.updated_at, ci.selected
FROM cart_items ci
JOIN carts c ON c.id = ci.cart_id
WHERE ci.id = $1 AND c.user_id = $2;

-- name: ListCartItemsDetailed :many
SELECT ci.id, ci.cart_id, ci.product_id, ci.variant_id, ci.quantity, ci.selected, ci.created_at, ci.updated_at,
       p.name AS product_name, p.slug AS product_slug, p.seller_id, pv.sku, pv.name AS variant_name,
       pv.price_cents, pv.currency, img.url AS image_url,
       COALESCE(inv.quantity_available - inv.quantity_reserved, 0)::int AS purchasable_quantity
FROM cart_items ci
JOIN products p ON p.id = ci.product_id
JOIN sellers s ON s.id = p.seller_id
JOIN product_variants pv ON pv.id = ci.variant_id
LEFT JOIN inventory_items inv ON inv.variant_id = pv.id
LEFT JOIN LATERAL (
  SELECT url FROM product_images
  WHERE product_id = p.id
  ORDER BY is_primary DESC, sort_order ASC, created_at ASC
  LIMIT 1
) img ON TRUE
WHERE ci.cart_id = $1
  AND p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active'
  AND pv.deleted_at IS NULL
  AND pv.is_active = TRUE
ORDER BY ci.created_at DESC;

-- name: UpsertCartItem :one
INSERT INTO cart_items (cart_id, product_id, variant_id, quantity, selected)
VALUES (sqlc.arg('cart_id'), sqlc.arg('product_id'), sqlc.arg('variant_id'), sqlc.arg('quantity'), TRUE)
ON CONFLICT (cart_id, variant_id) DO UPDATE
SET quantity = cart_items.quantity + EXCLUDED.quantity,
    selected = TRUE
RETURNING id, cart_id, product_id, variant_id, quantity, created_at, updated_at, selected;

-- name: UpdateCartItemQuantity :one
UPDATE cart_items ci
SET quantity = sqlc.arg('quantity')
FROM carts c
WHERE ci.cart_id = c.id
  AND ci.id = sqlc.arg('id')
  AND c.user_id = sqlc.arg('user_id')
RETURNING ci.id, ci.cart_id, ci.product_id, ci.variant_id, ci.quantity, ci.created_at, ci.updated_at, ci.selected;

-- name: DeleteCartItem :exec
DELETE FROM cart_items ci
USING carts c
WHERE ci.cart_id = c.id
  AND ci.id = $1
  AND c.user_id = $2;

-- name: ClearCart :exec
DELETE FROM cart_items
WHERE cart_id = $1;

-- name: SetAllCartItemsSelected :exec
UPDATE cart_items SET selected = $2
WHERE cart_id = $1;

-- name: SelectCartItems :exec
UPDATE cart_items SET selected = TRUE
WHERE cart_id = $1 AND id = ANY($2::uuid[]);

-- name: ApplyCartVoucher :one
UPDATE carts SET voucher_id = $2
WHERE id = $1 AND user_id = $3
RETURNING id, user_id, session_id, created_at, updated_at, voucher_id;

-- name: RemoveCartVoucher :one
UPDATE carts SET voucher_id = NULL
WHERE id = $1 AND user_id = $2 AND voucher_id = $3
RETURNING id, user_id, session_id, created_at, updated_at, voucher_id;

-- name: GetActiveVoucherByCode :one
SELECT id, code, title, description, discount_type, discount_value, min_order_cents, max_discount_cents, usage_limit, usage_count, starts_at, ends_at, is_active, created_at, seller_id
FROM vouchers
WHERE upper(code) = upper($1)
  AND is_active = TRUE
  AND starts_at <= now()
  AND ends_at > now()
  AND (usage_limit IS NULL OR usage_count < usage_limit);

-- name: GetVoucherByID :one
SELECT id, code, title, description, discount_type, discount_value, min_order_cents, max_discount_cents, usage_limit, usage_count, starts_at, ends_at, is_active, created_at, seller_id
FROM vouchers
WHERE id = $1
  AND is_active = TRUE
  AND starts_at <= now()
  AND ends_at > now()
  AND (usage_limit IS NULL OR usage_count < usage_limit);

-- name: CreateCheckoutSession :one
INSERT INTO checkout_sessions (user_id, cart_id, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, voucher_id)
VALUES (sqlc.arg('user_id'), sqlc.arg('cart_id'), COALESCE(sqlc.narg('currency')::text, 'USD'), sqlc.arg('subtotal_cents'), sqlc.arg('discount_cents'), sqlc.arg('shipping_cents'), sqlc.arg('tax_cents'), sqlc.arg('total_cents'), sqlc.narg('shipping_address'), sqlc.narg('billing_address'), sqlc.narg('payment_method'), sqlc.narg('selected_shipping_option'), sqlc.arg('expires_at'), sqlc.narg('voucher_id'))
RETURNING id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id;

-- name: GetCheckoutSession :one
SELECT id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id
FROM checkout_sessions
WHERE id = $1;

-- name: GetCheckoutSessionForUser :one
SELECT id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id
FROM checkout_sessions
WHERE id = $1 AND user_id = $2;

-- name: LockCheckoutSessionForUser :one
SELECT id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id
FROM checkout_sessions
WHERE id = $1 AND user_id = $2
FOR UPDATE;

-- name: CreateCheckoutSessionItem :one
INSERT INTO checkout_session_items (checkout_session_id, seller_id, product_id, variant_id, quantity, unit_price_cents, total_price_cents, currency, metadata)
VALUES (sqlc.arg('checkout_session_id'), sqlc.arg('seller_id'), sqlc.arg('product_id'), sqlc.arg('variant_id'), sqlc.arg('quantity'), sqlc.arg('unit_price_cents'), sqlc.arg('total_price_cents'), sqlc.arg('currency'), sqlc.arg('metadata'))
RETURNING id, checkout_session_id, seller_id, product_id, variant_id, quantity, unit_price_cents, total_price_cents, currency, metadata, created_at;

-- name: ListCheckoutSessionItems :many
SELECT id, checkout_session_id, seller_id, product_id, variant_id, quantity, unit_price_cents, total_price_cents, currency, metadata, created_at
FROM checkout_session_items
WHERE checkout_session_id = $1
ORDER BY created_at ASC;

-- name: UpdateCheckoutAddress :one
UPDATE checkout_sessions
SET shipping_address = COALESCE(sqlc.narg('shipping_address')::jsonb, shipping_address),
    billing_address = COALESCE(sqlc.narg('billing_address')::jsonb, billing_address)
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id') AND status = 'open'
RETURNING id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id;

-- name: UpdateCheckoutShipping :one
UPDATE checkout_sessions
SET selected_shipping_option = sqlc.arg('selected_shipping_option'),
    shipping_cents = sqlc.arg('shipping_cents'),
    total_cents = GREATEST(subtotal_cents - discount_cents + sqlc.arg('shipping_cents') + tax_cents, 0)
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id') AND status = 'open'
RETURNING id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id;

-- name: UpdateCheckoutPaymentMethod :one
UPDATE checkout_sessions
SET payment_method = sqlc.arg('payment_method')
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id') AND status = 'open'
RETURNING id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id;

-- name: ApplyCheckoutVoucher :one
UPDATE checkout_sessions
SET voucher_id = sqlc.arg('voucher_id'),
    discount_cents = sqlc.arg('discount_cents'),
    total_cents = GREATEST(subtotal_cents - sqlc.arg('discount_cents') + shipping_cents + tax_cents, 0)
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id') AND status = 'open'
RETURNING id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id;

-- name: RemoveCheckoutVoucher :one
UPDATE checkout_sessions
SET voucher_id = NULL,
    discount_cents = 0,
    total_cents = subtotal_cents + shipping_cents + tax_cents
WHERE id = $1 AND user_id = $2 AND voucher_id = $3 AND status = 'open'
RETURNING id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id;

-- name: CompleteCheckoutSession :one
UPDATE checkout_sessions
SET status = 'completed', completed_at = now()
WHERE id = $1 AND user_id = $2 AND status = 'open'
RETURNING id, user_id, cart_id, status, currency, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, shipping_address, billing_address, payment_method, selected_shipping_option, expires_at, completed_at, created_at, updated_at, voucher_id;

-- name: LockInventoryByVariant :one
SELECT id, product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold, created_at, updated_at
FROM inventory_items
WHERE variant_id = $1
FOR UPDATE;

-- name: LockInventoryByID :one
SELECT id, product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold, created_at, updated_at
FROM inventory_items
WHERE id = $1
FOR UPDATE;

-- name: UpdateInventoryReserved :one
UPDATE inventory_items
SET quantity_reserved = quantity_reserved + sqlc.arg('delta_reserved')
WHERE id = sqlc.arg('id')
  AND quantity_reserved + sqlc.arg('delta_reserved') >= 0
  AND quantity_reserved + sqlc.arg('delta_reserved') <= quantity_available
RETURNING id, product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold, created_at, updated_at;

-- name: ConsumeInventoryReservation :one
UPDATE inventory_items
SET quantity_available = quantity_available - sqlc.arg('quantity'),
    quantity_reserved = quantity_reserved - sqlc.arg('quantity')
WHERE id = sqlc.arg('id')
  AND quantity_available - sqlc.arg('quantity') >= 0
  AND quantity_reserved - sqlc.arg('quantity') >= 0
  AND quantity_reserved - sqlc.arg('quantity') <= quantity_available - sqlc.arg('quantity')
RETURNING id, product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold, created_at, updated_at;

-- name: CreateInventoryReservation :one
INSERT INTO inventory_reservations (inventory_item_id, user_id, checkout_session_id, quantity, expires_at)
VALUES (sqlc.arg('inventory_item_id'), sqlc.arg('user_id'), sqlc.arg('checkout_session_id'), sqlc.arg('quantity'), sqlc.arg('expires_at'))
RETURNING id, inventory_item_id, user_id, checkout_session_id, order_id, quantity, expires_at, released_at, created_at;

-- name: ListReservationsForCheckout :many
SELECT id, inventory_item_id, user_id, checkout_session_id, order_id, quantity, expires_at, released_at, created_at
FROM inventory_reservations
WHERE checkout_session_id = $1 AND released_at IS NULL
ORDER BY created_at ASC;

-- name: ListReservationsForOrder :many
SELECT id, inventory_item_id, user_id, checkout_session_id, order_id, quantity, expires_at, released_at, created_at
FROM inventory_reservations
WHERE order_id = $1 AND released_at IS NULL
ORDER BY created_at ASC;

-- name: AttachReservationToOrder :one
UPDATE inventory_reservations
SET order_id = $2
WHERE id = $1 AND released_at IS NULL
RETURNING id, inventory_item_id, user_id, checkout_session_id, order_id, quantity, expires_at, released_at, created_at;

-- name: ListExpiredReservations :many
SELECT id, inventory_item_id, user_id, checkout_session_id, order_id, quantity, expires_at, released_at, created_at
FROM inventory_reservations
WHERE released_at IS NULL AND order_id IS NULL AND expires_at <= now()
ORDER BY expires_at ASC
LIMIT $1;

-- name: ListStaleUnpaidOrders :many
SELECT id, order_number, user_id, checkout_session_id, status, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, currency, shipping_address, billing_address, placed_at, paid_at, cancelled_at, completed_at, created_at, updated_at
FROM orders
WHERE status = 'pending_payment'
  AND cancelled_at IS NULL
  AND paid_at IS NULL
  AND placed_at <= now() - (sqlc.arg('older_than_minutes')::int || ' minutes')::interval
ORDER BY placed_at ASC
LIMIT sqlc.arg('limit');

-- name: MarkReservationReleased :one
UPDATE inventory_reservations
SET released_at = now()
WHERE id = $1 AND released_at IS NULL
RETURNING id, inventory_item_id, user_id, checkout_session_id, order_id, quantity, expires_at, released_at, created_at;

-- name: CreateInventoryMovement :one
INSERT INTO inventory_movements (inventory_item_id, movement_type, quantity, reason, reference_type, reference_id, actor_user_id, before_available, after_available, before_reserved, after_reserved)
VALUES (sqlc.arg('inventory_item_id'), sqlc.arg('movement_type'), sqlc.arg('quantity'), sqlc.narg('reason'), sqlc.narg('reference_type'), sqlc.narg('reference_id'), sqlc.narg('actor_user_id'), sqlc.narg('before_available'), sqlc.narg('after_available'), sqlc.narg('before_reserved'), sqlc.narg('after_reserved'))
RETURNING id, inventory_item_id, movement_type, quantity, reason, reference_type, reference_id, created_at, actor_user_id, before_available, after_available, before_reserved, after_reserved;

-- name: CreateOrder :one
INSERT INTO orders (order_number, user_id, checkout_session_id, status, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, currency, shipping_address, billing_address)
VALUES (sqlc.arg('order_number'), sqlc.arg('user_id'), sqlc.arg('checkout_session_id'), COALESCE(sqlc.narg('status')::order_status, 'pending_payment'), sqlc.arg('subtotal_cents'), sqlc.arg('discount_cents'), sqlc.arg('shipping_cents'), sqlc.arg('tax_cents'), sqlc.arg('total_cents'), COALESCE(sqlc.narg('currency')::text, 'USD'), sqlc.arg('shipping_address'), sqlc.narg('billing_address'))
RETURNING id, order_number, user_id, checkout_session_id, status, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, currency, shipping_address, billing_address, placed_at, paid_at, cancelled_at, completed_at, created_at, updated_at;

-- name: CreateOrderItemFromCheckoutItem :one
INSERT INTO order_items (order_id, seller_id, product_id, variant_id, product_name, variant_name, sku, image_url, quantity, unit_price_cents, total_price_cents, currency, status)
SELECT sqlc.arg('order_id'), csi.seller_id, csi.product_id, csi.variant_id,
       COALESCE(csi.metadata->>'product_name', 'Product') AS product_name,
       csi.metadata->>'variant_name' AS variant_name,
       csi.metadata->>'sku' AS sku,
       csi.metadata->>'image_url' AS image_url,
       csi.quantity, csi.unit_price_cents, csi.total_price_cents, csi.currency, 'pending_payment'::order_status
FROM checkout_session_items csi
WHERE csi.id = sqlc.arg('checkout_item_id')
RETURNING id, order_id, seller_id, product_id, variant_id, product_name, variant_name, sku, image_url, quantity, unit_price_cents, total_price_cents, currency, status, created_at;

-- name: CreateShipmentsForOrderSellers :many
INSERT INTO shipments (order_id, seller_id)
SELECT DISTINCT oi.order_id, oi.seller_id
FROM order_items oi
WHERE oi.order_id = $1
RETURNING id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at;

-- name: IncrementVoucherUsage :exec
UPDATE vouchers SET usage_count = usage_count + 1 WHERE id = $1;

-- name: CreateVoucherRedemption :one
INSERT INTO voucher_redemptions (voucher_id, user_id, order_id, discount_cents)
VALUES ($1, $2, $3, $4)
RETURNING id, voucher_id, user_id, order_id, discount_cents, redeemed_at;

-- name: GetOrderByID :one
SELECT id, order_number, user_id, checkout_session_id, status, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, currency, shipping_address, billing_address, placed_at, paid_at, cancelled_at, completed_at, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: GetOrderForUser :one
SELECT id, order_number, user_id, checkout_session_id, status, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, currency, shipping_address, billing_address, placed_at, paid_at, cancelled_at, completed_at, created_at, updated_at
FROM orders
WHERE id = $1 AND user_id = $2;

-- name: ListUserOrders :many
SELECT id, order_number, user_id, checkout_session_id, status, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, currency, shipping_address, billing_address, placed_at, paid_at, cancelled_at, completed_at, created_at, updated_at
FROM orders
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('status')::order_status IS NULL OR status = sqlc.narg('status')::order_status)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListOrderItems :many
SELECT id, order_id, seller_id, product_id, variant_id, product_name, variant_name, sku, image_url, quantity, unit_price_cents, total_price_cents, currency, status, created_at
FROM order_items
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: ListOrderCreditsBySeller :many
SELECT seller_id, SUM(total_price_cents)::bigint AS amount_cents
FROM order_items
WHERE order_id = $1
GROUP BY seller_id
ORDER BY seller_id;

-- name: ListOrderStatusHistory :many
SELECT id, order_id, old_status, new_status, note, created_by, created_at
FROM order_status_history
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = sqlc.arg('status'),
    paid_at = CASE WHEN sqlc.arg('status') = 'paid' THEN COALESCE(paid_at, now()) ELSE paid_at END,
    cancelled_at = CASE WHEN sqlc.arg('status') = 'cancelled' THEN COALESCE(cancelled_at, now()) ELSE cancelled_at END,
    completed_at = CASE WHEN sqlc.arg('status') = 'completed' THEN COALESCE(completed_at, now()) ELSE completed_at END
WHERE id = sqlc.arg('id')
RETURNING id, order_number, user_id, checkout_session_id, status, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, currency, shipping_address, billing_address, placed_at, paid_at, cancelled_at, completed_at, created_at, updated_at;

-- name: InsertOrderStatusHistory :one
INSERT INTO order_status_history (order_id, old_status, new_status, note, created_by)
VALUES (sqlc.arg('order_id'), sqlc.narg('old_status'), sqlc.arg('new_status'), sqlc.narg('note'), sqlc.narg('created_by'))
RETURNING id, order_id, old_status, new_status, note, created_by, created_at;

-- name: ListShipmentsForOrder :many
SELECT id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at
FROM shipments
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: ListPaymentsForOrder :many
SELECT id, order_id, payment_method_id, provider, provider_payment_id, status, amount_cents, currency, idempotency_key, metadata, paid_at, failed_at, created_at, updated_at
FROM payments
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: ListTrackingEventsForOrder :many
SELECT ste.id, ste.shipment_id, ste.status, ste.location, ste.message, ste.event_time, ste.created_at
FROM shipment_tracking_events ste
JOIN shipments s ON s.id = ste.shipment_id
WHERE s.order_id = $1
ORDER BY ste.event_time ASC;

-- name: GetOrderIDForOrderItemForUser :one
SELECT oi.order_id
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE oi.id = $1 AND o.user_id = $2;

-- name: GetOrderAndSellerForOrderItemForUser :one
SELECT oi.order_id, oi.seller_id
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE oi.id = $1 AND o.user_id = $2;

-- name: CountDeliveredShipmentsForOrderSeller :one
SELECT COUNT(1)::bigint
FROM shipments
WHERE order_id = $1 AND seller_id = $2 AND status = 'delivered';

-- name: CreateRefundRequest :one
INSERT INTO refunds (order_id, requested_by, reason, amount_cents, currency, metadata)
SELECT o.id, sqlc.arg('requested_by'), sqlc.narg('reason'), o.total_cents, o.currency, COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb)
FROM orders o
WHERE o.id = sqlc.arg('order_id') AND o.user_id = sqlc.arg('requested_by')
RETURNING id, order_id, payment_id, requested_by, status, reason, amount_cents, currency, reviewed_by, reviewed_at, processed_at, metadata, created_at, updated_at;

-- name: CreateReturnRequest :one
INSERT INTO return_requests (order_id, order_item_id, requested_by, reason, quantity, metadata)
SELECT oi.order_id, oi.id, sqlc.arg('requested_by'), sqlc.narg('reason'), LEAST(sqlc.arg('quantity')::int, oi.quantity), COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb)
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE oi.id = sqlc.arg('order_item_id') AND o.user_id = sqlc.arg('requested_by')
RETURNING id, order_id, order_item_id, requested_by, status, reason, quantity, reviewed_by, reviewed_at, received_at, metadata, created_at, updated_at;

-- name: ListSellerOrders :many
SELECT DISTINCT o.id, o.order_number, o.user_id, o.checkout_session_id, o.status, o.subtotal_cents, o.discount_cents, o.shipping_cents, o.tax_cents, o.total_cents, o.currency, o.shipping_address, o.billing_address, o.placed_at, o.paid_at, o.cancelled_at, o.completed_at, o.created_at, o.updated_at
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
WHERE oi.seller_id = sqlc.arg('seller_id')
  AND (sqlc.narg('status')::order_status IS NULL OR o.status = sqlc.narg('status')::order_status)
ORDER BY o.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetSellerOrder :one
SELECT DISTINCT o.id, o.order_number, o.user_id, o.checkout_session_id, o.status, o.subtotal_cents, o.discount_cents, o.shipping_cents, o.tax_cents, o.total_cents, o.currency, o.shipping_address, o.billing_address, o.placed_at, o.paid_at, o.cancelled_at, o.completed_at, o.created_at, o.updated_at
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
WHERE o.id = $1 AND oi.seller_id = $2;

-- name: ListSellerOrderItems :many
SELECT id, order_id, seller_id, product_id, variant_id, product_name, variant_name, sku, image_url, quantity, unit_price_cents, total_price_cents, currency, status, created_at
FROM order_items
WHERE order_id = $1 AND seller_id = $2
ORDER BY created_at ASC;

-- name: UpdateSellerOrderItemsStatus :exec
UPDATE order_items SET status = $3
WHERE order_id = $1 AND seller_id = $2;

-- name: UpdateShipmentForSeller :one
UPDATE shipments
SET status = sqlc.arg('status'),
    tracking_number = COALESCE(sqlc.narg('tracking_number')::text, tracking_number),
    shipped_at = CASE WHEN sqlc.arg('status') IN ('shipped', 'in_transit') THEN COALESCE(shipped_at, now()) ELSE shipped_at END,
    delivered_at = CASE WHEN sqlc.arg('status') = 'delivered' THEN COALESCE(delivered_at, now()) ELSE delivered_at END
WHERE order_id = sqlc.arg('order_id') AND seller_id = sqlc.arg('seller_id')
RETURNING id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at;

-- name: ReviewReturnForSeller :one
UPDATE return_requests rr
SET status = sqlc.arg('status'),
    reviewed_by = sqlc.arg('reviewed_by'),
    reviewed_at = now()
FROM order_items oi
WHERE rr.order_item_id = oi.id
  AND rr.order_id = sqlc.arg('order_id')
  AND oi.seller_id = sqlc.arg('seller_id')
RETURNING rr.id, rr.order_id, rr.order_item_id, rr.requested_by, rr.status, rr.reason, rr.quantity, rr.reviewed_by, rr.reviewed_at, rr.received_at, rr.metadata, rr.created_at, rr.updated_at;

-- name: MarkReturnReceivedForSeller :one
UPDATE return_requests rr
SET status = 'received',
    received_at = COALESCE(received_at, now()),
    updated_at = now()
FROM order_items oi
WHERE rr.order_item_id = oi.id
  AND rr.order_id = sqlc.arg('order_id')
  AND oi.seller_id = sqlc.arg('seller_id')
  AND rr.status = 'approved'
RETURNING rr.id, rr.order_id, rr.order_item_id, rr.requested_by, rr.status, rr.reason, rr.quantity, rr.reviewed_by, rr.reviewed_at, rr.received_at, rr.metadata, rr.created_at, rr.updated_at;

-- name: ListAdminOrders :many
SELECT id, order_number, user_id, checkout_session_id, status, subtotal_cents, discount_cents, shipping_cents, tax_cents, total_cents, currency, shipping_address, billing_address, placed_at, paid_at, cancelled_at, completed_at, created_at, updated_at
FROM orders
WHERE (sqlc.narg('status')::order_status IS NULL OR status = sqlc.narg('status')::order_status)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreateOrGetOrderInvoice :one
INSERT INTO order_invoices (order_id, invoice_number, payload)
VALUES (sqlc.arg('order_id'), sqlc.arg('invoice_number'), sqlc.arg('payload'))
ON CONFLICT (order_id) DO UPDATE SET payload = order_invoices.payload
RETURNING id, order_id, invoice_number, status, payload, generated_at, created_at;

-- name: GetOrderInvoiceForUser :one
SELECT oi.id, oi.order_id, oi.invoice_number, oi.status, oi.payload, oi.generated_at, oi.created_at
FROM order_invoices oi
JOIN orders o ON o.id = oi.order_id
WHERE oi.order_id = $1 AND o.user_id = $2;
