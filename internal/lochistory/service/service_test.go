package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/repository"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/service"
)

type mockRepo struct {
	saveErr     error
	historyData []repository.LocationRecord
	historyErr  error
}

func (m *mockRepo) SaveLocation(_ context.Context, _ repository.LocationRecord) error {
	return m.saveErr
}

func (m *mockRepo) GetLocationHistory(_ context.Context, _ string, _, _ time.Time) ([]repository.LocationRecord, error) {
	return m.historyData, m.historyErr
}

var base = time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC)

func TestGetDistance(t *testing.T) {
	tests := []struct {
		name            string
		input           service.DistanceInput
		history         []repository.LocationRecord
		historyErr      error
		expectErr       bool
		expectDistMin   float64
		expectDistMax   float64
	}{
		{
			name:  "single point - no movement",
			input: service.DistanceInput{Username: "john1", From: base, To: base.Add(24 * time.Hour)},
			history: []repository.LocationRecord{
				{Username: "john1", Latitude: 35.12314, Longitude: 27.64532, RecordedAt: base},
			},
			expectDistMin: 0,
			expectDistMax: 0,
		},
		{
			name:  "two points about 445km apart",
			input: service.DistanceInput{Username: "john1", From: base, To: base.Add(24 * time.Hour)},
			history: []repository.LocationRecord{
				{Username: "john1", Latitude: 35.12314, Longitude: 27.64532, RecordedAt: base},
				{Username: "john1", Latitude: 39.12355, Longitude: 27.64538, RecordedAt: base.Add(1 * time.Hour)},
			},
			expectDistMin: 440,
			expectDistMax: 450,
		},
		{
			name:  "round trip is roughly double one way",
			input: service.DistanceInput{Username: "john1", From: base, To: base.Add(24 * time.Hour)},
			history: []repository.LocationRecord{
				{Username: "john1", Latitude: 35.12314, Longitude: 27.64532, RecordedAt: base},
				{Username: "john1", Latitude: 39.12355, Longitude: 27.64538, RecordedAt: base.Add(1 * time.Hour)},
				{Username: "john1", Latitude: 35.12314, Longitude: 27.64532, RecordedAt: base.Add(2 * time.Hour)},
			},
			expectDistMin: 880,
			expectDistMax: 900,
		},
		{
			name:  "no records - zero distance",
			input: service.DistanceInput{Username: "john1", From: base, To: base.Add(24 * time.Hour)},
			history: nil,
			expectDistMin: 0,
			expectDistMax: 0,
		},
		{
			name:      "invalid username",
			input:     service.DistanceInput{Username: "ab", From: base},
			expectErr: true,
		},
		{
			name:      "from after to",
			input:     service.DistanceInput{Username: "john1", From: base.Add(24 * time.Hour), To: base},
			expectErr: true,
		},
		{
			name:       "repo error",
			input:      service.DistanceInput{Username: "john1", From: base, To: base.Add(24 * time.Hour)},
			historyErr: errors.New("db error"),
			expectErr:  true,
		},
		{
			name:  "default to param when zero",
			input: service.DistanceInput{Username: "john1", From: base},
			history: []repository.LocationRecord{
				{Username: "john1", Latitude: 35.12314, Longitude: 27.64532, RecordedAt: base},
			},
			expectDistMin: 0,
			expectDistMax: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{historyData: tt.history, historyErr: tt.historyErr}
			svc := service.New(repo)

			result, err := svc.GetDistance(context.Background(), tt.input)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.GreaterOrEqual(t, result.DistanceKm, tt.expectDistMin)
			assert.LessOrEqual(t, result.DistanceKm, tt.expectDistMax)
		})
	}
}

func TestSaveLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepo{}
		svc := service.New(repo)
		err := svc.SaveLocation(context.Background(), "john1", 35.12, 27.64, base)
		require.NoError(t, err)
	})

	t.Run("invalid username", func(t *testing.T) {
		repo := &mockRepo{}
		svc := service.New(repo)
		err := svc.SaveLocation(context.Background(), "ab", 35.12, 27.64, base)
		require.Error(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockRepo{saveErr: errors.New("write failed")}
		svc := service.New(repo)
		err := svc.SaveLocation(context.Background(), "john1", 35.12, 27.64, base)
		require.Error(t, err)
	})
}
