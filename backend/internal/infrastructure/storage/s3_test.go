package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/config"
)

func TestS3UsesEnvironmentQualifiedPrivateKeys(t *testing.T) {
	s, err := NewS3(S3Config{Endpoint: "https://storage.example.test", Region: "us-east-1", Bucket: "receipts", Prefix: "mypocket/receipts", Environment: "staging", AccessKeyID: "access", SecretAccessKey: "secret", ForcePathStyle: true})
	if err != nil {
		t.Fatal(err)
	}
	key, err := s.Key("owner-1", "batch-1", "attachment-1", "bank statement.pdf")
	if err != nil {
		t.Fatal(err)
	}
	want := "mypocket/receipts/staging/owners/owner-1/batches/batch-1/attachment-1/bank-statement.pdf"
	if key != want {
		t.Fatalf("key = %q, want %q", key, want)
	}
	if _, err := s.Key("owner", "batch", "attachment", "../bank statement.pdf"); err == nil {
		t.Fatal("expected traversal filename rejection")
	}
}

func TestS3LivePDFWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("S3_LIVE_TEST") != "1" {
		t.Skip("set S3_LIVE_TEST=1 to exercise the configured private bucket")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal(err)
	}
	cfg := config.Load()
	file := os.Getenv("S3_LIVE_TEST_FILE")
	data, err := os.ReadFile(file)
	if err != nil || len(data) == 0 {
		t.Fatal("live PDF fixture could not be read")
	}
	s, err := NewS3(S3Config{Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, Bucket: cfg.S3Bucket, Prefix: cfg.S3Prefix, Environment: cfg.AppEnv, AccessKeyID: cfg.S3AccessKeyID, SecretAccessKey: cfg.S3SecretAccessKey, ForcePathStyle: cfg.S3ForcePathStyle})
	if err != nil {
		t.Fatal(err)
	}
	key, err := s.Key("uat-owner", "pdf-"+time.Now().UTC().Format("20060102150405"), "fixture", "receipt.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Put(context.Background(), key, "application/pdf", data); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Delete(context.Background(), key); err != nil {
			t.Errorf("cleanup private fixture: %v", err)
		}
	}()
	if _, err := s.SignedGet(context.Background(), key); err != nil {
		t.Fatal(err)
	}
}

func TestS3PutAndSignedGetNeverUsePublicURL(t *testing.T) {
	var gotPutPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPutPath, gotAuth = r.URL.Path, r.Header.Get("Authorization")
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	s, err := NewS3(S3Config{Endpoint: server.URL, Region: "us-east-1", Bucket: "receipts", Prefix: "root", Environment: "development", AccessKeyID: "access", SecretAccessKey: "secret", ForcePathStyle: true})
	if err != nil {
		t.Fatal(err)
	}
	key, _ := s.Key("owner", "batch", "attachment", "receipt.pdf")
	if err := s.Put(context.Background(), key, "application/pdf", []byte("pdf")); err != nil {
		t.Fatal(err)
	}
	if gotPutPath != "/receipts/"+key {
		t.Fatalf("path = %q", gotPutPath)
	}
	if !strings.HasPrefix(gotAuth, "AWS4-HMAC-SHA256 ") {
		t.Fatalf("missing sigv4 authorization: %q", gotAuth)
	}
	url, err := s.SignedGet(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(url, "X-Amz-Signature=") || strings.Contains(url, "secret") {
		t.Fatalf("unexpected signed URL: %s", url)
	}
}
