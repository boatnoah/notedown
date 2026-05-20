-- name: AppendOperation :exec
WITH next_v AS (
    SELECT COALESCE(MAX(version), 0) + 1 AS v
    FROM operations
    WHERE document_id = $2
)
INSERT INTO operations (id, document_id, kind, char_offset, length, text, version, created_at)
SELECT $1, $2, $3, $4, $5, $6, v, $7
FROM next_v;

-- name: ListOperations :many
SELECT id, kind, char_offset, length, text, created_at
FROM operations
WHERE document_id = $1
ORDER BY version ASC;
