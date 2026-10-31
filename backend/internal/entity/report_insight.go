package entity

import (
	"fmt"
	"sort"
)

const (
	ReportInsightKindBudget   = "budget_over_limit"
	ReportInsightKindCategory = "category_change"
	ReportInsightKindNeutral  = "neutral"
)

func BuildReportInsights(report MonthlyReport) []ReportInsight {
	items := make([]ReportInsight, 0, 4)
	seen := map[string]struct{}{}
	for _, budget := range report.Budgets {
		if !budget.OverLimit {
			continue
		}
		id := "budget-over:" + budget.ID
		items = append(items, ReportInsight{
			ID: id, RuleVersion: ReportInsightRuleVersion, Kind: ReportInsightKindBudget,
			Title: "Ngân sách đã vượt giới hạn", Body: fmt.Sprintf("%s đã vượt ngân sách ghi nhận trong kỳ.", budget.Name),
			FactRefs: []string{budget.FactRef}, Scope: report.Scope, From: report.CurrentRange.From,
			EndExclusive: report.CurrentRange.EndExclusive, AsOf: report.AsOf, Priority: 0,
			Drilldown: ReportDrilldown{From: report.CurrentRange.From, EndExclusive: report.CurrentRange.EndExclusive, WalletID: report.Scope.WalletID, CategoryIDs: stringSlice(budget.CategoryID), Type: TransactionTypeExpense},
		})
		seen[budget.CategoryIDValue()] = struct{}{}
	}
	contributors := append([]ReportContributor(nil), report.Contributors...)
	sort.SliceStable(contributors, func(i, j int) bool {
		left, right := abs(contributors[i].Delta), abs(contributors[j].Delta)
		if left != right {
			return left > right
		}
		return contributors[i].FactRef < contributors[j].FactRef
	})
	for _, contributor := range contributors {
		if report.Comparison.CommonDays < 7 || contributor.CurrentCount < 5 || contributor.PreviousCount < 5 || abs(contributor.Delta) < 100000 {
			continue
		}
		if contributor.ChangeBPS == nil || abs(*contributor.ChangeBPS) < 1500 {
			continue
		}
		id := "category-change:" + contributor.FactRef
		if _, ok := seen[contributor.CategoryIDValue()]; ok {
			continue
		}
		direction := "tăng"
		if contributor.Delta < 0 {
			direction = "giảm"
		}
		items = append(items, ReportInsight{
			ID: id, RuleVersion: ReportInsightRuleVersion, Kind: ReportInsightKindCategory,
			Title: "Danh mục thay đổi đáng chú ý", Body: fmt.Sprintf("%s %s so với kỳ so sánh.", contributor.Name, direction),
			FactRefs: []string{contributor.FactRef}, Scope: report.Scope, From: report.ComparisonRange.From,
			EndExclusive: report.CurrentRange.EndExclusive, AsOf: report.AsOf, Priority: 1,
			Drilldown: ReportDrilldown{From: report.CurrentRange.From, EndExclusive: report.CurrentRange.EndExclusive, WalletID: report.Scope.WalletID, CategoryIDs: stringSlice(contributor.CategoryID), Type: TransactionTypeExpense},
		})
	}
	if len(items) == 0 && report.Totals.TransactionCount == 0 {
		items = append(items, ReportInsight{
			ID: "neutral:no-data", RuleVersion: ReportInsightRuleVersion, Kind: ReportInsightKindNeutral,
			Title: "Chưa có dữ liệu đủ để nhận xét", Body: "Kỳ này chưa có giao dịch được ghi nhận trong báo cáo.",
			FactRefs: []string{"totals.transaction_count"}, Scope: report.Scope, From: report.CurrentRange.From,
			EndExclusive: report.CurrentRange.EndExclusive, AsOf: report.AsOf, Priority: 2,
			Drilldown: report.Drilldown,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority < items[j].Priority
		}
		return items[i].ID < items[j].ID
	})
	if len(items) > 3 {
		items = items[:3]
	}
	return items
}

func (c ReportContributor) CategoryIDValue() string {
	if c.CategoryID == nil {
		return ""
	}
	return *c.CategoryID
}

func (b ReportBudgetPreview) CategoryIDValue() string {
	if b.CategoryID == nil {
		return ""
	}
	return *b.CategoryID
}

func stringSlice(value *string) []string {
	if value == nil || *value == "" {
		return nil
	}
	return []string{*value}
}

func abs(value int64) int64 {
	if value == -value {
		return MaxSafeMoneyAmount
	}
	if value < 0 {
		return -value
	}
	return value
}
