-- +goose Up
ALTER TABLE sessions ADD CONSTRAINT sessions_refresh_token_hash_key UNIQUE (refresh_token_hash);

-- +goose Down
ALTER TABLE sessions DROP CONSTRAINT sessions_refresh_token_hash_key;
