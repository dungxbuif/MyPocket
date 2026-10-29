package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
	insiderrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type InsiderPostgresRepository struct{ db *gorm.DB }

func NewInsiderPostgresRepository(db *gorm.DB) insiderrepo.InsiderReader {
	return &InsiderPostgresRepository{db: db}
}

func (r *InsiderPostgresRepository) ReadInsider(ownerID, month, walletID, categoryID string) (*entity.InsiderSummary, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" {
		return nil, insiderrepo.ErrNotFound
	}
	parsedMonth, err := entity.ParseMonth(month)
	if err != nil {
		return nil, entity.ErrFinanceDateInvalid
	}
	var result *entity.InsiderSummary
	err = r.db.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var user entity.User
		if err := tx.Where("id = ?", ownerID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return insiderrepo.ErrNotFound
			}
			return err
		}
		if strings.TrimSpace(user.Timezone) == "" || user.Timezone == "Local" {
			return entity.ErrFinanceRangeInvalid
		}
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			return entity.ErrFinanceRangeInvalid
		}
		start, next, err := entity.MonthRangeUTC(parsedMonth, location)
		if err != nil {
			return err
		}
		previousStart, previousNext, err := entity.MonthRangeUTC(parsedMonth.AddDate(0, -1, 0), location)
		if err != nil {
			return err
		}
		income, expense, err := r.readTotals(tx, ownerID, start, next, walletID, categoryID)
		if err != nil {
			return err
		}
		previousIncome, previousExpense, err := r.readTotals(tx, ownerID, previousStart, previousNext, walletID, categoryID)
		if err != nil {
			return err
		}
		_ = previousIncome
		now := time.Now().UTC()
		localNow := now.In(location)
		monthKey := parsedMonth.Format("2006-01")
		isCurrent := localNow.Format("2006-01") == monthKey
		days := calendarDays(start.In(location), next.In(location))
		if isCurrent {
			monthStartLocal := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, location)
			days = calendarDays(monthStartLocal, time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location).AddDate(0, 0, 1))
		}
		change, err := entity.PercentBPS(expense, previousExpense)
		if err != nil {
			return err
		}
		categories, err := r.readTopCategories(tx, ownerID, start, next, walletID, categoryID)
		if err != nil {
			return err
		}
		expenses, err := r.readTopExpenses(tx, ownerID, start, next, walletID, categoryID)
		if err != nil {
			return err
		}
		result = &entity.InsiderSummary{
			Month: monthKey, Timezone: user.Timezone, StartAt: start, NextStartAt: next, IsCurrent: isCurrent,
			Income: income, Expense: expense, Net: income - expense, PreviousExpense: previousExpense,
			ExpenseDelta: expense - previousExpense, ExpenseChangeBPS: change,
			AverageDailyExpense: entity.AverageDailySpend(expense, days), DaysConsidered: days,
			SpendingIncomeRatioBPS: entity.SpendingIncomeRatioBPS(expense, income), TopCategories: categories,
			TopExpenses: expenses, Estimated: false, WalletFilter: strings.TrimSpace(walletID), CategoryFilter: strings.TrimSpace(categoryID),
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type insiderTotals struct {
	Income  int64 `gorm:"column:income"`
	Expense int64 `gorm:"column:expense"`
}

func (r *InsiderPostgresRepository) readTotals(tx *gorm.DB, ownerID string, start, next time.Time, walletID, categoryID string) (int64, int64, error) {
	var totals insiderTotals
	db := r.scoped(tx, ownerID, start, next, walletID, categoryID).Select(`COALESCE(SUM(CASE WHEN t.type = ? THEN t.amount ELSE 0 END), 0) AS income, COALESCE(SUM(CASE WHEN t.type = ? THEN t.amount ELSE 0 END), 0) AS expense`, entity.TransactionTypeIncome, entity.TransactionTypeExpense)
	if err := db.Scan(&totals).Error; err != nil {
		return 0, 0, err
	}
	return totals.Income, totals.Expense, nil
}

type insiderCategoryRow struct {
	CategoryID *string `gorm:"column:category_id"`
	Name       string  `gorm:"column:name"`
	Amount     int64   `gorm:"column:amount"`
	Count      int     `gorm:"column:count"`
}

func (r *InsiderPostgresRepository) readTopCategories(tx *gorm.DB, ownerID string, start, next time.Time, walletID, categoryID string) ([]entity.InsiderCategoryTotal, error) {
	rows := make([]insiderCategoryRow, 0, 3)
	db := r.scoped(tx, ownerID, start, next, walletID, categoryID).Where("t.type = ?", entity.TransactionTypeExpense).Select(`t.category_id, COALESCE(c.name, 'Khoản chi chưa phân nhóm') AS name, SUM(t.amount) AS amount, COUNT(*) AS count`).Group("t.category_id, c.name").Order("amount DESC").Limit(3)
	if err := db.Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]entity.InsiderCategoryTotal, 0, len(rows))
	for _, row := range rows {
		result = append(result, entity.InsiderCategoryTotal{CategoryID: row.CategoryID, Name: row.Name, Amount: row.Amount, Count: row.Count})
	}
	return result, nil
}

type insiderExpenseRow struct {
	ID           string    `gorm:"column:id"`
	WalletID     string    `gorm:"column:wallet_id"`
	WalletName   string    `gorm:"column:wallet_name"`
	CategoryID   *string   `gorm:"column:category_id"`
	CategoryName string    `gorm:"column:category_name"`
	Amount       int64     `gorm:"column:amount"`
	OccurredAt   time.Time `gorm:"column:occurred_at"`
	Note         *string   `gorm:"column:note"`
}

func (r *InsiderPostgresRepository) readTopExpenses(tx *gorm.DB, ownerID string, start, next time.Time, walletID, categoryID string) ([]entity.InsiderExpenseRow, error) {
	rows := make([]insiderExpenseRow, 0, 5)
	db := r.scoped(tx, ownerID, start, next, walletID, categoryID).Where("t.type = ?", entity.TransactionTypeExpense).Select(`t.id, t.wallet_id, COALESCE(w.name, '') AS wallet_name, t.category_id, COALESCE(c.name, 'Khoản chi chưa phân nhóm') AS category_name, t.amount, t.occurred_at, t.note`).Order("t.amount DESC, t.occurred_at DESC, t.id DESC").Limit(5)
	if err := db.Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]entity.InsiderExpenseRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, entity.InsiderExpenseRow{ID: row.ID, WalletID: row.WalletID, WalletName: row.WalletName, CategoryID: row.CategoryID, CategoryName: row.CategoryName, Amount: row.Amount, OccurredAt: row.OccurredAt, Note: row.Note})
	}
	return result, nil
}

func (r *InsiderPostgresRepository) scoped(tx *gorm.DB, ownerID string, start, next time.Time, walletID, categoryID string) *gorm.DB {
	db := tx.Table("transactions AS t").Joins("LEFT JOIN categories AS c ON c.id = t.category_id").Joins("LEFT JOIN wallets AS w ON w.id = t.wallet_id").Where("t.owner_id = ? AND t.occurred_at >= ? AND t.occurred_at < ? AND t.included_in_reports = TRUE AND t.type IN (?, ?)", ownerID, start, next, entity.TransactionTypeIncome, entity.TransactionTypeExpense).Where("NOT (t.type = ? AND COALESCE(c.system_key, '') = ?)", entity.TransactionTypeIncome, "income_transfer_in").Where("NOT (t.type = ? AND COALESCE(c.system_key, '') = ?)", entity.TransactionTypeExpense, "expense_transfer_out")
	if strings.TrimSpace(walletID) != "" {
		db = db.Where("t.wallet_id = ?", strings.TrimSpace(walletID))
	}
	if strings.TrimSpace(categoryID) != "" {
		db = db.Where("t.category_id = ?", strings.TrimSpace(categoryID))
	}
	return db
}

func calendarDays(start, next time.Time) int {
	startUTC := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	nextUTC := time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
	days := int(nextUTC.Sub(startUTC).Hours() / 24)
	if days < 1 {
		return 1
	}
	return days
}
