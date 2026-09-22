package repository

import (
	"errors"
	"time"

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
