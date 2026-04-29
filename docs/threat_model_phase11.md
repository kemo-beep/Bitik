# Phase 11 Threat Model

## Scope

- Authentication and session lifecycle.
- Checkout/order placement and order transitions.
- Payment confirmation, webhooks, refunds, and seller wallet settlement paths.
- Seller/admin privileged workflows.

## Key Threats and Mitigations

## Auth + Session

- **Credential stuffing/brute force**
  - Mitigation: route-weighted rate limits and login lockout after repeated failures.
- **OTP guessing**
  - Mitigation: OTP verify failure lockout and OTP send-rate limits.
- **Refresh token replay**
  - Mitigation: refresh-token rotation and reuse detection with full-session revocation.
- **Cookie theft / CSRF abuse**
  - Mitigation: `HttpOnly` refresh cookie, `Secure` option, SameSite policy (strict in production).

## Checkout + Orders

- **Duplicate writes from retries/network failures**
  - Mitigation: idempotency middleware + Redis lock/replay cache.
- **Cross-account resource access**
  - Mitigation: buyer/seller scoped queries require authenticated identity and owner IDs.
- **Large-body DoS**
  - Mitigation: global request body cap and upload-specific limits.

## Payments + Webhooks

- **Forged provider callbacks**
  - Mitigation: HMAC signature verification for payment webhook ingestion.
- **Unauthorized status transitions**
  - Mitigation: role checks, provider-specific transition rules, and state-machine checks.
- **Financial mutation tampering**
  - Mitigation: audit + admin activity logs for payment and settlement state changes.

## Seller/Admin Workflows

- **Privilege misuse / destructive actions without trace**
  - Mitigation: RBAC middleware + mutation audit logs.
- **Sensitive configuration drift**
  - Mitigation: environment-based config controls and CI security scans.

## Residual Risks / Next Hardening

- Move lockout/risk counters from Redis-only signals to persisted risk-event tables for longer forensic retention.
- Add signed outgoing webhooks/events for downstream internal services.
- Add WAF/risk-engine scoring hooks for geo/device anomaly detection.
