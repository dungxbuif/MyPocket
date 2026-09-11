package portfolio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mypocket/internal/platform/commandtx"
)

type Repository struct {
	db *commandtx.Handle
}

func NewRepositoryInTx(tx *sql.Tx) *Repository { return &Repository{db: commandtx.Bound(tx)} }

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: commandtx.New(db)}
}

func (r *Repository) commandCreatePosition(ctx context.Context, userID string, input CreatePositionInput) (Position, error) {
	input, err := ValidateCreatePosition(input)
	if err != nil {
		return Position{}, err
	}
	include := true
	if input.IncludeInNetWorth != nil {
		include = *input.IncludeInNetWorth
	}
	var p Position
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO asset_positions (
			id, user_id, type, symbol, exchange, name, unit, pricing_mode,
			provider_key, provider_symbol, include_in_net_worth
		)
		VALUES (
			coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6,
			$7, $8, nullif($9, ''), nullif($10, ''), $11
		)
		RETURNING id::text, user_id::text, type, symbol, exchange, name, unit,
			reporting_currency, pricing_mode, coalesce(provider_key, ''),
			coalesce(provider_symbol, ''), include_in_net_worth, archived_at, version
	`, input.ID, userID, string(input.Type), input.Symbol, input.Exchange, input.Name, input.Unit, string(input.PricingMode), input.ProviderKey, input.ProviderSymbol, include).Scan(
		&p.ID, &p.UserID, &p.Type, &p.Symbol, &p.Exchange, &p.Name, &p.Unit,
		&p.ReportingCurrency, &p.PricingMode, &p.ProviderKey, &p.ProviderSymbol,
		&p.IncludeInNetWorth, &p.ArchivedAt, &p.Version,
	)
	if err != nil {
		return Position{}, fmt.Errorf("create asset position: %w", err)
	}
	p.Summary, err = summarize("0", 0, 0, nil)
	return p, err
}

func (r *Repository) ListPositions(ctx context.Context, userID string, includeArchived bool) ([]Position, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, type, symbol, exchange, name, unit,
			reporting_currency, pricing_mode, coalesce(provider_key, ''),
			coalesce(provider_symbol, ''), include_in_net_worth, archived_at, version
		FROM asset_positions
		WHERE user_id = $1 AND ($2 OR archived_at IS NULL)
		ORDER BY created_at DESC, id DESC
	`, userID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("list asset positions: %w", err)
	}
	defer rows.Close()
	var positions []Position
	for rows.Next() {
		var p Position
		if err := scanPosition(rows, &p); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	// A bound repository uses one connection; finish the row stream before
	// issuing the summary queries needed for the same consistent snapshot.
	for i := range positions {
		if err := r.attachSummary(ctx, userID, &positions[i], false); err != nil {
			return nil, err
		}
	}
	return positions, nil
}

func (r *Repository) GetPosition(ctx context.Context, userID, assetID string) (Position, error) {
	var p Position
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, type, symbol, exchange, name, unit,
			reporting_currency, pricing_mode, coalesce(provider_key, ''),
			coalesce(provider_symbol, ''), include_in_net_worth, archived_at, version
		FROM asset_positions
		WHERE id = $1 AND user_id = $2
	`, assetID, userID).Scan(
		&p.ID, &p.UserID, &p.Type, &p.Symbol, &p.Exchange, &p.Name, &p.Unit,
		&p.ReportingCurrency, &p.PricingMode, &p.ProviderKey, &p.ProviderSymbol,
		&p.IncludeInNetWorth, &p.ArchivedAt, &p.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Position{}, ErrForbidden
	}
	if err != nil {
		return Position{}, fmt.Errorf("get asset position: %w", err)
	}
	if err := r.attachSummary(ctx, userID, &p, true); err != nil {
		return Position{}, err
	}
	return p, nil
}

func (r *Repository) commandArchivePosition(ctx context.Context, userID, assetID string, baseVersion int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE asset_positions
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL AND ($3 = 0 OR version = $3)
	`, assetID, userID, baseVersion)
	if err != nil {
		return fmt.Errorf("archive asset position: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive asset position rows: %w", err)
	}
	if affected == 0 {
		if exists, err := r.exists(ctx, userID, assetID); err != nil {
			return err
		} else if exists {
			return ErrConflict
		}
		return ErrForbidden
	}
	return nil
}

func (r *Repository) commandAddTrade(ctx context.Context, userID, assetID string, input AddTradeInput) (Position, error) {
	input, err := ValidateAddTrade(input)
	if err != nil {
		return Position{}, err
	}
	quantity, _, err := normalizeDecimal(input.Quantity)
	if err != nil {
		return Position{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Position{}, fmt.Errorf("begin add asset trade: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockActivePosition(ctx, tx, userID, assetID, input.BaseVersion); err != nil {
		return Position{}, err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO asset_trades (
			id, user_id, asset_id, side, quantity, unit_price_vnd, fee_vnd, occurred_at, note
		)
		VALUES (coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5::numeric, $6, $7, $8, $9)
	`, input.ID, userID, assetID, string(input.Side), quantity, input.UnitPriceVND, input.FeeVND, input.OccurredAt.UTC(), input.Note)
	if err != nil {
		return Position{}, fmt.Errorf("insert asset trade: %w", err)
	}
	if err := replayPosition(ctx, tx, userID, assetID); err != nil {
		return Position{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE asset_positions SET updated_at = now(), version = version + 1 WHERE id = $1 AND user_id = $2`, assetID, userID); err != nil {
		return Position{}, fmt.Errorf("bump asset position after trade: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Position{}, fmt.Errorf("commit add asset trade: %w", err)
	}
	return r.GetPosition(ctx, userID, assetID)
}

func (r *Repository) commandUpdateTrade(ctx context.Context, userID, assetID, tradeID string, input UpdateTradeInput) (Position, error) {
	input, err := ValidateUpdateTrade(input)
	if err != nil {
		return Position{}, err
	}
	quantity, _, err := normalizeDecimal(input.Quantity)
	if err != nil {
		return Position{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Position{}, fmt.Errorf("begin update asset trade: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockActivePosition(ctx, tx, userID, assetID, 0); err != nil {
		return Position{}, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE asset_trades
		SET side = $4, quantity = $5::numeric, unit_price_vnd = $6, fee_vnd = $7,
			occurred_at = $8, note = $9, updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND asset_id = $3 AND archived_at IS NULL
			AND ($10 = 0 OR version = $10)
	`, tradeID, userID, assetID, string(input.Side), quantity, input.UnitPriceVND, input.FeeVND, input.OccurredAt.UTC(), input.Note, input.BaseVersion)
	if err != nil {
		return Position{}, fmt.Errorf("update asset trade: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Position{}, fmt.Errorf("update asset trade rows: %w", err)
	}
	if affected == 0 {
		return Position{}, ErrConflict
	}
	if err := replayPosition(ctx, tx, userID, assetID); err != nil {
		return Position{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE asset_positions SET updated_at = now(), version = version + 1 WHERE id = $1 AND user_id = $2`, assetID, userID); err != nil {
		return Position{}, fmt.Errorf("bump asset position after trade update: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Position{}, fmt.Errorf("commit update asset trade: %w", err)
	}
	return r.GetPosition(ctx, userID, assetID)
}

func (r *Repository) commandArchiveTrade(ctx context.Context, userID, assetID, tradeID string, baseVersion int64) (Position, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Position{}, fmt.Errorf("begin archive asset trade: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockActivePosition(ctx, tx, userID, assetID, 0); err != nil {
		return Position{}, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE asset_trades
		SET archived_at = now(), updated_at = now(), version = version + 1
		WHERE id = $1 AND user_id = $2 AND asset_id = $3 AND archived_at IS NULL
			AND ($4 = 0 OR version = $4)
	`, tradeID, userID, assetID, baseVersion)
	if err != nil {
		return Position{}, fmt.Errorf("archive asset trade: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Position{}, fmt.Errorf("archive asset trade rows: %w", err)
	}
	if affected == 0 {
		return Position{}, ErrConflict
	}
	if err := replayPosition(ctx, tx, userID, assetID); err != nil {
		return Position{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE asset_positions SET updated_at = now(), version = version + 1 WHERE id = $1 AND user_id = $2`, assetID, userID); err != nil {
		return Position{}, fmt.Errorf("bump asset position after trade archive: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Position{}, fmt.Errorf("commit archive asset trade: %w", err)
	}
	return r.GetPosition(ctx, userID, assetID)
}

func (r *Repository) commandAddPrice(ctx context.Context, userID, assetID string, input AddPriceInput) (Position, error) {
	input, err := ValidateAddPrice(input)
	if err != nil {
		return Position{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Position{}, fmt.Errorf("begin add asset price: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockActivePosition(ctx, tx, userID, assetID, input.BaseVersion); err != nil {
		return Position{}, err
	}
	if input.ProviderQuoteID != "" {
		var existing string
		err = tx.QueryRowContext(ctx, `
			SELECT asset_id::text
			FROM asset_price_history
			WHERE source = $1 AND provider_quote_id = $2
		`, input.Source, input.ProviderQuoteID).Scan(&existing)
		if err == nil {
			if existing != assetID {
				return Position{}, ErrConflict
			}
			if err := tx.Commit(); err != nil {
				return Position{}, fmt.Errorf("commit replay asset price: %w", err)
			}
			return r.GetPosition(ctx, userID, assetID)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return Position{}, fmt.Errorf("find provider quote: %w", err)
		}
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO asset_price_history (
			id, user_id, asset_id, unit_price_vnd, priced_at, source, provider_quote_id
		)
		VALUES (coalesce(nullif($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, nullif($7, ''))
	`, input.ID, userID, assetID, input.UnitPriceVND, input.PricedAt.UTC(), input.Source, input.ProviderQuoteID)
	if err != nil {
		return Position{}, fmt.Errorf("insert asset price: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE asset_positions SET updated_at = now(), version = version + 1 WHERE id = $1 AND user_id = $2`, assetID, userID); err != nil {
		return Position{}, fmt.Errorf("bump asset position after price: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Position{}, fmt.Errorf("commit add asset price: %w", err)
	}
	return r.GetPosition(ctx, userID, assetID)
}

func (r *Repository) Summary(ctx context.Context, userID string) (PortfolioSummary, error) {
	positions, err := r.ListPositions(ctx, userID, false)
	if err != nil {
		return PortfolioSummary{}, err
	}
	var s PortfolioSummary
	s.PositionCount = len(positions)
	for _, position := range positions {
		if !position.IncludeInNetWorth {
			continue
		}
		s.IncludedPositionCount++
		if position.Summary.MarketValueVND == nil {
			s.MissingPriceCount++
			continue
		}
		s.InvestmentMarketValueVND, err = addMoney(s.InvestmentMarketValueVND, *position.Summary.MarketValueVND)
		if err != nil {
			return PortfolioSummary{}, err
		}
	}
	return s, nil
}

func (r *Repository) AcquireWorkerLease(ctx context.Context, leaseKey string, owner string, ttlSeconds int64) (bool, error) {
	var acquired bool
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO worker_leases (lease_key, owner, expires_at)
		VALUES ($1, $2, now() + make_interval(secs => $3))
		ON CONFLICT (lease_key) DO UPDATE
		SET owner = EXCLUDED.owner,
			expires_at = EXCLUDED.expires_at,
			updated_at = now()
		WHERE worker_leases.expires_at <= now() OR worker_leases.owner = EXCLUDED.owner
		RETURNING true
	`, leaseKey, owner, ttlSeconds).Scan(&acquired)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("acquire portfolio worker lease: %w", err)
	}
	return acquired, nil
}

func (r *Repository) ListPriceRefreshCandidates(ctx context.Context, limit int) ([]PriceRefreshCandidate, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id::text, id::text, coalesce(provider_key, ''), coalesce(provider_symbol, ''), version
		FROM asset_positions
		WHERE archived_at IS NULL
			AND pricing_mode = 'automatic'
			AND provider_key IS NOT NULL
			AND provider_symbol IS NOT NULL
		ORDER BY updated_at ASC, id
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list portfolio price refresh candidates: %w", err)
	}
	defer rows.Close()
	var candidates []PriceRefreshCandidate
	for rows.Next() {
		var candidate PriceRefreshCandidate
		if err := rows.Scan(&candidate.UserID, &candidate.AssetID, &candidate.ProviderKey, &candidate.ProviderSymbol, &candidate.BaseVersion); err != nil {
			return nil, fmt.Errorf("scan portfolio price refresh candidate: %w", err)
		}
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

func (r *Repository) AddProviderPrice(ctx context.Context, candidate PriceRefreshCandidate, quote ProviderQuote) (Position, error) {
	return r.AddPrice(ctx, candidate.UserID, candidate.AssetID, AddPriceInput{
		UnitPriceVND:    quote.UnitPriceVND,
		PricedAt:        quote.PricedAt,
		Source:          "provider:" + quote.ProviderKey,
		ProviderQuoteID: quote.ProviderQuoteID,
		BaseVersion:     0,
	})
}

func (r *Repository) exists(ctx context.Context, userID, assetID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM asset_positions WHERE id = $1 AND user_id = $2)`, assetID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check asset position exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) attachSummary(ctx context.Context, userID string, p *Position, includeDetail bool) error {
	trades, quantity, costBasis, realized, err := r.loadTradesAndTotals(ctx, userID, p.ID)
	if err != nil {
		return err
	}
	latest, err := r.latestPrice(ctx, userID, p.ID)
	if err != nil {
		return err
	}
	p.LatestPrice = latest
	p.Summary, err = summarize(quantity, costBasis, realized, latest)
	if err != nil {
		return err
	}
	if includeDetail {
		p.Trades = trades
		history, err := r.priceHistory(ctx, userID, p.ID, 30)
		if err != nil {
			return err
		}
		p.PriceHistory = history
	}
	return nil
}

func (r *Repository) loadTradesAndTotals(ctx context.Context, userID, assetID string) ([]Trade, string, int64, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, asset_id::text, side, quantity::text, unit_price_vnd,
			fee_vnd, occurred_at, quantity_after::text, cost_basis_after_vnd,
			realized_pnl_vnd, note, version, archived_at
		FROM asset_trades
		WHERE user_id = $1 AND asset_id = $2 AND archived_at IS NULL
		ORDER BY occurred_at, created_at, id
	`, userID, assetID)
	if err != nil {
		return nil, "", 0, 0, fmt.Errorf("list asset trades: %w", err)
	}
	defer rows.Close()
	trades := []Trade{}
	quantity := "0"
	var costBasis int64
	var realized int64
	for rows.Next() {
		var trade Trade
		if err := rows.Scan(
			&trade.ID, &trade.UserID, &trade.AssetID, &trade.Side, &trade.Quantity,
			&trade.UnitPriceVND, &trade.FeeVND, &trade.OccurredAt, &trade.QuantityAfter,
			&trade.CostBasisAfterVND, &trade.RealizedPNLVND, &trade.Note, &trade.Version,
			&trade.ArchivedAt,
		); err != nil {
			return nil, "", 0, 0, fmt.Errorf("scan asset trade: %w", err)
		}
		trade.Quantity = ratToDecimalString(parseStoredDecimal(trade.Quantity))
		trade.QuantityAfter = ratToDecimalString(parseStoredDecimal(trade.QuantityAfter))
		quantity = trade.QuantityAfter
		costBasis = trade.CostBasisAfterVND
		realized, err = addMoney(realized, trade.RealizedPNLVND)
		if err != nil {
			return nil, "", 0, 0, err
		}
		trades = append(trades, trade)
	}
	if err := rows.Err(); err != nil {
		return nil, "", 0, 0, err
	}
	return trades, quantity, costBasis, realized, nil
}

func (r *Repository) latestPrice(ctx context.Context, userID, assetID string) (*PricePoint, error) {
	var price PricePoint
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, asset_id::text, unit_price_vnd, priced_at,
			source, coalesce(provider_quote_id, ''), created_at
		FROM asset_price_history
		WHERE user_id = $1 AND asset_id = $2
		ORDER BY priced_at DESC, created_at DESC, id DESC
		LIMIT 1
	`, userID, assetID).Scan(
		&price.ID, &price.UserID, &price.AssetID, &price.UnitPriceVND, &price.PricedAt,
		&price.Source, &price.ProviderQuoteID, &price.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("latest asset price: %w", err)
	}
	return &price, nil
}

func (r *Repository) priceHistory(ctx context.Context, userID, assetID string, limit int) ([]PricePoint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, asset_id::text, unit_price_vnd, priced_at,
			source, coalesce(provider_quote_id, ''), created_at
		FROM asset_price_history
		WHERE user_id = $1 AND asset_id = $2
		ORDER BY priced_at DESC, created_at DESC, id DESC
		LIMIT $3
	`, userID, assetID, limit)
	if err != nil {
		return nil, fmt.Errorf("asset price history: %w", err)
	}
	defer rows.Close()
	var prices []PricePoint
	for rows.Next() {
		var price PricePoint
		if err := rows.Scan(
			&price.ID, &price.UserID, &price.AssetID, &price.UnitPriceVND, &price.PricedAt,
			&price.Source, &price.ProviderQuoteID, &price.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan asset price: %w", err)
		}
		prices = append(prices, price)
	}
	return prices, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPosition(scanner rowScanner, p *Position) error {
	if err := scanner.Scan(
		&p.ID, &p.UserID, &p.Type, &p.Symbol, &p.Exchange, &p.Name, &p.Unit,
		&p.ReportingCurrency, &p.PricingMode, &p.ProviderKey, &p.ProviderSymbol,
		&p.IncludeInNetWorth, &p.ArchivedAt, &p.Version,
	); err != nil {
		return fmt.Errorf("scan asset position: %w", err)
	}
	return nil
}

func lockActivePosition(ctx context.Context, tx commandtx.Queryer, userID, assetID string, baseVersion int64) error {
	var version int64
	err := tx.QueryRowContext(ctx, `
		SELECT version
		FROM asset_positions
		WHERE id = $1 AND user_id = $2 AND archived_at IS NULL
		FOR UPDATE
	`, assetID, userID).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("lock asset position: %w", err)
	}
	if baseVersion > 0 && version != baseVersion {
		return ErrConflict
	}
	return nil
}

func replayPosition(ctx context.Context, tx commandtx.Queryer, userID, assetID string) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id::text, side, quantity::text, unit_price_vnd, fee_vnd
		FROM asset_trades
		WHERE user_id = $1 AND asset_id = $2 AND archived_at IS NULL
		ORDER BY occurred_at, created_at, id
	`, userID, assetID)
	if err != nil {
		return fmt.Errorf("load asset trades for replay: %w", err)
	}
	var trades []ledgerTrade
	for rows.Next() {
		var trade ledgerTrade
		if err := rows.Scan(&trade.ID, &trade.Side, &trade.Quantity, &trade.UnitPriceVND, &trade.FeeVND); err != nil {
			rows.Close()
			return fmt.Errorf("scan replay trade: %w", err)
		}
		trades = append(trades, trade)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	results, err := replayLedger(trades)
	if err != nil {
		return err
	}
	for i, result := range results {
		if _, err := tx.ExecContext(ctx, `
			UPDATE asset_trades
			SET quantity_after = $3::numeric,
				cost_basis_after_vnd = $4,
				realized_pnl_vnd = $5,
				updated_at = now()
			WHERE id = $1 AND user_id = $2
		`, trades[i].ID, userID, result.QuantityAfter, result.CostBasisAfterVND, result.RealizedPNLVND); err != nil {
			return fmt.Errorf("persist replay trade: %w", err)
		}
	}
	return nil
}
