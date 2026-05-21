-- name: CreateAuthSession :one
INSERT INTO sessions (user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, refresh_token_hash, created_at, expires_at;

-- name: GetAuthSessionByTokenHash :one
SELECT id, user_id, refresh_token_hash, created_at, expires_at
FROM sessions
WHERE refresh_token_hash = $1;

-- name: DeleteAuthSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteAuthSessionByTokenHash :exec
DELETE FROM sessions WHERE refresh_token_hash = $1;
