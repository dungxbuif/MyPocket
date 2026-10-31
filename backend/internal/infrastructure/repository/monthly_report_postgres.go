package repository

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MonthlyReportPostgresRepository struct{ db *gorm.DB }

func NewMonthlyReportPostgresRepository(db *gorm.DB) *MonthlyReportPostgresRepository {
	return &MonthlyReportPostgresRepository{db: db}
}

type reportQueryScope struct {
	Scope             entity.MonthlyReportScope
	Location          *time.Location
	Month             time.Time
	MonthStart        time.Time
	MonthNext         time.Time
	Current           entity.ReportRange
	Comparison        entity.ReportRange
	PreviousStart     time.Time
	PreviousNext      time.Time
	CommonDays        int
	IsCurrentMonth    bool
	WalletCurrency    string
	AccountLocalMonth time.Time
}

func (r *MonthlyReportPostgresRepository) ReadMonthly(ctx context.Context, ownerID string, scope entity.MonthlyReportScope, now time.Time) (entity.MonthlyReport, error) {
	var result entity.MonthlyReport
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		revision, err := r.revision(ctx, tx, ownerID, false)
		if err != nil {
			return err
		}
		result, err = r.readMonthlyInTx(ctx, tx, ownerID, scope, now, revision)
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return result, err
}

func (r *MonthlyReportPostgresRepository) readMonthlyInTx(ctx context.Context, tx *gorm.DB, ownerID string, scope entity.MonthlyReportScope, now time.Time, revision int64) (entity.MonthlyReport, error) {
	queryScope, err := r.resolveScope(ctx, tx, ownerID, scope, now)
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	current, err := r.aggregateTotals(ctx, tx, ownerID, queryScope.Scope.WalletID, queryScope.Current.From, queryScope.Current.EndExclusive)
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	comparisonCurrent, err := r.aggregateTotals(ctx, tx, ownerID, queryScope.Scope.WalletID, queryScope.Comparison.From, queryScope.Comparison.EndExclusive)
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	comparisonPrevious, err := r.aggregateTotals(ctx, tx, ownerID, queryScope.Scope.WalletID, queryScope.PreviousStart, queryScope.PreviousStart.Add(queryScope.Comparison.EndExclusive.Sub(queryScope.Comparison.From)))
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	comparison, err := buildReportComparison(comparisonCurrent, comparisonPrevious, queryScope.CommonDays)
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	series, err := r.series(ctx, tx, ownerID, queryScope)
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	categories, contributors, err := r.categoriesAndContributors(ctx, tx, ownerID, queryScope)
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	budgets, err := r.budgetPreviews(ctx, tx, ownerID, queryScope, current.Income)
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	jars, err := r.jarPreviews(ctx, tx, ownerID, queryScope, current.Income)
	if err != nil {
		return entity.MonthlyReport{}, err
	}
	quality := make([]string, 0, 3)
	if current.TransactionCount == 0 {
		quality = append(quality, "no_data")
	}
	if !comparison.Comparable {
		quality = append(quality, "comparison_not_comparable")
	}
	if queryScope.Scope.WalletID != "" && len(jars) > 0 {
		quality = append(quality, "jar_scope_account")
	}
	result := entity.MonthlyReport{
		Scope: queryScope.Scope, Timezone: queryScope.Location.String(), Currency: queryScope.WalletCurrency,
		AsOf: now.UTC(), DataRevision: revision, RuleVersion: entity.MonthlyReportRuleVersion,
		CurrentRange: queryScope.Current, ComparisonRange: queryScope.Comparison,
		Totals: current, Comparison: comparison, Series: series, Categories: categories,
		Contributors: contributors, Budgets: budgets, Jars: jars, QualityFlags: quality,
		Drilldown: entity.ReportDrilldown{From: queryScope.Current.From, EndExclusive: queryScope.Current.EndExclusive, WalletID: queryScope.Scope.WalletID},
	}
	result.Insights = entity.BuildReportInsights(result)
	return result, nil
}

func (r *MonthlyReportPostgresRepository) resolveScope(ctx context.Context, tx *gorm.DB, ownerID string, scope entity.MonthlyReportScope, now time.Time) (reportQueryScope, error) {
	var user entity.User
	if err := tx.WithContext(ctx).Where("id = ?", ownerID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return reportQueryScope{}, contract.ErrNotFound
		}
		return reportQueryScope{}, err
	}
	location, err := time.LoadLocation(strings.TrimSpace(user.Timezone))
	if err != nil {
		return reportQueryScope{}, entity.ErrMonthlyReportScopeInvalid
	}
	monthValue := strings.TrimSpace(scope.Month)
	localNow := now.In(location)
	if monthValue == "" {
		monthValue = localNow.Format("2006-01")
	}
	month, err := entity.ParseMonth(monthValue)
	if err != nil {
		return reportQueryScope{}, entity.ErrMonthlyReportScopeInvalid
	}
	monthYear, monthNumber, _ := month.Date()
	if monthYear > localNow.Year() || (monthYear == localNow.Year() && monthNumber > localNow.Month()) {
		return reportQueryScope{}, entity.ErrMonthlyReportFuture
	}
	granularity := entity.NormalizeMonthlyReportGranularity(scope.Granularity)
	if granularity == "" {
		return reportQueryScope{}, entity.ErrMonthlyReportScopeInvalid
	}
	walletID := strings.TrimSpace(scope.WalletID)
	currency := entity.WalletCurrencyVND
	if walletID != "" {
		var wallet entity.Wallet
		if err := tx.WithContext(ctx).Where("id = ? AND owner_id = ?", walletID, ownerID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return reportQueryScope{}, contract.ErrNotFound
			}
			return reportQueryScope{}, err
		}
		currency = wallet.Currency
	}
	monthStart, monthNext, err := entity.MonthRangeUTC(month, location)
	if err != nil {
		return reportQueryScope{}, entity.ErrMonthlyReportScopeInvalid
	}
	monthStartLocal := monthStart.In(location)
	monthNextLocal := monthNext.In(location)
	isCurrent := monthYear == localNow.Year() && monthNumber == localNow.Month()
	currentEnd := monthNext
	currentDays := calendarDayCount(monthStartLocal, monthNextLocal)
	if isCurrent {
		currentEnd = now.UTC()
		currentDays = calendarDayCount(monthStartLocal, time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location)) + 1
	}
	previousMonth := month.AddDate(0, -1, 0)
	previousStart, previousNext, err := entity.MonthRangeUTC(previousMonth, location)
	if err != nil {
		return reportQueryScope{}, entity.ErrMonthlyReportScopeInvalid
	}
	commonDays := currentDays
	if isCurrent {
		commonDays--
	}
	if commonDays < 0 {
		commonDays = 0
	}
	previousDays := calendarDayCount(previousStart.In(location), previousNext.In(location))
	if previousDays < commonDays {
		commonDays = previousDays
	}
	comparisonFrom := monthStart
	comparisonEnd := monthStart
	if commonDays > 0 {
		comparisonEnd = localDateAdd(monthStartLocal, commonDays).UTC()
	}
	comparison := entity.ReportRange{From: comparisonFrom, EndExclusive: comparisonEnd, Days: commonDays, IsPartial: isCurrent}
	current := entity.ReportRange{From: monthStart, EndExclusive: currentEnd, Days: currentDays, IsPartial: isCurrent}
	return reportQueryScope{
		Scope: entity.MonthlyReportScope{Month: monthValue, WalletID: walletID, Granularity: granularity}, Location: location,
		Month: month, MonthStart: monthStart, MonthNext: monthNext, Current: current, Comparison: comparison,
		PreviousStart: previousStart, PreviousNext: previousNext, CommonDays: commonDays, IsCurrentMonth: isCurrent,
		WalletCurrency: currency, AccountLocalMonth: monthStartLocal,
	}, nil
}

func (r *MonthlyReportPostgresRepository) revision(ctx context.Context, tx *gorm.DB, ownerID string, lock bool) (int64, error) {
	query := tx.WithContext(ctx).Table("report_revisions").Select("revision").Where("owner_id = ?", ownerID)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var row struct{ Revision int64 }
	if err := query.Scan(&row).Error; err != nil {
		return 0, err
	}
	if row.Revision < 1 {
		return 0, contract.ErrNotFound
	}
	return row.Revision, nil
}

type totalsRow struct {
	Income       sql.NullString
	Expense      sql.NullString
	Count        int64
	IncomeCount  int64
	ExpenseCount int64
}

func (r *MonthlyReportPostgresRepository) aggregateTotals(ctx context.Context, tx *gorm.DB, ownerID, walletID string, from, end time.Time) (entity.ReportTotals, error) {
	query := tx.WithContext(ctx).Table("transactions AS t").Where("t.owner_id = ? AND t.occurred_at >= ? AND t.occurred_at < ?", ownerID, from, end)
	query = applyCanonicalReportEligibility(query, "t")
	if walletID != "" {
		query = query.Where("t.wallet_id = ?", walletID)
	}
	var row totalsRow
	err := query.Select("COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END)::text, '0') AS income, COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END)::text, '0') AS expense, COUNT(*) AS count, COUNT(*) FILTER (WHERE t.type = 'income') AS income_count, COUNT(*) FILTER (WHERE t.type = 'expense') AS expense_count").Scan(&row).Error
	if err != nil {
		return entity.ReportTotals{}, err
	}
	income, err := parseReportMoney(row.Income.String)
	if err != nil {
		return entity.ReportTotals{}, err
	}
	expense, err := parseReportMoney(row.Expense.String)
	if err != nil {
		return entity.ReportTotals{}, err
	}
	net, err := entity.CheckedMoneyDifference(income, expense)
	if err != nil {
		return entity.ReportTotals{}, err
	}
	return entity.ReportTotals{Income: income, Expense: expense, Net: net, TransactionCount: row.Count, IncomeCount: row.IncomeCount, ExpenseCount: row.ExpenseCount}, nil
}

type dailyReportRow struct {
	Day     string
	Income  sql.NullString
	Expense sql.NullString
	Count   int64
}

func (r *MonthlyReportPostgresRepository) series(ctx context.Context, tx *gorm.DB, ownerID string, scope reportQueryScope) ([]entity.ReportSeriesPoint, error) {
	if scope.Current.EndExclusive.Before(scope.Current.From) || scope.Current.EndExclusive.Equal(scope.Current.From) {
		return []entity.ReportSeriesPoint{}, nil
	}
	query := tx.WithContext(ctx).Table("transactions AS t").Where("t.owner_id = ? AND t.occurred_at >= ? AND t.occurred_at < ?", ownerID, scope.Current.From, scope.Current.EndExclusive)
	query = applyCanonicalReportEligibility(query, "t")
	if scope.Scope.WalletID != "" {
		query = query.Where("t.wallet_id = ?", scope.Scope.WalletID)
	}
	rows := make([]dailyReportRow, 0, scope.Current.Days)
	err := query.Select("to_char(t.occurred_at AT TIME ZONE ?, 'YYYY-MM-DD') AS day, COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END)::text, '0') AS income, COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END)::text, '0') AS expense, COUNT(*) AS count", scope.Location.String()).Group("day").Order("day ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	daily := make(map[string]dailyReportRow, len(rows))
	for _, row := range rows {
		daily[row.Day] = row
	}
	points := make([]entity.ReportSeriesPoint, 0, scope.Current.Days)
	startLocal := scope.Current.From.In(scope.Location)
	for index := 0; index < scope.Current.Days; {
		bucketStart := time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day()+index, 0, 0, 0, 0, scope.Location)
		bucketEnd := bucketStart.AddDate(0, 0, 1)
		label := bucketStart.Format("2006-01-02")
		if scope.Scope.Granularity == "week" {
			monday := bucketStart.AddDate(0, 0, -((int(bucketStart.Weekday()) + 6) % 7))
			bucketStart = monday
			bucketEnd = monday.AddDate(0, 0, 7)
			label = monday.Format("2006-01-02")
		}
		if bucketStart.Before(scope.Current.From.In(scope.Location)) {
			bucketStart = scope.Current.From.In(scope.Location)
		}
		if bucketEnd.After(scope.Current.EndExclusive.In(scope.Location)) {
			bucketEnd = scope.Current.EndExclusive.In(scope.Location)
		}
		income, expense, count := int64(0), int64(0), int64(0)
		for day := bucketStart; day.Before(bucketEnd); day = day.AddDate(0, 0, 1) {
			row := daily[day.Format("2006-01-02")]
			parsedIncome, err := parseReportMoney(row.Income.String)
			if err != nil {
				return nil, err
			}
			parsedExpense, err := parseReportMoney(row.Expense.String)
			if err != nil {
				return nil, err
			}
			income, err = entity.CheckedMoneyAdd(income, parsedIncome)
			if err != nil {
				return nil, err
			}
			expense, err = entity.CheckedMoneyAdd(expense, parsedExpense)
			if err != nil {
				return nil, err
			}
			count += row.Count
		}
		net, err := entity.CheckedMoneyDifference(income, expense)
		if err != nil {
			return nil, err
		}
		points = append(points, entity.ReportSeriesPoint{Label: label, From: bucketStart.UTC(), EndExclusive: bucketEnd.UTC(), Income: income, Expense: expense, Net: net, TransactionCount: count})
		if scope.Scope.Granularity == "day" {
			index++
		} else {
			next := bucketEnd.In(scope.Location)
			index = calendarDayCount(startLocal, next)
		}
	}
	return points, nil
}

type categoryReportRow struct {
	CategoryID sql.NullString
	Amount     sql.NullString
	Count      int64
}

func (r *MonthlyReportPostgresRepository) categoriesAndContributors(ctx context.Context, tx *gorm.DB, ownerID string, scope reportQueryScope) ([]entity.ReportCategoryBreakdown, []entity.ReportContributor, error) {
	current, err := r.categoryAggregates(ctx, tx, ownerID, scope.Scope.WalletID, scope.Current.From, scope.Current.EndExclusive)
	if err != nil {
		return nil, nil, err
	}
	previous, err := r.categoryAggregates(ctx, tx, ownerID, scope.Scope.WalletID, scope.PreviousStart, scope.PreviousStart.Add(scope.Comparison.EndExclusive.Sub(scope.Comparison.From)))
	if err != nil {
		return nil, nil, err
	}
	var categories []entity.Category
	if err := tx.WithContext(ctx).Where("owner_id = ? OR owner_id IS NULL", ownerID).Find(&categories).Error; err != nil {
		return nil, nil, err
	}
	byID := make(map[string]entity.Category, len(categories))
	for _, category := range categories {
		byID[category.ID] = category
	}
	type rolled struct {
		id, name                                       string
		current, previous, currentCount, previousCount int64
	}
	rolledRows := map[string]*rolled{}
	roll := func(row categoryReportRow, target map[string]*rolled, currentSide bool) error {
		id := ""
		if row.CategoryID.Valid {
			id = row.CategoryID.String
		}
		rootID, name := reportCategoryRoot(id, byID)
		key := rootID + "|" + name
		item := target[key]
		if item == nil {
			item = &rolled{id: rootID, name: name}
			target[key] = item
		}
		amount, err := parseReportMoney(row.Amount.String)
		if err != nil {
			return err
		}
		if currentSide {
			item.current, err = entity.CheckedMoneyAdd(item.current, amount)
			item.currentCount += row.Count
		} else {
			item.previous, err = entity.CheckedMoneyAdd(item.previous, amount)
			item.previousCount += row.Count
		}
		return err
	}
	for _, row := range current {
		if err := roll(row, rolledRows, true); err != nil {
			return nil, nil, err
		}
	}
	for _, row := range previous {
		if err := roll(row, rolledRows, false); err != nil {
			return nil, nil, err
		}
	}
	items := make([]rolled, 0, len(rolledRows))
	for _, item := range rolledRows {
		items = append(items, *item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].current != items[j].current {
			return items[i].current > items[j].current
		}
		return items[i].id < items[j].id
	})
	breakdown := make([]entity.ReportCategoryBreakdown, 0, minInt(6, len(items)))
	for index, item := range items {
		if index >= 5 {
			break
		}
		breakdown = append(breakdown, entity.ReportCategoryBreakdown{CategoryID: categoryPointer(item.id), Name: item.name, Amount: item.current, TransactionCount: item.currentCount})
	}
	if len(items) > 5 {
		otherAmount, otherCount := int64(0), int64(0)
		for _, item := range items[5:] {
			otherAmount, err = entity.CheckedMoneyAdd(otherAmount, item.current)
			if err != nil {
				return nil, nil, err
			}
			otherCount += item.currentCount
		}
		breakdown = append(breakdown, entity.ReportCategoryBreakdown{Name: "Khác", Amount: otherAmount, TransactionCount: otherCount})
	}
	contributors := make([]entity.ReportContributor, 0, len(items))
	for _, item := range items {
		delta, err := entity.CheckedMoneyDifference(item.current, item.previous)
		if err != nil {
			return nil, nil, err
		}
		change, err := entity.CheckedMoneyBPSChange(item.current, item.previous)
		if err != nil {
			return nil, nil, err
		}
		contributors = append(contributors, entity.ReportContributor{CategoryID: categoryPointer(item.id), Name: item.name, CurrentAmount: item.current, PreviousAmount: item.previous, Delta: delta, CurrentCount: item.currentCount, PreviousCount: item.previousCount, ChangeBPS: change, FactRef: "category:" + item.id})
	}
	sort.SliceStable(contributors, func(i, j int) bool {
		left, right := absReport(contributors[i].Delta), absReport(contributors[j].Delta)
		if left != right {
			return left > right
		}
		return contributors[i].FactRef < contributors[j].FactRef
	})
	if len(contributors) > 10 {
		contributors = contributors[:10]
	}
	return breakdown, contributors, nil
}

func (r *MonthlyReportPostgresRepository) categoryAggregates(ctx context.Context, tx *gorm.DB, ownerID, walletID string, from, end time.Time) ([]categoryReportRow, error) {
	query := tx.WithContext(ctx).Table("transactions AS t").Where("t.owner_id = ? AND t.occurred_at >= ? AND t.occurred_at < ? AND t.type = ?", ownerID, from, end, entity.TransactionTypeExpense)
	query = applyCanonicalReportEligibility(query, "t")
	if walletID != "" {
		query = query.Where("t.wallet_id = ?", walletID)
	}
	rows := make([]categoryReportRow, 0, 16)
	err := query.Select("t.category_id, COALESCE(SUM(t.amount)::text, '0') AS amount, COUNT(*) AS count").Group("t.category_id").Scan(&rows).Error
	return rows, err
}

func (r *MonthlyReportPostgresRepository) budgetPreviews(ctx context.Context, tx *gorm.DB, ownerID string, scope reportQueryScope, income int64) ([]entity.ReportBudgetPreview, error) {
	var budgets []entity.Budget
	monthDate := time.Date(scope.Month.Year(), scope.Month.Month(), 1, 0, 0, 0, 0, time.UTC)
	if err := tx.WithContext(ctx).Where("owner_id = ? AND start_date <= ? AND end_date >= ?", ownerID, monthDate, monthDate).Order("id ASC").Limit(20).Find(&budgets).Error; err != nil {
		return nil, err
	}
	result := make([]entity.ReportBudgetPreview, 0, len(budgets))
	for _, budget := range budgets {
		if budget.WalletID != nil && scope.Scope.WalletID != "" && *budget.WalletID != scope.Scope.WalletID {
			continue
		}
		query := tx.WithContext(ctx).Table("transactions AS t").Where("t.owner_id = ? AND t.occurred_at >= ? AND t.occurred_at < ? AND t.type = ?", ownerID, scope.Current.From, scope.Current.EndExclusive, entity.TransactionTypeExpense)
		query = applyCanonicalReportEligibility(query, "t")
		if scope.Scope.WalletID != "" {
			query = query.Where("t.wallet_id = ?", scope.Scope.WalletID)
		}
		if budget.WalletID != nil {
			query = query.Where("t.wallet_id = ?", *budget.WalletID)
		}
		if budget.CategoryID != nil {
			ids := categoryDescendants(ctx, tx, ownerID, *budget.CategoryID)
			if len(ids) == 0 {
				ids = []string{*budget.CategoryID}
			}
			query = query.Where("t.category_id IN ?", ids)
		}
		var row struct{ Amount sql.NullString }
		if err := query.Select("COALESCE(SUM(t.amount)::text, '0') AS amount").Scan(&row).Error; err != nil {
			return nil, err
		}
		spent, err := parseReportMoney(row.Amount.String)
		if err != nil {
			return nil, err
		}
		remaining, err := entity.CheckedMoneyDifference(budget.LimitAmount, spent)
		if err != nil {
			return nil, err
		}
		result = append(result, entity.ReportBudgetPreview{ID: budget.ID, Name: budget.Name, LimitAmount: budget.LimitAmount, Spent: spent, Remaining: remaining, OverLimit: spent > budget.LimitAmount, WalletID: budget.WalletID, CategoryID: budget.CategoryID, FactRef: "budget:" + budget.ID})
	}
	return result, nil
}

func (r *MonthlyReportPostgresRepository) jarPreviews(ctx context.Context, tx *gorm.DB, ownerID string, scope reportQueryScope, income int64) ([]entity.ReportJarPreview, error) {
	monthDate := time.Date(scope.Month.Year(), scope.Month.Month(), 1, 0, 0, 0, 0, time.UTC)
	var configs []entity.JarMonthConfig
	if err := tx.WithContext(ctx).Where("owner_id = ? AND month = ? AND active = TRUE", ownerID, monthDate).Order("jar_id ASC").Limit(20).Find(&configs).Error; err != nil {
		return nil, err
	}
	result := make([]entity.ReportJarPreview, 0, len(configs))
	for _, config := range configs {
		query := tx.WithContext(ctx).Table("transactions AS t").Where("t.owner_id = ? AND t.occurred_at >= ? AND t.occurred_at < ? AND t.type = ? AND t.jar_id = ?", ownerID, scope.Current.From, scope.Current.EndExclusive, entity.TransactionTypeExpense, config.JarID)
		query = applyCanonicalReportEligibility(query, "t")
		var row struct{ Amount sql.NullString }
		if err := query.Select("COALESCE(SUM(t.amount)::text, '0') AS amount").Scan(&row).Error; err != nil {
			return nil, err
		}
		spent, err := parseReportMoney(row.Amount.String)
		if err != nil {
			return nil, err
		}
		var allocation *int64
		switch config.AllocationMode {
		case entity.JarAllocationFixed:
			if config.AllocationAmount != nil {
				value := *config.AllocationAmount
				allocation = &value
			}
		case entity.JarAllocationPercent:
			if config.AllocationPercentBPS != nil {
				value, err := entity.CheckedMoneyPercentOfBPS(income, *config.AllocationPercentBPS)
				if err != nil {
					return nil, err
				}
				allocation = &value
			}
		}
		result = append(result, entity.ReportJarPreview{ID: config.JarID, Name: config.Name, AllocationMode: config.AllocationMode, AllocationAmount: config.AllocationAmount, CalculatedAllocation: allocation, Spent: spent, Active: config.Active, Scope: "account", FactRef: "jar:" + config.JarID})
	}
	return result, nil
}

func categoryDescendants(ctx context.Context, tx *gorm.DB, ownerID, root string) []string {
	var categories []entity.Category
	if err := tx.WithContext(ctx).Where("owner_id = ? OR owner_id IS NULL", ownerID).Find(&categories).Error; err != nil {
		return nil
	}
	children := map[string][]string{}
	for _, category := range categories {
		if category.ParentID != nil {
			children[*category.ParentID] = append(children[*category.ParentID], category.ID)
		}
	}
	result := []string{root}
	for index := 0; index < len(result); index++ {
		result = append(result, children[result[index]]...)
	}
	return result
}

func reportCategoryRoot(id string, categories map[string]entity.Category) (string, string) {
	if id == "" {
		return "", "Khoản chi chưa phân nhóm"
	}
	current := id
	seen := map[string]struct{}{}
	for {
		category, ok := categories[current]
		if !ok {
			return current, "Khoản chi chưa phân nhóm"
		}
		if category.ParentID == nil || *category.ParentID == "" {
			return current, category.Name
		}
		if _, ok := seen[current]; ok {
			return current, category.Name
		}
		seen[current] = struct{}{}
		current = *category.ParentID
	}
}

func buildReportComparison(current, previous entity.ReportTotals, commonDays int) (entity.ReportComparison, error) {
	incomeDelta, err := entity.CheckedMoneyDifference(current.Income, previous.Income)
	if err != nil {
		return entity.ReportComparison{}, err
	}
	expenseDelta, err := entity.CheckedMoneyDifference(current.Expense, previous.Expense)
	if err != nil {
		return entity.ReportComparison{}, err
	}
	netDelta, err := entity.CheckedMoneyDifference(current.Net, previous.Net)
	if err != nil {
		return entity.ReportComparison{}, err
	}
	comparison := entity.ReportComparison{Comparable: commonDays >= 7 && current.TransactionCount >= 5 && previous.TransactionCount >= 5, CommonDays: commonDays, Current: current, Previous: previous, IncomeDelta: incomeDelta, ExpenseDelta: expenseDelta, NetDelta: netDelta}
	if change, err := entity.CheckedMoneyBPSChange(current.Expense, previous.Expense); err != nil {
		return entity.ReportComparison{}, err
	} else {
		comparison.ExpenseChangeBPS = change
	}
	if change, err := entity.CheckedMoneyBPSChange(current.Income, previous.Income); err != nil {
		return entity.ReportComparison{}, err
	} else {
		comparison.IncomeChangeBPS = change
	}
	return comparison, nil
}

func parseReportMoney(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parsed, ok := new(big.Int).SetString(value, 10)
	if !ok || parsed.Sign() < 0 || !parsed.IsInt64() || parsed.Int64() > entity.MaxSafeMoneyAmount {
		return 0, entity.ErrMonthlyReportOverflow
	}
	return parsed.Int64(), nil
}

func calendarDayCount(start, end time.Time) int {
	startUTC := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	endUTC := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)
	days := int(endUTC.Sub(startUTC).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func localDateAdd(start time.Time, days int) time.Time { return start.AddDate(0, 0, days) }

func categoryPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
func absReport(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
