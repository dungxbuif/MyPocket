package finance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWallet(ctx context.Context, userID string, input CreateWalletInput) (Wallet, error) {
	input, err := ValidateCreateWallet(input)
	if err != nil {
		return Wallet{}, err
	}

	var wallet Wallet
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO wallets (
			user_id,
			name,
			type,
			credit_limit_vnd,
			statement_day,
			payment_due_day
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, user_id::text, name, type, balance_vnd, include_in_total, is_default_ai, version
	`,
		userID,
		input.Name,
		string(input.Type),
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
		INSERT INTO categories (user_id, parent_id, kind, name, is_system)
		VALUES ($1, NULL, $2, $3, false)
		RETURNING id::text, user_id::text, coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system
	`, userID, string(input.Kind), input.Name).Scan(
		&category.ID,
		&category.UserID,
		&category.ParentID,
		&category.Kind,
		&category.Name,
		&category.SystemKey,
		&category.IsSystem,
	)
	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}
	return category, nil
}

func (r *Repository) ListCategories(ctx context.Context, userID string) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, coalesce(user_id::text, ''), coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system
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
		RETURNING id::text, user_id::text, coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system
	`, categoryID, userID, trimmed(input.Name)).Scan(
		&category.ID,
		&category.UserID,
		&category.ParentID,
		&category.Kind,
		&category.Name,
		&category.SystemKey,
		&category.IsSystem,
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

func (r *Repository) getCategoryForUser(ctx context.Context, userID string, categoryID string) (Category, error) {
	var category Category
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, coalesce(user_id::text, ''), coalesce(parent_id::text, ''), kind, name, coalesce(system_key, ''), is_system
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
