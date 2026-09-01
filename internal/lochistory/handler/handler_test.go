package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/handler"
	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/service"
)

type mockService struct {
	distanceResult *service.DistanceResult
	distanceErr    error
}

func (m *mockService) GetDistance(_ context.Context, _ service.DistanceInput) (*service.DistanceResult, error) {
	return m.distanceResult, m.distanceErr
}

func (m *mockService) SaveLocation(_ context.Context, _ string, _, _ float64, _ time.Time) error {
	return nil
}

func newRouter(h *handler.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/users/:username/distance", h.GetDistance)
	return r
}

func TestGetDistance(t *testing.T) {
	logger := zap.NewNop()
	base := time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC)

	okResult := &service.DistanceResult{
		Username:   "john1",
		DistanceKm: 445.87,
		From:       base,
		To:         base.Add(24 * time.Hour),
	}

	tests := []struct {
		name           string
		username       string
		query          string
		svcResult      *service.DistanceResult
		svcErr         error
		expectedStatus int
	}{
		{
			name:           "success with from and to",
			username:       "john1",
			query:          "?from=2021-09-01T00:00:00Z&to=2021-09-02T00:00:00Z",
			svcResult:      okResult,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with only from",
			username:       "john1",
			query:          "?from=2021-09-01T00:00:00Z",
			svcResult:      okResult,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success with offset timezone (url-encoded +)",
			username:       "john1",
			query:          "?from=2021-09-01T00:00:00%2B00:00",
			svcResult:      okResult,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid username too short",
			username:       "ab",
			query:          "?from=2021-09-01T00:00:00Z",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing from param",
			username:       "john1",
			query:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid from format",
			username:       "john1",
			query:          "?from=2021-09-01",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid to format",
			username:       "john1",
			query:          "?from=2021-09-01T00:00:00Z&to=2021-09-02",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			username:       "john1",
			query:          "?from=2021-09-01T00:00:00Z",
			svcErr:         errors.New("'from' must not be after 'to'"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{distanceResult: tt.svcResult, distanceErr: tt.svcErr}
			h := handler.New(svc, logger)
			router := newRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.username+"/distance"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
