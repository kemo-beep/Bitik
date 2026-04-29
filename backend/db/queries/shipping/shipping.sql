-- name: ListActiveShippingProviders :many
SELECT id, name, code, metadata, is_active, created_at, updated_at
FROM shipping_providers
WHERE is_active = TRUE
ORDER BY name ASC;

-- name: GetShippingProviderByCode :one
SELECT id, name, code, metadata, is_active, created_at, updated_at
FROM shipping_providers
WHERE code = $1;

-- name: GetShippingProviderByID :one
SELECT id, name, code, metadata, is_active, created_at, updated_at
FROM shipping_providers
WHERE id = $1;

-- name: CreateShippingProvider :one
INSERT INTO shipping_providers (name, code, metadata, is_active)
VALUES (sqlc.arg('name'), sqlc.arg('code'), COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb), COALESCE(sqlc.narg('is_active')::bool, TRUE))
RETURNING id, name, code, metadata, is_active, created_at, updated_at;

-- name: UpdateShippingProvider :one
UPDATE shipping_providers
SET name = COALESCE(sqlc.narg('name')::text, name),
    metadata = COALESCE(sqlc.narg('metadata')::jsonb, metadata),
    is_active = COALESCE(sqlc.narg('is_active')::bool, is_active),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, name, code, metadata, is_active, created_at, updated_at;

-- name: GetShipmentByID :one
SELECT id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at
FROM shipments
WHERE id = $1;

-- name: ListShipmentsForOrder :many
SELECT id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at
FROM shipments
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: ListShipmentsForSellerOrder :many
SELECT s.id, s.order_id, s.seller_id, s.provider_id, s.tracking_number, s.label_url, s.provider_metadata, s.status, s.shipped_at, s.delivered_at, s.created_at, s.updated_at
FROM shipments s
WHERE s.order_id = sqlc.arg('order_id') AND s.seller_id = sqlc.arg('seller_id')
ORDER BY s.created_at ASC;

-- name: ListShipmentsForSeller :many
SELECT id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at
FROM shipments
WHERE seller_id = sqlc.arg('seller_id')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpdateShipmentSellerFields :one
UPDATE shipments
SET provider_id = COALESCE(sqlc.narg('provider_id')::uuid, provider_id),
    tracking_number = COALESCE(sqlc.narg('tracking_number')::text, tracking_number),
    provider_metadata = COALESCE(sqlc.narg('provider_metadata')::jsonb, provider_metadata),
    updated_at = now()
WHERE id = sqlc.arg('id') AND seller_id = sqlc.arg('seller_id')
RETURNING id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at;

-- name: UpdateShipmentStatusForSeller :one
UPDATE shipments
SET status = sqlc.arg('status'),
    shipped_at = CASE WHEN sqlc.arg('status') IN ('shipped', 'in_transit') THEN COALESCE(shipped_at, now()) ELSE shipped_at END,
    delivered_at = CASE WHEN sqlc.arg('status') = 'delivered' THEN COALESCE(delivered_at, now()) ELSE delivered_at END,
    updated_at = now()
WHERE id = sqlc.arg('id') AND seller_id = sqlc.arg('seller_id')
RETURNING id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at;

-- name: UpdateShipmentStatusAdmin :one
UPDATE shipments
SET status = sqlc.arg('status'),
    shipped_at = CASE WHEN sqlc.arg('status') IN ('shipped', 'in_transit') THEN COALESCE(shipped_at, now()) ELSE shipped_at END,
    delivered_at = CASE WHEN sqlc.arg('status') = 'delivered' THEN COALESCE(delivered_at, now()) ELSE delivered_at END,
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, order_id, seller_id, provider_id, tracking_number, label_url, provider_metadata, status, shipped_at, delivered_at, created_at, updated_at;

-- name: CreateShipmentTrackingEvent :one
INSERT INTO shipment_tracking_events (shipment_id, status, location, message, event_time)
SELECT
  sqlc.arg('shipment_id'),
  sqlc.arg('status'),
  sqlc.narg('location'),
  sqlc.narg('message'),
  sqlc.arg('event_time')
WHERE NOT EXISTS (
  SELECT 1
  FROM shipment_tracking_events ste
  WHERE ste.shipment_id = sqlc.arg('shipment_id')
    AND ste.status = sqlc.arg('status')
    AND ste.event_time = sqlc.arg('event_time')
)
RETURNING id, shipment_id, status, location, message, event_time, created_at;

-- name: ListTrackingEventsForShipment :many
SELECT id, shipment_id, status, location, message, event_time, created_at
FROM shipment_tracking_events
WHERE shipment_id = $1
ORDER BY event_time ASC;

-- name: ListTrackingEventsForOrder :many
SELECT ste.id, ste.shipment_id, ste.status, ste.location, ste.message, ste.event_time, ste.created_at
FROM shipment_tracking_events ste
JOIN shipments s ON s.id = ste.shipment_id
WHERE s.order_id = $1
ORDER BY ste.event_time ASC;

-- name: CreateShipmentLabel :one
INSERT INTO shipment_labels (shipment_id, label_url, format, metadata)
VALUES (sqlc.arg('shipment_id'), sqlc.arg('label_url'), COALESCE(sqlc.narg('format')::text, 'pdf'), COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb))
RETURNING id, shipment_id, label_url, format, generated_at, metadata;

-- name: ListShipmentLabels :many
SELECT id, shipment_id, label_url, format, generated_at, metadata
FROM shipment_labels
WHERE shipment_id = $1
ORDER BY generated_at DESC;

-- name: CountUndeliveredShipmentsForOrder :one
SELECT COUNT(1)::bigint
FROM shipments
WHERE order_id = $1 AND status <> 'delivered';

-- name: ListDeliveredShipmentsWithoutTrackingEvents :many
SELECT s.id, s.order_id, s.seller_id, s.provider_id, s.tracking_number, s.label_url, s.provider_metadata, s.status, s.shipped_at, s.delivered_at, s.created_at, s.updated_at
FROM shipments s
WHERE s.status = 'delivered'
  AND NOT EXISTS (
    SELECT 1 FROM shipment_tracking_events ste
    WHERE ste.shipment_id = s.id
      AND ste.status = 'delivered'
  )
ORDER BY s.delivered_at ASC NULLS LAST
LIMIT sqlc.arg('limit');
