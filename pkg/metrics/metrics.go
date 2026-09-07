package metrics

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPRequestsTotal counts every HTTP request handled by the service.
// Labels: service (which binary), method (GET/PUT/…), path (route template), status (200/400/…).
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests received.",
	},
	[]string{"service", "method", "path", "status"},
)

// RequestCounter returns a Gin middleware that increments HTTPRequestsTotal
// after every request, using the route template (e.g. /api/v1/users/:username/location)
// rather than the raw URL so high-cardinality user values don't explode the metric.
func RequestCounter(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		HTTPRequestsTotal.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()),
		).Inc()
	}
}
