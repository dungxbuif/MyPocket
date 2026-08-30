package finance_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/finance"
	"mypocket/internal/platform/db"
)

func TestRepositoryListsOnlyUserWallets(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userA := createFinanceUser(t, conn, "owner-a@example.com")
	userB := createFinanceUser(t, conn, "owner-b@example.com")

	walletA, err := repo.CreateWallet(context.Background(), userA, finance.CreateWalletInput{
		Name: "Tiền mặt",
		Type: finance.WalletCash,
	})
	if err != nil {
		t.Fatalf("create wallet A: %v", err)
	}
	if _, err := repo.CreateWallet(context.Background(), userB, finance.CreateWalletInput{
		Name: "Ngân hàng",
		Type: finance.WalletBank,
	}); err != nil {
		t.Fatalf("create wallet B: %v", err)
	}

	wallets, err := repo.ListWallets(context.Background(), userA)
	if err != nil {
		t.Fatalf("list wallets: %v", err)
	}
	if len(wallets) != 1 {
		t.Fatalf("expected one wallet for user A, got %#v", wallets)
	}
	if wallets[0].ID != walletA.ID || wallets[0].Name != "Tiền mặt" {
		t.Fatalf("listed wrong wallet: %#v", wallets[0])
	}
}

func TestRepositoryAllowsOneActiveDefaultAIWalletPerUser(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "default-ai@example.com")
	first := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	second := createFinanceWallet(t, repo, userID, "Ví điện tử", finance.WalletEWallet)

	if err := repo.SetDefaultAIWallet(context.Background(), userID, first.ID); err != nil {
		t.Fatalf("set first default: %v", err)
	}
	if err := repo.SetDefaultAIWallet(context.Background(), userID, second.ID); err != nil {
		t.Fatalf("set second default: %v", err)
	}

	wallets, err := repo.ListWallets(context.Background(), userID)
	if err != nil {
		t.Fatalf("list wallets: %v", err)
	}
	defaults := 0
	for _, wallet := range wallets {
		if wallet.IsDefaultAI {
			defaults++
			if wallet.ID != second.ID {
				t.Fatalf("expected second wallet as default, got %#v", wallet)
			}
		}
	}
	if defaults != 1 {
		t.Fatalf("expected exactly one default AI wallet, got %d from %#v", defaults, wallets)
	}
}

func TestRepositoryUpdatesAndArchivesOnlyOwnedWallet(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userA := createFinanceUser(t, conn, "wallet-update-a@example.com")
	userB := createFinanceUser(t, conn, "wallet-update-b@example.com")
	wallet := createFinanceWallet(t, repo, userA, "Tiền mặt", finance.WalletCash)

	updated, err := repo.UpdateWallet(context.Background(), userA, wallet.ID, finance.UpdateWalletInput{
		Name:           "Tiền ăn",
		IncludeInTotal: ptrBool(false),
	})
	if err != nil {
		t.Fatalf("update wallet: %v", err)
	}
	if updated.Name != "Tiền ăn" || updated.IncludeInTotal {
		t.Fatalf("wallet update did not persist: %#v", updated)
	}

	if err := repo.ArchiveWallet(context.Background(), userB, wallet.ID); !errors.Is(err, finance.ErrForbidden) {
		t.Fatalf("expected other user archive forbidden, got %v", err)
	}
	if err := repo.ArchiveWallet(context.Background(), userA, wallet.ID); err != nil {
		t.Fatalf("archive wallet: %v", err)
	}
	wallets, err := repo.ListWallets(context.Background(), userA)
	if err != nil {
		t.Fatalf("list wallets: %v", err)
	}
	if len(wallets) != 0 {
		t.Fatalf("expected archived wallet hidden from list, got %#v", wallets)
	}
}

func TestRepositoryRejectsActivationForAnotherUsersWallet(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userA := createFinanceUser(t, conn, "activation-a@example.com")
	userB := createFinanceUser(t, conn, "activation-b@example.com")
	wallet := createFinanceWallet(t, repo, userA, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "expense_food")

	err := repo.SetWalletCategoryActive(context.Background(), userB, wallet.ID, categoryID, false)

	if !errors.Is(err, finance.ErrForbidden) {
		t.Fatalf("expected forbidden activation error, got %v", err)
	}
}

func TestRepositoryListsSystemAndOwnedCategories(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userA := createFinanceUser(t, conn, "category-a@example.com")
	userB := createFinanceUser(t, conn, "category-b@example.com")

	custom, err := repo.CreateCategory(context.Background(), userA, finance.CreateCategoryInput{
		Kind: finance.CategoryExpense,
		Name: "Cà phê",
	})
	if err != nil {
		t.Fatalf("create user category: %v", err)
	}
	if _, err := repo.CreateCategory(context.Background(), userB, finance.CreateCategoryInput{
		Kind: finance.CategoryExpense,
		Name: "Hidden",
	}); err != nil {
		t.Fatalf("create other user category: %v", err)
	}

	categories, err := repo.ListCategories(context.Background(), userA)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if !containsCategoryKey(categories, "expense_food") {
		t.Fatalf("expected system category in list: %#v", categories)
	}
	if !containsCategoryID(categories, custom.ID) {
		t.Fatalf("expected owned custom category in list: %#v", categories)
	}
	if containsCategoryName(categories, "Hidden") {
		t.Fatalf("list leaked another user's category: %#v", categories)
	}
}

func TestRepositoryUpdatesAndArchivesOnlyUserCategories(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "category-update@example.com")
	systemID := findSystemCategory(t, conn, "expense_food")
	custom, err := repo.CreateCategory(context.Background(), userID, finance.CreateCategoryInput{
		Kind: finance.CategoryExpense,
		Name: "Cà phê",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	updated, err := repo.UpdateCategory(context.Background(), userID, custom.ID, finance.UpdateCategoryInput{Name: "Cafe"})
	if err != nil {
		t.Fatalf("update category: %v", err)
	}
	if updated.Name != "Cafe" {
		t.Fatalf("expected updated category name, got %#v", updated)
	}
	if _, err := repo.UpdateCategory(context.Background(), userID, systemID, finance.UpdateCategoryInput{Name: "Tên mới"}); !errors.Is(err, finance.ErrSystemCategoryLocked) {
		t.Fatalf("expected system category lock, got %v", err)
	}
	if err := repo.ArchiveCategory(context.Background(), userID, systemID); !errors.Is(err, finance.ErrSystemCategoryLocked) {
		t.Fatalf("expected system category archive lock, got %v", err)
	}
	if err := repo.ArchiveCategory(context.Background(), userID, custom.ID); err != nil {
		t.Fatalf("archive custom category: %v", err)
	}
	categories, err := repo.ListCategories(context.Background(), userID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if containsCategoryID(categories, custom.ID) {
		t.Fatalf("expected archived category hidden from list: %#v", categories)
	}
}

func TestRepositoryUpsertsWalletCategoryActivation(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "activation-upsert@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "expense_food")

	if err := repo.SetWalletCategoryActive(context.Background(), userID, wallet.ID, categoryID, false); err != nil {
		t.Fatalf("disable category: %v", err)
	}
	if err := repo.SetWalletCategoryActive(context.Background(), userID, wallet.ID, categoryID, true); err != nil {
		t.Fatalf("enable category: %v", err)
	}

	var active bool
	var count int
	err := conn.QueryRowContext(context.Background(), `
		SELECT bool_or(active), count(*)
		FROM wallet_category_settings
		WHERE wallet_id = $1 AND category_id = $2 AND user_id = $3
	`, wallet.ID, categoryID, userID).Scan(&active, &count)
	if err != nil {
		t.Fatalf("load activation: %v", err)
	}
	if !active || count != 1 {
		t.Fatalf("expected one active setting, active=%v count=%d", active, count)
	}
}

func TestRepositoryCreatesIncomeExpenseAndAdjustmentTransactions(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "transactions@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	expenseCategoryID := findSystemCategory(t, conn, "expense_food")
	incomeCategoryID := findSystemCategory(t, conn, "income_salary")

	income, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "income-1",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     incomeCategoryID,
		AmountVND:      500_000,
		OccurredAt:     fixedFinanceTime(),
		Note:           " Lương phụ ",
	})
	if err != nil {
		t.Fatalf("create income: %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 500_000, 2)
	if income.BalanceAfterVND != 500_000 || income.Note != "Lương phụ" {
		t.Fatalf("income transaction did not persist normalized result: %#v", income)
	}

	expense, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "expense-1",
		Type:           finance.TransactionExpense,
		SourceWalletID: wallet.ID,
		CategoryID:     expenseCategoryID,
		AmountVND:      125_000,
		OccurredAt:     fixedFinanceTime().Add(time.Hour),
		WithPerson:     " Bạn ",
	})
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 375_000, 3)
	if expense.BalanceAfterVND != 375_000 || expense.WithPerson != "Bạn" {
		t.Fatalf("expense transaction did not persist normalized result: %#v", expense)
	}

	target := int64(100_000)
	adjustment, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey:   "adjustment-1",
		Type:             finance.TransactionAdjustment,
		SourceWalletID:   wallet.ID,
		TargetBalanceVND: &target,
		OccurredAt:       fixedFinanceTime().Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create adjustment: %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 100_000, 4)
	if adjustment.AmountVND != 100_000 || adjustment.BalanceAfterVND != 100_000 {
		t.Fatalf("adjustment should store target balance as amount: %#v", adjustment)
	}
}

func TestRepositoryCreatesTransferTransactionAtomically(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "transfer@example.com")
	source := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	destination := createFinanceWallet(t, repo, userID, "Ngân hàng", finance.WalletBank)
	creditWallet(t, conn, source.ID, 1_000_000)

	tx, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey:      "transfer-1",
		Type:                finance.TransactionTransfer,
		SourceWalletID:      source.ID,
		DestinationWalletID: destination.ID,
		AmountVND:           250_000,
		OccurredAt:          fixedFinanceTime(),
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	assertWalletBalance(t, conn, source.ID, 750_000, 3)
	assertWalletBalance(t, conn, destination.ID, 250_000, 2)
	if tx.BalanceAfterVND != 750_000 || tx.DestinationWalletID != destination.ID {
		t.Fatalf("transfer transaction did not persist expected state: %#v", tx)
	}
}

func TestRepositoryReplaysDuplicateTransactionIdempotencyKey(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "idempotent@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "income_salary")
	input := finance.CreateTransactionInput{
		IdempotencyKey: "idem-1",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      120_000,
		OccurredAt:     fixedFinanceTime(),
	}

	first, err := repo.CreateTransaction(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("create first transaction: %v", err)
	}
	second, err := repo.CreateTransaction(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("replay transaction: %v", err)
	}

	if second.ID != first.ID || second.BalanceAfterVND != first.BalanceAfterVND {
		t.Fatalf("idempotent replay returned different transaction: first=%#v second=%#v", first, second)
	}
	assertWalletBalance(t, conn, wallet.ID, 120_000, 2)
}

func TestRepositoryRejectsReusedTransactionIdempotencyKeyWithDifferentRequest(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "idempotent-conflict@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "income_salary")

	if _, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "idem-conflict-1",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      120_000,
		OccurredAt:     fixedFinanceTime(),
	}); err != nil {
		t.Fatalf("create first transaction: %v", err)
	}
	_, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "idem-conflict-1",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      130_000,
		OccurredAt:     fixedFinanceTime(),
	})

	if !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected validation error for idempotency conflict, got %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 120_000, 2)
}

func TestRepositoryRejectsTransactionForInactiveWalletCategory(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "inactive-category@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "expense_food")
	if err := repo.SetWalletCategoryActive(context.Background(), userID, wallet.ID, categoryID, false); err != nil {
		t.Fatalf("disable category: %v", err)
	}

	_, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "inactive-category-1",
		Type:           finance.TransactionExpense,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      50_000,
		OccurredAt:     fixedFinanceTime(),
	})

	if !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected validation error for inactive category, got %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 0, 1)
}

func TestRepositoryRejectsTransactionForAnotherUsersWallet(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userA := createFinanceUser(t, conn, "tx-owner@example.com")
	userB := createFinanceUser(t, conn, "tx-other@example.com")
	wallet := createFinanceWallet(t, repo, userA, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "income_salary")

	_, err := repo.CreateTransaction(context.Background(), userB, finance.CreateTransactionInput{
		IdempotencyKey: "forbidden-1",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      120_000,
		OccurredAt:     fixedFinanceTime(),
	})

	if !errors.Is(err, finance.ErrForbidden) {
		t.Fatalf("expected forbidden transaction error, got %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 0, 1)
}

func migratedFinancePostgres(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set; finance repository integration proof skipped")
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

func fixedFinanceTime() time.Time {
	return time.Date(2026, 8, 30, 10, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
}

func createFinanceUser(t *testing.T, conn *sql.DB, email string) string {
	t.Helper()

	var userID string
	err := conn.QueryRowContext(context.Background(), `
		INSERT INTO users (google_subject, email, email_verified)
		VALUES ($1, $2, true)
		RETURNING id::text
	`, "subject-"+email, email).Scan(&userID)
	if err != nil {
		t.Fatalf("create finance user: %v", err)
	}
	return userID
}

func createFinanceWallet(t *testing.T, repo *finance.Repository, userID string, name string, walletType finance.WalletType) finance.Wallet {
	t.Helper()

	wallet, err := repo.CreateWallet(context.Background(), userID, finance.CreateWalletInput{Name: name, Type: walletType})
	if err != nil {
		t.Fatalf("create wallet %s: %v", name, err)
	}
	return wallet
}

func findSystemCategory(t *testing.T, conn *sql.DB, systemKey string) string {
	t.Helper()

	var categoryID string
	err := conn.QueryRowContext(context.Background(), `
		SELECT id::text
		FROM categories
		WHERE system_key = $1
	`, systemKey).Scan(&categoryID)
	if err != nil {
		t.Fatalf("find system category %s: %v", systemKey, err)
	}
	return categoryID
}

func creditWallet(t *testing.T, conn *sql.DB, walletID string, amountVND int64) {
	t.Helper()

	result, err := conn.ExecContext(context.Background(), `
		UPDATE wallets
		SET balance_vnd = $2, version = version + 1, updated_at = now()
		WHERE id = $1
	`, walletID, amountVND)
	if err != nil {
		t.Fatalf("credit wallet: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("credit wallet rows affected: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected to credit one wallet, affected %d", affected)
	}
}

func assertWalletBalance(t *testing.T, conn *sql.DB, walletID string, wantBalanceVND int64, wantVersion int64) {
	t.Helper()

	var balanceVND int64
	var version int64
	err := conn.QueryRowContext(context.Background(), `
		SELECT balance_vnd, version
		FROM wallets
		WHERE id = $1
	`, walletID).Scan(&balanceVND, &version)
	if err != nil {
		t.Fatalf("load wallet balance: %v", err)
	}
	if balanceVND != wantBalanceVND || version != wantVersion {
		t.Fatalf("wallet balance/version = %d/%d, want %d/%d", balanceVND, version, wantBalanceVND, wantVersion)
	}
}

func ptrBool(value bool) *bool {
	return &value
}

func containsCategoryKey(categories []finance.Category, systemKey string) bool {
	for _, category := range categories {
		if category.SystemKey == systemKey {
			return true
		}
	}
	return false
}

func containsCategoryID(categories []finance.Category, id string) bool {
	for _, category := range categories {
		if category.ID == id {
			return true
		}
	}
	return false
}

func containsCategoryName(categories []finance.Category, name string) bool {
	for _, category := range categories {
		if category.Name == name {
			return true
		}
	}
	return false
}
