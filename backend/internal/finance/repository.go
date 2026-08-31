package finance

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
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

func (r *Repository) CreateWallet(ctx context.Context, userID string, input CreateWalletInput) (Wallet, error) {
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

func (r *Repository) UpdateWallet(ctx context.Context, userID string, walletID string, input UpdateWalletInput) (Wallet, error) {
	input, err := ValidateUpdateWallet(input)
	if err != nil {
		return Wallet{}, err
	}
	includeInTotal := true
	if input.IncludeInTotal != nil {
		includeInTotal = *input.IncludeInTotal
	}

	var wallet Wallet
	err = r.db.QueryRowContext(ctx, `
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
	return wallet, nil
}

func (r *Repository) ArchiveWallet(ctx context.Context, userID string, walletID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE wallets
		SET archived_at = now(), is_default_ai = false, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
	`, walletID, userID)
	if err != nil {
		return fmt.Errorf("archive wallet: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive wallet rows affected: %w", err)
	}
	if affected == 0 {
		return ErrForbidden
	}
	return nil
}

func (r *Repository) SetDefaultAIWallet(ctx context.Context, userID string, walletID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin set default wallet: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM wallets
			WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		)
	`, walletID, userID).Scan(&exists); err != nil {
		return fmt.Errorf("check wallet owner: %w", err)
	}
	if !exists {
		return ErrForbidden
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET is_default_ai = false, updated_at = now(), version = version + 1
		WHERE user_id = $1 AND is_default_ai = true
	`, userID); err != nil {
		return fmt.Errorf("clear default wallets: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE wallets
		SET is_default_ai = true, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2
	`, walletID, userID); err != nil {
		return fmt.Errorf("set default wallet: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit set default wallet: %w", err)
	}
	return nil
}

func (r *Repository) CreateCategory(ctx context.Context, userID string, input CreateCategoryInput) (Category, error) {
	input, err := ValidateCreateCategory(input)
	if err != nil {
		return Category{}, err
	}

	var category Category
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO categories (id, user_id, parent_id, kind, name, is_system)
		VALUES (coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, NULL, $3, $4, false)
		RETURNING id::text, user_id::text, coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system, version
	`, input.ID, userID, string(input.Kind), input.Name).Scan(
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

func (r *Repository) UpdateCategory(ctx context.Context, userID string, categoryID string, input UpdateCategoryInput) (Category, error) {
	category, err := r.getCategoryForUser(ctx, userID, categoryID)
	if err != nil {
		return Category{}, err
	}
	if err := ValidateCategoryUpdate(category, input); err != nil {
		return Category{}, err
	}

	err = r.db.QueryRowContext(ctx, `
		UPDATE categories
		SET name = $3, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		RETURNING id::text, user_id::text, coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system, version
	`, categoryID, userID, trimmed(input.Name)).Scan(
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
		return Category{}, fmt.Errorf("update category: %w", err)
	}
	return category, nil
}

func (r *Repository) ArchiveCategory(ctx context.Context, userID string, categoryID string) error {
	category, err := r.getCategoryForUser(ctx, userID, categoryID)
	if err != nil {
		return err
	}
	if category.IsSystem {
		return ErrSystemCategoryLocked
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE categories
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
	`, categoryID, userID)
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
		return ErrForbidden
	}
	return nil
}

func (r *Repository) SetWalletCategoryActive(ctx context.Context, userID string, walletID string, categoryID string, active bool) error {
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

	replayed, ok, err := loadIdempotentTransaction(ctx, tx, userID, input.IdempotencyKey, requestHash)
	if err != nil {
		return Transaction{}, err
	}
	if ok {
		if err := tx.Commit(); err != nil {
			return Transaction{}, fmt.Errorf("commit idempotent transaction replay: %w", err)
		}
		return replayed, nil
	}

	sourceBalance, err := lockWalletBalance(ctx, tx, userID, input.SourceWalletID)
	if err != nil {
		return Transaction{}, err
	}
	destinationBalance := int64(0)
	if input.Type == TransactionTransfer {
		destinationBalance, err = lockWalletBalance(ctx, tx, userID, input.DestinationWalletID)
		if err != nil {
			return Transaction{}, err
		}
	}
	if err := requireTransactionCategory(ctx, tx, userID, input.SourceWalletID, input.CategoryID, input.Type); err != nil {
		return Transaction{}, err
	}

	effect, err := ApplyAccountingEffect(AccountingInput{
		Type:                  input.Type,
		AmountVND:             input.AmountVND,
		SourceWalletID:        input.SourceWalletID,
		DestinationWalletID:   input.DestinationWalletID,
		SourceBalanceVND:      sourceBalance,
		DestinationBalanceVND: destinationBalance,
		TargetBalanceVND:      input.TargetBalanceVND,
	})
	if err != nil {
		return Transaction{}, err
	}

	if err := updateWalletBalance(ctx, tx, userID, input.SourceWalletID, effect.SourceBalanceVND); err != nil {
		return Transaction{}, err
	}
	if input.Type == TransactionTransfer {
		if err := updateWalletBalance(ctx, tx, userID, input.DestinationWalletID, effect.DestinationBalanceVND); err != nil {
			return Transaction{}, err
		}
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

	if err := tx.Commit(); err != nil {
		return Transaction{}, fmt.Errorf("commit create transaction: %w", err)
	}
	return transaction, nil
}

func (r *Repository) UpdateTransaction(ctx context.Context, userID string, transactionID string, input UpdateTransactionInput) (Transaction, error) {
	normalized, err := ValidateUpdateTransaction(input)
	if err != nil {
		return Transaction{}, err
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
	balances, err := lockWalletBalances(ctx, tx, userID, transactionWalletIDs(current, normalized))
	if err != nil {
		return Transaction{}, err
	}
	if err := requireTransactionCategory(ctx, tx, userID, normalized.SourceWalletID, normalized.CategoryID, normalized.Type); err != nil {
		return Transaction{}, err
	}

	reverseTransactionEffect(balances, current)
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
	if err := tx.Commit(); err != nil {
		return Transaction{}, fmt.Errorf("commit update transaction: %w", err)
	}
	return updated, nil
}

func (r *Repository) ArchiveTransaction(ctx context.Context, userID string, transactionID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin archive transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := loadActiveTransaction(ctx, tx, userID, transactionID)
	if err != nil {
		return err
	}
	balances, err := lockWalletBalances(ctx, tx, userID, transactionWalletIDs(current, CreateTransactionInput{}))
	if err != nil {
		return err
	}
	reverseTransactionEffect(balances, current)
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

func loadIdempotentTransaction(ctx context.Context, tx *sql.Tx, userID string, key string, requestHash string) (Transaction, bool, error) {
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

func loadActiveTransaction(ctx context.Context, tx *sql.Tx, userID string, transactionID string) (Transaction, error) {
	transaction, err := queryTransaction(ctx, tx, `
		SELECT
			id::text,
			user_id::text,
			type,
			source_wallet_id::text,
			coalesce(destination_wallet_id::text, ''),
			coalesce(category_id::text, ''),
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

func lockWalletBalance(ctx context.Context, tx *sql.Tx, userID string, walletID string) (int64, error) {
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

func lockWalletBalances(ctx context.Context, tx *sql.Tx, userID string, walletIDs []string) (map[string]int64, error) {
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

func requireTransactionCategory(ctx context.Context, tx *sql.Tx, userID string, walletID string, categoryID string, txType TransactionType) error {
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

func requireWalletCategoryActive(ctx context.Context, tx *sql.Tx, userID string, walletID string, categoryID string) error {
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

func updateWalletBalance(ctx context.Context, tx *sql.Tx, userID string, walletID string, balanceVND int64) error {
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
	return nil
}

func updateWalletBalances(ctx context.Context, tx *sql.Tx, userID string, balances map[string]int64) error {
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

func insertTransaction(ctx context.Context, tx *sql.Tx, userID string, input CreateTransactionInput, effect AccountingEffect) (Transaction, error) {
	var transaction Transaction
	returning := `
		INSERT INTO transactions (
			id,
			user_id,
			type,
			source_wallet_id,
			destination_wallet_id,
			category_id,
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
		VALUES (coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, $3, $4, NULLIF($5, '')::uuid, NULLIF($6, '')::uuid, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING
			id::text,
			user_id::text,
			type,
			source_wallet_id::text,
			coalesce(destination_wallet_id::text, ''),
			coalesce(category_id::text, ''),
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

func updateTransactionRow(ctx context.Context, tx *sql.Tx, userID string, transactionID string, input CreateTransactionInput, effect AccountingEffect) (Transaction, error) {
	transaction, err := queryTransaction(ctx, tx, `
		UPDATE transactions
		SET
			type = $3,
			source_wallet_id = $4,
			destination_wallet_id = NULLIF($5, '')::uuid,
			category_id = NULLIF($6, '')::uuid,
			amount_vnd = $7,
			balance_after_vnd = $8,
			source_delta_vnd = $9,
			destination_delta_vnd = $10,
			note = $11,
			with_person = $12,
			event_ref = $13,
			occurred_at = $14,
			excluded_from_reports = $15,
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

func queryTransaction(ctx context.Context, tx *sql.Tx, query string, args ...any) (Transaction, error) {
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

func reverseTransactionEffect(balances map[string]int64, transaction Transaction) {
	balances[transaction.SourceWalletID] -= transaction.SourceDeltaVND
	if transaction.DestinationWalletID != "" {
		balances[transaction.DestinationWalletID] -= transaction.DestinationDeltaVND
	}
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
