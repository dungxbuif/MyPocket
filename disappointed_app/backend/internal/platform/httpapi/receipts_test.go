package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/identity"
	"mypocket/internal/platform/httpapi"
)

func TestReceiptUploadAndDownloadAreUserScoped(t *testing.T) {
	cfg := authTestConfig()
	repo := &receiptRepoStub{}
	store := &receiptStoreStub{}
	identityRepo := &authRepoStub{user: identity.User{ID: "user_123", Email: "owner@example.com", EmailVerified: true}}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: identityRepo, ReceiptRepository: repo, ObjectStore: store})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/presign", strings.NewReader(`{"filename":"bill.jpg","content_type":"image/jpeg","size_bytes":123,"checksum_sha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`))
	addAuthCookie(t, req, cfg.CookieSecret, "user_123")
	addCSRF(req)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated || repo.created.UserID != "user_123" || store.putKey == "" {
		t.Fatalf("unexpected receipt upload: %d %s", res.Code, res.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/receipt_1/download", nil)
	addAuthCookie(t, req, cfg.CookieSecret, "user_123")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK || store.getKey != store.putKey {
		t.Fatalf("unexpected receipt download: %d %s", res.Code, res.Body.String())
	}
}

func TestReceiptPresignUsesAPIKeyOwner(t *testing.T) {
	cfg := authTestConfig()
	cfg.APIKeyHashSecret = "change-this-development-api-key-hash-secret-32-bytes"
	identityRepo := &authRepoStub{user: identity.User{ID: "user_123", Email: "agent@example.com", EmailVerified: true}}
	repo := &receiptRepoStub{}
	handler := httpapi.NewRouter(cfg, httpapi.Dependencies{IdentityRepository: identityRepo, APIKeyRepository: identityRepo, ReceiptRepository: repo, ObjectStore: &receiptStoreStub{}, AuthCache: &authCacheStub{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/presign", strings.NewReader(`{"filename":"bill.jpg","content_type":"image/jpeg","size_bytes":123,"checksum_sha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`))
	req.Header.Set("Authorization", "Bearer mpk_test")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated || repo.created.UserID != "user_123" {
		t.Fatalf("expected receipt scoped to API key owner, code=%d user=%q body=%s", res.Code, repo.created.UserID, res.Body.String())
	}
}

type receiptRepoStub struct{ created finance.ReceiptObject }

func (r *receiptRepoStub) CreateReceiptObject(_ context.Context, userID string, input finance.CreateReceiptObjectInput) (finance.ReceiptObject, error) {
	r.created = finance.ReceiptObject{ID: "receipt_1", UserID: userID, ObjectKey: input.ObjectKey, ContentType: input.ContentType, SizeBytes: input.SizeBytes, ChecksumSHA256: input.ChecksumSHA256, OriginalFilename: input.OriginalFilename, CreatedAt: time.Now()}
	return r.created, nil
}
func (r *receiptRepoStub) GetReceiptObject(_ context.Context, userID, id string) (finance.ReceiptObject, error) {
	if id != r.created.ID || userID != r.created.UserID {
		return finance.ReceiptObject{}, finance.ErrForbidden
	}
	return r.created, nil
}

type receiptStoreStub struct{ putKey, getKey string }

func (s *receiptStoreStub) PresignPut(_ context.Context, key, _ string, _ time.Duration) (string, error) {
	s.putKey = key
	return "https://storage.test/upload", nil
}
func (s *receiptStoreStub) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	s.getKey = key
	return "https://storage.test/download", nil
}
func (s *receiptStoreStub) DeleteObject(context.Context, string) error { return nil }
