-- name: CreateUser :one
INSERT INTO id.users (email, password_hash, name, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM id.users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM id.users
WHERE id = $1;

-- name: UpdateUserAvatar :exec
UPDATE id.users
SET avatar = $2, avatar_content_type = $3, updated_at = NOW()
WHERE id = $1;

-- name: GetUserAvatar :one
SELECT avatar, avatar_content_type FROM id.users
WHERE id = $1;
