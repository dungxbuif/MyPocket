package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type auditContextKey string

const (
	auditRequestIDKey auditContextKey = "mypocket.audit.request_id"
	auditActorKey     auditContextKey = "mypocket.audit.actor_kind"
)

func WithAuditContext(ctx context.Context, requestID, actorKind string) context.Context {
	ctx = context.WithValue(ctx, auditRequestIDKey, strings.TrimSpace(requestID))
	return context.WithValue(ctx, auditActorKey, strings.TrimSpace(actorKind))
}

func auditContextValues(ctx context.Context) (string, string) {
	requestID, _ := ctx.Value(auditRequestIDKey).(string)
	actor, _ := ctx.Value(auditActorKey).(string)
	return requestID, actor
}

var (
	ErrFeedbackTransition        = errors.New("feedback transition is not allowed")
	ErrFeedbackAgentInput        = errors.New("feedback agent input is invalid")
	ErrFeedbackScreenshotInvalid = errors.New("feedback screenshot is invalid")
	ErrFeedbackStorage           = errors.New("feedback screenshot storage is unavailable")
)

const maxFeedbackScreenshotBytes int64 = 1 << 20

type FeedbackInput struct {
	Type           string
	Title          string
	Description    string
	Screenshot     io.Reader
	ScreenshotMIME string
	ScreenshotSize int64
}

type FeedbackStorage interface {
	Key(owner, batch, attachment, filename string) (string, error)
	Put(context.Context, string, string, []byte) error
	Delete(context.Context, string) error
	SignedGet(context.Context, string) (string, error)
}

type FeedbackService struct {
	Feedback repository.FeedbackRepository
	Audit    repository.AuditSink
	Storage  FeedbackStorage
	Now      func() time.Time
}

func NewFeedbackService(feedback repository.FeedbackRepository, audit repository.AuditSink, now func() time.Time) *FeedbackService {
	return NewFeedbackServiceWithStorage(feedback, audit, nil, now)
}

func NewFeedbackServiceWithStorage(feedback repository.FeedbackRepository, audit repository.AuditSink, storage FeedbackStorage, now func() time.Time) *FeedbackService {
	if now == nil {
		now = time.Now
	}
	return &FeedbackService{Feedback: feedback, Audit: audit, Storage: storage, Now: now}
}

func (s *FeedbackService) Create(ctx context.Context, owner string, input FeedbackInput) (*entity.Feedback, error) {
	now := s.Now().UTC()
	row := &entity.Feedback{ID: uuid.NewString(), UserID: strings.TrimSpace(owner), Type: strings.TrimSpace(input.Type), Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description), Status: entity.FeedbackStatusOpen, CreatedAt: now, UpdatedAt: now}
	if row.UserID == "" {
		return nil, repository.ErrFeedbackInvalid
	}
	if err := row.Validate(); err != nil {
		return nil, err
	}
	var screenshotKey string
	if input.Screenshot != nil {
		if s.Storage == nil || input.ScreenshotMIME != "image/png" || input.ScreenshotSize <= 0 || input.ScreenshotSize > maxFeedbackScreenshotBytes {
			return nil, ErrFeedbackScreenshotInvalid
		}
		data, err := io.ReadAll(io.LimitReader(input.Screenshot, maxFeedbackScreenshotBytes+1))
		if err != nil || len(data) == 0 || int64(len(data)) > maxFeedbackScreenshotBytes {
			return nil, ErrFeedbackScreenshotInvalid
		}
		screenshotKey, err = s.Storage.Key(row.UserID, "feedback", row.ID, "screen.png")
		if err != nil {
			return nil, ErrFeedbackScreenshotInvalid
		}
		if err := s.Storage.Put(ctx, screenshotKey, input.ScreenshotMIME, data); err != nil {
			return nil, ErrFeedbackStorage
		}
		nowScreenshot := now
		row.ScreenshotObjectKey = screenshotKey
		row.ScreenshotMIMEType = input.ScreenshotMIME
		row.ScreenshotSizeBytes = int64(len(data))
		row.ScreenshotCreatedAt = &nowScreenshot
		row.ScreenshotAvailable = true
	}
	if err := s.Feedback.Create(ctx, row); err != nil {
		if screenshotKey != "" && s.Storage != nil {
			_ = s.Storage.Delete(context.WithoutCancel(ctx), screenshotKey)
		}
		return nil, err
	}
	s.record(ctx, repository.AuditEvent{Action: "feedback.created", FeedbackID: row.ID})
	return row, nil
}

func (s *FeedbackService) List(ctx context.Context, owner string) ([]entity.Feedback, error) {
	rows, err := s.Feedback.ListByOwner(ctx, owner)
	for i := range rows {
		rows[i].ScreenshotAvailable = rows[i].ScreenshotObjectKey != ""
	}
	return rows, err
}

func (s *FeedbackService) Get(ctx context.Context, owner, id string) (*entity.Feedback, error) {
	row, err := s.Feedback.FindByOwner(ctx, owner, id)
	if row != nil {
		row.ScreenshotAvailable = row.ScreenshotObjectKey != ""
	}
	return row, err
}

func (s *FeedbackService) Screenshot(ctx context.Context, owner, id string) (string, error) {
	row, err := s.Get(ctx, owner, id)
	if err != nil {
		return "", err
	}
	return s.signedScreenshot(ctx, row)
}

func (s *FeedbackService) AgentScreenshot(ctx context.Context, id string) (string, error) {
	row, err := s.Feedback.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return s.signedScreenshot(ctx, row)
}

func (s *FeedbackService) signedScreenshot(ctx context.Context, row *entity.Feedback) (string, error) {
	if s.Storage == nil || row == nil || row.ScreenshotObjectKey == "" {
		return "", repository.ErrNotFound
	}
	return s.Storage.SignedGet(ctx, row.ScreenshotObjectKey)
}

func (s *FeedbackService) AgentList(ctx context.Context, status string, limit int) ([]entity.Feedback, error) {
	status = strings.TrimSpace(status)
	if status != "" && !isFeedbackStatus(status) {
		return nil, ErrFeedbackAgentInput
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.Feedback.ListForAgent(ctx, status, limit)
	if err == nil {
		s.record(ctx, repository.AuditEvent{Action: "feedback.agent_read", FeedbackIDs: feedbackIDs(rows)})
	}
	return rows, err
}

func (s *FeedbackService) SetStatus(ctx context.Context, id, status string) (*entity.Feedback, error) {
	status = strings.TrimSpace(status)
	if status == entity.FeedbackStatusFixed || !isFeedbackStatus(status) {
		return nil, ErrFeedbackTransition
	}
	row, err := s.Feedback.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !entity.CanTransitionFeedback(row.Status, status) {
		return nil, ErrFeedbackTransition
	}
	updated, err := s.Feedback.UpdateStatus(ctx, id, status, s.Now().UTC())
	if err != nil {
		return nil, err
	}
	s.record(ctx, repository.AuditEvent{Action: "feedback.status_changed", FeedbackID: id, OldStatus: row.Status, NewStatus: status})
	return updated, nil
}

func (s *FeedbackService) record(ctx context.Context, event repository.AuditEvent) {
	event.RequestID, event.ActorKind = auditContextValues(ctx)
	if s.Audit != nil {
		_ = s.Audit.Record(ctx, event)
	}
}

func isFeedbackStatus(status string) bool {
	switch status {
	case entity.FeedbackStatusOpen, entity.FeedbackStatusTriaged, entity.FeedbackStatusInProgress, entity.FeedbackStatusFixed, entity.FeedbackStatusRejected:
		return true
	default:
		return false
	}
}

func feedbackIDs(rows []entity.Feedback) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}
