package web_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"__MODULE__/internal/storage"
	"__MODULE__/internal/store"
	"__MODULE__/internal/web"
	"__MODULE__/migrations"
)

type client struct {
	t       *testing.T
	handler http.Handler
	cookies []*http.Cookie
}

func newClient(t *testing.T, devMode bool) *client {
	t.Helper()
	db := openDB(t)
	srv := web.NewServer(nil)
	users := store.NewUserStore(db)
	sessions := store.NewSessionStore(db)
	web.NewAuthHandler(users, sessions, time.Hour, false, devMode, nil).Register(srv.Engine())
	return &client{t: t, handler: srv.Handler()}
}

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "__NAME__.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := storage.Migrate(db, migrations.FS); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

// do sends a request carrying whatever cookies previous responses set, so a
// sign-in is visible to the calls that follow it.
func (c *client) do(method, path string, body any) *httptest.ResponseRecorder {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, req)
	if set := w.Result().Cookies(); len(set) > 0 {
		c.cookies = set
	}
	return w
}

func TestDevModeRegistersWithAPseudoAndNoPassword(t *testing.T) {
	c := newClient(t, true)

	w := c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{"email": "Ada L"})
	if w.Code != http.StatusCreated {
		t.Fatalf("register = %d (%s), want 201", w.Code, w.Body.String())
	}

	// Registration signs the user straight in, so /me works immediately: no
	// verification step exists in development mode.
	if got := c.do(http.MethodGet, "/api/v1/me", nil); got.Code != http.StatusOK {
		t.Fatalf("me after dev register = %d (%s), want 200", got.Code, got.Body.String())
	}
}

func TestDevModeSignsInWithThePseudoAlone(t *testing.T) {
	c := newClient(t, true)
	c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{"email": "Ada L"})
	c.do(http.MethodPost, "/api/v1/auth/logout", nil)

	if w := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "ada l"}); w.Code != http.StatusOK {
		t.Fatalf("dev login without a password = %d (%s), want 200", w.Code, w.Body.String())
	}
}

func TestProdModeRequiresAnEmailAndAPassword(t *testing.T) {
	c := newClient(t, false)

	if w := c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "Ada L", "display_name": "Ada", "password": "longenough",
	}); w.Code != http.StatusBadRequest {
		t.Errorf("register with a pseudo = %d, want 400 outside dev mode", w.Code)
	}
	if w := c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "ada@example.com", "display_name": "Ada", "password": "short",
	}); w.Code != http.StatusBadRequest {
		t.Errorf("register with a short password = %d, want 400", w.Code)
	}
}

func TestProdModeHoldsSignInUntilVerified(t *testing.T) {
	c := newClient(t, false)

	w := c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "ada@example.com", "display_name": "Ada", "password": "longenough",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register = %d (%s), want 201", w.Code, w.Body.String())
	}
	// No session is issued before verification.
	if got := c.do(http.MethodGet, "/api/v1/me", nil); got.Code != http.StatusUnauthorized {
		t.Errorf("me right after prod register = %d, want 401", got.Code)
	}
	if got := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "ada@example.com", "password": "longenough",
	}); got.Code != http.StatusForbidden {
		t.Errorf("login before verification = %d, want 403", got.Code)
	}
}

func TestLoginRejectsABadPassword(t *testing.T) {
	c := newClient(t, true)
	c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "ada", "password": "longenough",
	})
	c.do(http.MethodPost, "/api/v1/auth/logout", nil)

	w := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "ada", "password": "wrongpassword",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("login with a wrong password = %d, want 401", w.Code)
	}
	// The message must not distinguish a missing account from a wrong password.
	var body struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error != "invalid credentials" {
		t.Errorf("error = %q, want the generic message", body.Error)
	}
}

func TestLogoutInvalidatesTheSession(t *testing.T) {
	c := newClient(t, true)
	c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{"email": "ada"})

	if w := c.do(http.MethodPost, "/api/v1/auth/logout", nil); w.Code != http.StatusNoContent {
		t.Fatalf("logout = %d, want 204", w.Code)
	}
	if w := c.do(http.MethodGet, "/api/v1/me", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("me after logout = %d, want 401", w.Code)
	}
}

func TestMeRequiresASession(t *testing.T) {
	c := newClient(t, true)
	if w := c.do(http.MethodGet, "/api/v1/me", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("me without a cookie = %d, want 401", w.Code)
	}
}

func TestVerifyRejectsAnUnknownToken(t *testing.T) {
	c := newClient(t, false)
	if w := c.do(http.MethodGet, "/api/v1/auth/verify?token=nope", nil); w.Code != http.StatusBadRequest {
		t.Errorf("verify with an unknown token = %d, want 400", w.Code)
	}
}

func TestInstanceExposesDevModeToTheClient(t *testing.T) {
	srv := web.NewServer(nil)
	web.NewInstanceHandler("__TITLE__", true).Register(srv.Engine())

	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/instance", nil))

	var body struct {
		Name    string `json:"name"`
		DevMode bool   `json:"dev_mode"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Name != "__TITLE__" || !body.DevMode {
		t.Errorf("instance = %+v, want the app name and dev_mode true", body)
	}
}

// newLoggingClient is the production-mode client plus a capture of the logger, so
// a test can read the verification link the server would have emailed.
func newLoggingClient(t *testing.T) (*client, *bytes.Buffer) {
	t.Helper()
	db := openDB(t)
	var logs bytes.Buffer
	srv := web.NewServer(nil)
	web.NewAuthHandler(
		store.NewUserStore(db), store.NewSessionStore(db), time.Hour, false, false,
		slog.New(slog.NewJSONHandler(&logs, nil)),
	).Register(srv.Engine())
	return &client{t: t, handler: srv.Handler()}, &logs
}

// verifyURL pulls the verification link out of the captured log lines.
func verifyURL(t *testing.T, logs *bytes.Buffer) string {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var entry struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err == nil && entry.URL != "" {
			return entry.URL
		}
	}
	t.Fatal("no verification link was logged")
	return ""
}

func TestProdModeVerifyThenSignIn(t *testing.T) {
	c, logs := newLoggingClient(t)

	if w := c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "ada@example.com", "display_name": "Ada", "password": "longenough",
	}); w.Code != http.StatusCreated {
		t.Fatalf("register = %d (%s), want 201", w.Code, w.Body.String())
	}

	if w := c.do(http.MethodGet, verifyURL(t, logs), nil); w.Code != http.StatusOK {
		t.Fatalf("verify = %d (%s), want 200", w.Code, w.Body.String())
	}
	if w := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "ADA@example.com", "password": "longenough",
	}); w.Code != http.StatusOK {
		t.Fatalf("login after verification = %d (%s), want 200", w.Code, w.Body.String())
	}
	if w := c.do(http.MethodGet, "/api/v1/me", nil); w.Code != http.StatusOK {
		t.Errorf("me after login = %d, want 200", w.Code)
	}
}

func TestRegisterRejectsADuplicate(t *testing.T) {
	dev := newClient(t, true)
	dev.do(http.MethodPost, "/api/v1/auth/register", map[string]string{"email": "ada"})
	if w := dev.do(http.MethodPost, "/api/v1/auth/register", map[string]string{"email": "Ada"}); w.Code != http.StatusConflict {
		t.Errorf("duplicate dev register = %d, want 409", w.Code)
	}

	prod := newClient(t, false)
	body := map[string]string{"email": "ada@example.com", "display_name": "Ada", "password": "longenough"}
	prod.do(http.MethodPost, "/api/v1/auth/register", body)
	if w := prod.do(http.MethodPost, "/api/v1/auth/register", body); w.Code != http.StatusConflict {
		t.Errorf("duplicate prod register = %d, want 409", w.Code)
	}
}

func TestDevModeStillNeedsAnIdentity(t *testing.T) {
	c := newClient(t, true)
	if w := c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{}); w.Code != http.StatusBadRequest {
		t.Errorf("register with nothing = %d, want 400", w.Code)
	}
	if w := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{}); w.Code != http.StatusUnauthorized {
		t.Errorf("login with nothing = %d, want 401", w.Code)
	}
}

func TestDevModeAcceptsARealEmailAsTheIdentity(t *testing.T) {
	c := newClient(t, true)
	// A value that already looks like an email is kept as-is rather than slugified
	// into <pseudo>@dev.local, so a dev account can carry a real address.
	if w := c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "ada@example.com", "display_name": "Ada",
	}); w.Code != http.StatusCreated {
		t.Fatalf("register = %d (%s), want 201", w.Code, w.Body.String())
	}
	c.do(http.MethodPost, "/api/v1/auth/logout", nil)
	if w := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "ada@example.com"}); w.Code != http.StatusOK {
		t.Errorf("login = %d, want 200", w.Code)
	}
}

func TestMalformedBodyIsARequestError(t *testing.T) {
	c := newClient(t, true)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("login with a malformed body = %d, want 400", w.Code)
	}
}

func TestAnUnknownCookieIsNotASession(t *testing.T) {
	c := newClient(t, true)
	c.cookies = []*http.Cookie{{Name: "__NAME___session", Value: "not-a-real-token"}}
	if w := c.do(http.MethodGet, "/api/v1/me", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("me with a forged cookie = %d, want 401", w.Code)
	}
}

func TestVerifyWithoutATokenIsARequestError(t *testing.T) {
	c := newClient(t, false)
	if w := c.do(http.MethodGet, "/api/v1/auth/verify", nil); w.Code != http.StatusBadRequest {
		t.Errorf("verify with no token = %d, want 400", w.Code)
	}
}
