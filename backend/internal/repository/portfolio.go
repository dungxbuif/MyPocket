package repository

import (
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrPortfolioAssetNotFound        = errors.New("portfolio asset not found")
	ErrPortfolioAssetConflict        = errors.New("portfolio asset already exists")
	ErrPortfolioAssetInvalid         = errors.New("portfolio asset is invalid")
	ErrPortfolioTradeNotFound        = errors.New("portfolio trade not found")
	ErrPortfolioTradeInvalid         = errors.New("portfolio trade is invalid")
	ErrPortfolioInsufficientQuantity = errors.New("portfolio position is insufficient for sell")
)

type PortfolioAssetInput struct {
	Symbol string
	Name   string
}

type PortfolioPriceUpdate struct {
	LatestPrice   *int64
	LatestPriceAt *time.Time
}

type PortfolioTradeInput struct {
	Side           string
	QuantityScaled int64
	UnitPrice      int64
	Fee            int64
	OccurredAt     time.Time
	Note           *string
}

type PortfolioRepository interface {
	ListAssets(ownerID string) ([]entity.PortfolioAsset, error)
	CreateAsset(ownerID string, input PortfolioAssetInput) (*entity.PortfolioAsset, error)
	UpdateAssetPrice(ownerID, assetID string, update PortfolioPriceUpdate) (*entity.PortfolioAsset, error)
	ListTrades(ownerID, assetID string) ([]entity.PortfolioTrade, error)
	CreateTrade(ownerID, assetID string, input PortfolioTradeInput) (*entity.PortfolioTrade, error)
	Summary(ownerID string) (*entity.PortfolioSummary, error)
}
