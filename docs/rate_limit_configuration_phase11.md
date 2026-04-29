# Phase 11 Rate-Limit Configuration

## Current Strategy

- Global limiter middleware enabled for API requests.
- Key dimensions:
  - client IP
  - authenticated user ID (if present)
  - route template (`FullPath`)
  - HTTP method
- Route weighting:
  - `3x` for auth-risk endpoints (`/auth/login`, `/auth/refresh-token`, `/auth/send-phone-otp`)
  - `2x` for mutation methods (`POST`, `PATCH`, `DELETE`)
  - `1x` default.

## Config Keys

- `rate_limit.requests_per_second`
- `rate_limit.burst`

## Auth Risk Lockout

- Login lockout keys:
  - `auth:login:fail:<email>`
  - `auth:login:lock:<email>`
- OTP verify lockout key:
  - `auth:otp:verify:fail:<user_id>:<phone>`

## Recommended Production Baseline

- `rate_limit.requests_per_second=20`
- `rate_limit.burst=40`
- `auth.max_login_failures=8`
- `auth.login_lockout_duration=20m`
- `auth.otp_max_verify_failures=5`

Tune based on observed traffic and false positive rates.
