package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"__MODULE__/config"
)

func write(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoadAppliesDefaults(t *testing.T) {
	cfg, err := config.Load(write(t, "server:\n  host: 127.0.0.1\n"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != __PORT__ {
		t.Errorf("port = %d, want __PORT__", cfg.Server.Port)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("log_level = %q, want info", cfg.LogLevel)
	}
	if cfg.Server.Addr() != "127.0.0.1:__PORT__" {
		t.Errorf("addr = %q", cfg.Server.Addr())
	}
}

func TestEnvOverridesFile(t *testing.T) {
	t.Setenv("__ENV_PREFIX___PORT", "9999")
	t.Setenv("__ENV_PREFIX___WORKER_INTERVAL", "30s")

	cfg, err := config.Load(write(t, "server:\n  port: 1234\n"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("port = %d, want 9999 from env", cfg.Server.Port)
	}
	if cfg.Worker.Interval != 30*time.Second {
		t.Errorf("interval = %s, want 30s", cfg.Worker.Interval)
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	if _, err := config.Load(write(t, "server:\n  port: 70000\n")); err == nil {
		t.Error("expected an error for an out-of-range port")
	}
	t.Setenv("__ENV_PREFIX___PORT", "not-a-number")
	if _, err := config.Load(write(t, "")); err == nil {
		t.Error("expected an error for a non-numeric __ENV_PREFIX___PORT")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := config.Load(filepath.Join(t.TempDir(), "absent.yaml")); err == nil {
		t.Error("expected an error for a missing config file")
	}
}

func TestEveryEnvOverrideIsWired(t *testing.T) {
	// One test per variable would be noise; what matters is that no override is
	// silently missing from applyEnv, which is easy to forget when adding a field.
	t.Setenv("__ENV_PREFIX___HOST", "10.0.0.1")
	t.Setenv("__ENV_PREFIX___PORT", "9001")
	t.Setenv("__ENV_PREFIX___LOG_LEVEL", "debug")
	t.Setenv("__ENV_PREFIX___WORKER_INTERVAL", "2m")

	cfg, err := config.Load(write(t, ""))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Host != "10.0.0.1" || cfg.Server.Port != 9001 {
		t.Errorf("addr = %q, want 10.0.0.1:9001", cfg.Server.Addr())
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("log_level = %q, want debug", cfg.LogLevel)
	}
	if cfg.Worker.Interval != 2*time.Minute {
		t.Errorf("interval = %s, want 2m", cfg.Worker.Interval)
	}
}

func TestRejectsAMalformedDuration(t *testing.T) {
	t.Setenv("__ENV_PREFIX___WORKER_INTERVAL", "soon")
	if _, err := config.Load(write(t, "")); err == nil {
		t.Error("expected an error for a non-duration __ENV_PREFIX___WORKER_INTERVAL")
	}
}
