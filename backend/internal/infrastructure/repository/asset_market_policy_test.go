package repository

import (
	"errors"
	"testing"

	"github.com/mypocket/backend/internal/entity"
	marketrepo "github.com/mypocket/backend/internal/repository"
)

func TestMarketBindingPolicyRequiresConfirmedMatchingUnit(t *testing.T) {
	unit := "chỉ"
	product := entity.MarketProduct{
		ID: "gold-btmc", ProviderKey: "btmc", ExternalProductID: "ring-buy",
		Name: "Nhẫn tròn trơn", Kind: entity.PortfolioAssetKindGold, Unit: "chỉ",
		Currency: "VND", ValuationSide: "buyback", Enabled: true,
	}
	for _, tc := range []struct {
		name    string
		asset   entity.PortfolioAsset
		product entity.MarketProduct
		want    error
	}{
		{name: "legacy unit is not confirmed", asset: entity.PortfolioAsset{UnitLabel: &unit}, product: product, want: marketrepo.ErrMarketBindingInvalid},
		{name: "catalog unit differs", asset: entity.PortfolioAsset{UnitLabel: &unit, UnitConfirmed: true}, product: func() entity.MarketProduct { p := product; p.Unit = "lượng"; return p }(), want: marketrepo.ErrMarketBindingUnitMismatch},
		{name: "exact product unit", asset: entity.PortfolioAsset{UnitLabel: &unit, UnitConfirmed: true}, product: product, want: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateMarketProductForAsset(tc.asset, tc.product); !errors.Is(err, tc.want) {
				t.Fatalf("validateMarketProductForAsset() error=%v, want=%v", err, tc.want)
			}
		})
	}
}
