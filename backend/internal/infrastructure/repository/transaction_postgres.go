package repository

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	transactionrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionPostgresRepository struct{ db *gorm.DB }

func NewTransactionPostgresRepository(db *gorm.DB) transactionrepo.TransactionRepository {
	return &TransactionPostgresRepository{db: db}
}

func (r *TransactionPostgresRepository) List(ownerID string) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	if err := r.db.Where("owner_id = ?", ownerID).Order("occurred_at DESC, created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	needsNames := false
	for _, item := range transactions {
		if item.JarID != nil {
			needsNames = true
			break
		}
	}
	if !needsNames {
		return transactions, nil
	}
	var user entity.User
	if err := r.db.Where("id = ?", ownerID).First(&user).Error; err != nil {
		return nil, err
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		return nil, err
	}
	var configs []entity.JarMonthConfig
	if err := r.db.Where("owner_id = ?", ownerID).Find(&configs).Error; err != nil {
		return nil, err
	}
	nameByConfig := make(map[string]string, len(configs))
	for _, config := range configs {
		nameByConfig[config.Month.String()[:7]+":"+config.JarID] = config.Name
	}
	for i := range transactions {
		if transactions[i].JarID != nil {
			month := transactions[i].OccurredAt.In(location).Format("2006-01")
			transactions[i].JarName = nameByConfig[month+":"+*transactions[i].JarID]
		}
	}
	return transactions, nil
}

func (r *TransactionPostgresRepository) Find(ownerID, id string) (*entity.Transaction, error) {
	var transaction entity.Transaction
	if err := r.db.Where("id = ? AND owner_id = ?", id, ownerID).First(&transaction).Error; err != nil {
		return nil, err
	}
	if err := r.fillJarName(ownerID, &transaction); err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionPostgresRepository) Create(transaction *entity.Transaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := validateJarAssignment(tx, transaction.OwnerID, transaction.JarID, transaction.OccurredAt, nil); err != nil {
			return err
		}
		return tx.Create(transaction).Error
	})
}

func (r *TransactionPostgresRepository) CreateAdjustment(transaction *entity.Transaction) error {
	if transaction == nil || transaction.OwnerID == "" || transaction.WalletID == "" || transaction.Amount <= 0 || transaction.Type != entity.TransactionTypeAdjustment || transaction.AdjustmentDirection == nil || (*transaction.AdjustmentDirection != entity.AdjustmentDirectionIncrease && *transaction.AdjustmentDirection != entity.AdjustmentDirectionDecrease) {
		return transactionrepo.ErrAdjustmentInvalid
	}
	transaction.CategoryID = nil
	transaction.JarID = nil
	transaction.TransferID = nil
	transaction.IncludedInReports = false
	return r.db.Transaction(func(tx *gorm.DB) error {
		var wallet entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", transaction.WalletID, transaction.OwnerID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return transactionrepo.ErrAdjustmentWalletInvalid
			}
			return err
		}
		if wallet.Type == entity.WalletTypeCredit {
			return transactionrepo.ErrAdjustmentWalletInvalid
		}
		insert := `INSERT INTO transactions (id, owner_id, wallet_id, type, amount, adjustment_direction, occurred_at, note, included_in_reports) VALUES (?, ?, ?, ?, ?, ?, ?, ?, false)`
		if err := tx.Exec(insert, transaction.ID, transaction.OwnerID, transaction.WalletID, transaction.Type, transaction.Amount, transaction.AdjustmentDirection, transaction.OccurredAt, transaction.Note).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", transaction.ID).First(transaction).Error
	})
}

func (r *TransactionPostgresRepository) CreateCreditEntry(ownerID string, input transactionrepo.CreditEntryInput) (*entity.Transaction, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(input.WalletID) == "" || input.Amount <= 0 || input.OccurredAt.IsZero() {
		return nil, transactionrepo.ErrCreditInvalid
	}
	typeForKind := map[string]string{
		entity.CreditKindPurchase: entity.TransactionTypeExpense,
		entity.CreditKindFee:      entity.TransactionTypeExpense,
		entity.CreditKindInterest: entity.TransactionTypeExpense,
		entity.CreditKindRefund:   entity.TransactionTypeIncome,
	}
	txnType, ok := typeForKind[input.Kind]
	if !ok {
		return nil, transactionrepo.ErrCreditInvalid
	}
	created := &entity.Transaction{ID: uuid.NewString(), OwnerID: ownerID, WalletID: input.WalletID, CategoryID: input.CategoryID, Type: txnType, Amount: input.Amount, CreditKind: &input.Kind, OccurredAt: input.OccurredAt.UTC(), Note: input.Note, IncludedInReports: true}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var wallet entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", input.WalletID, ownerID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return transactionrepo.ErrCreditWalletInvalid
			}
			return err
		}
		if wallet.Type != entity.WalletTypeCredit || wallet.CreditLimit == nil || *wallet.CreditLimit <= 0 {
			return transactionrepo.ErrCreditWalletInvalid
		}
		if input.CategoryID != nil {
			var category entity.Category
			if err := tx.Where("id = ? AND (is_system = true OR owner_id = ?)", *input.CategoryID, ownerID).First(&category).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return transactionrepo.ErrCreditInvalid
				}
				return err
			}
			if category.Kind != txnType {
				return transactionrepo.ErrCreditInvalid
			}
		}
		insert := `INSERT INTO transactions (id, owner_id, wallet_id, category_id, type, amount, credit_kind, occurred_at, note, included_in_reports) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, true)`
		if err := tx.Exec(insert, created.ID, created.OwnerID, created.WalletID, created.CategoryID, created.Type, created.Amount, created.CreditKind, created.OccurredAt, created.Note).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND owner_id = ?", created.ID, ownerID).First(created).Error
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (r *TransactionPostgresRepository) CreateCreditPayment(ownerID, creditWalletID, sourceWalletID string, amount int64, occurredAt time.Time, note *string) ([]entity.Transaction, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(creditWalletID) == "" || strings.TrimSpace(sourceWalletID) == "" || creditWalletID == sourceWalletID || amount <= 0 || occurredAt.IsZero() {
		return nil, transactionrepo.ErrCreditPaymentInvalid
	}
	var result []entity.Transaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var wallets []entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id IN ?", ownerID, []string{creditWalletID, sourceWalletID}).Find(&wallets).Error; err != nil {
			return err
		}
		if len(wallets) != 2 {
			return transactionrepo.ErrCreditPaymentInvalid
		}
		var creditWallet, sourceWallet *entity.Wallet
		for i := range wallets {
			wallet := &wallets[i]
			if wallet.ID == creditWalletID {
				creditWallet = wallet
			} else if wallet.ID == sourceWalletID {
				sourceWallet = wallet
			}
		}
		if creditWallet == nil || sourceWallet == nil || creditWallet.Type != entity.WalletTypeCredit || creditWallet.CreditLimit == nil || *creditWallet.CreditLimit <= 0 || sourceWallet.Type == entity.WalletTypeCredit {
			return transactionrepo.ErrCreditWalletInvalid
		}
		var categories []entity.Category
		if err := tx.Where("is_system = true AND system_key IN ?", []string{"expense_transfer_out", "income_transfer_in"}).Find(&categories).Error; err != nil {
			return err
		}
		categoryIDs := map[string]string{}
		for _, category := range categories {
			if category.SystemKey != nil {
				categoryIDs[*category.SystemKey] = category.ID
			}
		}
		if categoryIDs["expense_transfer_out"] == "" || categoryIDs["income_transfer_in"] == "" {
			return transactionrepo.ErrCreditPaymentInvalid
		}
		expenseCategoryID := categoryIDs["expense_transfer_out"]
		incomeCategoryID := categoryIDs["income_transfer_in"]
		paymentID := uuid.NewString()
		timestamp := occurredAt.UTC()
		source := entity.Transaction{ID: uuid.NewString(), OwnerID: ownerID, WalletID: sourceWalletID, CategoryID: &expenseCategoryID, Type: entity.TransactionTypeExpense, Amount: amount, OccurredAt: timestamp, Note: note, IncludedInReports: false, TransferID: &paymentID}
		creditKind := entity.CreditKindPayment
		credit := entity.Transaction{ID: uuid.NewString(), OwnerID: ownerID, WalletID: creditWalletID, CategoryID: &incomeCategoryID, Type: entity.TransactionTypeIncome, Amount: amount, CreditKind: &creditKind, CreditPaymentID: &paymentID, OccurredAt: timestamp, Note: note, IncludedInReports: false, TransferID: &paymentID}
		insert := `INSERT INTO transactions (id, owner_id, wallet_id, category_id, type, amount, credit_kind, credit_payment_id, occurred_at, note, included_in_reports, transfer_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, false, ?)`
		if err := tx.Exec(insert, source.ID, source.OwnerID, source.WalletID, source.CategoryID, source.Type, source.Amount, source.CreditKind, source.CreditPaymentID, source.OccurredAt, source.Note, source.TransferID).Error; err != nil {
			return err
		}
		if err := tx.Exec(insert, credit.ID, credit.OwnerID, credit.WalletID, credit.CategoryID, credit.Type, credit.Amount, credit.CreditKind, credit.CreditPaymentID, credit.OccurredAt, credit.Note, credit.TransferID).Error; err != nil {
			return err
		}
		if err := tx.Where("id IN ?", []string{source.ID, credit.ID}).Order("wallet_id ASC").Find(&result).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *TransactionPostgresRepository) ListCreditStatement(ownerID, walletID string, from, to *time.Time) (entity.CreditStatement, error) {
	statement := entity.CreditStatement{WalletID: walletID, Items: []entity.Transaction{}}
	var wallet entity.Wallet
	if err := r.db.Where("id = ? AND owner_id = ?", walletID, ownerID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return statement, transactionrepo.ErrCreditWalletInvalid
		}
		return statement, err
	}
	if wallet.Type != entity.WalletTypeCredit || wallet.CreditLimit == nil {
		return statement, transactionrepo.ErrCreditWalletInvalid
	}
	statement.CreditLimit = *wallet.CreditLimit
	statement.LastStatementBalance = wallet.LastStatementBalance
	statement.StatementDay = wallet.StatementDay
	statement.PaymentDueDay = wallet.PaymentDueDay
	statement.PaymentStatus = "not_configured"
	statement.Balance = wallet.OpeningBalance
	query := r.db.Where("owner_id = ? AND wallet_id = ? AND credit_kind IS NOT NULL", ownerID, walletID)
	if from != nil {
		query = query.Where("occurred_at >= ?", from.UTC())
	}
	if to != nil {
		query = query.Where("occurred_at < ?", to.UTC())
	}
	if err := query.Order("occurred_at DESC, created_at DESC").Find(&statement.Items).Error; err != nil {
		return statement, err
	}
	var all []entity.Transaction
	if err := r.db.Where("owner_id = ? AND wallet_id = ?", ownerID, walletID).Order("occurred_at ASC, created_at ASC").Find(&all).Error; err != nil {
		return statement, err
	}
	for _, row := range all {
		if row.Type == entity.TransactionTypeIncome {
			statement.Balance += row.Amount
		} else if row.Type == entity.TransactionTypeExpense {
			statement.Balance -= row.Amount
		}
	}
	statement.AvailableCredit = statement.CreditLimit + statement.Balance
	statement.AmountDue, statement.PaymentStatus, statement.PaymentDueAt = creditPaymentSummary(wallet, statement.Balance, time.Now().UTC(), r.db)
	return statement, nil
}

func creditPaymentSummary(wallet entity.Wallet, balance int64, now time.Time, db *gorm.DB) (int64, string, *time.Time) {
	if wallet.LastStatementBalance == nil || *wallet.LastStatementBalance <= 0 {
		return 0, "not_configured", nil
	}
	amountDue := *wallet.LastStatementBalance
	outstanding := -balance
	if outstanding < 0 {
		outstanding = 0
	}
	if outstanding < amountDue {
		amountDue = outstanding
	}
	if outstanding == 0 {
		return 0, "paid", nil
	}
	location := time.UTC
	var user entity.User
	if db != nil && db.Where("id = ?", wallet.OwnerID).First(&user).Error == nil {
		if loaded, err := time.LoadLocation(user.Timezone); err == nil && user.Timezone != "Local" {
			location = loaded
		}
	}
	localNow := now.In(location)
	dueDay := 0
	if wallet.PaymentDueDay != nil {
		dueDay = *wallet.PaymentDueDay
	}
	var dueAt *time.Time
	status := "due"
	if dueDay > 0 {
		lastDay := time.Date(localNow.Year(), localNow.Month()+1, 0, 0, 0, 0, 0, location).Day()
		if dueDay > lastDay {
			dueDay = lastDay
		}
		due := time.Date(localNow.Year(), localNow.Month(), dueDay, 23, 59, 59, 0, location)
		utc := due.UTC()
		dueAt = &utc
		if localNow.After(due) {
			status = "overdue"
		}
	}
	if amountDue < *wallet.LastStatementBalance {
		status = "partial"
	}
	return amountDue, status, dueAt
}

func (r *TransactionPostgresRepository) BulkDelete(ownerID string, ids []string) error {
	if ownerID == "" || len(ids) == 0 || len(ids) > 100 {
		return transactionrepo.ErrBulkDeleteInvalid
	}
	seen := make(map[string]struct{}, len(ids))
	for _, rawID := range ids {
		id := strings.TrimSpace(rawID)
		if id == "" {
			return transactionrepo.ErrBulkDeleteInvalid
		}
		if _, exists := seen[id]; exists {
			return transactionrepo.ErrBulkDeleteInvalid
		}
		seen[id] = struct{}{}
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var rows []entity.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id IN ?", ownerID, ids).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) != len(ids) {
			return transactionrepo.ErrBulkDeleteNotFound
		}
		for _, row := range rows {
			if row.TransferID != nil {
				return transactionrepo.ErrBulkDeleteLinked
			}
			if row.Type == entity.TransactionTypeAdjustment {
				return transactionrepo.ErrBulkDeleteAdjustment
			}
			if row.CreditKind != nil {
				return transactionrepo.ErrBulkDeleteCredit
			}
		}
		result := tx.Where("owner_id = ? AND id IN ?", ownerID, ids).Delete(&entity.Transaction{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(len(ids)) {
			return transactionrepo.ErrBulkDeleteNotFound
		}
		return nil
	})
}

func (r *TransactionPostgresRepository) CreateTransfer(ownerID string, source, destination *entity.Transaction) error {
	if source == nil || destination == nil || source.TransferID == nil || destination.TransferID == nil || *source.TransferID != *destination.TransferID {
		return transactionrepo.ErrTransferInvalid
	}
	if source.OwnerID != ownerID || destination.OwnerID != ownerID || source.WalletID == destination.WalletID {
		return transactionrepo.ErrTransferInvalid
	}
	if source.Type != entity.TransactionTypeExpense || destination.Type != entity.TransactionTypeIncome || source.Amount <= 0 || destination.Amount != source.Amount || source.JarID != nil || destination.JarID != nil || source.IncludedInReports || destination.IncludedInReports {
		return transactionrepo.ErrTransferInvalid
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var wallets []entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id IN ?", ownerID, []string{source.WalletID, destination.WalletID}).Find(&wallets).Error; err != nil {
			return err
		}
		if len(wallets) != 2 {
			return transactionrepo.ErrTransferWalletInvalid
		}
		for _, wallet := range wallets {
			if wallet.Type == entity.WalletTypeCredit {
				return transactionrepo.ErrTransferWalletInvalid
			}
		}
		var categories []entity.Category
		if err := tx.Where("is_system = true AND system_key IN ?", []string{"expense_transfer_out", "income_transfer_in"}).Find(&categories).Error; err != nil {
			return err
		}
		categoryIDs := map[string]string{}
		for _, category := range categories {
			if category.SystemKey != nil {
				categoryIDs[*category.SystemKey] = category.ID
			}
		}
		if source.CategoryID == nil || destination.CategoryID == nil || categoryIDs["expense_transfer_out"] != *source.CategoryID || categoryIDs["income_transfer_in"] != *destination.CategoryID {
			return transactionrepo.ErrTransferInvalid
		}
		// GORM's `default:true` tag omits a false zero value even when selected.
		// Use a parameterized insert so the transfer contract explicitly writes
		// included_in_reports=false, then hydrate timestamps for the response.
		insert := `INSERT INTO transactions (id, owner_id, wallet_id, category_id, type, amount, occurred_at, note, included_in_reports, transfer_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, false, ?)`
		for _, row := range []*entity.Transaction{source, destination} {
			if err := tx.Exec(insert, row.ID, row.OwnerID, row.WalletID, row.CategoryID, row.Type, row.Amount, row.OccurredAt, row.Note, row.TransferID).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("id = ?", source.ID).First(source).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", destination.ID).First(destination).Error
	})
}

func (r *TransactionPostgresRepository) UpdateTransfer(ownerID, transferID string, updates transactionrepo.TransferUpdate) ([]entity.Transaction, error) {
	if ownerID == "" || transferID == "" || updates.Amount <= 0 || updates.OccurredAt.IsZero() {
		return nil, transactionrepo.ErrTransferInvalid
	}
	var result []entity.Transaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = r.updateTransferPair(tx, ownerID, transferID, updates)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *TransactionPostgresRepository) DeleteTransfer(ownerID, transferID string) error {
	if ownerID == "" || transferID == "" {
		return transactionrepo.ErrTransferInvalid
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.deleteTransferPair(tx, ownerID, transferID)
	})
}

func (r *TransactionPostgresRepository) UpdateTransferIdempotent(ownerID, transferID string, updates transactionrepo.TransferUpdate, idempotencyKey, requestHash string) ([]entity.Transaction, bool, error) {
	if idempotencyKey == "" || requestHash == "" {
		return nil, false, transactionrepo.ErrTransferInvalid
	}
	var result []entity.Transaction
	replayed := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mutation, created, err := r.reserveMutation(tx, ownerID, "transfer.patch", idempotencyKey, requestHash)
		if err != nil {
			return err
		}
		if !created {
			if err := json.Unmarshal([]byte(mutation.ResponseJSON), &result); err != nil {
				return err
			}
			replayed = true
			return nil
		}
		result, err = r.updateTransferPair(tx, ownerID, transferID, updates)
		if err != nil {
			return err
		}
		return r.completeMutation(tx, mutation.ID, result)
	})
	return result, replayed, err
}

func (r *TransactionPostgresRepository) DeleteTransferIdempotent(ownerID, transferID, idempotencyKey, requestHash string) (bool, error) {
	if idempotencyKey == "" || requestHash == "" {
		return false, transactionrepo.ErrTransferInvalid
	}
	replayed := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mutation, created, err := r.reserveMutation(tx, ownerID, "transfer.delete", idempotencyKey, requestHash)
		if err != nil {
			return err
		}
		if !created {
			replayed = true
			return nil
		}
		if err := r.deleteTransferPair(tx, ownerID, transferID); err != nil {
			return err
		}
		return r.completeMutation(tx, mutation.ID, []entity.Transaction{})
	})
	return replayed, err
}

func (r *TransactionPostgresRepository) updateTransferPair(tx *gorm.DB, ownerID, transferID string, updates transactionrepo.TransferUpdate) ([]entity.Transaction, error) {
	if ownerID == "" || transferID == "" || updates.Amount <= 0 || updates.OccurredAt.IsZero() {
		return nil, transactionrepo.ErrTransferInvalid
	}
	rows, source, destination, err := r.lockTransferPair(tx, ownerID, transferID)
	if err != nil {
		return nil, err
	}
	if err := r.validateTransferWallets(tx, ownerID, source, destination); err != nil {
		return nil, err
	}
	if source.Amount != destination.Amount || source.Type != entity.TransactionTypeExpense || destination.Type != entity.TransactionTypeIncome || source.IncludedInReports || destination.IncludedInReports || source.JarID != nil || destination.JarID != nil {
		return nil, transactionrepo.ErrTransferPairInvalid
	}
	updated := tx.Model(&entity.Transaction{}).Where("owner_id = ? AND transfer_id = ?", ownerID, transferID).Updates(map[string]any{"amount": updates.Amount, "occurred_at": updates.OccurredAt, "note": updates.Note})
	if updated.Error != nil {
		return nil, updated.Error
	}
	if updated.RowsAffected != int64(len(rows)) {
		return nil, transactionrepo.ErrTransferPairInvalid
	}
	result := []entity.Transaction{*source, *destination}
	for i := range result {
		result[i].Amount = updates.Amount
		result[i].OccurredAt = updates.OccurredAt
		result[i].Note = updates.Note
	}
	return result, nil
}

func (r *TransactionPostgresRepository) deleteTransferPair(tx *gorm.DB, ownerID, transferID string) error {
	if ownerID == "" || transferID == "" {
		return transactionrepo.ErrTransferInvalid
	}
	rows, source, destination, err := r.lockTransferPair(tx, ownerID, transferID)
	if err != nil {
		return err
	}
	if err := r.validateTransferWallets(tx, ownerID, source, destination); err != nil {
		return err
	}
	if source.Amount != destination.Amount || source.Type != entity.TransactionTypeExpense || destination.Type != entity.TransactionTypeIncome || source.IncludedInReports || destination.IncludedInReports || source.JarID != nil || destination.JarID != nil {
		return transactionrepo.ErrTransferPairInvalid
	}
	deleted := tx.Where("owner_id = ? AND transfer_id = ?", ownerID, transferID).Delete(&entity.Transaction{})
	if deleted.Error != nil {
		return deleted.Error
	}
	if deleted.RowsAffected != int64(len(rows)) {
		return transactionrepo.ErrTransferPairInvalid
	}
	return nil
}

func (r *TransactionPostgresRepository) reserveMutation(tx *gorm.DB, ownerID, operation, key, requestHash string) (entity.TransactionMutation, bool, error) {
	mutation := entity.TransactionMutation{ID: uuid.NewString(), OwnerID: ownerID, Operation: operation, IdempotencyKey: key, RequestHash: requestHash, Status: "pending", ResponseJSON: "{}"}
	result := tx.Exec(`INSERT INTO transaction_mutations (id, owner_id, operation, idempotency_key, request_hash, status, response_json) VALUES (?, ?, ?, ?, ?, ?, ?::jsonb) ON CONFLICT (owner_id, operation, idempotency_key) DO NOTHING`, mutation.ID, mutation.OwnerID, mutation.Operation, mutation.IdempotencyKey, mutation.RequestHash, mutation.Status, mutation.ResponseJSON)
	if result.Error != nil {
		return entity.TransactionMutation{}, false, result.Error
	}
	if result.RowsAffected == 1 {
		return mutation, true, nil
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND operation = ? AND idempotency_key = ?", ownerID, operation, key).First(&mutation).Error; err != nil {
		return entity.TransactionMutation{}, false, err
	}
	if mutation.RequestHash != requestHash {
		return entity.TransactionMutation{}, false, transactionrepo.ErrMutationConflict
	}
	if mutation.Status != transactionrepo.TransactionMutationCompleted {
		return entity.TransactionMutation{}, false, transactionrepo.ErrMutationPending
	}
	return mutation, false, nil
}

func (r *TransactionPostgresRepository) completeMutation(tx *gorm.DB, id string, response any) error {
	encoded, err := json.Marshal(response)
	if err != nil {
		return err
	}
	result := tx.Model(&entity.TransactionMutation{}).Where("id = ?", id).Updates(map[string]any{"status": transactionrepo.TransactionMutationCompleted, "response_json": string(encoded)})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return transactionrepo.ErrMutationPending
	}
	return nil
}

func (r *TransactionPostgresRepository) lockTransferPair(tx *gorm.DB, ownerID, transferID string) ([]entity.Transaction, *entity.Transaction, *entity.Transaction, error) {
	var rows []entity.Transaction
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND transfer_id = ?", ownerID, transferID).Order("type ASC, id ASC")
	if err := query.Find(&rows).Error; err != nil {
		return nil, nil, nil, err
	}
	if len(rows) == 0 {
		return nil, nil, nil, transactionrepo.ErrTransferNotFound
	}
	if len(rows) != 2 {
		return nil, nil, nil, transactionrepo.ErrTransferPairInvalid
	}
	var source, destination *entity.Transaction
	for i := range rows {
		row := rows[i]
		switch row.Type {
		case entity.TransactionTypeExpense:
			if source != nil {
				return nil, nil, nil, transactionrepo.ErrTransferPairInvalid
			}
			source = &row
		case entity.TransactionTypeIncome:
			if destination != nil {
				return nil, nil, nil, transactionrepo.ErrTransferPairInvalid
			}
			destination = &row
		default:
			return nil, nil, nil, transactionrepo.ErrTransferPairInvalid
		}
	}
	if source == nil || destination == nil || source.TransferID == nil || destination.TransferID == nil || *source.TransferID != transferID || *destination.TransferID != transferID {
		return nil, nil, nil, transactionrepo.ErrTransferPairInvalid
	}
	return rows, source, destination, nil
}

func (r *TransactionPostgresRepository) validateTransferWallets(tx *gorm.DB, ownerID string, source, destination *entity.Transaction) error {
	if source.WalletID == destination.WalletID {
		return transactionrepo.ErrTransferWalletInvalid
	}
	var wallets []entity.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id IN ?", ownerID, []string{source.WalletID, destination.WalletID}).Find(&wallets).Error; err != nil {
		return err
	}
	if len(wallets) != 2 {
		return transactionrepo.ErrTransferWalletInvalid
	}
	for _, wallet := range wallets {
		if wallet.Type == entity.WalletTypeCredit {
			return transactionrepo.ErrTransferWalletInvalid
		}
	}
	var categories []entity.Category
	if err := tx.Where("is_system = true AND system_key IN ?", []string{"expense_transfer_out", "income_transfer_in"}).Find(&categories).Error; err != nil {
		return err
	}
	categoryIDs := map[string]string{}
	for _, category := range categories {
		if category.SystemKey != nil {
			categoryIDs[*category.SystemKey] = category.ID
		}
	}
	if source.CategoryID == nil || destination.CategoryID == nil || categoryIDs["expense_transfer_out"] != *source.CategoryID || categoryIDs["income_transfer_in"] != *destination.CategoryID {
		return transactionrepo.ErrTransferPairInvalid
	}
	return nil
}

func (r *TransactionPostgresRepository) Update(ownerID, id string, updates map[string]any) (*entity.Transaction, error) {
	var updated entity.Transaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing entity.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND owner_id = ?", id, ownerID).First(&existing).Error; err != nil {
			return err
		}
		next := existing
		if value, ok := updates["jar_id"]; ok {
			next.JarID = optionalString(value)
		}
		if value, ok := updates["occurred_at"].(time.Time); ok {
			next.OccurredAt = value
		}
		if err := validateJarAssignment(tx, ownerID, next.JarID, next.OccurredAt, &existing); err != nil {
			return err
		}
		result := tx.Model(&entity.Transaction{}).Where("id = ? AND owner_id = ?", id, ownerID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("id = ? AND owner_id = ?", id, ownerID).First(&updated).Error
	})
	if err != nil {
		return nil, err
	}
	if err := r.fillJarName(ownerID, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *TransactionPostgresRepository) Delete(ownerID, id string) error {
	result := r.db.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&entity.Transaction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("transaction not found")
	}
	return nil
}

func validateJarAssignment(tx *gorm.DB, owner string, jarID *string, occurredAt time.Time, existing *entity.Transaction) error {
	if jarID == nil {
		return nil
	}
	if existing != nil && existing.JarID != nil && *existing.JarID == *jarID && existing.OccurredAt.Equal(occurredAt) {
		return nil
	}
	var user entity.User
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", owner).First(&user).Error; err != nil {
		return err
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil || user.Timezone == "Local" {
		return transactionrepo.ErrJarInvalid
	}
	monthKey := occurredAt.In(location).Format("2006-01")
	month, err := entity.ParseMonth(monthKey)
	if err != nil {
		return transactionrepo.ErrJarInvalid
	}
	query := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("owner_id = ? AND jar_id = ? AND month = ?", owner, *jarID, entity.CalendarDate{Time: month})
	query = query.Where("active = true")
	var config entity.JarMonthConfig
	if err := query.First(&config).Error; err != nil {
		return transactionrepo.ErrJarInvalid
	}
	return nil
}

func optionalString(value any) *string {
	switch typed := value.(type) {
	case nil:
		return nil
	case *string:
		return typed
	case string:
		return &typed
	default:
		return nil
	}
}

func (r *TransactionPostgresRepository) fillJarName(owner string, transaction *entity.Transaction) error {
	if transaction.JarID == nil {
		return nil
	}
	var user entity.User
	if err := r.db.Where("id = ?", owner).First(&user).Error; err != nil {
		return err
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		return err
	}
	month := transaction.OccurredAt.In(location).Format("2006-01")
	monthDate, err := entity.ParseMonth(month)
	if err != nil {
		return err
	}
	var config entity.JarMonthConfig
	if err := r.db.Where("owner_id = ? AND jar_id = ? AND month = ?", owner, *transaction.JarID, entity.CalendarDate{Time: monthDate}).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	transaction.JarName = config.Name
	return nil
}
