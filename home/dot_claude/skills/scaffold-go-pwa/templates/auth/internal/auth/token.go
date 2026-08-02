package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// tokenBytes is the entropy of a raw session/verification token (256 bits).
const tokenBytes = 32

// NewToken returns a cryptographically random opaque token (raw, given to the
// client) and its SHA-256 hash (stored server-side). Only the hash is persisted,
// so a database leak does not expose usable tokens.
func NewToken() (raw, hash string, err error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("read token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, HashToken(raw), nil
}

// HashToken returns the SHA-256 hex digest of a raw token, for lookup/comparison.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
