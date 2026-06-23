-- +goose Up
CREATE TABLE IF NOT EXISTS translations (
    model_name        VARCHAR(100) NOT NULL,
    column_name       VARCHAR(100) NOT NULL,
    key_id            VARCHAR(255) NOT NULL,
    language_id       VARCHAR(10)  NOT NULL,
    translation_value TEXT,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (model_name, column_name, key_id, language_id)
);

CREATE INDEX IF NOT EXISTS idx_translation_lookup
    ON translations (model_name, key_id, language_id);

-- +goose Down
DROP TABLE IF EXISTS translations;
