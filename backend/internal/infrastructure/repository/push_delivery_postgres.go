package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	domain "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	pushDeliveryPending = "pending"
	pushDeliveryRetry   = "retry"
	pushDeliverySent    = "sent"
	pushDeliveryRevoked = "revoked"
	pushDeliveryFailed  = "failed"
)

type PushDeliveryPostgresRepository struct{ db *gorm.DB }

func NewPushDeliveryPostgresRepository(db *gorm.DB) domain.PushDeliveryRepository {
	return &PushDeliveryPostgresRepository{db: db}
}

type pushDeliveryCandidate struct {
	NotificationID string
	OwnerID        string
	PushTokenID    string
	Environment    string
}

// EnqueueNotificationDeliveries creates one app-owned route payload for every
// active token that does not already have a durable row for the notification.
// The query intentionally excludes notification title/body and financial data.
func (r *PushDeliveryPostgresRepository) EnqueueNotificationDeliveries(ctx context.Context, now time.Time, limit int) (int, error) {
	if r == nil || r.db == nil || limit <= 0 {
		return 0, gorm.ErrInvalidData
	}
	if limit > 100 {
		limit = 100
	}
	now = now.UTC().Truncate(time.Microsecond)
	created := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var candidates []pushDeliveryCandidate
		if err := tx.Raw(`
			SELECT n.id AS notification_id, n.owner_id, t.id AS push_token_id, t.environment
			FROM notifications n
			JOIN push_tokens t ON t.owner_id = n.owner_id AND t.revoked_at IS NULL
			LEFT JOIN push_deliveries d
			  ON d.notification_id = n.id
			 AND d.push_token_id = t.id
			 AND d.environment = t.environment
			WHERE n.created_at <= ? AND n.created_at >= t.created_at AND d.id IS NULL
			ORDER BY n.created_at ASC, n.id ASC, t.id ASC
			LIMIT ?`, now, limit).Scan(&candidates).Error; err != nil {
			return err
		}
		for _, candidate := range candidates {
			row := &entity.PushDelivery{
				ID:             uuid.NewString(),
				OwnerID:        candidate.OwnerID,
				NotificationID: candidate.NotificationID,
				PushTokenID:    candidate.PushTokenID,
				Environment:    candidate.Environment,
				CollapseID:     candidate.NotificationID,
				Payload:        NotificationPushPayload(candidate.NotificationID),
				Status:         pushDeliveryPending,
				AvailableAt:    now,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			result := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "notification_id"}, {Name: "push_token_id"}, {Name: "environment"}},
				DoNothing: true,
			}).Create(row)
			if result.Error != nil {
				return result.Error
			}
			created += int(result.RowsAffected)
		}
		return nil
	})
	return created, err
}

func NotificationPushPayload(notificationID string) []byte {
	payload, _ := json.Marshal(struct {
		APS      map[string]int `json:"aps"`
		MyPocket struct {
			Route          string `json:"route"`
			NotificationID string `json:"notification_id"`
		} `json:"mypocket"`
	}{
		APS: map[string]int{"content-available": 1},
		MyPocket: struct {
			Route          string `json:"route"`
			NotificationID string `json:"notification_id"`
		}{Route: "notifications", NotificationID: notificationID},
	})
	return payload
}

func (r *PushDeliveryPostgresRepository) ClaimPushDeliveries(ctx context.Context, now time.Time, limit int, lease time.Duration, environment string) ([]domain.PushDeliveryJob, error) {
	if r == nil || r.db == nil || limit <= 0 || lease <= 0 || strings.TrimSpace(environment) == "" {
		return nil, gorm.ErrInvalidData
	}
	if limit > 100 {
		limit = 100
	}
	now = now.UTC().Truncate(time.Microsecond)
	var jobs []domain.PushDeliveryJob
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []entity.PushDelivery
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where(
			"environment = ? AND status IN ? AND available_at <= ? AND (lease_until IS NULL OR lease_until <= ?)",
			environment, []string{pushDeliveryPending, pushDeliveryRetry}, now, now,
		).Order("available_at ASC, created_at ASC, id ASC").Limit(limit).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		until := now.Add(lease)
		ids := make([]string, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		if err := tx.Model(&entity.PushDelivery{}).Where("id IN ?", ids).Updates(map[string]any{
			"lease_until": until,
			"attempts":    gorm.Expr("attempts + 1"),
			"updated_at":  now,
		}).Error; err != nil {
			return err
		}
		jobs = make([]domain.PushDeliveryJob, 0, len(rows))
		for _, row := range rows {
			row.Payload = append([]byte(nil), row.Payload...)
			row.Attempts++
			jobs = append(jobs, domain.PushDeliveryJob{Delivery: row, LeaseUntil: until})
		}
		return nil
	})
	return jobs, err
}

func (r *PushDeliveryPostgresRepository) MarkPushDeliverySent(ctx context.Context, id string, leaseUntil, now time.Time) error {
	return r.updateLeased(ctx, id, leaseUntil, map[string]any{
		"status":       pushDeliverySent,
		"lease_until":  nil,
		"delivered_at": now.UTC(),
		"last_error":   nil,
		"updated_at":   now.UTC(),
	})
}

func (r *PushDeliveryPostgresRepository) MarkPushDeliveryRetry(ctx context.Context, id string, leaseUntil, availableAt time.Time, reason string) error {
	return r.updateLeased(ctx, id, leaseUntil, map[string]any{
		"status":       pushDeliveryRetry,
		"available_at": availableAt.UTC(),
		"lease_until":  nil,
		"last_error":   safePushDeliveryReason(reason),
		"updated_at":   availableAt.UTC(),
	})
}

func (r *PushDeliveryPostgresRepository) MarkPushDeliveryRevoked(ctx context.Context, id string, leaseUntil, now time.Time) error {
	return r.updateLeased(ctx, id, leaseUntil, map[string]any{
		"status":      pushDeliveryRevoked,
		"lease_until": nil,
		"last_error":  nil,
		"updated_at":  now.UTC(),
	})
}

func (r *PushDeliveryPostgresRepository) MarkPushDeliveryFailed(ctx context.Context, id string, leaseUntil, now time.Time, reason string) error {
	return r.updateLeased(ctx, id, leaseUntil, map[string]any{
		"status":      pushDeliveryFailed,
		"lease_until": nil,
		"last_error":  safePushDeliveryReason(reason),
		"updated_at":  now.UTC(),
	})
}

func (r *PushDeliveryPostgresRepository) updateLeased(ctx context.Context, id string, leaseUntil time.Time, updates map[string]any) error {
	if r == nil || r.db == nil || strings.TrimSpace(id) == "" || leaseUntil.IsZero() {
		return gorm.ErrInvalidData
	}
	result := r.db.WithContext(ctx).Model(&entity.PushDelivery{}).Where(
		"id = ? AND lease_until = ? AND status IN ?", id, leaseUntil.UTC(), []string{pushDeliveryPending, pushDeliveryRetry},
	).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPushDeliveryLeaseLost
	}
	return nil
}

func safePushDeliveryReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" || len(reason) > 128 {
		return "provider_error"
	}
	for _, char := range reason {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' && char != '-' && char != '.' {
			return "provider_error"
		}
	}
	return reason
}

func (r *PushTokenPostgresRepository) LoadForDelivery(ctx context.Context, owner, id string) (entity.PushToken, error) {
	var item entity.PushToken
	if r == nil || r.db == nil || strings.TrimSpace(owner) == "" || strings.TrimSpace(id) == "" {
		return item, gorm.ErrInvalidData
	}
	err := r.db.WithContext(ctx).Where("owner_id = ? AND id = ?", owner, id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, domain.ErrNotFound
	}
	return item, err
}

func (r *PushTokenPostgresRepository) RevokeIfCurrent(ctx context.Context, owner, id, token string, now time.Time) (bool, error) {
	if r == nil || r.db == nil || strings.TrimSpace(owner) == "" || strings.TrimSpace(id) == "" || token == "" {
		return false, gorm.ErrInvalidData
	}
	result := r.db.WithContext(ctx).Model(&entity.PushToken{}).Where(
		"owner_id = ? AND id = ? AND token = ? AND revoked_at IS NULL", owner, id, token,
	).Updates(map[string]any{"revoked_at": now.UTC(), "updated_at": now.UTC()})
	return result.RowsAffected > 0, result.Error
}

var _ domain.PushDeliveryRepository = (*PushDeliveryPostgresRepository)(nil)
var _ domain.PushTokenDeliveryRepository = (*PushTokenPostgresRepository)(nil)
