-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    duration BIGINT NOT NULL CHECK (duration > 0),
    description TEXT,
    user_id UUID NOT NULL,
    notify_before BIGINT CHECK (notify_before IS NULL OR notify_before >= 0)
);

CREATE INDEX IF NOT EXISTS idx_events_date ON events (date);
CREATE INDEX IF NOT EXISTS idx_events_user_id_date ON events (user_id, date);

-- +goose Down
DROP TABLE IF EXISTS events;
