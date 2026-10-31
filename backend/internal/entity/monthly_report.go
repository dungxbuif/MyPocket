package entity

import (
	"errors"
	"math/big"
	"time"
)

const (
	MonthlyReportRuleVersion = "monthly-report-v1"
	ReportInsightRuleVersion = "report-insight-v1"
	ReportSnapshotRetention  = 90 * 24 * time.Hour
	ReportAnalysisCooldown   = 60 * time.Second
)

var (
	ErrMonthlyReportScopeInvalid = errors.New("monthly report scope is invalid")
	ErrMonthlyReportFuture       = errors.New("monthly report cannot target a future month")
	ErrMonthlyReportOverflow     = errors.New("monthly report money overflow")
)

type MonthlyReportScope struct {
	Month       string `json:"month"`
	WalletID    string `json:"wallet_id,omitempty"`
	Granularity string `json:"granularity"`
}

type ReportRange struct {
	From         time.Time `json:"from"`
	EndExclusive time.Time `json:"end_exclusive"`
	Days         int       `json:"days"`
	IsPartial    bool      `json:"is_partial"`
}

type ReportTotals struct {
	Income           int64 `json:"income"`
	Expense          int64 `json:"expense"`
	Net              int64 `json:"net"`
	TransactionCount int64 `json:"transaction_count"`
	IncomeCount      int64 `json:"income_count"`
	ExpenseCount     int64 `json:"expense_count"`
}

type ReportComparison struct {
	Comparable       bool         `json:"comparable"`
	CommonDays       int          `json:"common_days"`
	Current          ReportTotals `json:"current"`
	Previous         ReportTotals `json:"previous"`
	IncomeDelta      int64        `json:"income_delta"`
	ExpenseDelta     int64        `json:"expense_delta"`
	NetDelta         int64        `json:"net_delta"`
	ExpenseChangeBPS *int64       `json:"expense_change_bps,omitempty"`
	IncomeChangeBPS  *int64       `json:"income_change_bps,omitempty"`
}

type ReportSeriesPoint struct {
	Label            string    `json:"label"`
	From             time.Time `json:"from"`
	EndExclusive     time.Time `json:"end_exclusive"`
	Income           int64     `json:"income"`
	Expense          int64     `json:"expense"`
	Net              int64     `json:"net"`
	TransactionCount int64     `json:"transaction_count"`
}

type ReportCategoryBreakdown struct {
	CategoryID       *string `json:"category_id,omitempty"`
	Name             string  `json:"name"`
	Amount           int64   `json:"amount"`
	TransactionCount int64   `json:"transaction_count"`
}

type ReportContributor struct {
	CategoryID     *string `json:"category_id,omitempty"`
	Name           string  `json:"name"`
	CurrentAmount  int64   `json:"current_amount"`
	PreviousAmount int64   `json:"previous_amount"`
	Delta          int64   `json:"delta"`
	CurrentCount   int64   `json:"current_count"`
	PreviousCount  int64   `json:"previous_count"`
	ChangeBPS      *int64  `json:"change_bps,omitempty"`
	FactRef        string  `json:"fact_ref"`
}

type ReportBudgetPreview struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	LimitAmount int64   `json:"limit_amount"`
	Spent       int64   `json:"spent"`
	Remaining   int64   `json:"remaining"`
	OverLimit   bool    `json:"over_limit"`
	WalletID    *string `json:"wallet_id,omitempty"`
	CategoryID  *string `json:"category_id,omitempty"`
	FactRef     string  `json:"fact_ref"`
}

type ReportJarPreview struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	AllocationMode       string `json:"allocation_mode"`
	AllocationAmount     *int64 `json:"allocation_amount,omitempty"`
	CalculatedAllocation *int64 `json:"calculated_allocation,omitempty"`
	Spent                int64  `json:"spent"`
	Active               bool   `json:"active"`
	Scope                string `json:"scope"`
	FactRef              string `json:"fact_ref"`
}

type ReportDrilldown struct {
	From         time.Time `json:"from"`
	EndExclusive time.Time `json:"end_exclusive"`
	WalletID     string    `json:"wallet_id,omitempty"`
	CategoryIDs  []string  `json:"category_ids,omitempty"`
	Type         string    `json:"type,omitempty"`
}

type ReportInsight struct {
	ID           string             `json:"id"`
	RuleVersion  string             `json:"rule_version"`
	Kind         string             `json:"kind"`
	Title        string             `json:"title"`
	Body         string             `json:"body"`
	FactRefs     []string           `json:"fact_refs"`
	Scope        MonthlyReportScope `json:"scope"`
	From         time.Time          `json:"from"`
	EndExclusive time.Time          `json:"end_exclusive"`
	AsOf         time.Time          `json:"as_of"`
	Drilldown    ReportDrilldown    `json:"drilldown"`
	Priority     int                `json:"-"`
}

type MonthlyReport struct {
	Scope           MonthlyReportScope        `json:"scope"`
	Timezone        string                    `json:"timezone"`
	Currency        string                    `json:"currency"`
	AsOf            time.Time                 `json:"as_of"`
	DataRevision    int64                     `json:"data_revision"`
	SnapshotID      string                    `json:"snapshot_id"`
	RuleVersion     string                    `json:"rule_version"`
	CurrentRange    ReportRange               `json:"current_range"`
	ComparisonRange ReportRange               `json:"comparison_range"`
	Totals          ReportTotals              `json:"totals"`
	Comparison      ReportComparison          `json:"comparison"`
	Series          []ReportSeriesPoint       `json:"series"`
	Categories      []ReportCategoryBreakdown `json:"categories"`
	Contributors    []ReportContributor       `json:"contributors"`
	Budgets         []ReportBudgetPreview     `json:"budgets"`
	Jars            []ReportJarPreview        `json:"jars"`
	QualityFlags    []string                  `json:"quality_flags"`
	Drilldown       ReportDrilldown           `json:"drilldown"`
	Insights        []ReportInsight           `json:"insights"`
}

func CheckedMoneyAdd(left, right int64) (int64, error) {
	return checkedMoneyOperation(left, right, false)
}

func CheckedMoneyDifference(left, right int64) (int64, error) {
	return checkedMoneyOperation(left, right, true)
}

func CheckedMoneyBPSChange(current, previous int64) (*int64, error) {
	if previous <= 0 {
		return nil, nil
	}
	numerator := new(big.Int).Mul(big.NewInt(current-previous), big.NewInt(10000))
	change := new(big.Int).Quo(numerator, big.NewInt(previous))
	if !change.IsInt64() {
		return nil, ErrMonthlyReportOverflow
	}
	value := change.Int64()
	return &value, nil
}

func CheckedMoneyPercentOfBPS(amount int64, bps int) (int64, error) {
	if amount < 0 || bps < 0 {
		return 0, ErrMonthlyReportOverflow
	}
	numerator := new(big.Int).Mul(big.NewInt(amount), big.NewInt(int64(bps)))
	numerator.Quo(numerator, big.NewInt(10000))
	if !numerator.IsInt64() || numerator.Int64() > MaxSafeMoneyAmount {
		return 0, ErrMonthlyReportOverflow
	}
	return numerator.Int64(), nil
}

func checkedMoneyOperation(left, right int64, subtract bool) (int64, error) {
	value := new(big.Int).SetInt64(left)
	operand := new(big.Int).SetInt64(right)
	if subtract {
		value.Sub(value, operand)
	} else {
		value.Add(value, operand)
	}
	if !value.IsInt64() || value.Int64() > MaxSafeMoneyAmount || value.Int64() < -MaxSafeMoneyAmount {
		return 0, ErrMonthlyReportOverflow
	}
	return value.Int64(), nil
}
