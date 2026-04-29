# Bitik Web Frontend Production Task Plan

## Goal

Build production-ready web applications for Bitik: a buyer storefront, seller center, and admin console that consume the backend APIs, support marketplace workflows end to end, and are secure, fast, accessible, observable, and maintainable.

## Source Inputs

- Architecture baseline: `docs/rough.md`
- API surface: `api_list.md`
- Database/business domain reference: `db_tables.md`
- Backend task contract: `docs/Backend_tasks.md`

## Recommended Web Architecture

- Framework: Next.js with TypeScript.
- Rendering: SSR/SSG for public storefront pages; client rendering where dashboards and authenticated workflows need it.
- Styling: Tailwind CSS or a comparable design-system-friendly approach.
- Data fetching: generated API client from OpenAPI plus TanStack Query or equivalent for caching/mutations.
- Forms: React Hook Form plus schema validation.
- State: server state through query cache; minimal client state through context/store only where needed.
- Auth: backend-issued access/refresh strategy agreed with backend; support secure cookie or bearer-token mode based on final decision.
- Media: direct/presigned upload support for R2/MinIO.
- Observability: Sentry, web vitals, structured frontend event logging.
- Testing: unit, component, accessibility, API contract, and Playwright E2E.

## Production Acceptance Criteria

- Buyer storefront supports browsing, search, product detail, cart, checkout, Wave manual instructions, POD checkout, order tracking, reviews, notifications, and account management.
- Seller center supports onboarding, shop profile, products, variants, inventory, orders, shipping, promotions, reviews, wallet/payouts, and analytics.
- Admin console supports users, sellers, products, orders, payments, shipping, promotions, moderation, CMS, RBAC, settings, audit logs, and system health.
- All user-facing forms validate client-side and handle backend errors consistently.
- UI handles loading, empty, error, offline, and permission-denied states.
- Web app is responsive across mobile, tablet, and desktop.
- Accessibility targets WCAG 2.1 AA for core flows.
- Critical pages meet production performance targets: fast LCP on storefront, minimal dashboard blocking, image optimization, and route-level code splitting.
- E2E tests cover critical buyer, seller, and admin journeys.
- Deployment is automated with environment-specific configuration.

## Phase 0: Product UX, Information Architecture, And Design System

### Requirements

- Define the web product surfaces and navigation:
  - buyer storefront
  - authenticated buyer account
  - seller center
  - admin console
- Align page inventory with backend API modules.
- Establish consistent UI, content, error handling, and empty-state patterns.

### Tasks

- Create sitemap and route map for buyer, seller, and admin areas.
- Define user roles and page access:
  - guest buyer
  - authenticated buyer
  - pending seller
  - active seller
  - suspended seller
  - admin
  - staff roles by permission
- Create design tokens:
  - colors
  - typography
  - spacing
  - radius
  - shadows
  - breakpoints
  - z-index rules
- Build shared components:
  - buttons
  - inputs
  - selects
  - checkboxes/radios
  - modals
  - drawers
  - tabs
  - breadcrumbs
  - pagination
  - tables
  - cards
  - badges
  - toasts
  - file upload
  - image gallery
  - money/date/status formatters
  - loading skeletons
  - empty/error states
- Define marketplace status UI:
  - order statuses
  - payment statuses
  - shipment statuses
  - seller statuses
  - product statuses
  - refund/return statuses
- Create responsive layout shells:
  - storefront header/footer
  - buyer account layout
  - seller dashboard layout
  - admin dashboard layout
- Add accessibility baseline:
  - keyboard navigation
  - focus rings
  - semantic landmarks
  - aria patterns for menus/dialogs
  - color contrast checks

### Deliverables

- Route map.
- Design system documentation.
- Shared component library.
- Status and formatter utilities.
- Accessibility checklist.

## Phase 1: Web App Foundation

### Requirements

- Establish production project setup, API integration, auth handling, environment configuration, and CI.

### Tasks

- Scaffold web app with TypeScript, linting, formatting, test runner, and build checks.
- Add environment configuration:
  - API base URL
  - asset/CDN base URL
  - Sentry DSN
  - feature flags
  - OAuth redirect URLs
- Generate typed API client from backend OpenAPI.
- Create API wrapper with:
  - request IDs
  - auth token/cookie handling
  - refresh retry behavior
  - typed error mapping
  - idempotency key support for checkout/payment mutations
  - upload support for presigned URLs
- Add route protection:
  - guest-only routes
  - buyer routes
  - seller routes
  - admin permission routes
- Add global app providers:
  - query cache
  - auth/session
  - theme
  - toasts
  - analytics
  - feature flags
- Add Sentry and web-vitals reporting.
- Add Storybook or component preview if desired.
- Add CI:
  - install
  - typecheck
  - lint
  - unit tests
  - component tests
  - production build
  - Playwright smoke tests

### Deliverables

- Web foundation repository/app.
- Typed API client.
- Auth/session provider.
- Protected route system.
- CI pipeline.

## Phase 2: Authentication, Account, And Buyer Profile

### Requirements

- Users can register, log in, log out, refresh sessions, recover passwords, verify email/phone, manage profile, addresses, sessions, and devices.

### Notes / Gaps

- **Notification preferences persistence**: UI can be implemented, but it requires a backend API (and OpenAPI spec) to store/read preferences. Until then, the web implementation should treat it as a UI-only stub.

### Tasks

- Build auth pages:
  - register
  - login
  - forgot password
  - reset password
  - verify email
  - phone OTP send/verify
  - OAuth callbacks for Google, Apple, Facebook
- Add session persistence and refresh handling.
- Add account pages:
  - profile
  - edit profile
  - addresses list/create/edit/delete/set-default
  - sessions/devices list and revoke
  - notification preferences
  - account deletion request/confirmation
- Add form-level validation and backend error rendering.
- Add rate-limit and retry UI for OTP/password flows.
- Add audit-sensitive UX for logout-all/devices and account deletion.

### Deliverables

- Complete auth/account web flows.
- Buyer account settings area.
- Auth E2E tests.
- Accessibility tests for forms and dialogs.

## Phase 3: Public Storefront And Discovery

### Requirements

- Guests and buyers can browse the marketplace, discover products, search, filter, view sellers, and inspect product details.

### Tasks

- Build storefront pages:
  - home
  - home sections
  - banners
  - category listing
  - category detail
  - brand listing/detail
  - seller public profile
  - seller product listing
  - product listing
  - product detail
  - related products
  - reviews listing
- Build search experience:
  - search page
  - query input
  - suggestions
  - trending searches
  - recent searches
  - filter panel
  - sort controls
  - pagination/infinite loading decision
  - click tracking
- Product detail requirements:
  - image gallery
  - title/description
  - price range and variant price
  - option/variant selector
  - stock display
  - seller card
  - rating/review summary
  - shipping/payment hints
  - add to cart
  - add to wishlist
  - related products
- Add SEO:
  - metadata
  - canonical URLs
  - Open Graph/Twitter cards
  - product structured data
  - sitemap
  - robots configuration
- Add performance work:
  - optimized images
  - route prefetch rules
  - cache strategy for public catalog data
  - skeleton loading states

### Deliverables

- Buyer storefront.
- Search and discovery UI.
- SEO implementation.
- Storefront E2E tests.
- Web performance baseline report.

## Phase 4: Cart, Wishlist, Checkout, Payments, And Orders

### Requirements

- Buyers can manage wishlist/cart, check out, select shipping and payment, view Wave payment instructions, use POD where eligible, track orders, request returns/refunds, and submit reviews.

### Tasks

- Build wishlist:
  - list items
  - add/remove product
  - move to cart
- Build cart:
  - cart page/drawer
  - item quantity updates
  - remove item
  - clear cart
  - merge guest cart after login
  - select items
  - apply/remove voucher
  - inventory and price-change warnings
- Build checkout:
  - create checkout session
  - address selection/edit
  - shipping option selection
  - payment method selection
  - voucher application
  - order summary
  - validation step
  - place order with idempotency key
- Build payment UX:
  - POD eligibility display and policy text
  - Wave manual payment instructions
  - exact amount and order reference
  - pending manual confirmation page
  - payment retry/cancel states
  - timeout or rejected payment state
- Build buyer orders:
  - order list with filters
  - order detail
  - order items
  - status history
  - tracking
  - invoice link
  - cancel order
  - confirm received
  - request refund/return
  - dispute
- Build reviews:
  - create review after eligible purchase
  - edit/delete review
  - image upload/delete
  - vote/report review
- Build notifications:
  - list
  - unread count
  - mark read/read all
  - delete
  - push token registration support if browser push is used

### Deliverables

- Complete buyer purchase flow.
- Wave manual and POD payment UI.
- Buyer order management.
- Checkout E2E tests.
- Payment edge-case tests.

## Phase 5: Seller Center

### Requirements

- Sellers can onboard, manage shop profile, create products, upload media, manage variants/options/inventory, handle orders/shipments, run promotions, reply to reviews, manage wallet/payouts, and view analytics.

### Tasks

- Build seller onboarding:
  - application form
  - application status
  - edit pending application
  - document upload/delete
- Build seller profile/settings:
  - shop info
  - logo upload
  - banner upload
  - seller settings
- Build seller dashboard:
  - stats
  - sales chart
  - top products
  - recent orders
  - low-stock alerts
- Build product management:
  - product list with filters/statuses
  - create/edit product
  - rich description input if needed
  - image upload/reorder/delete
  - variant options and values
  - variant create/edit/delete
  - publish/unpublish
  - duplicate
  - delete
- Build inventory management:
  - inventory list
  - item detail
  - quantity adjustment
  - movement history
  - low stock
  - bulk update
- Build order management:
  - seller orders list/detail
  - accept/reject
  - mark processing
  - pack
  - ship
  - cancel
  - refund action
  - approve/reject return
- Build shipping:
  - settings
  - shipment list/detail/create/edit
  - print label
  - mark shipped
  - tracking
- Build vouchers/promotions:
  - list/create/edit/delete
  - activate/deactivate
  - promotion management
- Build reviews:
  - list/detail
  - reply/edit/delete reply
  - report review
- Build wallet and payouts:
  - wallet summary
  - transactions
  - payout list/detail/request
  - bank account CRUD/default
- Build seller analytics:
  - sales
  - orders
  - products
  - customers
  - traffic
  - conversion
  - refunds

### Deliverables

- Seller center dashboard.
- Product and inventory management UI.
- Seller order/shipping flows.
- Seller wallet/payout UI.
- Seller E2E tests.

## Phase 6: Admin Console

### Requirements

- Admins and staff can manage the marketplace, approve sellers/products, moderate content, approve manual Wave payments, manage CMS, configure settings, and inspect system health.

### Tasks

- Build admin auth and protected admin shell.
- Build dashboard:
  - stats
  - sales chart
  - order chart
  - user growth
  - seller growth
- Build user management:
  - list/create/detail/edit/delete
  - ban/unban
  - verify email/phone
  - user orders
  - user activity
- Build seller management:
  - list/detail/edit
  - approve/reject
  - suspend/unsuspend
  - seller products/orders/payouts
- Build product management:
  - list/detail/edit
  - approve/reject
  - ban/unban
  - delete
  - reported products resolution
- Build categories and brands:
  - CRUD
  - category reorder
- Build orders:
  - list/detail/edit
  - cancel/refund
  - status history
  - payments
  - shipments
- Build payments:
  - payment list/detail
  - manual Wave approval/rejection
  - refund action
  - webhook event list/detail/reprocess for future providers
- Build shipping:
  - provider CRUD
  - shipment list/detail/edit/tracking
- Build vouchers/campaigns/flash sales.
- Build reviews and moderation:
  - reviews
  - hide/unhide/delete
  - reported reviews/users/sellers
  - moderation cases and resolution
- Build CMS:
  - pages
  - banners and reorder
  - FAQs
  - announcements
- Build RBAC:
  - roles
  - permissions
  - role permissions
  - user roles
- Build settings:
  - platform settings
  - feature flags
  - audit logs
  - activity logs
  - system health

### Deliverables

- Complete admin console.
- Manual Wave approval UI and audit trail.
- CMS management UI.
- Admin permission tests.
- Admin E2E tests.

## Phase 7: Real-Time, Notifications, And Chat UI

### Requirements

- Buyers and sellers can receive order, payment, chat, and notification updates through polling first and WebSocket where available.

### Tasks

- Build notification center and unread count.
- Add WebSocket client abstraction for:
  - `/ws/chat`
  - `/ws/notifications`
  - `/ws/order-status`
- Add reconnection, auth refresh, backoff, and fallback polling.
- Build buyer/seller chat:
  - conversation list
  - conversation detail
  - message send
  - attachment upload
  - read status
  - delete message/conversation behavior
- Add real-time order status updates on order detail pages.
- Add seller/admin notification indicators for pending actions.

### Deliverables

- Chat UI.
- Notification center.
- WebSocket client abstraction.
- Real-time fallback behavior tests.

## Phase 8: Quality, Accessibility, Localization, And Performance

### Requirements

- Web must be reliable, inclusive, fast, and ready for production growth.

### Tasks

- Add accessibility testing:
  - keyboard flows
  - screen reader labels
  - focus management
  - color contrast
  - modal/drawer behavior
- Add localization readiness:
  - string extraction
  - date/money formatting
  - language switch if required
  - RTL readiness if needed later
- Add performance optimizations:
  - image optimization
  - code splitting
  - lazy loading heavy admin/seller charts
  - cache policies
  - debounced search
  - pagination/virtualization for large tables
- Add error resilience:
  - error boundaries
  - retry affordances
  - offline/network lost indicators
  - stale session recovery
- Add analytics:
  - product view
  - search submit/click
  - add to cart
  - checkout start
  - place order
  - payment method selected
  - seller product create/publish
  - admin moderation actions

### Deliverables

- Accessibility report.
- Performance report.
- Analytics event map.
- Error-state inventory.

## Phase 9: Testing, Deployment, And Launch Readiness

### Requirements

- Web releases must be automated, tested, observable, and rollback-ready.

### Tasks

- Unit tests for utilities, API wrappers, validators, formatters, auth helpers, and state reducers.
- Component tests for shared UI, forms, tables, uploaders, status components, and dialogs.
- Contract tests against generated API types.
- Playwright E2E:
  - auth/register/login
  - browse/search/product detail
  - cart/checkout/place-order
  - Wave manual pending flow
  - POD checkout
  - order tracking
  - seller apply
  - seller create product
  - seller ship order
  - admin approve seller/product
  - admin approve Wave payment
  - admin moderation
- Add visual regression testing for core design-system components and key pages if desired.
- Add deployment pipeline:
  - preview deployments
  - staging deployment
  - production deployment
  - environment variable validation
  - smoke tests
  - rollback path
- Add production monitoring:
  - Sentry errors
  - web vitals
  - uptime checks
  - API error dashboards
  - checkout conversion dashboards

### Deliverables

- Automated test suite.
- CI/CD pipeline.
- Production deployment runbook.
- Web launch checklist.
- Monitoring dashboards.

## MVP Build Order

1. Web foundation, API client, auth/session, design system.
2. Buyer storefront: home, categories, products, search, product detail.
3. Buyer cart, checkout, Wave/POD payment UX, orders.
4. Seller onboarding, seller profile, product/variant/inventory management.
5. Seller orders and shipping.
6. Admin dashboard, users, sellers, products, orders, manual Wave approval.
7. Reviews, notifications, basic CMS, promotions.
8. Chat, real-time updates, analytics, performance and accessibility hardening.

## MVP Page Checklist

- [ ] Home
- [ ] Categories
- [ ] Product listing
- [ ] Product detail
- [ ] Search results
- [ ] Login/register
- [ ] Buyer profile
- [ ] Addresses
- [ ] Cart
- [ ] Checkout
- [ ] Wave payment instructions/pending confirmation
- [ ] POD checkout confirmation
- [ ] Buyer orders list/detail
- [ ] Review form
- [ ] Notifications
- [ ] Seller application
- [ ] Seller dashboard
- [ ] Seller profile/settings
- [ ] Seller products
- [ ] Seller product create/edit
- [ ] Seller inventory
- [ ] Seller orders
- [ ] Seller shipping
- [ ] Admin dashboard
- [ ] Admin users
- [ ] Admin sellers
- [ ] Admin products
- [ ] Admin orders
- [ ] Admin payments/manual Wave approvals
- [ ] Admin moderation

## Open Decisions

- Whether buyer storefront, seller center, and admin console live in one Next.js app or separate apps in a monorepo.
- Whether web auth uses secure HTTP-only cookies, bearer tokens, or a hybrid approach.
- Final brand direction and visual identity.
- Final launch languages, currency, country/city coverage, and tax/shipping copy.
- Whether seller center must be fully responsive for mobile web or desktop-first for v1.
