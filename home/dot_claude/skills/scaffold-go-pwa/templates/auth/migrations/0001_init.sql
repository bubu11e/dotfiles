-- __TITLE__ initial schema: identity and sessions.
-- Timestamps are ISO-8601 UTC TEXT set by the application.

CREATE TABLE users (
    id                INTEGER PRIMARY KEY,
    email             TEXT NOT NULL UNIQUE,
    -- Empty for a development-mode passwordless account (ADR-0004).
    password_hash     TEXT NOT NULL,
    display_name      TEXT NOT NULL,
    email_verified    INTEGER NOT NULL DEFAULT 0 CHECK (email_verified IN (0, 1)),
    -- SHA-256 of the emailed verification token; NULL once consumed.
    verify_token_hash TEXT,
    created_at        TEXT NOT NULL,
    updated_at        TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_users_verify_token
    ON users(verify_token_hash) WHERE verify_token_hash IS NOT NULL;

-- id is the SHA-256 of the raw cookie value, never the value itself, so a
-- database leak does not hand out usable sessions.
CREATE TABLE sessions (
    id         TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
