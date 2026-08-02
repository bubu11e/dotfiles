package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"__MODULE__/metrics"
)

// live is the liveness probe: the process is up and the HTTP server responds.
// It performs no dependency checks.
func (s *Server) live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

// readyz is the readiness probe: 200 when the service is functional, else 503 so
// an orchestrator holds traffic. It also mirrors the state onto the Ready gauge.
func (s *Server) readyz(c *gin.Context) {
	if s.ready == nil || s.ready() {
		metrics.Ready.Set(1)
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
		return
	}
	metrics.Ready.Set(0)
	c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
}

// version returns build information.
func (s *Server) version(c *gin.Context) {
	c.JSON(http.StatusOK, s.buildInfo)
}
