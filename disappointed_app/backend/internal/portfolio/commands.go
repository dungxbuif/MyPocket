package portfolio

import (
	"context"
	"mypocket/internal/platform/changefeed"
	"mypocket/internal/platform/commandtx"
)

func (r *Repository) CreatePosition(ctx context.Context, userID string, input CreatePositionInput) (Position, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Position, error) {
		bound := &Repository{db: h}

		value, err := bound.commandCreatePosition(ctx, userID, input)
		if err != nil {
			return Position{}, err
		}

		err = changefeed.Append(ctx, h, userID, "asset", value.ID, "create", value.Version, value)
		return value, err
	})
}

func (r *Repository) AddTrade(ctx context.Context, userID string, assetID string, input AddTradeInput) (Position, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Position, error) {
		bound := &Repository{db: h}
		before, err := bound.GetPosition(ctx, userID, assetID)
		if err != nil {
			return Position{}, err
		}
		value, err := bound.commandAddTrade(ctx, userID, assetID, input)
		if err != nil {
			return Position{}, err
		}
		if value.Version == before.Version {
			return value, nil
		}
		err = changefeed.Append(ctx, h, userID, "asset", value.ID, "add_trade", value.Version, value)
		return value, err
	})
}

func (r *Repository) UpdateTrade(ctx context.Context, userID string, assetID, tradeID string, input UpdateTradeInput) (Position, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Position, error) {
		bound := &Repository{db: h}
		before, err := bound.GetPosition(ctx, userID, assetID)
		if err != nil {
			return Position{}, err
		}
		value, err := bound.commandUpdateTrade(ctx, userID, assetID, tradeID, input)
		if err != nil {
			return Position{}, err
		}
		if value.Version == before.Version {
			return value, nil
		}
		err = changefeed.Append(ctx, h, userID, "asset", value.ID, "update_trade", value.Version, value)
		return value, err
	})
}

func (r *Repository) ArchiveTrade(ctx context.Context, userID string, assetID, tradeID string, baseVersion int64) (Position, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Position, error) {
		bound := &Repository{db: h}
		before, err := bound.GetPosition(ctx, userID, assetID)
		if err != nil {
			return Position{}, err
		}
		value, err := bound.commandArchiveTrade(ctx, userID, assetID, tradeID, baseVersion)
		if err != nil {
			return Position{}, err
		}
		if value.Version == before.Version {
			return value, nil
		}
		err = changefeed.Append(ctx, h, userID, "asset", value.ID, "archive_trade", value.Version, value)
		return value, err
	})
}

func (r *Repository) AddPrice(ctx context.Context, userID string, assetID string, input AddPriceInput) (Position, error) {
	return commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (Position, error) {
		bound := &Repository{db: h}
		before, err := bound.GetPosition(ctx, userID, assetID)
		if err != nil {
			return Position{}, err
		}
		value, err := bound.commandAddPrice(ctx, userID, assetID, input)
		if err != nil {
			return Position{}, err
		}
		if value.Version == before.Version {
			return value, nil
		}
		err = changefeed.Append(ctx, h, userID, "asset", value.ID, "add_price", value.Version, value)
		return value, err
	})
}

func (r *Repository) ArchivePosition(ctx context.Context, userID, assetID string, baseVersion int64) error {
	_, err := commandtx.Run(ctx, r.db, userID, func(h *commandtx.Handle) (struct{}, error) {
		bound := &Repository{db: h}
		if err := bound.commandArchivePosition(ctx, userID, assetID, baseVersion); err != nil {
			return struct{}{}, err
		}
		value, err := bound.GetPosition(ctx, userID, assetID)
		if err != nil {
			return struct{}{}, err
		}
		return struct{}{}, changefeed.Append(ctx, h, userID, "asset", assetID, "archive", value.Version, map[string]any{"id": assetID, "version": value.Version})
	})
	return err
}
