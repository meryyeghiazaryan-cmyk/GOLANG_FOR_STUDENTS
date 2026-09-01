package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/lochistory/service"
	"github.com/training/GOLANG_FOR_STUDENTS/pkg/validation"
)

// Handler holds the HTTP handlers for the location history service.
type Handler struct {
	svc    service.Service
	logger *zap.Logger
}

// New creates a new Handler.
func New(svc service.Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// GetDistance handles GET /api/v1/users/:username/distance?from=&to=
func (h *Handler) GetDistance(c *gin.Context) {
	username := c.Param("username")
	if err := validation.ValidateUsername(username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fromStr := c.Query("from")
	if fromStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'from' query parameter is required"})
		return
	}

	from, err := validation.ParseTime(fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var to time.Time
	if toStr := c.Query("to"); toStr != "" {
		to, err = validation.ParseTime(toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	result, err := h.svc.GetDistance(c.Request.Context(), service.DistanceInput{
		Username: username,
		From:     from,
		To:       to,
	})
	if err != nil {
		h.logger.Warn("get distance", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
