# __TITLE__ architecture

TODO: describe what __TITLE__ does and the high-level data flow.

## Layout

- `main.go` -- wiring: config load, logger, signal handling, graceful shutdown.
- `config/` -- YAML + `__ENV_PREFIX___*` env config with defaults and validation.
- `internal/web/` -- Gin HTTP server: probes, metrics, build info, and the API.
- `internal/spa/` -- serves the embedded front-end build.
- `frontend/` -- Vue 3 + Vite PWA, built into `internal/spa/dist`.
- `metrics/` -- Prometheus metrics and the HTTP middleware.
- `version/` -- build information (commit injected via ldflags).
- `worker/` -- ticker-based background task with graceful cancellation.

## Request routing

Gin matches the API and operational routes first. Everything else reaches
`internal/spa`, which serves a real asset when one exists and otherwise returns
`index.html` so the client-side router owns the path. `/api/*` and the probe
paths are reserved: an unknown path under them returns a JSON 404 rather than the
app shell, so a mistyped API call fails as an API call.

## Operational endpoints

- `GET /live` (+ `/health`) -- liveness; the process is up.
- `GET /ready` -- readiness; 200 when functional, 503 otherwise.
- `GET /version`, `GET /api/v1/version` -- build information.
- `GET /metrics` -- Prometheus metrics (HTTP + domain).

## Observability

Logs are JSON on stderr via `slog`, at the level set by `log_level`. Metrics are
prefixed `__NAME___`: HTTP count and latency (route label bounded by the Gin route
template), `__NAME___build_info`, and `__NAME___ready`.

## Lifecycle

The process installs a `signal.NotifyContext` for SIGINT/SIGTERM. On signal it
drains in-flight HTTP requests via `http.Server.Shutdown` (10s timeout) and the
worker goroutine returns when the context is cancelled.
