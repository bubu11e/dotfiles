# CLAUDE.md

This file guides Claude Code when working in this repository.

## Project Overview

__TITLE__ -- __DESCRIPTION__

TODO: expand once the domain is real. Read [CONTEXT.md](CONTEXT.md) before
touching domain code.

## Development Commands

`internal/spa/dist` is gitignored and embedded by `go:embed`, so the front-end
must be built before any Go command in a fresh checkout.

```bash
cd frontend && npm ci && npm run build && cd ..
go run . -config config.example.yaml   # dev server
go build -o __NAME__ .                 # static binary
go test ./... -cover                   # tests
go vet ./...
gofmt -l .
pre-commit run --all-files             # golangci-lint + hadolint + yamllint + gitleaks
cd frontend && npm test                # front-end unit tests
```

## Architecture

See [docs/architecture.md](docs/architecture.md).

- Gin HTTP server in `internal/web/`: `server.go` plus one file per handler group.
- Embedded Vue PWA in `frontend/`, built into `internal/spa/dist` and served by
  `internal/spa`.
- Background work in `worker/`, driven from `main.go` with graceful shutdown.
- Config in `config/` (YAML + `__ENV_PREFIX___*` env overrides).
- Prometheus metrics in `metrics/`; build info in `version/`.
- One package per business concept under `internal/`.

## Conventions

- Conventional Commits; explain the "why" in the body.
- TDD; keep coverage >= 80% on both sides.
- Small files; docs under `./docs`.
- No emojis in code, comments, or docs. The app identity emoji in `index.html`,
  the manifest, and the icons is the one exception, and it is data, not decoration.
- Structured JSON logs via `slog` only; never `fmt.Println` or the `log` package.
- Any non-obvious decision becomes an ADR under `docs/adr/`.
