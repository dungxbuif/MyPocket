package finance_test

import (
	"context"
	"errors"
	"testing"

	"mypocket/internal/finance"
)

func TestRepositoryCreatesAndUpdatesOwnedCategoryHierarchy(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	owner := createFinanceUser(t, conn, "hierarchy-owner@example.com")
	parent, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{
		Kind: finance.CategoryExpense,
		Name: "Sinh hoạt",
	})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	child, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{
		Kind:     finance.CategoryExpense,
		Name:     "Tiền điện",
		ParentID: parent.ID,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	if child.ParentID != parent.ID {
		t.Fatalf("child parent = %q, want %q", child.ParentID, parent.ID)
	}

	clearedParent := ""
	updated, err := repo.UpdateCategory(context.Background(), owner, child.ID, finance.UpdateCategoryInput{
		Name:        "Điện nước",
		ParentID:    &clearedParent,
		BaseVersion: child.Version,
	})
	if err != nil {
		t.Fatalf("clear child parent: %v", err)
	}
	if updated.Name != "Điện nước" || updated.ParentID != "" || updated.Version != child.Version+1 {
		t.Fatalf("unexpected updated category: %#v", updated)
	}
}

func TestRepositoryRejectsInvalidCategoryParents(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	owner := createFinanceUser(t, conn, "hierarchy-validation-owner@example.com")
	other := createFinanceUser(t, conn, "hierarchy-validation-other@example.com")
	foreignParent, err := repo.CreateCategory(context.Background(), other, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Foreign"})
	if err != nil {
		t.Fatalf("create foreign parent: %v", err)
	}
	archivedParent, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Archived"})
	if err != nil {
		t.Fatalf("create archived parent: %v", err)
	}
	if err := repo.ArchiveCategory(context.Background(), owner, archivedParent.ID, archivedParent.Version); err != nil {
		t.Fatalf("archive parent: %v", err)
	}
	incomeParent, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{Kind: finance.CategoryIncome, Name: "Income"})
	if err != nil {
		t.Fatalf("create income parent: %v", err)
	}
	expenseParent, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Expense"})
	if err != nil {
		t.Fatalf("create expense parent: %v", err)
	}
	child, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Child", ParentID: expenseParent.ID})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	cases := []struct {
		name    string
		input   finance.CreateCategoryInput
		wantErr error
	}{
		{name: "foreign", input: finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Child", ParentID: foreignParent.ID}, wantErr: finance.ErrForbidden},
		{name: "archived", input: finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Child", ParentID: archivedParent.ID}, wantErr: finance.ErrValidation},
		{name: "kind mismatch", input: finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Child", ParentID: incomeParent.ID}, wantErr: finance.ErrValidation},
		{name: "third level", input: finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Grandchild", ParentID: child.ID}, wantErr: finance.ErrValidation},
		{name: "self", input: finance.CreateCategoryInput{ID: "2eb9f171-9d13-486d-9ebf-722dd36b8cb3", Kind: finance.CategoryExpense, Name: "Self", ParentID: "2eb9f171-9d13-486d-9ebf-722dd36b8cb3"}, wantErr: finance.ErrValidation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := repo.CreateCategory(context.Background(), owner, tc.input); !errors.Is(err, tc.wantErr) {
				t.Fatalf("create error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestRepositoryRejectsCategoryCyclesStaleVersionAndSystemMutation(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	owner := createFinanceUser(t, conn, "hierarchy-update-owner@example.com")
	parent, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Parent"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	child, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Child", ParentID: parent.ID})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	if _, err := repo.UpdateCategory(context.Background(), owner, parent.ID, finance.UpdateCategoryInput{Name: parent.Name, ParentID: &child.ID, BaseVersion: parent.Version}); !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("descendant cycle error = %v, want validation", err)
	}
	if _, err := repo.UpdateCategory(context.Background(), owner, child.ID, finance.UpdateCategoryInput{Name: child.Name, ParentID: &child.ID, BaseVersion: child.Version}); !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("self cycle error = %v, want validation", err)
	}
	if _, err := repo.UpdateCategory(context.Background(), owner, child.ID, finance.UpdateCategoryInput{Name: "Renamed", BaseVersion: child.Version + 1}); !errors.Is(err, finance.ErrConflict) {
		t.Fatalf("stale version error = %v, want conflict", err)
	}

	systemID := findSystemCategory(t, conn, "expense_food")
	if _, err := repo.UpdateCategory(context.Background(), owner, systemID, finance.UpdateCategoryInput{Name: "Locked", BaseVersion: 1}); !errors.Is(err, finance.ErrSystemCategoryLocked) {
		t.Fatalf("system mutation error = %v, want system lock", err)
	}
}

func TestRepositoryListsWalletCategorySettingsWithActiveDefaultAndOwnerIsolation(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	owner := createFinanceUser(t, conn, "settings-owner@example.com")
	other := createFinanceUser(t, conn, "settings-other@example.com")
	wallet := createFinanceWallet(t, repo, owner, "Tiền mặt", finance.WalletBasic)
	category, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Cafe"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	settings, err := repo.ListWalletCategorySettings(context.Background(), owner, wallet.ID)
	if err != nil {
		t.Fatalf("list default settings: %v", err)
	}
	if setting, ok := findCategorySetting(settings, category.ID); !ok || !setting.Active {
		t.Fatalf("expected owned category active by default, got %#v", setting)
	}
	if err := repo.SetWalletCategoryActive(context.Background(), owner, wallet.ID, category.ID, false); err != nil {
		t.Fatalf("disable category: %v", err)
	}
	settings, err = repo.ListWalletCategorySettings(context.Background(), owner, wallet.ID)
	if err != nil {
		t.Fatalf("list updated settings: %v", err)
	}
	if setting, ok := findCategorySetting(settings, category.ID); !ok || setting.Active {
		t.Fatalf("expected owned category inactive after update, got %#v", setting)
	}
	if _, err := repo.ListWalletCategorySettings(context.Background(), other, wallet.ID); !errors.Is(err, finance.ErrForbidden) {
		t.Fatalf("foreign wallet settings error = %v, want forbidden", err)
	}
}

func TestRepositoryArchiveRetainsReferencedCategoryHistory(t *testing.T) {
	conn := migratedFinancePostgres(t)
	repo := finance.NewRepository(conn)
	owner := createFinanceUser(t, conn, "category-history-owner@example.com")
	wallet := createFinanceWallet(t, repo, owner, "Tiền mặt", finance.WalletBasic)
	category, err := repo.CreateCategory(context.Background(), owner, finance.CreateCategoryInput{Kind: finance.CategoryExpense, Name: "Cafe"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	transaction, err := repo.CreateTransaction(context.Background(), owner, finance.CreateTransactionInput{
		IdempotencyKey: "category-history-transaction",
		Type:           finance.TransactionExpense,
		SourceWalletID: wallet.ID,
		CategoryID:     category.ID,
		AmountVND:      45000,
		OccurredAt:     fixedFinanceTime(),
	})
	if err != nil {
		t.Fatalf("create referenced transaction: %v", err)
	}
	if err := repo.ArchiveCategory(context.Background(), owner, category.ID, category.Version+1); !errors.Is(err, finance.ErrConflict) {
		t.Fatalf("expected stale category archive conflict, got %v", err)
	}
	if err := repo.ArchiveCategory(context.Background(), owner, category.ID, category.Version); err != nil {
		t.Fatalf("archive referenced category: %v", err)
	}

	var categoryID string
	if err := conn.QueryRowContext(context.Background(), `SELECT category_id::text FROM transactions WHERE id = $1`, transaction.ID).Scan(&categoryID); err != nil {
		t.Fatalf("load transaction category reference: %v", err)
	}
	if categoryID != category.ID {
		t.Fatalf("transaction category = %q, want %q", categoryID, category.ID)
	}
	categories, err := repo.ListCategories(context.Background(), owner)
	if err != nil {
		t.Fatalf("list categories after archive: %v", err)
	}
	if containsCategoryID(categories, category.ID) {
		t.Fatalf("archived category remained visible: %#v", categories)
	}
}

func findCategorySetting(settings []finance.WalletCategorySetting, categoryID string) (finance.WalletCategorySetting, bool) {
	for _, setting := range settings {
		if setting.Category.ID == categoryID {
			return setting, true
		}
	}
	return finance.WalletCategorySetting{}, false
}
