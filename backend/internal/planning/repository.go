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

func (r *Repository) CreateEvent(ctx context.Context, userID string, input CreateEventInput) (EventSummary, error) {
	input, err := ValidateCreateEvent(input)
	if err != nil {
		return EventSummary{}, err
	}
	var event EventSummary
	var startsOn sql.NullTime
	var endsOn sql.NullTime
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO events (user_id, name, starts_on, ends_on, note)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, user_id::text, name, starts_on, ends_on, note, version
	`, userID, input.Name, input.StartsOn, input.EndsOn, input.Note).Scan(&event.ID, &event.UserID, &event.Name, &startsOn, &endsOn, &event.Note, &event.Version)
	if err != nil {
		return EventSummary{}, fmt.Errorf("create event: %w", err)
	}
	event.StartsOn = dateString(startsOn)
	event.EndsOn = dateString(endsOn)
	return event, nil
}

func (r *Repository) UpdateEvent(ctx context.Context, userID string, eventID string, input UpdateEventInput) (EventSummary, error) {
	input, err := ValidateUpdateEvent(input)
	if err != nil {
		return EventSummary{}, err
	}
	var event EventSummary
	var startsOn sql.NullTime
	var endsOn sql.NullTime
	err = r.db.QueryRowContext(ctx, `
		UPDATE events
		SET name = $3, starts_on = $4, ends_on = $5, note = $6, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		RETURNING id::text, user_id::text, name, starts_on, ends_on, note, version
	`, eventID, userID, input.Name, input.StartsOn, input.EndsOn, input.Note).Scan(&event.ID, &event.UserID, &event.Name, &startsOn, &endsOn, &event.Note, &event.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return EventSummary{}, ErrForbidden
	}
	if err != nil {
		return EventSummary{}, fmt.Errorf("update event: %w", err)
	}
	event.StartsOn = dateString(startsOn)
	event.EndsOn = dateString(endsOn)
	return r.eventWithTotals(ctx, userID, event)
}

func (r *Repository) ArchiveEvent(ctx context.Context, userID string, eventID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE events
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
	`, eventID, userID)
	if err != nil {
		return fmt.Errorf("archive event: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive event rows affected: %w", err)
	}
	if affected == 0 {
		return ErrForbidden
	}
	return nil
}

func (r *Repository) LinkEventTransaction(ctx context.Context, userID string, eventID string, transactionID string) error {
	if err := r.ensureActiveOwnerRow(ctx, "events", userID, eventID); err != nil {
		return err
	}
	if err := r.ensureActiveOwnerRow(ctx, "transactions", userID, transactionID); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO event_transactions (event_id, user_id, transaction_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id, transaction_id) DO NOTHING
	`, eventID, userID, transactionID); err != nil {
		return fmt.Errorf("link event transaction: %w", err)
	}
	return nil
}

func (r *Repository) ListEvents(ctx context.Context, userID string) ([]EventSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id::text, e.user_id::text, e.name, e.starts_on, e.ends_on, e.note, e.version,
		       coalesce(sum(t.amount_vnd) FILTER (WHERE t.archived_at IS NULL AND t.excluded_from_reports = false), 0),
		       count(t.id) FILTER (WHERE t.archived_at IS NULL AND t.excluded_from_reports = false)
		FROM events e
		LEFT JOIN event_transactions et ON et.event_id = e.id AND et.user_id = e.user_id
		LEFT JOIN transactions t ON t.id = et.transaction_id AND t.user_id = e.user_id
		WHERE e.user_id = $1 AND e.archived_at IS NULL
		GROUP BY e.id
		ORDER BY e.starts_on DESC, e.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	var events []EventSummary
	for rows.Next() {
		var event EventSummary
		var startsOn sql.NullTime
		var endsOn sql.NullTime
		if err := rows.Scan(&event.ID, &event.UserID, &event.Name, &startsOn, &endsOn, &event.Note, &event.Version, &event.TotalVND, &event.TransactionCount); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		event.StartsOn = dateString(startsOn)
		event.EndsOn = dateString(endsOn)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("event rows: %w", err)
	}
	return events, nil
}

func (r *Repository) CreateObligation(ctx context.Context, userID string, input CreateObligationInput) (ObligationSummary, error) {
	input, err := ValidateCreateObligation(input)
	if err != nil {
		return ObligationSummary{}, err
	}
	var obligation ObligationSummary
	var dueOn sql.NullTime
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO obligations (user_id, direction, principal_vnd, counterparty, due_on, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, user_id::text, direction, principal_vnd, counterparty, due_on, note, version
	`, userID, string(input.Direction), input.PrincipalVND, input.Counterparty, input.DueOn, input.Note).Scan(&obligation.ID, &obligation.UserID, &obligation.Direction, &obligation.PrincipalVND, &obligation.Counterparty, &dueOn, &obligation.Note, &obligation.Version)
	if err != nil {
		return ObligationSummary{}, fmt.Errorf("create obligation: %w", err)
	}
	obligation.DueOn = dateString(dueOn)
	obligation.RemainingVND = obligation.PrincipalVND
	return obligation, nil
}

func (r *Repository) UpdateObligation(ctx context.Context, userID string, obligationID string, input UpdateObligationInput) (ObligationSummary, error) {
	input, err := ValidateUpdateObligation(input)
	if err != nil {
		return ObligationSummary{}, err
	}
	repaid, err := r.obligationRepaid(ctx, userID, obligationID)
	if err != nil {
		return ObligationSummary{}, err
	}
	if repaid > input.PrincipalVND {
		return ObligationSummary{}, fmt.Errorf("%w: obligation principal cannot be below repayments", ErrValidation)
	}
	var obligation ObligationSummary
	var dueOn sql.NullTime
	err = r.db.QueryRowContext(ctx, `
		UPDATE obligations
		SET direction = $3, principal_vnd = $4, counterparty = $5, due_on = $6, note = $7, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		RETURNING id::text, user_id::text, direction, principal_vnd, counterparty, due_on, note, version
	`, obligationID, userID, string(input.Direction), input.PrincipalVND, input.Counterparty, input.DueOn, input.Note).Scan(&obligation.ID, &obligation.UserID, &obligation.Direction, &obligation.PrincipalVND, &obligation.Counterparty, &dueOn, &obligation.Note, &obligation.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return ObligationSummary{}, ErrForbidden
	}
	if err != nil {
		return ObligationSummary{}, fmt.Errorf("update obligation: %w", err)
	}
	obligation.DueOn = dateString(dueOn)
	return r.obligationWithTotals(ctx, userID, obligation)
}

func (r *Repository) ArchiveObligation(ctx context.Context, userID string, obligationID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE obligations
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
	`, obligationID, userID)
	if err != nil {
		return fmt.Errorf("archive obligation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive obligation rows affected: %w", err)
	}
	if affected == 0 {
		return ErrForbidden
	}
	return nil
}

func (r *Repository) LinkObligationRepayment(ctx context.Context, userID string, obligationID string, transactionID string) error {
	if err := r.ensureActiveOwnerRow(ctx, "obligations", userID, obligationID); err != nil {
		return err
	}
	exists, err := r.obligationRepaymentExists(ctx, userID, obligationID, transactionID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	amount, err := r.transactionAmount(ctx, userID, transactionID)
	if err != nil {
		return err
	}
	var principal int64
	var repaid int64
	err = r.db.QueryRowContext(ctx, `
		SELECT o.principal_vnd,
		       coalesce(sum(t.amount_vnd) FILTER (WHERE t.archived_at IS NULL), 0)
		FROM obligations o
		LEFT JOIN obligation_repayments repayment ON repayment.obligation_id = o.id AND repayment.user_id = o.user_id
		LEFT JOIN transactions t ON t.id = repayment.transaction_id AND t.user_id = o.user_id
		WHERE o.id = $1 AND o.user_id = $2 AND o.archived_at IS NULL
		GROUP BY o.id
	`, obligationID, userID).Scan(&principal, &repaid)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("obligation repayment total: %w", err)
	}
	if repaid+amount > principal {
		return fmt.Errorf("%w: repayment exceeds remaining principal", ErrValidation)
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO obligation_repayments (obligation_id, user_id, transaction_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (obligation_id, transaction_id) DO NOTHING
	`, obligationID, userID, transactionID); err != nil {
		return fmt.Errorf("link obligation repayment: %w", err)
	}
	return nil
}

func (r *Repository) ListObligations(ctx context.Context, userID string) ([]ObligationSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.id::text, o.user_id::text, o.direction, o.principal_vnd, o.counterparty, o.due_on, o.note, o.version,
		       coalesce(sum(t.amount_vnd) FILTER (WHERE t.archived_at IS NULL), 0)
		FROM obligations o
		LEFT JOIN obligation_repayments repayment ON repayment.obligation_id = o.id AND repayment.user_id = o.user_id
		LEFT JOIN transactions t ON t.id = repayment.transaction_id AND t.user_id = o.user_id
		WHERE o.user_id = $1 AND o.archived_at IS NULL
		GROUP BY o.id
		ORDER BY o.due_on, o.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list obligations: %w", err)
	}
	defer rows.Close()
	var obligations []ObligationSummary
	for rows.Next() {
		var obligation ObligationSummary
		var dueOn sql.NullTime
		if err := rows.Scan(&obligation.ID, &obligation.UserID, &obligation.Direction, &obligation.PrincipalVND, &obligation.Counterparty, &dueOn, &obligation.Note, &obligation.Version, &obligation.RepaidVND); err != nil {
			return nil, fmt.Errorf("scan obligation: %w", err)
		}
		obligation.DueOn = dateString(dueOn)
		obligation.RemainingVND = obligation.PrincipalVND - obligation.RepaidVND
		obligations = append(obligations, obligation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("obligation rows: %w", err)
	}
	return obligations, nil
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

func (r *Repository) ensureActiveOwnerRow(ctx context.Context, table string, userID string, id string) error {
	query := ""
	switch table {
	case "events":
		query = `SELECT EXISTS (SELECT 1 FROM events WHERE id = $1 AND user_id = $2 AND archived_at IS NULL)`
	case "obligations":
		query = `SELECT EXISTS (SELECT 1 FROM obligations WHERE id = $1 AND user_id = $2 AND archived_at IS NULL)`
	case "transactions":
		query = `SELECT EXISTS (SELECT 1 FROM transactions WHERE id = $1 AND user_id = $2 AND archived_at IS NULL)`
	default:
		return ErrForbidden
	}
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists); err != nil {
		return fmt.Errorf("check owner row: %w", err)
	}
	if !exists {
		return ErrForbidden
	}
	return nil
}

func (r *Repository) transactionAmount(ctx context.Context, userID string, transactionID string) (int64, error) {
	var amount int64
	err := r.db.QueryRowContext(ctx, `
		SELECT amount_vnd
		FROM transactions
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
	`, transactionID, userID).Scan(&amount)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrForbidden
	}
	if err != nil {
		return 0, fmt.Errorf("transaction amount: %w", err)
	}
	return amount, nil
}

func (r *Repository) obligationRepaymentExists(ctx context.Context, userID string, obligationID string, transactionID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM obligation_repayments
			WHERE obligation_id = $1 AND user_id = $2 AND transaction_id = $3
		)
	`, obligationID, userID, transactionID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check obligation repayment: %w", err)
	}
	return exists, nil
}

func (r *Repository) obligationRepaid(ctx context.Context, userID string, obligationID string) (int64, error) {
	var repaid int64
	err := r.db.QueryRowContext(ctx, `
		SELECT coalesce(sum(t.amount_vnd) FILTER (WHERE t.archived_at IS NULL), 0)
		FROM obligations o
		LEFT JOIN obligation_repayments repayment ON repayment.obligation_id = o.id AND repayment.user_id = o.user_id
		LEFT JOIN transactions t ON t.id = repayment.transaction_id AND t.user_id = o.user_id
		WHERE o.id = $1 AND o.user_id = $2 AND o.archived_at IS NULL
		GROUP BY o.id
	`, obligationID, userID).Scan(&repaid)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrForbidden
	}
	if err != nil {
		return 0, fmt.Errorf("obligation repaid: %w", err)
	}
	return repaid, nil
}

func (r *Repository) eventWithTotals(ctx context.Context, userID string, event EventSummary) (EventSummary, error) {
	err := r.db.QueryRowContext(ctx, `
		SELECT coalesce(sum(t.amount_vnd) FILTER (WHERE t.archived_at IS NULL AND t.excluded_from_reports = false), 0),
		       count(t.id) FILTER (WHERE t.archived_at IS NULL AND t.excluded_from_reports = false)
		FROM event_transactions et
		JOIN transactions t ON t.id = et.transaction_id AND t.user_id = et.user_id
		WHERE et.event_id = $1 AND et.user_id = $2
	`, event.ID, userID).Scan(&event.TotalVND, &event.TransactionCount)
	if err != nil {
		return EventSummary{}, fmt.Errorf("event totals: %w", err)
	}
	return event, nil
}

func (r *Repository) obligationWithTotals(ctx context.Context, userID string, obligation ObligationSummary) (ObligationSummary, error) {
	err := r.db.QueryRowContext(ctx, `
		SELECT coalesce(sum(t.amount_vnd) FILTER (WHERE t.archived_at IS NULL), 0)
		FROM obligation_repayments repayment
		JOIN transactions t ON t.id = repayment.transaction_id AND t.user_id = repayment.user_id
		WHERE repayment.obligation_id = $1 AND repayment.user_id = $2
	`, obligation.ID, userID).Scan(&obligation.RepaidVND)
	if err != nil {
		return ObligationSummary{}, fmt.Errorf("obligation totals: %w", err)
	}
	obligation.RemainingVND = obligation.PrincipalVND - obligation.RepaidVND
	return obligation, nil
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
