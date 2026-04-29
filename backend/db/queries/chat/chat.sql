-- name: GetOrCreateConversation :one
INSERT INTO chat_conversations (buyer_id, seller_id)
VALUES ($1, $2)
ON CONFLICT (buyer_id, seller_id)
DO UPDATE SET buyer_id = EXCLUDED.buyer_id
RETURNING id, buyer_id, seller_id, last_message_at, created_at;

-- name: ListBuyerConversations :many
SELECT c.id, c.buyer_id, c.seller_id, c.last_message_at, c.created_at,
       s.shop_name, s.slug, s.logo_url,
       (
         SELECT count(*)::bigint
         FROM chat_messages m
         JOIN sellers s2 ON s2.id = c.seller_id
         WHERE m.conversation_id = c.id
           AND m.sender_user_id = s2.user_id
           AND m.read_at IS NULL
       ) AS unread_count
FROM chat_conversations c
JOIN sellers s ON s.id = c.seller_id
WHERE c.buyer_id = $1
ORDER BY COALESCE(c.last_message_at, c.created_at) DESC
LIMIT $2 OFFSET $3;

-- name: ListSellerConversations :many
SELECT c.id, c.buyer_id, c.seller_id, c.last_message_at, c.created_at,
       COALESCE(up.display_name, '') AS buyer_name, u.email AS buyer_email,
       (
         SELECT count(*)::bigint
         FROM chat_messages m
         WHERE m.conversation_id = c.id
           AND m.sender_user_id = c.buyer_id
           AND m.read_at IS NULL
       ) AS unread_count
FROM chat_conversations c
JOIN users u ON u.id = c.buyer_id
LEFT JOIN user_profiles up ON up.user_id = u.id
WHERE c.seller_id = $1
ORDER BY COALESCE(c.last_message_at, c.created_at) DESC
LIMIT $2 OFFSET $3;

-- name: CountBuyerConversations :one
SELECT count(*)::bigint FROM chat_conversations WHERE buyer_id = $1;

-- name: CountSellerConversations :one
SELECT count(*)::bigint FROM chat_conversations WHERE seller_id = $1;

-- name: ListMessages :many
SELECT id, conversation_id, sender_user_id, message, attachment_url, read_at, created_at
FROM chat_messages
WHERE conversation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateMessage :one
INSERT INTO chat_messages (conversation_id, sender_user_id, message, attachment_url)
VALUES ($1, $2, $3, $4)
RETURNING id, conversation_id, sender_user_id, message, attachment_url, read_at, created_at;

-- name: TouchConversationLastMessageAt :exec
UPDATE chat_conversations
SET last_message_at = now()
WHERE id = $1;

-- name: MarkConversationRead :exec
UPDATE chat_messages
SET read_at = now()
WHERE conversation_id = $1
  AND sender_user_id <> $2
  AND read_at IS NULL;

-- name: DeleteMessageForUser :execrows
DELETE FROM chat_messages
WHERE chat_messages.id = $1
  AND conversation_id IN (
    SELECT c.id
    FROM chat_conversations c
    JOIN sellers s ON s.id = c.seller_id
    WHERE c.id = $2
      AND (c.buyer_id = $3 OR s.user_id = $3)
  );

-- name: DeleteConversationForUser :execrows
DELETE FROM chat_conversations
WHERE chat_conversations.id = $1
  AND chat_conversations.id IN (
    SELECT c.id
    FROM chat_conversations c
    JOIN sellers s ON s.id = c.seller_id
    WHERE c.id = $1
      AND (c.buyer_id = $2 OR s.user_id = $2)
  );

-- name: GetConversationWithSellerUser :one
SELECT c.id, c.buyer_id, c.seller_id, s.user_id AS seller_user_id
FROM chat_conversations c
JOIN sellers s ON s.id = c.seller_id
WHERE c.id = $1;

-- name: GetConversationForUser :one
SELECT c.id, c.buyer_id, c.seller_id, s.user_id AS seller_user_id
FROM chat_conversations c
JOIN sellers s ON s.id = c.seller_id
WHERE c.id = sqlc.arg('id')
  AND (c.buyer_id = sqlc.arg('user_id') OR s.user_id = sqlc.arg('user_id'));

