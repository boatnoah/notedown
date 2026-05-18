-- +goose Up
CREATE TABLE operations (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    kind        TEXT        NOT NULL CHECK (kind IN ('insert', 'delete')),
    offset      INTEGER     NOT NULL,
    length      INTEGER     NOT NULL DEFAULT 0,
    text        TEXT        NOT NULL DEFAULT '',
    version     BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_operations_document_id ON operations(document_id);
CREATE INDEX idx_operations_document_version ON operations(document_id, version);

-- +goose Down
DROP TABLE operations;
