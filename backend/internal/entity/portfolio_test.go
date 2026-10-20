package entity

import (
	"errors"
	"testing"
	"time"
)

func TestParsePortfolioQuantityUsesEightFixedDecimals(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int64
	}{
		{name: "whole", raw: "2", want: 200000000},
		{name: "fraction", raw: "2.125", want: 212500000},
		{name: "trailing zeros", raw: "0.00000001", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePortfolioQuantity(tt.raw)
			if err != nil {
				t.Fatalf("ParsePortfolioQuantity() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParsePortfolioQuantity() = %d, want %d", got, tt.want)
			}
			if formatted := FormatPortfolioQuantity(got); formatted != tt.raw {
				t.Fatalf("FormatPortfolioQuantity() = %q, want %q", formatted, tt.raw)
			}
		})
	}
}

func TestParsePortfolioQuantityRejectsMoreThanEightDecimals(t *testing.T) {
	if _, err := ParsePortfolioQuantity("1.000000001"); !errors.Is(err, ErrPortfolioQuantityInvalid) {
		t.Fatalf("ParsePortfolioQuantity() error = %v, want ErrPortfolioQuantityInvalid", err)
	}
}

func TestComputePortfolioPositionUsesWeightedAverageCost(t *testing.T) {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	trades := []PortfolioTrade{
		{ID: "buy-1", Side: PortfolioTradeBuy, QuantityScaled: 100000000, UnitPrice: 100, Fee: 10, OccurredAt: base},
		{ID: "buy-2", Side: PortfolioTradeBuy, QuantityScaled: 100000000, UnitPrice: 200, Fee: 0, OccurredAt: base.Add(time.Hour)},
		{ID: "sell-1", Side: PortfolioTradeSell, QuantityScaled: 50000000, UnitPrice: 250, Fee: 5, OccurredAt: base.Add(2 * time.Hour)},
	}
	position, err := ComputePortfolioPosition(trades)
	if err != nil {
		t.Fatalf("ComputePortfolioPosition() error = %v", err)
	}
	if position.QuantityScaled != 150000000 {
		t.Fatalf("quantity = %d, want 150000000", position.QuantityScaled)
	}
	if position.CostBasis != 232 {
		t.Fatalf("cost basis = %d, want 232", position.CostBasis)
	}
	if position.AverageCost != 155 {
		t.Fatalf("average cost = %d, want 155", position.AverageCost)
	}
	if position.RealizedPnL != 42 {
		t.Fatalf("realized pnl = %d, want 42", position.RealizedPnL)
	}
}

func TestComputePortfolioPositionRejectsSellOverPosition(t *testing.T) {
	_, err := ComputePortfolioPosition([]PortfolioTrade{{
		ID: "sell-1", Side: PortfolioTradeSell, QuantityScaled: QuantityScale, UnitPrice: 100, OccurredAt: time.Now().UTC(),
	}})
	if !errors.Is(err, ErrPortfolioInsufficientQuantity) {
		t.Fatalf("ComputePortfolioPosition() error = %v, want ErrPortfolioInsufficientQuantity", err)
	}
}

func TestComputePortfolioPositionSupportsExplicitCompensatingTradeWithoutRewritingHistory(t *testing.T) {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	original := PortfolioTrade{
		ID:             "buy-original",
		AssetID:        "asset-1",
		Side:           PortfolioTradeBuy,
		QuantityScaled: 2 * QuantityScale,
		UnitPrice:      100,
		OccurredAt:     base,
	}
	compensating := PortfolioTrade{
		ID:             "sell-compensating-original",
		AssetID:        "asset-1",
		Side:           PortfolioTradeSell,
		QuantityScaled: 2 * QuantityScale,
		UnitPrice:      100,
		OccurredAt:     base.Add(time.Minute),
		Note:           stringPtr("compensates buy-original; user-confirmed correction"),
	}
	history := []PortfolioTrade{original, compensating}
	position, err := ComputePortfolioPosition(history)
	if err != nil {
		t.Fatalf("explicit compensating trade must use the existing weighted-average rules: %v", err)
	}
	if position.QuantityScaled != 0 || position.CostBasis != 0 || position.RealizedPnL != 0 {
		t.Fatalf("compensation should neutralize the explicitly reversed position: %+v", position)
	}
	if len(history) != 2 || history[0].ID != original.ID || history[1].ID != compensating.ID {
		t.Fatalf("compensation must append history, not rewrite it: %+v", history)
	}
}

func TestApplyPortfolioTradeRevisionsRecomputesReplacementAndVoid(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	trades := []PortfolioTrade{
		{ID: "buy-1", AssetID: "asset", Side: PortfolioTradeBuy, QuantityScaled: 10 * QuantityScale, UnitPrice: 100, OccurredAt: base, CreatedAt: base},
		{ID: "buy-2", AssetID: "asset", Side: PortfolioTradeBuy, QuantityScaled: 10 * QuantityScale, UnitPrice: 200, OccurredAt: base.Add(time.Hour), CreatedAt: base.Add(time.Hour)},
		{ID: "sell-1", AssetID: "asset", Side: PortfolioTradeSell, QuantityScaled: 5 * QuantityScale, UnitPrice: 300, OccurredAt: base.Add(2 * time.Hour), CreatedAt: base.Add(2 * time.Hour)},
	}
	revised, err := ApplyPortfolioTradeRevisions(trades, []PortfolioTradeRevision{{
		SourceTradeID: "buy-1", Revision: 1, Operation: PortfolioTradeRevisionReplace,
		Side: PortfolioTradeBuy, QuantityScaled: 10 * QuantityScale, UnitPrice: 120,
		OccurredAt: base, CreatedAt: base.Add(3 * time.Hour), Reason: "corrected price",
	}})
	if err != nil {
		t.Fatalf("apply replacement: %v", err)
	}
	position, err := ComputePortfolioPosition(revised)
	if err != nil {
		t.Fatalf("compute replacement: %v", err)
	}
	if position.QuantityScaled != 15*QuantityScale || position.CostBasis != 2400 || position.RealizedPnL != 700 {
		t.Fatalf("replacement position=%+v, want quantity=15 cost=2400 realized=700", position)
	}
	voided, err := ApplyPortfolioTradeRevisions(trades, []PortfolioTradeRevision{{
		SourceTradeID: "buy-1", Revision: 1, Operation: PortfolioTradeRevisionVoid,
		CreatedAt: base.Add(3 * time.Hour), Reason: "duplicate entry",
	}})
	if err != nil {
		t.Fatalf("apply void: %v", err)
	}
	position, err = ComputePortfolioPosition(voided)
	if err != nil {
		t.Fatalf("compute void: %v", err)
	}
	if position.QuantityScaled != 5*QuantityScale || position.CostBasis != 1000 || position.RealizedPnL != 500 {
		t.Fatalf("void position=%+v, want quantity=5 cost=1000 realized=500", position)
	}
}

func TestApplyPortfolioTradeRevisionsRejectsDuplicateRevisionAndInvalidResult(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	trades := []PortfolioTrade{{ID: "buy-1", AssetID: "asset", Side: PortfolioTradeBuy, QuantityScaled: QuantityScale, UnitPrice: 100, OccurredAt: base, CreatedAt: base}}
	_, err := ApplyPortfolioTradeRevisions(trades, []PortfolioTradeRevision{
		{SourceTradeID: "buy-1", Revision: 1, Operation: PortfolioTradeRevisionVoid, CreatedAt: base.Add(time.Hour), Reason: "duplicate entry"},
		{SourceTradeID: "buy-1", Revision: 1, Operation: PortfolioTradeRevisionVoid, CreatedAt: base.Add(2 * time.Hour), Reason: "duplicate entry"},
	})
	if !errors.Is(err, ErrPortfolioRevisionConflict) {
		t.Fatalf("duplicate revision error=%v, want conflict", err)
	}
	_, err = ApplyPortfolioTradeRevisions(trades, []PortfolioTradeRevision{{
		SourceTradeID: "buy-1", Revision: 1, Operation: PortfolioTradeRevisionReplace,
		Side: PortfolioTradeSell, QuantityScaled: 2 * QuantityScale, UnitPrice: 100, OccurredAt: base, CreatedAt: base.Add(time.Hour),
		Reason: "wrong side",
	}})
	if !errors.Is(err, ErrPortfolioInsufficientQuantity) {
		t.Fatalf("invalid replacement error=%v, want insufficient quantity", err)
	}
}

func stringPtr(value string) *string { return &value }
