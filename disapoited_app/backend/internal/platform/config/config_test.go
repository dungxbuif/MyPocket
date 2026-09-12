package config_test

import (
	"strings"
	"testing"
	"time"

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

func TestLoadParsesAllowedLoginEmails(t *testing.T) {
	cfg, err := config.Load(map[string]string{
		"APP_ENV":              "test",
		"DATABASE_URL":         "postgres://localhost/mypocket",
		"PUBLIC_WEB_URL":       "http://localhost:5173",
		"COOKIE_SECRET":        "01234567890123456789012345678901",
		"CSRF_SECRET":          "abcdefghijklmnopqrstuvwxyz123456",
		"ALLOWED_LOGIN_EMAILS": " A@Example.com, b@example.com ,,",
	})
	if err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
	if len(cfg.AllowedLoginEmails) != 2 || cfg.AllowedLoginEmails[0] != "a@example.com" || cfg.AllowedLoginEmails[1] != "b@example.com" {
		t.Fatalf("wrong allowed emails: %#v", cfg.AllowedLoginEmails)
	}
}

func TestLoadRequiresCompleteS3ConfigWhenEndpointConfigured(t *testing.T) {
	_, err := config.Load(map[string]string{
		"APP_ENV":        "test",
		"DATABASE_URL":   "postgres://localhost/mypocket",
		"PUBLIC_WEB_URL": "http://localhost:5173",
		"COOKIE_SECRET":  "01234567890123456789012345678901",
		"CSRF_SECRET":    "abcdefghijklmnopqrstuvwxyz123456",
		"S3_ENDPOINT":    "http://localhost:9000",
		"S3_BUCKET":      "mypocket",
		"S3_SECRET_KEY":  "do-not-leak-this-secret",
	})
	if err == nil || !strings.Contains(err.Error(), "S3_ACCESS_KEY") {
		t.Fatalf("expected S3_ACCESS_KEY error, got %v", err)
	}
	if strings.Contains(err.Error(), "do-not-leak") {
		t.Fatalf("config error leaked S3 secret: %v", err)
	}
}

func TestProductionRequiresAuditAndJSONLogging(t *testing.T) {
	_, err := config.Load(map[string]string{
		"APP_ENV":        "production",
		"DATABASE_URL":   "postgres://localhost/mypocket",
		"PUBLIC_WEB_URL": "https://app.example.com",
		"COOKIE_SECRET":  "01234567890123456789012345678901",
		"CSRF_SECRET":    "abcdefghijklmnopqrstuvwxyz123456",
		"LOG_FORMAT":     "text",
	})
	if err == nil || !strings.Contains(err.Error(), "AUDIT_VIEWER_EMAIL") {
		t.Fatalf("expected audit viewer production error, got %v", err)
	}

	_, err = config.Load(map[string]string{
		"APP_ENV":             "production",
		"DATABASE_URL":        "postgres://localhost/mypocket",
		"PUBLIC_WEB_URL":      "https://app.example.com",
		"COOKIE_SECRET":       "01234567890123456789012345678901",
		"CSRF_SECRET":         "abcdefghijklmnopqrstuvwxyz123456",
		"AUDIT_VIEWER_EMAIL":  "owner@example.com",
		"AUDIT_HASH_SECRET":   "01234567890123456789012345678901",
		"API_KEY_HASH_SECRET": "01234567890123456789012345678901",
		"REDIS_URL":           "redis://localhost:6379/0",
		"LOG_FORMAT":          "text",
	})
	if err == nil || !strings.Contains(err.Error(), "LOG_FORMAT") {
		t.Fatalf("expected JSON logging production error, got %v", err)
	}
}

func TestProductionRejectsInsecurePublicWebURL(t *testing.T) {
	_, err := config.Load(map[string]string{
		"APP_ENV":             "production",
		"DATABASE_URL":        "postgres://localhost/mypocket",
		"PUBLIC_WEB_URL":      "http://app.example.com",
		"COOKIE_SECRET":       "01234567890123456789012345678901",
		"CSRF_SECRET":         "abcdefghijklmnopqrstuvwxyz123456",
		"AUDIT_VIEWER_EMAIL":  "owner@example.com",
		"AUDIT_HASH_SECRET":   "01234567890123456789012345678901",
		"API_KEY_HASH_SECRET": "01234567890123456789012345678901",
		"REDIS_URL":           "redis://localhost:6379/0",
		"LOG_FORMAT":          "json",
	})
	if err == nil || !strings.Contains(err.Error(), "PUBLIC_WEB_URL") {
		t.Fatalf("expected https public web URL error, got %v", err)
	}
}

func TestWebPushConfigurationMustBeCompleteAndEnabledInProduction(t *testing.T) {
	base := map[string]string{
		"APP_ENV":             "production",
		"DATABASE_URL":        "postgres://localhost/mypocket",
		"PUBLIC_WEB_URL":      "https://app.example.com",
		"COOKIE_SECRET":       "01234567890123456789012345678901",
		"CSRF_SECRET":         "abcdefghijklmnopqrstuvwxyz123456",
		"AUDIT_VIEWER_EMAIL":  "owner@example.com",
		"AUDIT_HASH_SECRET":   "01234567890123456789012345678901",
		"API_KEY_HASH_SECRET": "01234567890123456789012345678901",
		"REDIS_URL":           "redis://localhost:6379/0",
		"LOG_FORMAT":          "json",
	}
	if _, err := config.Load(base); err == nil || !strings.Contains(err.Error(), "WEB_PUSH_ENABLED") {
		t.Fatalf("expected production Web Push requirement, got %v", err)
	}
	base["WEB_PUSH_ENABLED"] = "true"
	base["VAPID_PUBLIC_KEY"] = "public"
	if _, err := config.Load(base); err == nil || !strings.Contains(err.Error(), "VAPID_PRIVATE_KEY") {
		t.Fatalf("expected complete VAPID configuration, got %v", err)
	}
	base["VAPID_PRIVATE_KEY"] = "private"
	base["VAPID_SUBJECT"] = "owner@example.com"
	if cfg, err := config.Load(base); err != nil || !cfg.WebPushEnabled {
		t.Fatalf("expected valid production Web Push config, cfg=%#v err=%v", cfg, err)
	}
}

func TestAPIRateLimitDefaultsAndRejectsUnsafeBounds(t *testing.T) {
	base := map[string]string{"DATABASE_URL": "postgres://localhost/mypocket", "PUBLIC_WEB_URL": "http://localhost", "COOKIE_SECRET": "01234567890123456789012345678901", "CSRF_SECRET": "abcdefghijklmnopqrstuvwxyz123456"}
	cfg, err := config.Load(base)
	if err != nil || cfg.APIRateLimitPerMinute != 120 {
		t.Fatalf("default rate limit cfg=%#v err=%v", cfg, err)
	}
	base["API_RATE_LIMIT_PER_MINUTE"] = "0"
	if _, err = config.Load(base); err == nil || !strings.Contains(err.Error(), "API_RATE_LIMIT_PER_MINUTE") {
		t.Fatalf("expected rate limit error, got %v", err)
	}
	base["API_RATE_LIMIT_PER_MINUTE"] = "10001"
	if _, err = config.Load(base); err == nil {
		t.Fatal("expected upper bound error")
	}
}

func TestLoadAIConfigurationIsFailClosedAndSecretFree(t *testing.T) {
	env := map[string]string{"DATABASE_URL": "postgres://localhost/mypocket", "PUBLIC_WEB_URL": "http://localhost", "COOKIE_SECRET": "01234567890123456789012345678901", "CSRF_SECRET": "abcdefghijklmnopqrstuvwxyz123456", "AI_ENABLED": "true", "AI_BASE_URL": "http://provider.test", "AI_API_KEY": "do-not-leak", "AI_MODEL": "model-a"}
	cfg, err := config.Load(env)
	if err != nil || !cfg.AI.Enabled || cfg.AI.Timeout.String() != "30s" || cfg.AI.MaxRetries != 2 {
		t.Fatalf("unexpected config %#v error=%v", cfg.AI, err)
	}
	env["AI_TIMEOUT"] = "500ms"
	_, err = config.Load(env)
	if err == nil || strings.Contains(err.Error(), "do-not-leak") {
		t.Fatalf("expected redacted timeout error, got %v", err)
	}
	env["AI_TIMEOUT"] = "30s"
	env["APP_ENV"] = "production"
	_, err = config.Load(env)
	if err == nil || !strings.Contains(err.Error(), "AI_BASE_URL must use https") || strings.Contains(err.Error(), "do-not-leak") {
		t.Fatalf("expected production TLS validation, got %v", err)
	}
}

func TestLoadOCRConfigurationIsBoundedAndSecretFree(t *testing.T) {
	env := map[string]string{"DATABASE_URL": "postgres://localhost/mypocket", "PUBLIC_WEB_URL": "http://localhost", "COOKIE_SECRET": "01234567890123456789012345678901", "CSRF_SECRET": "abcdefghijklmnopqrstuvwxyz123456", "S3_ENDPOINT": "http://s3.test", "S3_BUCKET": "bucket", "S3_ACCESS_KEY": "access", "S3_SECRET_KEY": "secret", "AI_ENABLED": "true", "AI_BASE_URL": "http://ai.test", "AI_API_KEY": "ai-secret", "AI_MODEL": "model", "OCR_ENABLED": "true", "OCR_BASE_URL": "http://ocr.test", "OCR_API_KEY": "do-not-leak"}
	cfg, err := config.Load(env)
	if err != nil || !cfg.OCR.Enabled || cfg.OCR.PollInterval != 5*time.Second {
		t.Fatalf("cfg=%#v err=%v", cfg.OCR, err)
	}
	env["OCR_MAX_PROCESSING_TIME"] = "30s"
	_, err = config.Load(env)
	if err == nil || strings.Contains(err.Error(), "do-not-leak") {
		t.Fatalf("expected redacted bound error: %v", err)
	}
}
