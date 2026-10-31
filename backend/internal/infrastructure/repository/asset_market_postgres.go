package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	marketrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewAssetMarketPostgresRepository(db *gorm.DB) marketrepo.MarketRepository {
	return &AssetMarketPostgresRepository{db: db}
}

type AssetMarketPostgresRepository struct{ db *gorm.DB }

func validateMarketProductForAsset(asset entity.PortfolioAsset, product entity.MarketProduct) error {
	if !product.Enabled || strings.TrimSpace(product.ID) == "" || strings.TrimSpace(product.ProviderKey) == "" || strings.TrimSpace(product.ExternalProductID) == "" || strings.TrimSpace(product.Currency) != "VND" || product.ValuationSide != "buyback" {
		return marketrepo.ErrMarketBindingInvalid
	}
	if !asset.UnitConfirmed || asset.UnitLabel == nil || strings.TrimSpace(*asset.UnitLabel) == "" {
		return marketrepo.ErrMarketBindingInvalid
	}
	if strings.TrimSpace(product.Unit) != strings.TrimSpace(*asset.UnitLabel) {
		return marketrepo.ErrMarketBindingUnitMismatch
	}
	return nil
}

func marketProductSource(product entity.MarketProduct) string {
	return strings.TrimSpace(product.ProviderKey) + ":" + strings.TrimSpace(product.ExternalProductID)
}

func (r *AssetMarketPostgresRepository) ListEnabledMarketProducts() ([]entity.MarketProduct, error) {
	if r == nil || r.db == nil {
		return nil, marketrepo.ErrMarketProductInvalid
	}
	rows := make([]entity.MarketProduct, 0)
	if err := r.db.Where("enabled = ?", true).Order("provider_key ASC, name ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *AssetMarketPostgresRepository) FindMarketProduct(productID string) (*entity.MarketProduct, error) {
	if r == nil || r.db == nil || strings.TrimSpace(productID) == "" {
		return nil, marketrepo.ErrMarketProductInvalid
	}
	var row entity.MarketProduct
	if err := r.db.Where("id = ?", strings.TrimSpace(productID)).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, marketrepo.ErrMarketProductNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (r *PortfolioPostgresRepository) UpdateAssetPriceSource(ownerID, assetID string, input marketrepo.PortfolioPriceSourceInput) (*entity.PortfolioAsset, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" || input.ExpectedVersion <= 0 {
		return nil, marketrepo.ErrPortfolioAssetInvalid
	}
	if input.PriceMode != entity.PortfolioPriceModeManual && input.PriceMode != entity.PortfolioPriceModeAuto {
		return nil, marketrepo.ErrMarketBindingInvalid
	}
	if input.PriceMode == entity.PortfolioPriceModeManual && input.MarketProductID != nil {
		return nil, marketrepo.ErrMarketBindingInvalid
	}
	if input.PriceMode == entity.PortfolioPriceModeAuto && (input.MarketProductID == nil || strings.TrimSpace(*input.MarketProductID) == "") {
		return nil, marketrepo.ErrMarketBindingInvalid
	}
	var row entity.PortfolioAsset
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, assetID).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return marketrepo.ErrPortfolioAssetNotFound
			}
			return err
		}
		if row.Version != input.ExpectedVersion {
			return marketrepo.ErrPortfolioAssetConflictVersion
		}
		if input.PriceMode == entity.PortfolioPriceModeAuto {
			productID := strings.TrimSpace(*input.MarketProductID)
			var product entity.MarketProduct
			if err := tx.Where("id = ?", productID).First(&product).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return marketrepo.ErrMarketProductNotFound
				}
				return err
			}
			if err := validateMarketProductForAsset(row, product); err != nil {
				return err
			}
			if row.MarketProductID != nil && strings.TrimSpace(*row.MarketProductID) != productID {
				var tradeCount int64
				if err := tx.Model(&entity.PortfolioTrade{}).Where("owner_id = ? AND asset_id = ?", ownerID, assetID).Count(&tradeCount).Error; err != nil {
					return err
				}
				if tradeCount > 0 {
					return marketrepo.ErrMarketBindingHasTrades
				}
			}
			if row.MarketProductID != nil && strings.TrimSpace(*row.MarketProductID) == productID && row.PriceMode == entity.PortfolioPriceModeAuto {
				return nil
			}
			row.MarketProductID = &productID
			row.PriceMode = entity.PortfolioPriceModeAuto
			row.LatestPrice = nil
			row.LatestPriceAt = nil
			row.LatestPriceSource = nil
		} else {
			if row.MarketProductID != nil {
				var product entity.MarketProduct
				if err := tx.Where("id = ?", *row.MarketProductID).First(&product).Error; err == nil {
					var quote entity.MarketQuote
					if quoteErr := tx.Where("product_id = ?", product.ID).First(&quote).Error; quoteErr == nil {
						price := quote.Price
						observedAt := quote.ObservedAt
						source := marketProductSource(product)
						row.LatestPrice, row.LatestPriceAt, row.LatestPriceSource = &price, &observedAt, &source
					}
				}
			}
			row.MarketProductID = nil
			row.PriceMode = entity.PortfolioPriceModeManual
		}
		row.Version++
		return tx.Model(&row).Updates(map[string]any{
			"market_product_id":   row.MarketProductID,
			"price_mode":          row.PriceMode,
			"latest_price":        row.LatestPrice,
			"latest_price_at":     row.LatestPriceAt,
			"latest_price_source": row.LatestPriceSource,
			"version":             row.Version,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *AssetMarketPostgresRepository) ClaimMarketPoll(providerKey string, slotAt, now time.Time, lease time.Duration) (marketrepo.MarketPollClaim, error) {
	if r == nil || r.db == nil || strings.TrimSpace(providerKey) == "" || slotAt.IsZero() || now.IsZero() || lease <= 0 {
		return marketrepo.MarketPollClaim{}, marketrepo.ErrMarketProductInvalid
	}
	var result marketrepo.MarketPollClaim
	err := r.db.Transaction(func(tx *gorm.DB) error {
		candidate := entity.MarketPollClaim{ProviderKey: providerKey, SlotAt: slotAt.UTC(), LeaseToken: uuid.NewString(), LeaseUntil: now.UTC().Add(lease), Status: "running"}
		if err := tx.Exec(`INSERT INTO market_poll_claims (provider_key, slot_at, lease_token, lease_until, status) VALUES (?, ?, ?, ?, ?) ON CONFLICT (provider_key, slot_at) DO NOTHING`, candidate.ProviderKey, candidate.SlotAt, candidate.LeaseToken, candidate.LeaseUntil, candidate.Status).Error; err != nil {
			return err
		}
		var row entity.MarketPollClaim
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_key = ? AND slot_at = ?", providerKey, slotAt.UTC()).First(&row).Error; err != nil {
			return err
		}
		result.Claim = row
		if row.Status == "completed" || row.Status == "failed" || row.LeaseUntil.After(now.UTC()) {
			return nil
		}
		row.LeaseToken = uuid.NewString()
		row.LeaseUntil = now.UTC().Add(lease)
		row.Status = "running"
		row.Error = nil
		if err := tx.Model(&row).Updates(map[string]any{"lease_token": row.LeaseToken, "lease_until": row.LeaseUntil, "status": row.Status, "error_message": nil}).Error; err != nil {
			return err
		}
		result.Claim, result.Acquired = row, true
		return nil
	})
	if err != nil {
		return marketrepo.MarketPollClaim{}, err
	}
	return result, nil
}

func (r *AssetMarketPostgresRepository) CompleteMarketPoll(claim entity.MarketPollClaim, quotes []entity.MarketQuote, now time.Time) (int, error) {
	if r == nil || r.db == nil || strings.TrimSpace(claim.ProviderKey) == "" || strings.TrimSpace(claim.LeaseToken) == "" || claim.SlotAt.IsZero() || now.IsZero() {
		return 0, marketrepo.ErrMarketPollLeaseLost
	}
	accepted := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current entity.MarketPollClaim
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_key = ? AND slot_at = ?", claim.ProviderKey, claim.SlotAt.UTC()).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return marketrepo.ErrMarketPollLeaseLost
			}
			return err
		}
		if current.Status != "running" || current.LeaseToken != claim.LeaseToken || !current.LeaseUntil.After(now.UTC()) {
			return marketrepo.ErrMarketPollLeaseLost
		}
		for _, quote := range quotes {
			if strings.TrimSpace(quote.ProductID) == "" || quote.Price <= 0 || quote.ObservedAt.IsZero() || quote.FetchedAt.IsZero() {
				return marketrepo.ErrMarketQuoteUnavailable
			}
			var product entity.MarketProduct
			if err := tx.Where("id = ? AND provider_key = ? AND enabled = ?", quote.ProductID, claim.ProviderKey, true).First(&product).Error; err != nil {
				return marketrepo.ErrMarketQuoteUnavailable
			}
			result := tx.Exec(`INSERT INTO market_quotes (product_id, price, observed_at, fetched_at) VALUES (?, ?, ?, ?) ON CONFLICT (product_id) DO UPDATE SET price = EXCLUDED.price, observed_at = EXCLUDED.observed_at, fetched_at = EXCLUDED.fetched_at WHERE market_quotes.observed_at < EXCLUDED.observed_at`, quote.ProductID, quote.Price, quote.ObservedAt.UTC(), quote.FetchedAt.UTC())
			if result.Error != nil {
				return result.Error
			}
			accepted += int(result.RowsAffected)
		}
		return tx.Model(&current).Updates(map[string]any{"status": "completed", "error_message": nil}).Error
	})
	if err != nil {
		return 0, err
	}
	return accepted, nil
}

func (r *AssetMarketPostgresRepository) FailMarketPoll(claim entity.MarketPollClaim, now time.Time, cause error) error {
	if r == nil || r.db == nil || strings.TrimSpace(claim.ProviderKey) == "" || strings.TrimSpace(claim.LeaseToken) == "" || claim.SlotAt.IsZero() || now.IsZero() {
		return marketrepo.ErrMarketPollLeaseLost
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current entity.MarketPollClaim
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_key = ? AND slot_at = ?", claim.ProviderKey, claim.SlotAt.UTC()).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return marketrepo.ErrMarketPollLeaseLost
			}
			return err
		}
		if current.Status != "running" || current.LeaseToken != claim.LeaseToken || !current.LeaseUntil.After(now.UTC()) {
			return marketrepo.ErrMarketPollLeaseLost
		}
		// Do not persist raw provider errors: transport errors can contain a
		// backend endpoint or credential-bearing query string.
		message := "market provider fetch failed"
		return tx.Model(&current).Updates(map[string]any{"status": "failed", "error_message": message}).Error
	})
	return err
}
