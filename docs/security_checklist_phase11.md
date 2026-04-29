# Phase 11 Security Checklist

## Runtime Controls

- [x] Password hashing with bcrypt for user passwords.
- [x] Request body limits via middleware (`security.max_request_body_bytes`).
- [x] Upload limits enforced in media handlers (`security.max_upload_bytes`).
- [x] Rate limiting keyed by IP + user + route + method.
- [x] Auth-risk weighting for high-risk routes (`/auth/login`, `/auth/refresh-token`, `/auth/send-phone-otp`).
- [x] Account login lockout after repeated failures.
- [x] OTP verify lockout after repeated invalid codes.
- [x] Webhook signature verification support for payment providers.
- [x] CORS credentials disabled automatically when wildcard origins are configured.
- [x] Secure refresh cookie policy with SameSite controls and stricter production defaults.

## Data + Compliance Controls

- [x] PII handling and deletion workflow documented.
- [x] Audit logging present for auth, orders, payments, shipping, seller/admin mutations (see audit report).
- [x] Backup and restore procedure documented for PostgreSQL and object-storage metadata.

## Supply Chain + Operations Controls

- [x] Dependency vulnerability scan in CI (`govulncheck`).
- [x] Static security scan in CI (`gosec`).
- [x] Filesystem/container-oriented vulnerability scan in CI (`trivy`).
- [x] DB pooling and timeout policy configured (`database.max_open_conns`, `connect_timeout`, `query_timeout`).
- [x] Graceful shutdown implemented for API and worker services.
