-- name: CreateUser :one
INSERT INTO users (name, email, username, password_hash, pfp)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, email, username, pfp, created_at;

-- name: GetUserByEmail :one
SELECT id, name, email, username, password_hash, pfp, created_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, name, email, username, pfp, created_at
FROM users
WHERE id = $1;
