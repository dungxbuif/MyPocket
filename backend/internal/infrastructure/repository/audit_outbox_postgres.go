package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuditOutboxPostgresRepository struct{ db *gorm.DB }

func NewAuditOutboxPostgresRepository(db *gorm.DB) contract.AuditOutboxRepository {
	return &AuditOutboxPostgresRepository{db: db}
}

func (r *AuditOutboxPostgresRepository) EnqueueAudit(ctx context.Context, stream, eventID string, payload []byte, availableAt time.Time) error {
	if r == nil || r.db == nil || strings.TrimSpace(stream) == "" || strings.TrimSpace(eventID) == "" || len(payload) == 0 {
		return gorm.ErrInvalidData
	}
	row := &entity.AuditOutbox{ID: uuid.NewString(), Stream: stream, EventID: eventID, Payload: append([]byte(nil), payload...), AvailableAt: availableAt.UTC()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "stream"}, {Name: "event_id"}}, DoNothing: true}).Create(row).Error
}

func (r *AuditOutboxPostgresRepository) ClaimAudit(ctx context.Context, now time.Time, limit int, lease time.Duration) ([]contract.AuditOutboxEntry, error) {
	if r == nil || r.db == nil || limit <= 0 || lease <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if limit > 500 {
		limit = 500
	}
	var out []contract.AuditOutboxEntry
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []entity.AuditOutbox
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where(
			"delivered_at IS NULL AND available_at <= ? AND (lease_until IS NULL OR lease_until <= ?)", now.UTC(), now.UTC(),
		).Order("available_at ASC, created_at ASC, id ASC").Limit(limit).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		until := now.UTC().Add(lease)
		ids := make([]string, 0, len(rows))
		for i := range rows {
			ids = append(ids, rows[i].ID)
		}
		if err := tx.Model(&entity.AuditOutbox{}).Where("id IN ?", ids).Updates(map[string]any{"lease_until": until, "attempts": gorm.Expr("attempts + 1")}).Error; err != nil {
			return err
		}
		out = make([]contract.AuditOutboxEntry, 0, len(rows))
		for _, row := range rows {
			out = append(out, contract.AuditOutboxEntry{ID: row.ID, Stream: row.Stream, EventID: row.EventID, Payload: append([]byte(nil), row.Payload...), Attempts: row.Attempts + 1, AvailableAt: row.AvailableAt, LeaseUntil: &until})
		}
		return nil
	})
	return out, err
}

func (r *AuditOutboxPostgresRepository) MarkAuditDelivered(ctx context.Context, id string, now time.Time) error {
	if strings.TrimSpace(id) == "" {
		return gorm.ErrInvalidData
	}
	result := r.db.WithContext(ctx).Model(&entity.AuditOutbox{}).Where("id = ? AND delivered_at IS NULL", id).Updates(map[string]any{"delivered_at": now.UTC(), "lease_until": nil, "last_error": nil})
	return result.Error
}

func (r *AuditOutboxPostgresRepository) MarkAuditFailed(ctx context.Context, id string, availableAt time.Time, reason string) error {
	if strings.TrimSpace(id) == "" {
		return gorm.ErrInvalidData
	}
	reason = strings.TrimSpace(reason)
	if len(reason) > 1000 {
		reason = reason[:1000]
	}
	result := r.db.WithContext(ctx).Model(&entity.AuditOutbox{}).Where("id = ? AND delivered_at IS NULL", id).Updates(map[string]any{"available_at": availableAt.UTC(), "lease_until": nil, "last_error": reason})
	return result.Error
}

var _ contract.AuditOutboxRepository = (*AuditOutboxPostgresRepository)(nil)
