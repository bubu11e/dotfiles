package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"__MODULE__/internal/store"
)

// sessionCookie is the name of the opaque session cookie.
const sessionCookie = "__NAME___session"

// SessionCookies owns the session cookie's lifetime and attributes, shared by the
// handler that mints a session and the middleware that renews one, so both write
// the cookie the same way.
type SessionCookies struct {
	ttl    time.Duration
	secure bool
}

// NewSessionCookies returns the cookie policy for sessions living for ttl. secure
// should be true when the app is served over HTTPS.
func NewSessionCookies(ttl time.Duration, secure bool) SessionCookies {
	return SessionCookies{ttl: ttl, secure: secure}
}

// TTL is the session lifetime, used as both the cookie MaxAge and the server-side
// expiry so the two never drift apart.
func (s SessionCookies) TTL() time.Duration { return s.ttl }

func (s SessionCookies) set(c *gin.Context, raw string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookie,
		Value:    raw,
		Path:     "/",
		MaxAge:   int(s.ttl.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s SessionCookies) clear(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: s.secure, SameSite: http.SameSiteLaxMode,
	})
}

// dueForRenewal reports whether sess has burned more than half its lifetime and
// should be extended. Half-life keeps the write to roughly one per user per
// ttl/2 while ensuring an active user never reaches the fixed expiry -- the
// previous behaviour, which signed people out mid-use with no way for the running
// SPA to notice.
func (s SessionCookies) dueForRenewal(sess *store.Session) bool {
	return time.Until(sess.ExpiresAt) < s.ttl/2
}
