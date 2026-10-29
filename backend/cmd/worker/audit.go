package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/mypocket/backend/internal/infrastructure/cache"
	"github.com/mypocket/backend/internal/repository"
)

type auditPublisher interface {
	PublishAudit(context.Context, string, string, []byte) error
}

// runAuditOutboxPass drains a bounded leased batch. A failed item is delayed
// with exponential backoff, while other owners/events continue independently.
func runAuditOutboxPass(ctx context.Context, outbox repository.AuditOutboxRepository, publisher auditPublisher, now time.Time) (int, error) {
	if outbox == nil || publisher == nil {
		return 0, nil
	}
	rows, err := outbox.ClaimAudit(ctx, now, 100, 2*time.Minute)
	if err != nil {
		return 0, err
	}
	var firstErr error
	delivered := 0
	for _, row := range rows {
		if err := publisher.PublishAudit(ctx, row.Stream, row.EventID, row.Payload); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("audit %s delivery failed: %w", row.ID, err)
			}
			if markErr := outbox.MarkAuditFailed(ctx, row.ID, now.Add(auditRetryDelay(row.Attempts)), err.Error()); markErr != nil && firstErr == nil {
				firstErr = markErr
			}
			continue
		}
		if err := outbox.MarkAuditDelivered(ctx, row.ID, now); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		delivered++
	}
	return delivered, firstErr
}

func auditRetryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	if attempts >= 9 {
		return 5 * time.Minute
	}
	delay := time.Duration(math.Pow(2, float64(attempts-1))) * time.Second
	if delay > 5*time.Minute {
		return 5 * time.Minute
	}
	return delay
}

var _ auditPublisher = (*cache.Redis)(nil)
