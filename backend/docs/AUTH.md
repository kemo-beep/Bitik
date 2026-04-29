# Authentication and sessions

This document describes Phase 2 auth behavior implemented in the Bitik API.

## Schema migrations

The API process can apply Goose migrations on startup when `database.auto_migrate` is true (`BITIK_DATABASE_AUTO_MIGRATE`, default true). Disable in production if your deploy pipeline runs `cmd/migrate` (or Goose) explicitly.

## Tokens

- **Access token**: JWT (HS256), short-lived (`auth.access_token_ttl`, default 15 minutes). Claims include subject (user UUID), issuer, expiry, and `roles` (role names from `user_roles` / `roles` at token issuance). Protected route middleware reloads the current user and role names from the database, so bans/deletions/role changes take effect before the JWT expires.
- **Refresh token**: Opaque random string, SHA-256 hashed at rest in `refresh_tokens.token_hash`. Returned once to the client; plaintext secrets are only logged when `auth.log_secrets` is explicitly enabled.

Configure `BITIK_AUTH_JWT_SECRET` (via `auth.jwt_secret`) to a long random value in every non-local environment.

## Refresh rotation and reuse

- Each successful **refresh** revokes the previous refresh row and inserts a new hash for the same `user_sessions` row when `session_id` is present.
- If a **revoked** refresh token is presented again, the service treats this as reuse: all active refresh tokens and sessions for that user are revoked and the client must sign in again.
- Revoking a session also revokes refresh tokens bound to that session. Refresh rejects tokens bound to revoked sessions.

## OAuth (Google / Facebook / Apple)

- CSRF **state** is stored in Redis (`oauth:state:{state}`) with a 10-minute TTL. OAuth initiation returns `503` if Redis is unavailable.
- Callback exchanges the code, loads the provider profile, upserts `oauth_identities`, and issues the same access/refresh pair as password login.
- OAuth email-based account linking requires provider-verified emails. Google checks `verified_email`; Facebook checks `verified` / `is_verified`; Apple checks `email_verified` from `id_token`.
- Apple uses an ES256 client secret generated from `auth.oauth_apple_team_id`, `auth.oauth_apple_key_id`, and `auth.oauth_apple_private_key_pem`.

## Phone OTP

- OTP is bcrypt-hashed in `phone_otp_attempts`. With Redis, rate limiting uses `INCR` per phone per hour (`auth.otp_max_per_hour`, default 5).
- Production delivery should inject `OTPSender`; plaintext OTP logging is disabled unless `auth.log_secrets` is true.

## Password reset and email verification

- **Forgot password** always responds `200` with `{ok:true}` to avoid email enumeration; tokens are written to `password_reset_tokens`. Production delivery should inject `EmailSender`; plaintext token logging is disabled unless `auth.log_secrets` is true.
- **Email verification** consumes `email_verification_tokens` and sets `users.email_verified`.

## RBAC (Casbin)

- Policies are loaded from the `casbin_rule` table at process start (`ptype = p`, subject = role name, object = path pattern, action `*` or HTTP verb).
- HTTP middleware checks **any** role on the JWT against Casbin for the request path and method. Admin and seller route groups attach this middleware in addition to JWT auth.

## Audit

Sensitive actions write rows to `audit_logs` (login, register, refresh reuse, OAuth, password flows, self-delete, etc.).

## Integration tests

JWT unit tests run in CI. Database refresh rotation is behind `-tags=integration` and `BITIK_DATABASE_URL`.
