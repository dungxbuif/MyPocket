package entity

import (
	"errors"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	QuantityScale      int64 = 100000000
	PortfolioTradeBuy        = "buy"
	PortfolioTradeSell       = "sell"
)

var (
	ErrPortfolioQuantityInvalid      = errors.New("portfolio quantity is invalid")
	ErrPortfolioTradeInvalid         = errors.New("portfolio trade is invalid")
	ErrPortfolioInsufficientQuantity = errors.New("portfolio position is insufficient for sell")
	ErrPortfolioOverflow             = errors.New("portfolio calculation overflow")
)

// PortfolioAsset is an owner-scoped investment instrument. Prices are VND
// snapshots and are deliberately optional: an unknown price is not zero.
type PortfolioAsset struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	OwnerID       string     `json:"owner_id" gorm:"index;not null"`
	Symbol        string     `json:"symbol" gorm:"not null"`
	Name          string     `json:"name" gorm:"not null"`
	LatestPrice   *int64     `json:"latest_price,omitempty"`
	LatestPriceAt *time.Time `json:"latest_price_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PortfolioTrade is immutable history. QuantityScaled stores quantity in
// fixed-point units; Quantity is a presentation field populated by services.
type PortfolioTrade struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OwnerID        string    `json:"owner_id" gorm:"index;not null"`
	AssetID        string    `json:"asset_id" gorm:"index;not null"`
	Side           string    `json:"side" gorm:"not null"`
	Quantity       string    `json:"quantity,omitempty" gorm:"-"`
	QuantityScaled int64     `json:"-" gorm:"column:quantity_scaled;not null"`
	UnitPrice      int64     `json:"unit_price" gorm:"not null"`
	Fee            int64     `json:"fee" gorm:"not null;default:0"`
	OccurredAt     time.Time `json:"occurred_at" gorm:"index;not null"`
	Note           *string   `json:"note,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PortfolioPosition struct {
	AssetID        string     `json:"asset_id"`
	Symbol         string     `json:"symbol,omitempty"`
	Name           string     `json:"name,omitempty"`
	Quantity       string     `json:"quantity"`
	QuantityScaled int64      `json:"-"`
	CostBasis      int64      `json:"cost_basis"`
	AverageCost    int64      `json:"average_cost"`
	RealizedPnL    int64      `json:"realized_pnl"`
	MarketValue    *int64     `json:"market_value,omitempty"`
	UnrealizedPnL  *int64     `json:"unrealized_pnl,omitempty"`
	LatestPrice    *int64     `json:"latest_price,omitempty"`
	LatestPriceAt  *time.Time `json:"latest_price_at,omitempty"`
}

type PortfolioSummary struct {
	Assets             []PortfolioAsset    `json:"assets"`
	Positions          []PortfolioPosition `json:"positions"`
	Trades             []PortfolioTrade    `json:"trades"`
	TotalCostBasis     int64               `json:"total_cost_basis"`
	TotalMarketValue   *int64              `json:"total_market_value,omitempty"`
	TotalUnrealizedPnL *int64              `json:"total_unrealized_pnl,omitempty"`
}

func ParsePortfolioQuantity(raw string) (int64, error) {
	if raw == "" || strings.TrimSpace(raw) != raw || strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") {
		return 0, ErrPortfolioQuantityInvalid
	}
	parts := strings.Split(raw, ".")
	if len(parts) > 2 || parts[0] == "" || (len(parts) == 2 && parts[1] == "") {
		return 0, ErrPortfolioQuantityInvalid
	}
	whole, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, ErrPortfolioQuantityInvalid
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > 8 {
		return 0, ErrPortfolioQuantityInvalid
	}
	frac += strings.Repeat("0", 8-len(frac))
	fracValue, err := strconv.ParseUint(frac, 10, 64)
	if err != nil {
		return 0, ErrPortfolioQuantityInvalid
	}
	if whole > uint64(math.MaxInt64)/uint64(QuantityScale) {
		return 0, ErrPortfolioQuantityInvalid
	}
	scaled := whole*uint64(QuantityScale) + fracValue
	if scaled == 0 || scaled > uint64(math.MaxInt64) {
		return 0, ErrPortfolioQuantityInvalid
	}
	return int64(scaled), nil
}

func FormatPortfolioQuantity(scaled int64) string {
	if scaled <= 0 {
		return "0"
	}
	whole := scaled / QuantityScale
	frac := scaled % QuantityScale
	if frac == 0 {
		return strconv.FormatInt(whole, 10)
	}
	fracText := strconv.FormatInt(frac+QuantityScale, 10)[1:]
	fracText = strings.TrimRight(fracText, "0")
	return strconv.FormatInt(whole, 10) + "." + fracText
}

func ComputePortfolioPosition(trades []PortfolioTrade) (PortfolioPosition, error) {
	ordered := append([]PortfolioTrade(nil), trades...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if !ordered[i].OccurredAt.Equal(ordered[j].OccurredAt) {
			return ordered[i].OccurredAt.Before(ordered[j].OccurredAt)
		}
		if !ordered[i].CreatedAt.Equal(ordered[j].CreatedAt) {
			return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
		}
		return ordered[i].ID < ordered[j].ID
	})
	quantity := big.NewInt(0)
	costBasis := big.NewInt(0)
	realized := big.NewInt(0)
	var assetID string
	for _, trade := range ordered {
		if assetID == "" {
			assetID = trade.AssetID
		}
		if trade.QuantityScaled <= 0 || trade.UnitPrice < 0 || trade.Fee < 0 || (trade.Side != PortfolioTradeBuy && trade.Side != PortfolioTradeSell) {
			return PortfolioPosition{}, ErrPortfolioTradeInvalid
		}
		qty := big.NewInt(trade.QuantityScaled)
		if trade.Side == PortfolioTradeBuy {
			purchase := roundedRatio(new(big.Int).Mul(qty, big.NewInt(trade.UnitPrice)), big.NewInt(QuantityScale))
			purchase.Add(purchase, big.NewInt(trade.Fee))
			quantity.Add(quantity, qty)
			costBasis.Add(costBasis, purchase)
			continue
		}
		if qty.Cmp(quantity) > 0 {
			return PortfolioPosition{}, ErrPortfolioInsufficientQuantity
		}
		costRemoved := roundedRatio(new(big.Int).Mul(qty, costBasis), quantity)
		proceeds := roundedRatio(new(big.Int).Mul(qty, big.NewInt(trade.UnitPrice)), big.NewInt(QuantityScale))
		proceeds.Sub(proceeds, big.NewInt(trade.Fee))
		proceeds.Sub(proceeds, costRemoved)
		realized.Add(realized, proceeds)
		quantity.Sub(quantity, qty)
		costBasis.Sub(costBasis, costRemoved)
	}
	quantityInt, err := bigInt64(quantity)
	if err != nil {
		return PortfolioPosition{}, err
	}
	costInt, err := bigInt64(costBasis)
	if err != nil {
		return PortfolioPosition{}, err
	}
	realizedInt, err := bigInt64(realized)
	if err != nil {
		return PortfolioPosition{}, err
	}
	average := int64(0)
	if quantity.Sign() > 0 {
		averageBig := roundedRatio(new(big.Int).Mul(costBasis, big.NewInt(QuantityScale)), quantity)
		average, err = bigInt64(averageBig)
		if err != nil {
			return PortfolioPosition{}, err
		}
	}
	return PortfolioPosition{AssetID: assetID, Quantity: FormatPortfolioQuantity(quantityInt), QuantityScaled: quantityInt, CostBasis: costInt, AverageCost: average, RealizedPnL: realizedInt}, nil
}

func PortfolioMarketValue(position PortfolioPosition, latestPrice *int64) (PortfolioPosition, error) {
	if latestPrice == nil {
		return position, nil
	}
	if *latestPrice < 0 {
		return PortfolioPosition{}, ErrPortfolioTradeInvalid
	}
	value := roundedRatio(new(big.Int).Mul(big.NewInt(position.QuantityScaled), big.NewInt(*latestPrice)), big.NewInt(QuantityScale))
	marketValue, err := bigInt64(value)
	if err != nil {
		return PortfolioPosition{}, err
	}
	unrealized := marketValue - position.CostBasis
	position.MarketValue = &marketValue
	position.UnrealizedPnL = &unrealized
	position.LatestPrice = latestPrice
	return position, nil
}

func roundedRatio(numerator, denominator *big.Int) *big.Int {
	if denominator.Sign() == 0 {
		return big.NewInt(0)
	}
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	if remainder.Sign() == 0 {
		return quotient
	}
	twice := new(big.Int).Abs(remainder)
	twice.Lsh(twice, 1)
	if twice.Cmp(new(big.Int).Abs(denominator)) >= 0 {
		if numerator.Sign() < 0 {
			quotient.Sub(quotient, big.NewInt(1))
		} else {
			quotient.Add(quotient, big.NewInt(1))
		}
	}
	return quotient
}

func bigInt64(value *big.Int) (int64, error) {
	if value == nil || !value.IsInt64() {
		return 0, ErrPortfolioOverflow
	}
	return value.Int64(), nil
}
