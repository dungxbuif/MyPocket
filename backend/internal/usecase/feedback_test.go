package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
)

type feedbackRepoStub struct {
	rows   map[string]entity.Feedback
	createErr error
}

func (s *feedbackRepoStub) Create(_ context.Context, row *entity.Feedback) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.rows[row.ID] = *row
	return nil
}
func (s *feedbackRepoStub) ListByOwner(_ context.Context, owner string) ([]entity.Feedback, error) {
	rows := make([]entity.Feedback, 0)
	for _, row := range s.rows {
		if row.UserID == owner {
			rows = append(rows, row)
		}
	}
	return rows, nil
}
func (s *feedbackRepoStub) FindByOwner(_ context.Context, owner, id string) (*entity.Feedback, error) {
	row, ok := s.rows[id]
	if !ok || row.UserID != owner {
		return nil, contract.ErrNotFound
	}
	return &row, nil
}
func (s *feedbackRepoStub) FindByID(_ context.Context, id string) (*entity.Feedback, error) {
	row, ok := s.rows[id]
	if !ok {
		return nil, contract.ErrNotFound
	}
	return &row, nil
}
func (s *feedbackRepoStub) ListForAgent(_ context.Context, status string, _ int) ([]entity.Feedback, error) {
	rows := make([]entity.Feedback, 0)
	for _, row := range s.rows {
		if status == "" || row.Status == status {
			rows = append(rows, row)
		}
	}
	return rows, nil
}
func (s *feedbackRepoStub) UpdateStatus(_ context.Context, id, status string, now time.Time) (*entity.Feedback, error) {
	row, ok := s.rows[id]
	if !ok {
		return nil, contract.ErrNotFound
	}
	row.Status, row.UpdatedAt = status, now
	s.rows[id] = row
	return &row, nil
}

type changelogRepoStub struct {
	feedback *feedbackRepoStub
	rows     map[string]entity.Changelog
	last     contract.AuditEvent
}

func (s *changelogRepoStub) Publish(_ context.Context, changelog *entity.Changelog, ids []string) error {
	for _, id := range ids {
		row, ok := s.feedback.rows[id]
		if !ok {
			return contract.ErrNotFound
		}
		if row.Status == entity.FeedbackStatusRejected || row.Status == entity.FeedbackStatusFixed {
			return contract.ErrFeedbackConflict
		}
		now := changelog.PublishedAt
		row.Status, row.FixedAt, row.ChangelogID, row.UpdatedAt = entity.FeedbackStatusFixed, &now, &changelog.ID, now
		s.feedback.rows[id] = row
	}
	s.rows[changelog.ID] = *changelog
	return nil
}
func (s *changelogRepoStub) ListPublic(context.Context, int) ([]entity.Changelog, error) {
	return nil, nil
}
func (s *changelogRepoStub) FindPublic(_ context.Context, id string) (*entity.Changelog, error) {
	row, ok := s.rows[id]
	if !ok {
		return nil, contract.ErrNotFound
	}
	return &row, nil
}

type auditStub struct{ events []contract.AuditEvent }

func (s *auditStub) Record(_ context.Context, event contract.AuditEvent) error {
	s.events = append(s.events, event)
	return nil
}

type feedbackStorageStub struct {
	key       string
	putCount  int
	deleteCount int
	signedCount int
	putErr    error
	signedURL string
}

func (s *feedbackStorageStub) Key(owner, batch, attachment, filename string) (string, error) {
	s.key = owner + "/" + batch + "/" + attachment + "/" + filename
	return s.key, nil
}
func (s *feedbackStorageStub) Put(context.Context, string, string, []byte) error {
	s.putCount++
	return s.putErr
}
func (s *feedbackStorageStub) Delete(context.Context, string) error {
	s.deleteCount++
	return nil
}
func (s *feedbackStorageStub) SignedGet(context.Context, string) (string, error) {
	s.signedCount++
	if s.signedURL == "" {
		s.signedURL = "https://storage.test/screenshot.png"
	}
	return s.signedURL, nil
}

func TestFeedbackCreateTrimsAndDefaultsOpen(t *testing.T) {
	feedback := &feedbackRepoStub{rows: map[string]entity.Feedback{}}
	service := NewFeedbackService(feedback, &auditStub{}, func() time.Time { return time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC) })
	row, err := service.Create(context.Background(), "owner", FeedbackInput{Type: entity.FeedbackTypeBug, Title: "  Wrong category ", Description: "  Details  "})
	if err != nil {
		t.Fatal(err)
	}
	if row.ID == "" || row.Status != entity.FeedbackStatusOpen || row.Title != "Wrong category" || row.Description != "Details" || row.UserID != "owner" {
		t.Fatalf("unexpected feedback: %+v", row)
	}
}

func TestFeedbackStatusRejectsDirectFixedAndTerminalTransitions(t *testing.T) {
	now := time.Now().UTC()
	feedback := &feedbackRepoStub{rows: map[string]entity.Feedback{"open": {ID: "open", UserID: "owner", Type: entity.FeedbackTypeBug, Title: "x", Description: "y", Status: entity.FeedbackStatusOpen, CreatedAt: now, UpdatedAt: now}, "fixed": {ID: "fixed", UserID: "owner", Type: entity.FeedbackTypeBug, Title: "x", Description: "y", Status: entity.FeedbackStatusFixed, CreatedAt: now, UpdatedAt: now}}}
	service := NewFeedbackService(feedback, &auditStub{}, time.Now)
	if _, err := service.SetStatus(context.Background(), "open", entity.FeedbackStatusFixed); !errors.Is(err, ErrFeedbackTransition) {
		t.Fatalf("direct fixed error=%v", err)
	}
	if _, err := service.SetStatus(context.Background(), "fixed", entity.FeedbackStatusInProgress); !errors.Is(err, ErrFeedbackTransition) {
		t.Fatalf("terminal transition error=%v", err)
	}
}

func TestFeedbackCreateStoresOptionalScreenshotAndMarksAvailability(t *testing.T) {
	feedback := &feedbackRepoStub{rows: map[string]entity.Feedback{}}
	storage := &feedbackStorageStub{}
	service := NewFeedbackServiceWithStorage(feedback, &auditStub{}, storage, func() time.Time { return time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC) })
	png := []byte("\x89PNG\r\n\x1a\nfeedback")
	row, err := service.Create(context.Background(), "owner", FeedbackInput{
		Type: entity.FeedbackTypeBug, Title: "screen", Description: "details",
		Screenshot: bytes.NewReader(png), ScreenshotMIME: "image/png", ScreenshotSize: int64(len(png)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if storage.putCount != 1 || row.ScreenshotObjectKey == "" || !row.ScreenshotAvailable {
		t.Fatalf("screenshot was not persisted: storage=%+v row=%+v", storage, row)
	}
}

func TestFeedbackCreateRejectsInvalidScreenshot(t *testing.T) {
	feedback := &feedbackRepoStub{rows: map[string]entity.Feedback{}}
	storage := &feedbackStorageStub{}
	service := NewFeedbackServiceWithStorage(feedback, &auditStub{}, storage, time.Now)
	_, err := service.Create(context.Background(), "owner", FeedbackInput{
		Type: entity.FeedbackTypeBug, Title: "screen", Description: "details",
		Screenshot: bytes.NewReader([]byte("not png")), ScreenshotMIME: "text/plain", ScreenshotSize: 7,
	})
	if !errors.Is(err, ErrFeedbackScreenshotInvalid) {
		t.Fatalf("expected invalid screenshot, got %v", err)
	}
	if storage.putCount != 0 {
		t.Fatalf("invalid screenshot must not be uploaded: %+v", storage)
	}
}

func TestFeedbackCreateDeletesScreenshotWhenRepositoryFails(t *testing.T) {
	feedback := &feedbackRepoStub{rows: map[string]entity.Feedback{}, createErr: errors.New("db down")}
	storage := &feedbackStorageStub{}
	service := NewFeedbackServiceWithStorage(feedback, &auditStub{}, storage, time.Now)
	png := []byte("\x89PNG\r\n\x1a\nfeedback")
	_, err := service.Create(context.Background(), "owner", FeedbackInput{
		Type: entity.FeedbackTypeBug, Title: "screen", Description: "details",
		Screenshot: bytes.NewReader(png), ScreenshotMIME: "image/png", ScreenshotSize: int64(len(png)),
	})
	if err == nil || storage.putCount != 1 || storage.deleteCount != 1 {
		t.Fatalf("expected cleanup after repository failure: err=%v storage=%+v", err, storage)
	}
}

func TestFeedbackScreenshotReturnsSignedURLOnlyForStoredScreenshot(t *testing.T) {
	feedback := &feedbackRepoStub{rows: map[string]entity.Feedback{
		"fb": {ID: "fb", UserID: "owner", ScreenshotObjectKey: "owners/owner/feedback/fb/screen.png", ScreenshotAvailable: true},
	}}
	storage := &feedbackStorageStub{}
	service := NewFeedbackServiceWithStorage(feedback, &auditStub{}, storage, time.Now)
	url, err := service.Screenshot(context.Background(), "owner", "fb")
	if err != nil || url != "https://storage.test/screenshot.png" || storage.signedCount != 1 {
		t.Fatalf("signed screenshot mismatch: url=%q err=%v storage=%+v", url, err, storage)
	}
}
