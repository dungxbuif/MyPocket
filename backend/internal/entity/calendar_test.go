package entity

import (
	"testing"
	"time"
)

func TestMonthRangeUsesAccountTimezoneAcrossDST(t *testing.T) {
	month, err := ParseMonth("2026-03")
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	start, next, err := MonthRangeUTC(month, loc)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("month start = %s, want %s", start, want)
	}
	if want := time.Date(2026, 4, 1, 7, 0, 0, 0, time.UTC); !next.Equal(want) {
		t.Fatalf("next month = %s, want %s", next, want)
	}
}

func TestMonthCompletionUsesAccountLocalMonth(t *testing.T) {
	month, err := ParseMonth("2026-09")
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("Pacific/Kiritimati")
	if err != nil {
		t.Fatal(err)
	}
	if IsMonthComplete(month, time.Date(2026, 9, 30, 9, 59, 0, 0, time.UTC), loc) {
		t.Fatal("month closed before the account-local month changed")
	}
	if !IsMonthComplete(month, time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC), loc) {
		t.Fatal("month should be complete at the account-local October boundary")
	}
}

func TestParseMonthRejectsInvalidLabels(t *testing.T) {
	for _, value := range []string{"", "2026-13", "26-09", "2026-9", "2026-09-01"} {
		if _, err := ParseMonth(value); err == nil {
			t.Errorf("ParseMonth(%q) unexpectedly succeeded", value)
		}
	}
}
