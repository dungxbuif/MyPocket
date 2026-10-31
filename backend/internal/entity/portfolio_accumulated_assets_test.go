package entity

import "testing"

func TestPortfolioMarketValueKeepsUnknownValuationDistinctFromZero(t *testing.T) {
	position := PortfolioPosition{AssetID: "asset-1", Quantity: "1", QuantityScaled: QuantityScale, CostBasis: 100}
	unknown, err := PortfolioMarketValue(position, nil)
	if err != nil {
		t.Fatal(err)
	}
	if unknown.MarketValue != nil || unknown.UnrealizedPnL != nil {
		t.Fatalf("unknown valuation must remain nil, got market=%v pnl=%v", unknown.MarketValue, unknown.UnrealizedPnL)
	}

	price := int64(0)
	known, err := PortfolioMarketValue(position, &price)
	if err != nil {
		t.Fatal(err)
	}
	if known.MarketValue == nil || *known.MarketValue != 0 || known.UnrealizedPnL == nil || *known.UnrealizedPnL != -100 {
		t.Fatalf("zero valuation must remain known, got %+v", known)
	}
}

func TestPortfolioMarketValueUsesFractionalQuantityWithoutUnitGuessing(t *testing.T) {
	position := PortfolioPosition{AssetID: "asset-1", Quantity: "0.5", QuantityScaled: QuantityScale / 2, CostBasis: 3_000_000}
	price := int64(8_000_000)
	valued, err := PortfolioMarketValue(position, &price)
	if err != nil {
		t.Fatal(err)
	}
	if valued.MarketValue == nil || *valued.MarketValue != 4_000_000 {
		t.Fatalf("market value=%v, want 4000000", valued.MarketValue)
	}
}
