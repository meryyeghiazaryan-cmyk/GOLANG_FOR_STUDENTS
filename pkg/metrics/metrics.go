package metrics

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPRequestsTotal counts every HTTP request handled by the service.
// Labels: service, method, path (route template), status.
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests received.",
	},
	[]string{"service", "method", "path", "status"},
)

// HTTPBadRequestsTotal counts requests that ended with a 4xx status code.
// Useful for alerting on validation / client errors without filtering the main counter.
// Labels: service, method, path.
var HTTPBadRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_bad_requests_total",
		Help: "Total number of HTTP requests that resulted in a 4xx response.",
	},
	[]string{"service", "method", "path"},
)

// RequestCounter returns a Gin middleware that increments both counters after
// every request. It uses the route template for path so per-user URLs don't
// create unbounded label cardinality.
func RequestCounter(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		path := c.FullPath()
		method := c.Request.Method

		HTTPRequestsTotal.WithLabelValues(
			serviceName,
			method,
			path,
			strconv.Itoa(status),
		).Inc()

		if status >= 400 && status < 500 {
			HTTPBadRequestsTotal.WithLabelValues(serviceName, method, path).Inc()
		}
	}
}
