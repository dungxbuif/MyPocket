package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type FeedbackPostgresRepository struct{ db *gorm.DB }

func NewFeedbackPostgresRepository(db *gorm.DB) contract.FeedbackRepository {
	return &FeedbackPostgresRepository{db: db}
}

func (r *FeedbackPostgresRepository) Create(ctx context.Context, feedback *entity.Feedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}

func (r *FeedbackPostgresRepository) ListByOwner(ctx context.Context, owner string) ([]entity.Feedback, error) {
	var rows []entity.Feedback
	err := r.db.WithContext(ctx).Where("user_id = ?", owner).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *FeedbackPostgresRepository) FindByOwner(ctx context.Context, owner, id string) (*entity.Feedback, error) {
	var row entity.Feedback
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, owner).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, contract.ErrNotFound
	}
	return &row, err
}

func (r *FeedbackPostgresRepository) ListForAgent(ctx context.Context, status string, limit int) ([]entity.Feedback, error) {
	var rows []entity.Feedback
	query := r.db.WithContext(ctx).Order("created_at ASC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *FeedbackPostgresRepository) UpdateStatus(ctx context.Context, id, status string, now time.Time) (*entity.Feedback, error) {
	var row entity.Feedback
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, contract.ErrNotFound
		}
		return nil, err
	}
	result := r.db.WithContext(ctx).Model(&entity.Feedback{}).Where("id = ?", id).Updates(map[string]any{
		"status": status, "updated_at": now,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, contract.ErrFeedbackConflict
	}
	return r.FindByID(ctx, id)
}

func (r *FeedbackPostgresRepository) FindByID(ctx context.Context, id string) (*entity.Feedback, error) {
	var row entity.Feedback
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, contract.ErrNotFound
	}
	return &row, err
}
