-- +goose Up
ALTER TABLE operations ADD CONSTRAINT operations_document_id_version_key UNIQUE (document_id, version);

-- +goose Down
ALTER TABLE operations DROP CONSTRAINT operations_document_id_version_key;
