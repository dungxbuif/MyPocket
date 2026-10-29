package usecase

import (
	"context"
	"errors"
	"time"

	port "github.com/mypocket/backend/internal/repository"
)

type AIEntryAttachmentCleanup struct {
	Entries port.AIEntryAttachmentCleanupRepository
	Storage AttachmentStorage
	Now     func() time.Time
}

func (c *AIEntryAttachmentCleanup) Run(ctx context.Context, limit int) (int, error) {
	if c.Entries == nil || c.Storage == nil {
		return 0, ErrAIUnavailable
	}
	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	items, err := c.Entries.ClaimExpiredAttachments(ctx, now, now.Add(-10*time.Minute), limit)
	if err != nil {
		return 0, err
	}
	deleted := 0
	var failures []error
	for _, item := range items {
		status := "deleted"
		if err := c.Storage.Delete(ctx, item.ObjectKey); err != nil {
			status = "delete_failed"
			failures = append(failures, errors.New("one or more private attachments could not be deleted"))
		} else {
			deleted++
		}
		if err := c.Entries.SetAttachmentDeleteStatus(ctx, item.ID, status); err != nil {
			failures = append(failures, err)
		}
	}
	return deleted, errors.Join(failures...)
}
