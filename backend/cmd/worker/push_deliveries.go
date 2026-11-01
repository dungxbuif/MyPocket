package main

import (
	"context"
	"errors"
	"time"

	"github.com/mypocket/backend/internal/infrastructure/push"
	"github.com/mypocket/backend/internal/repository"
)

const (
	pushDeliveryLease     = 2 * time.Minute
	pushDeliveryBatchSize = 4 // 4 × 15s provider timeout stays well inside the lease.
)

// runPushDeliveryPass first hands off inbox notifications to the durable
// queue, then drains only the configured APNs environment. All provider state
// changes are fenced by the lease returned from ClaimPushDeliveries.
func runPushDeliveryPass(ctx context.Context, deliveries repository.PushDeliveryRepository, tokens repository.PushTokenDeliveryRepository, transport push.Transport, environment string, now time.Time) (int, error) {
	if deliveries == nil || tokens == nil || transport == nil {
		return 0, nil
	}
	if _, err := deliveries.EnqueueNotificationDeliveries(ctx, now, 100); err != nil {
		return 0, err
	}
	jobs, err := deliveries.ClaimPushDeliveries(ctx, now, pushDeliveryBatchSize, pushDeliveryLease, environment)
	if err != nil {
		return 0, err
	}
	processed := 0
	var firstErr error
	for _, job := range jobs {
		row := job.Delivery
		token, tokenErr := tokens.LoadForDelivery(ctx, row.OwnerID, row.PushTokenID)
		if errors.Is(tokenErr, repository.ErrNotFound) {
			if markErr := deliveries.MarkPushDeliveryRevoked(ctx, row.ID, job.LeaseUntil, now); markErr != nil {
				firstErr = firstPushError(firstErr, markErr)
				continue
			}
			processed++
			continue
		}
		if tokenErr != nil {
			if markErr := deliveries.MarkPushDeliveryRetry(ctx, row.ID, job.LeaseUntil, now.Add(pushDeliveryRetryDelay(row.Attempts)), "token_lookup_error"); markErr != nil {
				firstErr = firstPushError(firstErr, markErr)
			} else {
				processed++
			}
			firstErr = firstPushError(firstErr, errors.New("push token lookup failed"))
			continue
		}
		if token.RevokedAt != nil || token.Token == "" || token.Environment != row.Environment {
			if markErr := deliveries.MarkPushDeliveryRevoked(ctx, row.ID, job.LeaseUntil, now); markErr != nil {
				firstErr = firstPushError(firstErr, markErr)
				continue
			}
			processed++
			continue
		}

		result, deliveryErr := push.Deliver(ctx, transport, push.Request{
			NotificationID: row.NotificationID,
			TokenID:        row.PushTokenID,
			DeviceToken:    token.Token,
			CollapseID:     row.CollapseID,
			Payload:        append([]byte(nil), row.Payload...),
		}, push.Attempt{Number: row.Attempts}, now)
		reason := pushDeliveryFailureReason(result, deliveryErr)
		switch result.State {
		case push.StateSent:
			if markErr := deliveries.MarkPushDeliverySent(ctx, row.ID, job.LeaseUntil, now); markErr != nil {
				firstErr = firstPushError(firstErr, markErr)
				continue
			}
		case push.StateRetry:
			retryAt := now.Add(pushDeliveryRetryDelay(row.Attempts))
			if result.RetryAt != nil {
				retryAt = result.RetryAt.UTC()
			}
			if markErr := deliveries.MarkPushDeliveryRetry(ctx, row.ID, job.LeaseUntil, retryAt, reason); markErr != nil {
				firstErr = firstPushError(firstErr, markErr)
				continue
			}
		case push.StateRevokeToken:
			if _, revokeErr := tokens.RevokeIfCurrent(ctx, row.OwnerID, row.PushTokenID, token.Token, now); revokeErr != nil {
				firstErr = firstPushError(firstErr, revokeErr)
				if markErr := deliveries.MarkPushDeliveryRetry(ctx, row.ID, job.LeaseUntil, now.Add(pushDeliveryRetryDelay(row.Attempts)), "token_revoke_error"); markErr != nil {
					firstErr = firstPushError(firstErr, markErr)
				}
				continue
			}
			if markErr := deliveries.MarkPushDeliveryRevoked(ctx, row.ID, job.LeaseUntil, now); markErr != nil {
				firstErr = firstPushError(firstErr, markErr)
				continue
			}
		case push.StateFailed:
			if markErr := deliveries.MarkPushDeliveryFailed(ctx, row.ID, job.LeaseUntil, now, reason); markErr != nil {
				firstErr = firstPushError(firstErr, markErr)
				continue
			}
		default:
			if markErr := deliveries.MarkPushDeliveryFailed(ctx, row.ID, job.LeaseUntil, now, "unknown_delivery_state"); markErr != nil {
				firstErr = firstPushError(firstErr, markErr)
				continue
			}
		}
		processed++
	}
	return processed, firstErr
}

func pushDeliveryFailureReason(result push.DeliveryResult, deliveryErr error) string {
	if result.FailureCode != "" {
		return result.FailureCode
	}
	if deliveryErr != nil {
		return "transport_error"
	}
	return "provider_error"
}

func pushDeliveryRetryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	if attempts >= 9 {
		return 5 * time.Minute
	}
	delay := time.Second << (attempts - 1)
	if delay > 5*time.Minute {
		return 5 * time.Minute
	}
	return delay
}

func firstPushError(first, next error) error {
	if first != nil {
		return first
	}
	return next
}
