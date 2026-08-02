package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// InstanceHandler serves the presentation settings the client needs before it
// has a session: the app name and whether auth runs in development mode, so the
// sign-in form can drop the password field.
type InstanceHandler struct {
	name    string
	devMode bool
}

// NewInstanceHandler wires an InstanceHandler.
func NewInstanceHandler(name string, devMode bool) *InstanceHandler {
	return &InstanceHandler{name: name, devMode: devMode}
}

// Register mounts the instance route. It is deliberately unauthenticated: the app
// name carries no secret and the client needs it before login.
func (h *InstanceHandler) Register(r gin.IRouter) {
	r.GET("/api/v1/instance", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"name": h.name, "dev_mode": h.devMode})
	})
}
