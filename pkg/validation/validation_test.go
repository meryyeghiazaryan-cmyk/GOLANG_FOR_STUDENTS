package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/training/GOLANG_FOR_STUDENTS/pkg/validation"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		expectErr bool
	}{
		{"valid lowercase", "john", false},
		{"valid mixed case", "John123", false},
		{"valid max length", "abcdefghijklmnop", false},
		{"exactly 4 chars", "abcd", false},
		{"exactly 16 chars", "abcdefghijklmnop", false},
		{"too short - 2 chars", "ab", true},
		{"too long - 17 chars", "abcdefghijklmnopq", true},
		{"special chars", "john!", true},
		{"has space", "john doe", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateUsername(tt.username)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		lat, lon  float64
		expectErr bool
	}{
		{"valid coordinates", 35.12314, 27.64532, false},
		{"zero zero", 0, 0, false},
		{"max latitude", 90, 0, false},
		{"min latitude", -90, 0, false},
		{"max longitude", 0, 180, false},
		{"min longitude", 0, -180, false},
		{"lat too high", 90.00001, 0, true},
		{"lat too low", -90.00001, 0, true},
		{"lon too high", 0, 180.00001, true},
		{"lon too low", 0, -180.00001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateCoordinates(tt.lat, tt.lon)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRadius(t *testing.T) {
	tests := []struct {
		name      string
		radius    float64
		expectErr bool
	}{
		{"positive radius", 100, false},
		{"small radius", 0.001, false},
		{"zero radius", 0, true},
		{"negative radius", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateRadius(tt.radius)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"valid UTC", "2021-09-02T11:26:18Z", false},
		{"valid +00:00 offset", "2021-09-02T11:26:18+00:00", false},
		{"valid negative offset", "2021-09-02T11:26:18-05:00", false},
		{"date only no time", "2021-09-02", true},
		{"not a date", "not-a-date", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validation.ParseTime(tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
