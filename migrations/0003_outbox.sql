-- +goose Up
CREATE TABLE IF NOT EXISTS outbox_events (
    id              UUID PRIMARY KEY,
    type            TEXT NOT NULL,
    content_type    TEXT NOT NULL,
    topic           TEXT NOT NULL,
    route_key       TEXT NOT NULL DEFAULT '',
    payload         BYTEA NOT NULL,
    headers         JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'pending',
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 10,
    last_error      TEXT,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at         TIMESTAMPTZ,
    leased_at       TIMESTAMPTZ,
    lease_id        UUID
);
CREATE INDEX IF NOT EXISTS outbox_events_pending_idx ON outbox_events (status, next_attempt_at);

-- Consumer-side audit log: one row per widget.created event the consumer
-- receives. The primary key is the CloudEvents id, so at-least-once redelivery
-- de-duplicates via ON CONFLICT DO NOTHING.
CREATE TABLE IF NOT EXISTS widget_event_log (
    event_id    UUID PRIMARY KEY,
    type        TEXT NOT NULL,
    payload     BYTEA NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS widget_event_log;
DROP TABLE IF EXISTS outbox_events;
