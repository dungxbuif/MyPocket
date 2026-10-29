package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationPostgresRepository struct{ db *gorm.DB }

func NewNotificationPostgresRepository(db *gorm.DB) *NotificationPostgresRepository {
	return &NotificationPostgresRepository{db: db}
}

// CreateIfAbsent inserts one inbox item without replacing an existing item
// with the same owner-scoped dedupe key. Worker retries can therefore safely
// replay the same event.
func (r *NotificationPostgresRepository) CreateIfAbsent(ctx context.Context, item *entity.Notification) (bool, error) {
	if item == nil || item.OwnerID == "" || item.DedupeKey == "" || item.Kind == "" || item.Title == "" || item.Body == "" {
		return false, gorm.ErrInvalidData
	}
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "owner_id"}, {Name: "dedupe_key"}},
		DoNothing: true,
	}).Create(item)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func (r *NotificationPostgresRepository) List(ctx context.Context, ownerID string) ([]entity.Notification, error) {
	items := make([]entity.Notification, 0)
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at DESC, id DESC").Find(&items).Error
	return items, err
}

func (r *NotificationPostgresRepository) MarkRead(ctx context.Context, ownerID, id string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&entity.Notification{}).Where("owner_id = ? AND id = ? AND read_at IS NULL", ownerID, id).Update("read_at", now)
	if result.Error != nil {
		return result.Error
	}
	// An already-read row is intentionally a success; clients can safely retry.
	if result.RowsAffected > 0 {
		return nil
	}
	var exists int64
	if err := r.db.WithContext(ctx).Model(&entity.Notification{}).Where("owner_id = ? AND id = ?", ownerID, id).Count(&exists).Error; err != nil {
		return err
	}
	if exists == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
