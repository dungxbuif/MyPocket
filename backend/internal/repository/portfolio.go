package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrPortfolioAssetNotFound        = errors.New("portfolio asset not found")
	ErrPortfolioAssetConflict        = errors.New("portfolio asset already exists")
	ErrPortfolioAssetHasTrades       = errors.New("portfolio asset has trades")
	ErrPortfolioAssetInvalid         = errors.New("portfolio asset is invalid")
	ErrPortfolioTradeNotFound        = errors.New("portfolio trade not found")
	ErrPortfolioTradeInvalid         = errors.New("portfolio trade is invalid")
	ErrPortfolioInsufficientQuantity = errors.New("portfolio position is insufficient for sell")
	ErrPortfolioAssetConflictVersion = errors.New("portfolio asset version is stale")
	ErrPortfolioCorrectionConflict   = errors.New("portfolio correction conflicts with existing request")
	ErrPortfolioCorrectionPending    = errors.New("portfolio correction is still pending")
	ErrPortfolioCorrectionInvalid    = errors.New("portfolio correction is invalid")
)

type PortfolioAssetInput struct {
	Symbol string
	Name   string
}

// PortfolioAssetUpdate changes descriptive metadata only. Trades remain
// append-only and are never rewritten when an asset is renamed.
type PortfolioAssetUpdate struct {
	Symbol *string
	Name   *string
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

type PortfolioTradeCorrectionInput struct {
	SourceTradeID        string
	ExpectedAssetVersion int64
	Operation            string
	Reason               string
	Replacement          *PortfolioTradeInput
}

type PortfolioTradeCorrectionResult struct {
	RevisionID       string                   `json:"revision_id"`
	AssetVersion     int64                    `json:"asset_version"`
	Before           entity.PortfolioPosition `json:"before"`
	After            entity.PortfolioPosition `json:"after"`
	AffectedTradeIDs []string                 `json:"affected_trade_ids"`
}

func PortfolioCorrectionHash(input PortfolioTradeCorrectionInput) string {
	data, _ := json.Marshal(input)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

type PortfolioRepository interface {
	ListAssets(ownerID string) ([]entity.PortfolioAsset, error)
	CreateAsset(ownerID string, input PortfolioAssetInput) (*entity.PortfolioAsset, error)
	UpdateAsset(ownerID, assetID string, update PortfolioAssetUpdate) (*entity.PortfolioAsset, error)
	DeleteAsset(ownerID, assetID string) error
	UpdateAssetPrice(ownerID, assetID string, update PortfolioPriceUpdate) (*entity.PortfolioAsset, error)
	ListTrades(ownerID, assetID string) ([]entity.PortfolioTrade, error)
	CreateTrade(ownerID, assetID string, input PortfolioTradeInput) (*entity.PortfolioTrade, error)
	PreviewTradeCorrection(ownerID, assetID string, input PortfolioTradeCorrectionInput) (*PortfolioTradeCorrectionResult, error)
	CorrectTrade(ownerID, assetID string, input PortfolioTradeCorrectionInput, idempotencyKey, requestHash string) (*PortfolioTradeCorrectionResult, bool, error)
	Summary(ownerID string) (*entity.PortfolioSummary, error)
}
