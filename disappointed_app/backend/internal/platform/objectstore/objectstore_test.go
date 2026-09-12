package objectstore_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"mypocket/internal/platform/config"
	"mypocket/internal/platform/objectstore"
)

func TestS3SmokeErrorsDoNotExposeCredentials(t *testing.T) {
	store, err := objectstore.NewS3(config.Config{
		S3Endpoint:  "http://127.0.0.1:1",
		S3Bucket:    "mypocket",
		S3AccessKey: "access-key-value",
		S3SecretKey: "secret-key-value",
	})
	if err != nil {
		t.Fatalf("new s3 store: %v", err)
	}

	err = store.PutSmokeObject(context.Background())
	if err == nil {
		t.Fatal("expected connection failure")
	}
	message := err.Error()
	if strings.Contains(message, "access-key-value") || strings.Contains(message, "secret-key-value") {
		t.Fatalf("object store error leaked credentials: %v", err)
	}
}

func TestS3PresignedURLsUseConfiguredEndpoint(t *testing.T) {
	store, err := objectstore.NewS3(config.Config{S3Endpoint: "https://storage.example.test", S3Bucket: "my-pocket", S3AccessKey: "access", S3SecretKey: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	putURL, err := store.PresignPut(context.Background(), "users/u/receipt.jpg", "image/jpeg", 15*time.Minute)
	if err != nil || !strings.HasPrefix(putURL, "https://storage.example.test/") {
		t.Fatalf("unexpected presigned put URL: %q (%v)", putURL, err)
	}
	getURL, err := store.PresignGet(context.Background(), "users/u/receipt.jpg", 10*time.Minute)
	if err != nil || !strings.HasPrefix(getURL, "https://storage.example.test/") {
		t.Fatalf("unexpected presigned get URL: %q (%v)", getURL, err)
	}
}

func TestS3SmokeObjectLifecycle(t *testing.T) {
	cfg := loadSmokeConfig(t)
	ensureSmokeBucket(t, cfg)
	store, err := objectstore.NewS3(cfg)
	if err != nil {
		t.Fatalf("new s3 store: %v", err)
	}

	if err := store.PutSmokeObject(context.Background()); err != nil {
		t.Fatalf("put smoke object: %v", err)
	}
	if err := store.DeleteSmokeObject(context.Background()); err != nil {
		t.Fatalf("delete smoke object: %v", err)
	}
	payload := []byte("dataset,id\nwallets,w1\n")
	key := "users/test/exports/lifecycle.csv"
	if err := store.PutObject(context.Background(), key, "text/csv", bytes.NewReader(payload), int64(len(payload))); err != nil {
		t.Fatalf("put lifecycle object: %v", err)
	}
	body, size, err := store.GetObject(context.Background(), key, 1024)
	if err != nil {
		t.Fatalf("get lifecycle object: %v", err)
	}
	defer body.Close()
	got, err := io.ReadAll(body)
	if err != nil || size != int64(len(payload)) || !bytes.Equal(got, payload) {
		t.Fatalf("lifecycle object mismatch size=%d body=%q err=%v", size, got, err)
	}
	if err := store.DeleteObject(context.Background(), key); err != nil {
		t.Fatalf("delete lifecycle object: %v", err)
	}
}

func ensureSmokeBucket(t *testing.T, cfg config.Config) {
	t.Helper()
	if os.Getenv("MYPOCKET_TEST_S3_SKIP_BUCKET_CREATE") == "true" {
		return
	}

	awsConfig, err := awscfg.LoadDefaultConfig(context.Background(),
		awscfg.WithRegion("us-east-1"),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
	)
	if err != nil {
		t.Fatalf("configure smoke bucket client: %v", err)
	}
	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.S3Endpoint)
		options.UsePathStyle = true
	})
	_, err = client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(cfg.S3Bucket),
	})
	if err == nil {
		return
	}
	var apiErr smithy.APIError
	if ok := errors.As(err, &apiErr); ok && (apiErr.ErrorCode() == "BucketAlreadyOwnedByYou" || apiErr.ErrorCode() == "BucketAlreadyExists") {
		return
	}
	t.Fatalf("create smoke bucket: %v", err)
}

func loadSmokeConfig(t *testing.T) config.Config {
	t.Helper()

	endpoint := os.Getenv("MYPOCKET_TEST_S3_ENDPOINT")
	bucket := os.Getenv("MYPOCKET_TEST_S3_BUCKET")
	accessKey := os.Getenv("MYPOCKET_TEST_S3_ACCESS_KEY")
	secretKey := os.Getenv("MYPOCKET_TEST_S3_SECRET_KEY")
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		t.Skip("MYPOCKET_TEST_S3_* is not set; S3-compatible smoke proof skipped")
	}

	return config.Config{
		S3Endpoint:  endpoint,
		S3Bucket:    bucket,
		S3AccessKey: accessKey,
		S3SecretKey: secretKey,
	}
}
