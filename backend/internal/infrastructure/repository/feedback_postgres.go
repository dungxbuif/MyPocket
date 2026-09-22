package repository

import (
	"context"
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	var updated entity.Feedback
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row entity.Feedback
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return contract.ErrNotFound
			}
			return err
		}
		// The usecase validates the transition before entering the repository,
		// but the row may have changed while another worker was completing a
		// transition. Re-check under the row lock so concurrent agents cannot
		// skip a lifecycle state.
		if !entity.CanTransitionFeedback(row.Status, status) {
			return contract.ErrFeedbackConflict
		}
		result := tx.Model(&entity.Feedback{}).Where("id = ?", id).Updates(map[string]any{
			"status": status, "updated_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return contract.ErrFeedbackConflict
		}
		updated = row
		updated.Status = status
		updated.UpdatedAt = now
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *FeedbackPostgresRepository) FindByID(ctx context.Context, id string) (*entity.Feedback, error) {
	var row entity.Feedback
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, contract.ErrNotFound
	}
	return &row, err
}
