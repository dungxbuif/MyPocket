package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv                string
	DatabaseURL           string
	PublicWebURL          string
	CookieSecret          string
	CSRFSecret            string
	S3Endpoint            string
	S3Bucket              string
	S3AccessKey           string
	S3SecretKey           string
	OAuthFixtureMode      bool
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
	AllowedLoginEmails    []string
	APIKeyHashSecret      string
	RedisURL              string
	AuditViewerEmail      string
	AuditHashSecret       string
	AuditRetentionDays    int
	LogFormat             string
	LogLevel              string
	HTTPAddr              string
	DocsDir               string
	WebPushEnabled        bool
	VAPIDPublicKey        string
	VAPIDPrivateKey       string
	VAPIDSubject          string
	APIRateLimitPerMinute int
}

func Load(env map[string]string) (Config, error) {
	cfg := Config{
		AppEnv:                valueOrDefault(env["APP_ENV"], "development"),
		DatabaseURL:           strings.TrimSpace(env["DATABASE_URL"]),
		PublicWebURL:          strings.TrimSpace(env["PUBLIC_WEB_URL"]),
		CookieSecret:          env["COOKIE_SECRET"],
		CSRFSecret:            env["CSRF_SECRET"],
		S3Endpoint:            strings.TrimSpace(env["S3_ENDPOINT"]),
		S3Bucket:              strings.TrimSpace(env["S3_BUCKET"]),
		S3AccessKey:           env["S3_ACCESS_KEY"],
		S3SecretKey:           env["S3_SECRET_KEY"],
		OAuthFixtureMode:      env["OAUTH_FIXTURE_MODE"] == "true",
		GoogleClientID:        strings.TrimSpace(env["GOOGLE_CLIENT_ID"]),
		GoogleClientSecret:    env["GOOGLE_CLIENT_SECRET"],
		GoogleRedirectURL:     strings.TrimSpace(env["GOOGLE_REDIRECT_URL"]),
		AllowedLoginEmails:    parseCSV(env["ALLOWED_LOGIN_EMAILS"]),
		APIKeyHashSecret:      env["API_KEY_HASH_SECRET"],
		RedisURL:              strings.TrimSpace(env["REDIS_URL"]),
		AuditViewerEmail:      strings.ToLower(strings.TrimSpace(env["AUDIT_VIEWER_EMAIL"])),
		AuditHashSecret:       env["AUDIT_HASH_SECRET"],
		AuditRetentionDays:    intOrDefault(env["AUDIT_RETENTION_DAYS"], 180),
		LogFormat:             valueOrDefault(env["LOG_FORMAT"], "text"),
		LogLevel:              valueOrDefault(env["LOG_LEVEL"], "info"),
		HTTPAddr:              valueOrDefault(env["HTTP_ADDR"], ":8080"),
		DocsDir:               strings.TrimSpace(env["DOCS_DIR"]),
		WebPushEnabled:        env["WEB_PUSH_ENABLED"] == "true",
		VAPIDPublicKey:        strings.TrimSpace(env["VAPID_PUBLIC_KEY"]),
		VAPIDPrivateKey:       env["VAPID_PRIVATE_KEY"],
		VAPIDSubject:          strings.TrimSpace(env["VAPID_SUBJECT"]),
		APIRateLimitPerMinute: intOrDefault(env["API_RATE_LIMIT_PER_MINUTE"], 120),
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
	if cfg.AuditRetentionDays <= 0 {
		return Config{}, fmt.Errorf("invalid config: AUDIT_RETENTION_DAYS must be positive")
	}
	if cfg.LogFormat != "text" && cfg.LogFormat != "json" {
		return Config{}, fmt.Errorf("invalid config: LOG_FORMAT must be text or json")
	}
	if cfg.LogLevel != "debug" && cfg.LogLevel != "info" && cfg.LogLevel != "warn" && cfg.LogLevel != "error" {
		return Config{}, fmt.Errorf("invalid config: LOG_LEVEL must be debug, info, warn, or error")
	}
	if cfg.WebPushEnabled {
		pushMissing := make([]string, 0, 3)
		if cfg.VAPIDPublicKey == "" {
			pushMissing = append(pushMissing, "VAPID_PUBLIC_KEY")
		}
		if cfg.VAPIDPrivateKey == "" {
			pushMissing = append(pushMissing, "VAPID_PRIVATE_KEY")
		}
		if cfg.VAPIDSubject == "" {
			pushMissing = append(pushMissing, "VAPID_SUBJECT")
		}
		if len(pushMissing) > 0 {
			return Config{}, fmt.Errorf("missing required config: %s", strings.Join(pushMissing, ", "))
		}
	} else if cfg.VAPIDPublicKey != "" || cfg.VAPIDPrivateKey != "" || cfg.VAPIDSubject != "" {
		return Config{}, fmt.Errorf("invalid config: WEB_PUSH_ENABLED must be true when VAPID settings are present")
	}
	if cfg.APIRateLimitPerMinute <= 0 || cfg.APIRateLimitPerMinute > 10_000 {
		return Config{}, fmt.Errorf("invalid config: API_RATE_LIMIT_PER_MINUTE must be between 1 and 10000")
	}
	if cfg.AppEnv == "production" {
		if !strings.HasPrefix(cfg.PublicWebURL, "https://") {
			return Config{}, fmt.Errorf("invalid config: PUBLIC_WEB_URL must use https in production")
		}
		if cfg.AuditViewerEmail == "" {
			return Config{}, fmt.Errorf("missing required config: AUDIT_VIEWER_EMAIL")
		}
		if len(cfg.AuditHashSecret) < 32 {
			return Config{}, fmt.Errorf("invalid config: AUDIT_HASH_SECRET must be at least 32 bytes")
		}
		if len(cfg.APIKeyHashSecret) < 32 {
			return Config{}, fmt.Errorf("invalid config: API_KEY_HASH_SECRET must be at least 32 bytes")
		}
		if cfg.RedisURL == "" {
			return Config{}, fmt.Errorf("missing required config: REDIS_URL")
		}
		if cfg.LogFormat != "json" {
			return Config{}, fmt.Errorf("invalid config: LOG_FORMAT must be json in production")
		}
		if !cfg.WebPushEnabled {
			return Config{}, fmt.Errorf("invalid config: WEB_PUSH_ENABLED must be true in production")
		}
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

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func intOrDefault(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
