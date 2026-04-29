-- name: AdminListPages :many
SELECT id, slug, title, body, status, published_at, created_by, updated_by, created_at, updated_at
FROM cms_pages
WHERE (sqlc.narg('status')::cms_status IS NULL OR status = sqlc.narg('status')::cms_status)
  AND (sqlc.narg('q')::text IS NULL OR slug ILIKE ('%' || sqlc.narg('q')::text || '%') OR title ILIKE ('%' || sqlc.narg('q')::text || '%'))
ORDER BY updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminGetPageByID :one
SELECT id, slug, title, body, status, published_at, created_by, updated_by, created_at, updated_at
FROM cms_pages
WHERE id = $1;

-- name: AdminCreatePage :one
INSERT INTO cms_pages (slug, title, body, status, published_at, created_by, updated_by)
VALUES (
  sqlc.arg('slug'),
  sqlc.arg('title'),
  sqlc.arg('body'),
  COALESCE(sqlc.narg('status')::cms_status, 'draft'),
  sqlc.narg('published_at')::timestamptz,
  sqlc.narg('actor_user_id')::uuid,
  sqlc.narg('actor_user_id')::uuid
)
RETURNING id, slug, title, body, status, published_at, created_by, updated_by, created_at, updated_at;

-- name: AdminUpdatePage :one
UPDATE cms_pages
SET slug = COALESCE(sqlc.narg('slug')::text, slug),
    title = COALESCE(sqlc.narg('title')::text, title),
    body = COALESCE(sqlc.narg('body')::text, body),
    status = COALESCE(sqlc.narg('status')::cms_status, status),
    published_at = COALESCE(sqlc.narg('published_at')::timestamptz, published_at),
    updated_by = COALESCE(sqlc.narg('actor_user_id')::uuid, updated_by)
WHERE id = sqlc.arg('id')
RETURNING id, slug, title, body, status, published_at, created_by, updated_by, created_at, updated_at;

-- name: AdminDeletePage :exec
DELETE FROM cms_pages
WHERE id = $1;

-- name: AdminListBanners :many
SELECT id, title, image_url, link_url, placement, sort_order, status, starts_at, ends_at, created_at, updated_at
FROM cms_banners
WHERE (sqlc.narg('status')::cms_status IS NULL OR status = sqlc.narg('status')::cms_status)
  AND (sqlc.narg('placement')::text IS NULL OR placement = sqlc.narg('placement')::text)
ORDER BY placement ASC, sort_order ASC, updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminGetBannerByID :one
SELECT id, title, image_url, link_url, placement, sort_order, status, starts_at, ends_at, created_at, updated_at
FROM cms_banners
WHERE id = $1;

-- name: AdminCreateBanner :one
INSERT INTO cms_banners (title, image_url, link_url, placement, sort_order, status, starts_at, ends_at)
VALUES (
  sqlc.arg('title'),
  sqlc.arg('image_url'),
  sqlc.narg('link_url')::text,
  COALESCE(sqlc.narg('placement')::text, 'home'),
  COALESCE(sqlc.narg('sort_order')::int, 0),
  COALESCE(sqlc.narg('status')::cms_status, 'draft'),
  sqlc.narg('starts_at')::timestamptz,
  sqlc.narg('ends_at')::timestamptz
)
RETURNING id, title, image_url, link_url, placement, sort_order, status, starts_at, ends_at, created_at, updated_at;

-- name: AdminUpdateBanner :one
UPDATE cms_banners
SET title = COALESCE(sqlc.narg('title')::text, title),
    image_url = COALESCE(sqlc.narg('image_url')::text, image_url),
    link_url = COALESCE(sqlc.narg('link_url')::text, link_url),
    placement = COALESCE(sqlc.narg('placement')::text, placement),
    sort_order = COALESCE(sqlc.narg('sort_order')::int, sort_order),
    status = COALESCE(sqlc.narg('status')::cms_status, status),
    starts_at = COALESCE(sqlc.narg('starts_at')::timestamptz, starts_at),
    ends_at = COALESCE(sqlc.narg('ends_at')::timestamptz, ends_at)
WHERE id = sqlc.arg('id')
RETURNING id, title, image_url, link_url, placement, sort_order, status, starts_at, ends_at, created_at, updated_at;

-- name: AdminDeleteBanner :exec
DELETE FROM cms_banners
WHERE id = $1;

-- name: AdminListFaqs :many
SELECT id, question, answer, category, sort_order, status, created_at, updated_at
FROM cms_faqs
WHERE (sqlc.narg('status')::cms_status IS NULL OR status = sqlc.narg('status')::cms_status)
  AND (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category')::text)
ORDER BY COALESCE(category, '') ASC, sort_order ASC, updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminGetFaqByID :one
SELECT id, question, answer, category, sort_order, status, created_at, updated_at
FROM cms_faqs
WHERE id = $1;

-- name: AdminCreateFaq :one
INSERT INTO cms_faqs (question, answer, category, sort_order, status)
VALUES (
  sqlc.arg('question'),
  sqlc.arg('answer'),
  sqlc.narg('category')::text,
  COALESCE(sqlc.narg('sort_order')::int, 0),
  COALESCE(sqlc.narg('status')::cms_status, 'draft')
)
RETURNING id, question, answer, category, sort_order, status, created_at, updated_at;

-- name: AdminUpdateFaq :one
UPDATE cms_faqs
SET question = COALESCE(sqlc.narg('question')::text, question),
    answer = COALESCE(sqlc.narg('answer')::text, answer),
    category = COALESCE(sqlc.narg('category')::text, category),
    sort_order = COALESCE(sqlc.narg('sort_order')::int, sort_order),
    status = COALESCE(sqlc.narg('status')::cms_status, status)
WHERE id = sqlc.arg('id')
RETURNING id, question, answer, category, sort_order, status, created_at, updated_at;

-- name: AdminDeleteFaq :exec
DELETE FROM cms_faqs
WHERE id = $1;

-- name: AdminListAnnouncements :many
SELECT id, title, body, audience, status, starts_at, ends_at, created_at, updated_at
FROM cms_announcements
WHERE (sqlc.narg('status')::cms_status IS NULL OR status = sqlc.narg('status')::cms_status)
  AND (sqlc.narg('audience')::text IS NULL OR audience = sqlc.narg('audience')::text)
ORDER BY COALESCE(starts_at, created_at) DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminGetAnnouncementByID :one
SELECT id, title, body, audience, status, starts_at, ends_at, created_at, updated_at
FROM cms_announcements
WHERE id = $1;

-- name: AdminCreateAnnouncement :one
INSERT INTO cms_announcements (title, body, audience, status, starts_at, ends_at)
VALUES (
  sqlc.arg('title'),
  sqlc.arg('body'),
  COALESCE(sqlc.narg('audience')::text, 'all'),
  COALESCE(sqlc.narg('status')::cms_status, 'draft'),
  sqlc.narg('starts_at')::timestamptz,
  sqlc.narg('ends_at')::timestamptz
)
RETURNING id, title, body, audience, status, starts_at, ends_at, created_at, updated_at;

-- name: AdminUpdateAnnouncement :one
UPDATE cms_announcements
SET title = COALESCE(sqlc.narg('title')::text, title),
    body = COALESCE(sqlc.narg('body')::text, body),
    audience = COALESCE(sqlc.narg('audience')::text, audience),
    status = COALESCE(sqlc.narg('status')::cms_status, status),
    starts_at = COALESCE(sqlc.narg('starts_at')::timestamptz, starts_at),
    ends_at = COALESCE(sqlc.narg('ends_at')::timestamptz, ends_at)
WHERE id = sqlc.arg('id')
RETURNING id, title, body, audience, status, starts_at, ends_at, created_at, updated_at;

-- name: AdminDeleteAnnouncement :exec
DELETE FROM cms_announcements
WHERE id = $1;

