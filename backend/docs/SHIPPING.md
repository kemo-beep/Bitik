# Shipping (Phase 7)

## Data model

- `shipping_providers`: carrier/provider catalog (e.g. `manual`, `local-courier`).
- `shipments`: per-order, per-seller shipment rows created at order placement (Phase 5).
- `shipment_tracking_events`: timeline events per shipment.
- `shipment_labels`: optional labels (Phase 7 supports a stub for `local-courier`).

## Shipment statuses

`shipment_status` enum includes: `pending`, `packed`, `shipped`, `in_transit`, `delivered`, `failed`, `returned`.

Seller transitions are enforced as a simple monotonic flow:\n- `pending → packed → shipped → in_transit → delivered`\n- `pending → shipped` is allowed.\n- No transitions out of `delivered`.\n+
Admin can override status for operational fixes.

## Provider adapters (v1)

- **manual**: no label support; seller enters tracking/status manually.\n- **local-courier**: label generation is a stub; tracking URL templates can be stored in `shipping_providers.metadata.tracking_url_template`.

## Worker: tracking ingestion

`POST /api/v1/internal/jobs/update-shipment-tracking`\n\nv1 behavior: ensures any `delivered` shipment has at least one `delivered` tracking event (idempotent by `(shipment_id,status,event_time)`).

