package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type AssistantDelegationPostgresRepository struct{ db *gorm.DB }

func NewAssistantDelegationPostgresRepository(db *gorm.DB) contract.AssistantDelegationRepository {
	return &AssistantDelegationPostgresRepository{db: db}
}
func (r *AssistantDelegationPostgresRepository) Create(ctx context.Context, value *entity.AssistantDelegation) error {
	return r.db.WithContext(ctx).Create(value).Error
}
func (r *AssistantDelegationPostgresRepository) FindByLookup(ctx context.Context, lookup string) (*entity.AssistantDelegation, error) {
	var value entity.AssistantDelegation
	err := r.db.WithContext(ctx).Where("lookup_id = ?", lookup).First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	return &value, err
}
func (r *AssistantDelegationPostgresRepository) FindByID(ctx context.Context, id string) (*entity.AssistantDelegation, error) {
	var value entity.AssistantDelegation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contract.ErrNotFound
	}
	return &value, err
}
func (r *AssistantDelegationPostgresRepository) ListByOwner(ctx context.Context, owner string) ([]entity.AssistantDelegation, error) {
	var values []entity.AssistantDelegation
	err := r.db.WithContext(ctx).Where("owner_id = ?", owner).Order("created_at DESC").Find(&values).Error
	return values, err
}
func (r *AssistantDelegationPostgresRepository) Revoke(ctx context.Context, owner, id string, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&entity.AssistantDelegation{}).Where("owner_id = ? AND id = ? AND revoked_at IS NULL", owner, id).Updates(map[string]any{"revoked_at": now, "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return contract.ErrNotFound
	}
	return nil
}
func (r *AssistantDelegationPostgresRepository) TouchLastUsed(ctx context.Context, id string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.AssistantDelegation{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{"last_used_at": now, "updated_at": now}).Error
}
