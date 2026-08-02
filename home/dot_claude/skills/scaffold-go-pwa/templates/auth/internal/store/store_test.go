package store_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"__MODULE__/internal/storage"
	"__MODULE__/internal/store"
	"__MODULE__/migrations"
)

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

func TestCreateAndLookupUser(t *testing.T) {
	users := store.NewUserStore(openDB(t))
	ctx := context.Background()

	created, err := users.Create(ctx, "  Ada@Example.COM ", "hash", "Ada")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Email != "ada@example.com" {
		t.Errorf("email = %q, want it trimmed and folded", created.Email)
	}
	if !created.EmailVerified {
		t.Error("Create should produce a verified account")
	}

	// Lookup is case-insensitive because the stored form is folded.
	found, err := users.GetByEmail(ctx, "ADA@EXAMPLE.COM")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("GetByEmail returned id %d, want %d", found.ID, created.ID)
	}
}

func TestCreateRejectsADuplicateEmail(t *testing.T) {
	users := store.NewUserStore(openDB(t))
	ctx := context.Background()
	if _, err := users.Create(ctx, "a@b.c", "h", "A"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := users.Create(ctx, "A@B.C", "h", "A"); !errors.Is(err, store.ErrEmailTaken) {
		t.Errorf("duplicate Create = %v, want ErrEmailTaken", err)
	}
}

func TestMissingUserIsNotFound(t *testing.T) {
	users := store.NewUserStore(openDB(t))
	if _, err := users.GetByEmail(context.Background(), "nobody@example.com"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("GetByEmail on a missing user = %v, want ErrNotFound", err)
	}
}

func TestVerifyConsumesTheToken(t *testing.T) {
	users := store.NewUserStore(openDB(t))
	ctx := context.Background()

	user, err := users.CreatePending(ctx, "a@b.c", "h", "A", "token-hash")
	if err != nil {
		t.Fatalf("CreatePending: %v", err)
	}
	if user.EmailVerified {
		t.Error("CreatePending should produce an unverified account")
	}
	if err := users.Verify(ctx, "token-hash"); err != nil {
		t.Fatalf("Verify: %v", err)
	}
	got, _ := users.GetByID(ctx, user.ID)
	if !got.EmailVerified {
		t.Error("account still unverified after Verify")
	}
	// Replaying the same link must fail: the token was cleared.
	if err := users.Verify(ctx, "token-hash"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("replayed Verify = %v, want ErrNotFound", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	db := openDB(t)
	users := store.NewUserStore(db)
	sessions := store.NewSessionStore(db)
	ctx := context.Background()

	user, err := users.Create(ctx, "a@b.c", "h", "A")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := sessions.Create(ctx, "hash", user.ID, time.Hour); err != nil {
		t.Fatalf("Create session: %v", err)
	}
	sess, err := sessions.GetValid(ctx, "hash")
	if err != nil {
		t.Fatalf("GetValid: %v", err)
	}
	if sess.UserID != user.ID {
		t.Errorf("session user = %d, want %d", sess.UserID, user.ID)
	}
	if err := sessions.Delete(ctx, "hash"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := sessions.GetValid(ctx, "hash"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("GetValid after Delete = %v, want ErrNotFound", err)
	}
}

func TestExpiredSessionIsRejectedAndReaped(t *testing.T) {
	db := openDB(t)
	users := store.NewUserStore(db)
	sessions := store.NewSessionStore(db)
	ctx := context.Background()

	user, _ := users.Create(ctx, "a@b.c", "h", "A")
	if _, err := sessions.Create(ctx, "hash", user.ID, -time.Minute); err != nil {
		t.Fatalf("Create session: %v", err)
	}
	if _, err := sessions.GetValid(ctx, "hash"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("GetValid on an expired session = %v, want ErrNotFound", err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE id='hash'").Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Error("an expired session should be deleted on the failed lookup")
	}
}

func TestUpdateDisplayNameAndPassword(t *testing.T) {
	users := store.NewUserStore(openDB(t))
	ctx := context.Background()

	user, err := users.Create(ctx, "a@b.c", "old-hash", "Ada")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := users.UpdateDisplayName(ctx, user.ID, "Ada Lovelace"); err != nil {
		t.Fatalf("UpdateDisplayName: %v", err)
	}
	if err := users.UpdatePassword(ctx, user.ID, "new-hash"); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	got, err := users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.DisplayName != "Ada Lovelace" {
		t.Errorf("display name = %q", got.DisplayName)
	}
	if got.PasswordHash != "new-hash" {
		t.Errorf("password hash = %q, want the updated one", got.PasswordHash)
	}
}

func TestDeletingAUserCascadesToItsSessions(t *testing.T) {
	db := openDB(t)
	users := store.NewUserStore(db)
	sessions := store.NewSessionStore(db)
	ctx := context.Background()

	user, _ := users.Create(ctx, "a@b.c", "h", "A")
	if _, err := sessions.Create(ctx, "hash", user.ID, time.Hour); err != nil {
		t.Fatalf("Create session: %v", err)
	}
	if _, err := db.Exec("DELETE FROM users WHERE id = ?", user.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if _, err := sessions.GetValid(ctx, "hash"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("session survived its user: %v", err)
	}
}
