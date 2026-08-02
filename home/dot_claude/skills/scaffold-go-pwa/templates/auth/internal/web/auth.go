package web

import (
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"__MODULE__/internal/auth"
	"__MODULE__/internal/store"
)

// minPasswordLen is the minimum password length enforced when a password is
// required (always in production; only when supplied in development mode).
const minPasswordLen = 8

// AuthHandler serves registration, login, logout, email verification, and the
// current-user endpoint for local email + password accounts. In development mode
// the password is optional and accounts are auto-verified (ADR-0004).
type AuthHandler struct {
	users        *store.UserStore
	sessions     *store.SessionStore
	sessionTTL   time.Duration
	secureCookie bool
	devMode      bool
	logger       *slog.Logger
}

// NewAuthHandler constructs an AuthHandler. secureCookie should be true when the
// app is served over HTTPS so the session cookie carries the Secure attribute.
// When devMode is true, registration skips the password requirement and the
// verification email; otherwise a verification link is issued (logged via logger).
func NewAuthHandler(users *store.UserStore, sessions *store.SessionStore, ttl time.Duration, secureCookie, devMode bool, logger *slog.Logger) *AuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthHandler{
		users: users, sessions: sessions, sessionTTL: ttl,
		secureCookie: secureCookie, devMode: devMode, logger: logger,
	}
}

// Register mounts the auth routes onto r.
func (h *AuthHandler) Register(r gin.IRouter) {
	r.POST("/api/v1/auth/register", h.register)
	r.POST("/api/v1/auth/login", h.login)
	r.POST("/api/v1/auth/logout", h.logout)
	r.GET("/api/v1/auth/verify", h.verify)

	authed := r.Group("/api/v1")
	authed.Use(RequireAuth(h.sessions, h.users))
	authed.GET("/me", h.me)
}

// credentials carries the identity field (an email in production, or a plain
// pseudo in development) plus an optional password. Nothing is bound as required:
// presence and format are enforced per-mode in the handlers, so development mode
// can sign in with only a pseudo.
type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	credentials
	DisplayName string `json:"display_name"`
}

type userResponse struct {
	ID            int64  `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	EmailVerified bool   `json:"email_verified"`
}

func toUserResponse(u *store.User) userResponse {
	return userResponse{
		ID: u.ID, Email: u.Email,
		DisplayName: u.DisplayName, EmailVerified: u.EmailVerified,
	}
}

func (h *AuthHandler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	// Password policy: required in production; optional in development, but still
	// length-checked when supplied.
	if !h.devMode && len(req.Password) < minPasswordLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}
	if req.Password != "" && len(req.Password) < minPasswordLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}

	email, displayName, msg := h.resolveRegisterIdentity(req)
	if msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	hash, err := hashOptionalPassword(req.Password)
	if err != nil {
		serverError(c)
		return
	}

	if h.devMode {
		h.registerDev(c, email, displayName, hash)
		return
	}
	h.registerProd(c, email, displayName, hash)
}

// resolveRegisterIdentity derives the (email, displayName) to store and returns a
// user-facing error message if the request is invalid for the current mode. In
// development the "email" field is an optional pseudo, so the account email is
// derived from it; in production a real, well-formed email is required.
func (h *AuthHandler) resolveRegisterIdentity(req registerRequest) (email, displayName, msg string) {
	pseudo := strings.TrimSpace(req.Email)
	displayName = strings.TrimSpace(req.DisplayName)
	if h.devMode {
		if displayName == "" {
			displayName = pseudo
		}
		if displayName == "" {
			return "", "", "a username is required"
		}
		return devIdentityEmail(firstNonEmpty(pseudo, displayName)), displayName, ""
	}
	if displayName == "" {
		return "", "", "a display name is required"
	}
	addr, err := mail.ParseAddress(pseudo)
	if err != nil {
		return "", "", "a valid email is required"
	}
	return addr.Address, displayName, ""
}

// registerDev creates an auto-verified account and signs the user straight in.
func (h *AuthHandler) registerDev(c *gin.Context, email, displayName, hash string) {
	user, err := h.users.Create(c.Request.Context(), email, hash, displayName)
	if errors.Is(err, store.ErrEmailTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": "that username is already taken"})
		return
	}
	if err != nil {
		serverError(c)
		return
	}
	if err := h.issueSession(c, user.ID); err != nil {
		serverError(c)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(user))
}

// registerProd creates an unverified account, issues a verification link (logged
// in lieu of an email until SMTP is wired), and does not sign the user in: the
// account stays inactive until the link is followed.
func (h *AuthHandler) registerProd(c *gin.Context, email, displayName, hash string) {
	raw, tokenHash, err := auth.NewToken()
	if err != nil {
		serverError(c)
		return
	}
	user, err := h.users.CreatePending(c.Request.Context(), email, hash, displayName, tokenHash)
	if errors.Is(err, store.ErrEmailTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	if err != nil {
		serverError(c)
		return
	}
	// TODO: deliver this by email. Until SMTP is wired the link is logged.
	h.logger.Info("verification link issued",
		"email", user.Email, "url", "/api/v1/auth/verify?token="+raw)
	c.JSON(http.StatusCreated, gin.H{
		"email_verified": false,
		"message":        "Account created. Check your email to verify your address before signing in.",
	})
}

func (h *AuthHandler) login(c *gin.Context) {
	var req credentials
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	// In dev the identity field is a pseudo; map it to the derived email.
	identity := strings.TrimSpace(req.Email)
	if h.devMode {
		identity = devIdentityEmail(identity)
	}
	if identity == "" {
		invalidCredentials(c)
		return
	}
	user, err := h.users.GetByEmail(c.Request.Context(), identity)
	if errors.Is(err, store.ErrNotFound) {
		invalidCredentials(c)
		return
	}
	if err != nil {
		serverError(c)
		return
	}
	// In development mode the password is optional: an empty password
	// authenticates by identity alone. Otherwise it must match.
	if !h.devMode || req.Password != "" {
		switch err := auth.VerifyPassword(req.Password, user.PasswordHash); {
		case errors.Is(err, auth.ErrMismatch):
			invalidCredentials(c)
			return
		case err != nil:
			serverError(c)
			return
		}
	}
	if !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "email not verified; check your inbox for the verification link"})
		return
	}
	if err := h.issueSession(c, user.ID); err != nil {
		serverError(c)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

// verify consumes a verification token and marks the account verified.
func (h *AuthHandler) verify(c *gin.Context) {
	raw := c.Query("token")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}
	err := h.users.Verify(c.Request.Context(), auth.HashToken(raw))
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired verification token"})
		return
	}
	if err != nil {
		serverError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Email verified. You can now sign in."})
}

func (h *AuthHandler) logout(c *gin.Context) {
	if raw, err := c.Cookie(sessionCookie); err == nil && raw != "" {
		_ = h.sessions.Delete(c.Request.Context(), auth.HashToken(raw))
	}
	h.clearCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) me(c *gin.Context) {
	user, ok := CurrentUser(c)
	if !ok {
		unauthorized(c)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

// issueSession mints a session, persists its hash, and sets the cookie.
func (h *AuthHandler) issueSession(c *gin.Context, userID int64) error {
	raw, hash, err := auth.NewToken()
	if err != nil {
		return err
	}
	if _, err := h.sessions.Create(c.Request.Context(), hash, userID, h.sessionTTL); err != nil {
		return err
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookie,
		Value:    raw,
		Path:     "/",
		MaxAge:   int(h.sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (h *AuthHandler) clearCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode,
	})
}

// hashOptionalPassword returns an argon2id hash, or an empty string for an empty
// password (development-mode passwordless accounts).
func hashOptionalPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	return auth.HashPassword(password)
}

// devIdentityEmail maps a development-mode pseudo to the email used as the
// account key. A value that already looks like an email is kept as-is; a bare
// pseudo is slugified into a stable "<pseudo>@dev.local" address so the
// email-keyed schema and login-by-pseudo both work without the user ever
// entering an email.
func devIdentityEmail(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	if strings.Contains(v, "@") {
		return v
	}
	return slugIdentity(v) + "@dev.local"
}

// slugIdentity lowercases value and keeps only [a-z0-9], collapsing other runs to
// a single hyphen. Empty input yields "user".
func slugIdentity(value string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(value) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "user"
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func invalidCredentials(c *gin.Context) {
	// Deliberately generic: do not reveal whether the account exists.
	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
}

func serverError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}
