package entity

import "testing"

func TestSpendingIncomeRatioBPSUsesZeroBaselineAsUnknown(t *testing.T) {
	if ratio := SpendingIncomeRatioBPS(1200, 0); ratio != nil {
		t.Fatalf("SpendingIncomeRatioBPS() = %v, want nil", *ratio)
	}
	ratio := SpendingIncomeRatioBPS(2500, 10000)
	if ratio == nil || *ratio != 2500 {
		t.Fatalf("SpendingIncomeRatioBPS() = %v, want 2500", ratio)
	}
}

func TestAverageDailySpendRoundsDeterministically(t *testing.T) {
	if got := AverageDailySpend(1000, 3); got != 333 {
		t.Fatalf("AverageDailySpend() = %d, want 333", got)
	}
	if got := AverageDailySpend(1000, 0); got != 0 {
		t.Fatalf("AverageDailySpend() with zero days = %d, want 0", got)
	}
}
