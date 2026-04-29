# Returns (Phase 7)

## Lifecycle

`return_status` enum: `requested → approved|rejected → received` (and later phases can add refund processing).

## Buyer rules

- Buyer creates a return request for a specific `order_item_id`.
- v1 enforcement: returns are available **after delivery** (at least one delivered shipment exists for the order).

## Seller rules

- Seller can approve/reject returns for their items.
- After an approved return, seller can mark it as **received**.\n\nEndpoints are implemented in the orders module:\n- Buyer: `POST /api/v1/buyer/orders/{order_id}/request-return`\n- Seller: `POST /api/v1/seller/orders/{order_id}/return/approve|reject|received`

