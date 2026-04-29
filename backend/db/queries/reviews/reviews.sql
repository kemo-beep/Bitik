-- name: GetReviewForUser :one
SELECT id, product_id, order_item_id, user_id, rating, title, body, is_verified_purchase, is_hidden, deleted_at, created_at, updated_at, seller_reply, seller_reply_at, seller_reply_by
FROM product_reviews
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: CreateReview :one
INSERT INTO product_reviews (product_id, order_item_id, user_id, rating, title, body, is_verified_purchase)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, product_id, order_item_id, user_id, rating, title, body, is_verified_purchase, is_hidden, deleted_at, created_at, updated_at, seller_reply, seller_reply_at, seller_reply_by;

-- name: UpdateReviewForUser :one
UPDATE product_reviews
SET rating = COALESCE(sqlc.narg('rating')::int, rating),
    title = COALESCE(sqlc.narg('title')::text, title),
    body = COALESCE(sqlc.narg('body')::text, body)
WHERE id = sqlc.arg('id')
  AND user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL
RETURNING id, product_id, order_item_id, user_id, rating, title, body, is_verified_purchase, is_hidden, deleted_at, created_at, updated_at, seller_reply, seller_reply_at, seller_reply_by;

-- name: SoftDeleteReviewForUser :exec
UPDATE product_reviews
SET deleted_at = now()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: ListReviewImages :many
SELECT id, review_id, url, sort_order
FROM product_review_images
WHERE review_id = $1
ORDER BY sort_order ASC, id ASC;

-- name: AddReviewImage :one
INSERT INTO product_review_images (review_id, url, sort_order)
VALUES ($1, $2, $3)
RETURNING id, review_id, url, sort_order;

-- name: DeleteReviewImage :exec
DELETE FROM product_review_images
WHERE id = $1 AND review_id = $2;

-- name: UpsertReviewVote :one
INSERT INTO review_votes (review_id, user_id, vote)
VALUES ($1, $2, $3)
ON CONFLICT (review_id, user_id)
DO UPDATE SET vote = EXCLUDED.vote
RETURNING review_id, user_id, vote, created_at;

-- name: DeleteReviewVote :exec
DELETE FROM review_votes WHERE review_id = $1 AND user_id = $2;

-- name: CreateReviewReport :one
INSERT INTO review_reports (review_id, reporter_user_id, reason)
VALUES ($1, $2, $3)
RETURNING id, review_id, reporter_user_id, reason, status, resolved_by, resolved_at, created_at;

-- name: ListOpenReviewReports :many
SELECT id, review_id, reporter_user_id, reason, status, resolved_by, resolved_at, created_at
FROM review_reports
WHERE status = 'open'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ResolveReviewReport :one
UPDATE review_reports
SET status = 'resolved',
    resolved_by = $2,
    resolved_at = now()
WHERE id = $1
RETURNING id, review_id, reporter_user_id, reason, status, resolved_by, resolved_at, created_at;

-- name: HideReviewAdmin :one
UPDATE product_reviews
SET is_hidden = $2
WHERE id = $1
RETURNING id, product_id, order_item_id, user_id, rating, title, body, is_verified_purchase, is_hidden, deleted_at, created_at, updated_at, seller_reply, seller_reply_at, seller_reply_by;

-- name: SoftDeleteReviewAdmin :exec
UPDATE product_reviews
SET deleted_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SetSellerReply :one
UPDATE product_reviews pr
SET seller_reply = $2,
    seller_reply_at = now(),
    seller_reply_by = $3
FROM products p
WHERE pr.id = $1
  AND pr.product_id = p.id
  AND p.seller_id = $4
  AND pr.deleted_at IS NULL
RETURNING pr.id, pr.product_id, pr.order_item_id, pr.user_id, pr.rating, pr.title, pr.body, pr.is_verified_purchase, pr.is_hidden, pr.deleted_at, pr.created_at, pr.updated_at, pr.seller_reply, pr.seller_reply_at, pr.seller_reply_by;

-- name: GetOrderItemForReviewVerification :one
SELECT oi.id, oi.order_id, oi.product_id
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE oi.id = $1
  AND o.user_id = $2;

