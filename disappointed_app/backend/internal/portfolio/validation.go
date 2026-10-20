package portfolio

import (
	"strings"
	"time"
)

func ValidateCreatePosition(input CreatePositionInput) (CreatePositionInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Symbol = strings.ToUpper(strings.TrimSpace(input.Symbol))
	input.Exchange = strings.ToUpper(strings.TrimSpace(input.Exchange))
	input.Unit = strings.ToLower(strings.TrimSpace(input.Unit))
	input.ProviderKey = strings.TrimSpace(input.ProviderKey)
	input.ProviderSymbol = strings.TrimSpace(input.ProviderSymbol)
	if input.PricingMode == "" {
		input.PricingMode = PricingManual
	}
	if !validAssetType(input.Type) || input.Name == "" || !validUnit(input.Type, input.Unit) {
		return CreatePositionInput{}, ErrValidation
	}
	if len(input.Name) > 120 || len(input.Symbol) > 32 || len(input.Exchange) > 24 || len(input.Unit) > 24 {
		return CreatePositionInput{}, ErrValidation
	}
	if input.PricingMode != PricingManual && input.PricingMode != PricingAutomatic {
		return CreatePositionInput{}, ErrValidation
	}
	if input.PricingMode == PricingAutomatic && (input.ProviderKey == "" || input.ProviderSymbol == "") {
		return CreatePositionInput{}, ErrValidation
	}
	return input, nil
}

func ValidateAddTrade(input AddTradeInput) (AddTradeInput, error) {
	input.Note = strings.TrimSpace(input.Note)
	if input.Side != TradeBuy && input.Side != TradeSell {
		return AddTradeInput{}, ErrValidation
	}
	if _, _, err := normalizeDecimal(input.Quantity); err != nil {
		return AddTradeInput{}, err
	}
	if input.UnitPriceVND < 0 || input.FeeVND < 0 || input.OccurredAt.IsZero() || len(input.Note) > 240 {
		return AddTradeInput{}, ErrValidation
	}
	return input, nil
}

func ValidateUpdateTrade(input UpdateTradeInput) (UpdateTradeInput, error) {
	add := AddTradeInput{
		Side:         input.Side,
		Quantity:     input.Quantity,
		UnitPriceVND: input.UnitPriceVND,
		FeeVND:       input.FeeVND,
		OccurredAt:   input.OccurredAt,
		Note:         input.Note,
	}
	if _, err := ValidateAddTrade(add); err != nil {
		return UpdateTradeInput{}, err
	}
	input.Note = strings.TrimSpace(input.Note)
	return input, nil
}

func ValidateAddPrice(input AddPriceInput) (AddPriceInput, error) {
	input.Source = strings.TrimSpace(input.Source)
	input.ProviderQuoteID = strings.TrimSpace(input.ProviderQuoteID)
	if input.UnitPriceVND <= 0 || input.PricedAt.IsZero() || input.Source == "" || len(input.Source) > 48 || len(input.ProviderQuoteID) > 160 {
		return AddPriceInput{}, ErrValidation
	}
	return input, nil
}

func validAssetType(t AssetType) bool {
	switch t {
	case AssetGold, AssetStock, AssetCrypto, AssetForeignCurrency, AssetOther:
		return true
	default:
		return false
	}
}

func validUnit(t AssetType, unit string) bool {
	if unit == "" {
		return false
	}
	switch t {
	case AssetGold:
		return unit == "gram" || unit == "tael" || unit == "ounce"
	case AssetStock:
		return unit == "share"
	case AssetCrypto:
		return unit == "token"
	case AssetForeignCurrency, AssetOther:
		return unit == "unit"
	default:
		return false
	}
}

func nowIfZero(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value
}
