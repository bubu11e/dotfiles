// Package storage opens the SQLite database and applies the embedded schema
// migrations. It uses the pure-Go modernc.org/sqlite driver so the binary keeps
// building with CGO_ENABLED=0 (see the Dockerfile / ADR-0003).
package storage

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"

	_ "modernc.org/sqlite"
)

// Open opens (creating if necessary) the SQLite database at path with the
// pragmas the single-instance topology needs: WAL journalling, enforced foreign
// keys, and a busy timeout to ride out the single writer.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)",
		path,
	)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	// One writer at a time: SQLite serialises writes anyway, and a bounded pool
	// turns lock contention into a wait instead of a SQLITE_BUSY error.
	db.SetMaxOpenConns(1)
	return db, nil
}

// Migrate applies every *.sql file in migrationFS in lexical order, recording
// each in schema_migrations so a second run is a no-op.
func Migrate(db *sql.DB, migrationFS fs.FS) error {
	if _, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	names, err := fs.Glob(migrationFS, "*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		var applied int
		if err := db.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?", name,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied > 0 {
			continue
		}
		body, err := fs.ReadFile(migrationFS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		// One transaction per file so a failure leaves no half-applied schema.
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.Exec(string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations(version) VALUES(?)", name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}
