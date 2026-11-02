package entity

import (
	"errors"
	"testing"
)

func reportString(value string) *string { return &value }

func reportBPS(value int64) *int64 { return &value }

func TestReportInsightsThresholdsAndBudgetPriority(t *testing.T) {
	report := MonthlyReport{
		Scope:        MonthlyReportScope{Month: "2026-10", Granularity: "day"},
		Totals:       ReportTotals{TransactionCount: 12},
		CurrentRange: ReportRange{Days: 31},
		Comparison:   ReportComparison{Comparable: true, CommonDays: 10},
		Budgets: []ReportBudgetPreview{
			{ID: "budget-1", Name: "Ăn uống", OverLimit: true, CategoryID: reportString("cat-1"), FactRef: "budget:budget-1"},
			{ID: "budget-2", Name: "Mua sắm", OverLimit: true, CategoryID: reportString("cat-2"), FactRef: "budget:budget-2"},
		},
		Contributors: []ReportContributor{
			{CategoryID: reportString("cat-1"), Name: "Ăn uống", Delta: 900000, CurrentCount: 5, PreviousCount: 5, ChangeBPS: reportBPS(2000), FactRef: "category:cat-1"},
			{CategoryID: reportString("cat-3"), Name: "Di chuyển", Delta: -500000, CurrentCount: 5, PreviousCount: 5, ChangeBPS: reportBPS(-1600), FactRef: "category:cat-3"},
			{CategoryID: reportString("cat-4"), Name: "Giải trí", Delta: 400000, CurrentCount: 5, PreviousCount: 5, ChangeBPS: reportBPS(1700), FactRef: "category:cat-4"},
			{CategoryID: reportString("cat-5"), Name: "Sức khỏe", Delta: 300000, CurrentCount: 5, PreviousCount: 5, ChangeBPS: reportBPS(1600), FactRef: "category:cat-5"},
		},
	}

	insights := BuildReportInsights(report)
	if len(insights) != 3 {
		t.Fatalf("expected max three grounded insights, got %d: %+v", len(insights), insights)
	}
	if insights[0].Kind != ReportInsightKindBudget || insights[1].Kind != ReportInsightKindBudget {
		t.Fatalf("budget warnings must be first: %+v", insights)
	}
	for _, insight := range insights {
		if insight.ID == "category-change:category:cat-1" {
			t.Fatal("category already represented by an over-limit budget must be deduplicated")
		}
	}
}

func TestReportInsightsDoNotInventComparisonFromZeroBaseline(t *testing.T) {
	report := MonthlyReport{
		Scope:      MonthlyReportScope{Month: "2026-10", Granularity: "day"},
		Totals:     ReportTotals{TransactionCount: 1},
		Comparison: ReportComparison{CommonDays: 31},
		Contributors: []ReportContributor{{
			CategoryID: reportString("cat-1"), Name: "Danh mục mới", Delta: 900000,
			CurrentCount: 5, PreviousCount: 5, ChangeBPS: nil, FactRef: "category:cat-1",
		}},
	}

	if insights := BuildReportInsights(report); len(insights) != 0 {
		t.Fatalf("zero baseline must not create a comparative insight: %+v", insights)
	}
}

func TestCheckedMoneyRejectsOverflowAndKeepsZeroBaselineExplicit(t *testing.T) {
	if _, err := CheckedMoneyAdd(MaxSafeMoneyAmount, 1); !errors.Is(err, ErrMonthlyReportOverflow) {
		t.Fatalf("expected checked addition overflow, got %v", err)
	}
	if _, err := CheckedMoneyDifference(-MaxSafeMoneyAmount, 1); !errors.Is(err, ErrMonthlyReportOverflow) {
		t.Fatalf("expected checked subtraction overflow, got %v", err)
	}
	change, err := CheckedMoneyBPSChange(100, 0)
	if err != nil || change != nil {
		t.Fatalf("zero baseline must be unavailable, got %v %v", change, err)
	}
}
