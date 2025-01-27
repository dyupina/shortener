-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL,
    short_url TEXT PRIMARY KEY,
    full_url TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_full_url ON urls (full_url);

-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_full_url;
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd


