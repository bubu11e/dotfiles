package spa

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsReservedPath(t *testing.T) {
	reserved := []string{"/api/", "/api/version", "/live", "/ready", "/health", "/version", "/metrics"}
	for _, p := range reserved {
		if !isReservedPath(p) {
			t.Errorf("isReservedPath(%q) = false, want true", p)
		}
	}
	// Client routes that merely share a prefix with a probe must stay unreserved.
	notReserved := []string{"/", "/login", "/settings", "/versions", "/liveliness", "/assets/app.js"}
	for _, p := range notReserved {
		if isReservedPath(p) {
			t.Errorf("isReservedPath(%q) = true, want false", p)
		}
	}
}

func newEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	e := gin.New()
	if err := Mount(e); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	return e
}

func do(t *testing.T, e *gin.Engine, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

func TestMountServesShellForUnknownGet(t *testing.T) {
	// A client-side route the server does not know must fall back to index.html so
	// the router can take over.
	w := do(t, newEngine(t), http.MethodGet, "/some/client/route")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestMountServesTheManifestWithItsOwnType(t *testing.T) {
	w := do(t, newEngine(t), http.MethodGet, "/manifest.webmanifest")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/manifest+json") {
		t.Errorf("content-type = %q, want application/manifest+json", ct)
	}
}

func TestMountReservedPathReturnsJSON404(t *testing.T) {
	w := do(t, newEngine(t), http.MethodGet, "/api/does-not-exist")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("content-type = %q, want application/json", ct)
	}
}

func TestMountRejectsNonGet(t *testing.T) {
	if w := do(t, newEngine(t), http.MethodPost, "/some/client/route"); w.Code != http.StatusNotFound {
		t.Errorf("POST to an unknown path = %d, want 404", w.Code)
	}
}
