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
	"__MODULE__/internal/web"
	"__MODULE__/metrics"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Background worker. Replace the task body with the real work; readiness
	// flips true once the first run completes, gating /ready.
	var ready atomic.Bool
	task := func(_ context.Context) error {
		// TODO: real work here.
		ready.Store(true)
		return nil
	}
	go worker.New(task, cfg.Worker.Interval, logger).Run(ctx)

	srv := web.NewServer(ready.Load)

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
