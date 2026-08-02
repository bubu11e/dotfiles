# 3. Keep state in SQLite inside the single artifact

Date: __DATE__

## Status

Accepted

## Context

__TITLE__ needs accounts and sessions. A separate database server would double
the deployment surface for what is a single-instance internal app, and the
service already ships as one binary (ADR-0001).

## Decision

Persist to a SQLite file, opened with the pure-Go `modernc.org/sqlite` driver,
with the schema embedded under `migrations/` and applied at startup.

## Consequences

- `CGO_ENABLED=0` keeps working, so the Docker image stays a static binary on
  Alpine with no toolchain in the runtime stage.
- Migrations are ordinary `.sql` files applied in lexical order and recorded in
  `schema_migrations`; running them twice is a no-op, so startup is safe to
  repeat.
- WAL journalling, enforced foreign keys, and a busy timeout are set on the DSN.
  The pool is capped at one connection: SQLite serialises writes anyway, and a
  cap turns contention into a wait instead of a `SQLITE_BUSY` error.
- One writer means one instance. Scaling out horizontally would mean replacing
  this decision, not tuning it.
- The database file belongs on a volume; the path in `config.example.yaml` is a
  relative one meant for a local run.
