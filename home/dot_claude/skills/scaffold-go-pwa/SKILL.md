---
name: scaffold-go-pwa
description: >-
  Scaffold a new Go service in the established house style: a Gin REST API with
  a Vue 3 PWA embedded in the binary, structured JSON slog
  logging, Prometheus metrics and /live + /ready + /version probes, YAML + env
  config, an emoji-derived icon and favicon set, an optional sign-in page with a
  dev mode, a multi-stage non-root Dockerfile, Woodpecker CI, Renovate, and a
  golangci-lint + hadolint + yamllint + gitleaks pre-commit stack. Use when the
  user wants to start a new Go app, API, service, PWA, or microservice.
---

# Scaffold a Go API with an embedded PWA

Generate a new project from the bundled templates, substituting per-project
placeholders. The result builds, tests, and passes `pre-commit` out of the box.

Every app this skill produces is a **PWA**: an installable manifest, a service
worker, and a full icon set are not optional extras here. So are **structured
JSON logs via `slog`** and a **Prometheus `/metrics` endpoint**. Do not offer to
drop any of the three.

## What it produces

A module laid out like the existing services:

- `main.go` — config load, JSON `slog` logger set as the default, `signal.NotifyContext`
  for SIGINT/SIGTERM, worker startup, HTTP server, graceful shutdown.
- `config/` — YAML (`gopkg.in/yaml.v3`) + `<PREFIX>_*` env overrides, defaults,
  validation, with tests.
- `internal/web/` — Gin engine (`ReleaseMode`, `Recovery`, gzip, metrics
  middleware) exposing `/live`, `/health`, `/ready`, `/version`,
  `/api/v1/version`, `/metrics`, with tests.
- `internal/spa/` — `go:embed all:dist` of the built front-end, served from
  `NoRoute` with an `index.html` fallback and reserved `/api/` + probe paths.
- `frontend/` — Vue 3 + Vite + `vite-plugin-pwa` (injectManifest), Vitest tests,
  light/dark tokens driven by `<html data-theme>`.
- `metrics/` — `promauto` metrics plus an in-house Gin middleware recording
  `<name>_http_requests_total` and `<name>_http_request_duration_seconds`
  (route label = `c.FullPath()` for bounded cardinality), `<name>_build_info`,
  `<name>_ready`.
- `version/` — build info; commit injected via `-ldflags -X`.
- `worker/` — ticker-based background task with an initial run and ctx cancel.
- `Dockerfile` (node build stage → `golang:1.26-alpine` → `alpine:3.24`,
  `CGO_ENABLED=0`, non-root uid 1000, `HEALTHCHECK` on `/live`),
  `.woodpecker.yml`, `.golangci.yml`, `.pre-commit-config.yaml`, `.yamllint`,
  `renovate.json`, `.gitignore`, `.gitattributes`, `.dockerignore`,
  `.editorconfig`, `config.example.yaml`, `README.md`, `CLAUDE.md`,
  `CONTEXT.md`, `docs/architecture.md`, `docs/adr/`.

With the sign-in page (asked, see below), it also produces `internal/auth/`
(argon2id + opaque tokens), `internal/store/` (users, sessions),
`internal/storage/` + `migrations/` (SQLite via `modernc.org/sqlite`, self-migrating),
`internal/web/auth.go` + `middleware.go` + `instance.go`, and a Vue `AuthView`
with a router guard.

## Ask the user

Ask only for what you cannot infer. **The sign-in question and the emoji are
mandatory — never assume either.**

1. **App name** — derive `__NAME__` (lowercase, `[a-z0-9_]`), `__TITLE__`
   (display), `__ENV_PREFIX__` (upper), `__MODULE__` = `forgejo.local/<name>`,
   `__REGISTRY__` = `docker-registry.home.mpli.fr/<name>`. State the derivations
   rather than asking four times.
2. **One-line description** — `__DESCRIPTION__`, used in the manifest, the
   `<meta name="description">`, the README, and the placeholder view.
3. **Identity emoji** — `__EMOJI__`. One emoji; it becomes the favicon and every
   install icon (ADR-0002). Suggest two or three fitting the description.
4. **Sign-in page?** — default yes. Yes brings the auth overlay: accounts,
   sessions, SQLite, and the dev mode. No leaves a single-user app with no
   identity concept at all.
5. **HTTP port** — default `8080`.
6. **Target directory** — default `../<name>`, a sibling of the current repo.

Colours have sane defaults; ask only if the user raises them:
`__ACCENT__` `#0088b0`, `__BG_COLOR__` `#f3f2f2`, `__DARK_BG_COLOR__` `#1b1d21`.

`__NAME__` becomes a Prometheus metric prefix, so it must match
`[a-zA-Z_][a-zA-Z0-9_]*` — underscores, never hyphens.

## Placeholders

Every file is run through this substitution.

| Token | Meaning | Example |
| --- | --- | --- |
| `__MODULE__` | Go module path | `forgejo.local/myapp` |
| `__NAME__` | binary + metric prefix | `myapp` |
| `__TITLE__` | display name | `MyApp` |
| `__ENV_PREFIX__` | env var prefix (UPPER) | `MYAPP` |
| `__PORT__` | default HTTP port | `8080` |
| `__REGISTRY__` | Docker image repo | `docker-registry.home.mpli.fr/myapp` |
| `__EMOJI__` | app identity emoji | `🚀` |
| `__DESCRIPTION__` | one-line description | `Tracks the things.` |
| `__ACCENT__` | accent colour | `#0088b0` |
| `__BG_COLOR__` | light background / theme-color | `#f3f2f2` |
| `__DARK_BG_COLOR__` | dark background | `#1b1d21` |
| `__DATE__` | ADR date, `YYYY-MM-DD` | `2026-08-02` |

The `__TOKEN__` form is deliberate: `{{ }}` would collide with Vue's mustache
interpolation and with chezmoi's own templating of this skill's files.

## Execution flow

### 1. Create the target and copy templates

```bash
mkdir -p <target> && cd <target> && git init   # if not already a repo
cp -R <skill>/templates/base/. <target>/
cp -R <skill>/templates/auth/. <target>/       # only if a sign-in page was chosen
```

The trailing `/.` matters: the templates include dotfiles. The auth overlay
deliberately overwrites `main.go`, `config/`, `config.example.yaml`,
`.gitignore`, and several `frontend/src` files. Copy base first, always.

If any `dot_*` files turn up in `<target>`, you are reading the templates from
the dotfiles repo rather than the deployed skill (chezmoi stores them prefixed
because it ignores source files beginning with `.`). Rename them:

```bash
find <target> -name 'dot_*' -not -path '*/.git/*' \
  -exec sh -c 'mv "$1" "$(dirname "$1")/.$(basename "$1" | cut -c5-)"' _ {} \;
```

### 2. Substitute

Use `perl -pi` — it behaves identically on macOS and Linux, unlike `sed -i`
whose empty-suffix argument macOS `xargs` silently drops.

```bash
sub() { find <target> -type f -not -path '*/.git/*' -print0 \
          | xargs -0 perl -pi -e "s{\Q$1\E}{$2}g"; }
sub '__MODULE__'        'forgejo.local/myapp'
sub '__NAME__'          'myapp'
sub '__TITLE__'         'MyApp'
sub '__ENV_PREFIX__'    'MYAPP'
sub '__PORT__'          '8080'
sub '__REGISTRY__'      'docker-registry.home.mpli.fr/myapp'
sub '__EMOJI__'         '🚀'
sub '__DESCRIPTION__'   'Tracks the things.'
sub '__ACCENT__'        '#0088b0'
sub '__BG_COLOR__'      '#f3f2f2'
sub '__DARK_BG_COLOR__' '#1b1d21'
sub '__DATE__'          "$(date +%F)"
```

Then confirm none remain: `grep -rn '__[A-Z_]*__' <target>` should print nothing
but Go's own `__WB_MANIFEST` in `frontend/src/sw.js`, which is workbox's and must
stay.

### 3. Generate the icons

```bash
swift <skill>/scripts/make-icons.swift \
  --emoji '🚀' --out <target>/frontend/public/icons --background '#f3f2f2'
rm <target>/frontend/public/icons/README.md
```

It writes `icon-192.png`, `icon-512.png`, `icon-maskable-512.png` and
`apple-touch-icon.png`, and it is macOS-only: Apple Color Emoji is the only
emoji font on a stock machine here. CI never runs it — the icons are committed
build inputs. If `swift` is unavailable, say so and leave the PNGs missing
rather than inventing artwork; the SVG data-URI favicon in `index.html` still
works, only installability suffers.

Look at one of the PNGs before moving on. A blank square means the emoji did not
render.

### 4. Build and validate

The front-end must be built first: `internal/spa/dist` is gitignored and
`go:embed all:dist` fails to compile without it.

```bash
cd <target>/frontend && npm install && npm run build && npm test && cd ..
go mod tidy
go build ./...
go vet ./...
gofmt -l .            # expect no output
go test ./... -cover
pre-commit install && pre-commit run --all-files
```

All of it must pass before you report done. `npm install` (not `ci`) on the
first run: there is no lockfile yet, and the one it writes is what
`.woodpecker.yml` and the Dockerfile then use.

### 5. Report next steps

Tell the user to: fill in the `TODO`s (description, `CONTEXT.md` glossary, real
worker task, domain metrics), set the Woodpecker `docker_registry_*` secrets,
turn `auth.dev_mode` off before any real deployment, and commit with
`/commit-and-push`.

## Variants

**Headless service (no UI).** Delete `frontend/`, `internal/spa/`, and the
`public/icons` step; drop the `spa` import and the `spa.Mount` block from
`main.go`; delete the `frontend` step from `.woodpecker.yml` and the node stage
from the `Dockerfile`. Everything else stands.

**Persistence without a sign-in page.** Take `internal/storage/`, `migrations/`,
and the `storage:` config section from the auth overlay, and replace the
`0001_init.sql` schema with the real domain tables.

## House conventions these templates encode

Do not quietly undo them when adapting the output.

- One package per business concept under `internal/`; `internal/web/` is
  `server.go` plus one file per handler group.
- Logs are JSON on stderr through `slog` at the configured level. Never
  `fmt.Println`, never the `log` package.
- Metrics are prefixed with the app name and the HTTP route label is
  `c.FullPath()`, so cardinality stays bounded. The middleware is in-house on
  purpose: `gin-contrib` has no Prometheus middleware.
- Config is YAML with `<PREFIX>_*` env overrides on top, defaults in code, and
  validation that runs on every load. The safe value is always the default.
- The commit SHA reaches `/version` through `-ldflags -X`, fed by Woodpecker's
  `CI_COMMIT_SHA` via `build_args_from_env`; the build context has no `.git`.
- Docker: multi-stage, `CGO_ENABLED=0`, non-root uid 1000, `HEALTHCHECK` on
  `/live`, `TZ=Europe/Paris`.
- Renovate extends `local>julien/renovate-config`; nothing else belongs in
  `renovate.json`.
- Conventional Commits, TDD with coverage >= 80% on both sides, small files,
  docs under `./docs`, no emojis in code or docs — the identity emoji is data,
  not decoration.
- Any non-obvious decision becomes an ADR under `docs/adr/`. The templates ship
  ADR-0001 (single binary), ADR-0002 (emoji identity), and with auth, ADR-0003
  (SQLite in the binary) and ADR-0004 (dev mode). Keep them; add to them.
