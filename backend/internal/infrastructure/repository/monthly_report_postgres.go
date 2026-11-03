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

func (r *MonthlyReportPostgresRepository) ReadMonthlyTransactions(ctx context.Context, ownerID string, scope entity.MonthlyReportScope, categoryID, kind string, limit, offset int, now time.Time) (entity.SavingsTransactionPage, error) {
	drilldown := entity.ReportDrilldown{Type: kind}
	if strings.TrimSpace(categoryID) != "" {
		drilldown.CategoryBucket = "legacy_exact"
		drilldown.CategoryIDs = []string{strings.TrimSpace(categoryID)}
	}
	return r.ReadMonthlyTransactionsByDrilldown(ctx, ownerID, scope, drilldown, limit, offset, now)
}

func (r *MonthlyReportPostgresRepository) ReadMonthlyTransactionsByDrilldown(ctx context.Context, ownerID string, scope entity.MonthlyReportScope, drilldown entity.ReportDrilldown, limit, offset int, now time.Time) (entity.SavingsTransactionPage, error) {
	kind := strings.TrimSpace(drilldown.Type)
	if limit < 1 || limit > 100 || offset < 0 || (kind != "" && kind != "all" && kind != entity.TransactionTypeIncome && kind != entity.TransactionTypeExpense) {
		return entity.SavingsTransactionPage{}, entity.ErrMonthlyReportScopeInvalid
	}
	queryScope, err := r.resolveScope(ctx, r.db, ownerID, scope, now)
	if err != nil {
		return entity.SavingsTransactionPage{}, err
	}
	from, end, walletID := queryScope.Current.From, queryScope.Current.EndExclusive, queryScope.Scope.WalletID
	// Targets come from an owned immutable snapshot. Their range may be a
	// matched-day comparison or a budget period crossing the month boundary.
	if !drilldown.From.IsZero() || !drilldown.EndExclusive.IsZero() {
		if drilldown.From.IsZero() || !drilldown.EndExclusive.After(drilldown.From) || drilldown.EndExclusive.After(now) {
			return entity.SavingsTransactionPage{}, entity.ErrMonthlyReportScopeInvalid
		}
		from, end = drilldown.From, drilldown.EndExclusive
	}
	if drilldown.WalletID != "" {
		if walletID != "" && walletID != drilldown.WalletID {
			return entity.SavingsTransactionPage{}, entity.ErrMonthlyReportScopeInvalid
		}
		walletID = drilldown.WalletID
	}
	query := r.db.WithContext(ctx).Model(&entity.Transaction{}).
		Where("owner_id = ? AND occurred_at >= ? AND occurred_at < ?", ownerID, from, end)
	query = applyCanonicalReportEligibility(query, "transactions")
	if strings.HasPrefix(drilldown.TargetID, "budget:") {
		query = query.Where("credit_payment_id IS NULL")
	}
	if walletID != "" {
		query = query.Where("wallet_id = ?", walletID)
	}
	switch drilldown.CategoryBucket {
	case "":
	case "legacy_exact":
		if len(drilldown.CategoryIDs) != 1 {
			return entity.SavingsTransactionPage{}, entity.ErrMonthlyReportScopeInvalid
		}
		query = query.Where("category_id = ?", drilldown.CategoryIDs[0])
	case "uncategorized":
		query = query.Where("category_id IS NULL")
	case "root", "other":
		if len(drilldown.CategoryIDs) == 0 {
			return entity.SavingsTransactionPage{}, entity.ErrMonthlyReportScopeInvalid
		}
		query = query.Where("category_id IN ?", drilldown.CategoryIDs)
	default:
		return entity.SavingsTransactionPage{}, entity.ErrMonthlyReportScopeInvalid
	}
	if kind == entity.TransactionTypeIncome || kind == entity.TransactionTypeExpense {
		query = query.Where("type = ?", kind)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return entity.SavingsTransactionPage{}, err
	}
	items := make([]entity.Transaction, 0, limit)
	if err := query.Order("occurred_at DESC, id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return entity.SavingsTransactionPage{}, err
	}
	return entity.SavingsTransactionPage{Items: items, Total: int(total), Limit: limit, Offset: offset, Month: queryScope.Scope.Month, Timezone: queryScope.Location.String(), AsOf: now.UTC()}, nil
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
	comparisonPreviousEnd := localDateAdd(queryScope.PreviousStart.In(queryScope.Location), queryScope.CommonDays).UTC()
	comparisonPrevious, err := r.aggregateTotals(ctx, tx, ownerID, queryScope.Scope.WalletID, queryScope.PreviousStart, comparisonPreviousEnd)
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
	previousRange := entity.ReportRange{From: queryScope.PreviousStart, EndExclusive: queryScope.PreviousStart, Days: queryScope.CommonDays, IsPartial: queryScope.IsCurrentMonth}
	if queryScope.CommonDays > 0 {
		previousRange.EndExclusive = localDateAdd(queryScope.PreviousStart.In(queryScope.Location), queryScope.CommonDays).UTC()
	}
	previousSeries, err := r.seriesForRange(ctx, tx, ownerID, queryScope, previousRange)
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
		Totals: current, Comparison: comparison, Series: series, PreviousSeries: previousSeries, Categories: categories,
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
	return r.seriesForRange(ctx, tx, ownerID, scope, scope.Current)
}

func (r *MonthlyReportPostgresRepository) seriesForRange(ctx context.Context, tx *gorm.DB, ownerID string, scope reportQueryScope, reportRange entity.ReportRange) ([]entity.ReportSeriesPoint, error) {
	if reportRange.EndExclusive.Before(reportRange.From) || reportRange.EndExclusive.Equal(reportRange.From) {
		return []entity.ReportSeriesPoint{}, nil
	}
	query := tx.WithContext(ctx).Table("transactions AS t").Where("t.owner_id = ? AND t.occurred_at >= ? AND t.occurred_at < ?", ownerID, reportRange.From, reportRange.EndExclusive)
	query = applyCanonicalReportEligibility(query, "t")
	if scope.Scope.WalletID != "" {
		query = query.Where("t.wallet_id = ?", scope.Scope.WalletID)
	}
	rows := make([]dailyReportRow, 0, reportRange.Days)
	err := query.Select("to_char(t.occurred_at AT TIME ZONE ?, 'YYYY-MM-DD') AS day, COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END)::text, '0') AS income, COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END)::text, '0') AS expense, COUNT(*) AS count", scope.Location.String()).Group("day").Order("day ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	daily := make(map[string]dailyReportRow, len(rows))
	for _, row := range rows {
		daily[row.Day] = row
	}
	points := make([]entity.ReportSeriesPoint, 0, reportRange.Days)
	startLocal := reportRange.From.In(scope.Location)
	for index := 0; index < reportRange.Days; {
		bucketStart := time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day()+index, 0, 0, 0, 0, scope.Location)
		bucketEnd := bucketStart.AddDate(0, 0, 1)
		label := bucketStart.Format("2006-01-02")
		if scope.Scope.Granularity == "week" {
			monday := bucketStart.AddDate(0, 0, -((int(bucketStart.Weekday()) + 6) % 7))
			bucketStart = monday
			bucketEnd = monday.AddDate(0, 0, 7)
			label = monday.Format("2006-01-02")
		}
		if bucketStart.Before(reportRange.From.In(scope.Location)) {
			bucketStart = reportRange.From.In(scope.Location)
		}
		if bucketEnd.After(reportRange.EndExclusive.In(scope.Location)) {
			bucketEnd = reportRange.EndExclusive.In(scope.Location)
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
		// A current-month range can end partway through a day. Converting that
		// clipped end to a day index would revisit the same final weekly bucket.
		if !bucketEnd.Before(reportRange.EndExclusive) {
			break
		}
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
	fullCurrent, err := r.categoryAggregates(ctx, tx, ownerID, scope.Scope.WalletID, scope.Current.From, scope.Current.EndExclusive)
	if err != nil {
		return nil, nil, err
	}
	comparableCurrent, err := r.categoryAggregates(ctx, tx, ownerID, scope.Scope.WalletID, scope.Comparison.From, scope.Comparison.EndExclusive)
	if err != nil {
		return nil, nil, err
	}
	previousEnd := localDateAdd(scope.PreviousStart.In(scope.Location), scope.CommonDays).UTC()
	previous, err := r.categoryAggregates(ctx, tx, ownerID, scope.Scope.WalletID, scope.PreviousStart, previousEnd)
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
	roll := func(rows []categoryReportRow, target map[string]*rolled, currentSide bool) error {
		for _, row := range rows {
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
			if err != nil {
				return err
			}
		}
		return nil
	}
	fullRows := map[string]*rolled{}
	if err := roll(fullCurrent, fullRows, true); err != nil {
		return nil, nil, err
	}
	comparisonRows := map[string]*rolled{}
	if err := roll(comparableCurrent, comparisonRows, true); err != nil {
		return nil, nil, err
	}
	if err := roll(previous, comparisonRows, false); err != nil {
		return nil, nil, err
	}
	items := make([]rolled, 0, len(fullRows))
	for _, item := range fullRows {
		items = append(items, *item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].current != items[j].current {
			return items[i].current > items[j].current
		}
		return items[i].id < items[j].id
	})
	drilldownFor := func(targetID, bucket string, rootIDs []string) entity.ReportDrilldown {
		categoryIDs := make([]string, 0, len(byID))
		if bucket == "root" || bucket == "other" {
			roots := make(map[string]struct{}, len(rootIDs))
			for _, id := range rootIDs {
				roots[id] = struct{}{}
			}
			for id := range byID {
				rootID, _ := reportCategoryRoot(id, byID)
				if _, ok := roots[rootID]; ok {
					categoryIDs = append(categoryIDs, id)
				}
			}
			sort.Strings(categoryIDs)
		}
		return entity.ReportDrilldown{TargetID: targetID, From: scope.Current.From, EndExclusive: scope.Current.EndExclusive, WalletID: scope.Scope.WalletID, CategoryIDs: categoryIDs, CategoryBucket: bucket, Type: entity.TransactionTypeExpense}
	}
	totalCategoryAmount := int64(0)
	for _, item := range items {
		totalCategoryAmount, err = entity.CheckedMoneyAdd(totalCategoryAmount, item.current)
		if err != nil {
			return nil, nil, err
		}
	}
	categorized := make([]rolled, 0, len(items))
	var uncategorized *rolled
	for index := range items {
		if items[index].id == "" {
			copy := items[index]
			uncategorized = &copy
		} else {
			categorized = append(categorized, items[index])
		}
	}
	breakdown := make([]entity.ReportCategoryBreakdown, 0, minInt(7, len(items)+1))
	for index, item := range categorized {
		if index >= 5 {
			break
		}
		targetID := "category:" + item.id
		breakdown = append(breakdown, entity.ReportCategoryBreakdown{CategoryID: categoryPointer(item.id), Name: item.name, Amount: item.current, TransactionCount: item.currentCount, ShareBPS: reportShareBPS(item.current, totalCategoryAmount), Bucket: "root", TargetID: targetID, Drilldown: drilldownFor(targetID, "root", []string{item.id})})
	}
	if uncategorized != nil {
		breakdown = append(breakdown, entity.ReportCategoryBreakdown{Name: uncategorized.name, Amount: uncategorized.current, TransactionCount: uncategorized.currentCount, ShareBPS: reportShareBPS(uncategorized.current, totalCategoryAmount), Bucket: "uncategorized", TargetID: "category:uncategorized", Drilldown: drilldownFor("category:uncategorized", "uncategorized", nil)})
	}
	if len(categorized) > 5 {
		otherAmount, otherCount := int64(0), int64(0)
		otherRoots := make([]string, 0, len(categorized)-5)
		for _, item := range categorized[5:] {
			otherAmount, err = entity.CheckedMoneyAdd(otherAmount, item.current)
			if err != nil {
				return nil, nil, err
			}
			otherCount += item.currentCount
			otherRoots = append(otherRoots, item.id)
		}
		breakdown = append(breakdown, entity.ReportCategoryBreakdown{Name: "Khác", Amount: otherAmount, TransactionCount: otherCount, ShareBPS: reportShareBPS(otherAmount, totalCategoryAmount), Bucket: "other", TargetID: "category:other", Drilldown: drilldownFor("category:other", "other", otherRoots)})
	}
	comparisonItems := make([]rolled, 0, len(comparisonRows))
	for _, item := range comparisonRows {
		comparisonItems = append(comparisonItems, *item)
	}
	contributors := make([]entity.ReportContributor, 0, len(comparisonItems))
	for _, item := range comparisonItems {
		delta, err := entity.CheckedMoneyDifference(item.current, item.previous)
		if err != nil {
			return nil, nil, err
		}
		change, err := entity.CheckedMoneyBPSChange(item.current, item.previous)
		if err != nil {
			return nil, nil, err
		}
		bucket, targetID := "root", "comparison:category:"+item.id
		rootIDs := []string{item.id}
		if item.id == "" {
			bucket, targetID, rootIDs = "uncategorized", "comparison:category:uncategorized", nil
		}
		drilldown := drilldownFor(targetID, bucket, rootIDs)
		drilldown.From = scope.Comparison.From
		drilldown.EndExclusive = scope.Comparison.EndExclusive
		contributors = append(contributors, entity.ReportContributor{CategoryID: categoryPointer(item.id), Name: item.name, CurrentAmount: item.current, PreviousAmount: item.previous, Delta: delta, CurrentCount: item.currentCount, PreviousCount: item.previousCount, ChangeBPS: change, FactRef: "category:" + item.id, TargetID: targetID, Drilldown: drilldown})
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

func reportShareBPS(amount, total int64) int64 {
	if amount <= 0 || total <= 0 {
		return 0
	}
	numerator := new(big.Int).Mul(big.NewInt(amount), big.NewInt(10000))
	return numerator.Quo(numerator, big.NewInt(total)).Int64()
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
	// Keep the owned history until after calculation: prior periods determine
	// carryover even when they do not overlap the displayed month.
	var budgets []entity.Budget
	if err := tx.WithContext(ctx).Where("owner_id = ?", ownerID).Order("start_date ASC, id ASC").Find(&budgets).Error; err != nil {
		return nil, err
	}
	var transactions []entity.Transaction
	query := tx.WithContext(ctx).Where("owner_id = ? AND occurred_at < ? AND type = ?", ownerID, scope.Current.EndExclusive, entity.TransactionTypeExpense)
	query = applyCanonicalReportEligibility(query, "transactions")
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	var categories []entity.Category
	if err := tx.WithContext(ctx).Where("owner_id = ? OR owner_id IS NULL", ownerID).Find(&categories).Error; err != nil {
		return nil, err
	}
	summary, err := entity.CalculateBudgetsChecked(budgets, transactions, categories, scope.Current.EndExclusive.Add(-time.Nanosecond), scope.Location)
	if err != nil {
		return nil, err
	}
	children := map[string][]string{}
	for _, category := range categories {
		if category.ParentID != nil {
			children[*category.ParentID] = append(children[*category.ParentID], category.ID)
		}
	}
	result := make([]entity.ReportBudgetPreview, 0)
	for _, budget := range summary.Items {
		start := time.Date(budget.StartDate.Year(), budget.StartDate.Month(), budget.StartDate.Day(), 0, 0, 0, 0, scope.Location).UTC()
		end := time.Date(budget.EndDate.Year(), budget.EndDate.Month(), budget.EndDate.Day(), 0, 0, 0, 0, scope.Location).AddDate(0, 0, 1).UTC()
		if !start.Before(scope.Current.EndExclusive) || !end.After(scope.Current.From) {
			continue
		}
		if budget.WalletID != nil && scope.Scope.WalletID != "" && *budget.WalletID != scope.Scope.WalletID {
			continue
		}
		if end.After(scope.Current.EndExclusive) {
			end = scope.Current.EndExclusive
		}
		// A crossing period keeps its canonical start; month-only spending would
		// understate its consumption. Carryover above always uses the full scope.
		// A budget is judged against its own scope. An all-wallet budget keeps all wallets in both
		// the calculation and its evidence, so a wallet-filtered report cannot report "not over
		// limit" by comparing one wallet's spending with the whole budget's limit.
		walletID := ""
		if budget.WalletID != nil {
			walletID = *budget.WalletID
		}
		drilldown := entity.ReportDrilldown{TargetID: "budget:" + budget.ID, From: start, EndExclusive: end, WalletID: walletID, Type: entity.TransactionTypeExpense}
		if budget.CategoryID != nil {
			drilldown.CategoryBucket = "root"
			drilldown.CategoryIDs = []string{*budget.CategoryID}
			seen := map[string]bool{*budget.CategoryID: true}
			for i := 0; i < len(drilldown.CategoryIDs); i++ {
				for _, id := range children[drilldown.CategoryIDs[i]] {
					if !seen[id] {
						seen[id] = true
						drilldown.CategoryIDs = append(drilldown.CategoryIDs, id)
					}
				}
			}
			sort.Strings(drilldown.CategoryIDs)
		}
		spent := budget.Spent
		remaining, err := entity.CheckedMoneyDifference(budget.EffectiveLimitAmount, spent)
		if err != nil {
			return nil, err
		}
		result = append(result, entity.ReportBudgetPreview{ID: budget.ID, Name: budget.Name, LimitAmount: budget.EffectiveLimitAmount, Spent: spent, Remaining: remaining, OverLimit: spent > budget.EffectiveLimitAmount, WalletID: budget.WalletID, CategoryID: budget.CategoryID, FactRef: "budget:" + budget.ID, From: start, EndExclusive: end, Drilldown: drilldown})
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
