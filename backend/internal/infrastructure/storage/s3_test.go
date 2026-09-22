package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	stdurl "net/url"
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

func TestS3RejectsUnsafeEnvironmentPrefixAndIdentifiers(t *testing.T) {
	base := S3Config{Endpoint: "https://storage.example.test", Region: "us-east-1", Bucket: "receipts", Environment: "development", AccessKeyID: "access", SecretAccessKey: "secret"}
	for _, environment := range []string{"../prod", "dev\nname", "prod/blue"} {
		cfg := base
		cfg.Environment = environment
		if _, err := NewS3(cfg); err == nil {
			t.Fatalf("unsafe environment accepted: %q", environment)
		}
	}
	for _, prefix := range []string{"root/../outside", "root\\outside", "root\nname"} {
		cfg := base
		cfg.Prefix = prefix
		if _, err := NewS3(cfg); err == nil {
			t.Fatalf("unsafe prefix accepted: %q", prefix)
		}
	}
	for _, field := range []struct{ name, value string }{{"bucket", "mybucket/other"}, {"region", "us-east\n1"}} {
		cfg := base
		if field.name == "bucket" {
			cfg.Bucket = field.value
		} else {
			cfg.Region = field.value
		}
		if _, err := NewS3(cfg); err == nil {
			t.Fatalf("unsafe %s accepted", field.name)
		}
	}
	s, err := NewS3(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{"../owner", "owner/name", "owner\nname"} {
		if _, err := s.Key(owner, "batch-1", "attachment-1", "receipt.pdf"); err == nil {
			t.Fatalf("unsafe owner identifier accepted: %q", owner)
		}
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
	signedURL, err := s.SignedGet(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := stdurl.Parse(signedURL)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if parsed.Host != strings.TrimPrefix(server.URL, "http://") || parsed.Path != "/receipts/"+key {
		t.Fatalf("unexpected path-style presigned URL: %s", signedURL)
	}
	if query.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" || query.Get("X-Amz-Expires") != "300" || query.Get("X-Amz-SignedHeaders") != "host" || query.Get("X-Amz-Signature") == "" {
		t.Fatalf("unexpected presign query: %v", query)
	}
	if strings.Contains(signedURL, "secret") {
		t.Fatalf("presigned URL exposed secret key: %s", signedURL)
	}
}
