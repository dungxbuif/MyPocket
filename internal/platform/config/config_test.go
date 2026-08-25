package config_test

import (
	"strings"
	"testing"

	"mypocket/internal/platform/config"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	_, err := config.Load(map[string]string{"APP_ENV": "test"})
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected missing DATABASE_URL error, got %v", err)
	}
}

func TestLoadRedactsSecretValues(t *testing.T) {
	_, err := config.Load(map[string]string{
		"APP_ENV":        "test",
		"DATABASE_URL":   "postgres://user:secret@localhost/db",
		"PUBLIC_WEB_URL": "http://localhost:5173",
		"COOKIE_SECRET":  "short",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("config error leaked secret: %v", err)
	}
}

func TestLoadAcceptsRequiredRuntimeValues(t *testing.T) {
	cfg, err := config.Load(map[string]string{
		"APP_ENV":        "test",
		"DATABASE_URL":   "postgres://localhost/mypocket",
		"PUBLIC_WEB_URL": "http://localhost:5173",
		"COOKIE_SECRET":  "01234567890123456789012345678901",
		"CSRF_SECRET":    "abcdefghijklmnopqrstuvwxyz123456",
	})
	if err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
	if cfg.AppEnv != "test" || cfg.DatabaseURL == "" || cfg.PublicWebURL == "" {
		t.Fatalf("loaded wrong config: %#v", cfg)
	}
}
