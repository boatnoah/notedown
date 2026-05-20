-- name: CreateUser :one
INSERT INTO users (name, email, username, password_hash, pfp)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, email, username, pfp, created_at;
