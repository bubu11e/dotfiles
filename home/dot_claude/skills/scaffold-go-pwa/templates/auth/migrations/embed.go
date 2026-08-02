// Package migrations holds the SQL schema migrations, embedded into the binary
// so the single artifact can self-migrate on startup (ADR-0003).
package migrations

import "embed"

// FS contains the ordered *.sql migration files. They are applied in lexical
// order, so prefix new files with a zero-padded sequence (0001_, 0002_, ...).
//
//go:embed *.sql
var FS embed.FS
