package web_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"__MODULE__/internal/auth"
	"__MODULE__/internal/store"
	"__MODULE__/internal/web"
)

// testSessionTTL is the lifetime the renewal tests build their server with; a
// session is extended once less than half of it remains.
const testSessionTTL = time.Hour

// newRenewalClient returns a client for /api/v1/me plus the session store, so a
// test can plant a session with an arbitrary remaining lifetime.
func newRenewalClient(t *testing.T) (*client, *store.SessionStore, int64) {
	t.Helper()
	db := openDB(t)
	users, sessions := store.NewUserStore(db), store.NewSessionStore(db)
	user, err := users.Create(context.Background(), "a@b.c", "", "A")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	srv := web.NewServer(nil)
	web.NewAuthHandler(users, sessions, testSessionTTL, false, true, nil).Register(srv.Engine())
	return &client{t: t, handler: srv.Handler()}, sessions, user.ID
}

// plantSession stores a session expiring in remaining and hands the client its
// cookie.
func plantSession(t *testing.T, c *client, sessions *store.SessionStore, userID int64, remaining time.Duration) string {
	t.Helper()
	raw, hash, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if _, err := sessions.Create(context.Background(), hash, userID, remaining); err != nil {
		t.Fatalf("Create session: %v", err)
	}
	c.cookies = []*http.Cookie{{Name: "__NAME___session", Value: raw}}
	return raw
}

func TestSessionRenewedPastHalfLife(t *testing.T) {
	c, sessions, userID := newRenewalClient(t)
	raw := plantSession(t, c, sessions, userID, 5*time.Minute)

	rec := c.do(http.MethodGet, "/api/v1/me", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("me = %d %s", rec.Code, rec.Body)
	}

	sess, err := sessions.GetValid(context.Background(), auth.HashToken(raw))
	if err != nil {
		t.Fatalf("GetValid: %v", err)
	}
	if remaining := time.Until(sess.ExpiresAt); remaining < testSessionTTL-time.Minute {
		t.Errorf("remaining after renewal = %v, want ~%v", remaining, testSessionTTL)
	}

	var refreshed *http.Cookie
	for _, got := range rec.Result().Cookies() {
		if got.Name == "__NAME___session" {
			refreshed = got
		}
	}
	if refreshed == nil {
		t.Fatal("no refreshed session cookie")
	}
	if refreshed.Value != raw {
		t.Error("cookie value changed; want the same opaque token")
	}
	if refreshed.MaxAge != int(testSessionTTL.Seconds()) {
		t.Errorf("cookie MaxAge = %d, want %d", refreshed.MaxAge, int(testSessionTTL.Seconds()))
	}
}

func TestSessionNotRenewedWhileFresh(t *testing.T) {
	c, sessions, userID := newRenewalClient(t)
	plantSession(t, c, sessions, userID, testSessionTTL)

	rec := c.do(http.MethodGet, "/api/v1/me", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("me = %d %s", rec.Code, rec.Body)
	}
	for _, got := range rec.Result().Cookies() {
		if got.Name == "__NAME___session" {
			t.Errorf("fresh session re-issued the cookie (MaxAge %d); want no write", got.MaxAge)
		}
	}
}

func TestExpiredSessionStillUnauthorized(t *testing.T) {
	c, sessions, userID := newRenewalClient(t)
	plantSession(t, c, sessions, userID, -time.Minute)

	if rec := c.do(http.MethodGet, "/api/v1/me", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("me on expired session = %d, want 401", rec.Code)
	}
}
