package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/lifecycle"
	"mypocket/internal/platform/config"
)

type lifecycleRepoStub struct{ jobs map[string]lifecycle.Job }

func (s *lifecycleRepoStub) CreateJob(_ context.Context, user string, kind lifecycle.Kind, _ string, request any) (lifecycle.Job, error) {
	body, _ := json.Marshal(request)
	job := lifecycle.Job{ID: "job-1", UserID: user, Kind: kind, Status: lifecycle.StatusQueued, Request: body, Version: 1}
	s.jobs[job.ID] = job
	return job, nil
}
func (s *lifecycleRepoStub) GetJob(_ context.Context, user, id string) (lifecycle.Job, error) {
	job, ok := s.jobs[id]
	if !ok || job.UserID != user {
		return lifecycle.Job{}, lifecycle.ErrNotFound
	}
	return job, nil
}
func (s *lifecycleRepoStub) ConfirmImport(_ context.Context, user, id string, version int64) (lifecycle.Job, error) {
	job, err := s.GetJob(context.Background(), user, id)
	if err != nil {
		return lifecycle.Job{}, err
	}
	if version != job.Version {
		return lifecycle.Job{}, lifecycle.ErrConflict
	}
	job.Status = lifecycle.StatusQueued
	job.Version++
	s.jobs[id] = job
	return job, nil
}
func (s *lifecycleRepoStub) Counts(context.Context, string) (map[string]int, error) {
	return map[string]int{"wallets": 1}, nil
}
func (s *lifecycleRepoStub) DisableUser(context.Context, string) error { return nil }

type lifecycleStoreStub struct{}

func (lifecycleStoreStub) PutObject(context.Context, string, string, io.Reader, int64) error {
	return nil
}
func (lifecycleStoreStub) GetObject(context.Context, string, int64) (io.ReadCloser, int64, error) {
	return io.NopCloser(strings.NewReader("")), 0, nil
}
func (lifecycleStoreStub) PresignGet(context.Context, string, time.Duration) (string, error) {
	return "https://storage.example/download", nil
}

func TestLifecycleExportRequiresIdempotencyAndCreatesOwnedJob(t *testing.T) {
	repo := &lifecycleRepoStub{jobs: map[string]lifecycle.Job{}}
	handler := exports(testLifecycleConfig(), repo)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/exports", strings.NewReader(`{"datasets":["transactions"]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "export-1")
	req = req.WithContext(withAuthenticatedUser(req.Context(), "user-a", "api_key"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if repo.jobs["job-1"].UserID != "user-a" {
		t.Fatal("job owner mismatch")
	}
}

func TestLifecycleRoutesHideForeignJobsAndObjects(t *testing.T) {
	repo := &lifecycleRepoStub{jobs: map[string]lifecycle.Job{"foreign": {ID: "foreign", UserID: "user-b", Kind: lifecycle.KindExport, Status: lifecycle.StatusCompleted, ResultObjectKey: "users/user-b/exports/foreign.csv"}}}
	handler := exportByID(testLifecycleConfig(), repo, lifecycleStoreStub{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/exports/foreign/download", nil)
	req = req.WithContext(withAuthenticatedUser(req.Context(), "user-a", "api_key"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign download status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestImportRejectsForeignObjectPrefix(t *testing.T) {
	repo := &lifecycleRepoStub{jobs: map[string]lifecycle.Job{}}
	handler := imports(testLifecycleConfig(), repo, lifecycleStoreStub{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/imports", strings.NewReader(`{"object_key":"users/user-b/imports/data.csv"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "import-1")
	req = req.WithContext(withAuthenticatedUser(req.Context(), "user-a", "api_key"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func testLifecycleConfig() config.Config {
	return config.Config{CookieSecret: strings.Repeat("c", 32), CSRFSecret: strings.Repeat("s", 32)}
}
