package validation

import (
	"fmt"
	"regexp"
	"time"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9]{4,16}$`)

// ValidateUsername checks that a username is 4-16 alphanumeric characters.
func ValidateUsername(username string) error {
	if !usernameRe.MatchString(username) {
		return fmt.Errorf("username must be 4-16 alphanumeric characters (a-zA-Z0-9)")
	}
	return nil
}

// ValidateCoordinates checks that latitude/longitude are within valid ranges.
func ValidateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}

// ValidateRadius checks that the radius is a positive number.
func ValidateRadius(radius float64) error {
	if radius <= 0 {
		return fmt.Errorf("radius must be a positive number")
	}
	return nil
}

// ParseTime parses an ISO 8601 (RFC3339) datetime string.
func ParseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid datetime format, expected ISO 8601 e.g. 2021-09-02T11:26:18+00:00")
	}
	return t, nil
}
