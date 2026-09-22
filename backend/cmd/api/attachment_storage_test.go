package main

import (
	"testing"

	"github.com/mypocket/backend/internal/config"
)

func TestAttachmentStorageIsOptionalOnlyInDevelopment(t *testing.T) {
	if store, err := newAttachmentStorage(config.Config{AppEnv: "development"}); err != nil || store != nil {
		t.Fatalf("empty development storage should disable files without failing: store=%v err=%v", store, err)
	}
	if _, err := newAttachmentStorage(config.Config{AppEnv: "production"}); err == nil {
		t.Fatal("production must reject startup without private attachment storage")
	}
	store, err := newAttachmentStorage(config.Config{AppEnv: "production", S3Endpoint: "https://storage.example.test", S3Region: "us-east-1", S3Bucket: "mypocket", S3AccessKeyID: "access", S3SecretAccessKey: "secret"})
	if err != nil || store == nil {
		t.Fatalf("complete storage config should initialize: store=%v err=%v", store, err)
	}
}
