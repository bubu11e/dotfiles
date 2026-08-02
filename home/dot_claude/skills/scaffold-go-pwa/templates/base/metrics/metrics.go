// Package metrics defines all Prometheus metrics for __NAME__ and a Gin
// middleware that records HTTP request count and latency. Metrics register with
// the default registry on first import; serve them via promhttp at /metrics.
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "__NAME___http_requests_total",
		Help: "Total HTTP requests by route, method, and status code.",
	}, []string{"route", "method", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "__NAME___http_request_duration_seconds",
		Help:    "HTTP request latency in seconds by route and method.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "method"})

	// BuildInfo carries version/commit/goversion as labels on a constant 1.
	BuildInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "__NAME___build_info",
		Help: "Build information exposed as a constant 1; read the labels.",
	}, []string{"version", "commit", "goversion"})

	// Ready is 1 when the service is functional, 0 otherwise.
	Ready = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "__NAME___ready",
		Help: "Readiness: 1 when the service is functional, 0 otherwise.",
	})

	// TODO: add domain metrics here, e.g.
	// WorkerRuns = promauto.NewCounterVec(prometheus.CounterOpts{
	//     Name: "__NAME___worker_runs_total",
	//     Help: "Background worker runs by result.",
	// }, []string{"result"})
)

// Middleware records __NAME___http_requests_total and
// __NAME___http_request_duration_seconds for every request. The route label is
// the Gin route template (c.FullPath()) so cardinality stays bounded; requests
// that match no route are bucketed as "unmatched".
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		method := c.Request.Method
		httpRequests.WithLabelValues(route, method, strconv.Itoa(c.Writer.Status())).Inc()
		httpDuration.WithLabelValues(route, method).Observe(time.Since(start).Seconds())
	}
}
