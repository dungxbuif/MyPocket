package planning

import (
	"fmt"
	"math"
	"math/big"
	"time"
)

func ValidateCreateBudget(input CreateBudgetInput) (CreateBudgetInput, error) {
	input.Name = trimmed(input.Name)
	if input.Name == "" {
		return CreateBudgetInput{}, fmt.Errorf("%w: budget name is required", ErrValidation)
	}
	if !validBudgetPeriod(input.PeriodType) {
		return CreateBudgetInput{}, fmt.Errorf("%w: unsupported budget period", ErrValidation)
	}
	if input.AmountVND <= 0 {
		return CreateBudgetInput{}, fmt.Errorf("%w: budget amount must be positive", ErrValidation)
	}
	if input.PeriodType == BudgetCustom {
		if input.CustomStart == nil || input.CustomEnd == nil {
			return CreateBudgetInput{}, fmt.Errorf("%w: custom budget dates are required", ErrValidation)
		}
		if dateOnly(*input.CustomEnd).Before(dateOnly(*input.CustomStart)) {
			return CreateBudgetInput{}, fmt.Errorf("%w: custom budget end must be after start", ErrValidation)
		}
	} else if input.CustomStart != nil || input.CustomEnd != nil {
		return CreateBudgetInput{}, fmt.Errorf("%w: custom dates require custom period", ErrValidation)
	}
	return input, nil
}

func ValidateUpdateBudget(input UpdateBudgetInput) (UpdateBudgetInput, error) {
	if input.BaseVersion <= 0 {
		return UpdateBudgetInput{}, fmt.Errorf("%w: base version is required", ErrValidation)
	}
	created, err := ValidateCreateBudget(CreateBudgetInput{
		Name: input.Name, PeriodType: input.PeriodType, AmountVND: input.AmountVND,
		CategoryIDs: input.CategoryIDs, CustomStart: input.CustomStart, CustomEnd: input.CustomEnd,
	})
	if err != nil {
		return UpdateBudgetInput{}, err
	}
	return UpdateBudgetInput{BaseVersion: input.BaseVersion, Name: created.Name, PeriodType: created.PeriodType, AmountVND: created.AmountVND, CategoryIDs: created.CategoryIDs, CustomStart: created.CustomStart, CustomEnd: created.CustomEnd}, nil
}

func PeriodWindow(period BudgetPeriodType, customStart *time.Time, customEnd *time.Time, now time.Time) (time.Time, time.Time, error) {
	loc := mustHoChiMinh()
	local := now.In(loc)
	switch period {
	case BudgetWeekly:
		start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
		daysFromMonday := (int(start.Weekday()) + 6) % 7
		start = start.AddDate(0, 0, -daysFromMonday)
		return start, start.AddDate(0, 0, 6), nil
	case BudgetMonthly:
		start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 1, -1), nil
	case BudgetQuarterly:
		month := time.Month(((int(local.Month())-1)/3)*3 + 1)
		start := time.Date(local.Year(), month, 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 3, -1), nil
	case BudgetYearly:
		start := time.Date(local.Year(), 1, 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(1, 0, -1), nil
	case BudgetCustom:
		if customStart == nil || customEnd == nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: custom budget dates are required", ErrValidation)
		}
		return dateOnlyInLocation(*customStart, loc), dateOnlyInLocation(*customEnd, loc), nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("%w: unsupported budget period", ErrValidation)
	}
}

func percentSpent(spent int64, amount int64) int64 {
	if amount <= 0 || spent <= 0 {
		return 0
	}
	percent := new(big.Int).Mul(big.NewInt(spent), big.NewInt(100))
	percent.Quo(percent, big.NewInt(amount))
	if !percent.IsInt64() {
		return math.MaxInt64
	}
	return percent.Int64()
}

func validBudgetPeriod(value BudgetPeriodType) bool {
	switch value {
	case BudgetWeekly, BudgetMonthly, BudgetQuarterly, BudgetYearly, BudgetCustom:
		return true
	default:
		return false
	}
}

func dateOnly(value time.Time) time.Time {
	return dateOnlyInLocation(value, time.UTC)
}

func dateOnlyInLocation(value time.Time, loc *time.Location) time.Time {
	local := value.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
}

func mustHoChiMinh() *time.Location {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		return time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)
	}
	return loc
}
