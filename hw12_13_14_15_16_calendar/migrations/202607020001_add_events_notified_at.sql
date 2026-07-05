-- +goose Up
ALTER TABLE events
    ADD COLUMN IF NOT EXISTS notified_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_events_notify_due
    ON events (date, notified_at)
    WHERE notify_before IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_events_notify_due;

ALTER TABLE events
    DROP COLUMN IF EXISTS notified_at;
