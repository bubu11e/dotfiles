# 1. Ship a single binary with the PWA embedded

Date: __DATE__

## Status

Accepted

## Context

__TITLE__ is a Go API with a web UI. Serving the UI from a separate origin (a
CDN, an nginx sidecar) means two artifacts to version, a CORS story, and a way
for the two halves to drift apart across a deploy.

## Decision

Build the front-end with Vite into `internal/spa/dist` and embed it with
`go:embed all:dist`. The Gin engine serves it from `NoRoute`: real assets
directly, everything else the `index.html` shell.

## Consequences

- One artifact. `docker build` produces the whole app; a deploy cannot leave the
  UI and the API on different versions.
- Same origin, so the session cookie needs no CORS or `SameSite=None` relaxation.
- `internal/spa/dist` is gitignored, so a fresh checkout cannot `go build` until
  `npm run build` has run. The Dockerfile and the Woodpecker pipeline both build
  the front-end in an earlier stage; the README says so for local work.
- `all:` on the embed directive is required: without it Go skips the files Vite
  emits whose names begin with `_`.
- Probe and `/api/` paths are reserved in the fallback handler, so a mistyped API
  path returns a JSON 404 instead of the HTML shell.
