package ai

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mypocket/backend/internal/config"
)

func TestLiveOCRPDFWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("OCR_LIVE_TEST") != "1" {
		t.Skip("set OCR_LIVE_TEST=1 to exercise OCR with a local PDF fixture")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal(err)
	}
	cfg := config.Load()
	data, err := os.ReadFile(os.Getenv("OCR_LIVE_TEST_FILE"))
	if err != nil || len(data) == 0 {
		t.Fatal("live PDF fixture could not be read")
	}
	c := NewClient(Config{OCRURL: cfg.OCRAPIURL, OCRKey: cfg.OCRAPIKey})
	text, err := c.ocr(context.Background(), Image{Name: "receipt.pdf", MIMEType: "application/pdf", Base64: base64.StdEncoding.EncodeToString(data)})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(text) == "" {
		t.Fatal("OCR returned empty text")
	}
}
