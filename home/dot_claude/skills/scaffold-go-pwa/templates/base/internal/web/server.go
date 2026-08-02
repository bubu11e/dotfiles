// Package web exposes the HTTP server: operational probes, Prometheus metrics,
// and build information, served with Gin.
package web

import (
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"__MODULE__/metrics"
	"__MODULE__/version"
)

// Server holds the dependencies for the HTTP handlers.
type Server struct {
	engine    *gin.Engine
	ready     func() bool
	buildInfo version.Info
}

// NewServer builds the Gin engine and mounts the standard routes:
//
//	GET /live, /health            – liveness (process up)
//	GET /ready                    – readiness (200 when ready, else 503)
//	GET /version, /api/v1/version – build information
//	GET /metrics                  – Prometheus metrics
//
// ready reports whether the service is functional; it gates /ready. A nil ready
// is treated as always-ready. Register additional routes via Engine().
func NewServer(ready func() bool) *Server {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gzip.Gzip(gzip.DefaultCompression))
	engine.Use(metrics.Middleware())

	s := &Server{engine: engine, ready: ready, buildInfo: version.Get()}

	engine.GET("/live", s.live)
	engine.GET("/health", s.live) // backward-compatible alias
	engine.GET("/ready", s.readyz)
	engine.GET("/version", s.version)
	engine.GET("/api/v1/version", s.version)
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return s
}

// Engine exposes the underlying router so callers can mount more routes before
// serving (e.g. the API, or the embedded PWA).
func (s *Server) Engine() *gin.Engine { return s.engine }

// Handler returns the underlying http.Handler.
func (s *Server) Handler() http.Handler { return s.engine }
