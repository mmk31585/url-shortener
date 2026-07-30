-- +goose Up
CREATE TABLE IF NOT EXISTS urls
    (
        id             SERIAL PRIMARY KEY                ,
        short_code     VARCHAR(255) NOT NULL             ,
        original_url   TEXT NOT NULL                     ,
        redirect_count BIGINT NOT NULL DEFAULT 0      ,
        created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        deleted_at     TIMESTAMPTZ
    )
;
ALTER TABLE urls ADD CONSTRAINT chk_short_code_length CHECK (LENGTH(short_code) = 8);
ALTER TABLE urls ADD CONSTRAINT chk_original_url CHECK (original_url <> '');
ALTER TABLE urls ADD CONSTRAINT chk_redirect_count CHECK (redirect_count >= 0);
-- +goose Down
DROP TABLE IF EXISTS urls;