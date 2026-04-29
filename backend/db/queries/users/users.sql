-- name: CreateUser :one
INSERT INTO users (email, phone, password_hash, status)
VALUES ($1, $2, $3, COALESCE($4, 'active'))
RETURNING id, email, phone, password_hash, status, email_verified, phone_verified, last_login_at, deleted_at, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, phone, password_hash, status, email_verified, phone_verified, last_login_at, deleted_at, created_at, updated_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, phone, password_hash, status, email_verified, phone_verified, last_login_at, deleted_at, created_at, updated_at
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- name: UpdateUserContact :one
UPDATE users
SET email = COALESCE(sqlc.narg('email')::citext, email),
    phone = COALESCE(sqlc.narg('phone')::text, phone),
    email_verified = CASE
      WHEN sqlc.narg('email')::citext IS NOT NULL AND sqlc.narg('email')::citext IS DISTINCT FROM email THEN FALSE
      ELSE email_verified
    END,
    phone_verified = CASE
      WHEN sqlc.narg('phone')::text IS NOT NULL AND sqlc.narg('phone')::text IS DISTINCT FROM phone THEN FALSE
      ELSE phone_verified
    END
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id, email, phone, password_hash, status, email_verified, phone_verified, last_login_at, deleted_at, created_at, updated_at;

-- name: UpdateUserProfile :one
INSERT INTO user_profiles (user_id, first_name, last_name, display_name, avatar_url, gender, birthdate, language, country_code, timezone)
VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, 'en'), $9, $10)
ON CONFLICT (user_id) DO UPDATE
SET first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    display_name = EXCLUDED.display_name,
    avatar_url = EXCLUDED.avatar_url,
    gender = EXCLUDED.gender,
    birthdate = EXCLUDED.birthdate,
    language = EXCLUDED.language,
    country_code = EXCLUDED.country_code,
    timezone = EXCLUDED.timezone
RETURNING id, user_id, first_name, last_name, display_name, avatar_url, gender, birthdate, language, country_code, timezone, created_at, updated_at;

-- name: ListUserAddresses :many
SELECT id, user_id, full_name, phone, country, state, city, district, postal_code, address_line1, address_line2, latitude, longitude, is_default, deleted_at, created_at, updated_at
FROM user_addresses
WHERE user_id = $1
  AND deleted_at IS NULL
ORDER BY is_default DESC, created_at DESC;

-- name: CreateUserDevice :one
INSERT INTO user_devices (user_id, device_id, platform, app_version, push_token, last_seen_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (user_id, device_id) DO UPDATE
SET platform = EXCLUDED.platform,
    app_version = EXCLUDED.app_version,
    push_token = EXCLUDED.push_token,
    last_seen_at = now(),
    revoked_at = NULL
RETURNING id, user_id, device_id, platform, app_version, push_token, last_seen_at, revoked_at, created_at, updated_at;

-- name: GetUserProfileByUserID :one
SELECT id, user_id, first_name, last_name, display_name, avatar_url, gender, birthdate, language, country_code, timezone, created_at, updated_at
FROM user_profiles
WHERE user_id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = now()
WHERE id = $1;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = now(),
    status = 'deleted'
WHERE id = $1
  AND deleted_at IS NULL;

-- name: SetUserEmailVerified :exec
UPDATE users
SET email_verified = TRUE
WHERE id = $1;

-- name: SetUserPhoneVerified :exec
UPDATE users
SET phone_verified = TRUE
WHERE id = $1;

-- name: UpdateUserPasswordHash :exec
UPDATE users
SET password_hash = $2
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListUserDevicesForUser :many
SELECT id, user_id, device_id, platform, app_version, push_token, last_seen_at, revoked_at, created_at, updated_at
FROM user_devices
WHERE user_id = $1
  AND deleted_at IS NULL
ORDER BY last_seen_at DESC;

-- name: RevokeUserDevice :exec
UPDATE user_devices
SET revoked_at = now()
WHERE user_id = $1
  AND device_id = $2
  AND revoked_at IS NULL;

-- name: AdminCountUsers :one
SELECT COUNT(1)::bigint
FROM users u
WHERE u.deleted_at IS NULL
  AND (sqlc.narg('status')::text IS NULL OR u.status = sqlc.narg('status')::text)
  AND (sqlc.narg('q')::text IS NULL OR u.email ILIKE ('%' || sqlc.narg('q')::text || '%'));

-- name: AdminListUsers :many
SELECT id, email, phone, password_hash, status, email_verified, phone_verified, last_login_at, deleted_at, created_at, updated_at
FROM users u
WHERE u.deleted_at IS NULL
  AND (sqlc.narg('status')::text IS NULL OR u.status = sqlc.narg('status')::text)
  AND (sqlc.narg('q')::text IS NULL OR u.email ILIKE ('%' || sqlc.narg('q')::text || '%'))
ORDER BY u.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminUpdateUserStatus :one
UPDATE users
SET status = sqlc.arg('status'),
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id, email, phone, password_hash, status, email_verified, phone_verified, last_login_at, deleted_at, created_at, updated_at;
