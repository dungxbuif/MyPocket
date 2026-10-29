package planning_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/finance"
	"mypocket/internal/planning"
	"mypocket/internal/platform/db"
)

func TestPeriodWindowUsesHoChiMinhBoundaries(t *testing.T) {
	now := time.Date(2026, 8, 31, 4, 0, 0, 0, time.UTC)
	start, end, err := planning.PeriodWindow(planning.BudgetMonthly, nil, nil, now)
	if err != nil {
		t.Fatalf("monthly period: %v", err)
	}
	if start.Format("2006-01-02") != "2026-08-01" || end.Format("2006-01-02") != "2026-08-31" {
		t.Fatalf("unexpected monthly window: %s %s", start, end)
	}
}

func TestBudgetProgressExcludesNonExpenseAndReportExcludedTransactions(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "budget-progress@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	expenseCategoryID := findPlanningSystemCategory(t, conn, "expense_food")
	incomeCategoryID := findPlanningSystemCategory(t, conn, "income_salary")
	occurred := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)

	budget, err := planningRepo.CreateBudget(context.Background(), owner, planning.CreateBudgetInput{Name: "Ăn uống", PeriodType: planning.BudgetMonthly, AmountVND: 500000, CategoryIDs: []string{expenseCategoryID}})
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "expense-1", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: expenseCategoryID, BudgetID: budget.ID, AmountVND: 450000, OccurredAt: occurred, Note: "Counted"})
	createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "income-1", Type: finance.TransactionIncome, SourceWalletID: wallet.ID, CategoryID: incomeCategoryID, AmountVND: 2000000, OccurredAt: occurred, Note: "Excluded income"})
	createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "excluded-1", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: expenseCategoryID, BudgetID: budget.ID, AmountVND: 250000, OccurredAt: occurred, Note: "Excluded report", ExcludedFromReports: true})
	progress, err := planningRepo.ListBudgetProgress(context.Background(), owner, occurred)
	if err != nil {
		t.Fatalf("list progress: %v", err)
	}
	if len(progress) != 1 || progress[0].Budget.ID != budget.ID {
		t.Fatalf("unexpected progress rows: %#v", progress)
	}
	if progress[0].SpentVND != 450000 || progress[0].RemainingVND != 50000 || progress[0].Percent != 90 || !progress[0].Alert80 || progress[0].Alert100 {
		t.Fatalf("wrong progress calculation: %#v", progress[0])
	}
	assertBudgetAlertCount(t, conn, budget.ID, 1)

	if _, err := planningRepo.ListBudgetProgress(context.Background(), owner, occurred); err != nil {
		t.Fatalf("list progress again: %v", err)
	}
	assertBudgetAlertCount(t, conn, budget.ID, 1)
}

func TestBudgetProgressCountsOnlyExplicitlyAssignedExpenseTransactions(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "budget-assignment@example.com")
	other := createPlanningUser(t, conn, "budget-assignment-other@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	wallet := createPlanningWallet(t, financeRepo, owner)
	destination, err := financeRepo.CreateWallet(context.Background(), owner, finance.CreateWalletInput{Name: "Ví nhận", Type: finance.WalletBasic})
	if err != nil {
		t.Fatalf("create destination wallet: %v", err)
	}
	otherWallet := createPlanningWallet(t, financeRepo, other)
	expenseCategoryID := findPlanningSystemCategory(t, conn, "expense_food")
	incomeCategoryID := findPlanningSystemCategory(t, conn, "income_salary")
	occurred := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)

	budget, err := planningRepo.CreateBudget(context.Background(), owner, planning.CreateBudgetInput{Name: "Hũ ăn uống", PeriodType: planning.BudgetMonthly, AmountVND: 500000, CategoryIDs: []string{expenseCategoryID}})
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	otherBudget, err := planningRepo.CreateBudget(context.Background(), other, planning.CreateBudgetInput{Name: "Other budget", PeriodType: planning.BudgetMonthly, AmountVND: 500000, CategoryIDs: []string{expenseCategoryID}})
	if err != nil {
		t.Fatalf("create other budget: %v", err)
	}

	assigned := createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "assigned-expense", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: expenseCategoryID, BudgetID: budget.ID, AmountVND: 125000, OccurredAt: occurred, Note: "Assigned"})
	createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "same-category-unassigned", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: expenseCategoryID, AmountVND: 300000, OccurredAt: occurred, Note: "Same category, not assigned"})
	createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "assigned-excluded", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: expenseCategoryID, BudgetID: budget.ID, AmountVND: 50000, OccurredAt: occurred, Note: "Excluded", ExcludedFromReports: true})
	createPlanningTransaction(t, financeRepo, owner, finance.CreateTransactionInput{IdempotencyKey: "assigned-outside-period", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: expenseCategoryID, BudgetID: budget.ID, AmountVND: 90000, OccurredAt: occurred.AddDate(0, -1, 0), Note: "Outside"})

	if _, err := financeRepo.CreateTransaction(context.Background(), owner, finance.CreateTransactionInput{IdempotencyKey: "budget-income", Type: finance.TransactionIncome, SourceWalletID: wallet.ID, CategoryID: incomeCategoryID, BudgetID: budget.ID, AmountVND: 100000, OccurredAt: occurred}); !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("income with budget must be validation error, got %v", err)
	}
	if _, err := financeRepo.CreateTransaction(context.Background(), owner, finance.CreateTransactionInput{IdempotencyKey: "budget-transfer", Type: finance.TransactionTransfer, SourceWalletID: wallet.ID, DestinationWalletID: destination.ID, BudgetID: budget.ID, AmountVND: 100000, OccurredAt: occurred}); !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("transfer with budget must be validation error, got %v", err)
	}
	if _, err := financeRepo.CreateTransaction(context.Background(), owner, finance.CreateTransactionInput{IdempotencyKey: "foreign-budget", Type: finance.TransactionExpense, SourceWalletID: wallet.ID, CategoryID: expenseCategoryID, BudgetID: otherBudget.ID, AmountVND: 100000, OccurredAt: occurred}); !errors.Is(err, finance.ErrForbidden) {
		t.Fatalf("foreign budget must be forbidden, got %v", err)
	}
	if _, err := financeRepo.CreateTransaction(context.Background(), other, finance.CreateTransactionInput{IdempotencyKey: "other-owner-own-budget", Type: finance.TransactionExpense, SourceWalletID: otherWallet.ID, CategoryID: expenseCategoryID, BudgetID: otherBudget.ID, AmountVND: 100000, OccurredAt: occurred}); err != nil {
		t.Fatalf("other owner own budget should be accepted: %v", err)
	}

	progress, err := planningRepo.ListBudgetProgress(context.Background(), owner, occurred)
	if err != nil {
		t.Fatalf("list progress: %v", err)
	}
	if len(progress) != 1 || progress[0].Budget.ID != budget.ID {
		t.Fatalf("unexpected progress rows: %#v", progress)
	}
	if progress[0].SpentVND != assigned.AmountVND || progress[0].RemainingVND != 375000 || progress[0].Percent != 25 {
		t.Fatalf("budget progress must count only assigned report-included current-period expenses: %#v", progress[0])
	}
}

func TestBudgetUpdateArchiveAndCategoryOwnership(t *testing.T) {
	conn := migratedPlanningPostgres(t)
	owner := createPlanningUser(t, conn, "budget-owner@example.com")
	other := createPlanningUser(t, conn, "budget-other@example.com")
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	category := createPlanningCategory(t, financeRepo, owner, "Cafe")
	otherCategory := createPlanningCategory(t, financeRepo, other, "Hidden")

	budget, err := planningRepo.CreateBudget(context.Background(), owner, planning.CreateBudgetInput{Name: "Cafe", PeriodType: planning.BudgetWeekly, AmountVND: 300000, CategoryIDs: []string{category.ID}})
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	if _, err := planningRepo.UpdateBudget(context.Background(), owner, budget.ID, planning.UpdateBudgetInput{BaseVersion: budget.Version, Name: "Invalid", PeriodType: planning.BudgetWeekly, AmountVND: 300000, CategoryIDs: []string{otherCategory.ID}}); !errors.Is(err, planning.ErrForbidden) {
		t.Fatalf("expected category ownership rejection, got %v", err)
	}
	updated, err := planningRepo.UpdateBudget(context.Background(), owner, budget.ID, planning.UpdateBudgetInput{BaseVersion: budget.Version, Name: "Cafe tháng", PeriodType: planning.BudgetMonthly, AmountVND: 600000})
	if err != nil {
		t.Fatalf("update budget: %v", err)
	}
	if updated.Name != "Cafe tháng" || !updated.AllCategories || updated.Version != 2 {
		t.Fatalf("unexpected updated budget: %#v", updated)
	}
	if _, err := planningRepo.UpdateBudget(context.Background(), owner, budget.ID, planning.UpdateBudgetInput{BaseVersion: budget.Version, Name: "Stale", PeriodType: planning.BudgetMonthly, AmountVND: 700000}); !errors.Is(err, planning.ErrVersionConflict) {
		t.Fatalf("expected stale budget update conflict, got %v", err)
	}
	if err := planningRepo.ArchiveBudget(context.Background(), owner, budget.ID, budget.Version); !errors.Is(err, planning.ErrVersionConflict) {
		t.Fatalf("expected stale budget archive conflict, got %v", err)
	}
	if err := planningRepo.ArchiveBudget(context.Background(), other, budget.ID, updated.Version); !errors.Is(err, planning.ErrForbidden) {
		t.Fatalf("expected other user archive rejection, got %v", err)
	}
	if err := planningRepo.ArchiveBudget(context.Background(), owner, budget.ID, updated.Version); err != nil {
		t.Fatalf("archive budget: %v", err)
	}
	progress, err := planningRepo.ListBudgetProgress(context.Background(), owner, time.Now())
	if err != nil {
		t.Fatalf("list after archive: %v", err)
	}
	if len(progress) != 0 {
		t.Fatalf("archived budget should be hidden: %#v", progress)
	}
}

func migratedPlanningPostgres(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set; planning repository integration proof skipped")
	}
	conn, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.ExecContext(context.Background(), `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if err := db.Migrate(context.Background(), conn, os.DirFS("../../migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return conn
}

func createPlanningUser(t *testing.T, conn *sql.DB, email string) string {
	t.Helper()
	var id string
	if err := conn.QueryRowContext(context.Background(), `
		INSERT INTO users (google_subject, email, email_verified, display_name)
		VALUES ($1, $2, true, $2)
		RETURNING id::text
	`, "google-"+email, email).Scan(&id); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}

func createPlanningWallet(t *testing.T, repo *finance.Repository, userID string) finance.Wallet {
	t.Helper()
	wallet, err := repo.CreateWallet(context.Background(), userID, finance.CreateWalletInput{Name: "Tiền mặt", Type: finance.WalletBasic})
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	return wallet
}

func createPlanningCategory(t *testing.T, repo *finance.Repository, userID string, name string) finance.Category {
	t.Helper()
	category, err := repo.CreateCategory(context.Background(), userID, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: name})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	return category
}

func createPlanningTransaction(t *testing.T, repo *finance.Repository, userID string, input finance.CreateTransactionInput) finance.Transaction {
	t.Helper()
	transaction, err := repo.CreateTransaction(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("create transaction %s: %v", input.IdempotencyKey, err)
	}
	return transaction
}

func findPlanningSystemCategory(t *testing.T, conn *sql.DB, key string) string {
	t.Helper()
	var id string
	if err := conn.QueryRowContext(context.Background(), `SELECT id::text FROM categories WHERE system_key = $1`, key).Scan(&id); err != nil {
		t.Fatalf("find system category %s: %v", key, err)
	}
	return id
}

func assertBudgetAlertCount(t *testing.T, conn *sql.DB, budgetID string, want int) {
	t.Helper()
	var got int
	if err := conn.QueryRowContext(context.Background(), `SELECT count(*) FROM budget_alerts WHERE budget_id = $1`, budgetID).Scan(&got); err != nil {
		t.Fatalf("count alerts: %v", err)
	}
	if got != want {
		t.Fatalf("expected %d budget alerts, got %d", want, got)
	}
}
