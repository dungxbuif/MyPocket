package finance

import (
	"context"
	"mypocket/internal/platform/changefeed"
	"mypocket/internal/platform/commandtx"
)

func (r *Repository) CreateWallet(ctx context.Context, userID string, input CreateWalletInput) (Wallet, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Wallet, error) {
		value, err := (&Repository{db: h}).commandCreateWallet(ctx, userID, input)
		if err != nil {
			return Wallet{}, err
		}
		err = changefeed.Append(ctx, h, userID, "wallet", value.ID, "create", value.Version, value)
		return value, err
	})
}

func (r *Repository) UpdateWallet(ctx context.Context, userID string, walletID string, input UpdateWalletInput) (Wallet, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Wallet, error) {
		value, err := (&Repository{db: h}).commandUpdateWallet(ctx, userID, walletID, input)
		if err != nil {
			return Wallet{}, err
		}
		err = changefeed.Append(ctx, h, userID, "wallet", value.ID, "update", value.Version, value)
		return value, err
	})
}

func (r *Repository) CreateCategory(ctx context.Context, userID string, input CreateCategoryInput) (Category, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Category, error) {
		value, err := (&Repository{db: h}).commandCreateCategory(ctx, userID, input)
		if err != nil {
			return Category{}, err
		}
		err = changefeed.Append(ctx, h, userID, "category", value.ID, "create", value.Version, value)
		return value, err
	})
}

func (r *Repository) UpdateCategory(ctx context.Context, userID string, categoryID string, input UpdateCategoryInput) (Category, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Category, error) {
		value, err := (&Repository{db: h}).commandUpdateCategory(ctx, userID, categoryID, input)
		if err != nil {
			return Category{}, err
		}
		err = changefeed.Append(ctx, h, userID, "category", value.ID, "update", value.Version, value)
		return value, err
	})
}

func (r *Repository) ArchiveWallet(ctx context.Context, userID, walletID string, baseVersion int64) error {
	_, err := commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (struct{}, error) {
		if err := (&Repository{db: h}).commandArchiveWallet(ctx, userID, walletID, baseVersion); err != nil {
			return struct{}{}, err
		}
		return struct{}{}, changefeed.Append(ctx, h, userID, "wallet", walletID, "archive", baseVersion+1, map[string]any{"id": walletID, "version": baseVersion + 1})
	})
	return err
}

func (r *Repository) ArchiveCategory(ctx context.Context, userID, categoryID string, baseVersion int64) error {
	_, err := commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (struct{}, error) {
		if err := (&Repository{db: h}).commandArchiveCategory(ctx, userID, categoryID, baseVersion); err != nil {
			return struct{}{}, err
		}
		return struct{}{}, changefeed.Append(ctx, h, userID, "category", categoryID, "archive", baseVersion+1, map[string]any{"id": categoryID, "version": baseVersion + 1})
	})
	return err
}

func (r *Repository) SetDefaultAIWallet(ctx context.Context, userID, walletID string, baseVersion int64) error {
	_, err := commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (struct{}, error) {
		bound := &Repository{db: h}
		before, err := bound.ListWallets(ctx, userID)
		if err != nil {
			return struct{}{}, err
		}
		if err = bound.commandSetDefaultAIWallet(ctx, userID, walletID, baseVersion); err != nil {
			return struct{}{}, err
		}
		after, err := bound.ListWallets(ctx, userID)
		if err != nil {
			return struct{}{}, err
		}
		versions := map[string]int64{}
		for _, w := range before {
			versions[w.ID] = w.Version
		}
		for _, w := range after {
			if versions[w.ID] != w.Version {
				if err = changefeed.Append(ctx, h, userID, "wallet", w.ID, "update", w.Version, w); err != nil {
					return struct{}{}, err
				}
			}
		}
		return struct{}{}, nil
	})
	return err
}
func (r *Repository) SetWalletCategoryActive(ctx context.Context, userID, walletID, categoryID string, active bool) error {
	_, err := commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (struct{}, error) {
		if err := (&Repository{db: h}).commandSetWalletCategoryActive(ctx, userID, walletID, categoryID, active); err != nil {
			return struct{}{}, err
		}
		return struct{}{}, changefeed.Append(ctx, h, userID, "category", categoryID, "set_category_active", 0, map[string]any{"wallet_id": walletID, "active": active})
	})
	return err
}
