package storage_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"

	"__MODULE__/internal/storage"
	"__MODULE__/migrations"
)

func openMigrated(t *testing.T) *sql.DB {
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

func TestMigrateCreatesSchema(t *testing.T) {
	db := openMigrated(t)
	for _, table := range []string{"users", "sessions"} {
		var name string
		if err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name); err != nil {
			t.Errorf("table %q missing: %v", table, err)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := openMigrated(t)
	if err := storage.Migrate(db, migrations.FS); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM schema_migrations WHERE version='0001_init.sql'",
	).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("migration recorded %d times, want 1", count)
	}
}

func TestForeignKeysEnforced(t *testing.T) {
	db := openMigrated(t)
	// A session pointing at a missing user must fail, which only happens with the
	// foreign_keys pragma actually on.
	_, err := db.Exec(
		"INSERT INTO sessions(id, user_id, created_at, expires_at) VALUES(?,?,?,?)",
		"tok", 999, "now", "later",
	)
	if err == nil {
		t.Error("expected a foreign key violation, got nil")
	}
}

func TestMigrateRollsBackAFailingFile(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "__NAME__.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	bad := fstest.MapFS{
		"0001_ok.sql":  {Data: []byte("CREATE TABLE a (id INTEGER PRIMARY KEY);")},
		"0002_bad.sql": {Data: []byte("CREATE TABLE b (id INTEGER PRIMARY KEY); NOT SQL;")},
	}
	if err := storage.Migrate(db, bad); err == nil {
		t.Fatal("expected an error from the malformed migration")
	}
	// The good file before it stays applied; the failing one leaves nothing behind,
	// which is what the per-file transaction buys.
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='b'").Scan(&count); err != nil {
		t.Fatalf("inspect schema: %v", err)
	}
	if count != 0 {
		t.Error("the failing migration left table b behind")
	}
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM schema_migrations WHERE version='0002_bad.sql'").Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 0 {
		t.Error("a failed migration must not be recorded as applied")
	}
}

func TestOpenRejectsAnUnusablePath(t *testing.T) {
	// A directory where the file should be: the driver only notices on first use,
	// which is why Open pings rather than trusting sql.Open.
	if _, err := storage.Open(t.TempDir()); err == nil {
		t.Error("expected an error opening a directory as a database")
	}
}
