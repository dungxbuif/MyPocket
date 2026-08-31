package planning

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateBudget(ctx context.Context, userID string, input CreateBudgetInput) (Budget, error) {
	input, err := ValidateCreateBudget(input)
	if err != nil {
		return Budget{}, err
	}
	if err := r.validateExpenseCategories(ctx, userID, input.CategoryIDs); err != nil {
		return Budget{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Budget{}, fmt.Errorf("begin create budget: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var budget Budget
	var customStart sql.NullTime
	var customEnd sql.NullTime
	err = tx.QueryRowContext(ctx, `
		INSERT INTO budgets (user_id, name, period_type, amount_vnd, all_categories, custom_start, custom_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, user_id::text, name, period_type, amount_vnd, all_categories, custom_start, custom_end, version
	`, userID, input.Name, string(input.PeriodType), input.AmountVND, len(input.CategoryIDs) == 0, nullableDate(input.CustomStart), nullableDate(input.CustomEnd)).Scan(
		&budget.ID,
		&budget.UserID,
		&budget.Name,
		&budget.PeriodType,
		&budget.AmountVND,
		&budget.AllCategories,
		&customStart,
		&customEnd,
		&budget.Version,
	)
	if err != nil {
		return Budget{}, fmt.Errorf("create budget: %w", err)
	}
	if err := insertBudgetCategories(ctx, tx, userID, budget.ID, input.CategoryIDs); err != nil {
		return Budget{}, err
	}
	if err := tx.Commit(); err != nil {
		return Budget{}, fmt.Errorf("commit create budget: %w", err)
	}
	budget.CategoryIDs = input.CategoryIDs
	budget.CustomStart = dateString(customStart)
	budget.CustomEnd = dateString(customEnd)
	return budget, nil
}

func (r *Repository) ListBudgetProgress(ctx context.Context, userID string, now time.Time) ([]BudgetProgress, error) {
	budgets, err := r.listBudgets(ctx, userID)
	if err != nil {
		return nil, err
	}
	results := make([]BudgetProgress, 0, len(budgets))
	for _, budget := range budgets {
		progress, err := r.progressForBudget(ctx, userID, budget, now)
		if err != nil {
			return nil, err
		}
		if err := r.ensureBudgetAlerts(ctx, userID, progress); err != nil {
			return nil, err
		}
		results = append(results, progress)
	}
	return results, nil
}

func (r *Repository) UpdateBudget(ctx context.Context, userID string, budgetID string, input UpdateBudgetInput) (Budget, error) {
	input, err := ValidateUpdateBudget(input)
	if err != nil {
		return Budget{}, err
	}
	if err := r.validateExpenseCategories(ctx, userID, input.CategoryIDs); err != nil {
		return Budget{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Budget{}, fmt.Errorf("begin update budget: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var budget Budget
	var customStart sql.NullTime
	var customEnd sql.NullTime
	err = tx.QueryRowContext(ctx, `
		UPDATE budgets
		SET name = $3,
		    period_type = $4,
		    amount_vnd = $5,
		    all_categories = $6,
		    custom_start = $7,
		    custom_end = $8,
		    updated_at = now(),
		    version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		RETURNING id::text, user_id::text, name, period_type, amount_vnd, all_categories, custom_start, custom_end, version
	`, budgetID, userID, input.Name, string(input.PeriodType), input.AmountVND, len(input.CategoryIDs) == 0, nullableDate(input.CustomStart), nullableDate(input.CustomEnd)).Scan(
		&budget.ID,
		&budget.UserID,
		&budget.Name,
		&budget.PeriodType,
		&budget.AmountVND,
		&budget.AllCategories,
		&customStart,
		&customEnd,
		&budget.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Budget{}, ErrForbidden
	}
	if err != nil {
		return Budget{}, fmt.Errorf("update budget: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM budget_categories WHERE budget_id = $1 AND user_id = $2`, budgetID, userID); err != nil {
		return Budget{}, fmt.Errorf("delete budget categories: %w", err)
	}
	if err := insertBudgetCategories(ctx, tx, userID, budget.ID, input.CategoryIDs); err != nil {
		return Budget{}, err
	}
	if err := tx.Commit(); err != nil {
		return Budget{}, fmt.Errorf("commit update budget: %w", err)
	}
	budget.CategoryIDs = input.CategoryIDs
	budget.CustomStart = dateString(customStart)
	budget.CustomEnd = dateString(customEnd)
	return budget, nil
}

func (r *Repository) ArchiveBudget(ctx context.Context, userID string, budgetID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE budgets
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
	`, budgetID, userID)
	if err != nil {
		return fmt.Errorf("archive budget: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive budget rows affected: %w", err)
	}
	if affected == 0 {
		return ErrForbidden
	}
	return nil
}

func (r *Repository) listBudgets(ctx context.Context, userID string) ([]Budget, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id::text, b.user_id::text, b.name, b.period_type, b.amount_vnd, b.all_categories, b.custom_start, b.custom_end, b.version,
		       coalesce(string_agg(bc.category_id::text, ',' ORDER BY bc.category_id::text), '')
		FROM budgets b
		LEFT JOIN budget_categories bc ON bc.budget_id = b.id
		WHERE b.user_id = $1 AND b.archived_at IS NULL
		GROUP BY b.id
		ORDER BY b.created_at, b.id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var budget Budget
		var customStart sql.NullTime
		var customEnd sql.NullTime
		var categories string
		if err := rows.Scan(&budget.ID, &budget.UserID, &budget.Name, &budget.PeriodType, &budget.AmountVND, &budget.AllCategories, &customStart, &customEnd, &budget.Version, &categories); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		budget.CustomStart = dateString(customStart)
		budget.CustomEnd = dateString(customEnd)
		budget.CategoryIDs = splitCSV(categories)
		budgets = append(budgets, budget)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("budget rows: %w", err)
	}
	return budgets, nil
}

func (r *Repository) progressForBudget(ctx context.Context, userID string, budget Budget, now time.Time) (BudgetProgress, error) {
	start, end, err := PeriodWindow(budget.PeriodType, parseBudgetDate(budget.CustomStart), parseBudgetDate(budget.CustomEnd), now)
	if err != nil {
		return BudgetProgress{}, err
	}
	nextDay := end.AddDate(0, 0, 1)
	var spent int64
	err = r.db.QueryRowContext(ctx, `
		SELECT coalesce(sum(t.amount_vnd), 0)
		FROM transactions t
		WHERE t.user_id = $1
		  AND t.type = 'expense'
		  AND t.archived_at IS NULL
		  AND t.excluded_from_reports = false
		  AND t.occurred_at >= $2
		  AND t.occurred_at < $3
		  AND (
		      $4
		      OR EXISTS (
		          SELECT 1
		          FROM budget_categories bc
		          WHERE bc.budget_id = $5 AND bc.category_id = t.category_id
		      )
		  )
	`, userID, start.UTC(), nextDay.UTC(), budget.AllCategories, budget.ID).Scan(&spent)
	if err != nil {
		return BudgetProgress{}, fmt.Errorf("budget progress: %w", err)
	}
	percent := percentSpent(spent, budget.AmountVND)
	return BudgetProgress{
		Budget:       budget,
		PeriodStart:  start.Format("2006-01-02"),
		PeriodEnd:    end.Format("2006-01-02"),
		SpentVND:     spent,
		RemainingVND: budget.AmountVND - spent,
		Percent:      percent,
		Alert80:      percent >= 80,
		Alert100:     percent >= 100,
	}, nil
}

func (r *Repository) ensureBudgetAlerts(ctx context.Context, userID string, progress BudgetProgress) error {
	for _, threshold := range []int{80, 100} {
		if progress.Percent < int64(threshold) {
			continue
		}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO budget_alerts (user_id, budget_id, threshold, period_start, period_end)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (budget_id, threshold, period_start) DO NOTHING
		`, userID, progress.Budget.ID, threshold, progress.PeriodStart, progress.PeriodEnd); err != nil {
			return fmt.Errorf("insert budget alert: %w", err)
		}
	}
	return nil
}

func (r *Repository) validateExpenseCategories(ctx context.Context, userID string, categoryIDs []string) error {
	for _, categoryID := range categoryIDs {
		categoryID = trimmed(categoryID)
		if categoryID == "" {
			return fmt.Errorf("%w: category id is required", ErrValidation)
		}
		var exists bool
		if err := r.db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM categories
				WHERE id = $1
				  AND kind = 'expense'
				  AND archived_at IS NULL
				  AND (is_system OR user_id = $2)
			)
		`, categoryID, userID).Scan(&exists); err != nil {
			return fmt.Errorf("check budget category: %w", err)
		}
		if !exists {
			return ErrForbidden
		}
	}
	return nil
}

func insertBudgetCategories(ctx context.Context, tx *sql.Tx, userID string, budgetID string, categoryIDs []string) error {
	for _, categoryID := range categoryIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO budget_categories (budget_id, user_id, category_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (budget_id, category_id) DO NOTHING
		`, budgetID, userID, trimmed(categoryID)); err != nil {
			return fmt.Errorf("insert budget category: %w", err)
		}
	}
	return nil
}

func nullableDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return dateOnly(*value).Format("2006-01-02")
}

func dateString(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	var out []string
	start := 0
	for idx := 0; idx <= len(value); idx++ {
		if idx == len(value) || value[idx] == ',' {
			item := trimmed(value[start:idx])
			if item != "" {
				out = append(out, item)
			}
			start = idx + 1
		}
	}
	return out
}

func parseBudgetDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &parsed
}
