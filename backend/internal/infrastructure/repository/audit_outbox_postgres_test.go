package repository

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/infrastructure/db"
)

func TestAuditOutboxPostgresEnqueueLeaseRetryAndDelivery(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	database, err := db.NewPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	database.Exec("DELETE FROM audit_outbox WHERE event_id LIKE 'test-%'")
	defer database.Exec("DELETE FROM audit_outbox WHERE event_id LIKE 'test-%'")
	repo := NewAuditOutboxPostgresRepository(database)
	ctx := context.Background()
	stream, eventID := "mypocket:audit:feedback", "test-"+uuid.NewString()
	now := time.Now().UTC().Truncate(time.Microsecond)
	payload := []byte(`{"action":"feedback.created","event_id":"` + eventID + `"}`)
	if err := repo.EnqueueAudit(ctx, stream, eventID, payload, now); err != nil {
		t.Fatal(err)
	}
	if err := repo.EnqueueAudit(ctx, stream, eventID, payload, now); err != nil {
		t.Fatal(err)
	}
	rows, err := repo.ClaimAudit(ctx, now.Add(time.Second), 10, time.Minute)
	if err != nil || len(rows) != 1 {
		t.Fatalf("claim rows=%d err=%v", len(rows), err)
	}
	var gotPayload, wantPayload map[string]any
	_ = json.Unmarshal(rows[0].Payload, &gotPayload)
	_ = json.Unmarshal(payload, &wantPayload)
	if rows[0].Attempts != 1 || rows[0].EventID != eventID || !reflect.DeepEqual(gotPayload, wantPayload) {
		t.Fatalf("unexpected claimed row: attempts=%d event=%q want=%q payload=%q want=%q", rows[0].Attempts, rows[0].EventID, eventID, rows[0].Payload, payload)
	}
	if retry, err := repo.ClaimAudit(ctx, now.Add(2*time.Second), 10, time.Minute); err != nil || len(retry) != 0 {
		t.Fatalf("leased row should not be reclaimed: rows=%d err=%v", len(retry), err)
	}
	if err := repo.MarkAuditFailed(ctx, rows[0].ID, now.Add(2*time.Minute), "redis unavailable"); err != nil {
		t.Fatal(err)
	}
	retry, err := repo.ClaimAudit(ctx, now.Add(2*time.Minute), 10, time.Minute)
	if err != nil || len(retry) != 1 || retry[0].Attempts != 2 {
		t.Fatalf("failed row should retry with incremented attempts: rows=%+v err=%v", retry, err)
	}
	if err := repo.MarkAuditDelivered(ctx, retry[0].ID, now.Add(3*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if delivered, err := repo.ClaimAudit(ctx, now.Add(4*time.Minute), 10, time.Minute); err != nil || len(delivered) != 0 {
		t.Fatalf("delivered row should not be claimed: rows=%d err=%v", len(delivered), err)
	}
	// Keep the disposable fixture clean without touching any application data.
	database.Exec("DELETE FROM audit_outbox WHERE event_id = ?", eventID)
}
