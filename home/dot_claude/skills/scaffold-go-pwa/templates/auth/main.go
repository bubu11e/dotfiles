// Command __NAME__ TODO: one-line description of the service.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"__MODULE__/config"
	"__MODULE__/internal/spa"
	"__MODULE__/internal/storage"
	"__MODULE__/internal/store"
	"__MODULE__/internal/web"
	"__MODULE__/metrics"
	"__MODULE__/migrations"
	"__MODULE__/version"
	"__MODULE__/worker"
)

func main() {
	configPath := flag.String("config", envOr("__ENV_PREFIX___CONFIG", "config.yaml"),
		"path to the YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ver := version.Get()
	metrics.BuildInfo.WithLabelValues(ver.Version, version.ShortCommit(ver.Commit), ver.GoVersion).Set(1)
	logger.Info("__NAME__ starting",
		"version", ver.Version, "commit", version.ShortCommit(ver.Commit), "go", ver.GoVersion)

	if cfg.Auth.DevMode {
		// Loud on purpose: dev mode makes the password optional and skips email
		// verification, so it must never go unnoticed in a real deployment.
		logger.Warn("authentication is in development mode: passwords are optional and accounts are auto-verified")
	}

	db, err := storage.Open(cfg.Storage.DBPath)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()
	if err := storage.Migrate(db, migrations.FS); err != nil {
		logger.Error("migrate database", "error", err)
		os.Exit(1)
	}
	logger.Info("storage ready", "db", cfg.Storage.DBPath)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Background heartbeat: each tick gates /ready on database health. Add the
	// real periodic work to the task body.
	var ready atomic.Bool
	task := func(ctx context.Context) error {
		if err := db.PingContext(ctx); err != nil {
			ready.Store(false)
			return err
		}
		ready.Store(true)
		return nil
	}
	go worker.New(task, cfg.Worker.Interval, logger).Run(ctx)

	srv := web.NewServer(ready.Load)

	users := store.NewUserStore(db)
	sessions := store.NewSessionStore(db)

	web.NewAuthHandler(users, sessions, cfg.Auth.SessionTTL, cfg.Auth.SecureCookies, cfg.Auth.DevMode, logger).
		Register(srv.Engine())
	web.NewInstanceHandler(cfg.Instance.Name, cfg.Auth.DevMode).Register(srv.Engine())

	// Serve the embedded PWA for everything the API and probes do not claim.
	if err := spa.Mount(srv.Engine()); err != nil {
		logger.Error("mount spa", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{Addr: cfg.Server.Addr(), Handler: srv.Handler()}

	// Graceful shutdown: on SIGINT/SIGTERM, drain in-flight requests. The worker
	// goroutine returns when ctx is cancelled.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown", "error", err)
		}
	}()

	logger.Info("server starting", "addr", cfg.Server.Addr())
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
}
