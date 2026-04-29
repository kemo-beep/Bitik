# Bitik Web Accessibility Checklist (Phase 0 baseline)

Target: WCAG 2.1 AA for core flows.

## Global baseline

- [x] `<html lang="en">` set in `app/layout.tsx`.
- [x] `viewport` defined; `themeColor` matches light/dark.
- [x] Skip link to `#main-content` rendered before any nav.
- [x] Polite live region (`#app-live-region`) provided for async announcements.
- [x] `:focus-visible` outline ring is visible against background and dark mode.
- [x] No CSS removes focus outlines globally.
- [x] Color tokens contain explicit `*-foreground` pairs to keep contrast in both themes.

## Layout shells

- [x] `header role="banner"`, `main role="main"`, `footer role="contentinfo"` per surface.
- [x] Storefront `<nav aria-label="Primary">`; footer sections use `<nav aria-label="...">`.
- [x] Sidebars use `<aside aria-label="...">` and link items use `aria-current="page"`.
- [x] Sticky headers use a non-trapping `position: sticky` with sufficient contrast and backdrop blur fallback.

## Components

- [x] `Button` (base-ui) renders a real `<button>` by default; when used as a link, `nativeButton={false}` with `render={<a/Link>}` so semantics match.
- [x] `Input`, `Textarea`, `Select` use associated `Label`; placeholder is not a label.
- [x] Checkbox / Radio / Switch from `@base-ui/react` ship with proper `role` and `aria-checked` state; keyboard activates with Space.
- [x] Dialog and Sheet (base-ui Dialog) trap focus, return focus on close, escape to dismiss.
- [x] Dropdown menu, Tooltip, Popover use base-ui positioner with arrow + roving tab index.
- [x] Tabs use roving tab index and `aria-selected` via base-ui Tabs.
- [x] Breadcrumb component sets `aria-label="breadcrumb"` and current page has `aria-current="page"`.
- [x] Pagination root sets `aria-label="pagination"` and active link sets `aria-current="page"`.
- [x] Toaster (`sonner`) announces in a polite live region with status icons.
- [x] StatusBadge uses semantic color + label text; not color-only.
- [x] FileUpload exposes hidden `<input type="file">`, drop zone is a real `<button>`, errors are announced via `role="alert"`.
- [x] ImageGallery has `role="region" aria-roledescription="image gallery"`, arrow keys navigate, thumbnails are tabs with `aria-selected`.
- [x] ErrorState uses `role="alert" aria-live="polite"`.

## Forms

- [ ] Every field has a visible label or `aria-label`.
- [ ] Validation errors render adjacent and are referenced via `aria-describedby` on the field.
- [ ] Submit button shows loading state without losing focus.
- [ ] OTP fields (`input-otp`) preserve paste behavior and announce remaining characters.
- [ ] Required fields are marked both visually and with `aria-required`.

## Navigation and keyboard

- [ ] Tab order follows reading order on every page.
- [ ] All actionable elements reachable by keyboard; no `tabindex > 0`.
- [ ] Modals: open returns focus to trigger on close; first focus on dialog title or first input.
- [ ] Sheet/drawer: same focus rules as modal.
- [ ] Menu: arrow keys move, Home/End jump, Esc closes, type-ahead works.
- [ ] Search: Enter submits, Esc clears, arrow keys navigate suggestions when present.

## Color and contrast

- [x] Tokens defined in oklch with explicit foreground pairs.
- [ ] Manual contrast verification on style-guide page in both themes (text 4.5:1, large text 3:1, UI components 3:1).
- [ ] Status colors do not rely on hue alone; always paired with label text.

## Media

- [ ] Every product image has descriptive `alt`; decorative images use `alt=""`.
- [ ] Video, when added, ships with captions and a transcript link.

## Internationalization

- [ ] `lang` updates when locale switches.
- [ ] No text baked into images.
- [ ] `formatMoney`/`formatDate`/`formatRelative` accept locale; defaults to `en-US` and is configurable from a single place.

## Tooling

- [ ] axe-core or Playwright a11y assertions added in Phase 9 testing.
- [ ] Storybook (or component preview) optional, runs a11y addon when adopted.
- [ ] Lighthouse: run on home, product detail, cart, checkout, seller dashboard, admin dashboard before launch.
- [ ] Manual audit with VoiceOver (macOS) and NVDA (Windows) on auth, checkout, Wave manual confirmation, seller order ship, admin Wave approval.

Items marked `[x]` are baseline by Phase 0 layout/component work. `[ ]` items are required acceptance criteria for Phases 2–9.
