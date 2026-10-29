package repository

import (
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

func TestAdvanceRecurringMonthlyPreservesAnchorAndClampsShortMonth(t *testing.T) {
	location := time.FixedZone("ICT", 7*60*60)
	schedule := &entity.RecurringSchedule{Frequency: "monthly", Interval: 1, AnchorDay: 31, AnchorMonth: 1}
	due := time.Date(2026, time.January, 31, 9, 30, 0, 0, location).UTC()
	next := advanceRecurring(schedule, due, location).In(location)
	if next.Format("2006-01-02 15:04") != "2026-02-28 09:30" {
		t.Fatalf("next monthly occurrence = %s", next)
	}
	following := advanceRecurring(schedule, next.UTC(), location).In(location)
	if following.Format("2006-01-02 15:04") != "2026-03-31 09:30" {
		t.Fatalf("anchor day was not restored after clamp: %s", following)
	}
}

func TestAdvanceRecurringYearlyPreservesMonthAndDay(t *testing.T) {
	location := time.FixedZone("ICT", 7*60*60)
	schedule := &entity.RecurringSchedule{Frequency: "yearly", Interval: 1, AnchorDay: 29, AnchorMonth: 2}
	due := time.Date(2024, time.February, 29, 8, 0, 0, 0, location).UTC()
	next := advanceRecurring(schedule, due, location).In(location)
	if next.Format("2006-01-02 15:04") != "2025-02-28 08:00" {
		t.Fatalf("next yearly occurrence = %s", next)
	}
}
