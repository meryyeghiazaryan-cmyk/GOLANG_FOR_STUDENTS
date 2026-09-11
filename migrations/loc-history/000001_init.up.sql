CREATE TABLE IF NOT EXISTS location_history (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(16) NOT NULL,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL
);

-- Index for efficient time-range queries per user.
CREATE INDEX IF NOT EXISTS idx_location_history_username_time
    ON location_history (username, recorded_at);
