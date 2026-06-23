-- +goose Up
CREATE TABLE IF NOT EXISTS audit_log (
    id          BIGSERIAL PRIMARY KEY,
    model_type  TEXT NOT NULL,
    model_id    TEXT NOT NULL,
    version     INT NOT NULL,
    method      TEXT NOT NULL DEFAULT '',
    path        TEXT NOT NULL DEFAULT '',
    payload     JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  TEXT NOT NULL DEFAULT '',
    UNIQUE (model_type, model_id, version)
);

CREATE INDEX IF NOT EXISTS audit_log_model_idx ON audit_log (model_type, model_id, version);

-- +goose Down
DROP TABLE IF EXISTS audit_log;
