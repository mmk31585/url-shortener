-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_short_code
    ON urls(short_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_urls_created_at
    ON urls(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_urls_short_code;
DROP INDEX IF EXISTS idx_urls_created_at;