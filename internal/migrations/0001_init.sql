-- +goose Up
CREATE TABLE IF NOT EXISTS widgets (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS widgets_name_key ON widgets (name);

-- +goose Down
DROP TABLE IF EXISTS widgets;
