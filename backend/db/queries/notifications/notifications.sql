-- name: CountUnreadNotifications :one
SELECT count(*)::bigint
FROM notifications
WHERE user_id = $1
  AND read_at IS NULL;

-- name: ListNotifications :many
SELECT id, user_id, type, title, body, data, read_at, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountNotifications :one
SELECT count(*)::bigint
FROM notifications
WHERE user_id = $1;

-- name: MarkNotificationRead :one
UPDATE notifications
SET read_at = COALESCE(read_at, now())
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, type, title, body, data, read_at, created_at;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = now()
WHERE user_id = $1
  AND read_at IS NULL;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: GetNotificationPreferences :one
SELECT user_id, email_enabled, sms_enabled, push_enabled, marketing_enabled, quiet_hours, created_at, updated_at
FROM notification_preferences
WHERE user_id = $1;

-- name: UpsertNotificationPreferences :one
INSERT INTO notification_preferences (user_id, email_enabled, sms_enabled, push_enabled, marketing_enabled, quiet_hours)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id)
DO UPDATE SET email_enabled = EXCLUDED.email_enabled,
              sms_enabled = EXCLUDED.sms_enabled,
              push_enabled = EXCLUDED.push_enabled,
              marketing_enabled = EXCLUDED.marketing_enabled,
              quiet_hours = EXCLUDED.quiet_hours
RETURNING user_id, email_enabled, sms_enabled, push_enabled, marketing_enabled, quiet_hours, created_at, updated_at;

-- name: ListPushTokens :many
SELECT id, user_id, token, platform, created_at
FROM push_tokens
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreatePushToken :one
INSERT INTO push_tokens (user_id, token, platform)
VALUES ($1, $2, $3)
ON CONFLICT (token)
DO UPDATE SET user_id = EXCLUDED.user_id,
              platform = EXCLUDED.platform
RETURNING id, user_id, token, platform, created_at;

-- name: DeletePushToken :exec
DELETE FROM push_tokens
WHERE user_id = $1
  AND token = $2;

-- name: CreateInAppNotification :one
INSERT INTO notifications (user_id, type, title, body, data)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, type, title, body, data, read_at, created_at;

