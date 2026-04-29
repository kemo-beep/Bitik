-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at, session_id)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at, session_id;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, session_id
FROM refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = $1
  AND revoked_at IS NULL;

-- name: RevokeAllRefreshTokensForUser :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = $1
  AND revoked_at IS NULL;

-- name: RevokeRefreshTokensForSession :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE session_id = $1
  AND user_id = $2
  AND revoked_at IS NULL;

-- name: CreateUserSession :one
INSERT INTO user_sessions (user_id, device_id, user_agent, ip_address, platform, push_token)
VALUES (
  $1,
  $2,
  $3,
  CASE
    WHEN sqlc.narg('ip_text')::text = '' THEN NULL
    ELSE CAST(sqlc.narg('ip_text')::text AS inet)
  END,
  $4,
  $5
)
RETURNING id, user_id, device_id, user_agent, ip_address, platform, push_token, last_seen_at, revoked_at, created_at, updated_at;

-- name: GetUserSessionByIDForUser :one
SELECT id, user_id, device_id, user_agent, ip_address, platform, push_token, last_seen_at, revoked_at, created_at, updated_at
FROM user_sessions
WHERE id = $1
  AND user_id = $2
  AND revoked_at IS NULL;

-- name: GetUserSessionByID :one
SELECT id, user_id, device_id, user_agent, ip_address, platform, push_token, last_seen_at, revoked_at, created_at, updated_at
FROM user_sessions
WHERE id = $1
  AND user_id = $2;

-- name: ListUserSessionsForUser :many
SELECT id, user_id, device_id, user_agent, ip_address, platform, push_token, last_seen_at, revoked_at, created_at, updated_at
FROM user_sessions
WHERE user_id = $1
ORDER BY last_seen_at DESC;

-- name: RevokeUserSession :exec
UPDATE user_sessions
SET revoked_at = now()
WHERE id = $1
  AND user_id = $2
  AND revoked_at IS NULL;

-- name: RevokeAllUserSessions :exec
UPDATE user_sessions
SET revoked_at = now()
WHERE user_id = $1
  AND revoked_at IS NULL;

-- name: TouchUserSession :exec
UPDATE user_sessions
SET last_seen_at = now()
WHERE id = $1
  AND user_id = $2
  AND revoked_at IS NULL;

-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (user_id, email, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, email, token_hash, expires_at, consumed_at, created_at;

-- name: ConsumeEmailVerificationToken :one
UPDATE email_verification_tokens
SET consumed_at = now()
WHERE token_hash = $1
  AND consumed_at IS NULL
  AND expires_at > now()
RETURNING id, user_id, email, token_hash, expires_at, consumed_at, created_at;

-- name: CreatePhoneOTPAttempt :one
INSERT INTO phone_otp_attempts (user_id, phone, otp_hash, purpose, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, phone, otp_hash, purpose, attempts, max_attempts, expires_at, verified_at, created_at;

-- name: GetLatestPendingPhoneOTPForUser :one
SELECT id, user_id, phone, otp_hash, purpose, attempts, max_attempts, expires_at, verified_at, created_at
FROM phone_otp_attempts
WHERE user_id = $1
  AND purpose = $2
  AND verified_at IS NULL
  AND expires_at > now()
ORDER BY created_at DESC
LIMIT 1;

-- name: GetPhoneOTPAttemptByID :one
SELECT id, user_id, phone, otp_hash, purpose, attempts, max_attempts, expires_at, verified_at, created_at
FROM phone_otp_attempts
WHERE id = $1;

-- name: VerifyPhoneOTPAttempt :one
UPDATE phone_otp_attempts
SET verified_at = now()
WHERE id = $1
  AND verified_at IS NULL
  AND expires_at > now()
  AND attempts < max_attempts
RETURNING id, user_id, phone, otp_hash, purpose, attempts, max_attempts, expires_at, verified_at, created_at;

-- name: IncrementPhoneOTPAttempts :exec
UPDATE phone_otp_attempts
SET attempts = attempts + 1
WHERE id = $1
  AND verified_at IS NULL;

-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, consumed_at, created_at;

-- name: ConsumePasswordResetToken :one
UPDATE password_reset_tokens
SET consumed_at = now()
WHERE token_hash = $1
  AND consumed_at IS NULL
  AND expires_at > now()
RETURNING id, user_id, token_hash, expires_at, consumed_at, created_at;

-- name: UpsertOAuthIdentity :one
INSERT INTO oauth_identities (user_id, provider, provider_user_id, email, raw_profile)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (provider, provider_user_id) DO UPDATE
SET email = EXCLUDED.email,
    raw_profile = EXCLUDED.raw_profile
RETURNING id, user_id, provider, provider_user_id, email, raw_profile, created_at, updated_at;
