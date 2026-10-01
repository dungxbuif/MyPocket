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
