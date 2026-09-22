package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

var (
	ErrFeedbackTransition = errors.New("feedback transition is not allowed")
	ErrFeedbackAgentInput = errors.New("feedback agent input is invalid")
)

type FeedbackInput struct {
	Type        string
	Title       string
	Description string
}

type FeedbackService struct {
	Feedback repository.FeedbackRepository
	Audit    repository.AuditSink
	Now      func() time.Time
}

func NewFeedbackService(feedback repository.FeedbackRepository, audit repository.AuditSink, now func() time.Time) *FeedbackService {
	if now == nil {
		now = time.Now
	}
	return &FeedbackService{Feedback: feedback, Audit: audit, Now: now}
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
	if err := s.Feedback.Create(ctx, row); err != nil {
		return nil, err
	}
	s.record(ctx, repository.AuditEvent{Action: "feedback.created", FeedbackID: row.ID})
	return row, nil
}

func (s *FeedbackService) List(ctx context.Context, owner string) ([]entity.Feedback, error) {
	return s.Feedback.ListByOwner(ctx, owner)
}

func (s *FeedbackService) Get(ctx context.Context, owner, id string) (*entity.Feedback, error) {
	return s.Feedback.FindByOwner(ctx, owner, id)
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
