package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
)

func TestReportAnalysisPayloadPersistsProvenanceAndFallback(t *testing.T) {
	analysis := &entity.ReportAnalysis{ProviderCalled: true, Summary: "Mua sắm vượt 350.000 đ", ErrorCode: "provider_output_invalid"}
	payload, err := analysisPayload(analysis)
	if err != nil {
		t.Fatal(err)
	}
	if payload["provider_called"] != true || payload["summary"] != analysis.Summary || payload["error_code"] != analysis.ErrorCode {
		t.Fatalf("lost provenance: %+v", payload)
	}
}

func TestReportAnalysisCompletionRestoresEvidenceAndProvenance(t *testing.T) {
	db := openReportsPostgresTestDB(t)
	now := time.Now().UTC()
	ownerID := insertReportsTestUser(t, db, "Asia/Ho_Chi_Minh")
	snapshotID, analysisID := uuid.NewString(), uuid.NewString()
	insertExpiredReportSnapshot(t, db, ownerID, snapshotID, now.Add(entity.ReportSnapshotRetention))
	insertQueuedReportAnalysis(t, db, ownerID, snapshotID, analysisID, now)
	if err := db.Table("report_analyses").Where("id = ?", analysisID).Updates(map[string]any{"status": "running", "lease_token": "lease", "created_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	analysis := &entity.ReportAnalysis{ProviderCalled: true, Summary: "Mua sắm vượt 350.000 đ", ErrorCode: "provider_output_invalid", Observations: []entity.ReportInsight{{ID: "budget-over:shopping", Title: "Mua sắm", Metrics: []entity.ReportMetric{{Label: "Vượt", Amount: 350000}}, Drilldown: entity.ReportDrilldown{TargetID: "budget:shopping"}}}, Usage: map[string]any{"total_tokens": 42}}
	repo := NewReportAnalysisPostgresRepository(db)
	if err := repo.Complete(t.Context(), analysisID, "lease", now, analysis); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetOwned(t.Context(), ownerID, analysisID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.ProviderCalled || got.Summary != analysis.Summary || got.ErrorCode != analysis.ErrorCode || len(got.Observations) != 1 || got.Observations[0].Metrics[0].Amount != 350000 || got.Observations[0].Drilldown.TargetID != "budget:shopping" {
		t.Fatalf("completion lost facts/provenance: %+v", got)
	}
}

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
		if err := db.Exec(`INSERT INTO report_analyses(id, owner_id, snapshot_id, job_key, client_request_id, payload_hash, analysis_fingerprint, revision, rule_version, prompt_version, provider_profile_version, status, observations, suggestions, fact_refs, usage, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?, 'advisor-v1', ?, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, '{}'::jsonb, ?, ?)`, id, ownerID, snapshotID, id+":job", id+":request", id+":hash", id+":fingerprint", entity.ReportAnalysisRuleVersion, entity.ReportAnalysisPromptVersion, statuses[index], old, old).Error; err != nil {
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
