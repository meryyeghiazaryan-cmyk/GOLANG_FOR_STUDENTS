//go:build integration
// +build integration

// Package integration contains end-to-end tests that require running
// PostgreSQL databases and both microservices.
// Run with: go test -tags integration ./test/integration/...
package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var locMgmtBaseURL string

func TestMain(m *testing.M) {
	locMgmtBaseURL = getEnvOrDefault("LOC_MGMT_URL", "http://localhost:8080")
	os.Exit(m.Run())
}

func TestUpdateAndSearchIntegration(t *testing.T) {
	dbURL := getEnvOrDefault(
		"LOC_MGMT_DB_URL",
		"postgres://postgres:password@localhost:5432/loc_management?sslmode=disable",
	)
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up test data.
	_, err = db.Exec("DELETE FROM users_location WHERE username LIKE 'itest%'")
	require.NoError(t, err)

	t.Run("update location", func(t *testing.T) {
		body := map[string]interface{}{"latitude": 35.12314, "longitude": 27.64532}
		resp := doRequest(t, http.MethodPut, locMgmtBaseURL+"/api/v1/users/itest001/location", body)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("search finds updated user", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/users?lat=35.12314&lon=27.64532&radius=10", locMgmtBaseURL)
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.GreaterOrEqual(t, int(result["total"].(float64)), 1)
	})

	t.Run("search outside radius returns no results", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/users?lat=0&lon=0&radius=1", locMgmtBaseURL)
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var result map[string]interface{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		assert.Equal(t, float64(0), result["total"])
	})

	t.Run("invalid username returns 400", func(t *testing.T) {
		body := map[string]interface{}{"latitude": 35.12314, "longitude": 27.64532}
		resp := doRequest(t, http.MethodPut, locMgmtBaseURL+"/api/v1/users/ab/location", body)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func doRequest(t *testing.T, method, url string, body interface{}) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
