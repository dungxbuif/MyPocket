package config

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AIBaseURL          string
	AIAPIKey           string
	AIModel            string
	OCRAPIURL          string
	OCRAPIKey          string
	S3Endpoint         string
	S3Region           string
	S3Bucket           string
	S3Prefix           string
	S3AccessKeyID      string
	S3SecretAccessKey  string
	S3ForcePathStyle   bool
	AppEnv             string
	HTTPAddr           string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	JWTTTL             time.Duration
	OAuthFixtureMode   bool
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	CORSAllowedOrigins []string
	AllowedLoginEmails []string
	SeedEmail          string
	SeedName           string
	SeedPassword       string
}

func Load() Config {
	return Config{
		AIBaseURL:          getenv("AI_BASE_URL", ""),
		AIAPIKey:           getenv("AI_API_KEY", ""),
		AIModel:            getenv("AI_MODEL", ""),
		OCRAPIURL:          getenv("OCR_API_URL", "https://ocr.dungxbuif.com"),
		OCRAPIKey:          getenv("OCR_API_KEY", ""),
		S3Endpoint:         getenv("S3_ENDPOINT", ""),
		S3Region:           getenv("S3_REGION", ""),
		S3Bucket:           getenv("S3_BUCKET", ""),
		S3Prefix:           getenv("S3_PREFIX", "mypocket/receipts"),
		S3AccessKeyID:      getenv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey:  getenv("S3_SECRET_ACCESS_KEY", ""),
		S3ForcePathStyle:   getenvBool("S3_FORCE_PATH_STYLE", true),
		AppEnv:             getenv("APP_ENV", "development"),
		HTTPAddr:           getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:        getenv("DATABASE_URL", "postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable"),
		RedisURL:           getenv("REDIS_URL", "redis://127.0.0.1:6379/0"),
		JWTSecret:          getenv("JWT_SECRET", "change-this-development-jwt-secret-32-bytes"),
		JWTTTL:             getDuration("JWT_TTL_SECONDS", 3600),
		OAuthFixtureMode:   getenvBool("OAUTH_FIXTURE_MODE", false),
		GoogleClientID:     getenv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getenv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getenv("GOOGLE_REDIRECT_URL", ""),
		CORSAllowedOrigins: parseCSV(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:4173,http://127.0.0.1:4173")),
		AllowedLoginEmails: parseCSV(getenv("ALLOWED_LOGIN_EMAILS", "")),
		SeedEmail:          getenv("SEED_USER_EMAIL", "admin@mypocket.local"),
		SeedName:           getenv("SEED_USER_NAME", "MyPocket Admin"),
		SeedPassword:       getenv("SEED_USER_PASSWORD", "12345678"),
	}
}

func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultSeconds int64) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(defaultSeconds) * time.Second
	}
	tokenTTL, err := strconv.ParseInt(value, 10, 64)
	if err != nil || tokenTTL <= 0 {
		return time.Duration(defaultSeconds) * time.Second
	}
	return time.Duration(tokenTTL) * time.Second
}

func getenvBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseCSV(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	r := csv.NewReader(strings.NewReader(value))
	items, err := r.Read()
	if err != nil || len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if normalized := strings.ToLower(strings.TrimSpace(item)); normalized != "" {
			result = append(result, normalized)
		}
	}
	return result
}
