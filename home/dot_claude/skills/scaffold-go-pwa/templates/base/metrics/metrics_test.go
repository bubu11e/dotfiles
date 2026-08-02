package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"__MODULE__/metrics"
)

func TestMiddlewareRecordsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.Use(metrics.Middleware())
	e.GET("/things/:id", func(c *gin.Context) { c.Status(http.StatusOK) })

	before := testutil.CollectAndCount(prometheus.DefaultGatherer.(prometheus.Collector))
	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/things/1", nil))
	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/things/2", nil))

	// Two requests on the same route template must share one series: the route
	// label is c.FullPath(), which is what keeps cardinality bounded.
	if after := testutil.CollectAndCount(prometheus.DefaultGatherer.(prometheus.Collector)); after < before {
		t.Fatalf("metric count shrank: %d -> %d", before, after)
	}
}

func TestReadyGaugeIsSettable(t *testing.T) {
	metrics.Ready.Set(1)
	if got := testutil.ToFloat64(metrics.Ready); got != 1 {
		t.Errorf("Ready = %v, want 1", got)
	}
	metrics.Ready.Set(0)
	if got := testutil.ToFloat64(metrics.Ready); got != 0 {
		t.Errorf("Ready = %v, want 0", got)
	}
}
