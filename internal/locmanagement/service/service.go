package service

import (
	"context"
	"fmt"
	"time"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/grpcclient"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/repository"
	"github.com/training/GOLANG_FOR_STUDENTS/pkg/validation"
)

// UpdateLocationInput holds the data for updating a user's location.
type UpdateLocationInput struct {
	Username  string
	Latitude  float64
	Longitude float64
}

// SearchInput holds parameters for a radius-based user search.
type SearchInput struct {
	Latitude  float64
	Longitude float64
	RadiusKm  float64
	Page      int
	Size      int
}

// UserDTO is the data transfer object returned to the API layer.
type UserDTO struct {
	Username  string  `json:"username"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SearchResult contains paginated search output.
type SearchResult struct {
	Users []UserDTO `json:"users"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

// Service defines the business logic for location management.
type Service interface {
	UpdateLocation(ctx context.Context, input UpdateLocationInput) error
	SearchByRadius(ctx context.Context, input SearchInput) (*SearchResult, error)
}

type service struct {
	repo       repository.Repository
	grpcClient grpcclient.Client
}

// New creates a new location management service.
func New(repo repository.Repository, grpcClient grpcclient.Client) Service {
	return &service{repo: repo, grpcClient: grpcClient}
}

func (s *service) UpdateLocation(ctx context.Context, input UpdateLocationInput) error {
	if err := validation.ValidateUsername(input.Username); err != nil {
		return fmt.Errorf("validate username: %w", err)
	}
	if err := validation.ValidateCoordinates(input.Latitude, input.Longitude); err != nil {
		return fmt.Errorf("validate coordinates: %w", err)
	}

	loc := repository.UserLocation{
		Username:  input.Username,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
	}
	if err := s.repo.Upsert(ctx, loc); err != nil {
		return fmt.Errorf("persist location: %w", err)
	}

	if err := s.grpcClient.SaveLocation(ctx, input.Username, input.Latitude, input.Longitude, time.Now()); err != nil {
		return fmt.Errorf("notify history service: %w", err)
	}

	return nil
}

func (s *service) SearchByRadius(ctx context.Context, input SearchInput) (*SearchResult, error) {
	if err := validation.ValidateCoordinates(input.Latitude, input.Longitude); err != nil {
		return nil, fmt.Errorf("validate coordinates: %w", err)
	}
	if err := validation.ValidateRadius(input.RadiusKm); err != nil {
		return nil, fmt.Errorf("validate radius: %w", err)
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Size < 1 || input.Size > 100 {
		input.Size = 10
	}

	locs, total, err := s.repo.FindByRadius(ctx, input.Latitude, input.Longitude, input.RadiusKm, input.Page, input.Size)
	if err != nil {
		return nil, fmt.Errorf("find users by radius: %w", err)
	}

	users := make([]UserDTO, 0, len(locs))
	for _, l := range locs {
		users = append(users, UserDTO{
			Username:  l.Username,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		})
	}

	return &SearchResult{
		Users: users,
		Total: total,
		Page:  input.Page,
		Size:  input.Size,
	}, nil
}
