package httpapi

import (
	"testing"
	"time"

	"mypocket/internal/analytics"
)

func TestComparisonPeriodsReturnsSixChronologicalEqualRanges(t *testing.T) {
	loc := time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)
	current := analytics.Filter{
		From:     time.Date(2026, 8, 1, 0, 0, 0, 0, loc),
		To:       time.Date(2026, 8, 31, 23, 59, 59, 999999999, loc),
		WalletID: "wallet-1",
	}

	periods := comparisonPeriods(current)

	if len(periods) != 6 {
		t.Fatalf("period count=%d, want 6", len(periods))
	}
	if !periods[5].From.Equal(current.From) || !periods[5].To.Equal(current.To) {
		t.Fatalf("last period is not current: %#v", periods[5])
	}
	span := current.To.Sub(current.From)
	for i, period := range periods {
		if period.WalletID != current.WalletID || period.To.Sub(period.From) != span {
			t.Fatalf("period %d lost scope or duration: %#v", i, period)
		}
		if i > 0 && !period.From.Equal(periods[i-1].To.Add(time.Nanosecond)) {
			t.Fatalf("period %d is not contiguous: previous=%#v current=%#v", i, periods[i-1], period)
		}
	}
}
