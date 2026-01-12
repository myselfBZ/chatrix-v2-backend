-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (
    username,
    email,
    password_hash
) 
VALUES ( 
    $1,
    $2,
    $3
) RETURNING *;

		-- ListUsers(ctx context.Context) ([]queries.User, error)
		-- SearchUsers(ctx context.Context, username string) ([]queries.User, error)
		-- UpdateUserLastSeen(ctx context.Context, id uuid.UUID) error

-- name: ListUsers :many
SELECT * FROM users;

-- name: SearchUsers :many
SELECT *
FROM users 
WHERE username ILIKE $1 || '%'
LIMIT 20;

-- name: UpdateUserLastSeen :exec
UPDATE users
    SET last_seen = CURRENT_TIMESTAMP
    WHERE id = $1;
