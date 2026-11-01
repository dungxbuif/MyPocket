package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	}
	analysis := &entity.ReportAnalysis{ID: uuid.NewString(), OwnerID: principal.OwnerID, SnapshotID: snapshot.ID, JobKey: jobKey, ClientRequestID: request.ClientRequestID, PayloadHash: hex.EncodeToString(hash[:]), AnalysisFingerprint: fingerprint, Revision: snapshot.DataRevision, RuleVersion: entity.ReportAnalysisRuleVersion, PromptVersion: entity.ReportAnalysisPromptVersion, ProviderProfileVersion: "advisor-v1", Status: status, ErrorCode: errorCode, CreatedAt: now, UpdatedAt: now, RetainedUntil: now.Add(entity.ReportSnapshotRetention)}
	if status == entity.ReportAnalysisStatusCompleted {
		applyDeterministicAnalysis(analysis, snapshot.Report, "no_data")
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
	payload := struct {
		Scope      entity.MonthlyReportScope `json:"scope"`
		Timezone   string                    `json:"timezone"`
		MonthStart time.Time                 `json:"month_start"`
		Comparison entity.ReportRange        `json:"comparison"`
		Facts      []entity.ReportFact       `json:"facts"`
		Rule       string                    `json:"rule"`
		Prompt     string                    `json:"prompt"`
		Profile    string                    `json:"profile"`
	}{scope, snapshot.Timezone, snapshot.Report.CurrentRange.From, snapshot.Report.ComparisonRange, facts.Facts, entity.ReportAnalysisRuleVersion, entity.ReportAnalysisPromptVersion, "advisor-v1"}
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
		if err != nil {
			_ = w.Analyses.Fail(ctx, lease.ID, lease.LeaseToken, now, "provider_failed")
			continue
		}
		completed := *lease
		completed.ProviderCalled = true
		completed.Usage = providerResult.Usage
		if err := entity.ValidateReportProviderResult(providerResult, facts); err != nil {
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
	analysis.Summary = "Phân tích dựa trên dữ liệu đã ghi nhận trong báo cáo."
	analysis.FactRefs = make([]string, 0, len(result.Observations)+len(result.Suggestions))
	for _, observation := range result.Observations {
		text := insightTextForFact(report, observation.FactRef, observation.Direction, observation.Text)
		analysis.Observations = append(analysis.Observations, entity.ReportInsight{ID: "analysis-observation:" + observation.FactRef, RuleVersion: entity.ReportAnalysisRuleVersion, Kind: "observation", Title: "Quan sát từ dữ liệu", Body: text, FactRefs: []string{observation.FactRef}, Scope: report.Scope, From: report.CurrentRange.From, EndExclusive: report.CurrentRange.EndExclusive, AsOf: report.AsOf, Drilldown: report.Drilldown})
		analysis.FactRefs = appendUnique(analysis.FactRefs, observation.FactRef)
	}
	for _, suggestion := range result.Suggestions {
		text := strings.TrimSpace(suggestion.Text)
		if text == "" {
			text = "Theo dõi dữ liệu này trong kỳ tiếp theo."
		}
		analysis.Suggestions = append(analysis.Suggestions, entity.ReportInsight{ID: "analysis-suggestion:" + suggestion.FactRef, RuleVersion: entity.ReportAnalysisRuleVersion, Kind: "suggestion", Title: "Gợi ý theo dõi", Body: text, FactRefs: []string{suggestion.FactRef}, Scope: report.Scope, From: report.CurrentRange.From, EndExclusive: report.CurrentRange.EndExclusive, AsOf: report.AsOf, Drilldown: report.Drilldown})
		analysis.FactRefs = appendUnique(analysis.FactRefs, suggestion.FactRef)
	}
	if len(analysis.Observations) == 0 && len(analysis.Suggestions) == 0 {
		applyDeterministicAnalysis(analysis, report, "provider_empty")
	}
}

func applyDeterministicAnalysis(analysis *entity.ReportAnalysis, report entity.MonthlyReport, reason string) {
	analysis.Summary = "Phân tích định lượng vẫn khả dụng từ báo cáo đã ghi nhận."
	analysis.ErrorCode = reason
	analysis.Observations = append([]entity.ReportInsight(nil), report.Insights...)
	if len(analysis.Observations) > 3 {
		analysis.Observations = analysis.Observations[:3]
	}
	for _, insight := range analysis.Observations {
		analysis.FactRefs = appendUnique(analysis.FactRefs, insight.FactRefs...)
	}
}

func insightTextForFact(report entity.MonthlyReport, factRef, direction, providerText string) string {
	for _, insight := range report.Insights {
		for _, ref := range insight.FactRefs {
			if ref == factRef {
				return insight.Body
			}
		}
	}
	if strings.TrimSpace(providerText) != "" {
		return strings.TrimSpace(providerText)
	}
	if strings.TrimSpace(direction) == "" {
		return "Có thay đổi được ghi nhận trong dữ liệu báo cáo."
	}
	return "Dữ liệu báo cáo cho thấy một thay đổi " + strings.TrimSpace(direction) + "."
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
