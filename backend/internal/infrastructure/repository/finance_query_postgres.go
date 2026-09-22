package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrFinanceUnsupported = errors.New("finance query is unsupported")
	ErrFinanceScope       = errors.New("finance query scope is invalid")
	ErrFinanceCursor      = errors.New("finance cursor is invalid")
)

type financeCursor struct {
	Version    int       `json:"v"`
	OccurredAt time.Time `json:"occurred_at"`
	ID         string    `json:"id"`
	ScopeHash  string    `json:"scope_hash"`
}

func encodeFinanceCursor(occurredAt time.Time, id, scopeHash string) (string, error) {
	if id == "" || scopeHash == "" {
		return "", ErrFinanceCursor
	}
	payload, err := json.Marshal(financeCursor{Version: 1, OccurredAt: occurredAt.UTC(), ID: id, ScopeHash: scopeHash})
	if err != nil {
		return "", ErrFinanceCursor
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeFinanceCursor(value, expectedScope string) (financeCursor, error) {
	if value == "" || len(value) > 2048 || expectedScope == "" {
		return financeCursor{}, ErrFinanceCursor
	}
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return financeCursor{}, ErrFinanceCursor
	}
	var cursor financeCursor
	if err := json.Unmarshal(payload, &cursor); err != nil || cursor.Version != 1 || cursor.ID == "" || cursor.ScopeHash != expectedScope || cursor.OccurredAt.IsZero() {
		return financeCursor{}, ErrFinanceCursor
	}
	return cursor, nil
}

func financeScopeHash(kind string, filter entity.NormalizedFinanceFilter) string {
	payload, _ := json.Marshal(struct {
		Kind   string                         `json:"kind"`
		Filter entity.NormalizedFinanceFilter `json:"filter"`
	}{Kind: kind, Filter: filter})
	digest := sha256.Sum256(payload)
	return fmt.Sprintf("%x", digest[:])
}

type FinanceQueryPostgresRepository struct{ db *gorm.DB }

func NewFinanceQueryPostgresRepository(db *gorm.DB) contract.FinanceReader {
	return &FinanceQueryPostgresRepository{db: db}
}

func (r *FinanceQueryPostgresRepository) ReadBundle(ctx context.Context, ownerID string, queries []entity.NormalizedQuery) (entity.FactBundle, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || len(queries) == 0 {
		return entity.FactBundle{}, contract.ErrNotFound
	}
	bundle := entity.FactBundle{ID: uuid.NewString(), AsOf: time.Now().UTC(), Results: make([]entity.FinanceResult, 0, len(queries))}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL statement_timeout = '2000ms'").Error; err != nil {
			return err
		}
		var user entity.User
		if err := tx.Where("id = ?", ownerID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return contract.ErrNotFound
			}
			return err
		}
		if strings.TrimSpace(user.Timezone) == "" || user.Timezone == "Local" {
			return ErrFinanceScope
		}
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			return ErrFinanceScope
		}
		for _, query := range queries {
			result, err := r.readQuery(ctx, tx, ownerID, location, bundle.AsOf, query)
			if err != nil {
				return err
			}
			bundle.Results = append(bundle.Results, result)
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return entity.FactBundle{}, err
	}
	return bundle, nil
}

type financeSummaryView struct {
	Income  int64 `json:"income"`
	Expense int64 `json:"expense"`
	Net     int64 `json:"net"`
	Count   int64 `json:"count"`
}

func (r *FinanceQueryPostgresRepository) readQuery(ctx context.Context, tx *gorm.DB, ownerID string, location *time.Location, asOf time.Time, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	result := entity.FinanceResult{QueryKey: query.Key, Status: "ok", Facts: []entity.Fact{}, WarningCodes: []string{}}
	if query.Kind == "compare_spending_periods" {
		return r.readCompare(ctx, tx, ownerID, location, asOf, query)
	}
	if query.Kind != "get_finance_summary" {
		if query.Kind == "search_transactions" {
			return r.readSearch(ctx, tx, ownerID, location, asOf, query)
		}
		switch query.Kind {
		case "get_transaction":
			return r.readTransaction(ctx, tx, ownerID, location, asOf, query)
		case "get_wallet_balances":
			return r.readWalletBalances(ctx, tx, ownerID, location, asOf, query)
		case "get_goal_progress":
			return r.readGoalProgress(ctx, tx, ownerID, location, asOf, query)
		case "get_budget_progress":
			return r.readBudgetProgress(ctx, tx, ownerID, location, asOf, query)
		case "get_jar_progress":
			return r.readJarProgress(ctx, tx, ownerID, location, asOf, query)
		}
		return entity.FinanceResult{QueryKey: query.Key, Status: "unsupported", Facts: []entity.Fact{}, WarningCodes: []string{"tool_unsupported"}}, nil
	}
	filter := query.Filter
	// Summary is a report view by contract; excluded ledger rows remain
	// searchable but never affect income/expense totals.
	filter.ReportScope = entity.FinanceReportIncluded
	if filter.Dates.StartAt.IsZero() || filter.Dates.EndExclusive.IsZero() {
		return entity.FinanceResult{}, entity.ErrFinanceRangeInvalid
	}
	result.Source = entity.SourceScope{Ref: query.Key, Timezone: location.String(), StartAt: &filter.Dates.StartAt, EndExclusive: &filter.Dates.EndExclusive, Filter: filter.FinanceFilter, AsOf: asOf}
	queryDB, err := r.scopedTransactions(ctx, tx, ownerID, filter)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	var aggregate struct {
		Income  int64 `gorm:"column:income"`
		Expense int64 `gorm:"column:expense"`
		Count   int64 `gorm:"column:count"`
	}
	selectSQL := `COALESCE(SUM(CASE WHEN type = ? THEN amount ELSE 0 END), 0) AS income,
        COALESCE(SUM(CASE WHEN type = ? THEN amount ELSE 0 END), 0) AS expense,
        COUNT(*) AS count`
	if err := queryDB.Select(selectSQL, entity.TransactionTypeIncome, entity.TransactionTypeExpense).Scan(&aggregate).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	view := financeSummaryView{Income: aggregate.Income, Expense: aggregate.Expense, Net: aggregate.Income - aggregate.Expense, Count: aggregate.Count}
	encoded, err := json.Marshal(view)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	result.ViewKind = "finance_summary"
	result.View = encoded
	currency := entity.WalletCurrencyVND
	result.Facts = []entity.Fact{
		{ID: query.Key + ".income", Kind: "money", Integer: int64Ptr(view.Income), Currency: &currency, SourceRef: query.Key},
		{ID: query.Key + ".expense", Kind: "money", Integer: int64Ptr(view.Expense), Currency: &currency, SourceRef: query.Key},
		{ID: query.Key + ".net", Kind: "money", Integer: int64Ptr(view.Net), Currency: &currency, SourceRef: query.Key},
	}
	return result, nil
}

type financeComparisonView struct {
	CurrentExpense   int64  `json:"current_expense"`
	PreviousExpense  int64  `json:"previous_expense"`
	Delta            int64  `json:"delta"`
	ChangeBPS        *int64 `json:"change_bps,omitempty"`
	PreviousBaseline string `json:"previous_baseline"`
}

func (r *FinanceQueryPostgresRepository) readCompare(ctx context.Context, tx *gorm.DB, ownerID string, location *time.Location, asOf time.Time, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	if query.CompareRange == nil {
		return entity.FinanceResult{}, entity.ErrFinanceRangeInvalid
	}
	current := query
	current.Kind = "get_finance_summary"
	current.Key = query.Key + ".current"
	previousFilter := query.Filter
	previousDates, err := entity.NormalizeDateRange(*query.CompareRange, location)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	previousFilter.Dates = previousDates
	previous := current
	previous.Key = query.Key + ".previous"
	previous.Filter = previousFilter
	currentResult, err := r.readQuery(ctx, tx, ownerID, location, asOf, current)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	previousResult, err := r.readQuery(ctx, tx, ownerID, location, asOf, previous)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	var currentView, previousView financeSummaryView
	if err := json.Unmarshal(currentResult.View, &currentView); err != nil {
		return entity.FinanceResult{}, err
	}
	if err := json.Unmarshal(previousResult.View, &previousView); err != nil {
		return entity.FinanceResult{}, err
	}
	change, err := entity.PercentBPS(currentView.Expense, previousView.Expense)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	view := financeComparisonView{CurrentExpense: currentView.Expense, PreviousExpense: previousView.Expense, Delta: currentView.Expense - previousView.Expense, ChangeBPS: change, PreviousBaseline: "expense"}
	encoded, err := json.Marshal(view)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return entity.FinanceResult{QueryKey: query.Key, Status: "ok", ViewKind: "spending_comparison", View: encoded, Facts: []entity.Fact{{ID: query.Key + ".delta", Kind: "money", Integer: int64Ptr(view.Delta), SourceRef: query.Key}}, Source: entity.SourceScope{Ref: query.Key, Timezone: location.String(), StartAt: &query.Filter.Dates.StartAt, EndExclusive: &query.Filter.Dates.EndExclusive, Filter: query.Filter.FinanceFilter, AsOf: asOf}}, nil
}

type financeTransactionView struct {
	ID         string    `json:"id"`
	WalletID   string    `json:"wallet_id"`
	CategoryID *string   `json:"category_id,omitempty"`
	Type       string    `json:"type"`
	Amount     int64     `json:"amount"`
	OccurredAt time.Time `json:"occurred_at"`
	Note       *string   `json:"note,omitempty"`
}

type financeSearchView struct {
	Items      []financeTransactionView `json:"items"`
	TotalCount int64                    `json:"total_count"`
	NextCursor string                   `json:"next_cursor,omitempty"`
}

type financeTransactionDetailView struct {
	ID                string    `json:"id"`
	WalletID          string    `json:"wallet_id"`
	WalletName        string    `json:"wallet_name,omitempty"`
	CategoryID        *string   `json:"category_id,omitempty"`
	CategoryName      string    `json:"category_name,omitempty"`
	JarID             *string   `json:"jar_id,omitempty"`
	Type              string    `json:"type"`
	Amount            int64     `json:"amount"`
	OccurredAt        time.Time `json:"occurred_at"`
	Note              *string   `json:"note,omitempty"`
	IncludedInReports bool      `json:"included_in_reports"`
	TransferID        *string   `json:"transfer_id,omitempty"`
}

func (r *FinanceQueryPostgresRepository) readTransaction(ctx context.Context, tx *gorm.DB, ownerID string, location *time.Location, asOf time.Time, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	if len(query.ObjectIDs) != 1 || strings.TrimSpace(query.ObjectIDs[0]) == "" {
		return entity.FinanceResult{}, ErrFinanceScope
	}
	var row entity.Transaction
	if err := tx.WithContext(ctx).Where("id = ? AND owner_id = ?", query.ObjectIDs[0], ownerID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			encoded, marshalErr := json.Marshal(map[string]any{"found": false, "transaction_id": query.ObjectIDs[0]})
			if marshalErr != nil {
				return entity.FinanceResult{}, marshalErr
			}
			return entity.FinanceResult{QueryKey: query.Key, Status: "not_found", ViewKind: "transaction_detail", View: encoded, WarningCodes: []string{"transaction_not_found"}, Source: entity.SourceScope{Ref: query.Key, Timezone: location.String(), AsOf: asOf}}, nil
		}
		return entity.FinanceResult{}, err
	}
	view := financeTransactionDetailView{ID: row.ID, WalletID: row.WalletID, CategoryID: row.CategoryID, JarID: row.JarID, Type: row.Type, Amount: row.Amount, OccurredAt: row.OccurredAt, Note: row.Note, IncludedInReports: row.IncludedInReports, TransferID: row.TransferID}
	var wallet entity.Wallet
	if err := tx.WithContext(ctx).Where("id = ? AND owner_id = ?", row.WalletID, ownerID).First(&wallet).Error; err == nil {
		view.WalletName = wallet.Name
	}
	if row.CategoryID != nil {
		var category entity.Category
		if err := tx.WithContext(ctx).Where("id = ? AND (owner_id = ? OR owner_id IS NULL)", *row.CategoryID, ownerID).First(&category).Error; err == nil {
			view.CategoryName = category.Name
		}
	}
	encoded, err := json.Marshal(view)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return entity.FinanceResult{QueryKey: query.Key, Status: "ok", ViewKind: "transaction_detail", View: encoded, Facts: []entity.Fact{{ID: query.Key + ".amount", Kind: "money", Integer: int64Ptr(row.Amount), Currency: stringPtr(wallet.Currency), SourceRef: query.Key}}, Source: entity.SourceScope{Ref: query.Key, Timezone: location.String(), AsOf: asOf}}, nil
}

func (r *FinanceQueryPostgresRepository) readSearch(ctx context.Context, tx *gorm.DB, ownerID string, location *time.Location, asOf time.Time, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	filter := query.Filter
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit < 1 || filter.Limit > 50 {
		return entity.FinanceResult{}, entity.ErrFinanceRangeInvalid
	}
	queryDB, err := r.scopedTransactions(ctx, tx, ownerID, filter)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	var total int64
	if err := queryDB.Count(&total).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	scopeHash := financeScopeHash(query.Kind, filter)
	if query.Cursor != "" {
		cursor, err := decodeFinanceCursor(query.Cursor, scopeHash)
		if err != nil {
			return entity.FinanceResult{}, err
		}
		queryDB = queryDB.Where("(transactions.occurred_at, transactions.id) < (?, ?)", cursor.OccurredAt, cursor.ID)
	}
	var rows []entity.Transaction
	if err := queryDB.Order("transactions.occurred_at DESC, transactions.id DESC").Limit(filter.Limit + 1).Find(&rows).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	view := financeSearchView{Items: make([]financeTransactionView, 0, len(rows)), TotalCount: total}
	if len(rows) > filter.Limit {
		last := rows[filter.Limit-1]
		cursor, err := encodeFinanceCursor(last.OccurredAt, last.ID, scopeHash)
		if err != nil {
			return entity.FinanceResult{}, err
		}
		view.NextCursor = cursor
		rows = rows[:filter.Limit]
	}
	for _, row := range rows {
		view.Items = append(view.Items, financeTransactionView{ID: row.ID, WalletID: row.WalletID, CategoryID: row.CategoryID, Type: row.Type, Amount: row.Amount, OccurredAt: row.OccurredAt, Note: row.Note})
	}
	encoded, err := json.Marshal(view)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	result := entity.FinanceResult{
		QueryKey: query.Key,
		Status:   "ok",
		Facts:    []entity.Fact{{ID: query.Key + ".total_count", Kind: "count", Integer: int64Ptr(total), SourceRef: query.Key}},
		ViewKind: "transaction_search",
		View:     encoded,
		Source:   entity.SourceScope{Ref: query.Key, Timezone: location.String(), StartAt: &filter.Dates.StartAt, EndExclusive: &filter.Dates.EndExclusive, Filter: filter.FinanceFilter, AsOf: asOf},
	}
	return result, nil
}

type financeWalletView struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Currency        string `json:"currency"`
	CurrentBalance  int64  `json:"current_balance"`
	OpeningBalance  int64  `json:"opening_balance"`
	TargetAmount    *int64 `json:"target_amount,omitempty"`
	TargetRemaining *int64 `json:"target_remaining,omitempty"`
}

func (r *FinanceQueryPostgresRepository) readWalletBalances(ctx context.Context, tx *gorm.DB, ownerID string, location *time.Location, asOf time.Time, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	var wallets []entity.Wallet
	walletQuery := tx.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at ASC, id ASC")
	if len(query.ObjectIDs) > 0 {
		walletQuery = walletQuery.Where("id IN ?", query.ObjectIDs)
	}
	if err := walletQuery.Find(&wallets).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	var rows []entity.Transaction
	if err := tx.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&rows).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	balances := make(map[string]int64, len(wallets))
	for _, wallet := range wallets {
		balances[wallet.ID] = wallet.OpeningBalance
	}
	for _, row := range rows {
		if _, ok := balances[row.WalletID]; !ok {
			continue
		}
		if row.Type == entity.TransactionTypeIncome {
			balances[row.WalletID] += row.Amount
		} else if row.Type == entity.TransactionTypeExpense {
			balances[row.WalletID] -= row.Amount
		}
	}
	items := make([]financeWalletView, 0, len(wallets))
	for _, wallet := range wallets {
		item := financeWalletView{ID: wallet.ID, Name: wallet.Name, Type: wallet.Type, Currency: wallet.Currency, CurrentBalance: balances[wallet.ID], OpeningBalance: wallet.OpeningBalance, TargetAmount: wallet.TargetAmount}
		if wallet.TargetAmount != nil {
			remaining := *wallet.TargetAmount - item.CurrentBalance
			if remaining < 0 {
				remaining = 0
			}
			item.TargetRemaining = &remaining
		}
		items = append(items, item)
	}
	view := map[string]any{"items": items, "scope": "selected"}
	encoded, err := json.Marshal(view)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return entity.FinanceResult{QueryKey: query.Key, Status: "ok", ViewKind: "wallet_balances", View: encoded, Facts: []entity.Fact{{ID: query.Key + ".count", Kind: "count", Integer: int64Ptr(int64(len(items))), SourceRef: query.Key}}, Source: entity.SourceScope{Ref: query.Key, Timezone: location.String(), Filter: query.Filter.FinanceFilter, AsOf: asOf}}, nil
}

func (r *FinanceQueryPostgresRepository) readGoalProgress(ctx context.Context, tx *gorm.DB, ownerID string, location *time.Location, asOf time.Time, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	goalQuery := tx.WithContext(ctx).Where("owner_id = ? AND type = ?", ownerID, entity.WalletTypeGoal).Order("created_at ASC, id ASC")
	if len(query.ObjectIDs) > 0 {
		goalQuery = goalQuery.Where("id IN ?", query.ObjectIDs)
	}
	var goals []entity.Wallet
	if err := goalQuery.Find(&goals).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	walletResult, err := r.readWalletBalances(ctx, tx, ownerID, location, asOf, query)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	var wallets struct {
		Items []financeWalletView `json:"items"`
	}
	if err := json.Unmarshal(walletResult.View, &wallets); err != nil {
		return entity.FinanceResult{}, err
	}
	goalIDs := make(map[string]struct{}, len(goals))
	for _, goal := range goals {
		goalIDs[goal.ID] = struct{}{}
	}
	items := make([]financeWalletView, 0, len(goals))
	for _, item := range wallets.Items {
		if _, ok := goalIDs[item.ID]; ok {
			items = append(items, item)
		}
	}
	encoded, err := json.Marshal(map[string]any{"items": items})
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return entity.FinanceResult{QueryKey: query.Key, Status: "ok", ViewKind: "goal_progress", View: encoded, Facts: []entity.Fact{{ID: query.Key + ".count", Kind: "count", Integer: int64Ptr(int64(len(items))), SourceRef: query.Key}}, Source: entity.SourceScope{Ref: query.Key, Timezone: location.String(), Filter: query.Filter.FinanceFilter, AsOf: asOf}}, nil
}

func (r *FinanceQueryPostgresRepository) readBudgetProgress(ctx context.Context, tx *gorm.DB, ownerID string, location *time.Location, asOf time.Time, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	budgetQuery := tx.WithContext(ctx).Where("owner_id = ?", ownerID).Order("start_date ASC, id ASC")
	if len(query.ObjectIDs) > 0 {
		budgetQuery = budgetQuery.Where("id IN ?", query.ObjectIDs)
	}
	var budgets []entity.Budget
	if err := budgetQuery.Find(&budgets).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	var transactions []entity.Transaction
	if err := tx.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&transactions).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	var categories []entity.Category
	if err := tx.WithContext(ctx).Where("owner_id = ? OR owner_id IS NULL", ownerID).Find(&categories).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	// Evaluate an active budget against the same snapshot timestamp used by the
	// rest of the bundle. This avoids a budget boundary moving while a bundle
	// is being assembled.
	now := asOf.In(location)
	if query.ActiveOn != "" {
		parsed, err := time.ParseInLocation("2006-01-02", query.ActiveOn, location)
		if err != nil {
			return entity.FinanceResult{}, entity.ErrFinanceDateInvalid
		}
		now = parsed
	}
	summary := entity.CalculateBudgetsInLocation(budgets, transactions, categories, now, location)
	encoded, err := json.Marshal(summary)
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return entity.FinanceResult{QueryKey: query.Key, Status: "ok", ViewKind: "budget_progress", View: encoded, Facts: []entity.Fact{{ID: query.Key + ".spent", Kind: "money", Integer: int64Ptr(summary.Spent), SourceRef: query.Key}}, Source: entity.SourceScope{Ref: query.Key, Timezone: location.String(), Filter: query.Filter.FinanceFilter, AsOf: asOf}}, nil
}

func (r *FinanceQueryPostgresRepository) readJarProgress(ctx context.Context, tx *gorm.DB, ownerID string, location *time.Location, asOf time.Time, query entity.NormalizedQuery) (entity.FinanceResult, error) {
	if query.Month == "" {
		return entity.FinanceResult{}, entity.ErrFinanceDateInvalid
	}
	month, err := entity.ParseMonth(query.Month)
	if err != nil {
		return entity.FinanceResult{}, entity.ErrFinanceDateInvalid
	}
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, location)
	next := start.AddDate(0, 1, 0)
	var configs []entity.JarMonthConfig
	configQuery := tx.WithContext(ctx).Where("owner_id = ? AND month = ? AND active = true", ownerID, entity.CalendarDate{Time: month})
	if len(query.ObjectIDs) > 0 {
		configQuery = configQuery.Where("jar_id IN ?", query.ObjectIDs)
	}
	if err := configQuery.Order("jar_id ASC").Find(&configs).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	var transactions []entity.Transaction
	if err := tx.WithContext(ctx).Where("owner_id = ? AND occurred_at >= ? AND occurred_at < ? AND type = ? AND included_in_reports = true", ownerID, start.UTC(), next.UTC(), entity.TransactionTypeExpense).Find(&transactions).Error; err != nil {
		return entity.FinanceResult{}, err
	}
	spent := make(map[string]int64)
	for _, row := range transactions {
		if row.JarID != nil {
			spent[*row.JarID] += row.Amount
		}
	}
	type jarView struct {
		JarID            string `json:"jar_id"`
		Name             string `json:"name"`
		AllocationAmount *int64 `json:"allocation_amount,omitempty"`
		Spent            int64  `json:"spent"`
		ConfigOrigin     string `json:"config_origin"`
	}
	items := make([]jarView, 0, len(configs))
	for _, config := range configs {
		items = append(items, jarView{JarID: config.JarID, Name: config.Name, AllocationAmount: config.AllocationAmount, Spent: spent[config.JarID], ConfigOrigin: "stored"})
	}
	origin := "stored"
	if len(configs) == 0 {
		origin = "unconfigured"
	}
	encoded, err := json.Marshal(map[string]any{"month": query.Month, "timezone": location.String(), "config_origin": origin, "items": items})
	if err != nil {
		return entity.FinanceResult{}, err
	}
	return entity.FinanceResult{QueryKey: query.Key, Status: "ok", ViewKind: "jar_progress", View: encoded, Facts: []entity.Fact{{ID: query.Key + ".count", Kind: "count", Integer: int64Ptr(int64(len(items))), SourceRef: query.Key}}, Source: entity.SourceScope{Ref: query.Key, Timezone: location.String(), Filter: query.Filter.FinanceFilter, AsOf: asOf}}, nil
}

func (r *FinanceQueryPostgresRepository) scopedTransactions(ctx context.Context, tx *gorm.DB, ownerID string, filter entity.NormalizedFinanceFilter) (*gorm.DB, error) {
	query := tx.WithContext(ctx).Model(&entity.Transaction{}).
		Where("transactions.owner_id = ?", ownerID).
		Where("transactions.occurred_at >= ? AND transactions.occurred_at < ?", filter.Dates.StartAt, filter.Dates.EndExclusive)
	if filter.ReportScope == entity.FinanceReportIncluded {
		query = query.Where("transactions.included_in_reports = true")
	} else if filter.ReportScope == entity.FinanceReportExcluded {
		query = query.Where("transactions.included_in_reports = false")
	}
	if filter.Type != entity.FinanceTypeAll {
		query = query.Where("transactions.type = ?", filter.Type)
	}
	if len(filter.WalletIDs) > 0 {
		query = query.Where("transactions.wallet_id IN ?", filter.WalletIDs)
	}
	if filter.MinAmount != nil {
		query = query.Where("transactions.amount >= ?", *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		query = query.Where("transactions.amount <= ?", *filter.MaxAmount)
	}
	if filter.NoteContains != "" {
		query = query.Where("strpos(lower(coalesce(transactions.note, '')), lower(?)) > 0", filter.NoteContains)
	}
	if filter.ReportScope == entity.FinanceReportIncluded {
		// Internal transfers are not report activity, even if legacy rows were
		// accidentally flagged as included. System transfer categories are also
		// excluded by key to cover older data without a transfer_id.
		query = query.Where("transactions.transfer_id IS NULL")
		query = query.Where("NOT EXISTS (SELECT 1 FROM categories transfer_categories WHERE transfer_categories.id = transactions.category_id AND transfer_categories.system_key IN (?, ?))", "income_transfer_in", "expense_transfer_out")
	}
	if len(filter.CategoryIDs) > 0 {
		ids, err := r.expandCategoryIDs(tx, ownerID, filter.CategoryIDs, filter.IncludeChildren)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return query.Where("1 = 0"), nil
		}
		query = query.Where("transactions.category_id IN ?", ids)
	}
	return query, nil
}

func (r *FinanceQueryPostgresRepository) expandCategoryIDs(tx *gorm.DB, ownerID string, roots []string, includeChildren bool) ([]string, error) {
	if !includeChildren {
		var count int64
		if err := tx.Model(&entity.Category{}).Where("id IN ? AND (owner_id = ? OR owner_id IS NULL)", roots, ownerID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count != int64(len(roots)) {
			return nil, ErrFinanceScope
		}
		return roots, nil
	}
	var categories []entity.Category
	if err := tx.Where("owner_id = ? OR owner_id IS NULL", ownerID).Find(&categories).Error; err != nil {
		return nil, err
	}
	children := make(map[string][]string, len(categories))
	known := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		known[category.ID] = struct{}{}
		if category.ParentID != nil {
			children[*category.ParentID] = append(children[*category.ParentID], category.ID)
		}
	}
	result := make([]string, 0, len(roots))
	seen := make(map[string]struct{}, len(roots))
	var visit func(string)
	visit = func(id string) {
		if _, ok := known[id]; !ok {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		result = append(result, id)
		for _, child := range children[id] {
			visit(child)
		}
	}
	for _, root := range roots {
		visit(root)
	}
	if len(result) == 0 {
		return nil, ErrFinanceScope
	}
	return result, nil
}

func int64Ptr(value int64) *int64 { return &value }

func stringPtr(value string) *string { return &value }
