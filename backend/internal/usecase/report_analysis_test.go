package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type reportSnapshotRepositoryStub struct {
	snapshot *entity.ReportSnapshot
}

func (s *reportSnapshotRepositoryStub) ReadOrCreate(context.Context, string, entity.MonthlyReportScope, time.Time) (*entity.ReportSnapshot, error) {
	return s.snapshot, nil
}
func (s *reportSnapshotRepositoryStub) GetOwned(_ context.Context, ownerID, snapshotID string, _ time.Time) (*entity.ReportSnapshot, error) {
	if s.snapshot == nil || ownerID != s.snapshot.OwnerID || snapshotID != s.snapshot.ID {
		return nil, repository.ErrNotFound
	}
	return s.snapshot, nil
}
func (s *reportSnapshotRepositoryStub) CurrentRevision(context.Context, string) (int64, error) {
	return 1, nil
}
func (s *reportSnapshotRepositoryStub) GetRetainedOwned(ctx context.Context, ownerID, snapshotID string, now time.Time) (*entity.ReportSnapshot, error) {
	return s.GetOwned(ctx, ownerID, snapshotID, now)
}
func (s *reportSnapshotRepositoryStub) DeleteExpired(context.Context, time.Time, int) (int, error) {
	return 0, nil
}

type reportAnalysisRepositoryStub struct {
	created          *entity.ReportAnalysis
	createReplay     bool
	existing         *entity.ReportAnalysis
	lease            *entity.ReportAnalysis
	completed        *entity.ReportAnalysis
	failedCode       string
	recoverCalled    bool
	providerCallSeen int
}

func (r *reportAnalysisRepositoryStub) CreateOrGet(_ context.Context, analysis *entity.ReportAnalysis) (*entity.ReportAnalysis, bool, error) {
	if r.existing != nil {
		return r.existing, r.createReplay, nil
	}
	r.created = analysis
	return analysis, false, nil
}
func (r *reportAnalysisRepositoryStub) GetOwned(context.Context, string, string) (*entity.ReportAnalysis, error) {
	if r.created == nil {
		return nil, repository.ErrNotFound
	}
	return r.created, nil
}

func TestReportFollowUpContextKeepsSavedScopeAndRejectsMismatchedIdentifiers(t *testing.T) {
	snapshot := reportSnapshotForAnalysis("owner", "snapshot", 10)
	snapshot.Report.Scope.WalletID = "wallet-a"
	analysis := &entity.ReportAnalysis{ID: "analysis", SnapshotID: "snapshot", Status: entity.ReportAnalysisStatusCompleted, RetainedUntil: snapshot.AsOf.Add(time.Hour), Observations: []entity.ReportInsight{{ID: "finding", Title: "Mua sắm vượt 350.000 đ"}}}
	analysis.Suggestions = []entity.ReportInsight{{ID: "legacy-suggestion", Body: "Gợi ý được lưu trước đó"}}
	service := NewReportAnalysisService(&reportAnalysisRepositoryStub{created: analysis}, &reportSnapshotRepositoryStub{snapshot: snapshot}, nil)
	service.Now = func() time.Time { return snapshot.AsOf }
	request := ReportFollowUpRequest{AnalysisID: "analysis", SnapshotID: "snapshot", FindingID: "finding"}
	result, err := service.FollowUpContext(context.Background(), Principal{OwnerID: "owner"}, request)
	if err != nil || result["scope"].(entity.MonthlyReportScope).WalletID != "wallet-a" {
		t.Fatalf("context %+v %v", result, err)
	}
	for _, bad := range []ReportFollowUpRequest{{AnalysisID: "analysis", SnapshotID: "different", FindingID: "finding"}, {AnalysisID: "analysis", SnapshotID: "snapshot", FindingID: "missing"}} {
		if _, err := service.FollowUpContext(context.Background(), Principal{OwnerID: "owner"}, bad); err == nil {
			t.Fatal("mismatched context accepted")
		}
	}
	if _, err := service.FollowUpContext(context.Background(), Principal{OwnerID: "foreign"}, request); err == nil {
		t.Fatal("foreign snapshot accepted")
	}
	legacyRequest := request
	legacyRequest.FindingID = "legacy-suggestion"
	legacy, err := service.FollowUpContext(context.Background(), Principal{OwnerID: "owner"}, legacyRequest)
	if err != nil || legacy["findings"].([]entity.ReportInsight)[0].ID != "legacy-suggestion" {
		t.Fatalf("legacy suggestion rejected: %+v %v", legacy, err)
	}
	text := messageText(entity.AdvisorMessage{Parts: []entity.AdvisorPart{{Type: "text", Text: "Khoản nào cần xem?"}, {Type: "report_context", Data: result}}})
	if !strings.Contains(text, "wallet-a") || !strings.Contains(text, "finding") {
		t.Fatal("saved context absent from model input")
	}
}
func (r *reportAnalysisRepositoryStub) Cancel(context.Context, string, string, time.Time) (*entity.ReportAnalysis, error) {
	return nil, repository.ErrNotFound
}
func (r *reportAnalysisRepositoryStub) ClaimNext(context.Context, time.Time, time.Duration) (*entity.ReportAnalysis, error) {
	if r.lease == nil {
		return nil, repository.ErrNotFound
	}
	lease := r.lease
	r.lease = nil
	return lease, nil
}
func (r *reportAnalysisRepositoryStub) RecoverStale(context.Context, time.Time, int) error {
	r.recoverCalled = true
	return nil
}
func (r *reportAnalysisRepositoryStub) Complete(_ context.Context, _ string, _ string, _ time.Time, analysis *entity.ReportAnalysis) error {
	r.completed = analysis
	return nil
}
func (r *reportAnalysisRepositoryStub) Fail(_ context.Context, _ string, _ string, _ time.Time, code string) error {
	r.failedCode = code
	return nil
}
func (r *reportAnalysisRepositoryStub) DeleteExpired(context.Context, time.Time, int) (int, error) {
	return 0, nil
}
func (r *reportAnalysisRepositoryStub) LatestForScope(context.Context, string, entity.MonthlyReportScope, string, time.Time) (*entity.ReportAnalysis, *entity.ReportAnalysis, error) {
	return r.completed, nil, nil
}
func (r *reportAnalysisRepositoryStub) RenewLease(context.Context, string, string, time.Time, time.Duration) error {
	return nil
}

type reportAnalysisProviderStub struct {
	available bool
	calls     int
	result    entity.ReportProviderResult
}

func (p *reportAnalysisProviderStub) Available() bool { return p.available }
func (p *reportAnalysisProviderStub) Analyze(context.Context, entity.ReportFactBundle) (entity.ReportProviderResult, error) {
	p.calls++
	return p.result, nil
}

func reportSnapshotForAnalysis(ownerID, snapshotID string, transactionCount int64) *entity.ReportSnapshot {
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	return &entity.ReportSnapshot{
		ID: snapshotID, OwnerID: ownerID, DataRevision: 4, AsOf: now,
		Report: entity.MonthlyReport{
			Scope: entity.MonthlyReportScope{Month: "2026-10", Granularity: "day"},
			AsOf:  now, Totals: entity.ReportTotals{Income: 100, Expense: 50, Net: 50, TransactionCount: transactionCount},
		},
	}
}

func TestReportAnalysisBeginCompletesNoDataWithoutProviderCall(t *testing.T) {
	snapshot := reportSnapshotForAnalysis("owner-1", "snapshot-1", 0)
	analyses := &reportAnalysisRepositoryStub{}
	service := NewReportAnalysisService(analyses, &reportSnapshotRepositoryStub{snapshot: snapshot}, nil)
	service.Now = func() time.Time { return snapshot.AsOf }

	analysis, replay, err := service.Begin(context.Background(), Principal{OwnerID: "owner-1"}, ReportAnalysisBeginRequest{SnapshotID: snapshot.ID, ClientRequestID: "request-1"})
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	if replay || analysis.Status != entity.ReportAnalysisStatusCompleted || analysis.ErrorCode != "no_data" || analyses.created.ProviderCalled {
		t.Fatalf("no-data analysis = %+v replay=%v", analysis, replay)
	}
}

func TestReportAnalysisBeginSkipsProviderWhenEvidenceIsInsufficient(t *testing.T) {
	snapshot := reportSnapshotForAnalysis("owner-1", "snapshot-1", 1)
	analyses := &reportAnalysisRepositoryStub{}
	provider := &reportAnalysisProviderStub{available: false}
	service := NewReportAnalysisService(analyses, &reportSnapshotRepositoryStub{snapshot: snapshot}, provider)

	result, _, err := service.Begin(context.Background(), Principal{OwnerID: "owner-1"}, ReportAnalysisBeginRequest{SnapshotID: snapshot.ID, ClientRequestID: "request-1"})
	if err != nil || result.Status != entity.ReportAnalysisStatusCompleted || result.ErrorCode != "insufficient_evidence" || provider.calls != 0 {
		t.Fatalf("insufficient evidence result=%+v calls=%d err=%v", result, provider.calls, err)
	}
}

func TestReportAnalysisBeginRejectsOldQueryRules(t *testing.T) {
	snapshot := reportSnapshotForAnalysis("owner", "snapshot", 10)
	snapshot.RuleVersion = "monthly-report-v2"
	analyses := &reportAnalysisRepositoryStub{}
	service := NewReportAnalysisService(analyses, &reportSnapshotRepositoryStub{snapshot: snapshot}, nil)
	_, _, err := service.Begin(context.Background(), Principal{OwnerID: "owner"}, ReportAnalysisBeginRequest{SnapshotID: "snapshot", ClientRequestID: "new-analysis"})
	if err != entity.ErrReportSnapshotStale || analyses.created != nil {
		t.Fatalf("old calculation accepted: %v %+v", err, analyses.created)
	}
}

func TestProviderSelectionRendersExactServerEvidenceAndScopedTarget(t *testing.T) {
	report := reportSnapshotForAnalysis("owner", "snapshot", 10).Report
	report.Budgets = []entity.ReportBudgetPreview{{ID: "shopping", Name: "Mua sắm", FactRef: "budget:shopping", LimitAmount: 1500000, Spent: 1850000, Remaining: -350000, OverLimit: true}}
	result := entity.ReportProviderResult{Observations: []entity.ReportProviderObservation{{FindingID: "budget-over:shopping", Intent: "review_plan"}}}
	analysis := &entity.ReportAnalysis{}
	applyProviderAnalysis(analysis, report, result)
	if len(analysis.Observations) != 1 || len(analysis.Suggestions) != 0 {
		t.Fatalf("output %+v", analysis)
	}
	item := analysis.Observations[0]
	if !strings.Contains(item.Title, "350.000") || item.Drilldown.TargetID != "budget:shopping" || item.Metrics[0].Amount != 1850000 {
		t.Fatalf("evidence %+v", item)
	}
	if !strings.Contains(analysis.Summary, "Chưa đủ") {
		t.Fatalf("comparison guard absent: %s", analysis.Summary)
	}
}

func TestReportFallbackPreservesDistinctEvidenceKinds(t *testing.T) {
	report := reportSnapshotForAnalysis("owner", "snapshot", 10).Report
	for _, id := range []string{"a", "b", "c"} {
		report.Budgets = append(report.Budgets, entity.ReportBudgetPreview{ID: id, Name: id, LimitAmount: 1000000, Spent: 2000000, OverLimit: true})
	}
	categoryID := "food"
	bps := int64(10000)
	report.Comparison = entity.ReportComparison{Comparable: true, CommonDays: 10}
	report.Contributors = []entity.ReportContributor{{CategoryID: &categoryID, Name: "Ăn uống", FactRef: "category:food", CurrentAmount: 2000000, PreviousAmount: 1000000, Delta: 1000000, CurrentCount: 10, PreviousCount: 10, ChangeBPS: &bps}}
	analysis := &entity.ReportAnalysis{}
	applyDeterministicAnalysis(analysis, report, "provider_empty")
	found := false
	for _, item := range analysis.Observations {
		if item.Kind == entity.ReportInsightKindCategory {
			found = true
		}
	}
	if len(analysis.Observations) != 3 || !found {
		t.Fatalf("fallback lost category evidence: %+v", analysis.Observations)
	}
}

func TestReportAnalysisFingerprintIgnoresRenderTimeButTracksEligibilityAndNames(t *testing.T) {
	snapshot := reportSnapshotForAnalysis("owner", "snapshot", 10)
	snapshot.Report.Timezone = "Asia/Ho_Chi_Minh"
	snapshot.Report.CurrentRange = entity.ReportRange{From: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), EndExclusive: snapshot.AsOf, IsPartial: true}
	snapshot.Report.Budgets = []entity.ReportBudgetPreview{{ID: "b", Name: "Ăn uống", Spent: 200, LimitAmount: 100, OverLimit: true}}
	fingerprint := func() string {
		facts, _ := entity.BuildReportFactBundle(*snapshot)
		value, err := reportAnalysisFingerprint(*snapshot, facts)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	original := fingerprint()
	snapshot.Report.AsOf = snapshot.Report.AsOf.Add(time.Hour)
	snapshot.Report.CurrentRange.EndExclusive = snapshot.Report.AsOf
	snapshot.Report.Scope.Granularity = "week"
	if fingerprint() != original {
		t.Fatal("render time/granularity invalidates cache")
	}
	snapshot.Report.CurrentRange.EndExclusive = snapshot.Report.CurrentRange.EndExclusive.AddDate(0, 0, 1)
	if fingerprint() == original {
		t.Fatal("calendar coverage must invalidate cache")
	}
	snapshot.Report.CurrentRange.EndExclusive = snapshot.Report.CurrentRange.EndExclusive.AddDate(0, 0, -1)
	snapshot.Report.Budgets[0].Name = "Tên mới"
	if fingerprint() == original {
		t.Fatal("renamed evidence must invalidate cache")
	}
	original = fingerprint()
	snapshot.Report.Comparison.Comparable = true
	if fingerprint() == original {
		t.Fatal("eligibility must invalidate cache")
	}
}

func TestReportAnalysisBeginReturnsExistingReplayWithoutCreatingAnotherJob(t *testing.T) {
	snapshot := reportSnapshotForAnalysis("owner-1", "snapshot-1", 1)
	existing := &entity.ReportAnalysis{ID: "analysis-1", Status: entity.ReportAnalysisStatusQueued}
	analyses := &reportAnalysisRepositoryStub{existing: existing, createReplay: true}
	provider := &reportAnalysisProviderStub{available: true}
	service := NewReportAnalysisService(analyses, &reportSnapshotRepositoryStub{snapshot: snapshot}, provider)

	analysis, replay, err := service.Begin(context.Background(), Principal{OwnerID: "owner-1"}, ReportAnalysisBeginRequest{SnapshotID: snapshot.ID, ClientRequestID: "request-1"})
	if err != nil || !replay || analysis != existing {
		t.Fatalf("replay result=%+v replay=%v err=%v", analysis, replay, err)
	}
	if provider.calls != 0 || analyses.created != nil {
		t.Fatalf("replay started provider or created duplicate: calls=%d created=%+v", provider.calls, analyses.created)
	}
}

func TestReportAnalysisWorkerFallsBackWhenProviderOutputIsInvalid(t *testing.T) {
	snapshot := reportSnapshotForAnalysis("owner-1", "snapshot-1", 1)
	analyses := &reportAnalysisRepositoryStub{lease: &entity.ReportAnalysis{ID: "analysis-1", OwnerID: "owner-1", SnapshotID: snapshot.ID, LeaseToken: "lease-1", Status: entity.ReportAnalysisStatusRunning}}
	provider := &reportAnalysisProviderStub{available: true, result: entity.ReportProviderResult{Observations: []entity.ReportProviderObservation{{FactRef: "totals.expense", Text: "Chi tăng 12 phần trăm"}}}}
	worker := NewReportAnalysisWorker(analyses, &reportSnapshotRepositoryStub{snapshot: snapshot}, provider)
	worker.Now = func() time.Time { return snapshot.AsOf }

	processed, err := worker.RunPass(context.Background(), 1)
	if err != nil || processed != 1 {
		t.Fatalf("RunPass() processed=%d err=%v", processed, err)
	}
	if !analyses.recoverCalled || provider.calls != 1 || analyses.failedCode != "" || analyses.completed == nil {
		t.Fatalf("worker state recover=%v provider_calls=%d failed=%q completed=%+v", analyses.recoverCalled, provider.calls, analyses.failedCode, analyses.completed)
	}
	if analyses.completed.Status != entity.ReportAnalysisStatusCompleted || analyses.completed.ErrorCode != "provider_output_invalid" {
		t.Fatalf("invalid provider completion = %+v", analyses.completed)
	}
}
