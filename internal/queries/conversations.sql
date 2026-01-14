-- name: GetConversationsByUserID :many
SELECT 
    conversations.id AS conversation_id,
    users.id, 
    users.last_seen, 
    users.username
FROM 
    conversations 
JOIN 
    users 
    ON users.id IN (conversations.user1, conversations.user2)
WHERE 
    -- 1. Ensure the conversation belongs to you
    (conversations.user1 = $1 OR conversations.user2 = $1)
    -- 2. CRITICAL: Filter out your own user record from the final list
    AND users.id != $1;

-- name: GetConversationByMembers :one
SELECT * FROM conversations WHERE (user1 = $1 AND user2 = $2) OR (user1 = $2 AND user2 = $1);

-- name: CreateConversation :one
INSERT INTO conversations(user1, user2) VALUES($1, $2) RETURNING *;

-- name: DeleteConversation :one
DELETE FROM conversations WHERE id = $1 RETURNING *;
