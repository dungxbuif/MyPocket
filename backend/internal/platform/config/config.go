package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppEnv           string
	DatabaseURL      string
	PublicWebURL     string
	CookieSecret     string
	CSRFSecret       string
	S3Endpoint       string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
	OAuthFixtureMode bool
	HTTPAddr         string
}

func Load(env map[string]string) (Config, error) {
	cfg := Config{
		AppEnv:           valueOrDefault(env["APP_ENV"], "development"),
		DatabaseURL:      strings.TrimSpace(env["DATABASE_URL"]),
		PublicWebURL:     strings.TrimSpace(env["PUBLIC_WEB_URL"]),
		CookieSecret:     env["COOKIE_SECRET"],
		CSRFSecret:       env["CSRF_SECRET"],
		S3Endpoint:       strings.TrimSpace(env["S3_ENDPOINT"]),
		S3Bucket:         strings.TrimSpace(env["S3_BUCKET"]),
		S3AccessKey:      env["S3_ACCESS_KEY"],
		S3SecretKey:      env["S3_SECRET_KEY"],
		OAuthFixtureMode: env["OAUTH_FIXTURE_MODE"] == "true",
		HTTPAddr:         valueOrDefault(env["HTTP_ADDR"], ":8080"),
	}

	missing := make([]string, 0, 4)
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.PublicWebURL == "" {
		missing = append(missing, "PUBLIC_WEB_URL")
	}
	if cfg.CookieSecret == "" {
		missing = append(missing, "COOKIE_SECRET")
	}
	if cfg.CSRFSecret == "" {
		missing = append(missing, "CSRF_SECRET")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}

	if len(cfg.CookieSecret) < 32 {
		return Config{}, fmt.Errorf("invalid config: COOKIE_SECRET must be at least 32 bytes")
	}
	if len(cfg.CSRFSecret) < 32 {
		return Config{}, fmt.Errorf("invalid config: CSRF_SECRET must be at least 32 bytes")
	}
	if cfg.S3Endpoint != "" {
		s3Missing := make([]string, 0, 3)
		if cfg.S3Bucket == "" {
			s3Missing = append(s3Missing, "S3_BUCKET")
		}
		if cfg.S3AccessKey == "" {
			s3Missing = append(s3Missing, "S3_ACCESS_KEY")
		}
		if cfg.S3SecretKey == "" {
			s3Missing = append(s3Missing, "S3_SECRET_KEY")
		}
		if len(s3Missing) > 0 {
			return Config{}, fmt.Errorf("missing required config: %s", strings.Join(s3Missing, ", "))
		}
	}
	if cfg.AppEnv == "production" && cfg.OAuthFixtureMode {
		return Config{}, fmt.Errorf("invalid config: OAUTH_FIXTURE_MODE is not allowed in production")
	}

	return cfg, nil
}

func LoadFromEnv() (Config, error) {
	return Load(mapFromEnv(os.Environ()))
}

func mapFromEnv(values []string) map[string]string {
	env := make(map[string]string, len(values))
	for _, pair := range values {
		key, value, ok := strings.Cut(pair, "=")
		if ok {
			env[key] = value
		}
	}
	return env
}

func valueOrDefault(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
