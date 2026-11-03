package repository

import (
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

func TestBudgetSelectionLimitDoesNotCountCarryoverTwice(t *testing.T) {
	series := "weekly"
	day := func(d int) entity.CalendarDate {
		return entity.CalendarDate{Time: time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC)}
	}
	week1 := entity.Budget{SeriesID: &series, StartDate: day(1), EndDate: day(7), BaseLimitAmount: 1_000_000, EffectiveLimitAmount: 1_000_000}
	week2 := entity.Budget{SeriesID: &series, StartDate: day(8), EndDate: day(14), BaseLimitAmount: 1_000_000, CarryoverAmount: 600_000, EffectiveLimitAmount: 1_600_000}

	both, err := budgetSelectionLimit([]entity.Budget{week1, week2})
	if err != nil || both != 2_000_000 {
		t.Fatalf("both periods selected: total %d err %v, want 2,000,000", both, err)
	}
	onlySecond, err := budgetSelectionLimit([]entity.Budget{week2})
	if err != nil || onlySecond != 1_600_000 {
		t.Fatalf("carryover from an unselected period must count: total %d err %v, want 1,600,000", onlySecond, err)
	}
}
