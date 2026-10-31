package entity

import (
	"errors"
	"testing"
	"time"
)

func TestBuildReportFactBundlePreservesSnapshotIdentityAndCapsFacts(t *testing.T) {
	snapshot := ReportSnapshot{
		ID:           "snapshot-1",
		DataRevision: 7,
		AsOf:         time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
		Scope:        MonthlyReportScope{Month: "2026-10", Granularity: "day"},
		Report: MonthlyReport{
			Scope:        MonthlyReportScope{Month: "2026-10", Granularity: "day"},
			AsOf:         time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
			Totals:       ReportTotals{Income: 100, Expense: 50, Net: 50, TransactionCount: 2},
			Contributors: []ReportContributor{{FactRef: "category:coffee", CurrentAmount: 50, Delta: 10}},
		},
	}
	bundle, err := BuildReportFactBundle(snapshot)
	if err != nil {
		t.Fatalf("BuildReportFactBundle() error = %v", err)
	}
	if bundle.SnapshotID != snapshot.ID || bundle.Revision != snapshot.DataRevision || bundle.Scope != snapshot.Report.Scope {
		t.Fatalf("bundle identity = %+v", bundle)
	}
	if len(bundle.Facts) == 0 || bundle.Facts[0].ID != "totals.income" {
		t.Fatalf("facts = %+v", bundle.Facts)
	}

	tooMany := snapshot
	tooMany.Report.Contributors = make([]ReportContributor, 20)
	for index := range tooMany.Report.Contributors {
		tooMany.Report.Contributors[index] = ReportContributor{FactRef: "category:" + string(rune('a'+index)), CurrentAmount: 1, Delta: 1}
	}
	if _, err := BuildReportFactBundle(tooMany); !errors.Is(err, ErrReportAnalysisFactsInvalid) {
		t.Fatalf("too many facts error = %v, want ErrReportAnalysisFactsInvalid", err)
	}
}

func TestValidateReportProviderResultRejectsUnsupportedReferencesAndNumbers(t *testing.T) {
	facts := ReportFactBundle{Facts: []ReportFact{{ID: "totals.expense", SourceRef: "totals.expense"}}}
	for name, result := range map[string]ReportProviderResult{
		"unknown fact":  {Observations: []ReportProviderObservation{{FactRef: "totals.income", Text: "Theo dõi nhịp chi"}}},
		"numeric prose": {Observations: []ReportProviderObservation{{FactRef: "totals.expense", Text: "Chi tăng 12 phần trăm"}}},
		"too many observations": {Observations: []ReportProviderObservation{
			{FactRef: "totals.expense"}, {FactRef: "totals.expense"}, {FactRef: "totals.expense"}, {FactRef: "totals.expense"},
		}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateReportProviderResult(result, facts); !errors.Is(err, ErrReportAnalysisOutputInvalid) {
				t.Fatalf("ValidateReportProviderResult() error = %v, want ErrReportAnalysisOutputInvalid", err)
			}
		})
	}

	valid := ReportProviderResult{Observations: []ReportProviderObservation{{FactRef: "totals.expense", Text: "Theo dõi nhịp chi trong kỳ tiếp theo"}}}
	if err := ValidateReportProviderResult(valid, facts); err != nil {
		t.Fatalf("valid provider result rejected: %v", err)
	}
}
