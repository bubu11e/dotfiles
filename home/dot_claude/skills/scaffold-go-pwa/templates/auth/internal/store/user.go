package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// User is an account. PasswordHash is empty for a development-mode passwordless
// account (ADR-0004).
type User struct {
	ID            int64
	Email         string
	PasswordHash  string
	DisplayName   string
	EmailVerified bool
}

// UserStore persists accounts.
type UserStore struct{ db *sql.DB }

// NewUserStore wires a UserStore.
func NewUserStore(db *sql.DB) *UserStore { return &UserStore{db: db} }

// Create inserts an already-verified account and returns it. Used by
// development-mode registration, which has no verification step.
func (s *UserStore) Create(ctx context.Context, email, passwordHash, displayName string) (*User, error) {
	return s.insert(ctx, email, passwordHash, displayName, true, nil)
}

// CreatePending inserts an unverified account holding the hash of the
// verification token that will activate it.
func (s *UserStore) CreatePending(ctx context.Context, email, passwordHash, displayName, tokenHash string) (*User, error) {
	return s.insert(ctx, email, passwordHash, displayName, false, &tokenHash)
}

func (s *UserStore) insert(ctx context.Context, email, passwordHash, displayName string, verified bool, tokenHash *string) (*User, error) {
	// Email is the account key and is matched case-insensitively, so it is stored
	// folded rather than compared with LOWER() on every lookup.
	email = strings.ToLower(strings.TrimSpace(email))
	ts := now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users(email, password_hash, display_name, email_verified, verify_token_hash, created_at, updated_at)
		 VALUES(?,?,?,?,?,?,?)`,
		email, passwordHash, displayName, boolToInt(verified), tokenHash, ts, ts)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("user id: %w", err)
	}
	return &User{
		ID: id, Email: email, PasswordHash: passwordHash,
		DisplayName: displayName, EmailVerified: verified,
	}, nil
}

// GetByEmail looks an account up by its email, case-insensitively.
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.get(ctx, "email = ?", strings.ToLower(strings.TrimSpace(email)))
}

// GetByID looks an account up by its primary key.
func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.get(ctx, "id = ?", id)
}

func (s *UserStore) get(ctx context.Context, where string, arg any) (*User, error) {
	var u User
	var verified int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, display_name, email_verified FROM users WHERE `+where, arg,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &verified)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}
	u.EmailVerified = verified == 1
	return &u, nil
}

// Verify consumes a verification token hash, marking the account verified. The
// token is cleared so it cannot be replayed.
func (s *UserStore) Verify(ctx context.Context, tokenHash string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET email_verified = 1, verify_token_hash = NULL, updated_at = ?
		 WHERE verify_token_hash = ?`, now(), tokenHash)
	if err != nil {
		return fmt.Errorf("verify user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("verify user: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateDisplayName renames an account.
func (s *UserStore) UpdateDisplayName(ctx context.Context, id int64, name string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET display_name = ?, updated_at = ? WHERE id = ?", name, now(), id)
	if err != nil {
		return fmt.Errorf("update display name: %w", err)
	}
	return nil
}

// UpdatePassword replaces the stored password hash.
func (s *UserStore) UpdatePassword(ctx context.Context, id int64, hash string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", hash, now(), id)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isUniqueViolation reports whether err is a UNIQUE constraint failure. The
// modernc driver surfaces it only in the message, so there is no sentinel to
// compare against.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(strings.ToUpper(err.Error()), "UNIQUE CONSTRAINT FAILED")
}
