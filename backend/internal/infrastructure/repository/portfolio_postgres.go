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
	row := &entity.PortfolioAsset{ID: uuid.NewString(), OwnerID: ownerID, Symbol: symbol, Name: name}
	if err := r.db.Create(row).Error; err != nil {
		if portfolioUniqueViolation(err) {
			return nil, portfoliorepo.ErrPortfolioAssetConflict
		}
		return nil, err
	}
	return row, nil
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
	row.LatestPrice = update.LatestPrice
	row.LatestPriceAt = update.LatestPriceAt
	if update.LatestPrice == nil {
		row.LatestPriceAt = nil
	}
	if err := r.db.Model(&row).Updates(map[string]any{"latest_price": row.LatestPrice, "latest_price_at": row.LatestPriceAt}).Error; err != nil {
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
	result := &entity.PortfolioSummary{Assets: assets, Positions: make([]entity.PortfolioPosition, 0, len(assets)), Trades: []entity.PortfolioTrade{}}
	var totalMarket int64
	pricedCount := 0
	for _, asset := range assets {
		position, err := entity.ComputePortfolioPosition(byAsset[asset.ID])
		if err != nil {
			return nil, err
		}
		position.AssetID, position.Symbol, position.Name = asset.ID, asset.Symbol, asset.Name
		position, err = entity.PortfolioMarketValue(position, asset.LatestPrice)
		if err != nil {
			return nil, err
		}
		result.Positions = append(result.Positions, position)
		if position.CostBasis > 0 && result.TotalCostBasis > math.MaxInt64-position.CostBasis {
			return nil, entity.ErrPortfolioOverflow
		}
		result.TotalCostBasis += position.CostBasis
		if position.MarketValue != nil {
			pricedCount++
			if *position.MarketValue > 0 && totalMarket > math.MaxInt64-*position.MarketValue {
				return nil, entity.ErrPortfolioOverflow
			}
			totalMarket += *position.MarketValue
		}
	}
	if pricedCount == len(assets) && len(assets) > 0 {
		result.TotalMarketValue = &totalMarket
		totalUnrealized := totalMarket - result.TotalCostBasis
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
