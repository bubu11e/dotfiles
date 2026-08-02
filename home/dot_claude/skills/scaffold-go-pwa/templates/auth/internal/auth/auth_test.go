package auth_test

import (
	"errors"
	"strings"
	"testing"

	"__MODULE__/internal/auth"
)

func TestHashPasswordRoundTrip(t *testing.T) {
	hash, err := auth.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := auth.VerifyPassword("correct horse battery staple", hash); err != nil {
		t.Errorf("VerifyPassword on the right password: %v", err)
	}
	if err := auth.VerifyPassword("wrong", hash); !errors.Is(err, auth.ErrMismatch) {
		t.Errorf("VerifyPassword on the wrong password = %v, want ErrMismatch", err)
	}
}

func TestHashPasswordIsSalted(t *testing.T) {
	a, _ := auth.HashPassword("same")
	b, _ := auth.HashPassword("same")
	if a == b {
		t.Error("two hashes of the same password are identical, so the salt is not random")
	}
}

func TestVerifyPasswordRejectsMalformedHashes(t *testing.T) {
	for _, encoded := range []string{"", "not-a-hash", "$argon2i$v=19$m=1,t=1,p=1$c2FsdA$a2V5"} {
		err := auth.VerifyPassword("x", encoded)
		if err == nil || errors.Is(err, auth.ErrMismatch) {
			t.Errorf("VerifyPassword(%q) = %v, want a format error", encoded, err)
		}
	}
}

func TestNewTokenIsRandomAndHashed(t *testing.T) {
	raw, hash, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if len(raw) < 40 {
		t.Errorf("raw token is only %d chars, want 256 bits of entropy", len(raw))
	}
	if strings.Contains(hash, raw) || hash == raw {
		t.Error("the stored hash must not contain the raw token")
	}
	if auth.HashToken(raw) != hash {
		t.Error("HashToken does not reproduce the hash returned by NewToken")
	}
	other, _, _ := auth.NewToken()
	if other == raw {
		t.Error("two tokens collided")
	}
}
