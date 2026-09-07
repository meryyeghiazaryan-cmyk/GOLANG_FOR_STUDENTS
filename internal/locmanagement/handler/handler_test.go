package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/handler"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/service"
	"github.com/training/GOLANG_FOR_STUDENTS/pkg/validation"
)

type mockService struct {
	updateErr    error
	searchResult *service.SearchResult
	searchErr    error
}

func (m *mockService) UpdateLocation(_ context.Context, _ service.UpdateLocationInput) error {
	return m.updateErr
}

func (m *mockService) SearchByRadius(_ context.Context, _ service.SearchInput) (*service.SearchResult, error) {
	return m.searchResult, m.searchErr
}

func newRouter(h *handler.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/api/v1/users/:username/location", h.UpdateLocation)
	r.GET("/api/v1/users", h.SearchByRadius)
	return r
}

func TestUpdateLocation(t *testing.T) {
	logger := zap.NewNop()

	lat, lon := 35.12314, 27.64532

	tests := []struct {
		name           string
		username       string
		body           interface{}
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           "success",
			username:       "john1",
			body:           map[string]interface{}{"latitude": lat, "longitude": lon},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "username too short",
			username:       "ab",
			body:           map[string]interface{}{"latitude": lat, "longitude": lon},
			serviceErr:     validation.NewError("username must be 4-16 alphanumeric characters (a-zA-Z0-9)"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "username with special chars",
			username:       "john!",
			body:           map[string]interface{}{"latitude": lat, "longitude": lon},
			serviceErr:     validation.NewError("username must be 4-16 alphanumeric characters (a-zA-Z0-9)"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing latitude",
			username:       "john1",
			body:           map[string]interface{}{"longitude": lon},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing longitude",
			username:       "john1",
			body:           map[string]interface{}{"latitude": lat},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			username:       "john1",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "latitude out of range returns 400 not 500",
			username:       "john1",
			body:           map[string]interface{}{"latitude": 90.12314, "longitude": lon},
			serviceErr:     validation.NewError("latitude must be between -90 and 90"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "longitude out of range returns 400 not 500",
			username:       "john1",
			body:           map[string]interface{}{"latitude": lat, "longitude": 181.0},
			serviceErr:     validation.NewError("longitude must be between -180 and 180"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			username:       "john1",
			body:           map[string]interface{}{"latitude": lat, "longitude": lon},
			serviceErr:     errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{updateErr: tt.serviceErr}
			h := handler.New(svc, logger)
			router := newRouter(h)

			body, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+tt.username+"/location", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSearchByRadius(t *testing.T) {
	logger := zap.NewNop()

	okResult := &service.SearchResult{
		Users: []service.UserDTO{{Username: "john1", Latitude: 35.12, Longitude: 27.64}},
		Total: 1,
		Page:  1,
		Size:  10,
	}

	tests := []struct {
		name           string
		query          string
		serviceResult  *service.SearchResult
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           "success",
			query:          "?lat=35.12&lon=27.64&radius=100&page=1&size=10",
			serviceResult:  okResult,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing lat",
			query:          "?lon=27.64&radius=100",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing lon",
			query:          "?lat=35.12&radius=100",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing radius",
			query:          "?lat=35.12&lon=27.64",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-numeric lat",
			query:          "?lat=abc&lon=27.64&radius=100",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service validation error",
			query:          "?lat=35.12&lon=27.64&radius=100",
			serviceErr:     validation.NewError("radius must be a positive number"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service persistence error",
			query:          "?lat=35.12&lon=27.64&radius=100",
			serviceErr:     errors.New("db timeout"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "default pagination",
			query:          "?lat=35.12&lon=27.64&radius=100",
			serviceResult:  okResult,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{searchResult: tt.serviceResult, searchErr: tt.serviceErr}
			h := handler.New(svc, logger)
			router := newRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users"+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
