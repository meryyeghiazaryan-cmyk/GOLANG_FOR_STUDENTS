package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/repository"
	"github.com/training/GOLANG_FOR_STUDENTS/pkg/validation"
)

const earthRadiusKm = 6371.0

// DistanceInput holds the parameters for a distance calculation request.
type DistanceInput struct {
	Username string
	From     time.Time
	To       time.Time
}

// DistanceResult holds the calculated travel distance for a user.
type DistanceResult struct {
	Username   string    `json:"username"`
	DistanceKm float64   `json:"distance_km"`
	From       time.Time `json:"from"`
	To         time.Time `json:"to"`
}

// Service defines the business logic for location history.
type Service interface {
	GetDistance(ctx context.Context, input DistanceInput) (*DistanceResult, error)
	SaveLocation(ctx context.Context, username string, lat, lon float64, ts time.Time) error
}

type service struct {
	repo repository.Repository
}

// New creates a new location history service.
func New(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) SaveLocation(ctx context.Context, username string, lat, lon float64, ts time.Time) error {
	if err := validation.ValidateUsername(username); err != nil {
		return fmt.Errorf("validate username: %w", err)
	}
	if err := validation.ValidateCoordinates(lat, lon); err != nil {
		return fmt.Errorf("validate coordinates: %w", err)
	}

	rec := repository.LocationRecord{
		Username:   username,
		Latitude:   lat,
		Longitude:  lon,
		RecordedAt: ts,
	}
	if err := s.repo.SaveLocation(ctx, rec); err != nil {
		return fmt.Errorf("save location to history: %w", err)
	}
	return nil
}

func (s *service) GetDistance(ctx context.Context, input DistanceInput) (*DistanceResult, error) {
	if err := validation.ValidateUsername(input.Username); err != nil {
		return nil, fmt.Errorf("validate username: %w", err)
	}

	// Default time range is 1 day from the 'from' time.
	if input.To.IsZero() {
		input.To = input.From.Add(24 * time.Hour)
	}

	if input.From.After(input.To) {
		return nil, validation.NewError("'from' must not be after 'to'")
	}

	records, err := s.repo.GetLocationHistory(ctx, input.Username, input.From, input.To)
	if err != nil {
		return nil, fmt.Errorf("retrieve location history: %w", err)
	}

	totalKm := 0.0
	for i := 1; i < len(records); i++ {
		totalKm += haversine(
			records[i-1].Latitude, records[i-1].Longitude,
			records[i].Latitude, records[i].Longitude,
		)
	}

	return &DistanceResult{
		Username:   input.Username,
		DistanceKm: math.Round(totalKm*100) / 100,
		From:       input.From,
		To:         input.To,
	}, nil
}

// haversine calculates the great-circle distance in km between two points.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	lat1R := lat1 * math.Pi / 180
	lat2R := lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1R)*math.Cos(lat2R)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}
