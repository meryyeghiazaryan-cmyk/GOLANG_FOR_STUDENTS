package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// LocationRecord represents a single location event stored in history.
type LocationRecord struct {
	Username   string
	Latitude   float64
	Longitude  float64
	RecordedAt time.Time
}

// Repository defines the storage operations for location history.
type Repository interface {
	SaveLocation(ctx context.Context, record LocationRecord) error
	GetLocationHistory(ctx context.Context, username string, from, to time.Time) ([]LocationRecord, error)
}

type postgresRepo struct {
	db *sql.DB
}

// New creates a new PostgreSQL-backed history repository.
func New(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

// SaveLocation inserts a new location event into the history table.
func (r *postgresRepo) SaveLocation(ctx context.Context, record LocationRecord) error {
	const q = `
		INSERT INTO location_history (username, latitude, longitude, recorded_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := r.db.ExecContext(ctx, q, record.Username, record.Latitude, record.Longitude, record.RecordedAt); err != nil {
		return fmt.Errorf("insert location history: %w", err)
	}
	return nil
}

// GetLocationHistory returns all location records for a user in the given time range,
// ordered by recorded_at ascending.
func (r *postgresRepo) GetLocationHistory(ctx context.Context, username string, from, to time.Time) ([]LocationRecord, error) {
	const q = `
		SELECT username, latitude, longitude, recorded_at
		FROM location_history
		WHERE username = $1 AND recorded_at >= $2 AND recorded_at <= $3
		ORDER BY recorded_at ASC
	`
	rows, err := r.db.QueryContext(ctx, q, username, from, to)
	if err != nil {
		return nil, fmt.Errorf("query location history: %w", err)
	}
	defer rows.Close()

	var records []LocationRecord
	for rows.Next() {
		var rec LocationRecord
		if err := rows.Scan(&rec.Username, &rec.Latitude, &rec.Longitude, &rec.RecordedAt); err != nil {
			return nil, fmt.Errorf("scan location record: %w", err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate location history: %w", err)
	}

	return records, nil
}
