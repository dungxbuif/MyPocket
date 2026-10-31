package repository

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openReportsRetentionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func insertReportRetentionUser(t *testing.T, db *gorm.DB) string {
	t.Helper()
	ownerID := uuid.NewString()
	if err := db.Create(&entity.User{ID: ownerID, Email: ownerID + "@reports-retention.test", GoogleSubject: ownerID, Timezone: "Asia/Ho_Chi_Minh", TimezoneConfirmed: true}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM "user" WHERE id = ?`, ownerID) })
	return ownerID
}

func insertExpiredReportSnapshot(t *testing.T, db *gorm.DB, ownerID string, id string, now time.Time) {
	t.Helper()
	createdAt := now.Add(-entity.ReportSnapshotRetention - time.Hour)
	if err := db.Exec(`INSERT INTO report_snapshots(id, owner_id, month, wallet_id, granularity, timezone, as_of_day, data_revision, rule_version, as_of, expires_at, created_at, payload)
VALUES (?, ?, '2026-10', '', 'day', 'UTC', ?, 1, ?, ?, ?, ?, ?::jsonb)`, id, ownerID, createdAt.Format("2006-01-02"), entity.MonthlyReportRuleVersion, createdAt, now.Add(-time.Minute), createdAt, `{"scope":{"month":"2026-10","granularity":"day"},"totals":{"transaction_count":1}}`).Error; err != nil {
		t.Fatal(err)
	}
}

func insertQueuedReportAnalysis(t *testing.T, db *gorm.DB, ownerID, snapshotID, id string, now time.Time) {
	t.Helper()
	if err := db.Exec(`INSERT INTO report_analyses(id, owner_id, snapshot_id, job_key, client_request_id, payload_hash, revision, rule_version, prompt_version, provider_profile_version, status, observations, suggestions, fact_refs, usage, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, 'advisor-v1', ?, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, '{}'::jsonb, ?, ?)`, id, ownerID, snapshotID, id+":job", id+":request", id+":hash", entity.ReportAnalysisRuleVersion, entity.ReportAnalysisPromptVersion, entity.ReportAnalysisStatusQueued, now.Add(-entity.ReportSnapshotRetention-time.Hour), now.Add(-entity.ReportSnapshotRetention-time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
}

func TestReportSnapshotPostgresGetOwnedIsOwnerScopedAndExpiryAware(t *testing.T) {
	db := openReportsRetentionTestDB(t)
	now := time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC)
	ownerID := insertReportRetentionUser(t, db)
	otherID := insertReportRetentionUser(t, db)
	activeID := uuid.NewString()
	insertExpiredReportSnapshot(t, db, ownerID, activeID, now.Add(entity.ReportSnapshotRetention))

	repository := NewReportSnapshotPostgresRepository(db)
	got, err := repository.GetOwned(t.Context(), ownerID, activeID, now)
	if err != nil {
		t.Fatalf("GetOwned active snapshot: %v", err)
	}
	if got.ID != activeID || got.OwnerID != ownerID || got.Report.Scope.Month != "2026-10" {
		t.Fatalf("decoded snapshot = %+v", got)
	}
	if _, err := repository.GetOwned(t.Context(), otherID, activeID, now); !errors.Is(err, contract.ErrNotFound) {
		t.Fatalf("cross-owner snapshot lookup error = %v, want not found", err)
	}

	expiredID := uuid.NewString()
	insertExpiredReportSnapshot(t, db, ownerID, expiredID, now)
	if _, err := repository.GetOwned(t.Context(), ownerID, expiredID, now); !errors.Is(err, contract.ErrSnapshotExpired) {
		t.Fatalf("expired snapshot lookup error = %v, want snapshot expired", err)
	}
}

func TestReportSnapshotPostgresRetentionSkipsSnapshotsWithLiveAnalysis(t *testing.T) {
	db := openReportsRetentionTestDB(t)
	now := time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC)
	ownerID := insertReportsRetentionUserWithCleanup(t, db)
	liveID, expiredID := uuid.NewString(), uuid.NewString()
	insertExpiredReportSnapshot(t, db, ownerID, liveID, now)
	insertExpiredReportSnapshot(t, db, ownerID, expiredID, now)
	analysisID := uuid.NewString()
	insertQueuedReportAnalysis(t, db, ownerID, liveID, analysisID, now)

	deleted, err := NewReportSnapshotPostgresRepository(db).DeleteExpired(t.Context(), now, 10)
	if err != nil {
		t.Fatalf("DeleteExpired snapshots: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted snapshots = %d, want 1", deleted)
	}
	var remaining int64
	if err := db.Table("report_snapshots").Where("id IN ?", []string{liveID, expiredID}).Count(&remaining).Error; err != nil {
		t.Fatal(err)
	}
	if remaining != 1 {
		t.Fatalf("remaining snapshots = %d, want live snapshot only", remaining)
	}
	var liveAnalysis int64
	if err := db.Table("report_analyses").Where("id = ?", analysisID).Count(&liveAnalysis).Error; err != nil {
		t.Fatal(err)
	}
	if liveAnalysis != 1 {
		t.Fatal("live analysis must be preserved with its snapshot")
	}
}

func TestReportAnalysisPostgresRetentionPreservesQueuedAndRunningJobs(t *testing.T) {
	db := openReportsRetentionTestDB(t)
	now := time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC)
	ownerID := insertReportsRetentionUserWithCleanup(t, db)
	snapshotID := uuid.NewString()
	insertExpiredReportSnapshot(t, db, ownerID, snapshotID, now.Add(2*entity.ReportSnapshotRetention))
	old := now.Add(-entity.ReportSnapshotRetention - time.Hour)
	ids := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
	statuses := []string{entity.ReportAnalysisStatusCompleted, entity.ReportAnalysisStatusQueued, entity.ReportAnalysisStatusRunning}
	for index, id := range ids {
		if err := db.Exec(`INSERT INTO report_analyses(id, owner_id, snapshot_id, job_key, client_request_id, payload_hash, revision, rule_version, prompt_version, provider_profile_version, status, observations, suggestions, fact_refs, usage, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, 'advisor-v1', ?, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, '{}'::jsonb, ?, ?)`, id, ownerID, snapshotID, id+":job", id+":request", id+":hash", entity.ReportAnalysisRuleVersion, entity.ReportAnalysisPromptVersion, statuses[index], old, old).Error; err != nil {
			t.Fatal(err)
		}
	}

	deleted, err := NewReportAnalysisPostgresRepository(db).DeleteExpired(t.Context(), now, 10)
	if err != nil {
		t.Fatalf("DeleteExpired analyses: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted analyses = %d, want completed job only", deleted)
	}
	var remaining int64
	if err := db.Table("report_analyses").Where("id IN ?", ids[1:]).Count(&remaining).Error; err != nil {
		t.Fatal(err)
	}
	if remaining != 2 {
		t.Fatalf("remaining live analyses = %d, want 2", remaining)
	}
}

func insertReportsRetentionUserWithCleanup(t *testing.T, db *gorm.DB) string {
	t.Helper()
	ownerID := uuid.NewString()
	if err := db.Create(&entity.User{ID: ownerID, Email: fmt.Sprintf("%s@reports-retention.test", ownerID), GoogleSubject: ownerID, Timezone: "UTC", TimezoneConfirmed: true}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM "user" WHERE id = ?`, ownerID) })
	return ownerID
}
