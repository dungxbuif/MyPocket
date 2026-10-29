package entity

import (
	"math/big"
	"sort"
	"time"
)

type BudgetForecast struct {
	Status             string `json:"status"`
	Method             string `json:"method"`
	ElapsedDays        int    `json:"elapsed_days"`
	TotalDays          int    `json:"total_days"`
	ProjectedSpent     *int64 `json:"projected_spent"`
	ProjectedRemaining *int64 `json:"projected_remaining"`
	OverBudget         *bool  `json:"over_budget"`
}

type BudgetDailySpend struct {
	Date  string `json:"date"`
	Spent int64  `json:"spent"`
}

type BudgetReport struct {
	BudgetID             string             `json:"budget_id"`
	Timezone             string             `json:"timezone"`
	AsOf                 time.Time          `json:"as_of"`
	StartDate            CalendarDate       `json:"start_date"`
	EndDate              CalendarDate       `json:"end_date"`
	BaseLimitAmount      int64              `json:"base_limit_amount"`
	CarryoverAmount      int64              `json:"carryover_amount"`
	EffectiveLimitAmount int64              `json:"effective_limit_amount"`
	Spent                int64              `json:"spent"`
	Remaining            int64              `json:"remaining"`
	Forecast             BudgetForecast     `json:"forecast"`
	Daily                []BudgetDailySpend `json:"daily"`
}

// MatchBudgetTransactions is shared by the report and its drilldown. Calendar
// labels use UTC solely for arithmetic; inclusion instants use the account zone.
func MatchBudgetTransactions(b Budget, rows []Transaction, categories []Category, now time.Time, location *time.Location) []Transaction {
	if location == nil {
		location = time.UTC
	}
	parents := map[string]string{}
	for _, c := range categories {
		if c.ParentID != nil {
			parents[c.ID] = *c.ParentID
		}
	}
	start, end := budgetReportDates(b)
	out := []Transaction{}
	for _, row := range rows {
		if row.OwnerID != b.OwnerID || row.Type != TransactionTypeExpense || !row.IncludedInReports || row.TransferID != nil || row.CreditPaymentID != nil || row.OccurredAt.After(now) {
			continue
		}
		label := reportCalendarDay(row.OccurredAt, location)
		if label.Before(start.Time) || label.After(end.Time) {
			continue
		}
		if b.WalletID != nil && row.WalletID != *b.WalletID {
			continue
		}
		if b.CategoryID != nil {
			if row.CategoryID == nil {
				continue
			}
			id := *row.CategoryID
			seen := map[string]bool{}
			for id != "" && id != *b.CategoryID && !seen[id] {
				seen[id] = true
				id = parents[id]
			}
			if id != *b.CategoryID {
				continue
			}
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].OccurredAt.After(out[j].OccurredAt)
	})
	return out
}

func CalculateBudgetReport(b Budget, rows []Transaction, categories []Category, now time.Time, location *time.Location) (BudgetReport, error) {
	if location == nil {
		return BudgetReport{}, ErrFinanceRangeInvalid
	}
	start, end := budgetReportDates(b)
	if start.IsZero() || end.IsZero() || end.Before(start.Time) {
		return BudgetReport{}, ErrFinanceRangeInvalid
	}
	effective := b.EffectiveLimitAmount
	if effective == 0 {
		effective = b.LimitAmount
	}
	if effective < 0 || effective > MaxSafeMoneyAmount || b.CarryoverAmount < 0 {
		return BudgetReport{}, ErrFinanceAmountInvalid
	}
	result := BudgetReport{BudgetID: b.ID, Timezone: location.String(), AsOf: now.UTC(), StartDate: start, EndDate: end, BaseLimitAmount: b.LimitAmount, CarryoverAmount: b.CarryoverAmount, EffectiveLimitAmount: effective, Daily: []BudgetDailySpend{}}
	total := int(end.Sub(start.Time).Hours()/24) + 1
	today := reportCalendarDay(now, location)
	elapsed := int(today.Sub(start.Time).Hours()/24) + 1
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > total {
		elapsed = total
	}
	result.Forecast = BudgetForecast{Status: "estimated", Method: "linear_daily_v1", ElapsedDays: elapsed, TotalDays: total}
	byDay := map[string]int64{}
	for _, row := range MatchBudgetTransactions(b, rows, categories, now, location) {
		if row.Amount < 0 || row.Amount > MaxSafeMoneyAmount-result.Spent {
			return BudgetReport{}, ErrFinanceAmountInvalid
		}
		result.Spent += row.Amount
		byDay[reportCalendarDay(row.OccurredAt, location).Format("2006-01-02")] += row.Amount
	}
	result.Remaining = effective - result.Spent
	for i := 0; i < elapsed; i++ {
		label := start.AddDate(0, 0, i).Format("2006-01-02")
		result.Daily = append(result.Daily, BudgetDailySpend{Date: label, Spent: byDay[label]})
	}
	if today.Before(start.Time) {
		result.Forecast.Status = "not_started"
		return result, nil
	}
	projected := result.Spent
	if today.After(end.Time) {
		result.Forecast.Status = "actual"
	} else {
		// Ceiling with big.Int avoids intermediate overflow for large valid money.
		numerator := new(big.Int).Mul(big.NewInt(result.Spent), big.NewInt(int64(total)))
		numerator.Add(numerator, big.NewInt(int64(elapsed-1)))
		numerator.Quo(numerator, big.NewInt(int64(elapsed)))
		if !numerator.IsInt64() || numerator.Int64() > MaxSafeMoneyAmount {
			return BudgetReport{}, ErrFinanceAmountInvalid
		}
		projected = numerator.Int64()
	}
	projectedRemaining := effective - projected
	over := projected > effective
	result.Forecast.ProjectedSpent = &projected
	result.Forecast.ProjectedRemaining = &projectedRemaining
	result.Forecast.OverBudget = &over
	return result, nil
}

func budgetReportDates(b Budget) (CalendarDate, CalendarDate) {
	start, end := b.StartDate, b.EndDate
	if start.IsZero() && !b.StartAt.IsZero() {
		start = CalendarDate{Time: reportCalendarDay(b.StartAt, time.UTC)}
	}
	if end.IsZero() && !b.EndAt.IsZero() {
		end = CalendarDate{Time: reportCalendarDay(b.EndAt.Add(-time.Microsecond), time.UTC)}
	}
	return start, end
}

func reportCalendarDay(at time.Time, location *time.Location) time.Time {
	local := at.In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}
