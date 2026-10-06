package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type NotificationPostgresRepository struct{ db *gorm.DB }

func NewNotificationPostgresRepository(db *gorm.DB) contract.NotificationRepository {
	return &NotificationPostgresRepository{db: db}
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
