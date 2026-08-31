package sync

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func RequestHash(mutation Mutation) (string, error) {
	body, err := json.Marshal(struct {
		DeviceID    string          `json:"device_id"`
		Sequence    int64           `json:"sequence"`
		EntityType  EntityType      `json:"entity_type"`
		EntityID    string          `json:"entity_id"`
		Operation   Operation       `json:"operation"`
		BaseVersion int64           `json:"base_version"`
		Payload     json.RawMessage `json:"payload"`
	}{
		DeviceID:    mutation.DeviceID,
		Sequence:    mutation.Sequence,
		EntityType:  mutation.EntityType,
		EntityID:    mutation.EntityID,
		Operation:   mutation.Operation,
		BaseVersion: mutation.BaseVersion,
		Payload:     normalizedPayload(mutation.Payload),
	})
	if err != nil {
		return "", fmt.Errorf("hash mutation: %w", err)
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func (r *Repository) LoadMutationResult(ctx context.Context, userID string, mutationID string) (requestHash string, result MutationResult, found bool, err error) {
	var resultJSON []byte
	err = r.db.QueryRowContext(ctx, `
		SELECT request_hash, result_json
		FROM sync_mutations
		WHERE user_id = $1 AND mutation_id = $2
	`, userID, mutationID).Scan(&requestHash, &resultJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return "", MutationResult{}, false, nil
	}
	if err != nil {
		return "", MutationResult{}, false, fmt.Errorf("load sync mutation: %w", err)
	}
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return "", MutationResult{}, false, fmt.Errorf("decode sync mutation result: %w", err)
	}
	return requestHash, result, true, nil
}

func (r *Repository) StoreMutationResult(ctx context.Context, userID string, mutationID string, requestHash string, result MutationResult) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal sync mutation result: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO sync_mutations (user_id, mutation_id, request_hash, result_json)
		VALUES ($1, $2, $3, $4)
	`, userID, mutationID, requestHash, resultJSON)
	if err != nil {
		return fmt.Errorf("store sync mutation result: %w", err)
	}
	return nil
}

func (r *Repository) AppendChange(ctx context.Context, userID string, entityType EntityType, entityID string, operation Operation, version int64, payload json.RawMessage) (Change, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Change{}, fmt.Errorf("begin sync change: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var nextCursor int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO sync_cursors (user_id, next_cursor)
		VALUES ($1, 1)
		ON CONFLICT (user_id) DO NOTHING
		RETURNING next_cursor
	`, userID).Scan(&nextCursor)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `
			SELECT next_cursor
			FROM sync_cursors
			WHERE user_id = $1
			FOR UPDATE
		`, userID).Scan(&nextCursor)
	}
	if err != nil {
		return Change{}, fmt.Errorf("reserve sync cursor: %w", err)
	}

	change := Change{
		Cursor:     nextCursor,
		EntityType: entityType,
		EntityID:   entityID,
		Operation:  operation,
		Version:    version,
		Payload:    normalizedPayload(payload),
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO sync_changes (user_id, cursor, entity_type, entity_id, operation, version, payload_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`, userID, nextCursor, entityType, entityID, operation, version, change.Payload).Scan(&change.CreatedAt)
	if err != nil {
		return Change{}, fmt.Errorf("insert sync change: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE sync_cursors
		SET next_cursor = $2
		WHERE user_id = $1
	`, userID, nextCursor+1); err != nil {
		return Change{}, fmt.Errorf("advance sync cursor: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Change{}, fmt.Errorf("commit sync change: %w", err)
	}
	return change, nil
}

func (r *Repository) ListChanges(ctx context.Context, userID string, after int64, limit int) (ChangesResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT cursor, entity_type, entity_id, operation, version, payload_json, created_at
		FROM sync_changes
		WHERE user_id = $1 AND cursor > $2
		ORDER BY cursor
		LIMIT $3
	`, userID, after, limit)
	if err != nil {
		return ChangesResult{}, fmt.Errorf("list sync changes: %w", err)
	}
	defer rows.Close()

	var result ChangesResult
	result.NextCursor = after
	for rows.Next() {
		var change Change
		if err := rows.Scan(&change.Cursor, &change.EntityType, &change.EntityID, &change.Operation, &change.Version, &change.Payload, &change.CreatedAt); err != nil {
			return ChangesResult{}, fmt.Errorf("scan sync change: %w", err)
		}
		result.Changes = append(result.Changes, change)
		result.NextCursor = change.Cursor
	}
	if err := rows.Err(); err != nil {
		return ChangesResult{}, fmt.Errorf("sync change rows: %w", err)
	}
	return result, nil
}

func (r *Repository) CurrentCursor(ctx context.Context, userID string) (int64, error) {
	var nextCursor int64
	err := r.db.QueryRowContext(ctx, `
		SELECT next_cursor
		FROM sync_cursors
		WHERE user_id = $1
	`, userID).Scan(&nextCursor)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("current sync cursor: %w", err)
	}
	return nextCursor - 1, nil
}

func (r *Repository) EntityVersionAndPayload(ctx context.Context, userID string, entityType EntityType, entityID string) (int64, json.RawMessage, bool, error) {
	switch entityType {
	case EntityWallet:
		return r.walletVersionAndPayload(ctx, userID, entityID)
	case EntityCategory:
		return r.categoryVersionAndPayload(ctx, userID, entityID)
	case EntityTransaction:
		return r.transactionVersionAndPayload(ctx, userID, entityID)
	default:
		return 0, nil, false, fmt.Errorf("%w: unsupported entity_type", ErrValidation)
	}
}

func (r *Repository) walletVersionAndPayload(ctx context.Context, userID string, walletID string) (int64, json.RawMessage, bool, error) {
	var version int64
	var payload []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT version, to_jsonb(row) - 'user_id'
		FROM (
			SELECT id::text, user_id::text, name, type, balance_vnd, include_in_total, is_default_ai, version
			FROM wallets
			WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		) row
	`, walletID, userID).Scan(&version, &payload)
	return versionPayloadResult(err, payload)
}

func (r *Repository) categoryVersionAndPayload(ctx context.Context, userID string, categoryID string) (int64, json.RawMessage, bool, error) {
	var version int64
	var payload []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT version, to_jsonb(row) - 'user_id'
		FROM (
			SELECT id::text, coalesce(user_id::text, '') AS user_id, coalesce(parent_id::text, '') AS parent_id, kind, name, coalesce(system_key, '') AS system_key, is_system, version
			FROM categories
			WHERE id = $1 AND archived_at IS NULL AND (is_system OR user_id = $2)
		) row
	`, categoryID, userID).Scan(&version, &payload)
	return versionPayloadResult(err, payload)
}

func (r *Repository) transactionVersionAndPayload(ctx context.Context, userID string, transactionID string) (int64, json.RawMessage, bool, error) {
	var version int64
	var payload []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT version, to_jsonb(row) - 'user_id'
		FROM (
			SELECT id::text, user_id::text, type, source_wallet_id::text, coalesce(destination_wallet_id::text, '') AS destination_wallet_id, coalesce(category_id::text, '') AS category_id, amount_vnd, coalesce(balance_after_vnd, 0) AS balance_after_vnd, note, with_person, event_ref, occurred_at, excluded_from_reports, version
			FROM transactions
			WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		) row
	`, transactionID, userID).Scan(&version, &payload)
	return versionPayloadResult(err, payload)
}

func versionPayloadResult(err error, payload []byte) (int64, json.RawMessage, bool, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil, false, nil
	}
	if err != nil {
		return 0, nil, false, err
	}
	var versionProbe struct {
		Version int64 `json:"version"`
	}
	if err := json.Unmarshal(payload, &versionProbe); err != nil {
		return 0, nil, false, fmt.Errorf("decode entity version: %w", err)
	}
	return versionProbe.Version, json.RawMessage(payload), true, nil
}

func normalizedPayload(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 || string(payload) == "null" {
		return json.RawMessage(`{}`)
	}
	return payload
}
