-- name: AddContact :one
INSERT INTO contacts ( 
    user_id, 
    contact_user_id,  
    nickname
) VALUES (
    $1,
    $2,
    $3
) RETURNING *;

-- name: GetContactsByUserID :many
SELECT 
    u.id, 
    u.username, 
    u.email, 
    u.created_at, 
    u.last_seen
FROM users u
INNER JOIN contacts c ON u.id = c.contact_user_id
WHERE c.user_id = $1
ORDER BY u.username ASC;

-- name: DeleteContact :one
DELETE FROM contacts WHERE contact_user_id = $1 AND user_id = $2 RETURNING *;

