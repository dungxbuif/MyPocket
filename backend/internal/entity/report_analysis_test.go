package entity

import (
	"strings"
	"testing"
	"time"
)

func evidenceReport() MonthlyReport {
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	return MonthlyReport{Scope: MonthlyReportScope{Month: "2026-11"}, Timezone: "Asia/Ho_Chi_Minh", CurrentRange: ReportRange{From: from, EndExclusive: from.AddDate(0, 0, 2)}, Totals: ReportTotals{Income: 48000000, Expense: 12367000, Net: 35633000, TransactionCount: 17}, Comparison: ReportComparison{CommonDays: 1}, Budgets: []ReportBudgetPreview{{ID: "shopping", Name: "Mua sắm", LimitAmount: 1500000, Spent: 1850000, Remaining: -350000, OverLimit: true, FactRef: "budget:shopping", CategoryID: reportString("shopping-category")}}}
}

func TestAnalysisCandidatesKeepBudgetEvidenceAndRejectEarlyTrend(t *testing.T) {
	report := evidenceReport()
	for i := 0; i < 40; i++ {
		report.Contributors = append(report.Contributors, ReportContributor{FactRef: "category:x", Name: "Ăn uống", CurrentAmount: 2000000, PreviousAmount: 1000000, Delta: 1000000, CurrentCount: 10, PreviousCount: 10, ChangeBPS: reportBPS(10000)})
	}
	bundle, err := BuildReportFactBundle(ReportSnapshot{ID: "snapshot", DataRevision: 7, Report: report})
	if err != nil || bundle.SnapshotID != "snapshot" || bundle.Revision != 7 {
		t.Fatalf("identity/error: %+v %v", bundle, err)
	}
	if len(bundle.Candidates) != 1 {
		t.Fatalf("early trend or lost budget: %+v", bundle.Candidates)
	}
	finding := bundle.Candidates[0]
	if !strings.Contains(finding.Title, "350.000") || finding.Metrics[0].Amount != 1850000 || finding.Drilldown.TargetID != "budget:shopping" || finding.Drilldown.CategoryIDs[0] != "shopping-category" {
		t.Fatalf("wrong budget evidence: %+v", finding)
	}
	if bundle.Comparable || bundle.CommonDays != 1 {
		t.Fatal("comparison quality lost")
	}
}

func TestAnalysisSelectionOnlyAcceptsDistinctEligibleCandidates(t *testing.T) {
	bundle, _ := BuildReportFactBundle(ReportSnapshot{Report: evidenceReport()})
	valid := ReportProviderObservation{FindingID: bundle.Candidates[0].ID, Intent: "review_plan"}
	if err := ValidateReportProviderResult(ReportProviderResult{Observations: []ReportProviderObservation{valid}}, bundle); err != nil {
		t.Fatal(err)
	}
	for name, rows := range map[string][]ReportProviderObservation{
		"unknown":            {{FindingID: "category-change:fake", Intent: "inspect_evidence"}},
		"duplicate":          {valid, valid},
		"unsupported intent": {{FindingID: valid.FindingID, Intent: "buy_gold"}},
		"invented claims":    {{FindingID: valid.FindingID, Intent: "review_plan", Text: "Bạn đã chi 999 đồng"}},
		"legacy prose":       {{FactRef: "totals.net", Direction: "down", Text: "Số dư giảm"}},
	} {
		t.Run(name, func(t *testing.T) {
			if ValidateReportProviderResult(ReportProviderResult{Observations: rows}, bundle) == nil {
				t.Fatal("invalid output accepted")
			}
		})
	}
}

func TestAnalysisConcentrationIsNotTrendAndDoesNotDuplicateBudget(t *testing.T) {
	report := evidenceReport()
	report.Categories = []ReportCategoryBreakdown{{CategoryID: reportString("shopping-category"), Name: "Mua sắm", TargetID: "category:shopping", Amount: 1850000, ShareBPS: 3000, TransactionCount: 1}, {Name: "Nhà ở", TargetID: "category:rent", Amount: 6000000, ShareBPS: 4800, TransactionCount: 1}}
	items := BuildReportAnalysisCandidates(report)
	if len(items) != 2 || items[1].Kind != "category_concentration" || !strings.Contains(items[1].Body, "48,0%") {
		t.Fatalf("candidates %+v", items)
	}
	if abs(0) != 0 {
		t.Fatal("zero must remain zero when ranking impacts")
	}
}

func TestAnalysisCandidateDiversitySurvivesInsiderDisplayLimit(t *testing.T) {
	report := evidenceReport()
	for _, id := range []string{"travel", "housing"} {
		report.Budgets = append(report.Budgets, ReportBudgetPreview{ID: id, Name: id, LimitAmount: 1000000, Spent: 2000000, OverLimit: true, FactRef: "budget:" + id, CategoryID: reportString(id)})
	}
	report.Comparison = ReportComparison{Comparable: true, CommonDays: 10}
	report.ComparisonRange = ReportRange{From: report.CurrentRange.From, EndExclusive: report.CurrentRange.From.AddDate(0, 0, 1)}
	report.Contributors = []ReportContributor{{CategoryID: reportString("food"), Name: "Ăn uống", FactRef: "category:food", CurrentAmount: 2000000, PreviousAmount: 1000000, Delta: 1000000, CurrentCount: 10, PreviousCount: 10, ChangeBPS: reportBPS(10000)}}
	if len(BuildReportInsights(report)) != 3 {
		t.Fatal("Insider display cap changed")
	}
	found := false
	for _, item := range BuildReportAnalysisCandidates(report) {
		if item.ID == "category-change:category:food" {
			found = true
			if !item.EndExclusive.Equal(report.ComparisonRange.EndExclusive) {
				t.Fatal("finding metadata exceeds matched evidence window")
			}
		}
	}
	if !found {
		t.Fatal("budget warnings starved eligible category evidence")
	}
}
