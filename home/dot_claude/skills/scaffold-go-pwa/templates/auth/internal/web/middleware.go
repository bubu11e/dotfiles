package web

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"__MODULE__/internal/auth"
	"__MODULE__/internal/store"
)

// sessionCookie is the name of the opaque session cookie.
const sessionCookie = "__NAME___session"

// contextUserKey is where RequireAuth stashes the resolved user.
const contextUserKey = "auth.user"

// RequireAuth returns middleware that resolves the session cookie to a user and
// stores it in the context. It aborts with 401 if there is no valid session.
func RequireAuth(sessions *store.SessionStore, users *store.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(sessionCookie)
		if err != nil || raw == "" {
			unauthorized(c)
			return
		}
		sess, err := sessions.GetValid(c.Request.Context(), auth.HashToken(raw))
		if errors.Is(err, store.ErrNotFound) {
			unauthorized(c)
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "session lookup failed"})
			return
		}
		user, err := users.GetByID(c.Request.Context(), sess.UserID)
		if err != nil {
			// The session points at a missing user: treat it as unauthenticated.
			unauthorized(c)
			return
		}
		c.Set(contextUserKey, user)
		c.Next()
	}
}

// CurrentUser returns the authenticated user previously stored by RequireAuth.
func CurrentUser(c *gin.Context) (*store.User, bool) {
	v, ok := c.Get(contextUserKey)
	if !ok {
		return nil, false
	}
	user, ok := v.(*store.User)
	return user, ok
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
}
