package lifecycle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"mypocket/internal/finance"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateJob(ctx context.Context, userID string, kind Kind, idempotencyKey string, request any) (Job, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if userID == "" || idempotencyKey == "" || !validKind(kind) {
		return Job{}, ErrInvalid
	}
	body, err := json.Marshal(request)
	if err != nil {
		return Job{}, ErrInvalid
	}
	var id string
	err = r.db.QueryRowContext(ctx, `INSERT INTO data_jobs (user_id,kind,status,idempotency_key,request) VALUES ($1,$2,'queued',$3,$4) ON CONFLICT (user_id,kind,idempotency_key) DO UPDATE SET idempotency_key=EXCLUDED.idempotency_key RETURNING id::text`, userID, kind, idempotencyKey, body).Scan(&id)
	if err != nil {
		return Job{}, fmt.Errorf("create lifecycle job: %w", err)
	}
	return r.GetJob(ctx, userID, id)
}

func (r *Repository) GetJob(ctx context.Context, userID, id string) (Job, error) {
	var job Job
	var result []byte
	var objectKey, errorCode sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id::text,user_id::text,kind,status,request,result,result_object_key,attempts,version,error_code,next_attempt_at,created_at,updated_at FROM data_jobs WHERE id=$1 AND user_id=$2`, id, userID).Scan(&job.ID, &job.UserID, &job.Kind, &job.Status, &job.Request, &result, &objectKey, &job.Attempts, &job.Version, &errorCode, &job.NextAttemptAt, &job.CreatedAt, &job.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, ErrNotFound
	}
	if err != nil {
		return Job{}, fmt.Errorf("get lifecycle job: %w", err)
	}
	job.Result, job.ResultObjectKey, job.ErrorCode = result, objectKey.String, errorCode.String
	return job, nil
}

func (r *Repository) ClaimDue(ctx context.Context) (Job, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Job{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var id, userID string
	err = tx.QueryRowContext(ctx, `SELECT id::text,user_id::text FROM data_jobs WHERE status IN ('queued','running') AND next_attempt_at<=now() ORDER BY next_attempt_at,created_at FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, ErrNotFound
	}
	if err != nil {
		return Job{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE data_jobs SET status='running',attempts=attempts+1,version=version+1,updated_at=now(),next_attempt_at=now()+interval '5 minutes' WHERE id=$1`, id); err != nil {
		return Job{}, err
	}
	if err = tx.Commit(); err != nil {
		return Job{}, err
	}
	return r.GetJob(ctx, userID, id)
}

func (r *Repository) Complete(ctx context.Context, job Job, result any, objectKey string) error {
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return r.transition(ctx, job, StatusCompleted, body, objectKey, "")
}

func (r *Repository) AwaitConfirmation(ctx context.Context, job Job, preview ImportPreview) error {
	body, err := json.Marshal(preview)
	if err != nil {
		return err
	}
	return r.transition(ctx, job, StatusAwaitingConfirmation, body, "", "")
}

func (r *Repository) ConfirmImport(ctx context.Context, userID, id string, version int64) (Job, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE data_jobs SET status='queued',version=version+1,next_attempt_at=now(),updated_at=now() WHERE id=$1 AND user_id=$2 AND kind='import' AND status='awaiting_confirmation' AND version=$3`, id, userID, version)
	if err != nil {
		return Job{}, err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return Job{}, ErrConflict
	}
	return r.GetJob(ctx, userID, id)
}

func (r *Repository) Fail(ctx context.Context, job Job, code string) error {
	return r.transition(ctx, job, StatusFailed, nil, "", code)
}

func (r *Repository) Retry(ctx context.Context, job Job, code string) error {
	if job.Attempts >= 5 {
		return r.Fail(ctx, job, code)
	}
	result, err := r.db.ExecContext(ctx, `UPDATE data_jobs SET status='queued',error_code=$3,next_attempt_at=now()+(attempts*interval '1 minute'),version=version+1,updated_at=now() WHERE id=$1 AND user_id=$2 AND version=$4`, job.ID, job.UserID, code, job.Version)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (r *Repository) transition(ctx context.Context, job Job, status Status, result []byte, objectKey, code string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE data_jobs SET status=$3,result=$4,result_object_key=NULLIF($5,''),error_code=NULLIF($6,''),version=version+1,updated_at=now() WHERE id=$1 AND user_id=$2 AND version=$7`, job.ID, job.UserID, status, result, objectKey, code, job.Version)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (r *Repository) Counts(ctx context.Context, userID string) (map[string]int, error) {
	tables := []string{"wallets", "transactions", "budgets", "events", "obligations", "recurring_schedules", "asset_positions", "receipt_objects"}
	result := make(map[string]int, len(tables))
	for _, table := range tables {
		var count int
		if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE user_id=$1`, userID).Scan(&count); err != nil {
			return nil, err
		}
		result[table] = count
	}
	return result, nil
}

func (r *Repository) OwnedReferences(ctx context.Context, userID string) (map[string]bool, map[string]bool, error) {
	wallets, categories := map[string]bool{}, map[string]bool{}
	rows, err := r.db.QueryContext(ctx, `SELECT id::text FROM wallets WHERE user_id=$1 AND archived_at IS NULL`, userID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, nil, err
		}
		wallets[id] = true
	}
	rows.Close()
	rows, err = r.db.QueryContext(ctx, `SELECT id::text FROM categories WHERE (user_id=$1 OR user_id IS NULL) AND archived_at IS NULL`, userID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, nil, err
		}
		categories[id] = true
	}
	rows.Close()
	return wallets, categories, nil
}

func (r *Repository) ApplyImport(ctx context.Context, job Job) (int, error) {
	var preview ImportPreview
	if err := json.Unmarshal(job.Result, &preview); err != nil || !preview.Confirmable || len(preview.Errors) > 0 {
		return 0, ErrInvalid
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	repo := finance.NewRepositoryInTx(tx)
	for _, row := range preview.ValidRows {
		_, err = repo.CreateTransaction(ctx, job.UserID, finance.CreateTransactionInput{IdempotencyKey: fmt.Sprintf("import:%s:%d", job.ID, row.Row), Type: finance.TransactionType(row.Type), SourceWalletID: row.SourceWalletID, DestinationWalletID: row.DestinationWalletID, CategoryID: row.CategoryID, AmountVND: row.AmountVND, OccurredAt: row.OccurredAt, Note: row.Note, ExcludedFromReports: row.ExcludedFromReports})
		if err != nil {
			return 0, fmt.Errorf("apply import row %d: %w", row.Row, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(preview.ValidRows), nil
}

func (r *Repository) DisableUser(ctx context.Context, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `UPDATE users SET disabled_at=coalesce(disabled_at,now()),updated_at=now() WHERE id=$1`, userID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE api_keys SET revoked_at=coalesce(revoked_at,now()),updated_at=now() WHERE user_id=$1`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ResetUserData(ctx context.Context, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, table := range []string{"push_subscriptions", "notifications", "transaction_drafts", "recurring_occurrences", "recurring_schedules", "budget_alerts", "budget_categories", "budgets", "event_transactions", "obligation_repayments", "events", "obligations", "asset_price_history", "asset_trades", "asset_positions", "sync_changes", "sync_mutations", "sync_cursors", "finance_idempotency_keys", "transactions", "wallet_category_settings", "receipt_objects", "wallets", "categories"} {
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE user_id=$1`, userID); err != nil {
			return fmt.Errorf("reset %s: %w", table, err)
		}
	}
	return tx.Commit()
}

func (r *Repository) DeleteUserData(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1 AND disabled_at IS NOT NULL`, userID)
	return err
}

func validKind(kind Kind) bool {
	return kind == KindImport || kind == KindExport || kind == KindReset || kind == KindDelete
}

func RetryAt(at time.Time) time.Time { return at.Add(5 * time.Minute) }
