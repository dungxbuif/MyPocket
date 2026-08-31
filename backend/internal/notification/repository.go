package notification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrForbidden = errors.New("notification resource is not owned by user")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateNotice(ctx context.Context, input CreateNoticeInput) (Notice, bool, error) {
	if strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.DedupeKey) == "" {
		return Notice{}, false, fmt.Errorf("user, title, and dedupe key are required")
	}
	var n Notice
	var readAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO notifications (user_id, kind, title, body, source_type, source_id, dedupe_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, dedupe_key) DO NOTHING
		RETURNING id::text, user_id::text, kind, title, body, source_type, source_id, read_at, created_at
	`, input.UserID, strings.TrimSpace(input.Kind), strings.TrimSpace(input.Title), input.Body, input.SourceType, input.SourceID, input.DedupeKey).Scan(
		&n.ID, &n.UserID, &n.Kind, &n.Title, &n.Body, &n.SourceType, &n.SourceID, &readAt, &n.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Notice{}, false, nil
	}
	if err != nil {
		return Notice{}, false, fmt.Errorf("create notice: %w", err)
	}
	n.ReadAt = nullableTime(readAt)
	return n, true, nil
}

func (r *Repository) List(ctx context.Context, userID string, limit int, before time.Time) ([]Notice, bool, error) {
	if limit <= 0 || limit > MaxPageSize {
		limit = MaxPageSize
	}
	args := []any{userID, limit + 1}
	query := `SELECT id::text, user_id::text, kind, title, body, source_type, source_id, read_at, created_at FROM notifications WHERE user_id = $1`
	if !before.IsZero() {
		query += ` AND created_at < $2`
		args = []any{userID, before, limit + 1}
	}
	query += ` ORDER BY created_at DESC, id DESC LIMIT $` + fmt.Sprint(len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	items := make([]Notice, 0, limit)
	for rows.Next() {
		var n Notice
		var readAt sql.NullTime
		if err := rows.Scan(&n.ID, &n.UserID, &n.Kind, &n.Title, &n.Body, &n.SourceType, &n.SourceID, &readAt, &n.CreatedAt); err != nil {
			return nil, false, fmt.Errorf("scan notification: %w", err)
		}
		n.ReadAt = nullableTime(readAt)
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return items, hasMore, nil
}

func (r *Repository) MarkRead(ctx context.Context, userID, id string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE notifications SET read_at = COALESCE(read_at, now()) WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrForbidden
	}
	return nil
}

func (r *Repository) UpsertPushSubscription(ctx context.Context, userID string, input CreatePushSubscriptionInput) (PushSubscription, error) {
	if strings.TrimSpace(input.Endpoint) == "" || strings.TrimSpace(input.P256DH) == "" || strings.TrimSpace(input.Auth) == "" {
		return PushSubscription{}, fmt.Errorf("push endpoint and keys are required")
	}
	var s PushSubscription
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, endpoint) DO UPDATE SET p256dh = EXCLUDED.p256dh, auth = EXCLUDED.auth, expires_at = EXCLUDED.expires_at, failure_count = 0, next_attempt_at = now(), updated_at = now()
		RETURNING id::text, endpoint, expires_at, created_at
	`, userID, input.Endpoint, input.P256DH, input.Auth, input.ExpiresAt).Scan(&s.ID, &s.Endpoint, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return PushSubscription{}, fmt.Errorf("upsert push subscription: %w", err)
	}
	return s, nil
}

func (r *Repository) DeletePushSubscription(ctx context.Context, userID, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM push_subscriptions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete push subscription: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrForbidden
	}
	return nil
}

func (r *Repository) DuePushSubscriptions(ctx context.Context, now time.Time, limit int) ([]DeliverySubscription, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `SELECT user_id::text, id::text, endpoint, p256dh, auth FROM push_subscriptions WHERE next_attempt_at <= $1 AND (expires_at IS NULL OR expires_at > $1) AND failure_count < $2 ORDER BY next_attempt_at, id LIMIT $3`, now, MaxDeliveryAttempts, limit)
	if err != nil {
		return nil, fmt.Errorf("list due push subscriptions: %w", err)
	}
	defer rows.Close()
	result := make([]DeliverySubscription, 0, limit)
	for rows.Next() {
		var s DeliverySubscription
		if err := rows.Scan(&s.UserID, &s.ID, &s.Endpoint, &s.P256DH, &s.Auth); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// CreateDueNotices materializes authoritative planning signals without touching accounting balances.
func (r *Repository) CreateDueNotices(ctx context.Context, now time.Time) (int, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (user_id, kind, title, body, source_type, source_id, dedupe_key)
		SELECT ba.user_id, 'budget_threshold', 'Ngân sách cần chú ý',
			CASE WHEN ba.threshold >= 100 THEN 'Ngân sách đã vượt 100%' ELSE 'Ngân sách đã chạm 80%' END,
			'budget', ba.budget_id::text, 'budget-alert:' || ba.budget_id::text || ':' || ba.threshold::text || ':' || ba.period_start::text
		FROM budget_alerts ba
		ON CONFLICT (user_id, dedupe_key) DO NOTHING;
		INSERT INTO notifications (user_id, kind, title, body, source_type, source_id, dedupe_key)
		SELECT o.user_id, 'obligation_due', 'Khoản cần thanh toán', o.counterparty || ' đến hạn ' || o.due_on::text,
			'obligation', o.id::text, 'obligation-due:' || o.id::text || ':' || o.due_on::text
		FROM obligations o WHERE o.archived_at IS NULL AND o.due_on <= ($1 AT TIME ZONE 'Asia/Ho_Chi_Minh')::date;
		INSERT INTO notifications (user_id, kind, title, body, source_type, source_id, dedupe_key)
		SELECT d.user_id, 'transaction_draft', 'Có bản nháp giao dịch', COALESCE(NULLIF(d.note, ''), 'Bản nháp lặp đang chờ duyệt'),
			'draft', d.id::text, 'draft:' || d.id::text
		FROM transaction_drafts d WHERE d.status = 'pending';
	`, now)
	if err != nil {
		return 0, fmt.Errorf("create due notices: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *Repository) RecordDeliveryFailure(ctx context.Context, id string, expired bool, now time.Time) error {
	if expired {
		_, err := r.db.ExecContext(ctx, `DELETE FROM push_subscriptions WHERE id = $1`, id)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE push_subscriptions SET failure_count = failure_count + 1, next_attempt_at = $2 + LEAST((failure_count + 1) * INTERVAL '1 minute', INTERVAL '1 hour'), updated_at = $2 WHERE id = $1 AND failure_count < $3`, id, now, MaxDeliveryAttempts)
	return err
}

func (r *Repository) RecordDeliverySuccess(ctx context.Context, id string, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE push_subscriptions SET failure_count = 0, next_attempt_at = $2, updated_at = $2 WHERE id = $1`, id, now)
	return err
}

func nullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}
