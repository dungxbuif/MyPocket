package entity

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	ReportAnalysisStatusQueued    = "queued"
	ReportAnalysisStatusRunning   = "running"
	ReportAnalysisStatusCompleted = "completed"
	ReportAnalysisStatusFailed    = "failed"
	ReportAnalysisStatusCancelled = "cancelled"
	ReportAnalysisRuleVersion     = "report-analysis-v3"
	ReportAnalysisPromptVersion   = "report-analysis-prompt-v3"
	ReportAnalysisLease           = 2 * time.Minute
	ReportAnalysisDeadline        = 5 * time.Minute
)

var (
	ErrReportAnalysisOutputInvalid = errors.New("report analysis provider output is invalid")
	ErrReportAnalysisFactsInvalid  = errors.New("report analysis facts are invalid")
)

type ReportFact struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Value     int64  `json:"value,omitempty"`
	Text      string `json:"text,omitempty"`
	SourceRef string `json:"source_ref"`
}

type ReportFactBundle struct {
	SnapshotID      string             `json:"snapshot_id"`
	Revision        int64              `json:"revision"`
	Scope           MonthlyReportScope `json:"scope"`
	AsOf            time.Time          `json:"as_of"`
	Facts           []ReportFact       `json:"facts"`
	Candidates      []ReportInsight    `json:"candidates"`
	Comparable      bool               `json:"comparable"`
	CommonDays      int                `json:"common_days"`
	CurrentRange    ReportRange        `json:"current_range"`
	ComparisonRange ReportRange        `json:"comparison_range"`
	QualityFlags    []string           `json:"quality_flags"`
}

type ReportProviderObservation struct {
	FindingID string `json:"finding_id"`
	Intent    string `json:"intent"`
	FactRef   string `json:"fact_ref"`
	Direction string `json:"direction"`
	Text      string `json:"text"`
}

type ReportProviderSuggestion struct {
	FactRef string `json:"fact_ref"`
	Text    string `json:"text"`
}

type ReportProviderResult struct {
	Observations []ReportProviderObservation `json:"observations"`
	Suggestions  []ReportProviderSuggestion  `json:"suggestions"`
	Usage        map[string]any              `json:"usage,omitempty"`
}

type ReportAnalysis struct {
	ID                     string          `json:"id" gorm:"primaryKey"`
	OwnerID                string          `json:"-" gorm:"index;not null"`
	SnapshotID             string          `json:"snapshot_id" gorm:"index;not null"`
	JobKey                 string          `json:"-"`
	ClientRequestID        string          `json:"-"`
	PayloadHash            string          `json:"-"`
	AnalysisFingerprint    string          `json:"analysis_fingerprint"`
	Revision               int64           `json:"revision"`
	RuleVersion            string          `json:"rule_version"`
	PromptVersion          string          `json:"prompt_version"`
	ProviderProfileVersion string          `json:"provider_profile_version"`
	Status                 string          `json:"status"`
	ErrorCode              string          `json:"error_code,omitempty"`
	ProviderCalled         bool            `json:"provider_called"`
	Summary                string          `json:"summary,omitempty"`
	Observations           []ReportInsight `json:"observations,omitempty" gorm:"-"`
	Suggestions            []ReportInsight `json:"suggestions,omitempty" gorm:"-"`
	FactRefs               []string        `json:"fact_refs,omitempty" gorm:"-"`
	Usage                  map[string]any  `json:"usage,omitempty" gorm:"-"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	StartedAt              *time.Time      `json:"started_at,omitempty"`
	FinishedAt             *time.Time      `json:"finished_at,omitempty"`
	RetainedUntil          time.Time       `json:"retained_until" gorm:"-"`
	LeaseUntil             *time.Time      `json:"-"`
	LeaseToken             string          `json:"-"`
	DeadlineAt             *time.Time      `json:"-"`
	ObservationPayload     json.RawMessage `json:"-" gorm:"column:observations"`
	SuggestionPayload      json.RawMessage `json:"-" gorm:"column:suggestions"`
	FactRefsPayload        json.RawMessage `json:"-" gorm:"column:fact_refs"`
	UsagePayload           json.RawMessage `json:"-" gorm:"column:usage"`
}

func (ReportAnalysis) TableName() string { return "report_analyses" }

func BuildReportFactBundle(snapshot ReportSnapshot) (ReportFactBundle, error) {
	report := snapshot.Report
	facts := make([]ReportFact, 0, 30)
	appendFact := func(fact ReportFact) bool {
		if len(facts) >= 30 {
			return false
		}
		facts = append(facts, fact)
		return true
	}
	for _, item := range []struct {
		id    string
		value int64
	}{
		{"totals.income", report.Totals.Income}, {"totals.expense", report.Totals.Expense},
		{"totals.net", report.Totals.Net}, {"totals.transaction_count", report.Totals.TransactionCount},
	} {
		appendFact(ReportFact{ID: item.id, Kind: "money_or_count", Value: item.value, SourceRef: item.id})
	}
	return ReportFactBundle{SnapshotID: snapshot.ID, Revision: snapshot.DataRevision, Scope: report.Scope, AsOf: report.AsOf, Facts: facts, Candidates: BuildReportAnalysisCandidates(report), Comparable: report.Comparison.Comparable, CommonDays: report.Comparison.CommonDays, CurrentRange: report.CurrentRange, ComparisonRange: report.ComparisonRange, QualityFlags: report.QualityFlags}, nil
}

func ValidateReportProviderResult(result ReportProviderResult, facts ReportFactBundle) error {
	known := make(map[string]struct{}, len(facts.Candidates))
	for _, candidate := range facts.Candidates {
		known[candidate.ID] = struct{}{}
	}
	if len(result.Observations) > 3 || len(result.Suggestions) != 0 {
		return ErrReportAnalysisOutputInvalid
	}
	seen := map[string]bool{}
	for _, item := range result.Observations {
		if _, ok := known[item.FindingID]; !ok || seen[item.FindingID] {
			return ErrReportAnalysisOutputInvalid
		}
		seen[item.FindingID] = true
		if item.Intent != "inspect_evidence" && item.Intent != "review_plan" {
			return ErrReportAnalysisOutputInvalid
		}
		// Free-form claims are not part of the selection contract.
		if item.Text != "" || item.Direction != "" || item.FactRef != "" {
			return ErrReportAnalysisOutputInvalid
		}
	}
	return nil
}
