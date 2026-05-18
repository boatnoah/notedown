-- +goose Up
CREATE TYPE pfp_preset AS ENUM ('blue', 'green', 'red', 'yellow', 'purple', 'orange');

CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    username      TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    pfp           pfp_preset  NOT NULL DEFAULT 'blue',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
DROP TYPE pfp_preset;
