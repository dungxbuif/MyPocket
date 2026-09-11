package finance

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"mypocket/internal/platform/changefeed"
	"mypocket/internal/platform/commandtx"
	"sort"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *commandtx.Handle
}

func NewRepositoryInTx(tx *sql.Tx) *Repository { return &Repository{db: commandtx.Bound(tx)} }

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: commandtx.New(db)}
}

func (r *Repository) CreateReceiptObject(ctx context.Context, userID string, input CreateReceiptObjectInput) (ReceiptObject, error) {
	input, err := ValidateCreateReceiptObject(input)
	if err != nil {
		return ReceiptObject{}, err
	}
	var receipt ReceiptObject
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO receipt_objects (user_id, object_key, content_type, size_bytes, checksum_sha256, original_filename)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, user_id::text, object_key, content_type, size_bytes, checksum_sha256, original_filename, created_at
	`, userID, input.ObjectKey, input.ContentType, input.SizeBytes, input.ChecksumSHA256, input.OriginalFilename).Scan(
		&receipt.ID, &receipt.UserID, &receipt.ObjectKey, &receipt.ContentType, &receipt.SizeBytes, &receipt.ChecksumSHA256, &receipt.OriginalFilename, &receipt.CreatedAt,
	)
	if err != nil {
		return ReceiptObject{}, fmt.Errorf("create receipt object: %w", err)
	}
	return receipt, nil
}

func (r *Repository) GetReceiptObject(ctx context.Context, userID, receiptID string) (ReceiptObject, error) {
	var receipt ReceiptObject
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, object_key, content_type, size_bytes, checksum_sha256, original_filename, created_at
		FROM receipt_objects WHERE id = $1 AND user_id = $2
	`, receiptID, userID).Scan(&receipt.ID, &receipt.UserID, &receipt.ObjectKey, &receipt.ContentType, &receipt.SizeBytes, &receipt.ChecksumSHA256, &receipt.OriginalFilename, &receipt.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ReceiptObject{}, ErrForbidden
	}
	if err != nil {
		return ReceiptObject{}, fmt.Errorf("get receipt object: %w", err)
	}
	return receipt, nil
}

func (r *Repository) commandCreateWallet(ctx context.Context, userID string, input CreateWalletInput) (Wallet, error) {
	input, err := ValidateCreateWallet(input)
	if err != nil {
		return Wallet{}, err
	}
	var wallet Wallet
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO wallets (
			id,
			user_id,
			name,
			type,
			balance_vnd,
			include_in_total,
			credit_limit_vnd,
			statement_day,
			payment_due_day
		)
		VALUES (coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, coalesce($6, true), $7, $8, $9)
		RETURNING id::text, user_id::text, name, type, balance_vnd, include_in_total, is_default_ai, version
	`,
		input.ID,
		userID,
		input.Name,
		string(input.Type),
		input.BalanceVND,
		input.IncludeInTotal,
		input.CreditLimitVND,
		input.StatementDay,
		input.PaymentDueDay,
	).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Type,
		&wallet.BalanceVND,
		&wallet.IncludeInTotal,
		&wallet.IsDefaultAI,
		&wallet.Version,
	)
	if err != nil {
		return Wallet{}, fmt.Errorf("create wallet: %w", err)
	}
	return wallet, nil
}

func (r *Repository) ListWallets(ctx context.Context, userID string) ([]Wallet, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, name, type, balance_vnd, include_in_total, is_default_ai, version
		FROM wallets
		WHERE user_id = $1 AND archived_at IS NULL
		ORDER BY created_at, id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list wallets: %w", err)
	}
	defer rows.Close()

	var wallets []Wallet
	for rows.Next() {
		var wallet Wallet
		if err := rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.Name,
			&wallet.Type,
			&wallet.BalanceVND,
			&wallet.IncludeInTotal,
			&wallet.IsDefaultAI,
			&wallet.Version,
		); err != nil {
			return nil, fmt.Errorf("scan wallet: %w", err)
		}
		wallets = append(wallets, wallet)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("wallet rows: %w", err)
	}
	return wallets, nil
}

func (r *Repository) commandUpdateWallet(ctx context.Context, userID string, walletID string, input UpdateWalletInput) (Wallet, error) {
	input, err := ValidateUpdateWallet(input)
	if err != nil {
		return Wallet{}, err
	}
	if input.BaseVersion <= 0 {
		return Wallet{}, fmt.Errorf("%w: base version is required", ErrValidation)
	}
	includeInTotal := true
	if input.IncludeInTotal != nil {
		includeInTotal = *input.IncludeInTotal
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Wallet{}, fmt.Errorf("begin update wallet: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var currentVersion int64
	err = tx.QueryRowContext(ctx, `
		SELECT version FROM wallets
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		FOR UPDATE
	`, walletID, userID).Scan(&currentVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return Wallet{}, ErrForbidden
	}
	if err != nil {
		return Wallet{}, fmt.Errorf("lock wallet update: %w", err)
	}
	if input.BaseVersion != currentVersion {
		return Wallet{}, ErrConflict
	}

	var wallet Wallet
	err = tx.QueryRowContext(ctx, `
		UPDATE wallets
		SET name = $3, include_in_total = $4, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		RETURNING id::text, user_id::text, name, type, balance_vnd, include_in_total, is_default_ai, version
	`, walletID, userID, input.Name, includeInTotal).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Name,
		&wallet.Type,
		&wallet.BalanceVND,
		&wallet.IncludeInTotal,
		&wallet.IsDefaultAI,
		&wallet.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Wallet{}, ErrForbidden
	}
	if err != nil {
		return Wallet{}, fmt.Errorf("update wallet: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Wallet{}, fmt.Errorf("commit update wallet: %w", err)
	}
	return wallet, nil
}

func (r *Repository) commandArchiveWallet(ctx context.Context, userID string, walletID string, baseVersion int64) error {
	if baseVersion <= 0 {
		return fmt.Errorf("%w: base version is required", ErrValidation)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE wallets
		SET archived_at = now(), is_default_ai = false, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $3
	`, walletID, userID, baseVersion)
	if err != nil {
		return fmt.Errorf("archive wallet: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive wallet rows affected: %w", err)
	}
	if affected == 0 {
		var active bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM wallets WHERE id = $1 AND user_id = $2 AND archived_at IS NULL)`, walletID, userID).Scan(&active); err != nil {
			return fmt.Errorf("check archived wallet version: %w", err)
		}
		if active {
			return ErrConflict
		}
		return ErrForbidden
	}
	return nil
}

func (r *Repository) commandSetDefaultAIWallet(ctx context.Context, userID string, walletID string, baseVersion int64) error {
	if baseVersion <= 0 {
		return fmt.Errorf("%w: base version is required", ErrValidation)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin set default wallet: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var currentVersion int64
	if err := tx.QueryRowContext(ctx, `
		SELECT version FROM wallets
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		FOR UPDATE
	`, walletID, userID).Scan(&currentVersion); errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	} else if err != nil {
		return fmt.Errorf("check wallet owner: %w", err)
	}
	if baseVersion != currentVersion {
		return ErrConflict
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET is_default_ai = false, updated_at = now(), version = version + 1
		WHERE user_id = $1 AND id <> $2 AND archived_at IS NULL AND is_default_ai = true
	`, userID, walletID); err != nil {
		return fmt.Errorf("clear default wallet: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET is_default_ai = true, updated_at = now(), version = version + 1
		WHERE user_id = $1 AND id = $2 AND archived_at IS NULL AND is_default_ai = false
	`, userID, walletID); err != nil {
		return fmt.Errorf("set default wallet: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit set default wallet: %w", err)
	}
	return nil
}

func (r *Repository) commandCreateCategory(ctx context.Context, userID string, input CreateCategoryInput) (Category, error) {
	input, err := ValidateCreateCategory(input)
	if err != nil {
		return Category{}, err
	}
	if input.ID != "" && input.ParentID == input.ID {
		return Category{}, fmt.Errorf("%w: category cannot be its own parent", ErrValidation)
	}
	if input.ParentID != "" {
		if err := validateCategoryParent(ctx, r.db, userID, "", input.ParentID, input.Kind); err != nil {
			return Category{}, err
		}
	}

	var category Category
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO categories (id, user_id, parent_id, kind, name, is_system)
		VALUES (coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, nullif($3, '')::uuid, $4, $5, false)
		RETURNING id::text, user_id::text, coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system, version
	`, input.ID, userID, input.ParentID, string(input.Kind), input.Name).Scan(
		&category.ID,
		&category.UserID,
		&category.ParentID,
		&category.Kind,
		&category.Name,
		&category.SystemKey,
		&category.IsSystem,
		&category.Version,
	)
	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}
	return category, nil
}

func (r *Repository) ListCategories(ctx context.Context, userID string) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, coalesce(user_id::text, ''), coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system, version
		FROM categories
		WHERE archived_at IS NULL AND (is_system OR user_id = $1)
		ORDER BY is_system DESC, kind, name, id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(
			&category.ID,
			&category.UserID,
			&category.ParentID,
			&category.Kind,
			&category.Name,
			&category.SystemKey,
			&category.IsSystem,
			&category.Version,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category rows: %w", err)
	}
	return categories, nil
}

func (r *Repository) commandUpdateCategory(ctx context.Context, userID string, categoryID string, input UpdateCategoryInput) (Category, error) {
	category, err := r.getCategoryForUser(ctx, userID, categoryID)
	if err != nil {
		return Category{}, err
	}
	if err := ValidateCategoryUpdate(category, input); err != nil {
		return Category{}, err
	}
	if input.BaseVersion <= 0 {
		return Category{}, fmt.Errorf("%w: base version is required", ErrValidation)
	}
	if input.BaseVersion != category.Version {
		return Category{}, ErrConflict
	}
	parentID := category.ParentID
	if input.ParentID != nil {
		parentID = trimmed(*input.ParentID)
	}
	if parentID == categoryID {
		return Category{}, fmt.Errorf("%w: category cannot be its own parent", ErrValidation)
	}
	if parentID != "" {
		if err := validateCategoryParent(ctx, r.db, userID, categoryID, parentID, category.Kind); err != nil {
			return Category{}, err
		}
	}

	err = r.db.QueryRowContext(ctx, `
		UPDATE categories
		SET name = $3, parent_id = nullif($4, '')::uuid, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $5
		RETURNING id::text, user_id::text, coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system, version
	`, categoryID, userID, trimmed(input.Name), parentID, input.BaseVersion).Scan(
		&category.ID,
		&category.UserID,
		&category.ParentID,
		&category.Kind,
		&category.Name,
		&category.SystemKey,
		&category.IsSystem,
		&category.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		if input.BaseVersion > 0 {
			return Category{}, ErrConflict
		}
		return Category{}, ErrForbidden
	}
	if err != nil {
		return Category{}, fmt.Errorf("update category: %w", err)
	}
	return category, nil
}

func (r *Repository) commandArchiveCategory(ctx context.Context, userID string, categoryID string, baseVersion int64) error {
	if baseVersion <= 0 {
		return fmt.Errorf("%w: base version is required", ErrValidation)
	}
	category, err := r.getCategoryForUser(ctx, userID, categoryID)
	if err != nil {
		return err
	}
	if category.IsSystem {
		return ErrSystemCategoryLocked
	}
	if category.Version != baseVersion {
		return ErrConflict
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE categories
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND version = $3
	`, categoryID, userID, baseVersion)
	if err != nil {
		if isForeignKeyViolation(err) {
			return fmt.Errorf("%w: category has history", ErrValidation)
		}
		return fmt.Errorf("archive category: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive category rows affected: %w", err)
	}
	if affected == 0 {
		return ErrConflict
	}
	return nil
}

func (r *Repository) commandSetWalletCategoryActive(ctx context.Context, userID string, walletID string, categoryID string, active bool) error {
	if err := r.requireWalletOwner(ctx, userID, walletID); err != nil {
		return err
	}
	if err := r.requireCategoryVisible(ctx, userID, categoryID); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO wallet_category_settings (wallet_id, category_id, user_id, active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (wallet_id, category_id)
		DO UPDATE SET active = EXCLUDED.active, updated_at = now()
	`, walletID, categoryID, userID, active)
	if err != nil {
		return fmt.Errorf("set wallet category active: %w", err)
	}
	return nil
}

func (r *Repository) ListWalletCategorySettings(ctx context.Context, userID string, walletID string) ([]WalletCategorySetting, error) {
	if err := r.requireWalletOwner(ctx, userID, walletID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			c.id::text,
			coalesce(c.user_id::text, ''),
			coalesce(c.parent_id::text, ''),
			c.kind,
			c.name,
			coalesce(c.system_key, ''),
			c.is_system,
			c.version,
			coalesce(s.active, true)
		FROM categories c
		LEFT JOIN wallet_category_settings s
			ON s.wallet_id = $2 AND s.category_id = c.id AND s.user_id = $1
		WHERE c.archived_at IS NULL AND (c.is_system OR c.user_id = $1)
		ORDER BY c.is_system DESC, c.kind, c.name, c.id
	`, userID, walletID)
	if err != nil {
		return nil, fmt.Errorf("list wallet category settings: %w", err)
	}
	defer rows.Close()

	var settings []WalletCategorySetting
	for rows.Next() {
		var setting WalletCategorySetting
		if err := rows.Scan(
			&setting.Category.ID,
			&setting.Category.UserID,
			&setting.Category.ParentID,
			&setting.Category.Kind,
			&setting.Category.Name,
			&setting.Category.SystemKey,
			&setting.Category.IsSystem,
			&setting.Category.Version,
			&setting.Active,
		); err != nil {
			return nil, fmt.Errorf("scan wallet category setting: %w", err)
		}
		settings = append(settings, setting)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("wallet category setting rows: %w", err)
	}
	return settings, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, userID string, input CreateTransactionInput) (Transaction, error) {
	input, err := ValidateCreateTransaction(input)
	if err != nil {
		return Transaction{}, err
	}
	requestHash, err := transactionRequestHash(input)
	if err != nil {
		return Transaction{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Transaction{}, fmt.Errorf("begin create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	transaction, err := createTransactionInTx(ctx, tx, userID, input, requestHash)
	if err != nil {
		return Transaction{}, err
	}
	if err := tx.Commit(); err != nil {
		return Transaction{}, fmt.Errorf("commit create transaction: %w", err)
	}
	return transaction, nil
}

// CreateTransactionInTx posts a transaction inside a caller-owned SQL
// transaction. The caller is responsible for committing or rolling back tx.
func CreateTransactionInTx(ctx context.Context, tx *sql.Tx, userID string, input CreateTransactionInput) (Transaction, error) {
	input, err := ValidateCreateTransaction(input)
	if err != nil {
		return Transaction{}, err
	}
	requestHash, err := transactionRequestHash(input)
	if err != nil {
		return Transaction{}, err
	}
	return createTransactionInTx(ctx, tx, userID, input, requestHash)
}

func createTransactionInTx(ctx context.Context, tx commandtx.Queryer, userID string, input CreateTransactionInput, requestHash string) (Transaction, error) {
	if err := commandtx.LockUser(ctx, tx, userID); err != nil {
		return Transaction{}, err
	}
	replayed, ok, err := loadIdempotentTransaction(ctx, tx, userID, input.IdempotencyKey, requestHash)
	if err != nil {
		return Transaction{}, err
	}
	if ok {
		return replayed, nil
	}
	if input.ReceiptObjectID != "" {
		if err := requireReceiptOwnership(ctx, tx, userID, input.ReceiptObjectID); err != nil {
			return Transaction{}, err
		}
	}

	balances, err := lockWalletBalances(ctx, tx, userID, []string{input.SourceWalletID, input.DestinationWalletID})
	if err != nil {
		return Transaction{}, err
	}
	if err := requireTransactionCategory(ctx, tx, userID, input.SourceWalletID, input.CategoryID, input.Type); err != nil {
		return Transaction{}, err
	}

	effect, err := ApplyAccountingEffect(AccountingInput{
		Type:                  input.Type,
		AmountVND:             input.AmountVND,
		SourceWalletID:        input.SourceWalletID,
		DestinationWalletID:   input.DestinationWalletID,
		SourceBalanceVND:      balances[input.SourceWalletID],
		DestinationBalanceVND: balances[input.DestinationWalletID],
		TargetBalanceVND:      input.TargetBalanceVND,
	})
	if err != nil {
		return Transaction{}, err
	}

	balances[input.SourceWalletID] = effect.SourceBalanceVND
	if input.Type == TransactionTransfer {
		balances[input.DestinationWalletID] = effect.DestinationBalanceVND
	}
	if err := updateWalletBalances(ctx, tx, userID, balances); err != nil {
		return Transaction{}, err
	}

	transaction, err := insertTransaction(ctx, tx, userID, input, effect)
	if err != nil {
		return Transaction{}, err
	}
	responseJSON, err := json.Marshal(transaction)
	if err != nil {
		return Transaction{}, fmt.Errorf("marshal transaction replay response: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO finance_idempotency_keys (user_id, key, request_hash, response_status, response_json)
		VALUES ($1, $2, $3, 201, $4)
	`, userID, input.IdempotencyKey, requestHash, responseJSON); err != nil {
		return Transaction{}, fmt.Errorf("store transaction idempotency: %w", err)
	}
	if err := changefeed.Append(ctx, tx, userID, "transaction", transaction.ID, "create", transaction.Version, transaction); err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}

func (r *Repository) UpdateTransaction(ctx context.Context, userID string, transactionID string, input UpdateTransactionInput) (Transaction, error) {
	normalized, err := ValidateUpdateTransaction(input)
	if err != nil {
		return Transaction{}, err
	}
	if input.BaseVersion <= 0 {
		return Transaction{}, fmt.Errorf("%w: base version is required", ErrValidation)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Transaction{}, fmt.Errorf("begin update transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := loadActiveTransaction(ctx, tx, userID, transactionID)
	if err != nil {
		return Transaction{}, err
	}
	if input.BaseVersion != current.Version {
		return Transaction{}, ErrConflict
	}
	balances, err := lockWalletBalances(ctx, tx, userID, transactionWalletIDs(current, normalized))
	if err != nil {
		return Transaction{}, err
	}
	if err := requireTransactionCategory(ctx, tx, userID, normalized.SourceWalletID, normalized.CategoryID, normalized.Type); err != nil {
		return Transaction{}, err
	}
	if normalized.ReceiptObjectID != "" {
		if err := requireReceiptOwnership(ctx, tx, userID, normalized.ReceiptObjectID); err != nil {
			return Transaction{}, err
		}
	}

	if err := reverseTransactionEffect(balances, current); err != nil {
		return Transaction{}, err
	}
	effect, err := ApplyAccountingEffect(AccountingInput{
		Type:                  normalized.Type,
		AmountVND:             normalized.AmountVND,
		SourceWalletID:        normalized.SourceWalletID,
		DestinationWalletID:   normalized.DestinationWalletID,
		SourceBalanceVND:      balances[normalized.SourceWalletID],
		DestinationBalanceVND: balances[normalized.DestinationWalletID],
		TargetBalanceVND:      normalized.TargetBalanceVND,
	})
	if err != nil {
		return Transaction{}, err
	}
	balances[normalized.SourceWalletID] = effect.SourceBalanceVND
	if normalized.Type == TransactionTransfer {
		balances[normalized.DestinationWalletID] = effect.DestinationBalanceVND
	}
	if err := updateWalletBalances(ctx, tx, userID, balances); err != nil {
		return Transaction{}, err
	}

	updated, err := updateTransactionRow(ctx, tx, userID, transactionID, normalized, effect)
	if err != nil {
		return Transaction{}, err
	}
	if err := changefeed.Append(ctx, tx, userID, "transaction", updated.ID, "update", updated.Version, updated); err != nil {
		return Transaction{}, err
	}
	if err := tx.Commit(); err != nil {
		return Transaction{}, fmt.Errorf("commit update transaction: %w", err)
	}
	return updated, nil
}

func (r *Repository) ArchiveTransaction(ctx context.Context, userID string, transactionID string, baseVersion int64) error {
	if baseVersion <= 0 {
		return fmt.Errorf("%w: base version is required", ErrValidation)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin archive transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := loadActiveTransaction(ctx, tx, userID, transactionID)
	if err != nil {
		return err
	}
	if baseVersion != current.Version {
		return ErrConflict
	}
	balances, err := lockWalletBalances(ctx, tx, userID, transactionWalletIDs(current, CreateTransactionInput{}))
	if err != nil {
		return err
	}
	if err := reverseTransactionEffect(balances, current); err != nil {
		return err
	}
	if err := updateWalletBalances(ctx, tx, userID, balances); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE transactions
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
	`, transactionID, userID)
	if err != nil {
		return fmt.Errorf("archive transaction: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive transaction rows affected: %w", err)
	}
	if affected == 0 {
		return ErrForbidden
	}
	if err := changefeed.Append(ctx, tx, userID, "transaction", transactionID, "archive", current.Version+1, map[string]any{"id": transactionID, "version": current.Version + 1}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit archive transaction: %w", err)
	}
	return nil
}

func (r *Repository) ListTransactions(ctx context.Context, userID string, filters TransactionFilters) ([]Transaction, error) {
	walletID := trimmed(filters.WalletID)
	categoryID := trimmed(filters.CategoryID)
	txType := string(filters.Type)
	query := "%" + trimmed(filters.Query) + "%"
	if query == "%%" {
		query = ""
	}
	var excluded any
	if filters.ExcludedFromReports != nil {
		excluded = *filters.ExcludedFromReports
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id::text,
			user_id::text,
			type,
			source_wallet_id::text,
			coalesce(destination_wallet_id::text, ''),
			coalesce(category_id::text, ''),
			coalesce(receipt_object_id::text, ''),
			amount_vnd,
			coalesce(balance_after_vnd, 0),
			source_delta_vnd,
			destination_delta_vnd,
			occurred_at,
			note,
			with_person,
			event_ref,
			excluded_from_reports,
			version
		FROM transactions
		WHERE user_id = $1
			AND ($2 = '' OR source_wallet_id::text = $2 OR coalesce(destination_wallet_id::text, '') = $2)
			AND ($3 = '' OR coalesce(category_id::text, '') = $3)
			AND ($4 = '' OR type = $4)
			AND ($5::timestamptz IS NULL OR occurred_at >= $5)
			AND ($6::timestamptz IS NULL OR occurred_at <= $6)
			AND ($7 = '' OR note ILIKE $7 OR with_person ILIKE $7 OR event_ref ILIKE $7)
			AND ($8::boolean IS NULL OR excluded_from_reports = $8)
			AND ($9 OR archived_at IS NULL)
		ORDER BY occurred_at DESC, created_at DESC, id DESC
	`, userID, walletID, categoryID, txType, filters.DateFrom, filters.DateTo, query, excluded, filters.IncludeArchivedItems)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("transaction rows: %w", err)
	}
	return transactions, nil
}

func loadIdempotentTransaction(ctx context.Context, tx commandtx.Queryer, userID string, key string, requestHash string) (Transaction, bool, error) {
	var storedHash string
	var responseJSON []byte
	err := tx.QueryRowContext(ctx, `
		SELECT request_hash, response_json
		FROM finance_idempotency_keys
		WHERE user_id = $1 AND key = $2
		FOR UPDATE
	`, userID, key).Scan(&storedHash, &responseJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, false, nil
	}
	if err != nil {
		return Transaction{}, false, fmt.Errorf("load transaction idempotency: %w", err)
	}
	if storedHash != requestHash {
		return Transaction{}, false, fmt.Errorf("%w: idempotency key reused with different request", ErrValidation)
	}
	var transaction Transaction
	if err := json.Unmarshal(responseJSON, &transaction); err != nil {
		return Transaction{}, false, fmt.Errorf("decode transaction idempotency response: %w", err)
	}
	return transaction, true, nil
}

func loadActiveTransaction(ctx context.Context, tx commandtx.Queryer, userID string, transactionID string) (Transaction, error) {
	if err := commandtx.LockUser(ctx, tx, userID); err != nil {
		return Transaction{}, err
	}
	transaction, err := queryTransaction(ctx, tx, `
		SELECT
			id::text,
			user_id::text,
			type,
			source_wallet_id::text,
			coalesce(destination_wallet_id::text, ''),
			coalesce(category_id::text, ''),
			coalesce(receipt_object_id::text, ''),
			amount_vnd,
			coalesce(balance_after_vnd, 0),
			source_delta_vnd,
			destination_delta_vnd,
			occurred_at,
			note,
			with_person,
			event_ref,
			excluded_from_reports,
			version
		FROM transactions
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		FOR UPDATE
	`, transactionID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, ErrForbidden
	}
	if err != nil {
		return Transaction{}, fmt.Errorf("load transaction: %w", err)
	}
	return transaction, nil
}

// GetTransactionInTx loads an owned transaction, including an archived one,
// inside a caller-owned SQL transaction.
func GetTransactionInTx(ctx context.Context, tx *sql.Tx, userID string, transactionID string) (Transaction, error) {
	transaction, err := queryTransaction(ctx, tx, `
		SELECT
			id::text,
			user_id::text,
			type,
			source_wallet_id::text,
			coalesce(destination_wallet_id::text, ''),
			coalesce(category_id::text, ''),
			coalesce(receipt_object_id::text, ''),
			amount_vnd,
			coalesce(balance_after_vnd, 0),
			source_delta_vnd,
			destination_delta_vnd,
			occurred_at,
			note,
			with_person,
			event_ref,
			excluded_from_reports,
			version
		FROM transactions
		WHERE id = $1 AND user_id = $2
	`, transactionID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, ErrForbidden
	}
	if err != nil {
		return Transaction{}, fmt.Errorf("load transaction: %w", err)
	}
	return transaction, nil
}

func lockWalletBalance(ctx context.Context, tx commandtx.Queryer, userID string, walletID string) (int64, error) {
	var balanceVND int64
	err := tx.QueryRowContext(ctx, `
		SELECT balance_vnd
		FROM wallets
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		FOR UPDATE
	`, walletID, userID).Scan(&balanceVND)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrForbidden
	}
	if err != nil {
		return 0, fmt.Errorf("lock wallet balance: %w", err)
	}
	return balanceVND, nil
}

func lockWalletBalances(ctx context.Context, tx commandtx.Queryer, userID string, walletIDs []string) (map[string]int64, error) {
	unique := make(map[string]struct{}, len(walletIDs))
	for _, walletID := range walletIDs {
		walletID = trimmed(walletID)
		if walletID != "" {
			unique[walletID] = struct{}{}
		}
	}
	ordered := make([]string, 0, len(unique))
	for walletID := range unique {
		ordered = append(ordered, walletID)
	}
	sort.Strings(ordered)

	balances := make(map[string]int64, len(ordered))
	for _, walletID := range ordered {
		balance, err := lockWalletBalance(ctx, tx, userID, walletID)
		if err != nil {
			return nil, err
		}
		balances[walletID] = balance
	}
	return balances, nil
}

func requireTransactionCategory(ctx context.Context, tx commandtx.Queryer, userID string, walletID string, categoryID string, txType TransactionType) error {
	if txType == TransactionTransfer || txType == TransactionAdjustment {
		return nil
	}

	var kind CategoryKind
	err := tx.QueryRowContext(ctx, `
		SELECT kind
		FROM categories
		WHERE id = $1
			AND archived_at IS NULL
			AND (is_system OR user_id = $2)
	`, categoryID, userID).Scan(&kind)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("load transaction category: %w", err)
	}

	if txType == TransactionIncome && kind != CategoryIncome {
		return fmt.Errorf("%w: income requires income category", ErrValidation)
	}
	if txType == TransactionExpense && kind != CategoryExpense {
		return fmt.Errorf("%w: expense requires expense category", ErrValidation)
	}

	return requireWalletCategoryActive(ctx, tx, userID, walletID, categoryID)
}

func requireWalletCategoryActive(ctx context.Context, tx commandtx.Queryer, userID string, walletID string, categoryID string) error {
	var active bool
	err := tx.QueryRowContext(ctx, `
		SELECT active
		FROM wallet_category_settings
		WHERE wallet_id = $1 AND category_id = $2 AND user_id = $3
	`, walletID, categoryID, userID).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load wallet category setting: %w", err)
	}
	if !active {
		return fmt.Errorf("%w: category is inactive for wallet", ErrValidation)
	}
	return nil
}

func updateWalletBalance(ctx context.Context, tx commandtx.Queryer, userID string, walletID string, balanceVND int64) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET balance_vnd = $3, version = version + 1, updated_at = now()
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
	`, walletID, userID, balanceVND)
	if err != nil {
		return fmt.Errorf("update wallet balance: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update wallet balance rows affected: %w", err)
	}
	if affected == 0 {
		return ErrForbidden
	}
	var wallet Wallet
	if err := tx.QueryRowContext(ctx, `SELECT id::text,user_id::text,name,type,balance_vnd,include_in_total,is_default_ai,version FROM wallets WHERE id=$1 AND user_id=$2`, walletID, userID).Scan(&wallet.ID, &wallet.UserID, &wallet.Name, &wallet.Type, &wallet.BalanceVND, &wallet.IncludeInTotal, &wallet.IsDefaultAI, &wallet.Version); err != nil {
		return err
	}
	return changefeed.Append(ctx, tx, userID, "wallet", wallet.ID, "update", wallet.Version, wallet)
}

func updateWalletBalances(ctx context.Context, tx commandtx.Queryer, userID string, balances map[string]int64) error {
	walletIDs := make([]string, 0, len(balances))
	for walletID := range balances {
		walletIDs = append(walletIDs, walletID)
	}
	sort.Strings(walletIDs)
	for _, walletID := range walletIDs {
		if err := updateWalletBalance(ctx, tx, userID, walletID, balances[walletID]); err != nil {
			return err
		}
	}
	return nil
}

func insertTransaction(ctx context.Context, tx commandtx.Queryer, userID string, input CreateTransactionInput, effect AccountingEffect) (Transaction, error) {
	var transaction Transaction
	returning := `
		INSERT INTO transactions (
			id,
			user_id,
			type,
			source_wallet_id,
			 destination_wallet_id,
			 category_id,
			receipt_object_id,
			amount_vnd,
			balance_after_vnd,
			source_delta_vnd,
			destination_delta_vnd,
			note,
			with_person,
			event_ref,
			occurred_at,
			excluded_from_reports
		)
		VALUES (coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, $3, $4, NULLIF($5, '')::uuid, NULLIF($6, '')::uuid, NULLIF($7, '')::uuid, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING
			id::text,
			user_id::text,
			type,
			source_wallet_id::text,
			coalesce(destination_wallet_id::text, ''),
			coalesce(category_id::text, ''),
			coalesce(receipt_object_id::text, ''),
			amount_vnd,
			coalesce(balance_after_vnd, 0),
			source_delta_vnd,
			destination_delta_vnd,
			occurred_at,
			note,
			with_person,
			event_ref,
			excluded_from_reports,
			version
	`
	err := tx.QueryRowContext(ctx, returning,
		input.ID,
		userID,
		string(input.Type),
		input.SourceWalletID,
		input.DestinationWalletID,
		input.CategoryID,
		input.ReceiptObjectID,
		input.AmountVND,
		effect.SourceBalanceVND,
		effect.SourceDeltaVND,
		effect.DestinationDeltaVND,
		input.Note,
		input.WithPerson,
		input.EventRef,
		input.OccurredAt,
		input.ExcludedFromReports,
	).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.Type,
		&transaction.SourceWalletID,
		&transaction.DestinationWalletID,
		&transaction.CategoryID,
		&transaction.ReceiptObjectID,
		&transaction.AmountVND,
		&transaction.BalanceAfterVND,
		&transaction.SourceDeltaVND,
		&transaction.DestinationDeltaVND,
		&transaction.OccurredAt,
		&transaction.Note,
		&transaction.WithPerson,
		&transaction.EventRef,
		&transaction.ExcludedFromReports,
		&transaction.Version,
	)
	if err != nil {
		return Transaction{}, fmt.Errorf("insert transaction: %w", err)
	}
	return transaction, nil
}

func updateTransactionRow(ctx context.Context, tx commandtx.Queryer, userID string, transactionID string, input CreateTransactionInput, effect AccountingEffect) (Transaction, error) {
	transaction, err := queryTransaction(ctx, tx, `
		UPDATE transactions
		SET
			type = $3,
			source_wallet_id = $4,
			destination_wallet_id = NULLIF($5, '')::uuid,
			category_id = NULLIF($6, '')::uuid,
			receipt_object_id = NULLIF($7, '')::uuid,
			amount_vnd = $8,
			balance_after_vnd = $9,
			source_delta_vnd = $10,
			destination_delta_vnd = $11,
			note = $12,
			with_person = $13,
			event_ref = $14,
			occurred_at = $15,
			excluded_from_reports = $16,
			updated_at = now(),
			version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		RETURNING
			id::text,
			user_id::text,
			type,
			source_wallet_id::text,
			coalesce(destination_wallet_id::text, ''),
			coalesce(category_id::text, ''),
			coalesce(receipt_object_id::text, ''),
			amount_vnd,
			coalesce(balance_after_vnd, 0),
			source_delta_vnd,
			destination_delta_vnd,
			occurred_at,
			note,
			with_person,
			event_ref,
			excluded_from_reports,
			version
	`,
		transactionID,
		userID,
		string(input.Type),
		input.SourceWalletID,
		input.DestinationWalletID,
		input.CategoryID,
		input.ReceiptObjectID,
		input.AmountVND,
		effect.SourceBalanceVND,
		effect.SourceDeltaVND,
		effect.DestinationDeltaVND,
		input.Note,
		input.WithPerson,
		input.EventRef,
		input.OccurredAt,
		input.ExcludedFromReports,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, ErrForbidden
	}
	if err != nil {
		return Transaction{}, fmt.Errorf("update transaction row: %w", err)
	}
	return transaction, nil
}

type transactionScanner interface {
	Scan(dest ...any) error
}

func queryTransaction(ctx context.Context, tx commandtx.Queryer, query string, args ...any) (Transaction, error) {
	return scanTransaction(tx.QueryRowContext(ctx, query, args...))
}

func scanTransaction(scanner transactionScanner) (Transaction, error) {
	var transaction Transaction
	err := scanner.Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.Type,
		&transaction.SourceWalletID,
		&transaction.DestinationWalletID,
		&transaction.CategoryID,
		&transaction.ReceiptObjectID,
		&transaction.AmountVND,
		&transaction.BalanceAfterVND,
		&transaction.SourceDeltaVND,
		&transaction.DestinationDeltaVND,
		&transaction.OccurredAt,
		&transaction.Note,
		&transaction.WithPerson,
		&transaction.EventRef,
		&transaction.ExcludedFromReports,
		&transaction.Version,
	)
	if err != nil {
		return Transaction{}, err
	}
	return transaction, nil
}

func requireReceiptOwnership(ctx context.Context, tx commandtx.Queryer, userID, receiptID string) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM receipt_objects WHERE id = $1 AND user_id = $2)`, receiptID, userID).Scan(&exists); err != nil {
		return fmt.Errorf("check receipt ownership: %w", err)
	}
	if !exists {
		return ErrForbidden
	}
	return nil
}

func transactionWalletIDs(current Transaction, next CreateTransactionInput) []string {
	walletIDs := []string{current.SourceWalletID, current.DestinationWalletID}
	if next.SourceWalletID != "" {
		walletIDs = append(walletIDs, next.SourceWalletID)
	}
	if next.DestinationWalletID != "" {
		walletIDs = append(walletIDs, next.DestinationWalletID)
	}
	return walletIDs
}

func reverseTransactionEffect(balances map[string]int64, transaction Transaction) error {
	sourceBalance, ok := checkedSub(balances[transaction.SourceWalletID], transaction.SourceDeltaVND)
	if !ok {
		return fmt.Errorf("%w: reversing transaction would overflow source balance", ErrValidation)
	}
	balances[transaction.SourceWalletID] = sourceBalance
	if transaction.DestinationWalletID != "" {
		destinationBalance, ok := checkedSub(balances[transaction.DestinationWalletID], transaction.DestinationDeltaVND)
		if !ok {
			return fmt.Errorf("%w: reversing transaction would overflow destination balance", ErrValidation)
		}
		balances[transaction.DestinationWalletID] = destinationBalance
	}
	return nil
}

func (r *Repository) getCategoryForUser(ctx context.Context, userID string, categoryID string) (Category, error) {
	var category Category
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, coalesce(user_id::text, ''), coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system, version
		FROM categories
		WHERE id = $1 AND archived_at IS NULL AND (is_system OR user_id = $2)
	`, categoryID, userID).Scan(
		&category.ID,
		&category.UserID,
		&category.ParentID,
		&category.Kind,
		&category.Name,
		&category.SystemKey,
		&category.IsSystem,
		&category.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Category{}, ErrForbidden
	}
	if err != nil {
		return Category{}, fmt.Errorf("load category: %w", err)
	}
	return category, nil
}

type categoryQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func validateCategoryParent(ctx context.Context, queryer categoryQueryer, userID, categoryID, parentID string, kind CategoryKind) error {
	var parent Category
	var archivedAt sql.NullTime
	err := queryer.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, coalesce(parent_id::text, ''), kind, name, is_system, version, archived_at
		FROM categories
		WHERE id::text = $1 AND user_id = $2
	`, parentID, userID).Scan(
		&parent.ID,
		&parent.UserID,
		&parent.ParentID,
		&parent.Kind,
		&parent.Name,
		&parent.IsSystem,
		&parent.Version,
		&archivedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("load category parent: %w", err)
	}
	if archivedAt.Valid {
		return fmt.Errorf("%w: category parent is archived", ErrValidation)
	}
	if parent.Kind != kind {
		return fmt.Errorf("%w: category parent kind must match", ErrValidation)
	}
	if categoryID != "" {
		var cycle bool
		err := queryer.QueryRowContext(ctx, `
			WITH RECURSIVE descendants(id) AS (
				SELECT id
				FROM categories
				WHERE parent_id::text = $1 AND user_id = $2
				UNION
				SELECT c.id
				FROM categories c
				JOIN descendants d ON c.parent_id = d.id
				WHERE c.user_id = $2
			)
			SELECT EXISTS (SELECT 1 FROM descendants WHERE id::text = $3)
		`, categoryID, userID, parentID).Scan(&cycle)
		if err != nil {
			return fmt.Errorf("check category parent cycle: %w", err)
		}
		if cycle {
			return fmt.Errorf("%w: category parent cannot be a descendant", ErrValidation)
		}
	}
	if parent.ParentID != "" {
		return fmt.Errorf("%w: category depth cannot exceed two levels", ErrValidation)
	}
	return nil
}

func (r *Repository) requireWalletOwner(ctx context.Context, userID string, walletID string) error {
	var owner string
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id::text
		FROM wallets
		WHERE id = $1 AND archived_at IS NULL
	`, walletID).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) || owner != userID {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("load wallet owner: %w", err)
	}
	return nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func (r *Repository) requireCategoryVisible(ctx context.Context, userID string, categoryID string) error {
	var visible bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE id = $1
				AND archived_at IS NULL
				AND (is_system OR user_id = $2)
		)
	`, categoryID, userID).Scan(&visible)
	if err != nil {
		return fmt.Errorf("check category visibility: %w", err)
	}
	if !visible {
		return ErrForbidden
	}
	return nil
}
