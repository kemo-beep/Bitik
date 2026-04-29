# Order State Machine

Phase 5 uses one parent `orders` row with seller-scoped `order_items` and one shipment per seller. Parent order status is the buyer/admin-facing rollup; seller actions also update that seller's item rows.

## Parent Status Flow

- `pending_payment` -> `paid` after payment authorization/capture or seller acceptance for manual flows.
- `paid` -> `processing` when a seller starts fulfillment.
- `processing` -> `shipped` when shipment tracking is added.
- `shipped` -> `delivered` by shipment tracking integrations or admin updates.
- `delivered` or `shipped` -> `completed` when the buyer confirms receipt.
- `pending_payment`, `paid`, or `processing` -> `cancelled` by buyer, seller, or admin policy.
- `paid`, `processing`, `shipped`, `delivered`, or `disputed` -> `refunded`.
- Any non-`pending_payment` order can move to `disputed`.

Terminal states are `cancelled`, `completed`, and `refunded`.

## Audit Rules

Every status mutation writes `order_status_history` with old status, new status, actor, note, and timestamp. Mutating checkout and order endpoints are covered by the existing idempotency middleware through `Idempotency-Key`.

## Checkout And Inventory

Checkout creation snapshots selected cart items into `checkout_session_items`, reserves inventory with row locks, and records `reserve` movements. Placing an order is transactional: it locks the checkout session, creates the order and order items, attaches reservations to the order, creates seller shipments, records history, and completes the checkout session.

Expired reservations are released by the `release-expired-inventory` worker, which decrements `quantity_reserved`, marks reservations released, and records `release` movements.

## Multi-Seller Rules

The schema uses one parent order with seller-scoped `order_items` and `shipments`. Seller order endpoints are scoped through the authenticated seller id and can only see or mutate their own items and shipment. Parent rollup is conservative: seller shipment moves the parent to `shipped`, buyer confirmation moves it to `completed`, and admin can override via audited status mutation.
