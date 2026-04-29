# Error-State Inventory (Phase 8)

## Runtime Boundaries
- Global app boundary: `app/global-error.tsx`
- Section boundaries:
  - `app/(storefront)/error.tsx`
  - `app/account/error.tsx`
  - `app/seller/error.tsx`
  - `app/admin/error.tsx`

## API/Session Failure States
- Access token expired + refresh failure:
  - Captured in `lib/api/bitik-fetch.ts` via `bitik.session.expired.v1`.
  - User-facing stale session banner in `components/system/network-and-session-banner.tsx`.
- Generic request failures:
  - Standardized retry surface via `components/ui/error-state.tsx`.

## Connectivity States
- Offline network loss:
  - Online/offline events shown via `components/system/network-and-session-banner.tsx`.
  - User informed that some actions are unavailable until reconnect.

## Recovery Affordances
- Retry actions on boundaries and inline error states.
- Explicit sign-in redirect action when stale session is detected.

## Remaining Follow-up
- Add per-feature retry counters/telemetry for repeated failures.
- Expand stale-session messaging into auth-specific pages for richer remediation details.

