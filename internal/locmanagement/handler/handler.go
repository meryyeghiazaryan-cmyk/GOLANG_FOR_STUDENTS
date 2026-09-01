package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/training/GOLANG_FOR_STUDENTS/internal/locmanagement/service"
	"github.com/training/GOLANG_FOR_STUDENTS/pkg/validation"
)

// Handler holds the HTTP handlers for the location management service.
type Handler struct {
	svc    service.Service
	logger *zap.Logger
}

// New creates a new Handler.
func New(svc service.Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

type updateLocationRequest struct {
	Latitude  *float64 `json:"latitude"  binding:"required"`
	Longitude *float64 `json:"longitude" binding:"required"`
}

// UpdateLocation handles PUT /api/v1/users/:username/location
func (h *Handler) UpdateLocation(c *gin.Context) {
	username := c.Param("username")
	if err := validation.ValidateUsername(username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req updateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	input := service.UpdateLocationInput{
		Username:  username,
		Latitude:  *req.Latitude,
		Longitude: *req.Longitude,
	}
	if err := h.svc.UpdateLocation(c.Request.Context(), input); err != nil {
		h.logger.Error("update location", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "location updated"})
}

// SearchByRadius handles GET /api/v1/users?lat=&lon=&radius=&page=&size=
func (h *Handler) SearchByRadius(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat must be a valid number"})
		return
	}
	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lon must be a valid number"})
		return
	}
	radius, err := strconv.ParseFloat(c.Query("radius"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "radius must be a valid number"})
		return
	}

	page := 1
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	size := 10
	if s, err := strconv.Atoi(c.Query("size")); err == nil && s > 0 {
		size = s
	}

	result, err := h.svc.SearchByRadius(c.Request.Context(), service.SearchInput{
		Latitude:  lat,
		Longitude: lon,
		RadiusKm:  radius,
		Page:      page,
		Size:      size,
	})
	if err != nil {
		h.logger.Warn("search by radius", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
