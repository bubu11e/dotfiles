// Package worker runs a periodic background task with the same lifecycle as the
// rest of the service: an initial run, then a tick loop that stops cleanly when
// the context is cancelled (on SIGINT/SIGTERM).
package worker

import (
	"context"
	"log/slog"
	"time"
)

// Task is the unit of work executed on each tick.
type Task func(ctx context.Context) error

// Worker runs Task immediately, then on every interval until ctx is cancelled.
type Worker struct {
	task     Task
	interval time.Duration
	logger   *slog.Logger
}

// New wires a Worker. A nil logger falls back to slog.Default().
func New(task Task, interval time.Duration, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{task: task, interval: interval, logger: logger}
}

// Run performs an initial execution, then ticks until ctx is cancelled. It never
// returns an error: a failed task is logged and the loop continues so a
// transient failure does not take the process down.
func (w *Worker) Run(ctx context.Context) {
	w.runOnce(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker stopping")
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *Worker) runOnce(ctx context.Context) {
	if err := w.task(ctx); err != nil {
		w.logger.Warn("worker task failed", "err", err)
	}
}
