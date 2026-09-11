package planning

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"mypocket/internal/finance"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

var ErrDraftAlreadyResolved = errors.New("transaction draft already resolved")
var ErrDraftVersionConflict = errors.New("transaction draft version conflict")

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
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $9
		RETURNING id::text, user_id::text, name, period_type, amount_vnd, all_categories, custom_start, custom_end, version
	`, budgetID, userID, input.Name, string(input.PeriodType), input.AmountVND, len(input.CategoryIDs) == 0, nullableDate(input.CustomStart), nullableDate(input.CustomEnd), input.BaseVersion).Scan(
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
		var currentVersion int64
		lookupErr := tx.QueryRowContext(ctx, `SELECT version FROM budgets WHERE id = $1 AND user_id = $2 AND archived_at IS NULL`, budgetID, userID).Scan(&currentVersion)
		if lookupErr == nil {
			return Budget{}, ErrVersionConflict
		}
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

func (r *Repository) ArchiveBudget(ctx context.Context, userID string, budgetID string, baseVersion int64) error {
	if baseVersion <= 0 {
		return fmt.Errorf("%w: base version is required", ErrValidation)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE budgets
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $3
	`, budgetID, userID, baseVersion)
	if err != nil {
		return fmt.Errorf("archive budget: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive budget rows affected: %w", err)
	}
	if affected == 0 {
		var currentVersion int64
		if lookupErr := r.db.QueryRowContext(ctx, `SELECT version FROM budgets WHERE id = $1 AND user_id = $2 AND archived_at IS NULL`, budgetID, userID).Scan(&currentVersion); lookupErr == nil {
			return ErrVersionConflict
		}
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
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $7
		RETURNING id::text, user_id::text, name, starts_on, ends_on, note, version
	`, eventID, userID, input.Name, input.StartsOn, input.EndsOn, input.Note, input.BaseVersion).Scan(&event.ID, &event.UserID, &event.Name, &startsOn, &endsOn, &event.Note, &event.Version)
	if errors.Is(err, sql.ErrNoRows) {
		var currentVersion int64
		if lookupErr := r.db.QueryRowContext(ctx, `SELECT version FROM events WHERE id = $1 AND user_id = $2 AND archived_at IS NULL`, eventID, userID).Scan(&currentVersion); lookupErr == nil {
			return EventSummary{}, ErrVersionConflict
		}
		return EventSummary{}, ErrForbidden
	}
	if err != nil {
		return EventSummary{}, fmt.Errorf("update event: %w", err)
	}
	event.StartsOn = dateString(startsOn)
	event.EndsOn = dateString(endsOn)
	return r.eventWithTotals(ctx, userID, event)
}

func (r *Repository) ArchiveEvent(ctx context.Context, userID string, eventID string, baseVersion int64) error {
	if baseVersion <= 0 {
		return fmt.Errorf("%w: base version is required", ErrValidation)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE events
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $3
	`, eventID, userID, baseVersion)
	if err != nil {
		return fmt.Errorf("archive event: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive event rows affected: %w", err)
	}
	if affected == 0 {
		var currentVersion int64
		if lookupErr := r.db.QueryRowContext(ctx, `SELECT version FROM events WHERE id = $1 AND user_id = $2 AND archived_at IS NULL`, eventID, userID).Scan(&currentVersion); lookupErr == nil {
			return ErrVersionConflict
		}
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ObligationSummary{}, fmt.Errorf("begin update obligation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var currentVersion int64
	err = tx.QueryRowContext(ctx, `
		SELECT version FROM obligations
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		FOR UPDATE
	`, obligationID, userID).Scan(&currentVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return ObligationSummary{}, ErrForbidden
	}
	if err != nil {
		return ObligationSummary{}, fmt.Errorf("lock obligation update: %w", err)
	}
	if currentVersion != input.BaseVersion {
		return ObligationSummary{}, ErrVersionConflict
	}
	var repaid int64
	if err := tx.QueryRowContext(ctx, `
		SELECT coalesce(sum(t.amount_vnd) FILTER (WHERE t.archived_at IS NULL), 0)
		FROM obligation_repayments repayment
		JOIN transactions t ON t.id = repayment.transaction_id AND t.user_id = repayment.user_id
		WHERE repayment.obligation_id = $1 AND repayment.user_id = $2
	`, obligationID, userID).Scan(&repaid); err != nil {
		return ObligationSummary{}, fmt.Errorf("obligation repaid: %w", err)
	}
	if repaid > input.PrincipalVND {
		return ObligationSummary{}, fmt.Errorf("%w: obligation principal cannot be below repayments", ErrValidation)
	}
	var obligation ObligationSummary
	var dueOn sql.NullTime
	err = tx.QueryRowContext(ctx, `
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
	if err := tx.Commit(); err != nil {
		return ObligationSummary{}, fmt.Errorf("commit update obligation: %w", err)
	}
	obligation.DueOn = dateString(dueOn)
	return r.obligationWithTotals(ctx, userID, obligation)
}

func (r *Repository) ArchiveObligation(ctx context.Context, userID string, obligationID string, baseVersion int64) error {
	if baseVersion <= 0 {
		return fmt.Errorf("%w: base version is required", ErrValidation)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE obligations
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $3
	`, obligationID, userID, baseVersion)
	if err != nil {
		return fmt.Errorf("archive obligation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive obligation rows affected: %w", err)
	}
	if affected == 0 {
		var currentVersion int64
		if lookupErr := r.db.QueryRowContext(ctx, `SELECT version FROM obligations WHERE id = $1 AND user_id = $2 AND archived_at IS NULL`, obligationID, userID).Scan(&currentVersion); lookupErr == nil {
			return ErrVersionConflict
		}
		return ErrForbidden
	}
	return nil
}

func (r *Repository) LinkObligationRepayment(ctx context.Context, userID string, obligationID string, transactionID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin link obligation repayment: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var principal int64
	var direction ObligationDirection
	err = tx.QueryRowContext(ctx, `
		SELECT principal_vnd, direction
		FROM obligations
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		FOR UPDATE
	`, obligationID, userID).Scan(&principal, &direction)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("lock obligation repayment: %w", err)
	}

	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM obligation_repayments
			WHERE obligation_id = $1 AND user_id = $2 AND transaction_id = $3
		)
	`, obligationID, userID, transactionID).Scan(&exists); err != nil {
		return fmt.Errorf("check obligation repayment: %w", err)
	}
	if exists {
		return tx.Commit()
	}

	var amount int64
	var transactionType finance.TransactionType
	err = tx.QueryRowContext(ctx, `
		SELECT amount_vnd, type
		FROM transactions
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		FOR UPDATE
	`, transactionID, userID).Scan(&amount, &transactionType)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("load repayment transaction: %w", err)
	}
	expectedType := finance.TransactionExpense
	if direction == ObligationLent {
		expectedType = finance.TransactionIncome
	}
	if transactionType != expectedType {
		return fmt.Errorf("%w: %s obligation repayment requires %s transaction", ErrValidation, direction, expectedType)
	}
	var linkedElsewhere bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM obligation_repayments
			WHERE user_id = $1 AND transaction_id = $2 AND obligation_id <> $3
		)
	`, userID, transactionID, obligationID).Scan(&linkedElsewhere); err != nil {
		return fmt.Errorf("check repayment reuse: %w", err)
	}
	if linkedElsewhere {
		return fmt.Errorf("%w: transaction is already linked to another obligation", ErrValidation)
	}

	var repaid int64
	if err := tx.QueryRowContext(ctx, `
		SELECT coalesce(sum(t.amount_vnd) FILTER (WHERE t.archived_at IS NULL), 0)
		FROM obligation_repayments repayment
		JOIN transactions t ON t.id = repayment.transaction_id AND t.user_id = repayment.user_id
		WHERE repayment.obligation_id = $1 AND repayment.user_id = $2
	`, obligationID, userID).Scan(&repaid); err != nil {
		return fmt.Errorf("obligation repayment total: %w", err)
	}
	if repaid > principal || amount > principal-repaid {
		return fmt.Errorf("%w: repayment exceeds remaining principal", ErrValidation)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO obligation_repayments (obligation_id, user_id, transaction_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (obligation_id, transaction_id) DO NOTHING
	`, obligationID, userID, transactionID); err != nil {
		return fmt.Errorf("link obligation repayment: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit obligation repayment: %w", err)
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

func (r *Repository) CreateRecurringSchedule(ctx context.Context, userID string, input CreateRecurringScheduleInput) (RecurringSchedule, error) {
	input, startsAt, err := ValidateCreateRecurringSchedule(input)
	if err != nil {
		return RecurringSchedule{}, err
	}
	if err := r.validateRecurringFinanceRefs(ctx, userID, input); err != nil {
		return RecurringSchedule{}, err
	}
	var schedule RecurringSchedule
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO recurring_schedules (user_id, name, frequency, timezone, starts_at, next_occurs_at, transaction_type, source_wallet_id, destination_wallet_id, category_id, amount_vnd, note)
		VALUES ($1, $2, $3, $4, $5, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id::text, user_id::text, name, frequency, timezone, starts_at, next_occurs_at, transaction_type, source_wallet_id::text, coalesce(destination_wallet_id::text, ''), coalesce(category_id::text, ''), amount_vnd, note, version
	`, userID, input.Name, string(input.Frequency), input.Timezone, startsAt.UTC(), string(input.Type), input.SourceWalletID, nullString(input.DestinationWalletID), nullString(input.CategoryID), input.AmountVND, input.Note).Scan(
		&schedule.ID, &schedule.UserID, &schedule.Name, &schedule.Frequency, &schedule.Timezone, &schedule.StartsAt, &schedule.NextOccursAt, &schedule.Type, &schedule.SourceWalletID, &schedule.DestinationWalletID, &schedule.CategoryID, &schedule.AmountVND, &schedule.Note, &schedule.Version,
	)
	if err != nil {
		return RecurringSchedule{}, fmt.Errorf("create recurring schedule: %w", err)
	}
	return schedule, nil
}

func (r *Repository) ListRecurringSchedules(ctx context.Context, userID string) ([]RecurringSchedule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, name, frequency, timezone, starts_at, next_occurs_at, transaction_type, source_wallet_id::text, coalesce(destination_wallet_id::text, ''), coalesce(category_id::text, ''), amount_vnd, note, version
		FROM recurring_schedules
		WHERE user_id = $1 AND archived_at IS NULL
		ORDER BY next_occurs_at, created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list recurring schedules: %w", err)
	}
	defer rows.Close()
	var schedules []RecurringSchedule
	for rows.Next() {
		var schedule RecurringSchedule
		if err := rows.Scan(&schedule.ID, &schedule.UserID, &schedule.Name, &schedule.Frequency, &schedule.Timezone, &schedule.StartsAt, &schedule.NextOccursAt, &schedule.Type, &schedule.SourceWalletID, &schedule.DestinationWalletID, &schedule.CategoryID, &schedule.AmountVND, &schedule.Note, &schedule.Version); err != nil {
			return nil, fmt.Errorf("scan recurring schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recurring schedule rows: %w", err)
	}
	return schedules, nil
}

func (r *Repository) ArchiveRecurringSchedule(ctx context.Context, userID string, scheduleID string, baseVersion int64) error {
	if baseVersion <= 0 {
		return fmt.Errorf("%w: base version is required", ErrValidation)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE recurring_schedules
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $3
	`, scheduleID, userID, baseVersion)
	if err != nil {
		return fmt.Errorf("archive recurring schedule: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive recurring schedule rows affected: %w", err)
	}
	if affected == 0 {
		var currentVersion int64
		if lookupErr := r.db.QueryRowContext(ctx, `SELECT version FROM recurring_schedules WHERE id = $1 AND user_id = $2 AND archived_at IS NULL`, scheduleID, userID).Scan(&currentVersion); lookupErr == nil {
			return ErrVersionConflict
		}
		return ErrForbidden
	}
	return nil
}

func (r *Repository) ListTransactionDrafts(ctx context.Context, userID string) ([]TransactionDraft, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, coalesce(schedule_id::text, ''), occurrence_key, transaction_type, source_wallet_id::text, coalesce(destination_wallet_id::text, ''), coalesce(category_id::text, ''), amount_vnd, occurred_at, note, status, coalesce(confirmed_transaction_id::text, ''), version
		FROM transaction_drafts
		WHERE user_id = $1
		ORDER BY occurred_at DESC, created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list transaction drafts: %w", err)
	}
	defer rows.Close()
	var drafts []TransactionDraft
	for rows.Next() {
		draft, err := scanTransactionDraft(rows)
		if err != nil {
			return nil, fmt.Errorf("scan transaction draft: %w", err)
		}
		drafts = append(drafts, draft)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("transaction draft rows: %w", err)
	}
	return drafts, nil
}

func (r *Repository) ConfirmTransactionDraft(ctx context.Context, userID string, draftID string, input ConfirmTransactionDraftInput) (TransactionDraftDecision, error) {
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.Note = strings.TrimSpace(input.Note)
	if input.Version <= 0 || input.AmountVND <= 0 || input.IdempotencyKey == "" {
		return TransactionDraftDecision{}, fmt.Errorf("%w: version, positive amount, and idempotency key are required", ErrValidation)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TransactionDraftDecision{}, fmt.Errorf("begin confirm transaction draft: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	draft, err := lockTransactionDraft(ctx, tx, userID, draftID)
	if err != nil {
		return TransactionDraftDecision{}, err
	}
	if draft.Status != "pending" {
		if draft.Status != "confirmed" || draft.AmountVND != input.AmountVND || draft.Note != input.Note || draft.ConfirmedTransactionID == "" {
			return TransactionDraftDecision{}, ErrDraftAlreadyResolved
		}
		transaction, err := finance.GetTransactionInTx(ctx, tx, userID, draft.ConfirmedTransactionID)
		if err != nil {
			return TransactionDraftDecision{}, fmt.Errorf("load confirmed draft transaction: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return TransactionDraftDecision{}, fmt.Errorf("commit confirmed draft replay: %w", err)
		}
		return TransactionDraftDecision{Draft: draft, Transaction: &transaction}, nil
	}
	if draft.Version != input.Version {
		return TransactionDraftDecision{}, ErrDraftVersionConflict
	}

	transaction, err := finance.CreateTransactionInTx(ctx, tx, userID, finance.CreateTransactionInput{
		IdempotencyKey:      draftFinanceIdempotencyKey(draft.ID, input.IdempotencyKey),
		Type:                draft.Type,
		SourceWalletID:      draft.SourceWalletID,
		DestinationWalletID: draft.DestinationWalletID,
		CategoryID:          draft.CategoryID,
		AmountVND:           input.AmountVND,
		OccurredAt:          draft.OccurredAt,
		Note:                input.Note,
	})
	if err != nil {
		return TransactionDraftDecision{}, mapDraftFinanceError(err)
	}
	draft, err = updateConfirmedDraft(ctx, tx, userID, draft.ID, input, transaction.ID)
	if err != nil {
		return TransactionDraftDecision{}, err
	}
	if err := tx.Commit(); err != nil {
		return TransactionDraftDecision{}, fmt.Errorf("commit confirm transaction draft: %w", err)
	}
	return TransactionDraftDecision{Draft: draft, Transaction: &transaction}, nil
}

func (r *Repository) RejectTransactionDraft(ctx context.Context, userID string, draftID string, input RejectTransactionDraftInput) (TransactionDraftDecision, error) {
	if input.Version <= 0 {
		return TransactionDraftDecision{}, fmt.Errorf("%w: version is required", ErrValidation)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TransactionDraftDecision{}, fmt.Errorf("begin reject transaction draft: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	draft, err := lockTransactionDraft(ctx, tx, userID, draftID)
	if err != nil {
		return TransactionDraftDecision{}, err
	}
	if draft.Status != "pending" {
		if draft.Status != "rejected" {
			return TransactionDraftDecision{}, ErrDraftAlreadyResolved
		}
		if err := tx.Commit(); err != nil {
			return TransactionDraftDecision{}, fmt.Errorf("commit rejected draft replay: %w", err)
		}
		return TransactionDraftDecision{Draft: draft}, nil
	}
	if draft.Version != input.Version {
		return TransactionDraftDecision{}, ErrDraftVersionConflict
	}
	draft, err = updateRejectedDraft(ctx, tx, userID, draft.ID, input.Version)
	if err != nil {
		return TransactionDraftDecision{}, err
	}
	if err := tx.Commit(); err != nil {
		return TransactionDraftDecision{}, fmt.Errorf("commit reject transaction draft: %w", err)
	}
	return TransactionDraftDecision{Draft: draft}, nil
}

func lockTransactionDraft(ctx context.Context, tx *sql.Tx, userID string, draftID string) (TransactionDraft, error) {
	draft, err := scanTransactionDraft(tx.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, coalesce(schedule_id::text, ''), occurrence_key, transaction_type, source_wallet_id::text, coalesce(destination_wallet_id::text, ''), coalesce(category_id::text, ''), amount_vnd, occurred_at, note, status, coalesce(confirmed_transaction_id::text, ''), version
		FROM transaction_drafts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`, draftID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return TransactionDraft{}, ErrForbidden
	}
	if err != nil {
		return TransactionDraft{}, fmt.Errorf("lock transaction draft: %w", err)
	}
	return draft, nil
}

func updateConfirmedDraft(ctx context.Context, tx *sql.Tx, userID string, draftID string, input ConfirmTransactionDraftInput, transactionID string) (TransactionDraft, error) {
	draft, err := scanTransactionDraft(tx.QueryRowContext(ctx, `
		UPDATE transaction_drafts
		SET amount_vnd = $4, note = $5, status = 'confirmed', confirmed_transaction_id = $6, version = version + 1, updated_at = now()
		WHERE id = $1 AND user_id = $2 AND status = 'pending' AND version = $3
		RETURNING id::text, user_id::text, coalesce(schedule_id::text, ''), occurrence_key, transaction_type, source_wallet_id::text, coalesce(destination_wallet_id::text, ''), coalesce(category_id::text, ''), amount_vnd, occurred_at, note, status, coalesce(confirmed_transaction_id::text, ''), version
	`, draftID, userID, input.Version, input.AmountVND, input.Note, transactionID))
	if errors.Is(err, sql.ErrNoRows) {
		return TransactionDraft{}, ErrDraftVersionConflict
	}
	if err != nil {
		return TransactionDraft{}, fmt.Errorf("finalize confirmed transaction draft: %w", err)
	}
	return draft, nil
}

func updateRejectedDraft(ctx context.Context, tx *sql.Tx, userID string, draftID string, version int64) (TransactionDraft, error) {
	draft, err := scanTransactionDraft(tx.QueryRowContext(ctx, `
		UPDATE transaction_drafts
		SET status = 'rejected', version = version + 1, updated_at = now()
		WHERE id = $1 AND user_id = $2 AND status = 'pending' AND version = $3
		RETURNING id::text, user_id::text, coalesce(schedule_id::text, ''), occurrence_key, transaction_type, source_wallet_id::text, coalesce(destination_wallet_id::text, ''), coalesce(category_id::text, ''), amount_vnd, occurred_at, note, status, coalesce(confirmed_transaction_id::text, ''), version
	`, draftID, userID, version))
	if errors.Is(err, sql.ErrNoRows) {
		return TransactionDraft{}, ErrDraftVersionConflict
	}
	if err != nil {
		return TransactionDraft{}, fmt.Errorf("finalize rejected transaction draft: %w", err)
	}
	return draft, nil
}

type transactionDraftScanner interface {
	Scan(dest ...any) error
}

func scanTransactionDraft(scanner transactionDraftScanner) (TransactionDraft, error) {
	var draft TransactionDraft
	err := scanner.Scan(&draft.ID, &draft.UserID, &draft.ScheduleID, &draft.OccurrenceKey, &draft.Type, &draft.SourceWalletID, &draft.DestinationWalletID, &draft.CategoryID, &draft.AmountVND, &draft.OccurredAt, &draft.Note, &draft.Status, &draft.ConfirmedTransactionID, &draft.Version)
	return draft, err
}

func draftFinanceIdempotencyKey(draftID string, clientKey string) string {
	sum := sha256.Sum256([]byte("mypocket:transaction-draft-confirm:v1\x00" + draftID + "\x00" + clientKey))
	return hex.EncodeToString(sum[:])
}

func mapDraftFinanceError(err error) error {
	switch {
	case errors.Is(err, finance.ErrValidation):
		return fmt.Errorf("%w: draft transaction is invalid", ErrValidation)
	case errors.Is(err, finance.ErrForbidden):
		return fmt.Errorf("%w: draft transaction references are unavailable", ErrValidation)
	default:
		return err
	}
}

func (r *Repository) AcquireWorkerLease(ctx context.Context, leaseKey string, owner string, ttl time.Duration, now time.Time) (bool, error) {
	leaseKey = trimmed(leaseKey)
	owner = trimmed(owner)
	if leaseKey == "" || owner == "" || ttl <= 0 {
		return false, fmt.Errorf("%w: invalid worker lease", ErrValidation)
	}
	var acquired bool
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO worker_leases (lease_key, owner, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (lease_key) DO UPDATE
		SET owner = EXCLUDED.owner,
		    expires_at = EXCLUDED.expires_at,
		    updated_at = now()
		WHERE worker_leases.expires_at <= $4 OR worker_leases.owner = $2
		RETURNING true
	`, leaseKey, owner, now.Add(ttl), now).Scan(&acquired)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("acquire worker lease: %w", err)
	}
	return acquired, nil
}

func (r *Repository) ProcessDueRecurringSchedules(ctx context.Context, workerID string, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		limit = 25
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin recurring processing: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
		SELECT id::text, user_id::text, name, frequency, timezone, starts_at, next_occurs_at, transaction_type, source_wallet_id::text, coalesce(destination_wallet_id::text, ''), coalesce(category_id::text, ''), amount_vnd, note, version
		FROM recurring_schedules
		WHERE archived_at IS NULL AND next_occurs_at <= $1
		ORDER BY next_occurs_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, now.UTC(), limit)
	if err != nil {
		return 0, fmt.Errorf("query due recurring schedules: %w", err)
	}
	var schedules []RecurringSchedule
	for rows.Next() {
		var schedule RecurringSchedule
		if err := rows.Scan(&schedule.ID, &schedule.UserID, &schedule.Name, &schedule.Frequency, &schedule.Timezone, &schedule.StartsAt, &schedule.NextOccursAt, &schedule.Type, &schedule.SourceWalletID, &schedule.DestinationWalletID, &schedule.CategoryID, &schedule.AmountVND, &schedule.Note, &schedule.Version); err != nil {
			_ = rows.Close()
			return 0, fmt.Errorf("scan due recurring schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Close(); err != nil {
		return 0, fmt.Errorf("close due recurring schedules: %w", err)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("due recurring rows: %w", err)
	}

	processed := 0
	attempted := 0
	for _, schedule := range schedules {
		next := schedule.NextOccursAt
		for !next.After(now.UTC()) && attempted < limit {
			key := fmt.Sprintf("recurring:%s:%s", schedule.ID, next.UTC().Format(time.RFC3339))
			inserted, err := insertRecurringDraft(ctx, tx, schedule, key, next.UTC())
			if err != nil {
				return 0, err
			}
			if inserted {
				processed++
			}
			attempted++
			next = NextOccurrence(next, schedule.Frequency, schedule.Timezone)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE recurring_schedules
			SET next_occurs_at = $1, updated_at = now(), version = version + 1
			WHERE id = $2
		`, next.UTC(), schedule.ID); err != nil {
			return 0, fmt.Errorf("advance recurring schedule: %w", err)
		}
		if attempted >= limit {
			break
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit recurring processing: %w", err)
	}
	_ = workerID
	return processed, nil
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

func (r *Repository) validateRecurringFinanceRefs(ctx context.Context, userID string, input CreateRecurringScheduleInput) error {
	if err := r.ensureActiveOwnerRow(ctx, "wallets", userID, input.SourceWalletID); err != nil {
		return err
	}
	if input.DestinationWalletID != "" {
		if err := r.ensureActiveOwnerRow(ctx, "wallets", userID, input.DestinationWalletID); err != nil {
			return err
		}
	}
	if input.CategoryID != "" {
		var kind finance.CategoryKind
		if err := r.db.QueryRowContext(ctx, `
			SELECT kind
			FROM categories
			WHERE id = $1
			  AND archived_at IS NULL
			  AND (is_system OR user_id = $2)
		`, input.CategoryID, userID).Scan(&kind); errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else if err != nil {
			return fmt.Errorf("check schedule category: %w", err)
		}
		if (input.Type == finance.TransactionIncome && kind != finance.CategoryIncome) ||
			(input.Type == finance.TransactionExpense && kind != finance.CategoryExpense) {
			return fmt.Errorf("%w: schedule category kind does not match transaction type", ErrValidation)
		}
	}
	return nil
}

func (r *Repository) ensureActiveOwnerRow(ctx context.Context, table string, userID string, id string) error {
	query := ""
	switch table {
	case "wallets":
		query = `SELECT EXISTS (SELECT 1 FROM wallets WHERE id = $1 AND user_id = $2 AND archived_at IS NULL)`
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

func insertRecurringDraft(ctx context.Context, tx *sql.Tx, schedule RecurringSchedule, occurrenceKey string, occursAt time.Time) (bool, error) {
	result, err := tx.ExecContext(ctx, `
		WITH occurrence AS (
			INSERT INTO recurring_occurrences (user_id, schedule_id, occurrence_key, occurs_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (schedule_id, occurrence_key) DO NOTHING
			RETURNING occurrence_key
		)
		INSERT INTO transaction_drafts (user_id, schedule_id, occurrence_key, transaction_type, source_wallet_id, destination_wallet_id, category_id, amount_vnd, occurred_at, note)
		SELECT $1, $2, $3, $5, $6, $7, $8, $9, $4, $10
		FROM occurrence
		ON CONFLICT (user_id, occurrence_key) DO NOTHING
	`, schedule.UserID, schedule.ID, occurrenceKey, occursAt, string(schedule.Type), schedule.SourceWalletID, nullString(schedule.DestinationWalletID), nullString(schedule.CategoryID), schedule.AmountVND, schedule.Note)
	if err != nil {
		return false, fmt.Errorf("insert recurring draft: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("recurring draft rows affected: %w", err)
	}
	return affected > 0, nil
}

func nullString(value string) any {
	if trimmed(value) == "" {
		return nil
	}
	return value
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
