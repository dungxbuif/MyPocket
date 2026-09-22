package repository

import (
	"context"
	"sort"
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChangelogPostgresRepository struct{ db *gorm.DB }

func NewChangelogPostgresRepository(db *gorm.DB) contract.ChangelogRepository {
	return &ChangelogPostgresRepository{db: db}
}

func (r *ChangelogPostgresRepository) Publish(ctx context.Context, changelog *entity.Changelog, feedbackIDs []string) error {
	ids := append([]string(nil), feedbackIDs...)
	sort.Strings(ids)
	if len(ids) == 0 {
		return contract.ErrFeedbackInvalid
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []entity.Feedback
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id ASC").Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) != len(ids) {
			return contract.ErrNotFound
		}
		for _, row := range rows {
			if row.Status == entity.FeedbackStatusFixed || row.Status == entity.FeedbackStatusRejected || row.ChangelogID != nil {
				return contract.ErrFeedbackConflict
			}
		}
		if changelog.PublishedAt.IsZero() {
			changelog.PublishedAt = time.Now().UTC()
		}
		if err := tx.Create(changelog).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, id := range ids {
			result := tx.Model(&entity.Feedback{}).Where("id = ?", id).Updates(map[string]any{
				"status": entity.FeedbackStatusFixed, "fixed_at": now, "changelog_id": changelog.ID, "updated_at": now,
			})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return contract.ErrFeedbackConflict
			}
		}
		return nil
	})
}

func (r *ChangelogPostgresRepository) ListPublic(ctx context.Context, limit int) ([]entity.Changelog, error) {
	var rows []entity.Changelog
	err := r.db.WithContext(ctx).Order("published_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *ChangelogPostgresRepository) FindPublic(ctx context.Context, id string) (*entity.Changelog, error) {
	var row entity.Changelog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, contract.ErrNotFound
	}
	return &row, err
}
