package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Append(ctx context.Context, event Event) error {
	if r == nil || r.db == nil {
		return nil
	}
	event = normalizeEvent(event)
	if event.Action == "" || event.CorrelationID == "" {
		return ErrValidation
	}
	metadata := event.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	var actor any
	if event.ActorUserID != "" {
		actor = event.ActorUserID
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_events (
			correlation_id, actor_user_id, actor_email_hash, action, entity_type, entity_id,
			outcome, severity, source, error_code, request_method, request_path,
			ip_hash, user_agent_hash, metadata_json
		)
		VALUES ($1, nullif($2, '')::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, event.CorrelationID, actor, event.ActorEmailHash, event.Action, event.EntityType, event.EntityID,
		event.Outcome, event.Severity, event.Source, event.ErrorCode, event.RequestMethod,
		event.RequestPath, event.IPHash, event.UserAgentHash, metadata)
	if err != nil {
		return fmt.Errorf("append audit event: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, query Query) ([]Event, error) {
	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	args := []any{}
	clauses := []string{"1=1"}
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(condition, len(args)))
	}
	if query.CorrelationID != "" {
		add("correlation_id = $%d", query.CorrelationID)
	}
	if query.Action != "" {
		add("action = $%d", query.Action)
	}
	if query.Severity != "" {
		add("severity = $%d", query.Severity)
	}
	if !query.From.IsZero() {
		add("occurred_at >= $%d", query.From.UTC())
	}
	if !query.To.IsZero() {
		add("occurred_at <= $%d", query.To.UTC())
	}
	if !query.Before.IsZero() {
		add("occurred_at < $%d", query.Before.UTC())
	}
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, occurred_at, correlation_id, coalesce(actor_user_id::text, ''),
			actor_email_hash, action, entity_type, entity_id, outcome, severity, source,
			error_code, request_method, request_path, ip_hash, user_agent_hash,
			metadata_json, created_at
		FROM audit_events
		WHERE `+strings.Join(clauses, " AND ")+`
		ORDER BY occurred_at DESC, id DESC
		LIMIT $`+fmt.Sprint(len(args)), args...)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()
	events := []Event{}
	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.ID, &event.OccurredAt, &event.CorrelationID, &event.ActorUserID,
			&event.ActorEmailHash, &event.Action, &event.EntityType, &event.EntityID,
			&event.Outcome, &event.Severity, &event.Source, &event.ErrorCode,
			&event.RequestMethod, &event.RequestPath, &event.IPHash, &event.UserAgentHash,
			&event.Metadata, &event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *Repository) PurgeExpired(ctx context.Context, retentionDays int, limit int) (int, error) {
	if retentionDays <= 0 {
		retentionDays = 180
	}
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM audit_events
		WHERE id IN (
			SELECT id
			FROM audit_events
			WHERE occurred_at < now() - make_interval(days => $1)
			ORDER BY occurred_at
			LIMIT $2
		)
	`, retentionDays, limit)
	if err != nil {
		return 0, fmt.Errorf("purge audit events: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("purge audit rows affected: %w", err)
	}
	return int(count), nil
}

func (r *Repository) AcquireWorkerLease(ctx context.Context, leaseKey string, owner string, ttlSeconds int64) (bool, error) {
	leaseKey = strings.TrimSpace(leaseKey)
	owner = strings.TrimSpace(owner)
	if leaseKey == "" || owner == "" || ttlSeconds <= 0 {
		return false, fmt.Errorf("%w: invalid worker lease", ErrValidation)
	}
	var acquired bool
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO worker_leases (lease_key, owner, expires_at)
		VALUES ($1, $2, now() + make_interval(secs => $3))
		ON CONFLICT (lease_key) DO UPDATE
		SET owner = EXCLUDED.owner,
			expires_at = EXCLUDED.expires_at,
			updated_at = now()
		WHERE worker_leases.expires_at <= now() OR worker_leases.owner = EXCLUDED.owner
		RETURNING true
	`, leaseKey, owner, ttlSeconds).Scan(&acquired)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("acquire audit worker lease: %w", err)
	}
	return acquired, nil
}

func normalizeEvent(event Event) Event {
	event.CorrelationID = strings.TrimSpace(event.CorrelationID)
	event.Action = strings.TrimSpace(event.Action)
	event.EntityType = strings.TrimSpace(event.EntityType)
	event.EntityID = strings.TrimSpace(event.EntityID)
	event.ErrorCode = strings.TrimSpace(event.ErrorCode)
	event.RequestMethod = strings.ToUpper(strings.TrimSpace(event.RequestMethod))
	event.RequestPath = strings.TrimSpace(event.RequestPath)
	if event.Outcome == "" {
		event.Outcome = OutcomeSuccess
	}
	if event.Severity == "" {
		event.Severity = SeverityInfo
	}
	if event.Source == "" {
		event.Source = SourceAPI
	}
	return event
}

var ErrValidation = errors.New("audit validation failed")
