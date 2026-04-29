# Bitik Web Design System (Phase 0)

## Stack

- Next.js 16 + React 19 + TypeScript
- Tailwind CSS v4 with `@theme inline`
- shadcn-style component layer over `@base-ui/react` primitives
- `lucide-react` icons
- `next-themes` for light/dark
- `sonner` toast notifications

## Tokens

Defined in `bitik-web/app/globals.css` as CSS variables, exposed to Tailwind via `@theme inline`.

### Color (semantic)

`background`, `foreground`, `card`, `popover`, `muted`, `accent`, `border`, `input`, `ring`, `primary`, `secondary`, `sidebar*`, `chart-1..5`, plus status colors:

| Token | Use |
|-------|-----|
| `--success`, `--success-foreground` | success state, confirmations |
| `--info`, `--info-foreground` | informational |
| `--warning`, `--warning-foreground` | warnings, manual confirmation |
| `--destructive`, `--destructive-foreground` | errors, destructive actions |
| `--pending`, `--pending-foreground` | neutral pending state |

Both light and dark are defined. Status badges in `components/ui/status-badge.tsx` map domain status values onto these tones via `lib/status.ts`.

### Radius

`--radius` (0.625rem base) with derived `--radius-sm`, `--radius-md`, `--radius-lg`, `--radius-xl`, `--radius-2xl`, `--radius-3xl`, `--radius-4xl`.

### Shadows

`--shadow-xs`, `--shadow-sm`, `--shadow-md`, `--shadow-lg`, `--shadow-xl`, `--shadow-2xl`. Use Tailwind utilities `shadow-xs` … `shadow-2xl` mapped to these.

### Z-index layers

| Token | Use |
|-------|-----|
| `--z-base: 0` | default |
| `--z-dropdown: 30` | inline menus, comboboxes |
| `--z-sticky: 40` | sticky headers |
| `--z-fixed: 50` | fixed nav |
| `--z-overlay: 70` | sheet/modal backdrop |
| `--z-modal: 80` | dialog content |
| `--z-popover: 90` | floating popovers |
| `--z-toast: 100` | toasts |
| `--z-tooltip: 110` | tooltips |

### Breakpoints

`--breakpoint-xs 22rem`, `sm 40rem`, `md 48rem`, `lg 64rem`, `xl 80rem`, `2xl 96rem`. Use Tailwind responsive prefixes.

### Typography

- `--font-sans` (Inter)
- `--font-mono` (Geist Mono)
- `--font-heading` mirrors sans by default; can be re-pointed for marketing surfaces.

### Spacing

Tailwind v4 defaults. Use `gap-*`, `p-*`, `space-*`. No custom spacing tokens to keep simple.

## Component library

Shared UI is under `bitik-web/components/ui/` (shadcn-style):

| Group | Components |
|-------|-----------|
| Buttons | `button` |
| Inputs | `input`, `textarea`, `label`, `field`, `input-group`, `input-otp`, `native-select`, `select`, `combobox`, `radio-group`, `checkbox`, `switch`, `toggle`, `toggle-group`, `slider` |
| Surfaces | `card`, `sheet`, `dialog`, `drawer`, `popover`, `hover-card`, `tooltip`, `accordion`, `collapsible`, `tabs`, `alert`, `alert-dialog` |
| Navigation | `breadcrumb`, `pagination`, `navigation-menu`, `menubar`, `dropdown-menu`, `context-menu`, `command`, `sidebar` |
| Data | `table`, `chart`, `progress`, `calendar`, `carousel` |
| Feedback | `badge`, `status-badge` (Bitik), `skeleton`, `spinner`, `kbd`, `sonner` (toaster), `empty`, `error-state` (Bitik) |
| Media | `avatar`, `aspect-ratio`, `image-gallery` (Bitik), `file-upload` (Bitik) |
| Layout | `separator`, `scroll-area`, `resizable` |

Bitik-specific:

- `status-badge.tsx` — domain-typed status badge backed by `lib/status.ts`.
- `error-state.tsx` — full-card error placeholder with retry.
- `file-upload.tsx` — drag-and-drop multi-file picker with size/file-count validation.
- `image-gallery.tsx` — accessible main image + thumbnails with keyboard navigation.

Surface chrome:

- `components/storefront/header.tsx`, `components/storefront/footer.tsx`
- `components/account/sidebar.tsx`
- `components/seller/sidebar.tsx`
- `components/admin/sidebar.tsx`
- `components/shared/sidebar-nav.tsx` — shared item renderer with `aria-current="page"`
- `components/shared/dashboard-shell.tsx` — header + sidebar + main shell for seller/admin
- `components/shared/page-placeholder.tsx` — Phase 0 stub helper

## Status UI

Domain status enums and their UI metadata live in `bitik-web/lib/status.ts`:

| Kind | Source | UI |
|------|--------|----|
| Order | `OrderStatus` | `<StatusBadge kind="order" value=... />` |
| Payment | `PaymentStatus` | `<StatusBadge kind="payment" ... />` |
| Shipment | `ShipmentStatus` | `<StatusBadge kind="shipment" ... />` |
| Seller | `SellerStatus` | `<StatusBadge kind="seller" ... />` |
| Product | `ProductStatus` | `<StatusBadge kind="product" ... />` |
| Refund | `RefundStatus` | `<StatusBadge kind="refund" ... />` |
| Return | `ReturnStatus` | `<StatusBadge kind="return" ... />` |

Tones map: `neutral → muted`, `info → info`, `success → success`, `warning → warning`, `danger → destructive`, `pending → pending`.

## Formatters

`bitik-web/lib/format.ts`:

- `formatMoney(value, { locale?, currency?, compact? })`
- `formatMoneyRange(min, max, opts?)`
- `formatNumber(value, opts?)`
- `formatDate(value, opts?)` (default: `Mar 15, 2026`)
- `formatDateTime(value, opts?)`
- `formatRelative(value, { now? })` — `2 hours ago`, `in 5 minutes`
- `formatPercent(value, opts?)`
- `truncate(value, max)`
- `initials(name)` for avatar fallback

Default currency is `MMK`; pass `{ currency: "USD" }` if needed.

## Theme

`next-themes` with class-based dark mode. `ThemeProvider` lives in `components/theme-provider.tsx` and is wrapped in the root layout. A `d` keypress toggles modes (skipped while typing in inputs).

## Reference page

`/style-guide` renders all key components against current tokens. Use it as the visual smoke-test page when working on theme or token changes.
