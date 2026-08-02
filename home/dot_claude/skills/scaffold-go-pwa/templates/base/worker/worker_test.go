package worker_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"__MODULE__/worker"
)

func TestRunExecutesImmediatelyAndStopsOnCancel(t *testing.T) {
	var runs atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker.New(func(context.Context) error {
			runs.Add(1)
			return nil
		}, time.Hour, nil).Run(ctx)
		close(done)
	}()

	// The initial run happens before the first tick, so a one-hour interval still
	// yields exactly one execution.
	waitFor(t, func() bool { return runs.Load() == 1 })
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

func TestRunSurvivesAFailingTask(t *testing.T) {
	var runs atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.New(func(context.Context) error {
		runs.Add(1)
		return errors.New("boom")
	}, 10*time.Millisecond, nil).Run(ctx)

	// A failing task must not stop the loop: the tick after the failure still runs.
	waitFor(t, func() bool { return runs.Load() >= 2 })
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met before the deadline")
}
