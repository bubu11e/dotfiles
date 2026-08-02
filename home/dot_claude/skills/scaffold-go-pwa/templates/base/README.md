# __TITLE__ __EMOJI__

__DESCRIPTION__

A single Go binary: a Gin REST API with the Vue PWA embedded, structured JSON
logs, Prometheus metrics, and operational probes.

## Development

The front-end is built into `internal/spa/dist`, which `go:embed` reads. That
directory is gitignored, so build it once before any Go command:

```bash
cd frontend && npm ci && npm run build && cd ..

go run . -config config.example.yaml   # dev server on :__PORT__
go build -o __NAME__ .                 # static binary
go test ./... -cover
go vet ./...
gofmt -l .
pre-commit run --all-files             # golangci-lint + hadolint + yamllint + gitleaks
```

For front-end work, `npm run dev` in `frontend/` serves with hot reload and
proxies `/api` to the Go backend on `:__PORT__`.

```bash
cd frontend
npm run dev
npm test
npm run test:coverage
```

## Docker

```bash
docker build -t __NAME__ .
docker run -p __PORT__:__PORT__ __NAME__
```

The image builds the front-end in its own stage, so no local `npm` is needed.

## Configuration

Settings live in `config.yaml` (see `config.example.yaml`). Every value is
overridable via `__ENV_PREFIX___*` environment variables.

## Operational endpoints

| Path | Purpose |
| --- | --- |
| `/live`, `/health` | Liveness |
| `/ready` | Readiness (503 when not ready) |
| `/version`, `/api/v1/version` | Build information |
| `/metrics` | Prometheus metrics |

Everything else falls through to the embedded PWA.

## Icons

The app identity is the emoji __EMOJI__. `frontend/public/icons/` holds the
rasterised PNGs (install, maskable, apple-touch); the browser tab uses an inline
SVG favicon built from the same emoji. See `docs/adr/0002-emoji-derived-identity.md`
to regenerate them.
