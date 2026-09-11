package finance_test

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"os"
	"strings"
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

func TestRepositoryStoresUserScopedReceiptMetadata(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	owner := createFinanceUser(t, conn, "receipt-owner@example.com")
	other := createFinanceUser(t, conn, "receipt-other@example.com")
	receipt, err := repo.CreateReceiptObject(context.Background(), owner, finance.CreateReceiptObjectInput{ObjectKey: "users/owner/receipt.jpg", ContentType: "image/jpeg", SizeBytes: 1234, ChecksumSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", OriginalFilename: "hoa-don.jpg"})
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	loaded, err := repo.GetReceiptObject(context.Background(), owner, receipt.ID)
	if err != nil || loaded.ObjectKey != receipt.ObjectKey {
		t.Fatalf("load receipt: %#v %v", loaded, err)
	}
	if _, err := repo.GetReceiptObject(context.Background(), other, receipt.ID); !errors.Is(err, finance.ErrForbidden) {
		t.Fatalf("expected scoped receipt access denial, got %v", err)
	}
}

func TestRepositoryRejectsInvalidReceiptMetadata(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	owner := createFinanceUser(t, conn, "receipt-invalid@example.com")
	if _, err := repo.CreateReceiptObject(context.Background(), owner, finance.CreateReceiptObjectInput{ObjectKey: "x", ContentType: "image/jpeg", SizeBytes: 1, ChecksumSHA256: "bad"}); !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRepositoryAllowsOneActiveDefaultAIWalletPerUser(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "default-ai@example.com")
	first := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	second := createFinanceWallet(t, repo, userID, "Ví điện tử", finance.WalletEWallet)

	if err := repo.SetDefaultAIWallet(context.Background(), userID, first.ID, first.Version); err != nil {
		t.Fatalf("set first default: %v", err)
	}
	if err := repo.SetDefaultAIWallet(context.Background(), userID, first.ID, first.Version); !errors.Is(err, finance.ErrConflict) {
		t.Fatalf("stale default selection error = %v, want conflict", err)
	}
	if err := repo.SetDefaultAIWallet(context.Background(), userID, second.ID, second.Version); err != nil {
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
		BaseVersion:    wallet.Version,
		Name:           "Tiền ăn",
		IncludeInTotal: ptrBool(false),
	})
	if err != nil {
		t.Fatalf("update wallet: %v", err)
	}
	if updated.Name != "Tiền ăn" || updated.IncludeInTotal {
		t.Fatalf("wallet update did not persist: %#v", updated)
	}

	if err := repo.ArchiveWallet(context.Background(), userB, wallet.ID, updated.Version); !errors.Is(err, finance.ErrForbidden) {
		t.Fatalf("expected other user archive forbidden, got %v", err)
	}
	if err := repo.ArchiveWallet(context.Background(), userA, wallet.ID, updated.Version); err != nil {
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

func TestRepositoryRejectsStaleWalletArchive(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "wallet-archive-stale@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	updated, err := repo.UpdateWallet(context.Background(), userID, wallet.ID, finance.UpdateWalletInput{BaseVersion: wallet.Version, Name: "Ví mới", IncludeInTotal: ptrBool(true)})
	if err != nil {
		t.Fatalf("update wallet: %v", err)
	}
	if err := repo.ArchiveWallet(context.Background(), userID, wallet.ID, wallet.Version); !errors.Is(err, finance.ErrConflict) {
		t.Fatalf("stale wallet archive error = %v, want conflict", err)
	}
	if err := repo.ArchiveWallet(context.Background(), userID, wallet.ID, updated.Version); err != nil {
		t.Fatalf("current-version wallet archive: %v", err)
	}
}

func TestRepositoryRejectsStaleWalletUpdate(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "wallet-stale@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)

	updated, err := repo.UpdateWallet(context.Background(), userID, wallet.ID, finance.UpdateWalletInput{BaseVersion: wallet.Version, Name: "Ví mới", IncludeInTotal: ptrBool(true)})
	if err != nil {
		t.Fatalf("first wallet update: %v", err)
	}
	_, err = repo.UpdateWallet(context.Background(), userID, wallet.ID, finance.UpdateWalletInput{BaseVersion: wallet.Version, Name: "Ghi đè cũ", IncludeInTotal: ptrBool(false)})
	if !errors.Is(err, finance.ErrConflict) {
		t.Fatalf("expected stale wallet conflict after version %d became %d, got %v", wallet.Version, updated.Version, err)
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

	updated, err := repo.UpdateCategory(context.Background(), userID, custom.ID, finance.UpdateCategoryInput{Name: "Cafe", BaseVersion: custom.Version})
	if err != nil {
		t.Fatalf("update category: %v", err)
	}
	if updated.Name != "Cafe" {
		t.Fatalf("expected updated category name, got %#v", updated)
	}
	if _, err := repo.UpdateCategory(context.Background(), userID, systemID, finance.UpdateCategoryInput{Name: "Tên mới"}); !errors.Is(err, finance.ErrSystemCategoryLocked) {
		t.Fatalf("expected system category lock, got %v", err)
	}
	if err := repo.ArchiveCategory(context.Background(), userID, systemID, 1); !errors.Is(err, finance.ErrSystemCategoryLocked) {
		t.Fatalf("expected system category archive lock, got %v", err)
	}
	if err := repo.ArchiveCategory(context.Background(), userID, custom.ID, updated.Version); err != nil {
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

func TestRepositoryIdempotencyHashIncludesReceiptReference(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "idempotent-receipt@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "expense_food")
	receipt := func(key string) finance.ReceiptObject {
		object, err := repo.CreateReceiptObject(context.Background(), userID, finance.CreateReceiptObjectInput{
			ObjectKey: key, ContentType: "image/jpeg", SizeBytes: 10, ChecksumSHA256: strings.Repeat("a", 64), OriginalFilename: key + ".jpg",
		})
		if err != nil {
			t.Fatalf("create receipt %s: %v", key, err)
		}
		return object
	}
	firstReceipt := receipt("receipt-first")
	secondReceipt := receipt("receipt-second")
	base := finance.CreateTransactionInput{
		IdempotencyKey: "idem-receipt", Type: finance.TransactionExpense, SourceWalletID: wallet.ID,
		CategoryID: categoryID, ReceiptObjectID: firstReceipt.ID, AmountVND: 10_000, OccurredAt: fixedFinanceTime(),
	}
	if _, err := repo.CreateTransaction(context.Background(), userID, base); err != nil {
		t.Fatalf("create first transaction: %v", err)
	}
	base.ReceiptObjectID = secondReceipt.ID
	if _, err := repo.CreateTransaction(context.Background(), userID, base); !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("reused key with a different receipt error = %v, want validation", err)
	}
	assertWalletBalance(t, conn, wallet.ID, -10_000, 2)
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

func TestRepositoryUpdatesTransactionByReversingAndReapplyingEffect(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "tx-edit@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	incomeCategoryID := findSystemCategory(t, conn, "income_salary")
	expenseCategoryID := findSystemCategory(t, conn, "expense_food")
	created, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "edit-base-1",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     incomeCategoryID,
		AmountVND:      500_000,
		OccurredAt:     fixedFinanceTime(),
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	updated, err := repo.UpdateTransaction(context.Background(), userID, created.ID, finance.UpdateTransactionInput{
		BaseVersion:         created.Version,
		Type:                finance.TransactionExpense,
		SourceWalletID:      wallet.ID,
		CategoryID:          expenseCategoryID,
		AmountVND:           125_000,
		OccurredAt:          fixedFinanceTime().Add(3 * time.Hour),
		Note:                "Bữa tối",
		ExcludedFromReports: true,
	})
	if err != nil {
		t.Fatalf("update transaction: %v", err)
	}

	assertWalletBalance(t, conn, wallet.ID, -125_000, 3)
	if updated.Type != finance.TransactionExpense || updated.AmountVND != 125_000 || updated.BalanceAfterVND != -125_000 || updated.Version != 2 {
		t.Fatalf("updated transaction has wrong accounting state: %#v", updated)
	}
	if !updated.ExcludedFromReports || updated.Note != "Bữa tối" {
		t.Fatalf("updated transaction metadata not persisted: %#v", updated)
	}
}

func TestRepositoryRejectsStaleTransactionUpdateWithoutChangingBalance(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "tx-stale-update@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	incomeCategoryID := findSystemCategory(t, conn, "income_salary")
	created, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "stale-update-base",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     incomeCategoryID,
		AmountVND:      100_000,
		OccurredAt:     fixedFinanceTime(),
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	first, err := repo.UpdateTransaction(context.Background(), userID, created.ID, finance.UpdateTransactionInput{
		BaseVersion:    created.Version,
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     incomeCategoryID,
		AmountVND:      200_000,
		OccurredAt:     fixedFinanceTime(),
	})
	if err != nil {
		t.Fatalf("first update: %v", err)
	}
	_, err = repo.UpdateTransaction(context.Background(), userID, created.ID, finance.UpdateTransactionInput{
		BaseVersion:    created.Version,
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     incomeCategoryID,
		AmountVND:      300_000,
		OccurredAt:     fixedFinanceTime(),
	})
	if !errors.Is(err, finance.ErrConflict) {
		t.Fatalf("stale update error = %v, want conflict", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 200_000, 3)
	if first.Version != created.Version+1 {
		t.Fatalf("first update version = %d, want %d", first.Version, created.Version+1)
	}
}

func TestRepositoryArchivesTransactionByReversingEffectOnce(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "tx-archive@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "income_salary")
	created, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "archive-base-1",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     categoryID,
		AmountVND:      200_000,
		OccurredAt:     fixedFinanceTime(),
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	if err := repo.ArchiveTransaction(context.Background(), userID, created.ID, created.Version); err != nil {
		t.Fatalf("archive transaction: %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 0, 3)
	if err := repo.ArchiveTransaction(context.Background(), userID, created.ID, created.Version); !errors.Is(err, finance.ErrForbidden) {
		t.Fatalf("expected second archive to be forbidden, got %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 0, 3)
}

func TestRepositoryRejectsArchiveWhenReversalWouldOverflowBalance(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "tx-archive-overflow@example.com")
	wallet, err := repo.CreateWallet(context.Background(), userID, finance.CreateWalletInput{
		Name:       "Ví sát biên",
		Type:       finance.WalletCash,
		BalanceVND: math.MaxInt64 - 1,
	})
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	expenseCategoryID := findSystemCategory(t, conn, "expense_food")
	incomeCategoryID := findSystemCategory(t, conn, "income_salary")
	expense, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "archive-overflow-expense",
		Type:           finance.TransactionExpense,
		SourceWalletID: wallet.ID,
		CategoryID:     expenseCategoryID,
		AmountVND:      1,
		OccurredAt:     fixedFinanceTime(),
	})
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}
	if _, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{
		IdempotencyKey: "archive-overflow-income",
		Type:           finance.TransactionIncome,
		SourceWalletID: wallet.ID,
		CategoryID:     incomeCategoryID,
		AmountVND:      2,
		OccurredAt:     fixedFinanceTime(),
	}); err != nil {
		t.Fatalf("create income: %v", err)
	}

	err = repo.ArchiveTransaction(context.Background(), userID, expense.ID, expense.Version)
	if !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("archive overflow error = %v, want validation", err)
	}
	assertWalletBalance(t, conn, wallet.ID, math.MaxInt64, 3)
	assertFinanceRowCount(t, conn, `SELECT count(*) FROM transactions WHERE id = $1 AND archived_at IS NULL`, expense.ID, 1)
}

func TestRepositoryRejectsStaleTransactionArchive(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userID := createFinanceUser(t, conn, "tx-archive-stale@example.com")
	wallet := createFinanceWallet(t, repo, userID, "Tiền mặt", finance.WalletCash)
	categoryID := findSystemCategory(t, conn, "income_salary")
	created, err := repo.CreateTransaction(context.Background(), userID, finance.CreateTransactionInput{IdempotencyKey: "archive-stale", Type: finance.TransactionIncome, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 200_000, OccurredAt: fixedFinanceTime()})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	updated, err := repo.UpdateTransaction(context.Background(), userID, created.ID, finance.UpdateTransactionInput{BaseVersion: created.Version, Type: finance.TransactionIncome, SourceWalletID: wallet.ID, CategoryID: categoryID, AmountVND: 300_000, OccurredAt: fixedFinanceTime()})
	if err != nil {
		t.Fatalf("update transaction: %v", err)
	}
	if err := repo.ArchiveTransaction(context.Background(), userID, created.ID, created.Version); !errors.Is(err, finance.ErrConflict) {
		t.Fatalf("stale archive error = %v, want conflict", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 300_000, 3)
	if err := repo.ArchiveTransaction(context.Background(), userID, created.ID, updated.Version); err != nil {
		t.Fatalf("current-version archive: %v", err)
	}
	assertWalletBalance(t, conn, wallet.ID, 0, 4)
}

func TestRepositoryListsTransactionsWithFilters(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	userA := createFinanceUser(t, conn, "tx-list-a@example.com")
	userB := createFinanceUser(t, conn, "tx-list-b@example.com")
	walletA := createFinanceWallet(t, repo, userA, "Tiền mặt", finance.WalletCash)
	walletB := createFinanceWallet(t, repo, userA, "Ngân hàng", finance.WalletBank)
	otherWallet := createFinanceWallet(t, repo, userB, "Other", finance.WalletCash)
	incomeCategoryID := findSystemCategory(t, conn, "income_salary")
	expenseCategoryID := findSystemCategory(t, conn, "expense_food")
	base := fixedFinanceTime()

	coffee, err := repo.CreateTransaction(context.Background(), userA, finance.CreateTransactionInput{
		IdempotencyKey: "list-coffee",
		Type:           finance.TransactionExpense,
		SourceWalletID: walletA.ID,
		CategoryID:     expenseCategoryID,
		AmountVND:      45_000,
		OccurredAt:     base.Add(time.Hour),
		Note:           "Cafe sáng",
	})
	if err != nil {
		t.Fatalf("create coffee: %v", err)
	}
	_, err = repo.CreateTransaction(context.Background(), userA, finance.CreateTransactionInput{
		IdempotencyKey: "list-income",
		Type:           finance.TransactionIncome,
		SourceWalletID: walletB.ID,
		CategoryID:     incomeCategoryID,
		AmountVND:      1_000_000,
		OccurredAt:     base.Add(2 * time.Hour),
		Note:           "Lương",
	})
	if err != nil {
		t.Fatalf("create income: %v", err)
	}
	excluded, err := repo.CreateTransaction(context.Background(), userA, finance.CreateTransactionInput{
		IdempotencyKey:      "list-excluded",
		Type:                finance.TransactionExpense,
		SourceWalletID:      walletA.ID,
		CategoryID:          expenseCategoryID,
		AmountVND:           20_000,
		OccurredAt:          base.Add(3 * time.Hour),
		Note:                "Cafe excluded",
		ExcludedFromReports: true,
	})
	if err != nil {
		t.Fatalf("create excluded: %v", err)
	}
	_, err = repo.CreateTransaction(context.Background(), userB, finance.CreateTransactionInput{
		IdempotencyKey: "list-other-user",
		Type:           finance.TransactionIncome,
		SourceWalletID: otherWallet.ID,
		CategoryID:     incomeCategoryID,
		AmountVND:      9_000_000,
		OccurredAt:     base.Add(4 * time.Hour),
		Note:           "Hidden",
	})
	if err != nil {
		t.Fatalf("create other user transaction: %v", err)
	}

	excludedOnly := true
	got, err := repo.ListTransactions(context.Background(), userA, finance.TransactionFilters{
		WalletID:             walletA.ID,
		CategoryID:           expenseCategoryID,
		Type:                 finance.TransactionExpense,
		DateFrom:             ptrTime(base),
		DateTo:               ptrTime(base.Add(4 * time.Hour)),
		Query:                "cafe",
		ExcludedFromReports:  &excludedOnly,
		IncludeArchivedItems: false,
	})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(got) != 1 || got[0].ID != excluded.ID {
		t.Fatalf("expected only excluded cafe transaction, got %#v; coffee=%s", got, coffee.ID)
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

func ptrTime(value time.Time) *time.Time {
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
