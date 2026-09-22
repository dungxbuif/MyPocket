package entity

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"
)

const MaxSafeMoneyAmount int64 = 9007199254740991

var (
	ErrFinanceDateInvalid   = errors.New("finance date is invalid")
	ErrFinanceRangeInvalid  = errors.New("finance date range is invalid")
	ErrFinanceRangeTooLarge = errors.New("finance date range exceeds twelve months")
	ErrFinanceAmountInvalid = errors.New("finance amount is invalid")
)

const (
	FinanceTypeAll     = "all"
	FinanceTypeIncome  = "income"
	FinanceTypeExpense = "expense"

	FinanceReportAll      = "all"
	FinanceReportIncluded = "included"
	FinanceReportExcluded = "excluded"
)

type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type NormalizedDateRange struct {
	From         string    `json:"from"`
	To           string    `json:"to"`
	StartAt      time.Time `json:"start_at"`
	EndExclusive time.Time `json:"end_exclusive"`
}

type FinanceFilter struct {
	Range           DateRange `json:"range"`
	WalletIDs       []string  `json:"wallet_ids"`
	CategoryIDs     []string  `json:"category_ids"`
	IncludeChildren bool      `json:"include_children"`
	Type            string    `json:"type"`
	ReportScope     string    `json:"report_scope"`
	NoteContains    string    `json:"note_contains"`
	MinAmount       *int64    `json:"min_amount,omitempty"`
	MaxAmount       *int64    `json:"max_amount,omitempty"`
	Limit           int       `json:"limit,omitempty"`
}

type NormalizedFinanceFilter struct {
	FinanceFilter
	Dates NormalizedDateRange `json:"dates"`
}

type NormalizedQuery struct {
	Key          string                  `json:"key"`
	Kind         string                  `json:"kind"`
	Filter       NormalizedFinanceFilter `json:"filter"`
	CompareRange *DateRange              `json:"compare_range,omitempty"`
	ObjectIDs    []string                `json:"object_ids,omitempty"`
	Month        string                  `json:"month,omitempty"`
	ActiveOn     string                  `json:"active_on,omitempty"`
	Limit        int                     `json:"limit,omitempty"`
	Cursor       string                  `json:"cursor,omitempty"`
}

type SourceScope struct {
	Ref          string        `json:"ref"`
	Timezone     string        `json:"timezone"`
	StartAt      *time.Time    `json:"start_at,omitempty"`
	EndExclusive *time.Time    `json:"end_exclusive,omitempty"`
	Filter       FinanceFilter `json:"filter"`
	AsOf         time.Time     `json:"as_of"`
}

type Fact struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	Integer   *int64  `json:"integer,omitempty"`
	Text      *string `json:"text,omitempty"`
	Currency  *string `json:"currency,omitempty"`
	SourceRef string  `json:"source_ref"`
}

type FinanceResult struct {
	QueryKey     string          `json:"query_key"`
	Status       string          `json:"status"`
	Facts        []Fact          `json:"facts"`
	ViewKind     string          `json:"view_kind"`
	View         json.RawMessage `json:"view"`
	Source       SourceScope     `json:"source"`
	WarningCodes []string        `json:"warning_codes,omitempty"`
}

type FactBundle struct {
	ID      string          `json:"id"`
	AsOf    time.Time       `json:"as_of"`
	Results []FinanceResult `json:"results"`
}

func NormalizeFinanceFilter(input FinanceFilter, location *time.Location) (NormalizedFinanceFilter, error) {
	dates, err := NormalizeDateRange(input.Range, location)
	if err != nil {
		return NormalizedFinanceFilter{}, err
	}
	if input.Type == "" {
		input.Type = FinanceTypeAll
	}
	if input.ReportScope == "" {
		input.ReportScope = FinanceReportAll
	}
	if input.Type != FinanceTypeAll && input.Type != FinanceTypeIncome && input.Type != FinanceTypeExpense {
		return NormalizedFinanceFilter{}, ErrFinanceRangeInvalid
	}
	if input.ReportScope != FinanceReportAll && input.ReportScope != FinanceReportIncluded && input.ReportScope != FinanceReportExcluded {
		return NormalizedFinanceFilter{}, ErrFinanceRangeInvalid
	}
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.Limit < 1 || input.Limit > 50 {
		return NormalizedFinanceFilter{}, ErrFinanceRangeInvalid
	}
	input.WalletIDs = dedupeBoundedIDs(input.WalletIDs)
	input.CategoryIDs = dedupeBoundedIDs(input.CategoryIDs)
	input.NoteContains = strings.TrimSpace(input.NoteContains)
	if len(input.NoteContains) > 200 {
		return NormalizedFinanceFilter{}, ErrFinanceRangeInvalid
	}
	if input.MinAmount != nil {
		if err := ValidateMoneyAmount(*input.MinAmount); err != nil {
			return NormalizedFinanceFilter{}, err
		}
	}
	if input.MaxAmount != nil {
		if err := ValidateMoneyAmount(*input.MaxAmount); err != nil {
			return NormalizedFinanceFilter{}, err
		}
	}
	if input.MinAmount != nil && input.MaxAmount != nil && *input.MinAmount > *input.MaxAmount {
		return NormalizedFinanceFilter{}, ErrFinanceRangeInvalid
	}
	return NormalizedFinanceFilter{FinanceFilter: input, Dates: dates}, nil
}

func dedupeBoundedIDs(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, value := range input {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if len(value) > 128 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
		if len(out) == 100 {
			break
		}
	}
	return out
}

// NormalizeDateRange converts inclusive account-local calendar labels to UTC
// instants. End is computed as the next local midnight, which preserves 23/25
// hour days around DST transitions.
func NormalizeDateRange(input DateRange, location *time.Location) (NormalizedDateRange, error) {
	if location == nil {
		return NormalizedDateRange{}, ErrFinanceRangeInvalid
	}
	from, err := parseCalendarDate(input.From)
	if err != nil {
		return NormalizedDateRange{}, err
	}
	to, err := parseCalendarDate(input.To)
	if err != nil {
		return NormalizedDateRange{}, err
	}
	fromLocal := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, location)
	toLocal := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, location)
	if toLocal.Before(fromLocal) {
		return NormalizedDateRange{}, ErrFinanceRangeInvalid
	}
	if toLocal.After(fromLocal.AddDate(1, 0, 0)) {
		return NormalizedDateRange{}, ErrFinanceRangeTooLarge
	}
	return NormalizedDateRange{
		From:         input.From,
		To:           input.To,
		StartAt:      fromLocal.UTC(),
		EndExclusive: toLocal.AddDate(0, 0, 1).UTC(),
	}, nil
}

func parseCalendarDate(value string) (time.Time, error) {
	if strings.TrimSpace(value) != value || len(value) != len("2006-01-02") {
		return time.Time{}, ErrFinanceDateInvalid
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, ErrFinanceDateInvalid
	}
	return parsed, nil
}

func ValidateMoneyAmount(amount int64) error {
	if amount < 0 || amount > MaxSafeMoneyAmount {
		return ErrFinanceAmountInvalid
	}
	return nil
}

// PercentBPS returns percentage change in signed basis points. A non-positive
// baseline is intentionally represented as nil so callers never render
// Infinity or an invented percentage.
func PercentBPS(current, baseline int64) (*int64, error) {
	if current < 0 || baseline < 0 {
		return nil, ErrFinanceAmountInvalid
	}
	if baseline == 0 {
		return nil, nil
	}
	numerator := new(big.Int).Mul(big.NewInt(current-baseline), big.NewInt(10000))
	denominator := big.NewInt(baseline)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	// Round half away from zero to make display stable across clients.
	if remainder.Sign() != 0 {
		twice := new(big.Int).Abs(remainder)
		twice.Mul(twice, big.NewInt(2))
		if twice.Cmp(denominator) >= 0 {
			if numerator.Sign() < 0 {
				quotient.Sub(quotient, big.NewInt(1))
			} else {
				quotient.Add(quotient, big.NewInt(1))
			}
		}
	}
	if !quotient.IsInt64() {
		return nil, ErrFinanceAmountInvalid
	}
	value := quotient.Int64()
	return &value, nil
}
