// Package config loads and validates __NAME__ configuration from a YAML file,
// with __ENV_PREFIX___* environment overrides applied on top.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level application configuration.
type Config struct {
	Server   Server   `yaml:"server"`
	LogLevel string   `yaml:"log_level"`
	Worker   Worker   `yaml:"worker"`
	Storage  Storage  `yaml:"storage"`
	Auth     Auth     `yaml:"auth"`
	Instance Instance `yaml:"instance"`
}

// Server holds HTTP server settings.
type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Addr returns the host:port the server listens on.
func (s Server) Addr() string { return fmt.Sprintf("%s:%d", s.Host, s.Port) }

// Worker holds background worker settings.
type Worker struct {
	Interval time.Duration `yaml:"interval"`
}

// Storage holds the single-instance persistence locations.
type Storage struct {
	DBPath string `yaml:"db_path"`
}

// Auth holds authentication settings for the local email + password accounts.
type Auth struct {
	SessionTTL    time.Duration `yaml:"session_ttl"`
	SecureCookies bool          `yaml:"secure_cookies"` // set true when served over HTTPS
	// DevMode relaxes auth for local development: the password becomes optional
	// and accounts are auto-verified (no verification email). Leave it false in
	// production, where a password and a verified email are both required.
	DevMode bool `yaml:"dev_mode"`
}

// Instance holds the presentation settings the client reads before it has a
// session.
type Instance struct {
	Name string `yaml:"name"`
}

// Load reads the YAML file at path over the built-in defaults, overlays
// __ENV_PREFIX___* env overrides, then validates the result.
func Load(path string) (*Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.applyEnv(); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Server:   Server{Host: "0.0.0.0", Port: __PORT__},
		LogLevel: "info",
		Worker:   Worker{Interval: time.Minute},
		Storage:  Storage{DBPath: "__NAME__.db"},
		Auth:     Auth{SessionTTL: 30 * 24 * time.Hour, SecureCookies: false},
		Instance: Instance{Name: "__TITLE__"},
	}
}

func (c *Config) applyEnv() error {
	if v, ok := os.LookupEnv("__ENV_PREFIX___HOST"); ok {
		c.Server.Host = v
	}
	if v, ok := os.LookupEnv("__ENV_PREFIX___PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("__ENV_PREFIX___PORT %q: %w", v, err)
		}
		c.Server.Port = port
	}
	if v, ok := os.LookupEnv("__ENV_PREFIX___LOG_LEVEL"); ok {
		c.LogLevel = v
	}
	if v, ok := os.LookupEnv("__ENV_PREFIX___WORKER_INTERVAL"); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("__ENV_PREFIX___WORKER_INTERVAL %q: %w", v, err)
		}
		c.Worker.Interval = d
	}
	if v, ok := os.LookupEnv("__ENV_PREFIX___DB_PATH"); ok {
		c.Storage.DBPath = v
	}
	if v, ok := os.LookupEnv("__ENV_PREFIX___SESSION_TTL"); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("__ENV_PREFIX___SESSION_TTL %q: %w", v, err)
		}
		c.Auth.SessionTTL = d
	}
	if v, ok := os.LookupEnv("__ENV_PREFIX___SECURE_COOKIES"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("__ENV_PREFIX___SECURE_COOKIES %q: %w", v, err)
		}
		c.Auth.SecureCookies = b
	}
	if v, ok := os.LookupEnv("__ENV_PREFIX___DEV_MODE"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("__ENV_PREFIX___DEV_MODE %q: %w", v, err)
		}
		c.Auth.DevMode = b
	}
	if v, ok := os.LookupEnv("__ENV_PREFIX___INSTANCE_NAME"); ok {
		c.Instance.Name = v
	}
	return nil
}

func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port %d out of range", c.Server.Port)
	}
	if c.Worker.Interval <= 0 {
		return fmt.Errorf("worker.interval must be positive")
	}
	if c.Storage.DBPath == "" {
		return fmt.Errorf("storage.db_path must not be empty")
	}
	if c.Auth.SessionTTL <= 0 {
		return fmt.Errorf("auth.session_ttl must be positive")
	}
	return nil
}
