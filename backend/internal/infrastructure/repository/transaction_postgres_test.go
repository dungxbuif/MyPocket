package repository

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	transactionrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestValidateJarAssignmentPreservesUnchangedHistoricalLink(t *testing.T) {
	instant := time.Date(2026, time.February, 1, 0, 30, 0, 0, time.UTC)
	jarID := "jar-stable-id"
	existing := &entity.Transaction{JarID: &jarID, OccurredAt: instant}
	// A timezone change can re-bucket the same UTC instant into a month with no
	// config. Saving unrelated fields must not need a database month lookup.
	if err := validateJarAssignment(nil, "owner", &jarID, instant, existing); err != nil {
		t.Fatalf("unchanged stable-jar link should be preserved without a new-month config: %v", err)
	}
}

func TestCreditPaymentSummaryTracksDuePartialAndOverdue(t *testing.T) {
	statementBalance := int64(250000)
	dueDay := 5
	wallet := entity.Wallet{OwnerID: "owner", LastStatementBalance: &statementBalance, PaymentDueDay: &dueDay}
	_, status, dueAt := creditPaymentSummary(wallet, -250000, time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC), nil)
	if status != "due" || dueAt == nil {
		t.Fatalf("expected due status, got %q at=%v", status, dueAt)
	}
	amount, status, _ := creditPaymentSummary(wallet, -100000, time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC), nil)
	if amount != 100000 || status != "partial" {
		t.Fatalf("expected partial amount/status, got %d/%q", amount, status)
	}
	_, status, _ = creditPaymentSummary(wallet, -250000, time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC), nil)
	if status != "overdue" {
		t.Fatalf("expected overdue status, got %q", status)
	}
}

func TestCreditLedgerKeepsSignedBalanceAndAtomicPaymentPair(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	if err := db.Create(&entity.User{ID: owner, Email: owner + "@credit.test", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh"}).Error; err != nil {
		t.Fatal(err)
	}
	limit := int64(1000000)
	creditWallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Thẻ", Type: entity.WalletTypeCredit, Currency: entity.WalletCurrencyVND, CreditLimit: &limit}
	sourceWallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Thanh toán", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}
	if err := db.Create(&creditWallet).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&sourceWallet).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ?", owner).Delete(&entity.Transaction{})
		db.Where("owner_id = ?", owner).Delete(&entity.Wallet{})
		db.Where("id = ?", owner).Delete(&entity.User{})
	})
	repo := &TransactionPostgresRepository{db: db}
	created, err := repo.CreateCreditEntry(owner, transactionrepo.CreditEntryInput{WalletID: creditWallet.ID, Kind: entity.CreditKindPurchase, Amount: 100, OccurredAt: time.Now().UTC()})
	if err != nil || created == nil || created.CreditKind == nil || *created.CreditKind != entity.CreditKindPurchase {
		t.Fatalf("create credit entry: row=%#v err=%v", created, err)
	}
	statement, err := repo.ListCreditStatement(owner, creditWallet.ID, nil, nil)
	if err != nil || statement.Balance != -100 || statement.AvailableCredit != 999900 || len(statement.Items) != 1 {
		t.Fatalf("unexpected statement after purchase: %#v err=%v", statement, err)
	}
	rows, err := repo.CreateCreditPayment(owner, creditWallet.ID, sourceWallet.ID, 50, time.Now().UTC(), nil)
	if err != nil || len(rows) != 2 {
		t.Fatalf("expected atomic payment pair, rows=%#v err=%v", rows, err)
	}
	statement, err = repo.ListCreditStatement(owner, creditWallet.ID, nil, nil)
	if err != nil || statement.Balance != -50 || statement.AvailableCredit != 999950 || len(statement.Items) != 2 {
		t.Fatalf("unexpected statement after payment: %#v err=%v", statement, err)
	}
}

func TestCreateTransferPersistsExactlyTwoLinkedRows(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	user := entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh"}
	if err = db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	sourceWallet, destinationWallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Nguồn", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}, entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Đích", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}
	if err = db.Create(&sourceWallet).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&destinationWallet).Error; err != nil {
		t.Fatal(err)
	}
	var categories []entity.Category
	if err = db.Where("is_system = true AND system_key IN ?", []string{"expense_transfer_out", "income_transfer_in"}).Find(&categories).Error; err != nil || len(categories) != 2 {
		t.Fatalf("expected seeded transfer categories, got %d, err=%v", len(categories), err)
	}
	byKey := map[string]string{}
	for _, category := range categories {
		if category.SystemKey != nil {
			byKey[*category.SystemKey] = category.ID
		}
	}
	transferID := uuid.NewString()
	sourceCategory, destinationCategory := byKey["expense_transfer_out"], byKey["income_transfer_in"]
	source := &entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: sourceWallet.ID, CategoryID: &sourceCategory, Type: entity.TransactionTypeExpense, Amount: 7500, OccurredAt: time.Now().UTC(), IncludedInReports: false, TransferID: &transferID}
	destination := &entity.Transaction{ID: uuid.NewString(), OwnerID: owner, WalletID: destinationWallet.ID, CategoryID: &destinationCategory, Type: entity.TransactionTypeIncome, Amount: 7500, OccurredAt: source.OccurredAt, IncludedInReports: false, TransferID: &transferID}
	t.Cleanup(func() {
		db.Where("owner_id = ?", owner).Delete(&entity.Transaction{})
		db.Where("owner_id = ?", owner).Delete(&entity.Wallet{})
		db.Where("id = ?", owner).Delete(&entity.User{})
	})
	if err = NewTransactionPostgresRepository(db).(interface {
		CreateTransfer(string, *entity.Transaction, *entity.Transaction) error
	}).CreateTransfer(owner, source, destination); err != nil {
		t.Fatalf("create transfer: %v", err)
	}
	var rows []entity.Transaction
	if err = db.Where("transfer_id = ?", transferID).Order("type").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Amount != 7500 || rows[1].Amount != 7500 {
		t.Fatalf("expected two linked rows, got %+v", rows)
	}
	for _, row := range rows {
		if row.IncludedInReports || row.TransferID == nil || *row.TransferID != transferID {
			t.Fatalf("invalid transfer row: %+v", row)
		}
	}
}

func TestUpdateTransferMutatesBothRowsAndRejectsWrongOwnerOrMalformedPair(t *testing.T) {
	db, repo, owner, transferID, sourceID, destinationID := transferFixture(t)
	updatedAt := time.Date(2026, time.September, 25, 5, 0, 0, 0, time.UTC)
	note := "đã cập nhật"
	rows, err := repo.UpdateTransfer(owner, transferID, transactionrepo.TransferUpdate{Amount: 9100, OccurredAt: updatedAt, Note: &note})
	if err != nil {
		t.Fatalf("update transfer: %v", err)
	}
	if len(rows) != 2 || rows[0].Amount != 9100 || rows[1].Amount != 9100 || rows[0].Note == nil || *rows[0].Note != note {
		t.Fatalf("expected both rows updated, got %+v", rows)
	}
	if err := repo.DeleteTransfer("another-owner", transferID); !errors.Is(err, transactionrepo.ErrTransferNotFound) {
		t.Fatalf("wrong owner should not find transfer, got %v", err)
	}
	if err := db.Where("id = ?", destinationID).Delete(&entity.Transaction{}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateTransfer(owner, transferID, transactionrepo.TransferUpdate{Amount: 1, OccurredAt: updatedAt}); !errors.Is(err, transactionrepo.ErrTransferPairInvalid) {
		t.Fatalf("malformed pair should be rejected, got %v", err)
	}
	var source entity.Transaction
	if err := db.Where("id = ?", sourceID).First(&source).Error; err != nil {
		t.Fatal(err)
	}
	if source.Amount != 9100 {
		t.Fatalf("malformed update should not mutate remaining row: %+v", source)
	}
}

func TestDeleteTransferRemovesBothRowsAndRejectsMalformedPair(t *testing.T) {
	db, repo, owner, transferID, _, _ := transferFixture(t)
	if err := repo.DeleteTransfer(owner, transferID); err != nil {
		t.Fatalf("delete transfer: %v", err)
	}
	var count int64
	if err := db.Model(&entity.Transaction{}).Where("owner_id = ? AND transfer_id = ?", owner, transferID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected both transfer rows deleted, count=%d", count)
	}

	_, repo, owner, transferID, sourceID, destinationID := transferFixture(t)
	if err := db.Where("id = ?", destinationID).Delete(&entity.Transaction{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteTransfer(owner, transferID); !errors.Is(err, transactionrepo.ErrTransferPairInvalid) {
		t.Fatalf("malformed pair should be rejected, got %v", err)
	}
	if err := db.Model(&entity.Transaction{}).Where("id = ?", sourceID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("malformed delete should not remove remaining row, count=%d", count)
	}
}

func transferFixture(t *testing.T) (*gorm.DB, *TransactionPostgresRepository, string, string, string, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	user := entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner, Timezone: "Asia/Ho_Chi_Minh"}
	if err = db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	sourceWallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Nguồn", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}
	destinationWallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "Đích", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}
	if err = db.Create(&sourceWallet).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&destinationWallet).Error; err != nil {
		t.Fatal(err)
	}
	var categories []entity.Category
	if err = db.Where("is_system = true AND system_key IN ?", []string{"expense_transfer_out", "income_transfer_in"}).Find(&categories).Error; err != nil || len(categories) != 2 {
		t.Fatalf("expected seeded transfer categories, got %d, err=%v", len(categories), err)
	}
	byKey := map[string]string{}
	for _, category := range categories {
		if category.SystemKey != nil {
			byKey[*category.SystemKey] = category.ID
		}
	}
	transferID := uuid.NewString()
	sourceID, destinationID := uuid.NewString(), uuid.NewString()
	sourceCategory, destinationCategory := byKey["expense_transfer_out"], byKey["income_transfer_in"]
	source := &entity.Transaction{ID: sourceID, OwnerID: owner, WalletID: sourceWallet.ID, CategoryID: &sourceCategory, Type: entity.TransactionTypeExpense, Amount: 7500, OccurredAt: time.Now().UTC(), IncludedInReports: false, TransferID: &transferID}
	destination := &entity.Transaction{ID: destinationID, OwnerID: owner, WalletID: destinationWallet.ID, CategoryID: &destinationCategory, Type: entity.TransactionTypeIncome, Amount: 7500, OccurredAt: source.OccurredAt, IncludedInReports: false, TransferID: &transferID}
	repo := NewTransactionPostgresRepository(db).(*TransactionPostgresRepository)
	if err = repo.CreateTransfer(owner, source, destination); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ?", owner).Delete(&entity.Transaction{})
		db.Where("owner_id = ?", owner).Delete(&entity.Wallet{})
		db.Where("id = ?", owner).Delete(&entity.User{})
	})
	return db, repo, owner, transferID, sourceID, destinationID
}
