package analytics

import (
	"math"
	"testing"
	"time"
)

func TestNormalizeFilterUsesHoChiMinhCalendarDefaults(t *testing.T) {
	filter := NormalizeFilter(time.Time{}, time.Time{})
	if filter.From.Location().String() != "Asia/Ho_Chi_Minh" || filter.To.Location().String() != "Asia/Ho_Chi_Minh" {
		t.Fatalf("unexpected timezone: %s %s", filter.From.Location(), filter.To.Location())
	}
	if filter.From.Day() != 1 || filter.To.Day() < 28 {
		t.Fatalf("unexpected month boundaries: %s %s", filter.From, filter.To)
	}
}

func TestRoundDecimalMoneyCheckedRejectsOverflow(t *testing.T) {
	if _, err := roundDecimalMoneyChecked("9223372036854775807", 2); err == nil {
		t.Fatal("expected portfolio market value overflow to be rejected")
	}
	got, err := roundDecimalMoneyChecked("1.5", 100)
	if err != nil || got != 150 {
		t.Fatalf("expected rounded market value 150, got %d, err %v", got, err)
	}
}

func TestAddMoneyCheckedRejectsOverflow(t *testing.T) {
	if _, err := addMoneyChecked(math.MaxInt64, 1); err == nil {
		t.Fatal("expected aggregate overflow to be rejected")
	}
	if got, err := addMoneyChecked(40, 2); err != nil || got != 42 {
		t.Fatalf("expected checked total 42, got %d, err %v", got, err)
	}
}

func TestElapsedPeriodDaysUsesTodayForActiveMonth(t *testing.T) {
	loc := time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)
	filter := Filter{
		From: time.Date(2026, 8, 1, 0, 0, 0, 0, loc),
		To:   time.Date(2026, 8, 31, 23, 59, 59, 0, loc),
	}
	if got := elapsedPeriodDays(filter, time.Date(2026, 8, 10, 12, 0, 0, 0, loc)); got != 10 {
		t.Fatalf("expected 10 elapsed days, got %d", got)
	}
	if got := elapsedPeriodDays(filter, filter.To); got != 31 {
		t.Fatalf("expected 31 completed days, got %d", got)
	}
}
