package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Session is a signed-in session. ID is the SHA-256 of the raw cookie value.
type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
}

// SessionStore persists sessions.
type SessionStore struct{ db *sql.DB }

// NewSessionStore wires a SessionStore.
func NewSessionStore(db *sql.DB) *SessionStore { return &SessionStore{db: db} }

// Create records a session under the hash of its raw token, expiring after ttl.
func (s *SessionStore) Create(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) (*Session, error) {
	expires := time.Now().UTC().Add(ttl)
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO sessions(id, user_id, created_at, expires_at) VALUES(?,?,?,?)",
		tokenHash, userID, now(), expires.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}
	return &Session{ID: tokenHash, UserID: userID, ExpiresAt: expires}, nil
}

// GetValid returns the unexpired session for a token hash. An expired session is
// deleted and reported as ErrNotFound, so expiry needs no separate sweep.
func (s *SessionStore) GetValid(ctx context.Context, tokenHash string) (*Session, error) {
	var sess Session
	var expires string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, user_id, expires_at FROM sessions WHERE id = ?", tokenHash,
	).Scan(&sess.ID, &sess.UserID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select session: %w", err)
	}
	sess.ExpiresAt, err = time.Parse(time.RFC3339, expires)
	if err != nil {
		return nil, fmt.Errorf("parse session expiry: %w", err)
	}
	if !time.Now().UTC().Before(sess.ExpiresAt) {
		_ = s.Delete(ctx, tokenHash)
		return nil, ErrNotFound
	}
	return &sess, nil
}

// Delete removes a session, signing that client out.
func (s *SessionStore) Delete(ctx context.Context, tokenHash string) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
