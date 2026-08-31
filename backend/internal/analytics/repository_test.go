package analytics

import (
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
