package portfolio

import (
	"errors"
	"math"
	"testing"
)

func TestReplayLedgerRejectsMoneyOverflow(t *testing.T) {
	cases := map[string][]ledgerTrade{
		"quantity product":   {{Side: TradeBuy, Quantity: "2", UnitPriceVND: math.MaxInt64}},
		"buy fee":            {{Side: TradeBuy, Quantity: "1", UnitPriceVND: math.MaxInt64, FeeVND: 1}},
		"accumulated basis":  {{Side: TradeBuy, Quantity: "1", UnitPriceVND: math.MaxInt64}, {Side: TradeBuy, Quantity: "1", UnitPriceVND: 1}},
		"realized loss":      {{Side: TradeBuy, Quantity: "1", UnitPriceVND: math.MaxInt64}, {Side: TradeSell, Quantity: "1", UnitPriceVND: 1, FeeVND: math.MaxInt64}},
		"accumulated profit": {{Side: TradeBuy, Quantity: "1", UnitPriceVND: 1}, {Side: TradeSell, Quantity: "1", UnitPriceVND: math.MaxInt64}, {Side: TradeBuy, Quantity: "1", UnitPriceVND: 1}, {Side: TradeSell, Quantity: "1", UnitPriceVND: math.MaxInt64}},
	}
	for name, trades := range cases {
		t.Run(name, func(t *testing.T) {
			if result, err := replayLedger(trades); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected overflow validation, got result=%+v err=%v", result, err)
			}
		})
	}
}

func TestReplayLedgerAcceptsMaximumRepresentableBasis(t *testing.T) {
	result, err := replayLedger([]ledgerTrade{{Side: TradeBuy, Quantity: "1", UnitPriceVND: math.MaxInt64 - 1, FeeVND: 1}})
	if err != nil || result[0].CostBasisAfterVND != math.MaxInt64 {
		t.Fatalf("boundary rejected: %+v %v", result, err)
	}
}

func TestSummaryRejectsOutOfRangeMarketValue(t *testing.T) {
	if result, err := summarize("2", 1, 0, &PricePoint{UnitPriceVND: math.MaxInt64}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation, got %+v %v", result, err)
	}
}
