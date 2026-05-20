-- name: UpsertDocument :exec
INSERT INTO documents (id, owner_id, title, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
    SET title      = EXCLUDED.title,
        updated_at = EXCLUDED.updated_at;

-- name: GetDocument :one
SELECT id, owner_id, title, created_at, updated_at
FROM documents
WHERE id = $1;

-- name: ListDocumentsByOwner :many
SELECT id, owner_id, title, created_at, updated_at
FROM documents
WHERE owner_id = $1
ORDER BY created_at DESC;
