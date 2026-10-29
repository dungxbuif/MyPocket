package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"gorm.io/gorm"
)

const (
	budgetNotificationKind = "budget.threshold.v1"
	budgetWarningPercent   = 80
	budgetLimitPercent     = 100
)

type budgetNotificationCandidate struct {
	OwnerID   string
	BudgetID  string
	Threshold int
	DedupeKey string
	Title     string
	Body      string
}

// budgetNotificationCandidates is deliberately pure so threshold, timezone,
// and exclusion rules can be tested without a worker or provider.
func budgetNotificationCandidates(budget entity.Budget, transactions []entity.Transaction, categories []entity.Category, now time.Time, location *time.Location) ([]budgetNotificationCandidate, error) {
	if location == nil || budget.OwnerID == "" || budget.ID == "" || budget.LimitAmount <= 0 {
		return nil, nil
	}
	localToday := entity.CalendarDate{Time: time.Date(now.In(location).Year(), now.In(location).Month(), now.In(location).Day(), 0, 0, 0, 0, time.UTC)}
	if budget.StartDate.IsZero() || budget.EndDate.IsZero() || localToday.Before(budget.StartDate.Time) || localToday.After(budget.EndDate.Time) {
		return nil, nil
	}
	report, err := entity.CalculateBudgetReport(budget, transactions, categories, now, location)
	if err != nil {
		return nil, err
	}
	limit := report.EffectiveLimitAmount
	if limit <= 0 || report.Spent <= 0 {
		return nil, nil
	}
	result := make([]budgetNotificationCandidate, 0, 2)
	for _, threshold := range []int{budgetWarningPercent, budgetLimitPercent} {
		if report.Spent < thresholdAmount(limit, threshold) {
			continue
		}
		period := budget.StartDate.Time.Format("2006-01-02") + ":" + budget.EndDate.Time.Format("2006-01-02")
		kind := "sắp chạm hạn mức"
		if threshold == budgetLimitPercent {
			kind = "đã chạm hạn mức"
		}
		result = append(result, budgetNotificationCandidate{
			OwnerID:   budget.OwnerID,
			BudgetID:  budget.ID,
			Threshold: threshold,
			DedupeKey: fmt.Sprintf("%s:%s:%d", budget.ID, period, threshold),
			Title:     "Ngân sách " + budget.Name + " " + kind,
			Body:      fmt.Sprintf("Bạn đã dùng %d%% ngân sách trong kỳ hiện tại.", threshold),
		})
	}
	return result, nil
}

func thresholdAmount(limit int64, percent int) int64 {
	// ceil(limit * percent / 100), without overflowing int64.
	return (limit*int64(percent) + 99) / 100
}

// runBudgetNotificationPass materializes only in-app inbox rows. The unique
// owner/dedupe index makes this safe across retries and overlapping workers.
func runBudgetNotificationPass(database *gorm.DB, now time.Time) error {
	var budgets []entity.Budget
	// Load the small budget set and apply the account-timezone active-window
	// check in budgetNotificationCandidates. Filtering with a UTC timestamp in
	// SQL would incorrectly drop a budget on its local end date.
	if err := database.Find(&budgets).Error; err != nil {
		return err
	}
	notifications := repo.NewNotificationPostgresRepository(database)
	for _, budget := range budgets {
		var user entity.User
		if err := database.Where("id = ?", budget.OwnerID).First(&user).Error; err != nil {
			return err
		}
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			return err
		}
		var transactions []entity.Transaction
		if err := database.Where("owner_id = ?", budget.OwnerID).Find(&transactions).Error; err != nil {
			return err
		}
		var categories []entity.Category
		if err := database.Where("owner_id = ? OR is_system = true", budget.OwnerID).Find(&categories).Error; err != nil {
			return err
		}
		candidates, err := budgetNotificationCandidates(budget, transactions, categories, now, location)
		if err != nil {
			return fmt.Errorf("budget %s notification evaluation: %w", budget.ID, err)
		}
		for _, candidate := range candidates {
			_, err := notifications.CreateIfAbsent(context.Background(), &entity.Notification{ID: uuid.NewString(), OwnerID: candidate.OwnerID, Kind: budgetNotificationKind, Title: candidate.Title, Body: candidate.Body, DedupeKey: candidate.DedupeKey})
			if err != nil {
				return fmt.Errorf("budget %s notification insert: %w", budget.ID, err)
			}
		}
	}
	return nil
}
