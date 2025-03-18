-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;

CREATE TABLE IF NOT EXISTS urls (
    id SERIAL,
    user_id TEXT,
    short_url TEXT,
    original_url TEXT NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (user_id, short_url)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls (original_url);

-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_original_url;
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd


