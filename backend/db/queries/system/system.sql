-- name: GetPlatformSetting :one
SELECT key, value, description, is_public, created_at, updated_at
FROM platform_settings
WHERE key = $1;

-- name: UpsertPlatformSetting :one
INSERT INTO platform_settings (key, value, description, is_public)
VALUES ($1, $2, $3, $4)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    description = EXCLUDED.description,
    is_public = EXCLUDED.is_public
RETURNING key, value, description, is_public, created_at, updated_at;

-- name: ListFeatureFlags :many
SELECT key, description, enabled, rules, created_at, updated_at
FROM feature_flags
ORDER BY key ASC;

-- name: UpsertFeatureFlag :one
INSERT INTO feature_flags (key, description, enabled, rules)
VALUES (sqlc.arg('key'), sqlc.narg('description'), sqlc.arg('enabled'), COALESCE(sqlc.narg('rules')::jsonb, '{}'::jsonb))
ON CONFLICT (key) DO UPDATE
SET description = EXCLUDED.description,
    enabled = EXCLUDED.enabled,
    rules = EXCLUDED.rules
RETURNING key, description, enabled, rules, created_at, updated_at;

-- name: ListPlatformSettings :many
SELECT key, value, description, is_public, created_at, updated_at
FROM platform_settings
ORDER BY key ASC;

-- name: CreateAuditLog :one
INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, actor_user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at;

-- name: CreateModerationReport :one
INSERT INTO moderation_reports (reporter_user_id, target_type, target_id, reason)
VALUES ($1, $2, $3, $4)
RETURNING id, reporter_user_id, target_type, target_id, reason, status, created_at;

-- name: AdminCountModerationReports :one
SELECT COUNT(1)::bigint
FROM moderation_reports mr
WHERE (sqlc.narg('status')::moderation_status IS NULL OR mr.status = sqlc.narg('status')::moderation_status)
  AND (sqlc.narg('target_type')::text IS NULL OR mr.target_type = sqlc.narg('target_type')::text);

-- name: AdminListModerationReports :many
SELECT id, reporter_user_id, target_type, target_id, reason, status, created_at
FROM moderation_reports mr
WHERE (sqlc.narg('status')::moderation_status IS NULL OR mr.status = sqlc.narg('status')::moderation_status)
  AND (sqlc.narg('target_type')::text IS NULL OR mr.target_type = sqlc.narg('target_type')::text)
ORDER BY mr.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminGetModerationReportByID :one
SELECT id, reporter_user_id, target_type, target_id, reason, status, created_at
FROM moderation_reports
WHERE id = $1;

-- name: AdminUpdateModerationReportStatus :one
UPDATE moderation_reports
SET status = sqlc.arg('status')::moderation_status
WHERE id = sqlc.arg('id')
RETURNING id, reporter_user_id, target_type, target_id, reason, status, created_at;

-- name: AdminCountModerationCases :one
SELECT COUNT(1)::bigint
FROM moderation_cases mc
WHERE (sqlc.narg('status')::moderation_status IS NULL OR mc.status = sqlc.narg('status')::moderation_status);

-- name: AdminListModerationCases :many
SELECT id, report_id, assigned_to, status, resolution, resolved_at, created_at, updated_at
FROM moderation_cases mc
WHERE (sqlc.narg('status')::moderation_status IS NULL OR mc.status = sqlc.narg('status')::moderation_status)
ORDER BY mc.updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminGetModerationCaseByID :one
SELECT id, report_id, assigned_to, status, resolution, resolved_at, created_at, updated_at
FROM moderation_cases
WHERE id = $1;

-- name: AdminCreateModerationCase :one
INSERT INTO moderation_cases (report_id, assigned_to, status, resolution, resolved_at)
VALUES (
  sqlc.arg('report_id'),
  sqlc.narg('assigned_to'),
  COALESCE(sqlc.narg('status')::moderation_status, 'under_review'::moderation_status),
  sqlc.narg('resolution'),
  sqlc.narg('resolved_at')
)
RETURNING id, report_id, assigned_to, status, resolution, resolved_at, created_at, updated_at;

-- name: AdminUpdateModerationCase :one
UPDATE moderation_cases
SET assigned_to = COALESCE(sqlc.narg('assigned_to')::uuid, assigned_to),
    status = COALESCE(sqlc.narg('status')::moderation_status, status),
    resolution = COALESCE(sqlc.narg('resolution')::text, resolution),
    resolved_at = CASE
      WHEN sqlc.narg('resolved_at')::timestamptz IS NOT NULL THEN sqlc.narg('resolved_at')::timestamptz
      WHEN sqlc.narg('status')::moderation_status IN ('resolved', 'dismissed') THEN now()
      ELSE resolved_at
    END
WHERE id = sqlc.arg('id')
RETURNING id, report_id, assigned_to, status, resolution, resolved_at, created_at, updated_at;

-- name: AdminCountAuditLogs :one
SELECT COUNT(1)::bigint
FROM audit_logs al
WHERE (sqlc.narg('actor_user_id')::uuid IS NULL OR al.actor_user_id = sqlc.narg('actor_user_id')::uuid)
  AND (sqlc.narg('action')::text IS NULL OR al.action = sqlc.narg('action')::text)
  AND (sqlc.narg('entity_type')::text IS NULL OR al.entity_type = sqlc.narg('entity_type')::text)
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR al.created_at >= sqlc.narg('from_time')::timestamptz)
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR al.created_at <= sqlc.narg('to_time')::timestamptz);

-- name: AdminListAuditLogs :many
SELECT id, actor_user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
FROM audit_logs al
WHERE (sqlc.narg('actor_user_id')::uuid IS NULL OR al.actor_user_id = sqlc.narg('actor_user_id')::uuid)
  AND (sqlc.narg('action')::text IS NULL OR al.action = sqlc.narg('action')::text)
  AND (sqlc.narg('entity_type')::text IS NULL OR al.entity_type = sqlc.narg('entity_type')::text)
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR al.created_at >= sqlc.narg('from_time')::timestamptz)
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR al.created_at <= sqlc.narg('to_time')::timestamptz)
ORDER BY al.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminCountAdminActivityLogs :one
SELECT COUNT(1)::bigint
FROM admin_activity_logs aal
WHERE (sqlc.narg('admin_user_id')::uuid IS NULL OR aal.admin_user_id = sqlc.narg('admin_user_id')::uuid)
  AND (sqlc.narg('action')::text IS NULL OR aal.action = sqlc.narg('action')::text)
  AND (sqlc.narg('entity_type')::text IS NULL OR aal.entity_type = sqlc.narg('entity_type')::text)
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR aal.created_at >= sqlc.narg('from_time')::timestamptz)
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR aal.created_at <= sqlc.narg('to_time')::timestamptz);

-- name: AdminListAdminActivityLogs :many
SELECT id, admin_user_id, action, entity_type, entity_id, metadata, ip_address, user_agent, created_at
FROM admin_activity_logs aal
WHERE (sqlc.narg('admin_user_id')::uuid IS NULL OR aal.admin_user_id = sqlc.narg('admin_user_id')::uuid)
  AND (sqlc.narg('action')::text IS NULL OR aal.action = sqlc.narg('action')::text)
  AND (sqlc.narg('entity_type')::text IS NULL OR aal.entity_type = sqlc.narg('entity_type')::text)
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR aal.created_at >= sqlc.narg('from_time')::timestamptz)
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR aal.created_at <= sqlc.narg('to_time')::timestamptz)
ORDER BY aal.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreateEventLog :one
INSERT INTO event_logs (user_id, event_name, entity_type, entity_id, metadata, ip_address)
VALUES (
  sqlc.narg('user_id'),
  sqlc.arg('event_name'),
  sqlc.narg('entity_type'),
  sqlc.narg('entity_id'),
  COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb),
  sqlc.narg('ip_address')
)
RETURNING id, user_id, event_name, entity_type, entity_id, metadata, ip_address, created_at;

-- name: EnqueueAnalyticsEvent :exec
INSERT INTO analytics_event_queue (event_log_id)
VALUES ($1)
ON CONFLICT (event_log_id) DO NOTHING;

-- name: DequeueAnalyticsEvents :many
WITH picked AS (
  SELECT q.id
  FROM analytics_event_queue q
  WHERE q.status = 'pending'
    AND q.available_at <= now()
  ORDER BY q.created_at ASC
  LIMIT sqlc.arg('limit')
  FOR UPDATE SKIP LOCKED
)
UPDATE analytics_event_queue q
SET status = 'processing',
    attempts = q.attempts + 1,
    updated_at = now()
FROM picked
WHERE q.id = picked.id
RETURNING q.id, q.event_log_id, q.status, q.attempts, q.last_error, q.available_at, q.processed_at, q.created_at, q.updated_at;

-- name: GetEventLogByID :one
SELECT id, user_id, event_name, entity_type, entity_id, metadata, ip_address, created_at
FROM event_logs
WHERE id = $1;

-- name: MarkAnalyticsEventProcessed :exec
UPDATE analytics_event_queue
SET status = 'processed',
    processed_at = now(),
    updated_at = now(),
    last_error = NULL
WHERE id = $1;

-- name: MarkAnalyticsEventFailed :exec
UPDATE analytics_event_queue
SET status = 'pending',
    last_error = sqlc.narg('last_error'),
    available_at = now() + sqlc.arg('retry_after')::interval,
    updated_at = now()
WHERE id = sqlc.arg('id');

-- name: UpsertAdminMetricDaily :exec
INSERT INTO admin_metrics_daily (metric_date, event_name, total_count, unique_users)
VALUES (
  sqlc.arg('metric_date'),
  sqlc.arg('event_name'),
  sqlc.arg('total_count'),
  sqlc.arg('unique_users')
)
ON CONFLICT (metric_date, event_name) DO UPDATE
SET total_count = EXCLUDED.total_count,
    unique_users = EXCLUDED.unique_users,
    updated_at = now();

-- name: AggregateEventLogDay :one
SELECT
  COUNT(1)::bigint AS total_count,
  COUNT(DISTINCT user_id)::bigint AS unique_users
FROM event_logs
WHERE event_name = sqlc.arg('event_name')
  AND created_at >= date_trunc('day', sqlc.arg('day')::timestamptz)
  AND created_at < date_trunc('day', sqlc.arg('day')::timestamptz) + interval '1 day';

-- name: ListAdminMetricsDaily :many
SELECT metric_date, event_name, total_count, unique_users, created_at, updated_at
FROM admin_metrics_daily
WHERE metric_date >= sqlc.arg('from_date')
  AND metric_date <= sqlc.arg('to_date')
ORDER BY metric_date ASC, event_name ASC;

-- name: AdminDashboardOverview :one
SELECT
  (SELECT COUNT(1)::bigint FROM users u WHERE u.deleted_at IS NULL) AS users_total,
  (SELECT COUNT(1)::bigint FROM sellers s WHERE s.deleted_at IS NULL) AS sellers_total,
  (SELECT COUNT(1)::bigint FROM products p WHERE p.deleted_at IS NULL) AS products_total,
  (SELECT COUNT(1)::bigint FROM orders) AS orders_total,
  (SELECT COUNT(1)::bigint FROM payments) AS payments_total,
  (SELECT COUNT(1)::bigint FROM shipments) AS shipments_total;

-- name: CreateAdminActivityLog :one
INSERT INTO admin_activity_logs (admin_user_id, action, entity_type, entity_id, metadata, ip_address, user_agent)
VALUES (sqlc.arg('admin_user_id'), sqlc.arg('action'), sqlc.narg('entity_type'), sqlc.narg('entity_id'), COALESCE(sqlc.narg('metadata')::jsonb, '{}'::jsonb), sqlc.narg('ip_address'), sqlc.narg('user_agent'))
RETURNING id, admin_user_id, action, entity_type, entity_id, metadata, ip_address, user_agent, created_at;
