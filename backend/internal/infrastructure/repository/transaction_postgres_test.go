package repository

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
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
