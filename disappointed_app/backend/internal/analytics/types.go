package analytics

import "time"

type Filter struct {
	From, To time.Time
	WalletID string
}
type Summary struct {
	GeneratedAt          time.Time `json:"generated_at"`
	Timezone             string    `json:"timezone"`
	From                 string    `json:"from"`
	To                   string    `json:"to"`
	DataVersion          int64     `json:"data_version"`
	IncomeVND            int64     `json:"income_vnd"`
	ExpenseVND           int64     `json:"expense_vnd"`
	NetIncomeVND         int64     `json:"net_income_vnd"`
	DailyAverageVND      int64     `json:"daily_average_vnd"`
	NotComparable        bool      `json:"not_comparable"`
	IncomeChangePercent  float64   `json:"income_change_percent,omitempty"`
	ExpenseChangePercent float64   `json:"expense_change_percent,omitempty"`
}
type CategoryTotal struct {
	CategoryID   string  `json:"category_id,omitempty"`
	CategoryName string  `json:"category_name"`
	AmountVND    int64   `json:"amount_vnd"`
	SharePercent float64 `json:"share_percent"`
}
type DailyTotal struct {
	Date             string `json:"date"`
	IncomeVND        int64  `json:"income_vnd"`
	ExpenseVND       int64  `json:"expense_vnd"`
	NetIncomeVND     int64  `json:"net_income_vnd"`
	CumulativeNetVND int64  `json:"cumulative_net_vnd"`
}
type Report struct {
	Summary    Summary         `json:"summary"`
	Categories []CategoryTotal `json:"categories,omitempty"`
	Daily      []DailyTotal    `json:"daily,omitempty"`
	Prior      *Summary        `json:"prior,omitempty"`
	Periods    []Summary       `json:"periods,omitempty"`
}
type Dashboard struct {
	Summary                  Summary             `json:"summary"`
	NetWorthVND              int64               `json:"net_worth_vnd"`
	WalletNetWorthVND        int64               `json:"wallet_net_worth_vnd"`
	InvestmentMarketValueVND int64               `json:"investment_market_value_vnd"`
	CombinedNetWorthVND      int64               `json:"combined_net_worth_vnd"`
	MissingAssetPriceCount   int                 `json:"missing_asset_price_count"`
	Wallets                  []Wallet            `json:"wallets"`
	Recent                   []RecentTransaction `json:"recent_transactions"`
}
type InsiderCategory struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name"`
	TransactionCount int64  `json:"transaction_count"`
}
type InsiderReport struct {
	SelectedCategory     *InsiderCategory `json:"selected_category,omitempty"`
	SpentVND             int64            `json:"spent_vnd"`
	AverageDailyVND      int64            `json:"average_daily_vnd"`
	PriorAverageDailyVND int64            `json:"prior_average_daily_vnd"`
	ChangePercent        float64          `json:"change_percent,omitempty"`
	NotComparable        bool             `json:"not_comparable"`
	ElapsedDays          int              `json:"elapsed_days"`
	GeneratedAt          time.Time        `json:"generated_at"`
	Timezone             string           `json:"timezone"`
	From                 string           `json:"from"`
	To                   string           `json:"to"`
	DataVersion          int64            `json:"data_version"`
}
type Wallet struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	BalanceVND     int64  `json:"balance_vnd"`
	IncludeInTotal bool   `json:"include_in_total"`
}
type RecentTransaction struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	AmountVND  int64     `json:"amount_vnd"`
	Note       string    `json:"note"`
	OccurredAt time.Time `json:"occurred_at"`
}
