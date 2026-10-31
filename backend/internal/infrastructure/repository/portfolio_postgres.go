package repository

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	portfoliorepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PortfolioPostgresRepository struct{ db *gorm.DB }

func NewPortfolioPostgresRepository(db *gorm.DB) portfoliorepo.PortfolioRepository {
	return &PortfolioPostgresRepository{db: db}
}

func (r *PortfolioPostgresRepository) ListAssets(ownerID string) ([]entity.PortfolioAsset, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	rows := make([]entity.PortfolioAsset, 0)
	if err := r.db.Where("owner_id = ?", ownerID).Order("symbol ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PortfolioPostgresRepository) CreateAsset(ownerID string, input portfoliorepo.PortfolioAssetInput) (*entity.PortfolioAsset, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	symbol, name := strings.TrimSpace(input.Symbol), strings.TrimSpace(input.Name)
	if symbol == "" || name == "" || len(symbol) > 32 || len(name) > 160 {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	priceMode := input.PriceMode
	if priceMode == "" {
		priceMode = entity.PortfolioPriceModeManual
		if input.MarketProductID != nil {
			priceMode = entity.PortfolioPriceModeAuto
		}
	}
	if priceMode != entity.PortfolioPriceModeManual && priceMode != entity.PortfolioPriceModeAuto {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	if priceMode == entity.PortfolioPriceModeAuto && input.MarketProductID == nil {
		return nil, portfoliorepo.ErrMarketBindingInvalid
	}
	row := &entity.PortfolioAsset{ID: uuid.NewString(), OwnerID: ownerID, Symbol: symbol, Name: name, PriceMode: priceMode}
	if input.UnitLabel != nil {
		unit := strings.TrimSpace(*input.UnitLabel)
		if unit == "" || len(unit) > 64 {
			return nil, portfoliorepo.ErrPortfolioAssetUnitInvalid
		}
		row.UnitLabel = &unit
		row.UnitConfirmed = true
	}
	if input.MarketProductID != nil {
		productID := strings.TrimSpace(*input.MarketProductID)
		if productID == "" || priceMode != entity.PortfolioPriceModeAuto || row.UnitLabel == nil || !row.UnitConfirmed {
			return nil, portfoliorepo.ErrMarketBindingInvalid
		}
		var product entity.MarketProduct
		if err := r.db.Where("id = ?", productID).First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, portfoliorepo.ErrMarketProductNotFound
			}
			return nil, err
		}
		if err := validateMarketProductForAsset(row, product); err != nil {
			return nil, err
		}
		row.MarketProductID = &productID
	}
	if err := r.db.Create(row).Error; err != nil {
		if portfolioUniqueViolation(err) {
			return nil, portfoliorepo.ErrPortfolioAssetConflict
		}
		return nil, err
	}
	return row, nil
}

func (r *PortfolioPostgresRepository) ConfirmAssetUnit(ownerID, assetID string, expectedVersion int64, unitLabel string) (*entity.PortfolioAsset, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" || expectedVersion <= 0 {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	unitLabel = strings.TrimSpace(unitLabel)
	if unitLabel == "" || len(unitLabel) > 64 {
		return nil, portfoliorepo.ErrPortfolioAssetUnitInvalid
	}
	var row entity.PortfolioAsset
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, assetID).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return portfoliorepo.ErrPortfolioAssetNotFound
			}
			return err
		}
		if row.Version != expectedVersion {
			return portfoliorepo.ErrPortfolioAssetConflictVersion
		}
		if row.UnitConfirmed && (row.UnitLabel == nil || strings.TrimSpace(*row.UnitLabel) != unitLabel) {
			var tradeCount int64
			if err := tx.Model(&entity.PortfolioTrade{}).Where("owner_id = ? AND asset_id = ?", ownerID, assetID).Count(&tradeCount).Error; err != nil {
				return err
			}
			if tradeCount > 0 {
				return portfoliorepo.ErrPortfolioAssetUnitConflict
			}
		}
		if row.UnitConfirmed && row.UnitLabel != nil && strings.TrimSpace(*row.UnitLabel) == unitLabel {
			return nil
		}
		row.UnitLabel = &unitLabel
		row.UnitConfirmed = true
		row.Version++
		return tx.Model(&row).Updates(map[string]any{
			"unit_label":     row.UnitLabel,
			"unit_confirmed": true,
			"version":        row.Version,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PortfolioPostgresRepository) UpdateAsset(ownerID, assetID string, update portfoliorepo.PortfolioAssetUpdate) (*entity.PortfolioAsset, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" || (update.Symbol == nil && update.Name == nil) {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	if update.Symbol != nil {
		value := strings.TrimSpace(*update.Symbol)
		if value == "" || len(value) > 32 {
			return nil, portfoliorepo.ErrPortfolioAssetInvalid
		}
	}
	if update.Name != nil {
		value := strings.TrimSpace(*update.Name)
		if value == "" || len(value) > 160 {
			return nil, portfoliorepo.ErrPortfolioAssetInvalid
		}
	}
	var row entity.PortfolioAsset
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, assetID).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return portfoliorepo.ErrPortfolioAssetNotFound
			}
			return err
		}
		changes := make(map[string]any, 2)
		if update.Symbol != nil {
			row.Symbol = strings.TrimSpace(*update.Symbol)
			changes["symbol"] = row.Symbol
		}
		if update.Name != nil {
			row.Name = strings.TrimSpace(*update.Name)
			changes["name"] = row.Name
		}
		if err := tx.Model(&row).Updates(changes).Error; err != nil {
			if portfolioUniqueViolation(err) {
				return portfoliorepo.ErrPortfolioAssetConflict
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PortfolioPostgresRepository) DeleteAsset(ownerID, assetID string) error {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" {
		return portfoliorepo.ErrPortfolioAssetInvalid
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var asset entity.PortfolioAsset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, assetID).First(&asset).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return portfoliorepo.ErrPortfolioAssetNotFound
			}
			return err
		}
		var tradeCount int64
		if err := tx.Model(&entity.PortfolioTrade{}).Where("owner_id = ? AND asset_id = ?", ownerID, assetID).Count(&tradeCount).Error; err != nil {
			return err
		}
		if tradeCount > 0 {
			return portfoliorepo.ErrPortfolioAssetHasTrades
		}
		return tx.Delete(&asset).Error
	})
}

func (r *PortfolioPostgresRepository) UpdateAssetPrice(ownerID, assetID string, update portfoliorepo.PortfolioPriceUpdate) (*entity.PortfolioAsset, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" || (update.LatestPrice != nil && *update.LatestPrice < 0) {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	var row entity.PortfolioAsset
	if err := r.db.Where("owner_id = ? AND id = ?", ownerID, assetID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, portfoliorepo.ErrPortfolioAssetNotFound
		}
		return nil, err
	}
	if row.PriceMode == entity.PortfolioPriceModeAuto {
		return nil, portfoliorepo.ErrPortfolioAssetAutomaticPrice
	}
	row.LatestPrice = update.LatestPrice
	row.LatestPriceAt = update.LatestPriceAt
	row.LatestPriceSource = nil
	if update.LatestPrice == nil {
		row.LatestPriceAt = nil
	}
	if err := r.db.Model(&row).Updates(map[string]any{"latest_price": row.LatestPrice, "latest_price_at": row.LatestPriceAt, "latest_price_source": nil}).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PortfolioPostgresRepository) ListTrades(ownerID, assetID string) ([]entity.PortfolioTrade, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	if err := r.ensureAsset(ownerID, assetID); err != nil {
		return nil, err
	}
	rows := make([]entity.PortfolioTrade, 0)
	if err := r.db.Where("owner_id = ? AND asset_id = ?", ownerID, assetID).Order("occurred_at ASC, created_at ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Quantity = entity.FormatPortfolioQuantity(rows[i].QuantityScaled)
	}
	return rows, nil
}

func (r *PortfolioPostgresRepository) CreateTrade(ownerID, assetID string, input portfoliorepo.PortfolioTradeInput) (*entity.PortfolioTrade, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" || input.QuantityScaled <= 0 || (input.Side != entity.PortfolioTradeBuy && input.Side != entity.PortfolioTradeSell) || input.UnitPrice < 0 || input.Fee < 0 || input.OccurredAt.IsZero() {
		return nil, portfoliorepo.ErrPortfolioTradeInvalid
	}
	var created entity.PortfolioTrade
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var asset entity.PortfolioAsset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, assetID).First(&asset).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return portfoliorepo.ErrPortfolioAssetNotFound
			}
			return err
		}
		var existing []entity.PortfolioTrade
		if err := tx.Where("owner_id = ? AND asset_id = ?", ownerID, assetID).Order("occurred_at ASC, created_at ASC, id ASC").Find(&existing).Error; err != nil {
			return err
		}
		candidate := entity.PortfolioTrade{ID: uuid.NewString(), OwnerID: ownerID, AssetID: assetID, Side: input.Side, QuantityScaled: input.QuantityScaled, UnitPrice: input.UnitPrice, Fee: input.Fee, OccurredAt: input.OccurredAt.UTC(), Note: input.Note, CreatedAt: time.Now().UTC()}
		_, computeErr := entity.ComputePortfolioPosition(append(existing, candidate))
		if errors.Is(computeErr, entity.ErrPortfolioInsufficientQuantity) {
			return portfoliorepo.ErrPortfolioInsufficientQuantity
		}
		if computeErr != nil {
			return portfoliorepo.ErrPortfolioTradeInvalid
		}
		if err := tx.Create(&candidate).Error; err != nil {
			return err
		}
		candidate.Quantity = entity.FormatPortfolioQuantity(candidate.QuantityScaled)
		created = candidate
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *PortfolioPostgresRepository) Summary(ownerID string) (*entity.PortfolioSummary, error) {
	assets, err := r.ListAssets(ownerID)
	if err != nil {
		return nil, err
	}
	trades := make([]entity.PortfolioTrade, 0)
	if err := r.db.Where("owner_id = ?", ownerID).Order("occurred_at ASC, created_at ASC, id ASC").Find(&trades).Error; err != nil {
		return nil, err
	}
	byAsset := make(map[string][]entity.PortfolioTrade, len(assets))
	for _, trade := range trades {
		byAsset[trade.AssetID] = append(byAsset[trade.AssetID], trade)
	}
	products := make(map[string]entity.MarketProduct)
	quotes := make(map[string]entity.MarketQuote)
	productIDs := make([]string, 0)
	for _, asset := range assets {
		if asset.MarketProductID != nil && strings.TrimSpace(*asset.MarketProductID) != "" {
			productIDs = append(productIDs, strings.TrimSpace(*asset.MarketProductID))
		}
	}
	if len(productIDs) > 0 {
		var marketProducts []entity.MarketProduct
		if err := r.db.Where("id IN ?", productIDs).Find(&marketProducts).Error; err != nil {
			return nil, err
		}
		for _, product := range marketProducts {
			products[product.ID] = product
		}
		var marketQuotes []entity.MarketQuote
		if err := r.db.Where("product_id IN ?", productIDs).Find(&marketQuotes).Error; err != nil {
			return nil, err
		}
		for _, quote := range marketQuotes {
			quotes[quote.ProductID] = quote
		}
	}
	result := &entity.PortfolioSummary{Assets: assets, Positions: make([]entity.PortfolioPosition, 0, len(assets)), Trades: []entity.PortfolioTrade{}}
	var knownMarket int64
	knownCount := 0
	now := time.Now().UTC()
	for _, asset := range assets {
		position, err := entity.ComputePortfolioPosition(byAsset[asset.ID])
		if err != nil {
			return nil, err
		}
		position.AssetID, position.Symbol, position.Name = asset.ID, asset.Symbol, asset.Name
		position.UnitLabel, position.PriceMode = asset.UnitLabel, asset.PriceMode
		if position.PriceMode == "" {
			position.PriceMode = entity.PortfolioPriceModeManual
		}
		price := asset.LatestPrice
		priceAsOf := asset.LatestPriceAt
		priceSource := asset.LatestPriceSource
		priceStatus := entity.PortfolioQuoteStatusUnknown
		if position.PriceMode == entity.PortfolioPriceModeAuto {
			price = nil
			if asset.MarketProductID != nil {
				if product, ok := products[*asset.MarketProductID]; ok {
					source := product.ProviderKey + ":" + product.ExternalProductID
					priceSource = &source
				}
				if quote, ok := quotes[*asset.MarketProductID]; ok && quote.Price > 0 && !quote.ObservedAt.After(now.Add(5*time.Minute)) {
					price = &quote.Price
					observedAt := quote.ObservedAt
					priceAsOf = &observedAt
					priceStatus = entity.PortfolioQuoteStatusFresh
					if now.Sub(quote.ObservedAt) > entity.PortfolioQuoteFreshness {
						priceStatus = entity.PortfolioQuoteStatusStale
					}
				}
			}
		} else if price != nil {
			priceStatus = entity.PortfolioQuoteStatusFresh
		}
		position.PriceSource, position.PriceAsOf, position.PriceStatus = priceSource, priceAsOf, priceStatus
		position, err = entity.PortfolioMarketValue(position, price)
		if err != nil {
			return nil, err
		}
		result.Positions = append(result.Positions, position)
		if position.CostBasis > 0 && result.TotalCostBasis > math.MaxInt64-position.CostBasis {
			return nil, entity.ErrPortfolioOverflow
		}
		result.TotalCostBasis += position.CostBasis
		if position.QuantityScaled > 0 && position.MarketValue == nil {
			result.UnknownAssetCount++
		}
		if position.QuantityScaled > 0 && position.PriceStatus == entity.PortfolioQuoteStatusStale {
			result.StaleAssetCount++
		}
		if position.MarketValue != nil {
			knownCount++
			if *position.MarketValue > 0 && knownMarket > math.MaxInt64-*position.MarketValue {
				return nil, entity.ErrPortfolioOverflow
			}
			knownMarket += *position.MarketValue
		}
	}
	if knownCount > 0 {
		result.KnownMarketValue = &knownMarket
	}
	if result.UnknownAssetCount == 0 && len(assets) > 0 {
		result.TotalMarketValue = &knownMarket
		totalUnrealized := knownMarket - result.TotalCostBasis
		result.TotalUnrealizedPnL = &totalUnrealized
	}
	return result, nil
}

func (r *PortfolioPostgresRepository) ensureAsset(ownerID, assetID string) error {
	var asset entity.PortfolioAsset
	if err := r.db.Select("id").Where("owner_id = ? AND id = ?", ownerID, assetID).First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return portfoliorepo.ErrPortfolioAssetNotFound
		}
		return err
	}
	return nil
}

func portfolioUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key") || strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
