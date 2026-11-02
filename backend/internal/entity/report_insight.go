package entity

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ReportInsightKindBudget   = "budget_over_limit"
	ReportInsightKindCategory = "category_change"
	ReportInsightKindNeutral  = "neutral"
)

func BuildReportInsights(report MonthlyReport) []ReportInsight {
	items := buildReportInsightCandidates(report)
	if len(items) > 3 {
		return items[:3]
	}
	return items
}

func buildReportInsightCandidates(report MonthlyReport) []ReportInsight {
	items := make([]ReportInsight, 0, 4)
	seen := map[string]struct{}{}
	for _, budget := range report.Budgets {
		if !budget.OverLimit {
			continue
		}
		id := "budget-over:" + budget.ID
		from, end := budget.From, budget.EndExclusive
		if from.IsZero() {
			from = report.CurrentRange.From
		}
		if end.IsZero() {
			end = report.CurrentRange.EndExclusive
		}
		target := budget.Drilldown
		if target.From.IsZero() {
			target = ReportDrilldown{TargetID: "budget:" + budget.ID, From: from, EndExclusive: end, WalletID: report.Scope.WalletID, CategoryIDs: stringSlice(budget.CategoryID), CategoryBucket: "root", Type: TransactionTypeExpense}
		}
		items = append(items, ReportInsight{
			ID: id, RuleVersion: ReportInsightRuleVersion, Kind: ReportInsightKindBudget,
			Title: budget.Name + " vượt " + formatReportVND(budget.Spent-budget.LimitAmount), Body: fmt.Sprintf("Đã chi %s trên hạn mức %s (gồm phần chuyển tiếp nếu có).", formatReportVND(budget.Spent), formatReportVND(budget.LimitAmount)),
			FactRefs: []string{budget.FactRef}, Scope: report.Scope, From: from,
			EndExclusive: end, AsOf: report.AsOf, Priority: 0,
			Drilldown:   target,
			Metrics:     []ReportMetric{{"Đã chi", budget.Spent}, {"Hạn mức", budget.LimitAmount}, {"Vượt", budget.Spent - budget.LimitAmount}},
			PeriodLabel: ReportPeriodLabel(from, end, report.Timezone),
			ActionHint:  "Xem các khoản chi trong ngân sách này để quyết định khoản nào có thể điều chỉnh.",
			Impact:      budget.Spent - budget.LimitAmount,
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
		if !report.Comparison.Comparable || report.Comparison.CommonDays < 7 || contributor.CurrentCount < 5 || contributor.PreviousCount < 5 || abs(contributor.Delta) < 100000 {
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
			Title: contributor.Name + " " + direction, Body: fmt.Sprintf("%s %s %s so với cùng %d ngày của kỳ trước.", contributor.Name, direction, formatReportVND(abs(contributor.Delta)), report.Comparison.CommonDays),
			FactRefs: []string{contributor.FactRef}, Scope: report.Scope, From: report.ComparisonRange.From,
			EndExclusive: report.ComparisonRange.EndExclusive, AsOf: report.AsOf, Priority: 1,
			Drilldown: contributor.Drilldown, Impact: abs(contributor.Delta),
			Metrics:     []ReportMetric{{"Kỳ này (cùng số ngày)", contributor.CurrentAmount}, {"Kỳ trước (cùng số ngày)", contributor.PreviousAmount}},
			PeriodLabel: fmt.Sprintf("%d ngày đầu kỳ · %s", report.Comparison.CommonDays, report.Scope.Month),
			ActionHint:  "Đối chiếu các khoản chi trong nhóm này trước khi điều chỉnh kế hoạch tháng tới.",
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
		if items[i].Impact != items[j].Impact {
			return items[i].Impact > items[j].Impact
		}
		return items[i].ID < items[j].ID
	})
	return items
}

func formatReportVND(value int64) string {
	digits := strconv.FormatInt(value, 10)
	groups := make([]string, 0, (len(digits)+2)/3)
	for len(digits) > 3 {
		groups = append([]string{digits[len(digits)-3:]}, groups...)
		digits = digits[:len(digits)-3]
	}
	groups = append([]string{digits}, groups...)
	return strings.Join(groups, ".") + " đ"
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
	if value == -1<<63 {
		return MaxSafeMoneyAmount
	}
	if value < 0 {
		return -value
	}
	return value
}

func ReportPeriodLabel(from, end time.Time, timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	if from.IsZero() || end.IsZero() {
		return ""
	}
	return from.In(loc).Format("02/01/2006") + " – " + end.Add(-time.Nanosecond).In(loc).Format("02/01/2006")
}

// BuildReportAnalysisCandidates preserves distinct evidence groups. Additional
// candidates explain current spending without pretending a short month is a trend.
func BuildReportAnalysisCandidates(report MonthlyReport) []ReportInsight {
	// The Insider display limit is not a candidate-generation limit. Reserve
	// capacity per evidence kind so budget warnings cannot starve comparisons.
	items := make([]ReportInsight, 0, 9)
	counts := map[string]int{}
	for _, candidate := range buildReportInsightCandidates(report) {
		if counts[candidate.Kind] >= 3 {
			continue
		}
		items = append(items, candidate)
		counts[candidate.Kind]++
	}
	seenCategories := map[string]bool{}
	for _, item := range items {
		for _, id := range item.Drilldown.CategoryIDs {
			seenCategories[id] = true
		}
	}
	if report.Totals.TransactionCount == 0 {
		return items
	}
	if report.Totals.Net < 0 {
		report.Drilldown.TargetID = "report:all"
		items = append(items, ReportInsight{ID: "cashflow:deficit", RuleVersion: ReportInsightRuleVersion, Kind: "cashflow_deficit", Title: "Chi nhiều hơn thu " + formatReportVND(-report.Totals.Net), Body: "Đây là chênh lệch thu–chi được ghi nhận trong kỳ, không phải số dư các ví.", FactRefs: []string{"totals.income", "totals.expense", "totals.net"}, Scope: report.Scope, From: report.CurrentRange.From, EndExclusive: report.CurrentRange.EndExclusive, AsOf: report.AsOf, Drilldown: report.Drilldown, Metrics: []ReportMetric{{"Thu nhập", report.Totals.Income}, {"Chi tiêu", report.Totals.Expense}}, PeriodLabel: ReportPeriodLabel(report.CurrentRange.From, report.CurrentRange.EndExclusive, report.Timezone), ActionHint: "Kiểm tra các khoản thu đã ghi đủ chưa và những khoản chi lớn trong kỳ.", Priority: 0, Impact: -report.Totals.Net})
	}
	categories := append([]ReportCategoryBreakdown(nil), report.Categories...)
	sort.SliceStable(categories, func(i, j int) bool {
		if categories[i].Amount == categories[j].Amount {
			return categories[i].TargetID < categories[j].TargetID
		}
		return categories[i].Amount > categories[j].Amount
	})
	added := 0
	for _, category := range categories {
		if added >= 2 {
			break
		}
		if category.Amount < 100000 || category.ShareBPS < 2000 || category.TransactionCount == 0 {
			continue
		}
		if category.CategoryID != nil && seenCategories[*category.CategoryID] {
			continue
		}
		items = append(items, ReportInsight{ID: "concentration:" + category.TargetID, RuleVersion: ReportInsightRuleVersion, Kind: "category_concentration", Title: category.Name + " chiếm phần lớn chi tiêu", Body: fmt.Sprintf("%s: %s qua %d giao dịch, chiếm %d,%d%% chi tiêu đã ghi nhận. Tỷ trọng này chưa nói lên khoản chi có cần thiết hay không.", category.Name, formatReportVND(category.Amount), category.TransactionCount, category.ShareBPS/100, (category.ShareBPS%100)/10), FactRefs: []string{category.TargetID}, Scope: report.Scope, From: report.CurrentRange.From, EndExclusive: report.CurrentRange.EndExclusive, AsOf: report.AsOf, Drilldown: category.Drilldown, Metrics: []ReportMetric{{"Nhóm này", category.Amount}, {"Tổng chi", report.Totals.Expense}}, PeriodLabel: ReportPeriodLabel(report.CurrentRange.From, report.CurrentRange.EndExclusive, report.Timezone), ActionHint: "Xem các giao dịch trong nhóm để phân biệt khoản chi một lần với chi tiêu thường xuyên.", Priority: 2, Impact: category.Amount})
		added++
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority < items[j].Priority
		}
		if items[i].Impact != items[j].Impact {
			return items[i].Impact > items[j].Impact
		}
		return items[i].ID < items[j].ID
	})
	return items
}
