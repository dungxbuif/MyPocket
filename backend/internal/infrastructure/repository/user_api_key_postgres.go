package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type UserAPIKeyPostgresRepository struct{ db *gorm.DB }

func NewUserAPIKeyPostgresRepository(db *gorm.DB) contract.UserAPIKeyRepository {
	return &UserAPIKeyPostgresRepository{db: db}
}

func (r *UserAPIKeyPostgresRepository) Create(ctx context.Context, key *entity.UserAPIKey) error {
	if r == nil || r.db == nil || key == nil || strings.TrimSpace(key.OwnerID) == "" {
		return entity.ErrAPIKeyInvalid
	}
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *UserAPIKeyPostgresRepository) FindByLookup(ctx context.Context, lookup string) (*entity.UserAPIKey, error) {
	var key entity.UserAPIKey
	err := r.db.WithContext(ctx).Where("lookup_id = ?", lookup).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrAPIKeyNotFound
	}
	return &key, err
}

func (r *UserAPIKeyPostgresRepository) FindByID(ctx context.Context, id string) (*entity.UserAPIKey, error) {
	var key entity.UserAPIKey
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrAPIKeyNotFound
	}
	return &key, err
}

func (r *UserAPIKeyPostgresRepository) ListByOwner(ctx context.Context, ownerID string) ([]entity.UserAPIKey, error) {
	var keys []entity.UserAPIKey
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at DESC, id DESC").Find(&keys).Error
	return keys, err
}

func (r *UserAPIKeyPostgresRepository) Revoke(ctx context.Context, ownerID, id string, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&entity.UserAPIKey{}).Where("owner_id = ? AND id = ? AND revoked_at IS NULL", ownerID, id).Updates(map[string]any{"revoked_at": now.UTC(), "updated_at": now.UTC()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return contract.ErrAPIKeyNotFound
	}
	return nil
}

func (r *UserAPIKeyPostgresRepository) TouchLastUsed(ctx context.Context, id string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.UserAPIKey{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{"last_used_at": now.UTC(), "updated_at": now.UTC()}).Error
}
