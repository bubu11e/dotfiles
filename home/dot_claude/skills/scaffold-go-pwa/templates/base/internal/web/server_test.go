package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"__MODULE__/internal/web"
)

func do(t *testing.T, s *web.Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestLivenessIsIndependentOfReadiness(t *testing.T) {
	s := web.NewServer(func() bool { return false })
	for _, path := range []string{"/live", "/health"} {
		if got := do(t, s, path).Code; got != http.StatusOK {
			t.Errorf("GET %s = %d, want 200 even when not ready", path, got)
		}
	}
}

func TestReadinessReflectsTheGate(t *testing.T) {
	ready := false
	s := web.NewServer(func() bool { return ready })

	if got := do(t, s, "/ready").Code; got != http.StatusServiceUnavailable {
		t.Errorf("GET /ready = %d, want 503 while not ready", got)
	}
	ready = true
	if got := do(t, s, "/ready").Code; got != http.StatusOK {
		t.Errorf("GET /ready = %d, want 200 once ready", got)
	}
}

func TestNilReadyMeansAlwaysReady(t *testing.T) {
	if got := do(t, web.NewServer(nil), "/ready").Code; got != http.StatusOK {
		t.Errorf("GET /ready with a nil gate = %d, want 200", got)
	}
}

func TestVersionEndpointsReturnBuildInfo(t *testing.T) {
	s := web.NewServer(nil)
	for _, path := range []string{"/version", "/api/v1/version"} {
		var info struct {
			Version   string `json:"version"`
			GoVersion string `json:"go_version"`
		}
		if err := json.Unmarshal(do(t, s, path).Body.Bytes(), &info); err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if info.Version == "" || info.GoVersion == "" {
			t.Errorf("GET %s returned an incomplete payload: %+v", path, info)
		}
	}
}

func TestMetricsEndpointExposesOurMetrics(t *testing.T) {
	s := web.NewServer(nil)
	// A served request first, so the HTTP middleware has something to report:
	// a counter vec exposes no series until a label combination is observed.
	do(t, s, "/live")

	body := do(t, s, "/metrics").Body.String()
	for _, name := range []string{"__NAME___http_requests_total", "__NAME___ready"} {
		if !strings.Contains(body, name) {
			t.Errorf("/metrics does not expose %s", name)
		}
	}
	// The route label must be the Gin template, not the concrete path, or
	// cardinality grows without bound.
	if !strings.Contains(body, `route="/live"`) {
		t.Error("/metrics does not label requests with the route template")
	}
}
