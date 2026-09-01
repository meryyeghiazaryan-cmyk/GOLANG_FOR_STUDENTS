package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/repository"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/service"
)

// --- mocks ---

type mockRepo struct {
	upsertErr  error
	findResult []repository.UserLocation
	findTotal  int
	findErr    error
}

func (m *mockRepo) Upsert(_ context.Context, _ repository.UserLocation) error {
	return m.upsertErr
}

func (m *mockRepo) FindByRadius(_ context.Context, _, _, _ float64, _, _ int) ([]repository.UserLocation, int, error) {
	return m.findResult, m.findTotal, m.findErr
}

type mockGRPC struct {
	saveErr error
}

func (m *mockGRPC) SaveLocation(_ context.Context, _ string, _, _ float64, _ time.Time) error {
	return m.saveErr
}

func (m *mockGRPC) Close() error { return nil }

// --- tests ---

func TestUpdateLocation(t *testing.T) {
	tests := []struct {
		name      string
		input     service.UpdateLocationInput
		repoErr   error
		grpcErr   error
		expectErr bool
	}{
		{
			name:  "success",
			input: service.UpdateLocationInput{Username: "john1", Latitude: 35.12314, Longitude: 27.64532},
		},
		{
			name:      "username too short",
			input:     service.UpdateLocationInput{Username: "ab", Latitude: 35.12314, Longitude: 27.64532},
			expectErr: true,
		},
		{
			name:      "username too long",
			input:     service.UpdateLocationInput{Username: "abcdefghijklmnopq", Latitude: 35.12314, Longitude: 27.64532},
			expectErr: true,
		},
		{
			name:      "latitude out of range",
			input:     service.UpdateLocationInput{Username: "john1", Latitude: 91.0, Longitude: 27.64532},
			expectErr: true,
		},
		{
			name:      "longitude out of range",
			input:     service.UpdateLocationInput{Username: "john1", Latitude: 35.12314, Longitude: 181.0},
			expectErr: true,
		},
		{
			name:      "repo error",
			input:     service.UpdateLocationInput{Username: "john1", Latitude: 35.12314, Longitude: 27.64532},
			repoErr:   errors.New("connection refused"),
			expectErr: true,
		},
		{
			name:      "grpc error",
			input:     service.UpdateLocationInput{Username: "john1", Latitude: 35.12314, Longitude: 27.64532},
			grpcErr:   errors.New("history service unavailable"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{upsertErr: tt.repoErr}
			grpc := &mockGRPC{saveErr: tt.grpcErr}
			svc := service.New(repo, grpc)

			err := svc.UpdateLocation(context.Background(), tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSearchByRadius(t *testing.T) {
	tests := []struct {
		name        string
		input       service.SearchInput
		repoResult  []repository.UserLocation
		repoTotal   int
		repoErr     error
		expectErr   bool
		expectCount int
	}{
		{
			name:  "success with results",
			input: service.SearchInput{Latitude: 35.12, Longitude: 27.64, RadiusKm: 100, Page: 1, Size: 10},
			repoResult: []repository.UserLocation{
				{Username: "john1", Latitude: 35.12, Longitude: 27.64},
				{Username: "jane2", Latitude: 35.11, Longitude: 27.65},
			},
			repoTotal:   2,
			expectCount: 2,
		},
		{
			name:        "no results",
			input:       service.SearchInput{Latitude: 35.12, Longitude: 27.64, RadiusKm: 1, Page: 1, Size: 10},
			repoResult:  nil,
			repoTotal:   0,
			expectCount: 0,
		},
		{
			name:      "invalid latitude",
			input:     service.SearchInput{Latitude: 100.0, Longitude: 27.64, RadiusKm: 100},
			expectErr: true,
		},
		{
			name:      "invalid longitude",
			input:     service.SearchInput{Latitude: 35.12, Longitude: 200.0, RadiusKm: 100},
			expectErr: true,
		},
		{
			name:      "negative radius",
			input:     service.SearchInput{Latitude: 35.12, Longitude: 27.64, RadiusKm: -5},
			expectErr: true,
		},
		{
			name:      "zero radius",
			input:     service.SearchInput{Latitude: 35.12, Longitude: 27.64, RadiusKm: 0},
			expectErr: true,
		},
		{
			name:    "repo error",
			input:   service.SearchInput{Latitude: 35.12, Longitude: 27.64, RadiusKm: 100, Page: 1, Size: 10},
			repoErr: errors.New("db timeout"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{
				findResult: tt.repoResult,
				findTotal:  tt.repoTotal,
				findErr:    tt.repoErr,
			}
			svc := service.New(repo, &mockGRPC{})

			result, err := svc.SearchByRadius(context.Background(), tt.input)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.repoTotal, result.Total)
			assert.Len(t, result.Users, tt.expectCount)
		})
	}
}
