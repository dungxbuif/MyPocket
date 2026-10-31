package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
)

func TestReportAnalysisPostgresRetentionPreservesQueuedAndRunningJobs(t *testing.T) {
	db := openReportsPostgresTestDB(t)
	now := time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC)
	ownerID := insertReportsTestUser(t, db, "UTC")
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
