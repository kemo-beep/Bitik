-- name: ListRecentSearchQueriesForUser :many
SELECT id, user_id, session_id, query, filters, created_at
FROM search_recent_queries
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListRecentSearchQueriesForSession :many
SELECT id, user_id, session_id, query, filters, created_at
FROM search_recent_queries
WHERE session_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ClearRecentSearchQueriesForUser :exec
DELETE FROM search_recent_queries
WHERE user_id = $1;

-- name: ClearRecentSearchQueriesForSession :exec
DELETE FROM search_recent_queries
WHERE session_id = $1;

-- name: ListTrendingSearchQueries :many
SELECT query, count(*)::bigint AS clicks
FROM search_click_events
WHERE created_at >= now() - sqlc.arg('window')::interval
GROUP BY query
ORDER BY clicks DESC, max(created_at) DESC
LIMIT sqlc.arg('limit');

-- name: ListIndexableProducts :many
SELECT p.id,
       p.seller_id,
       p.category_id,
       p.brand_id,
       p.name,
       p.slug,
       p.description,
       p.min_price_cents,
       p.max_price_cents,
       p.currency,
       p.total_sold,
       p.rating,
       p.review_count,
       p.published_at,
       p.updated_at
FROM products p
JOIN sellers s ON s.id = p.seller_id
WHERE p.deleted_at IS NULL
  AND p.status = 'active'
  AND s.deleted_at IS NULL
  AND s.status = 'active'
  AND (
    sqlc.narg('cursor_updated_at')::timestamptz IS NULL
    OR p.updated_at > sqlc.narg('cursor_updated_at')::timestamptz
    OR (p.updated_at = sqlc.narg('cursor_updated_at')::timestamptz AND p.id > sqlc.narg('cursor_id')::uuid)
  )
ORDER BY p.updated_at ASC, p.id ASC
LIMIT sqlc.arg('limit');

