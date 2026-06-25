-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    price      BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS products;
