# Bitik Mobile App Production Task Plan

## Goal

Build a production-ready Bitik mobile app for iOS and Android that supports buyer marketplace flows, authenticated account management, cart/checkout, manual Wave payment instructions, pay on delivery, orders, notifications, chat, reviews, and selected seller operations. The mobile app should be reliable on poor networks, secure with device-aware auth, app-store compliant, observable, and ready for future expansion.

## Source Inputs

- Architecture baseline: `docs/rough.md`
- API surface: `api_list.md`
- Database/business domain reference: `db_tables.md`
- Backend task contract: `docs/Backend_tasks.md`
- Web task contract: `docs/Web_frontend_tasks.md`

## Recommended Mobile Architecture

- Framework: React Native with Expo or bare React Native, chosen before implementation.
- Language: TypeScript.
- Navigation: typed stack/tab/deep-link navigation.
- API client: generated from backend OpenAPI where possible.
- Server state: TanStack Query or equivalent.
- Local state: minimal store for UI preferences, cart draft state, onboarding hints, and temporary checkout state.
- Secure storage: platform secure storage/keychain for refresh/session material where applicable.
- Push notifications: FCM/APNs through a provider abstraction.
- Media: native image picker/camera plus presigned upload flow.
- Observability: Sentry, crash reporting, performance traces, analytics events.
- Testing: unit, component, integration, Detox/Maestro or equivalent E2E, release smoke tests.

## Production Acceptance Criteria

- Buyers can register/login, browse, search, view product detail, manage cart, check out, select Wave manual or POD payment, track orders, receive notifications, chat with sellers, and submit reviews.
- Seller users can apply, view application/shop state, manage basic profile, view orders, ship orders, and handle urgent inventory/order tasks if included in v1.
- App handles loading, empty, error, permission-denied, offline, slow-network, and stale-session states.
- App supports secure session refresh, logout, device revocation awareness, and push token registration.
- App supports deep links for products, orders, payment pending pages, seller chat, password reset, email verification, and notifications.
- App uploads media safely using backend presigned upload APIs.
- App is accessible, localized-ready, and responsive across common phone sizes.
- App is ready for app-store release with signing, build profiles, privacy labels, screenshots, and crash monitoring.

## Phase 0: Product Scope, UX, And Native Strategy

### Requirements

- Define v1 mobile scope clearly: buyer-first app with seller companion capabilities, or separate buyer and seller apps.
- Decide whether admin console is web-only for v1. Recommended: keep full admin on web and expose only critical admin approval actions on mobile later if needed.
- Define app platform support, minimum OS versions, release channels, and app-store requirements.

### Tasks

- Decide mobile stack:
  - Expo managed
  - Expo prebuild
  - bare React Native
- Decide app packaging:
  - one app with buyer/seller role switching
  - separate buyer and seller apps
  - buyer app first, seller app later
- Create mobile route map:
  - guest
  - authenticated buyer
  - pending seller
  - active seller
- Define design system aligned with web:
  - tokens
  - typography
  - spacing
  - icons
  - status colors
  - touch target sizes
  - native haptics rules
- Define native permissions:
  - camera
  - photo library
  - push notifications
  - location if shipping/address discovery requires it
- Define offline/poor-network expectations.
- Define app analytics event map and privacy policy inputs.

### Deliverables

- Mobile product scope document.
- Stack decision.
- Route map.
- Native permission matrix.
- Mobile design system baseline.
- App-store readiness checklist.

## Phase 1: Mobile Foundation

### Requirements

- Establish a secure, typed, testable mobile app foundation that can consume backend APIs reliably.

### Tasks

- Scaffold TypeScript mobile app.
- Add linting, formatting, typecheck, test runner, and CI build checks.
- Configure environments:
  - local
  - staging
  - production
  - API base URL
  - asset/CDN URL
  - Sentry DSN
  - push provider config
  - feature flags
- Add app navigation:
  - root auth switch
  - guest tabs
  - buyer tabs
  - seller stack/tab if included
  - modal flows
  - deep-link handling
- Add API layer:
  - generated client from OpenAPI
  - request/response interceptors
  - auth/session headers
  - refresh flow
  - typed error mapping
  - idempotency key support
  - upload helper for presigned URLs
- Add app providers:
  - query cache
  - auth/session
  - feature flags
  - theme
  - toast/snackbar
  - analytics
  - error boundary
- Add secure storage abstraction.
- Add app update/version handling.
- Add crash reporting and performance tracing.

### Deliverables

- Mobile app foundation.
- Typed API client.
- Navigation shell.
- Secure auth storage.
- CI baseline.
- Crash/analytics setup.

## Phase 2: Authentication, Onboarding, And Account

### Requirements

- Mobile users can register, log in, verify contact methods, recover accounts, manage profile, addresses, devices, and notification preferences.

### Tasks

- Build auth screens:
  - welcome
  - register
  - login
  - forgot password
  - reset password
  - email verification
  - phone OTP send/verify
  - OAuth login for Google, Apple, Facebook where platform rules allow
- Add secure session lifecycle:
  - refresh token rotation
  - app foreground refresh
  - logout
  - revoked-session handling
  - banned/deleted account handling
- Build account screens:
  - profile
  - edit profile
  - avatar upload
  - address list
  - address create/edit/delete
  - set default address
  - sessions/devices list
  - revoke device/session
  - notification preferences
  - account deletion
- Add push token registration after consent.
- Add form validation and keyboard-safe layouts.
- Add biometric unlock only if it does not weaken backend session security.

### Deliverables

- Auth and account flows.
- Push token registration.
- Account management screens.
- Auth E2E tests.
- Session security test cases.

## Phase 3: Buyer Home, Catalog, Search, And Product Detail

### Requirements

- Buyers and guests can discover products through home, categories, banners, brands, seller pages, product lists, search, and product detail.

### Tasks

- Build buyer home:
  - banners
  - home sections
  - categories
  - featured products
  - campaigns/flash sales if available
- Build category and brand flows:
  - categories list
  - category detail
  - category products
  - brand list/detail
  - brand products
- Build seller public profile:
  - seller info
  - seller products
  - seller reviews
- Build product listing:
  - filters
  - sort
  - pagination/infinite list
  - pull to refresh
  - empty states
- Build search:
  - search input
  - suggestions
  - trending
  - recent searches
  - clear recent
  - result filters
  - click tracking
- Build product detail:
  - image carousel
  - zoom/gallery view
  - price display
  - variant option selector
  - stock availability
  - seller summary
  - description
  - rating and reviews
  - related products
  - add to cart
  - add to wishlist
  - share/deep link
- Add image caching and placeholder strategy.
- Add skeletons and retry states.

### Deliverables

- Buyer discovery experience.
- Product detail flow.
- Search flow.
- Deep links for product/seller/category pages.
- Catalog performance baseline.

## Phase 4: Wishlist, Cart, Checkout, Payments, And Orders

### Requirements

- Buyers can save products, manage cart, place orders, use Wave manual payment or POD, and track/manage orders.

### Tasks

- Build wishlist:
  - list
  - add/remove product
  - move to cart
- Build cart:
  - cart tab/screen
  - item quantity update
  - remove item
  - clear cart
  - merge guest cart after login
  - select items
  - voucher apply/remove
  - price/stock change warnings
- Build checkout:
  - checkout session creation
  - address selection/edit
  - shipping option selection
  - payment method selection
  - voucher application
  - order summary
  - validation
  - place order with idempotency key
- Build Wave manual payment UX:
  - display exact amount
  - display Bitik Wave business account instructions
  - display order reference/memo
  - copy reference/amount controls
  - pending confirmation screen
  - rejected/expired state
  - support upload/note evidence only if backend includes it
- Build POD UX:
  - eligibility messaging
  - city/order value restrictions
  - confirmation copy
  - delivery payment state
- Build payment methods if backend supports saved methods.
- Build buyer order screens:
  - list with status filters
  - detail
  - items
  - status history/timeline
  - tracking
  - invoice link/share
  - cancel
  - confirm received
  - request refund
  - request return
  - dispute
- Add local optimistic updates only where safe, never for payment/order final states.

### Deliverables

- Complete mobile purchase flow.
- Wave manual and POD payment screens.
- Buyer order management.
- Checkout and payment E2E tests.
- Offline/slow-network checkout behavior tests.

## Phase 5: Reviews, Notifications, Chat, And Real-Time Updates

### Requirements

- Buyers can receive notifications, chat with sellers, and leave/review feedback. Real-time behavior should use WebSocket where available and fallback to polling.

### Tasks

- Build reviews:
  - create review after purchase
  - review list
  - review detail
  - edit/delete review
  - image upload/delete
  - vote/report review
- Build notification center:
  - list
  - unread count
  - mark read
  - read all
  - delete
  - notification preferences
- Add push notification handling:
  - permission prompt
  - token registration
  - foreground notification handling
  - background tap routing
  - deep links to order/product/chat/payment pages
- Build chat:
  - conversations list
  - conversation detail
  - send message
  - image/file attachment
  - read receipts
  - delete message/conversation behavior
- Add WebSocket client:
  - `/ws/chat`
  - `/ws/notifications`
  - `/ws/order-status`
  - reconnect/backoff
  - auth refresh
  - fallback polling
- Add in-app order status update handling.

### Deliverables

- Reviews flow.
- Notification center and push handling.
- Chat experience.
- WebSocket client abstraction.
- Real-time fallback tests.

## Phase 6: Seller Companion Flows

### Requirements

- If seller features are included in the same app, sellers can handle high-value mobile workflows without needing the web seller center.

### Tasks

- Build seller role entry and status handling:
  - apply to sell
  - view application status
  - edit pending application
  - upload seller documents
  - pending/rejected/suspended states
- Build seller profile:
  - shop info
  - logo/banner upload
  - settings basics
- Build seller dashboard:
  - stats
  - recent orders
  - low-stock alerts
  - top products
- Build seller products:
  - product list
  - product detail
  - create/edit simple product
  - image upload
  - publish/unpublish
  - variant and option management if practical on mobile
- Build seller inventory:
  - inventory list/detail
  - quantity adjustment
  - movement history
  - low-stock list
- Build seller orders:
  - orders list/detail
  - accept/reject
  - mark processing
  - pack
  - ship
  - cancel
  - refund/return actions if allowed
- Build seller shipping:
  - shipment list/detail
  - create/update shipment
  - mark shipped
  - tracking
- Build seller reviews:
  - review list/detail
  - reply/report
- Build seller wallet:
  - wallet summary
  - transactions
  - payouts list/detail/request
  - bank accounts basic management

### Deliverables

- Seller mobile companion flows.
- Seller upload and order action tests.
- Role-switching UX.
- Permission and suspended-state tests.

## Phase 7: Native Platform Features

### Requirements

- Mobile app should feel native, handle platform lifecycle events, and protect user data.

### Tasks

- Deep links and universal/app links:
  - product
  - seller
  - order
  - payment pending
  - chat
  - notification
  - reset password
  - email verification
- Push notification categories/actions where useful.
- Native sharing for products and order invoices.
- Camera/photo library for:
  - avatar
  - review images
  - seller documents
  - product images
  - chat attachments
- File/image compression before upload.
- Secure clipboard handling for Wave reference copying.
- App lifecycle handling:
  - refresh data on foreground
  - pause WebSocket in background
  - resume subscriptions
  - session expiry prompt
- Optional location support:
  - address assistance
  - city eligibility for POD
  - only if privacy and product requirements justify it
- Accessibility:
  - labels
  - dynamic text
  - sufficient contrast
  - screen reader order
  - reduced motion handling

### Deliverables

- Deep-link implementation.
- Native upload implementation.
- Push integration.
- Accessibility pass.
- Platform permission documentation.

## Phase 8: Security, Privacy, And Reliability

### Requirements

- Mobile app must protect tokens, PII, order/payment data, and seller/admin actions.

### Tasks

- Store sensitive session material in secure storage.
- Avoid logging PII, tokens, payment references, or addresses.
- Redact Sentry breadcrumbs and network payloads.
- Add certificate pinning only if the operations team can maintain it safely.
- Protect payment/order action screens from accidental duplicate submissions.
- Add idempotency keys for checkout/payment/order mutations.
- Add jailbreak/root detection only if product risk justifies false positives.
- Add remote logout/revoked-device handling.
- Add app version minimum check for forced upgrades.
- Add privacy policy and data deletion links.
- Add local cache invalidation after logout.
- Add secure screen behavior for sensitive pages only if required by policy.

### Deliverables

- Mobile security checklist.
- Privacy checklist.
- Secure storage and logout tests.
- Sensitive logging audit.

## Phase 9: Testing, CI/CD, And Release Operations

### Requirements

- Mobile builds must be reproducible, tested, signed, observable, and releasable to internal, staging, and production channels.

### Tasks

- Unit tests:
  - formatters
  - validators
  - API error mapping
  - auth/session helpers
  - checkout helpers
  - status mapping
- Component tests:
  - forms
  - product cards
  - cart item
  - checkout summary
  - order timeline
  - upload controls
  - seller order actions
- Integration tests:
  - auth flow with mocked API
  - product listing/search
  - cart mutation
  - checkout place-order
  - Wave pending flow
  - push notification routing
  - deep links
- E2E tests:
  - register/login
  - browse/search/product detail
  - add to cart
  - checkout with Wave manual
  - checkout with POD
  - view order and tracking
  - submit review
  - chat with seller
  - seller apply
  - seller ship order
- CI/CD:
  - typecheck
  - lint
  - unit tests
  - build Android
  - build iOS
  - E2E smoke
  - upload internal build
  - promote to staging/production
- Release setup:
  - app identifiers
  - signing certificates
  - provisioning profiles
  - Android keystore
  - store listing
  - screenshots
  - privacy labels
  - release notes
  - phased rollout
  - rollback/hotfix strategy
- Monitoring:
  - crashes
  - API failure rates
  - checkout conversion
  - payment pending duration
  - push delivery/open rates
  - app startup performance

### Deliverables

- Automated test suite.
- Mobile build pipeline.
- Internal/staging/prod release channels.
- App-store submission assets.
- Mobile launch checklist.
- Incident and hotfix runbook.

## MVP Build Order

1. Mobile foundation, navigation, API client, auth/session, secure storage.
2. Buyer home, categories, product listing, search, product detail.
3. Cart, checkout, Wave manual payment, POD payment, buyer orders.
4. Account, addresses, notifications, push registration.
5. Reviews, wishlist, chat, real-time order updates.
6. Seller apply, seller dashboard, seller orders, basic inventory/product actions.
7. Native deep links, uploads, offline hardening, accessibility, app-store readiness.

## MVP Screen Checklist

- [ ] Welcome
- [ ] Register
- [ ] Login
- [ ] Forgot/reset password
- [ ] Email verification
- [ ] Phone OTP
- [ ] Buyer home
- [ ] Categories
- [ ] Product listing
- [ ] Product detail
- [ ] Search
- [ ] Wishlist
- [ ] Cart
- [ ] Checkout address
- [ ] Checkout shipping
- [ ] Checkout payment method
- [ ] Checkout summary
- [ ] Wave manual payment instructions
- [ ] Wave pending confirmation
- [ ] POD confirmation
- [ ] Buyer orders list
- [ ] Buyer order detail
- [ ] Tracking
- [ ] Review create/edit
- [ ] Notifications
- [ ] Chat conversations
- [ ] Chat detail
- [ ] Profile
- [ ] Addresses
- [ ] Notification preferences
- [ ] Sessions/devices
- [ ] Seller application
- [ ] Seller dashboard
- [ ] Seller products
- [ ] Seller inventory
- [ ] Seller orders
- [ ] Seller shipping

## Open Decisions

- Whether to use Expo managed, Expo prebuild, or bare React Native.
- Whether mobile v1 includes seller center features or ships buyer-only first.
- Whether admin actions are web-only for v1.
- Final push notification provider and notification taxonomy.
- Final launch countries, currencies, supported languages, and app-store regions.
- Whether location permission is required or avoidable for v1.
