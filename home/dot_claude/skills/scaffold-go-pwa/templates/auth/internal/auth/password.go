// Package auth provides password hashing and opaque session tokens for the local
// email + password accounts.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// argon2id parameters. Tuned for an interactive login on modest hardware; encoded
// into each hash so they can be raised later without invalidating old hashes.
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MiB
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

// ErrMismatch is returned by VerifyPassword when the password does not match.
var ErrMismatch = errors.New("auth: password mismatch")

// HashPassword returns an argon2id PHC-style encoded hash of password.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword reports whether password matches the encoded argon2id hash.
// Returns ErrMismatch on a valid hash that does not match, or another error if
// the encoded hash is malformed.
func VerifyPassword(password, encoded string) error {
	memory, time, threads, salt, key, err := decodeHash(encoded)
	if err != nil {
		return err
	}
	candidate := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(key)))
	if subtle.ConstantTimeCompare(key, candidate) == 1 {
		return nil
	}
	return ErrMismatch
}

func decodeHash(encoded string) (memory, time uint32, threads uint8, salt, key []byte, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return 0, 0, 0, nil, nil, errors.New("auth: invalid hash format")
	}
	var version int
	if _, err = fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return 0, 0, 0, nil, nil, errors.New("auth: incompatible hash version")
	}
	if _, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("auth: invalid hash params: %w", err)
	}
	if salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("auth: invalid salt: %w", err)
	}
	if key, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("auth: invalid key: %w", err)
	}
	return memory, time, threads, salt, key, nil
}
