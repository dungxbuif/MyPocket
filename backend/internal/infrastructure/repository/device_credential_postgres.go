package repository

import (
	"context"
	"errors"
	"github.com/mypocket/backend/internal/entity"
	domain "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type DeviceCredentialPostgresRepository struct{ db *gorm.DB }

func NewDeviceCredentialPostgresRepository(db *gorm.DB) domain.DeviceCredentialRepository {
	return &DeviceCredentialPostgresRepository{db: db}
}
func (r *DeviceCredentialPostgresRepository) Save(ctx context.Context, d *entity.DeviceCredential) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "owner_id"}, {Name: "device_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"device_name", "secret_hash", "password_fingerprint", "expires_at", "revoked_at", "updated_at"}),
		}).Create(d).Error
		if err != nil {
			return err
		}
		var persisted entity.DeviceCredential
		if err := tx.Where("owner_id = ? AND device_id = ?", d.OwnerID, d.DeviceID).Take(&persisted).Error; err != nil {
			return err
		}
		*d = persisted
		return nil
	})
}
func (r *DeviceCredentialPostgresRepository) FindByID(ctx context.Context, id string) (*entity.DeviceCredential, error) {
	var d entity.DeviceCredential
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}
func (r *DeviceCredentialPostgresRepository) Revoke(ctx context.Context, owner, id string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&entity.DeviceCredential{}).Where("owner_id = ? AND id = ?", owner, id).Updates(map[string]any{"revoked_at": now, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
