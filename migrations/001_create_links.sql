-- +goose Up
CREATE TABLE links (
    short_code   TEXT        PRIMARY KEY,
    original_url TEXT        NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE links;
