// Package store holds the persistence layer: one type per aggregate, each taking
// a *sql.DB and returning domain structs. SQL never leaves this package.
package store

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a lookup matches no row. Callers distinguish it
// from a real failure with errors.Is.
var ErrNotFound = errors.New("store: not found")

// ErrEmailTaken is returned when an email is already registered.
var ErrEmailTaken = errors.New("store: email already registered")

// now returns the timestamp format every table stores: ISO-8601 in UTC, so
// string ordering matches chronological ordering.
func now() string { return time.Now().UTC().Format(time.RFC3339) }
