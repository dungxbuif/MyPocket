package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

var (
	ErrReportAnalysisProviderUnavailable = errors.New("report analysis provider unavailable")
	ErrReportAnalysisBusy                = errors.New("another report analysis is active")
	ErrReportAnalysisRequestConflict     = errors.New("report analysis request conflicts with an existing request")
	ErrReportAnalysisCooldown            = errors.New("report analysis cooldown is active")
)

type ReportAnalysisProvider interface {
	Available() bool
	Analyze(context.Context, entity.ReportFactBundle) (entity.ReportProviderResult, error)
}

type ReportAnalysisBeginRequest struct {
	SnapshotID      string
	ClientRequestID string
}

type ReportFollowUpRequest struct {
	AnalysisID string `json:"analysis_id"`
	SnapshotID string `json:"snapshot_id"`
	FindingID  string `json:"finding_id,omitempty"`
}

// FollowUpContext accepts identifiers only. Every financial claim is reloaded
// from the owner's saved analysis, never trusted from client supplied prose.
func (s *ReportAnalysisService) FollowUpContext(ctx context.Context, principal Principal, input ReportFollowUpRequest) (map[string]any, error) {
	analysis, err := s.Get(ctx, principal, input.AnalysisID)
	if err != nil {
		return nil, err
	}
	if analysis.Status != entity.ReportAnalysisStatusCompleted || analysis.SnapshotID != input.SnapshotID || s.Snapshots == nil {
		return nil, repository.ErrNotFound
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	snapshot, err := s.Snapshots.GetRetainedOwned(ctx, principal.OwnerID, analysis.SnapshotID, now)
	if err != nil {
		return nil, err
	}
	findings := append(append([]entity.ReportInsight(nil), analysis.Observations...), analysis.Suggestions...)
	if input.FindingID != "" {
		allFindings := findings
		findings = nil
		for _, item := range allFindings {
			if item.ID == input.FindingID {
				findings = append(findings, item)
				break
			}
		}
		if len(findings) == 0 {
			return nil, repository.ErrNotFound
		}
	}
	return map[string]any{"analysis_id": analysis.ID, "snapshot_id": snapshot.ID, "scope": snapshot.Report.Scope, "timezone": snapshot.Timezone, "as_of": snapshot.AsOf, "summary": analysis.Summary, "findings": findings, "instruction": "Dữ liệu đã lưu tại as_of, không phải dữ liệu hiện tại. Giữ nguyên phạm vi tháng/ví; dùng công cụ truy vấn khi cần số liệu mới. Không suy diễn nguyên nhân từ tổng tiền."}, nil
}

type ReportAnalysisService struct {
	Analyses  repository.ReportAnalysisRepository
	Snapshots repository.ReportSnapshotRepository
	Provider  ReportAnalysisProvider
	Now       func() time.Time
}

type ReportAnalysisLatest struct {
	Completed         *entity.ReportAnalysis `json:"completed"`
	Active            *entity.ReportAnalysis `json:"active"`
	Freshness         string                 `json:"freshness"`
	StaleReasons      []string               `json:"stale_reasons"`
	CurrentSnapshotID string                 `json:"current_snapshot_id"`
}

func NewReportAnalysisService(analyses repository.ReportAnalysisRepository, snapshots repository.ReportSnapshotRepository, provider ReportAnalysisProvider) *ReportAnalysisService {
	return &ReportAnalysisService{Analyses: analyses, Snapshots: snapshots, Provider: provider, Now: time.Now}
}

func (s *ReportAnalysisService) Begin(ctx context.Context, principal Principal, request ReportAnalysisBeginRequest) (*entity.ReportAnalysis, bool, error) {
	if s == nil || s.Analyses == nil || s.Snapshots == nil {
		return nil, false, repository.ErrNotFound
	}
	request.SnapshotID = strings.TrimSpace(request.SnapshotID)
	request.ClientRequestID = strings.TrimSpace(request.ClientRequestID)
	if request.SnapshotID == "" || len(request.SnapshotID) > 128 || request.ClientRequestID == "" || len(request.ClientRequestID) > 128 {
		return nil, false, entity.ErrMonthlyReportScopeInvalid
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	snapshot, err := s.Snapshots.GetOwned(ctx, principal.OwnerID, request.SnapshotID, now)
	if err != nil {
		return nil, false, err
	}
	if snapshot.RuleVersion != "" && snapshot.RuleVersion != entity.MonthlyReportRuleVersion {
		return nil, false, entity.ErrReportSnapshotStale
	}
	facts, err := entity.BuildReportFactBundle(*snapshot)
	if err != nil {
		return nil, false, err
	}
	fingerprint, err := reportAnalysisFingerprint(*snapshot, facts)
	if err != nil {
		return nil, false, err
	}
	jobKey := fingerprint + ":" + request.ClientRequestID
	hash := sha256.Sum256([]byte(snapshot.ID + ":" + fingerprint + ":" + request.ClientRequestID))
	status := entity.ReportAnalysisStatusQueued
	errorCode := ""
	if snapshot.Report.Totals.TransactionCount == 0 {
		status = entity.ReportAnalysisStatusCompleted
		errorCode = "no_data"
	} else if len(facts.Candidates) == 0 {
		status = entity.ReportAnalysisStatusCompleted
		errorCode = "insufficient_evidence"
	}
	analysis := &entity.ReportAnalysis{ID: uuid.NewString(), OwnerID: principal.OwnerID, SnapshotID: snapshot.ID, JobKey: jobKey, ClientRequestID: request.ClientRequestID, PayloadHash: hex.EncodeToString(hash[:]), AnalysisFingerprint: fingerprint, Revision: snapshot.DataRevision, RuleVersion: entity.ReportAnalysisRuleVersion, PromptVersion: entity.ReportAnalysisPromptVersion, ProviderProfileVersion: "advisor-v1", Status: status, ErrorCode: errorCode, CreatedAt: now, UpdatedAt: now, RetainedUntil: now.Add(entity.ReportSnapshotRetention)}
	if status == entity.ReportAnalysisStatusCompleted {
		applyDeterministicAnalysis(analysis, snapshot.Report, errorCode)
	}
	created, replay, err := s.Analyses.CreateOrGet(ctx, analysis)
	if err != nil {
		return nil, false, mapReportAnalysisRepositoryError(err)
	}
	return created, replay, nil
}

func (s *ReportAnalysisService) Latest(ctx context.Context, principal Principal, snapshotID string) (*ReportAnalysisLatest, error) {
	if s == nil || s.Analyses == nil || s.Snapshots == nil {
		return nil, repository.ErrNotFound
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	snapshot, err := s.Snapshots.GetOwned(ctx, principal.OwnerID, strings.TrimSpace(snapshotID), now)
	if err != nil {
		return nil, err
	}
	completed, active, err := s.Analyses.LatestForScope(ctx, principal.OwnerID, snapshot.Scope, snapshot.Timezone, now)
	if err != nil {
		return nil, err
	}
	result := &ReportAnalysisLatest{Completed: completed, Active: active, Freshness: "none", StaleReasons: []string{}, CurrentSnapshotID: snapshot.ID}
	if completed == nil {
		return result, nil
	}
	facts, err := entity.BuildReportFactBundle(*snapshot)
	if err != nil {
		return nil, err
	}
	fingerprint, err := reportAnalysisFingerprint(*snapshot, facts)
	if err != nil {
		return nil, err
	}
	if completed.AnalysisFingerprint == fingerprint {
		result.Freshness = "current"
	} else {
		result.Freshness = "stale"
		result.StaleReasons = []string{"facts_changed"}
		if completed.RuleVersion != entity.ReportAnalysisRuleVersion || completed.PromptVersion != entity.ReportAnalysisPromptVersion {
			result.StaleReasons = []string{"rules_changed"}
		}
	}
	return result, nil
}

func reportAnalysisFingerprint(snapshot entity.ReportSnapshot, facts entity.ReportFactBundle) (string, error) {
	scope := snapshot.Report.Scope
	scope.Granularity = ""
	// Evaluation time and chart granularity do not change the underlying claims.
	evidence := facts
	evidence.SnapshotID, evidence.Revision, evidence.AsOf = "", 0, time.Time{}
	evidence.Scope = scope
	evidence.Candidates = append([]entity.ReportInsight(nil), facts.Candidates...)
	// Exact sub-day cutoffs remain on the persisted evidence. The fingerprint
	// compares calendar coverage and financial claims, not incidental read time.
	location, locationErr := time.LoadLocation(snapshot.Report.Timezone)
	if locationErr != nil {
		location = time.UTC
	}
	calendarEnd := func(value time.Time) time.Time {
		if value.IsZero() {
			return value
		}
		day := value.Add(-time.Nanosecond).In(location)
		return time.Date(day.Year(), day.Month(), day.Day()+1, 0, 0, 0, 0, location).UTC()
	}
	evidence.CurrentRange.EndExclusive = calendarEnd(evidence.CurrentRange.EndExclusive)
	for i := range evidence.Candidates {
		evidence.Candidates[i].AsOf = time.Time{}
		evidence.Candidates[i].Scope.Granularity = ""
		evidence.Candidates[i].EndExclusive = calendarEnd(evidence.Candidates[i].EndExclusive)
		evidence.Candidates[i].Drilldown.EndExclusive = calendarEnd(evidence.Candidates[i].Drilldown.EndExclusive)
	}
	payload := struct {
		Scope      entity.MonthlyReportScope `json:"scope"`
		Timezone   string                    `json:"timezone"`
		MonthStart time.Time                 `json:"month_start"`
		Comparison entity.ReportRange        `json:"comparison"`
		Facts      entity.ReportFactBundle   `json:"facts"`
		Rule       string                    `json:"rule"`
		Prompt     string                    `json:"prompt"`
		Profile    string                    `json:"profile"`
	}{scope, snapshot.Timezone, snapshot.Report.CurrentRange.From, snapshot.Report.ComparisonRange, evidence, entity.ReportAnalysisRuleVersion, entity.ReportAnalysisPromptVersion, "advisor-v1"}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}

func (s *ReportAnalysisService) Get(ctx context.Context, principal Principal, analysisID string) (*entity.ReportAnalysis, error) {
	if s == nil || s.Analyses == nil {
		return nil, repository.ErrNotFound
	}
	analysis, err := s.Analyses.GetOwned(ctx, principal.OwnerID, strings.TrimSpace(analysisID))
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	if !analysis.RetainedUntil.After(now) {
		return nil, repository.ErrSnapshotExpired
	}
	return analysis, nil
}

func (s *ReportAnalysisService) Cancel(ctx context.Context, principal Principal, analysisID string) (*entity.ReportAnalysis, error) {
	if s == nil || s.Analyses == nil {
		return nil, repository.ErrNotFound
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	return s.Analyses.Cancel(ctx, principal.OwnerID, strings.TrimSpace(analysisID), now)
}

type ReportAnalysisWorker struct {
	Analyses  repository.ReportAnalysisRepository
	Snapshots repository.ReportSnapshotRepository
	Provider  ReportAnalysisProvider
	Now       func() time.Time
}

func NewReportAnalysisWorker(analyses repository.ReportAnalysisRepository, snapshots repository.ReportSnapshotRepository, provider ReportAnalysisProvider) *ReportAnalysisWorker {
	return &ReportAnalysisWorker{Analyses: analyses, Snapshots: snapshots, Provider: provider, Now: time.Now}
}

func (w *ReportAnalysisWorker) RunPass(ctx context.Context, limit int) (int, error) {
	if w == nil || w.Analyses == nil || w.Snapshots == nil {
		return 0, nil
	}
	now := time.Now().UTC()
	if w.Now != nil {
		now = w.Now().UTC()
	}
	if err := w.Analyses.RecoverStale(ctx, now, limit); err != nil {
		return 0, err
	}
	processed := 0
	for processed < limit {
		now = workerNow(w)
		lease, err := w.Analyses.ClaimNext(ctx, now, entity.ReportAnalysisLease)
		if errors.Is(err, repository.ErrNotFound) {
			break
		}
		if err != nil {
			return processed, err
		}
		processed++
		if w.Provider == nil || !w.Provider.Available() {
			_ = w.Analyses.Fail(ctx, lease.ID, lease.LeaseToken, now, "provider_unavailable")
			continue
		}
		snapshot, err := w.Snapshots.GetRetainedOwned(ctx, lease.OwnerID, lease.SnapshotID, workerNow(w))
		if err != nil {
			_ = w.Analyses.Fail(ctx, lease.ID, lease.LeaseToken, now, "snapshot_unavailable")
			continue
		}
		facts, err := entity.BuildReportFactBundle(*snapshot)
		if err != nil {
			_ = w.Analyses.Fail(ctx, lease.ID, lease.LeaseToken, now, "facts_invalid")
			continue
		}
		providerResult, err := w.analyzeWithLease(ctx, lease, facts)
		if err != nil && !errors.Is(err, entity.ErrReportAnalysisOutputInvalid) {
			_ = w.Analyses.Fail(ctx, lease.ID, lease.LeaseToken, now, "provider_failed")
			continue
		}
		completed := *lease
		completed.ProviderCalled = true
		completed.Usage = providerResult.Usage
		if validationErr := entity.ValidateReportProviderResult(providerResult, facts); err != nil || validationErr != nil {
			applyDeterministicAnalysis(&completed, snapshot.Report, "provider_output_invalid")
			completed.ErrorCode = "provider_output_invalid"
		} else {
			applyProviderAnalysis(&completed, snapshot.Report, providerResult)
		}
		completed.Status = entity.ReportAnalysisStatusCompleted
		if err := w.Analyses.Complete(ctx, lease.ID, lease.LeaseToken, workerNow(w), &completed); err != nil && !errors.Is(err, repository.ErrAnalysisLeaseLost) {
			return processed, err
		}
	}
	return processed, nil
}

func workerNow(w *ReportAnalysisWorker) time.Time {
	if w != nil && w.Now != nil {
		return w.Now().UTC()
	}
	return time.Now().UTC()
}

func (w *ReportAnalysisWorker) analyzeWithLease(ctx context.Context, lease *entity.ReportAnalysis, facts entity.ReportFactBundle) (entity.ReportProviderResult, error) {
	type outcome struct {
		result entity.ReportProviderResult
		err    error
	}
	providerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	result := make(chan outcome, 1)
	go func() { value, err := w.Provider.Analyze(providerCtx, facts); result <- outcome{value, err} }()
	ticker := time.NewTicker(entity.ReportAnalysisLease / 3)
	defer ticker.Stop()
	for {
		select {
		case value := <-result:
			return value.result, value.err
		case <-ticker.C:
			if err := w.Analyses.RenewLease(ctx, lease.ID, lease.LeaseToken, workerNow(w), entity.ReportAnalysisLease); err != nil {
				cancel()
				return entity.ReportProviderResult{}, err
			}
		case <-ctx.Done():
			return entity.ReportProviderResult{}, ctx.Err()
		}
	}
}

func applyProviderAnalysis(analysis *entity.ReportAnalysis, report entity.MonthlyReport, result entity.ReportProviderResult) {
	analysis.Summary = reportAnalysisSummary(report)
	analysis.Observations, analysis.Suggestions, analysis.FactRefs = nil, nil, nil
	candidates := map[string]entity.ReportInsight{}
	for _, candidate := range entity.BuildReportAnalysisCandidates(report) {
		candidates[candidate.ID] = candidate
	}
	for _, observation := range result.Observations {
		candidate, ok := candidates[observation.FindingID]
		if !ok {
			continue
		}
		if observation.Intent == "inspect_evidence" {
			candidate.ActionHint = "Mở các giao dịch liên quan để kiểm tra số liệu trước khi quyết định điều chỉnh."
		}
		analysis.Observations = append(analysis.Observations, candidate)
		analysis.FactRefs = appendUnique(analysis.FactRefs, candidate.FactRefs...)
	}
	if len(analysis.Observations) == 0 && len(analysis.Suggestions) == 0 {
		applyDeterministicAnalysis(analysis, report, "provider_empty")
	}
}

func applyDeterministicAnalysis(analysis *entity.ReportAnalysis, report entity.MonthlyReport, reason string) {
	analysis.Summary = reportAnalysisSummary(report)
	analysis.ErrorCode = reason
	candidates := entity.BuildReportAnalysisCandidates(report)
	analysis.Observations = nil
	seenKinds, seenIDs := map[string]bool{}, map[string]bool{}
	for _, candidate := range candidates {
		if len(analysis.Observations) == 3 {
			break
		}
		if seenKinds[candidate.Kind] {
			continue
		}
		analysis.Observations = append(analysis.Observations, candidate)
		seenKinds[candidate.Kind], seenIDs[candidate.ID] = true, true
	}
	for _, candidate := range candidates {
		if len(analysis.Observations) == 3 {
			break
		}
		if !seenIDs[candidate.ID] {
			analysis.Observations = append(analysis.Observations, candidate)
			seenIDs[candidate.ID] = true
		}
	}
	analysis.Suggestions, analysis.FactRefs = nil, nil
	for _, insight := range analysis.Observations {
		analysis.FactRefs = appendUnique(analysis.FactRefs, insight.FactRefs...)
	}
}

func reportAnalysisSummary(report entity.MonthlyReport) string {
	if report.Totals.TransactionCount == 0 {
		return "Kỳ này chưa có giao dịch để phân tích."
	}
	period := entity.ReportPeriodLabel(report.CurrentRange.From, report.CurrentRange.EndExclusive, report.Timezone)
	if !report.Comparison.Comparable {
		return fmt.Sprintf("Dữ liệu %s · %d giao dịch. Chưa đủ dữ liệu tương ứng để kết luận xu hướng so với tháng trước.", period, report.Totals.TransactionCount)
	}
	return fmt.Sprintf("Dữ liệu %s · %d giao dịch. So sánh dùng %d ngày đầu của mỗi kỳ; tổng thu–chi dùng toàn bộ kỳ đang xem.", period, report.Totals.TransactionCount, report.Comparison.CommonDays)
}

func appendUnique(values []string, additions ...string) []string {
	seen := make(map[string]struct{}, len(values)+len(additions))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	for _, value := range additions {
		if value != "" {
			if _, ok := seen[value]; !ok {
				values = append(values, value)
				seen[value] = struct{}{}
			}
		}
	}
	return values
}

func mapReportAnalysisRepositoryError(err error) error {
	switch {
	case errors.Is(err, repository.ErrAnalysisActive):
		return ErrReportAnalysisBusy
	case errors.Is(err, repository.ErrAnalysisCooldown):
		return ErrReportAnalysisCooldown
	case errors.Is(err, repository.ErrAnalysisRequestConflict):
		return ErrReportAnalysisRequestConflict
	default:
		return err
	}
}
