package entity

import (
	"testing"
	"time"
)

func TestNormalizeDateRangeUsesNextLocalMidnightAcrossDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	rangeValue, err := NormalizeDateRange(DateRange{From: "2026-03-08", To: "2026-03-08"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	if got := rangeValue.EndExclusive.Sub(rangeValue.StartAt); got != 23*time.Hour {
		t.Fatalf("DST day must be 23h, got %s", got)
	}
	if rangeValue.StartAt.Location() != time.UTC || rangeValue.EndExclusive.Location() != time.UTC {
		t.Fatalf("normalized instants must be UTC: %+v", rangeValue)
	}
}

func TestNormalizeDateRangeRejectsInvalidAndOverlongRanges(t *testing.T) {
	loc := time.FixedZone("ICT", 7*60*60)
	for _, input := range []DateRange{
		{From: "2026-02-30", To: "2026-03-01"},
		{From: "2026-03-01", To: "2026-02-28"},
		{From: "2025-01-01", To: "2026-02-01"},
	} {
		if _, err := NormalizeDateRange(input, loc); err == nil {
			t.Fatalf("expected invalid range error for %+v", input)
		}
	}
}

func TestValidateMoneyUsesJavaScriptSafeIntegerBoundary(t *testing.T) {
	if err := ValidateMoneyAmount(9007199254740991); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMoneyAmount(9007199254740992); err == nil {
		t.Fatal("amount above safe integer must be rejected")
	}
	if err := ValidateMoneyAmount(-1); err == nil {
		t.Fatal("negative amount must be rejected")
	}
}

func TestPercentBPSReturnsNilForZeroBaseline(t *testing.T) {
	percent, err := PercentBPS(120, 0)
	if err != nil {
		t.Fatal(err)
	}
	if percent != nil {
		t.Fatalf("zero baseline must be null, got %v", *percent)
	}
	percent, err = PercentBPS(150, 100)
	if err != nil || percent == nil || *percent != 5000 {
		t.Fatalf("expected 50.00%% as 5000 bps, got %v, %v", percent, err)
	}
}

func TestNormalizeFinanceFilterAppliesBoundedDefaultsAndRejectsUnknownType(t *testing.T) {
	loc := time.FixedZone("ICT", 7*60*60)
	filter, err := NormalizeFinanceFilter(FinanceFilter{
		Range:           DateRange{From: "2026-09-01", To: "2026-09-30"},
		WalletIDs:       []string{"w1", "w1", "w2"},
		CategoryIDs:     []string{"c1", "c1"},
		IncludeChildren: true,
		NoteContains:    "  Grab  ",
	}, loc)
	if err != nil {
		t.Fatal(err)
	}
	if filter.Type != FinanceTypeAll || filter.ReportScope != FinanceReportAll || filter.Limit != 20 {
		t.Fatalf("unexpected defaults: %+v", filter)
	}
	if len(filter.WalletIDs) != 2 || len(filter.CategoryIDs) != 1 || filter.NoteContains != "Grab" {
		t.Fatalf("normalization did not deduplicate/trim: %+v", filter)
	}
	for _, bad := range []string{"transfer", "unknown"} {
		input := FinanceFilter{Range: DateRange{From: "2026-09-01", To: "2026-09-30"}, Type: bad}
		if _, err := NormalizeFinanceFilter(input, loc); err == nil {
			t.Fatalf("expected invalid type for %q", bad)
		}
	}
}
