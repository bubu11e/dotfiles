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
	if cfg.Server.Addr() != "127.0.0.1:__PORT__" {
		t.Errorf("addr = %q", cfg.Server.Addr())
	}
	if cfg.Storage.DBPath != "__NAME__.db" {
		t.Errorf("db_path = %q", cfg.Storage.DBPath)
	}
	if cfg.Instance.Name != "__TITLE__" {
		t.Errorf("instance.name = %q", cfg.Instance.Name)
	}
	// Dev mode must default off: the safe value is the one you get by omission.
	if cfg.Auth.DevMode {
		t.Error("auth.dev_mode defaults to true, want false")
	}
}

func TestEnvOverridesFile(t *testing.T) {
	t.Setenv("__ENV_PREFIX___PORT", "9999")
	t.Setenv("__ENV_PREFIX___DEV_MODE", "true")
	t.Setenv("__ENV_PREFIX___SESSION_TTL", "15m")

	cfg, err := config.Load(write(t, "server:\n  port: 1234\nauth:\n  dev_mode: false\n"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("port = %d, want 9999 from env", cfg.Server.Port)
	}
	if !cfg.Auth.DevMode {
		t.Error("dev_mode = false, want the env override to win over the file")
	}
	if cfg.Auth.SessionTTL != 15*time.Minute {
		t.Errorf("session_ttl = %s, want 15m", cfg.Auth.SessionTTL)
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	if _, err := config.Load(write(t, "server:\n  port: 70000\n")); err == nil {
		t.Error("expected an error for an out-of-range port")
	}
	if _, err := config.Load(write(t, "storage:\n  db_path: ''\n")); err == nil {
		t.Error("expected an error for an empty db_path")
	}
	if _, err := config.Load(write(t, "auth:\n  session_ttl: 0s\n")); err == nil {
		t.Error("expected an error for a zero session_ttl")
	}
	t.Setenv("__ENV_PREFIX___DEV_MODE", "maybe")
	if _, err := config.Load(write(t, "")); err == nil {
		t.Error("expected an error for a non-boolean __ENV_PREFIX___DEV_MODE")
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
	t.Setenv("__ENV_PREFIX___DB_PATH", "/data/app.db")
	t.Setenv("__ENV_PREFIX___SESSION_TTL", "1h")
	t.Setenv("__ENV_PREFIX___SECURE_COOKIES", "true")
	t.Setenv("__ENV_PREFIX___DEV_MODE", "true")
	t.Setenv("__ENV_PREFIX___INSTANCE_NAME", "Renamed")

	cfg, err := config.Load(write(t, ""))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := map[string]any{
		"host": cfg.Server.Host, "port": cfg.Server.Port, "log_level": cfg.LogLevel,
		"interval": cfg.Worker.Interval, "db_path": cfg.Storage.DBPath,
		"session_ttl": cfg.Auth.SessionTTL, "secure_cookies": cfg.Auth.SecureCookies,
		"dev_mode": cfg.Auth.DevMode, "instance_name": cfg.Instance.Name,
	}
	want := map[string]any{
		"host": "10.0.0.1", "port": 9001, "log_level": "debug",
		"interval": 2 * time.Minute, "db_path": "/data/app.db",
		"session_ttl": time.Hour, "secure_cookies": true,
		"dev_mode": true, "instance_name": "Renamed",
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Errorf("%s = %v, want %v", key, got[key], wantValue)
		}
	}
}

func TestRejectsMalformedDurationsAndPorts(t *testing.T) {
	for _, env := range []string{"__ENV_PREFIX___WORKER_INTERVAL", "__ENV_PREFIX___SESSION_TTL"} {
		t.Setenv(env, "soon")
		if _, err := config.Load(write(t, "")); err == nil {
			t.Errorf("expected an error for a non-duration %s", env)
		}
		t.Setenv(env, "")
	}
	t.Setenv("__ENV_PREFIX___WORKER_INTERVAL", "1m")
	t.Setenv("__ENV_PREFIX___SESSION_TTL", "1h")
	t.Setenv("__ENV_PREFIX___SECURE_COOKIES", "sometimes")
	if _, err := config.Load(write(t, "")); err == nil {
		t.Error("expected an error for a non-boolean __ENV_PREFIX___SECURE_COOKIES")
	}
}
