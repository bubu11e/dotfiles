// Package spa embeds the built front-end and serves it from the Go binary, so the
// whole app ships as a single self-contained artifact (ADR-0001).
package spa

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// dist is produced by `npm run build` in frontend/ (vite build.outDir points
// here) and is gitignored: a checkout has no dist/, so the Dockerfile and CI
// build the front-end before the Go build. "all:" is required to include the
// hashed asset files Vite emits under directories Go would otherwise skip.
//
//go:embed all:dist
var dist embed.FS

// Mount serves the embedded PWA from the engine's NoRoute handler: real assets
// are served directly, and any other GET falls back to index.html so the
// client-side router owns every bookmarkable path. API and operational routes are
// matched by their own handlers first, so this only catches what they do not.
func Mount(engine *gin.Engine) error {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return err
	}
	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		return err
	}
	fileServer := http.FileServer(http.FS(sub))

	engine.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if c.Request.Method != http.MethodGet || isReservedPath(p) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if _, statErr := fs.Stat(sub, strings.TrimPrefix(p, "/")); statErr == nil && p != "/" {
			// Go's mime table has no entry for .webmanifest, and a manifest served
			// as octet-stream is ignored, which silently costs installability.
			if strings.HasSuffix(p, ".webmanifest") {
				c.Header("Content-Type", "application/manifest+json")
			}
			// The worker is re-fetched on every navigation to check for an update;
			// a cached copy would pin the app to an old one.
			if p == "/sw.js" {
				c.Header("Cache-Control", "no-cache")
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
	return nil
}

func isReservedPath(p string) bool {
	if strings.HasPrefix(p, "/api/") {
		return true
	}
	// Operational endpoints have no sub-tree of their own, so match them exactly or
	// at a path boundary. A plain HasPrefix would wrongly reserve client routes that
	// merely start with the same letters (e.g. "/versions", "/liveliness").
	for _, ep := range []string{"/live", "/ready", "/health", "/version", "/metrics"} {
		if p == ep || strings.HasPrefix(p, ep+"/") {
			return true
		}
	}
	return false
}
