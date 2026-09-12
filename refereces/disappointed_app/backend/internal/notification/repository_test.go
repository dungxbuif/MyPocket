package notification_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/notification"
	"mypocket/internal/platform/db"
)

func TestPushSubscriptionIsDueOnlyForANewInboxNotice(t *testing.T) {
	conn := migratedNotificationPostgres(t)
	repo := notification.NewRepository(conn)
	ctx := context.Background()
	var userID string
	if err := conn.QueryRowContext(ctx, `INSERT INTO users (google_subject, email, email_verified) VALUES ('push-owner', 'push-owner@example.com', true) RETURNING id::text`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	subscription, err := repo.UpsertPushSubscription(ctx, userID, notification.CreatePushSubscriptionInput{Endpoint: "https://push.example/sub", P256DH: "p256dh", Auth: "auth"})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	assertDueSubscriptionCount(t, repo, now, 0)

	if _, _, err := repo.CreateNotice(ctx, notification.CreateNoticeInput{UserID: userID, Kind: "budget", Title: "Budget", DedupeKey: "budget-1"}); err != nil {
		t.Fatal(err)
	}
	assertDueSubscriptionCount(t, repo, now.Add(time.Minute), 1)
	if err := repo.RecordDeliverySuccess(ctx, subscription.ID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	assertDueSubscriptionCount(t, repo, now.Add(2*time.Minute), 0)

	time.Sleep(time.Millisecond)
	if _, _, err := repo.CreateNotice(ctx, notification.CreateNoticeInput{UserID: userID, Kind: "draft", Title: "Draft", DedupeKey: "draft-1"}); err != nil {
		t.Fatal(err)
	}
	assertDueSubscriptionCount(t, repo, now.Add(3*time.Minute), 1)
}

func assertDueSubscriptionCount(t *testing.T, repo *notification.Repository, now time.Time, want int) {
	t.Helper()
	due, err := repo.DuePushSubscriptions(context.Background(), now, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(due) != want {
		t.Fatalf("due subscriptions=%d, want %d: %#v", len(due), want, due)
	}
}

func migratedNotificationPostgres(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set; notification integration proof skipped")
	}
	conn, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.ExecContext(context.Background(), `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(context.Background(), conn, os.DirFS("../../migrations")); err != nil {
		t.Fatal(err)
	}
	return conn
}
