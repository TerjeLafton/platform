-- name: CreateUser :one
INSERT INTO id.users (email, password_hash, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM id.users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM id.users
WHERE id = $1;
