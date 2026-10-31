package usecase

import (
	"context"
	"errors"
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

func TestReportAnalysisBeginRequiresConfiguredProviderForNonEmptyReport(t *testing.T) {
	snapshot := reportSnapshotForAnalysis("owner-1", "snapshot-1", 1)
	analyses := &reportAnalysisRepositoryStub{}
	provider := &reportAnalysisProviderStub{available: false}
	service := NewReportAnalysisService(analyses, &reportSnapshotRepositoryStub{snapshot: snapshot}, provider)

	_, _, err := service.Begin(context.Background(), Principal{OwnerID: "owner-1"}, ReportAnalysisBeginRequest{SnapshotID: snapshot.ID, ClientRequestID: "request-1"})
	if !errors.Is(err, ErrReportAnalysisProviderUnavailable) {
		t.Fatalf("Begin() error = %v, want ErrReportAnalysisProviderUnavailable", err)
	}
	if analyses.created != nil || provider.calls != 0 {
		t.Fatalf("provider unavailable path created=%+v calls=%d", analyses.created, provider.calls)
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
