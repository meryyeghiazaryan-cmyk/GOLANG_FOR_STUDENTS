CREATE TABLE IF NOT EXISTS users_location (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(16) UNIQUE NOT NULL,
    latitude   DOUBLE PRECISION NOT NULL,
    longitude  DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index to speed up coordinate-based proximity queries.
CREATE INDEX IF NOT EXISTS idx_users_location_coords
    ON users_location (latitude, longitude);
