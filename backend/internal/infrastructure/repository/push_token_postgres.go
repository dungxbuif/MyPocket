package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mypocket/backend/internal/entity"
	domain "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PushTokenPostgresRepository struct{ db *gorm.DB }

func NewPushTokenPostgresRepository(db *gorm.DB) domain.PushTokenRepository {
	return &PushTokenPostgresRepository{db: db}
}

func (r *PushTokenPostgresRepository) Save(ctx context.Context, item *entity.PushToken) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.PushToken
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("platform = ? AND environment = ? AND token = ?", item.Platform, item.Environment, item.Token).Take(&existing).Error
		if err == nil && existing.OwnerID != item.OwnerID {
			return domain.ErrPushTokenConflict
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		result := tx.Where("owner_id = ? AND device_id = ? AND platform = ? AND environment = ?", item.OwnerID, item.DeviceID, item.Platform, item.Environment).First(&existing)
		if result.Error == nil {
			updates := map[string]any{"device_name": item.DeviceName, "token": item.Token, "revoked_at": nil, "updated_at": time.Now().UTC()}
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					return domain.ErrPushTokenConflict
				}
				return err
			}
			return tx.Where("id = ?", existing.ID).Take(item).Error
		}
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		if err := tx.Create(item).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrPushTokenConflict
			}
			return err
		}
		return nil
	})
}

func (r *PushTokenPostgresRepository) Revoke(ctx context.Context, owner, id string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&entity.PushToken{}).Where("owner_id = ? AND id = ?", owner, id).Updates(map[string]any{"revoked_at": now, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
