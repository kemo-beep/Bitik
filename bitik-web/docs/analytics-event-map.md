# Analytics Event Map (Phase 8)

## Backend Ingestion Endpoint
- `POST /api/v1/analytics/events` (authenticated marketplace user)
- Handler: `backend/internal/adminsvc/phase9_handlers.go` via `HandleIngestAnalyticsEvent`
- Route registration: `backend/internal/adminsvc/service.go`

## Event Catalog

| Event name | Payload schema | Trigger location | Owner |
| --- | --- | --- | --- |
| `product.view` | `{ product_id, seller_id }` | `app/(storefront)/products/product-detail-client.tsx` | Web storefront |
| `search.submit` | `{ q }` | `app/(storefront)/search/search-client.tsx` | Web storefront |
| `search.click` | `{ q, index }` | `app/(storefront)/search/search-client.tsx` | Web storefront |
| `checkout.start` | `{}` | `app/(storefront)/checkout/checkout-client.tsx` | Web checkout |
| `checkout.place_order` | `{ checkout_session_id, order_id }` | `app/(storefront)/checkout/checkout-client.tsx` | Web checkout |
| `payment.method_selected` | `{ payment_method }` | `app/(storefront)/checkout/checkout-client.tsx` | Web checkout |
| `seller.product_create` | `{}` | `app/seller/products/seller-products-client.tsx` | Seller center |
| `seller.product_publish` | `{ product_id }` | `app/seller/products/seller-products-client.tsx` | Seller center |
| `admin.moderation_action` | `{ action, product_id }` | `app/admin/products/page.tsx` | Admin console |

## Transport Layer
- Frontend tracker: `lib/analytics.tsx`
- Event constants: `lib/analytics-events.ts`
- Network transport: authenticated `bitikFetch` to `/analytics/events`

## Enabling client instrumentation
- Tracking runs only when `NEXT_PUBLIC_ANALYTICS_ENABLED` is truthy (`1` / `true` / etc.); see `lib/env.ts`.
- Set it at **Next build** time for environments that should emit events (including CI builds used for Playwright).
- Local Playwright `webServer` in `playwright.config.ts` sets `NEXT_PUBLIC_ANALYTICS_ENABLED=1` for the dev server.
- GitHub Actions `web-ci.yml` sets the same variable for Playwright smoke and `@critical` runs.

