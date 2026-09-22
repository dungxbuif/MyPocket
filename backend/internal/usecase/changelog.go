package usecase

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type ChangelogInput struct {
	FeedbackIDs []string
	Version     string
	Title       string
	Description string
}

type ChangelogService struct {
	Feedback   repository.FeedbackRepository
	Changelogs repository.ChangelogRepository
	Audit      repository.AuditSink
	Now        func() time.Time
}

func NewChangelogService(feedback repository.FeedbackRepository, changelogs repository.ChangelogRepository, audit repository.AuditSink, now func() time.Time) *ChangelogService {
	if now == nil {
		now = time.Now
	}
	return &ChangelogService{Feedback: feedback, Changelogs: changelogs, Audit: audit, Now: now}
}

func (s *ChangelogService) Publish(ctx context.Context, input ChangelogInput) (*entity.Changelog, error) {
	ids := uniqueIDs(input.FeedbackIDs)
	row := &entity.Changelog{ID: uuid.NewString(), Version: strings.TrimSpace(input.Version), Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description), PublishedAt: s.Now().UTC()}
	if len(ids) == 0 {
		return nil, repository.ErrFeedbackInvalid
	}
	if err := row.Validate(); err != nil {
		return nil, err
	}
	if err := s.Changelogs.Publish(ctx, row, ids); err != nil {
		return nil, err
	}
	if s.Audit != nil {
		sorted := append([]string(nil), ids...)
		sort.Strings(sorted)
		_ = s.Audit.Record(ctx, repository.AuditEvent{Action: "changelog.published", ChangelogID: row.ID, FeedbackIDs: sorted, Version: row.Version})
	}
	return row, nil
}

func (s *ChangelogService) List(ctx context.Context, limit int) ([]entity.Changelog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.Changelogs.ListPublic(ctx, limit)
}

func (s *ChangelogService) Get(ctx context.Context, id string) (*entity.Changelog, error) {
	return s.Changelogs.FindPublic(ctx, id)
}

func uniqueIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
