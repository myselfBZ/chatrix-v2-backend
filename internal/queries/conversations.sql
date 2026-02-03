-- name: GetConversationsByUserID :many
SELECT 
    c.id AS conversation_id,
    u.id, 
    u.last_seen, 
    u.username,
    (
        SELECT COUNT(m.id) 
        FROM messages m 
        WHERE m.is_read = FALSE 
          AND m.conversation_id = c.id
          AND m.sender_id != $1 
    ) AS unread_msg_count
FROM 
    conversations c
JOIN 
    users u 
    ON u.id IN (c.user1, c.user2)
WHERE 
    (c.user1 = $1 OR c.user2 = $1)
    AND u.id != $1;


-- name: GetConversationByMembers :one
SELECT * FROM conversations WHERE (user1 = $1 AND user2 = $2) OR (user1 = $2 AND user2 = $1);

-- name: CreateConversation :one
INSERT INTO conversations(user1, user2) VALUES($1, $2) RETURNING *;

-- name: DeleteConversation :one
DELETE FROM conversations WHERE id = $1 RETURNING *;
