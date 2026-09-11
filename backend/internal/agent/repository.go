package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("agent run not found")
var ErrConflict = errors.New("agent idempotency conflict")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateRun(ctx context.Context, userID, key string, kind Kind, text string) (Run, error) {
	if err := ValidateCreate(kind, text, key); err != nil {
		return Run{}, err
	}
	return scanRun(r.db.QueryRowContext(ctx, `
		INSERT INTO agent_runs (user_id, idempotency_key, kind, request_text)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, idempotency_key) DO UPDATE SET idempotency_key = EXCLUDED.idempotency_key
		WHERE agent_runs.kind = EXCLUDED.kind AND agent_runs.request_text = EXCLUDED.request_text
		RETURNING id::text, user_id::text, kind, status, request_text, coalesce(response_text,''), coalesce(error_code,''), attempts, created_at, updated_at, next_attempt_at, coalesce(lease_owner,''), coalesce(lease_expires_at, 'epoch'::timestamptz)
	`, userID, strings.TrimSpace(key), kind, strings.TrimSpace(text)))
}

func (r *Repository) GetRun(ctx context.Context, userID, id string) (Run, error) {
	run, err := scanRun(r.db.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, kind, status, request_text, coalesce(response_text,''), coalesce(error_code,''), attempts, created_at, updated_at, next_attempt_at, coalesce(lease_owner,''), coalesce(lease_expires_at, 'epoch'::timestamptz)
		FROM agent_runs WHERE id=$1 AND user_id=$2
	`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Run{}, ErrNotFound
	}
	if err != nil {
		return Run{}, fmt.Errorf("get agent run: %w", err)
	}
	run.DraftIDs, err = r.draftIDs(ctx, run.ID, run.UserID)
	return run, err
}

func (r *Repository) ClaimDue(ctx context.Context, owner string, now time.Time, lease time.Duration) (Run, bool, error) {
	run, err := scanRun(r.db.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT id FROM agent_runs
			WHERE next_attempt_at <= $1 AND (status='queued' OR (status='processing' AND lease_expires_at <= $1))
			ORDER BY next_attempt_at, id FOR UPDATE SKIP LOCKED LIMIT 1
		)
		UPDATE agent_runs r SET status='processing', attempts=r.attempts+1, lease_owner=$2, lease_expires_at=$3, updated_at=$1
		FROM candidate WHERE r.id=candidate.id
		RETURNING r.id::text, r.user_id::text, r.kind, r.status, r.request_text, coalesce(r.response_text,''), coalesce(r.error_code,''), r.attempts, r.created_at, r.updated_at, r.next_attempt_at, coalesce(r.lease_owner,''), coalesce(r.lease_expires_at, 'epoch'::timestamptz)
	`, now, owner, now.Add(lease)))
	if errors.Is(err, sql.ErrNoRows) {
		return Run{}, false, nil
	}
	if err != nil {
		return Run{}, false, fmt.Errorf("claim agent run: %w", err)
	}
	return run, true, nil
}

func (r *Repository) Context(ctx context.Context, userID string) (string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 'wallet', id::text, name FROM wallets WHERE user_id=$1 AND archived_at IS NULL
		UNION ALL
		SELECT 'category', id::text, name FROM categories WHERE (user_id=$1 OR user_id IS NULL) AND archived_at IS NULL
		ORDER BY 1, 3
	`, userID)
	if err != nil {
		return "", fmt.Errorf("load agent context: %w", err)
	}
	defer rows.Close()
	type item struct{ Type, ID, Name string }
	items := []item{}
	for rows.Next() {
		var value item
		if err := rows.Scan(&value.Type, &value.ID, &value.Name); err != nil {
			return "", fmt.Errorf("scan agent context: %w", err)
		}
		items = append(items, value)
	}
	var income, expense int64
	var from, to time.Time
	if err := r.db.QueryRowContext(ctx, `
		SELECT now() - interval '30 days', now(),
		       coalesce(sum(amount_vnd) FILTER (WHERE type='income'),0),
		       coalesce(sum(amount_vnd) FILTER (WHERE type='expense'),0)
		FROM transactions
		WHERE user_id=$1 AND archived_at IS NULL AND excluded_from_reports=false
		  AND occurred_at >= now() - interval '30 days' AND occurred_at <= now()
	`, userID).Scan(&from, &to, &income, &expense); err != nil {
		return "", fmt.Errorf("load agent aggregate: %w", err)
	}
	body, err := json.Marshal(map[string]any{
		"owned_references": items,
		"aggregate":        map[string]any{"scope_from": from.UTC().Format(time.RFC3339), "scope_to": to.UTC().Format(time.RFC3339), "income_vnd": income, "expense_vnd": expense, "net_vnd": income - expense},
	})
	if err != nil {
		return "", fmt.Errorf("encode agent context: %w", err)
	}
	return string(body), rows.Err()
}

func (r *Repository) Complete(ctx context.Context, run Run, result ModelResult, provenance map[string]any) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin agent completion: %w", err)
	}
	defer tx.Rollback()
	if run.Kind == KindTransactionDraft {
		if result.Transaction == nil {
			return fmt.Errorf("%w: missing transaction", ErrValidation)
		}
		proposal := result.Transaction
		if err := validateOwnedReferences(ctx, tx, run.UserID, proposal); err != nil {
			return err
		}
		occurred, _ := time.Parse(time.RFC3339, proposal.OccurredAt)
		var destination, category any
		if proposal.DestinationWalletID != nil {
			destination = *proposal.DestinationWalletID
		}
		if proposal.CategoryID != nil {
			category = *proposal.CategoryID
		}
		proof, _ := json.Marshal(provenance)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO transaction_drafts (user_id, occurrence_key, transaction_type, source_wallet_id, destination_wallet_id, category_id, amount_vnd, occurred_at, note, agent_run_id, provenance)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (user_id, occurrence_key) DO NOTHING
		`, run.UserID, "agent:"+run.ID, proposal.Type, proposal.SourceWalletID, destination, category, proposal.AmountVND, occurred, proposal.Note, run.ID, proof); err != nil {
			return fmt.Errorf("insert agent draft: %w", err)
		}
	}
	response := strings.TrimSpace(result.Answer)
	if _, err := tx.ExecContext(ctx, `UPDATE agent_runs SET status='completed', response_text=$1, error_code=NULL, lease_owner=NULL, lease_expires_at=NULL, updated_at=now() WHERE id=$2 AND user_id=$3 AND status='processing'`, response, run.ID, run.UserID); err != nil {
		return fmt.Errorf("complete agent run: %w", err)
	}
	return tx.Commit()
}

func (r *Repository) Fail(ctx context.Context, run Run, code string, retry bool) error {
	status, next := StatusFailed, time.Now().UTC()
	if retry && run.Attempts < 3 {
		status, next = StatusQueued, time.Now().UTC().Add(time.Duration(run.Attempts)*time.Minute)
	}
	_, err := r.db.ExecContext(ctx, `UPDATE agent_runs SET status=$1, error_code=$2, next_attempt_at=$3, lease_owner=NULL, lease_expires_at=NULL, updated_at=now() WHERE id=$4 AND user_id=$5`, status, code, next, run.ID, run.UserID)
	return err
}

func validateOwnedReferences(ctx context.Context, tx *sql.Tx, userID string, proposal *ProposedTransaction) error {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM wallets WHERE id=$1 AND user_id=$2 AND archived_at IS NULL`, proposal.SourceWalletID, userID).Scan(&count); err != nil || count != 1 {
		return fmt.Errorf("%w: source wallet unavailable", ErrValidation)
	}
	if proposal.DestinationWalletID != nil {
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM wallets WHERE id=$1 AND user_id=$2 AND archived_at IS NULL`, *proposal.DestinationWalletID, userID).Scan(&count); err != nil || count != 1 {
			return fmt.Errorf("%w: destination wallet unavailable", ErrValidation)
		}
	}
	if proposal.CategoryID != nil {
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM categories WHERE id=$1 AND (user_id=$2 OR user_id IS NULL) AND archived_at IS NULL`, *proposal.CategoryID, userID).Scan(&count); err != nil || count != 1 {
			return fmt.Errorf("%w: category unavailable", ErrValidation)
		}
	}
	return nil
}

func (r *Repository) draftIDs(ctx context.Context, runID, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id::text FROM transaction_drafts WHERE agent_run_id=$1 AND user_id=$2 ORDER BY id`, runID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

type scanner interface{ Scan(...any) error }

func scanRun(row scanner) (Run, error) {
	var run Run
	err := row.Scan(&run.ID, &run.UserID, &run.Kind, &run.Status, &run.RequestText, &run.ResponseText, &run.ErrorCode, &run.Attempts, &run.CreatedAt, &run.UpdatedAt, &run.NextAttemptAt, &run.LeaseOwner, &run.LeaseExpiresAt)
	return run, err
}
