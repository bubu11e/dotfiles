# Build the embedded PWA first; Vite emits into internal/spa/dist, which the Go
# build embeds (ADR-0001). Building it here keeps the image reproducible outside
# CI, since internal/spa/dist is gitignored and absent from a checkout.
FROM node:24-alpine AS frontend

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
# Outputs to /app/internal/spa/dist (vite build.outDir is ../internal/spa/dist).
RUN npm run build

# ---

FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Bring in the built PWA so the //go:embed all:dist directive has files to embed.
COPY --from=frontend /app/internal/spa/dist ./internal/spa/dist

# Commit is supplied at build time (the build context has no .git, so the Go
# toolchain cannot embed vcs.revision itself). CI passes the Woodpecker built-in
# CI_COMMIT_SHA via build_args_from_env; a manual
# `docker build --build-arg COMMIT=$(git rev-parse HEAD)` works too.
ARG CI_COMMIT_SHA=""
ARG COMMIT="${CI_COMMIT_SHA}"
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X __MODULE__/version.Commit=${COMMIT}" \
    -o /bin/__NAME__ .

# ---

FROM alpine:3.24

# ca-certificates: outbound HTTPS. tzdata + TZ: correct local wall-clock time.
# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1000 app

USER app
WORKDIR /app

COPY --from=builder /bin/__NAME__ /bin/__NAME__
COPY config.example.yaml /app/config.yaml

ENV TZ=Europe/Paris \
    __ENV_PREFIX___PORT=__PORT__

EXPOSE __PORT__

# Liveness only; uses the build-time default port. If you run with a different
# __ENV_PREFIX___PORT, override the healthcheck accordingly.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- "http://localhost:__PORT__/live" >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/bin/__NAME__"]
