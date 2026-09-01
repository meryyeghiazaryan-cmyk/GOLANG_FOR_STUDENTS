package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// UserLocation holds the current location of a user.
type UserLocation struct {
	Username  string
	Latitude  float64
	Longitude float64
}

// Repository defines the storage operations for location management.
type Repository interface {
	Upsert(ctx context.Context, loc UserLocation) error
	FindByRadius(ctx context.Context, lat, lon, radiusKm float64, page, size int) ([]UserLocation, int, error)
}

type postgresRepo struct {
	db *sql.DB
}

// New creates a new PostgreSQL-backed repository.
func New(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

// Upsert inserts or updates a user's current location.
func (r *postgresRepo) Upsert(ctx context.Context, loc UserLocation) error {
	const q = `
		INSERT INTO users_location (username, latitude, longitude, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (username) DO UPDATE
		  SET latitude = EXCLUDED.latitude,
		      longitude = EXCLUDED.longitude,
		      updated_at = NOW()
	`
	if _, err := r.db.ExecContext(ctx, q, loc.Username, loc.Latitude, loc.Longitude); err != nil {
		return fmt.Errorf("upsert user location: %w", err)
	}
	return nil
}

// FindByRadius returns users within radiusKm of (lat, lon) with pagination.
// Uses the Haversine formula to calculate distances.
func (r *postgresRepo) FindByRadius(ctx context.Context, lat, lon, radiusKm float64, page, size int) ([]UserLocation, int, error) {
	const haversine = `
		6371 * 2 * ASIN(SQRT(
			POWER(SIN((RADIANS($1) - RADIANS(latitude)) / 2), 2) +
			COS(RADIANS($1)) * COS(RADIANS(latitude)) *
			POWER(SIN((RADIANS($2) - RADIANS(longitude)) / 2), 2)
		))
	`
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM users_location WHERE (%s) <= $3", haversine)

	var total int
	if err := r.db.QueryRowContext(ctx, countQ, lat, lon, radiusKm).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users by radius: %w", err)
	}

	offset := (page - 1) * size
	dataQ := fmt.Sprintf(`
		SELECT username, latitude, longitude
		FROM users_location
		WHERE (%s) <= $3
		ORDER BY username
		LIMIT $4 OFFSET $5
	`, haversine)

	rows, err := r.db.QueryContext(ctx, dataQ, lat, lon, radiusKm, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query users by radius: %w", err)
	}
	defer rows.Close()

	var locs []UserLocation
	for rows.Next() {
		var l UserLocation
		if err := rows.Scan(&l.Username, &l.Latitude, &l.Longitude); err != nil {
			return nil, 0, fmt.Errorf("scan user location row: %w", err)
		}
		locs = append(locs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users result: %w", err)
	}

	return locs, total, nil
}
