-- name: CreateMessage :one
INSERT INTO messages (
    sender_id, 
    conversation_id, 
    content
) VALUES (
    $1, 
    (
        SELECT id FROM conversations 
        WHERE (user1 = $1 AND user2 = $2) 
           OR (user1 = $2 AND user2 = $1)
        LIMIT 1
    ), 
    $3
)
RETURNING *;

-- name: GetMessagesByConversationID :many
-- Ordered by created_at DESC for easy pagination
SELECT *
FROM messages
WHERE conversation_id = $1 
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: MarkMessagesAsRead :many
-- Marks all messages sent FROM the contact TO the current user as read
UPDATE messages
SET is_read = TRUE
WHERE sender_id = $1 
  AND conversation_id = $2 
  AND is_read = FALSE RETURNING id;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1;
