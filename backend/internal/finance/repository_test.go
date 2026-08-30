package finance_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

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
