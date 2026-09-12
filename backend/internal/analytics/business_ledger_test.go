package analytics_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/analytics"
	"mypocket/internal/finance"
	"mypocket/internal/planning"
	"mypocket/internal/platform/db"
	"mypocket/internal/portfolio"
)

type ledgerExpectation struct {
	WalletBalances map[string]int64
	IncomeVND      int64
	ExpenseVND     int64
	NetVND         int64
}

func TestBusinessLedgerKeepsBalancesReportsDebtAndPortfolioConsistent(t *testing.T) {
	conn := migratedLedgerPostgres(t)
	ctx := context.Background()
	financeRepo := finance.NewRepository(conn)
	planningRepo := planning.NewRepository(conn)
	portfolioRepo := portfolio.NewRepository(conn)
	analyticsRepo := analytics.NewRepository(conn)
	userID := createLedgerUser(t, conn)
	foodID := ledgerCategoryID(t, conn, "expense_food")
	shoppingID := ledgerCategoryID(t, conn, "expense_shopping")
	incomeID := ledgerCategoryID(t, conn, "income_salary")
	occurredAt := time.Date(2026, time.September, 11, 3, 0, 0, 0, time.UTC)

	walletA := mustWallet(t, financeRepo, userID, "Tiền mặt")
	walletB := mustWallet(t, financeRepo, userID, "Ngân hàng")
	mustTransaction(t, financeRepo, userID, finance.CreateTransactionInput{IdempotencyKey: "ledger-income", Type: finance.TransactionIncome, SourceWalletID: walletA.ID, CategoryID: incomeID, AmountVND: 1_000_000, OccurredAt: occurredAt})
	mustTransaction(t, financeRepo, userID, finance.CreateTransactionInput{IdempotencyKey: "ledger-food", Type: finance.TransactionExpense, SourceWalletID: walletA.ID, CategoryID: foodID, AmountVND: 100_000, OccurredAt: occurredAt.Add(time.Minute)})
	mustTransaction(t, financeRepo, userID, finance.CreateTransactionInput{IdempotencyKey: "ledger-transfer", Type: finance.TransactionTransfer, SourceWalletID: walletA.ID, DestinationWalletID: walletB.ID, AmountVND: 200_000, OccurredAt: occurredAt.Add(2 * time.Minute)})
	target := int64(300_000)
	mustTransaction(t, financeRepo, userID, finance.CreateTransactionInput{IdempotencyKey: "ledger-adjustment", Type: finance.TransactionAdjustment, SourceWalletID: walletB.ID, TargetBalanceVND: &target, OccurredAt: occurredAt.Add(3 * time.Minute)})
	mustTransaction(t, financeRepo, userID, finance.CreateTransactionInput{IdempotencyKey: "ledger-excluded", Type: finance.TransactionExpense, SourceWalletID: walletA.ID, CategoryID: foodID, AmountVND: 50_000, OccurredAt: occurredAt.Add(4 * time.Minute), ExcludedFromReports: true})

	editable := mustTransaction(t, financeRepo, userID, finance.CreateTransactionInput{IdempotencyKey: "ledger-edit", Type: finance.TransactionExpense, SourceWalletID: walletA.ID, CategoryID: foodID, AmountVND: 20_000, OccurredAt: occurredAt.Add(5 * time.Minute)})
	if _, err := financeRepo.UpdateTransaction(ctx, userID, editable.ID, finance.UpdateTransactionInput{BaseVersion: editable.Version, Type: finance.TransactionExpense, SourceWalletID: walletB.ID, CategoryID: shoppingID, AmountVND: 30_000, OccurredAt: occurredAt.Add(6 * time.Minute)}); err != nil {
		t.Fatalf("update editable transaction: %v", err)
	}
	archived := mustTransaction(t, financeRepo, userID, finance.CreateTransactionInput{IdempotencyKey: "ledger-archive", Type: finance.TransactionExpense, SourceWalletID: walletA.ID, CategoryID: foodID, AmountVND: 40_000, OccurredAt: occurredAt.Add(7 * time.Minute)})
	if err := financeRepo.ArchiveTransaction(ctx, userID, archived.ID, archived.Version); err != nil {
		t.Fatalf("archive transaction: %v", err)
	}

	obligation, err := planningRepo.CreateObligation(ctx, userID, planning.CreateObligationInput{Direction: planning.ObligationBorrowed, PrincipalVND: 100_000, Counterparty: "Bạn A", DueOn: "2026-12-31"})
	if err != nil {
		t.Fatalf("create obligation: %v", err)
	}
	repayment := mustTransaction(t, financeRepo, userID, finance.CreateTransactionInput{IdempotencyKey: "ledger-repayment", Type: finance.TransactionExpense, SourceWalletID: walletA.ID, CategoryID: shoppingID, AmountVND: 60_000, OccurredAt: occurredAt.Add(8 * time.Minute)})
	if err := planningRepo.LinkObligationRepayment(ctx, userID, obligation.ID, repayment.ID); err != nil {
		t.Fatalf("link repayment: %v", err)
	}

	position, err := portfolioRepo.CreatePosition(ctx, userID, portfolio.CreatePositionInput{Type: portfolio.AssetGold, Name: "Vàng kiểm thử", Unit: "gram", PricingMode: portfolio.PricingManual})
	if err != nil {
		t.Fatalf("create position: %v", err)
	}
	position, err = portfolioRepo.AddTrade(ctx, userID, position.ID, portfolio.AddTradeInput{Side: portfolio.TradeBuy, Quantity: "2", UnitPriceVND: 90_000, OccurredAt: occurredAt.Add(9 * time.Minute), BaseVersion: position.Version})
	if err != nil {
		t.Fatalf("add trade: %v", err)
	}
	if _, err := portfolioRepo.AddPrice(ctx, userID, position.ID, portfolio.AddPriceInput{UnitPriceVND: 100_000, PricedAt: occurredAt.Add(10 * time.Minute), Source: "manual", BaseVersion: position.Version}); err != nil {
		t.Fatalf("add price: %v", err)
	}

	want := ledgerExpectation{
		WalletBalances: map[string]int64{walletA.ID: 590_000, walletB.ID: 270_000},
		IncomeVND:      1_000_000,
		ExpenseVND:     190_000,
		NetVND:         810_000,
	}
	for walletID, balance := range want.WalletBalances {
		assertLedgerWalletBalance(t, conn, walletID, balance)
	}
	filter := analytics.Filter{From: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)), To: time.Date(2026, time.September, 30, 23, 59, 59, 999999999, time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60))}
	summary, err := analyticsRepo.Summary(ctx, userID, filter)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.IncomeVND != want.IncomeVND || summary.ExpenseVND != want.ExpenseVND || summary.NetIncomeVND != want.NetVND {
		t.Fatalf("unexpected report summary: %+v", summary)
	}
	categories, err := analyticsRepo.Categories(ctx, userID, filter)
	if err != nil {
		t.Fatalf("categories: %v", err)
	}
	wantCategories := map[string]int64{foodID: 100_000, shoppingID: 90_000}
	for _, category := range categories {
		if want, ok := wantCategories[category.CategoryID]; ok {
			if category.AmountVND != want {
				t.Fatalf("category %s = %d, want %d", category.CategoryID, category.AmountVND, want)
			}
			delete(wantCategories, category.CategoryID)
		}
	}
	if len(wantCategories) != 0 {
		t.Fatalf("missing category totals: %+v; got %+v", wantCategories, categories)
	}
	dashboard, err := analyticsRepo.Dashboard(ctx, userID, filter)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if dashboard.WalletNetWorthVND != 860_000 || dashboard.InvestmentMarketValueVND != 200_000 || dashboard.CombinedNetWorthVND != 1_060_000 {
		t.Fatalf("unexpected net-worth totals: %+v", dashboard)
	}
}

func migratedLedgerPostgres(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set; business-ledger integration proof skipped")
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

func createLedgerUser(t *testing.T, conn *sql.DB) string {
	t.Helper()
	var userID string
	if err := conn.QueryRowContext(context.Background(), `INSERT INTO users (google_subject, email, email_verified) VALUES ('ledger-subject', 'ledger@example.com', true) RETURNING id::text`).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return userID
}

func ledgerCategoryID(t *testing.T, conn *sql.DB, systemKey string) string {
	t.Helper()
	var id string
	if err := conn.QueryRowContext(context.Background(), `SELECT id::text FROM categories WHERE system_key = $1`, systemKey).Scan(&id); err != nil {
		t.Fatalf("find category %s: %v", systemKey, err)
	}
	return id
}

func mustWallet(t *testing.T, repo *finance.Repository, userID, name string) finance.Wallet {
	t.Helper()
	wallet, err := repo.CreateWallet(context.Background(), userID, finance.CreateWalletInput{Name: name, Type: finance.WalletBasic})
	if err != nil {
		t.Fatalf("create wallet %s: %v", name, err)
	}
	return wallet
}

func mustTransaction(t *testing.T, repo *finance.Repository, userID string, input finance.CreateTransactionInput) finance.Transaction {
	t.Helper()
	transaction, err := repo.CreateTransaction(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("create transaction %s: %v", input.IdempotencyKey, err)
	}
	return transaction
}

func assertLedgerWalletBalance(t *testing.T, conn *sql.DB, walletID string, want int64) {
	t.Helper()
	var got int64
	if err := conn.QueryRowContext(context.Background(), `SELECT balance_vnd FROM wallets WHERE id = $1`, walletID).Scan(&got); err != nil {
		t.Fatalf("wallet balance: %v", err)
	}
	if got != want {
		t.Fatalf("wallet %s balance = %d, want %d", walletID, got, want)
	}
}
