# Bitik Web Route Map (Phase 0)

Surfaces split by route group / role. All routes are stubbed in `bitik-web/app/`.

## Surfaces

| Surface | Layout | Audience |
|--------|--------|---------|
| Storefront | `app/(storefront)/layout.tsx` | guest, buyer, seller, admin (browsing) |
| Auth | `app/(auth)/layout.tsx` | guest |
| Buyer account | `app/account/layout.tsx` | authenticated buyer (and any seller/admin) |
| Seller center | `app/seller/layout.tsx` | active or suspended seller, admin/staff |
| Admin console | `app/admin/layout.tsx` | admin / staff with permissions |

## Role access matrix

| Role | Storefront | Auth | Account | Seller apply | Seller center | Admin |
|------|:---:|:---:|:---:|:---:|:---:|:---:|
| guest | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| buyer | ✅ | — | ✅ | ✅ | ❌ | ❌ |
| seller_pending | ✅ | — | ✅ | ✅ (status only) | ❌ | ❌ |
| seller_active | ✅ | — | ✅ | ❌ | ✅ | ❌ |
| seller_suspended | ✅ | — | ✅ | ❌ | ✅ (read-only) | ❌ |
| admin | ✅ | — | ✅ | — | ✅ | ✅ |
| staff | ✅ | — | ✅ | — | optional | ✅ (by permission) |

`AREA_ACCESS` in `bitik-web/lib/roles.ts` is the single source of truth and should be used by route guards.

## Storefront routes

| Path | Page | Backend |
|------|------|---------|
| `/` | Home | `GET /public/home` `GET /public/banners` |
| `/categories` | Categories | `GET /public/categories` |
| `/categories/[id]` | Category detail | `GET /public/categories/{id}` `.../products` |
| `/brands` | Brands | `GET /public/brands` |
| `/brands/[id]` | Brand detail | `GET /public/brands/{id}` `.../products` |
| `/products` | All products | `GET /public/products` |
| `/products/[id]` | Product detail | `GET /public/products/{id}` `.../variants` `.../reviews` |
| `/sellers/[id]` | Seller storefront | `GET /public/sellers/{id}` `.../products` |
| `/search` | Search | `GET /public/search` |
| `/cart` | Cart | `GET /buyer/cart` (+ mutations) |
| `/wishlist` | Wishlist | buyer wishlist endpoints |
| `/checkout` | Checkout | `POST /buyer/checkout/sessions` |
| `/checkout/payment` | Payment | Wave manual + POD endpoints |
| `/checkout/pending` | Wave pending | manual confirmation wait |
| `/checkout/success/[orderId]` | Order placed | `GET /buyer/orders/{id}` |
| `/style-guide` | Design system reference | — |

## Auth routes

| Path | Page | Backend |
|------|------|---------|
| `/login` | Sign in | `POST /auth/login`, OAuth |
| `/register` | Register | `POST /auth/register` |
| `/forgot-password` | Forgot password | forgot/reset password |
| `/reset-password` | Reset password | reset endpoint |
| `/verify-email` | Verify email | email verification |
| `/verify-phone` | Verify phone | phone OTP |

## Buyer account routes

| Path | Page | Backend |
|------|------|---------|
| `/account` | Overview | `GET /users/me` |
| `/account/profile` | Profile | `PATCH /users/me` |
| `/account/addresses` | Addresses | buyer address endpoints |
| `/account/orders` | Orders | `GET /buyer/orders` |
| `/account/orders/[id]` | Order detail | `GET /buyer/orders/{id}` |
| `/account/notifications` | Notifications | `GET /buyer/notifications` |
| `/account/reviews` | Reviews | buyer reviews endpoints |
| `/account/sessions` | Sessions/devices | sessions/devices endpoints |
| `/account/security` | Security | password, 2FA, deletion |
| `/account/preferences` | Preferences | notification preferences |

## Seller center routes

| Path | Page | Backend |
|------|------|---------|
| `/seller` | Dashboard | seller dashboard endpoints |
| `/seller/apply` | Apply as seller | `POST /seller/apply` |
| `/seller/application` | Application status | `GET/PATCH /seller/application` |
| `/seller/products` | Products | `GET /seller/products` |
| `/seller/products/new` | Create product | `POST /seller/products` |
| `/seller/products/[id]` | Edit product | `PATCH /seller/products/{id}` |
| `/seller/inventory` | Inventory | seller inventory endpoints |
| `/seller/orders` | Orders | `GET /seller/orders` |
| `/seller/orders/[id]` | Order detail | seller order endpoints |
| `/seller/shipping` | Shipping | seller shipping endpoints |
| `/seller/promotions` | Promotions | seller promotions endpoints |
| `/seller/reviews` | Reviews | seller reviews endpoints |
| `/seller/wallet` | Wallet/payouts | seller wallet endpoints |
| `/seller/analytics` | Analytics | seller analytics endpoints |
| `/seller/profile` | Shop profile | `PATCH /seller/me/profile` |
| `/seller/settings` | Settings | `PATCH /seller/me/settings` |

## Admin console routes

| Path | Page | Backend |
|------|------|---------|
| `/admin` | Dashboard | `GET /admin/dashboard` |
| `/admin/users` | Users | `GET /admin/users` |
| `/admin/sellers` | Sellers | `GET /admin/sellers` |
| `/admin/products` | Products | `GET /admin/products` |
| `/admin/orders` | Orders | `GET /admin/orders` |
| `/admin/payments` | Payments | admin payment endpoints |
| `/admin/payments/wave` | Wave approvals | manual Wave approve/reject |
| `/admin/shipments` | Shipments | admin shipping endpoints |
| `/admin/promotions` | Promotions | admin promotions endpoints |
| `/admin/moderation` | Moderation | admin moderation endpoints |
| `/admin/cms/pages` | CMS | admin CMS endpoints |
| `/admin/rbac` | RBAC | admin RBAC endpoints |
| `/admin/settings` | Settings | settings + flags endpoints |
| `/admin/audit-logs` | Audit logs | `GET /admin/audit-logs` |
| `/admin/system/health` | System health | system health endpoints |

## Conventions

- Route map constants live in `bitik-web/lib/routes.ts`. Do not hardcode paths in components.
- Each protected surface should consume `canAccess(area, role)` from `lib/roles.ts` via a server-side guard once auth is wired.
- Public storefront prefers SSR/SSG; account/seller/admin default to client rendering with server data fetch boundaries.
