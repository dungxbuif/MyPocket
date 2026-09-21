package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalEnvLoadsSecretsWithoutOverridingProcess(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, ".env.local")
	if err := os.WriteFile(file, []byte("# local\nAI_API_KEY=fixture-key\nS3_BUCKET=receipts\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AI_API_KEY", "process-key")
	t.Setenv("S3_BUCKET", "")
	if err := LoadLocalEnv(file); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("AI_API_KEY") != "process-key" {
		t.Fatal("overwrote runtime secret")
	}
	if os.Getenv("S3_BUCKET") != "" {
		t.Fatal("overwrote explicit environment")
	}
	t.Setenv("OCR_API_URL", "https://fixture.invalid")
	t.Setenv("S3_ENDPOINT", "https://storage.invalid")
	t.Setenv("S3_REGION", "region")
	cfg := Load()
	if cfg.OCRAPIURL != "https://fixture.invalid" || cfg.S3Endpoint != "https://storage.invalid" || cfg.S3Region != "region" {
		t.Fatal("integration config not wired")
	}
	if err := LoadLocalEnv(filepath.Join(dir, "missing")); err != nil {
		t.Fatal(err)
	}
}
