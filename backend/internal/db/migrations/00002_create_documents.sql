-- +goose Up
CREATE TYPE share_mode AS ENUM ('private', 'read', 'edit');

CREATE TABLE documents (
    id         UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT       NOT NULL DEFAULT '',
    share_mode share_mode NOT NULL DEFAULT 'private',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_owner_id ON documents(owner_id);

-- +goose Down
DROP TABLE documents;
DROP TYPE share_mode;
