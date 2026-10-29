package portfolio

import "time"

type AssetType string
type PricingMode string
type TradeSide string

const (
	AssetGold            AssetType = "gold"
	AssetStock           AssetType = "stock"
	AssetCrypto          AssetType = "crypto"
	AssetForeignCurrency AssetType = "foreign_currency"
	AssetOther           AssetType = "other"

	PricingManual    PricingMode = "manual"
	PricingAutomatic PricingMode = "automatic"

	TradeBuy  TradeSide = "buy"
	TradeSell TradeSide = "sell"
)

type CreatePositionInput struct {
	ID                string
	Type              AssetType
	Symbol            string
	Exchange          string
	Name              string
	Unit              string
	PricingMode       PricingMode
	ProviderKey       string
	ProviderSymbol    string
	IncludeInNetWorth *bool
}

type UpdatePositionInput struct {
	Name              string
	Symbol            string
	Exchange          string
	Unit              string
	PricingMode       PricingMode
	ProviderKey       string
	ProviderSymbol    string
	IncludeInNetWorth *bool
	BaseVersion       int64
}

type Position struct {
	ID                string       `json:"id"`
	UserID            string       `json:"user_id"`
	Type              AssetType    `json:"type"`
	Symbol            string       `json:"symbol"`
	Exchange          string       `json:"exchange"`
	Name              string       `json:"name"`
	Unit              string       `json:"unit"`
	ReportingCurrency string       `json:"reporting_currency"`
	PricingMode       PricingMode  `json:"pricing_mode"`
	ProviderKey       string       `json:"provider_key,omitempty"`
	ProviderSymbol    string       `json:"provider_symbol,omitempty"`
	IncludeInNetWorth bool         `json:"include_in_net_worth"`
	ArchivedAt        *time.Time   `json:"archived_at,omitempty"`
	Version           int64        `json:"version"`
	Summary           AssetSummary `json:"summary"`
	LatestPrice       *PricePoint  `json:"latest_price,omitempty"`
	Trades            []Trade      `json:"trades,omitempty"`
	PriceHistory      []PricePoint `json:"price_history,omitempty"`
}

type AddTradeInput struct {
	ID           string
	Side         TradeSide
	Quantity     string
	UnitPriceVND int64
	FeeVND       int64
	OccurredAt   time.Time
	Note         string
	BaseVersion  int64
}

type UpdateTradeInput struct {
	Side         TradeSide
	Quantity     string
	UnitPriceVND int64
	FeeVND       int64
	OccurredAt   time.Time
	Note         string
	BaseVersion  int64
}

type Trade struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	AssetID           string     `json:"asset_id"`
	Side              TradeSide  `json:"side"`
	Quantity          string     `json:"quantity"`
	UnitPriceVND      int64      `json:"unit_price_vnd"`
	FeeVND            int64      `json:"fee_vnd"`
	OccurredAt        time.Time  `json:"occurred_at"`
	QuantityAfter     string     `json:"quantity_after"`
	CostBasisAfterVND int64      `json:"cost_basis_after_vnd"`
	RealizedPNLVND    int64      `json:"realized_pnl_vnd"`
	Note              string     `json:"note"`
	Version           int64      `json:"version"`
	ArchivedAt        *time.Time `json:"archived_at,omitempty"`
}

type AddPriceInput struct {
	ID              string
	UnitPriceVND    int64
	PricedAt        time.Time
	Source          string
	ProviderQuoteID string
	BaseVersion     int64
}

type PricePoint struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	AssetID         string    `json:"asset_id"`
	UnitPriceVND    int64     `json:"unit_price_vnd"`
	PricedAt        time.Time `json:"priced_at"`
	Source          string    `json:"source"`
	ProviderQuoteID string    `json:"provider_quote_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type AssetSummary struct {
	Quantity                string     `json:"quantity"`
	CostBasisVND            int64      `json:"cost_basis_vnd"`
	RealizedPNLVND          int64      `json:"realized_pnl_vnd"`
	CurrentUnitPriceVND     *int64     `json:"current_unit_price_vnd"`
	MarketValueVND          *int64     `json:"market_value_vnd"`
	UnrealizedPNLVND        *int64     `json:"unrealized_pnl_vnd"`
	UnrealizedPNLPercent    *string    `json:"unrealized_pnl_percent"`
	UnrealizedNotComparable bool       `json:"unrealized_not_comparable"`
	ValuationStatus         string     `json:"valuation_status"`
	PricedAt                *time.Time `json:"priced_at,omitempty"`
	PriceSource             string     `json:"price_source,omitempty"`
}

type PortfolioSummary struct {
	InvestmentMarketValueVND int64 `json:"investment_market_value_vnd"`
	MissingPriceCount        int   `json:"missing_price_count"`
	IncludedPositionCount    int   `json:"included_position_count"`
	PositionCount            int   `json:"position_count"`
}

type PriceRefreshCandidate struct {
	UserID         string
	AssetID        string
	ProviderKey    string
	ProviderSymbol string
	BaseVersion    int64
}

type ProviderQuote struct {
	ProviderKey     string
	ProviderSymbol  string
	UnitPriceVND    int64
	PricedAt        time.Time
	ProviderQuoteID string
}
