package portfolio

import "math/big"

type ledgerTrade struct {
	ID           string
	Side         TradeSide
	Quantity     string
	UnitPriceVND int64
	FeeVND       int64
}

type ledgerResult struct {
	QuantityAfter       string
	CostBasisAfterVND   int64
	RealizedPNLVND      int64
	FinalQuantity       string
	FinalCostBasisVND   int64
	TotalRealizedPNLVND int64
}

func replayLedger(trades []ledgerTrade) ([]ledgerResult, error) {
	quantity := new(big.Rat)
	costBasis := int64(0)
	totalRealized := int64(0)
	results := make([]ledgerResult, 0, len(trades))
	for _, trade := range trades {
		normalized, tradeQuantity, err := normalizeDecimal(trade.Quantity)
		if err != nil {
			return nil, err
		}
		_ = normalized
		realized := int64(0)
		switch trade.Side {
		case TradeBuy:
			quantity.Add(quantity, tradeQuantity)
			costBasis += roundQuantityMoney(tradeQuantity, trade.UnitPriceVND) + trade.FeeVND
		case TradeSell:
			if quantity.Cmp(tradeQuantity) < 0 {
				return nil, ErrOversell
			}
			if quantity.Sign() == 0 {
				return nil, ErrOversell
			}
			removedCostRat := new(big.Rat).Mul(big.NewRat(costBasis, 1), tradeQuantity)
			removedCostRat.Quo(removedCostRat, quantity)
			removedCost := roundRat(removedCostRat).Int64()
			proceeds := roundQuantityMoney(tradeQuantity, trade.UnitPriceVND) - trade.FeeVND
			realized = proceeds - removedCost
			totalRealized += realized
			quantity.Sub(quantity, tradeQuantity)
			costBasis -= removedCost
			if quantity.Sign() == 0 {
				costBasis = 0
			}
		default:
			return nil, ErrValidation
		}
		finalQuantity := ratToDecimalString(quantity)
		results = append(results, ledgerResult{
			QuantityAfter:       finalQuantity,
			CostBasisAfterVND:   costBasis,
			RealizedPNLVND:      realized,
			FinalQuantity:       finalQuantity,
			FinalCostBasisVND:   costBasis,
			TotalRealizedPNLVND: totalRealized,
		})
	}
	return results, nil
}

func summarize(quantityText string, costBasisVND int64, realizedPNLVND int64, latest *PricePoint) AssetSummary {
	if quantityText == "" {
		quantityText = "0"
	}
	summary := AssetSummary{
		Quantity:        quantityText,
		CostBasisVND:    costBasisVND,
		RealizedPNLVND:  realizedPNLVND,
		ValuationStatus: "missing_price",
	}
	if latest == nil {
		if costBasisVND == 0 {
			summary.UnrealizedNotComparable = true
		}
		return summary
	}
	quantity := parseStoredDecimal(quantityText)
	marketValue := roundQuantityMoney(quantity, latest.UnitPriceVND)
	unrealized := marketValue - costBasisVND
	summary.CurrentUnitPriceVND = &latest.UnitPriceVND
	summary.MarketValueVND = &marketValue
	summary.UnrealizedPNLVND = &unrealized
	summary.ValuationStatus = "current"
	summary.PricedAt = &latest.PricedAt
	summary.PriceSource = latest.Source
	if costBasisVND == 0 {
		summary.UnrealizedNotComparable = true
		return summary
	}
	summary.UnrealizedPNLPercent = percentString(unrealized, costBasisVND)
	return summary
}
