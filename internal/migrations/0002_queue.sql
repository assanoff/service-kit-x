-- +goose Up
CREATE TABLE IF NOT EXISTS queue_tasks (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    kind       TEXT NOT NULL,
    payload    BYTEA,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    run_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempts   INT NOT NULL DEFAULT 0,
    leased_at  TIMESTAMPTZ,
    lease_id   TEXT,
    last_error TEXT,
    done_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS queue_tasks_ready_idx ON queue_tasks (done_at, run_at);

-- +goose Down
DROP TABLE IF EXISTS queue_tasks;
