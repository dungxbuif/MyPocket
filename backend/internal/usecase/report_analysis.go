package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

type ReportAnalysisService struct {
	Analyses  repository.ReportAnalysisRepository
	Snapshots repository.ReportSnapshotRepository
	Provider  ReportAnalysisProvider
	Now       func() time.Time
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
	jobKey := fmt.Sprintf("%s:%d:%s:%s:%s", snapshot.ID, snapshot.DataRevision, entity.ReportAnalysisRuleVersion, entity.ReportAnalysisPromptVersion, "advisor-v1")
	hash := sha256.Sum256([]byte(jobKey + ":" + request.ClientRequestID))
	status := entity.ReportAnalysisStatusQueued
	errorCode := ""
	if snapshot.Report.Totals.TransactionCount == 0 {
		status = entity.ReportAnalysisStatusCompleted
		errorCode = "no_data"
	} else if s.Provider == nil || !s.Provider.Available() {
		return nil, false, ErrReportAnalysisProviderUnavailable
	}
	analysis := &entity.ReportAnalysis{ID: uuid.NewString(), OwnerID: principal.OwnerID, SnapshotID: snapshot.ID, JobKey: jobKey, ClientRequestID: request.ClientRequestID, PayloadHash: hex.EncodeToString(hash[:]), Revision: snapshot.DataRevision, RuleVersion: entity.ReportAnalysisRuleVersion, PromptVersion: entity.ReportAnalysisPromptVersion, ProviderProfileVersion: "advisor-v1", Status: status, ErrorCode: errorCode, CreatedAt: now, UpdatedAt: now}
	if status == entity.ReportAnalysisStatusCompleted {
		applyDeterministicAnalysis(analysis, snapshot.Report, "no_data")
	}
	created, replay, err := s.Analyses.CreateOrGet(ctx, analysis)
	if err != nil {
		return nil, false, mapReportAnalysisRepositoryError(err)
	}
	return created, replay, nil
}

func (s *ReportAnalysisService) Get(ctx context.Context, principal Principal, analysisID string) (*entity.ReportAnalysis, error) {
	if s == nil || s.Analyses == nil {
		return nil, repository.ErrNotFound
	}
	return s.Analyses.GetOwned(ctx, principal.OwnerID, strings.TrimSpace(analysisID))
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
		snapshot, err := w.Snapshots.GetOwned(ctx, lease.OwnerID, lease.SnapshotID, now)
		if err != nil {
			_ = w.Analyses.Fail(ctx, lease.ID, lease.LeaseToken, now, "snapshot_unavailable")
			continue
		}
		facts, err := entity.BuildReportFactBundle(*snapshot)
		if err != nil {
			_ = w.Analyses.Fail(ctx, lease.ID, lease.LeaseToken, now, "facts_invalid")
			continue
		}
		providerResult, err := w.Provider.Analyze(ctx, facts)
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
		if err := w.Analyses.Complete(ctx, lease.ID, lease.LeaseToken, now, &completed); err != nil && !errors.Is(err, repository.ErrAnalysisLeaseLost) {
			return processed, err
		}
	}
	return processed, nil
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
