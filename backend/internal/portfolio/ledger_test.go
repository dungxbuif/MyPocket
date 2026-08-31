package portfolio

import (
	"errors"
	"testing"
)

func TestReplayLedgerMovingAverageBuySellFeesAndRounding(t *testing.T) {
	results, err := replayLedger([]ledgerTrade{
		{Side: TradeBuy, Quantity: "2", UnitPriceVND: 70_000_000, FeeVND: 20_000},
		{Side: TradeBuy, Quantity: "1.5", UnitPriceVND: 72_000_000, FeeVND: 10_000},
		{Side: TradeSell, Quantity: "1.2", UnitPriceVND: 75_000_000, FeeVND: 15_000},
	})
	if err != nil {
		t.Fatalf("replay ledger: %v", err)
	}
	last := results[len(results)-1]
	if last.QuantityAfter != "2.3" {
		t.Fatalf("quantity after = %s, want 2.3", last.QuantityAfter)
	}
	if last.CostBasisAfterVND != 162_991_143 {
		t.Fatalf("cost basis after = %d, want 162991143", last.CostBasisAfterVND)
	}
	if last.RealizedPNLVND != 4_946_143 {
		t.Fatalf("realized pnl = %d, want 4946143", last.RealizedPNLVND)
	}
}

func TestReplayLedgerRejectsOversell(t *testing.T) {
	_, err := replayLedger([]ledgerTrade{
		{Side: TradeBuy, Quantity: "1", UnitPriceVND: 100},
		{Side: TradeSell, Quantity: "1.000000000001", UnitPriceVND: 100},
	})
	if !errors.Is(err, ErrOversell) {
		t.Fatalf("expected oversell, got %v", err)
	}
}

func TestNormalizeDecimalRejectsFloatyOrOverPreciseInput(t *testing.T) {
	if normalized, _, err := normalizeDecimal("0001.230000000000"); err != nil || normalized != "1.23" {
		t.Fatalf("normalize decimal = %q, %v", normalized, err)
	}
	for _, value := range []string{"", "1.", ".5", "1.0000000000001", "1e-3", "-1"} {
		if _, _, err := normalizeDecimal(value); !errors.Is(err, ErrValidation) {
			t.Fatalf("expected validation for %q, got %v", value, err)
		}
	}
}

func TestSummarizeMissingPriceAndZeroCostBasisStates(t *testing.T) {
	missing := summarize("10", 1_000_000, 0, nil)
	if missing.MarketValueVND != nil || missing.ValuationStatus != "missing_price" {
		t.Fatalf("missing summary should not invent market value: %#v", missing)
	}
	price := PricePoint{UnitPriceVND: 12_345, Source: "manual"}
	zeroCost := summarize("2", 0, 0, &price)
	if zeroCost.MarketValueVND == nil || *zeroCost.MarketValueVND != 24_690 {
		t.Fatalf("market value = %#v, want 24690", zeroCost.MarketValueVND)
	}
	if !zeroCost.UnrealizedNotComparable || zeroCost.UnrealizedPNLPercent != nil {
		t.Fatalf("zero cost basis should be not comparable: %#v", zeroCost)
	}
}
