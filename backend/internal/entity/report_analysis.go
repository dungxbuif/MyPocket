package entity

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	ReportAnalysisStatusQueued    = "queued"
	ReportAnalysisStatusRunning   = "running"
	ReportAnalysisStatusCompleted = "completed"
	ReportAnalysisStatusFailed    = "failed"
	ReportAnalysisStatusCancelled = "cancelled"
	ReportAnalysisRuleVersion     = "report-analysis-v2"
	ReportAnalysisPromptVersion   = "report-analysis-prompt-v2"
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
	SnapshotID string             `json:"snapshot_id"`
	Revision   int64              `json:"revision"`
	Scope      MonthlyReportScope `json:"scope"`
	AsOf       time.Time          `json:"as_of"`
	Facts      []ReportFact       `json:"facts"`
}

type ReportProviderObservation struct {
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
		{"comparison.income_delta", report.Comparison.IncomeDelta}, {"comparison.expense_delta", report.Comparison.ExpenseDelta},
		{"comparison.net_delta", report.Comparison.NetDelta},
	} {
		appendFact(ReportFact{ID: item.id, Kind: "money_or_count", Value: item.value, SourceRef: item.id})
	}
	for _, contributor := range report.Contributors {
		if len(facts)+2 > 30 {
			break
		}
		appendFact(ReportFact{ID: contributor.FactRef + ".current", Kind: "money", Value: contributor.CurrentAmount, SourceRef: contributor.FactRef})
		appendFact(ReportFact{ID: contributor.FactRef + ".delta", Kind: "money", Value: contributor.Delta, SourceRef: contributor.FactRef})
	}
	for _, budget := range report.Budgets {
		if len(facts)+2 > 30 {
			break
		}
		appendFact(ReportFact{ID: budget.FactRef + ".spent", Kind: "money", Value: budget.Spent, SourceRef: budget.FactRef})
		appendFact(ReportFact{ID: budget.FactRef + ".remaining", Kind: "money", Value: budget.Remaining, SourceRef: budget.FactRef})
	}
	for _, jar := range report.Jars {
		if !appendFact(ReportFact{ID: jar.FactRef + ".spent", Kind: "money", Value: jar.Spent, SourceRef: jar.FactRef}) {
			break
		}
	}
	return ReportFactBundle{SnapshotID: snapshot.ID, Revision: snapshot.DataRevision, Scope: report.Scope, AsOf: report.AsOf, Facts: facts}, nil
}

func ValidateReportProviderResult(result ReportProviderResult, facts ReportFactBundle) error {
	known := make(map[string]struct{}, len(facts.Facts))
	for _, fact := range facts.Facts {
		known[fact.ID] = struct{}{}
		known[fact.SourceRef] = struct{}{}
	}
	if len(result.Observations) > 3 || len(result.Suggestions) > 3 {
		return ErrReportAnalysisOutputInvalid
	}
	for _, item := range append(append([]ReportProviderObservation(nil), result.Observations...), convertSuggestions(result.Suggestions)...) {
		if _, ok := known[item.FactRef]; !ok {
			return ErrReportAnalysisOutputInvalid
		}
		if len([]rune(strings.TrimSpace(item.Text))) > 500 || containsDigit(item.Text) {
			return ErrReportAnalysisOutputInvalid
		}
	}
	return nil
}

func convertSuggestions(items []ReportProviderSuggestion) []ReportProviderObservation {
	result := make([]ReportProviderObservation, 0, len(items))
	for _, item := range items {
		result = append(result, ReportProviderObservation{FactRef: item.FactRef, Text: item.Text})
	}
	return result
}

func containsDigit(value string) bool {
	for _, r := range value {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}
