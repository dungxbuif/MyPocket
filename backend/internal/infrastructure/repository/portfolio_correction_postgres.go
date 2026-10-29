package repository

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	portfoliorepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *PortfolioPostgresRepository) PreviewTradeCorrection(ownerID, assetID string, input portfoliorepo.PortfolioTradeCorrectionInput) (*portfoliorepo.PortfolioTradeCorrectionResult, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" {
		return nil, portfoliorepo.ErrPortfolioAssetInvalid
	}
	var asset entity.PortfolioAsset
	if err := r.db.Where("owner_id = ? AND id = ?", ownerID, assetID).First(&asset).Error; err != nil {
		return nil, portfolioCorrectionMapDBError(err)
	}
	return r.previewCorrection(r.db, ownerID, asset, input)
}

func (r *PortfolioPostgresRepository) CorrectTrade(ownerID, assetID string, input portfoliorepo.PortfolioTradeCorrectionInput, idempotencyKey, requestHash string) (*portfoliorepo.PortfolioTradeCorrectionResult, bool, error) {
	if r == nil || r.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(assetID) == "" || !validPortfolioCorrectionKey(idempotencyKey) || strings.TrimSpace(requestHash) == "" {
		return nil, false, portfoliorepo.ErrPortfolioCorrectionInvalid
	}
	var result portfoliorepo.PortfolioTradeCorrectionResult
	replayed := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		mutation, created, err := reservePortfolioMutation(tx, ownerID, idempotencyKey, requestHash)
		if err != nil {
			return err
		}
		if !created {
			if mutation.Status != "completed" {
				return portfoliorepo.ErrPortfolioCorrectionPending
			}
			if err := json.Unmarshal([]byte(mutation.ResponseJSON), &result); err != nil {
				return err
			}
			replayed = true
			return nil
		}
		var asset entity.PortfolioAsset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, assetID).First(&asset).Error; err != nil {
			return portfolioCorrectionMapDBError(err)
		}
		if err := validatePortfolioCorrectionVersion(input.ExpectedAssetVersion, asset.Version); err != nil {
			return err
		}
		preview, err := r.previewCorrection(tx, ownerID, asset, input)
		if err != nil {
			return err
		}
		var source entity.PortfolioTrade
		if err := tx.Where("owner_id = ? AND asset_id = ? AND id = ?", ownerID, assetID, input.SourceTradeID).First(&source).Error; err != nil {
			return portfolioCorrectionMapDBError(err)
		}
		var count int64
		if err := tx.Model(&entity.PortfolioTradeRevision{}).Where("owner_id = ? AND source_trade_id = ?", ownerID, input.SourceTradeID).Count(&count).Error; err != nil {
			return err
		}
		revision := entity.PortfolioTradeRevision{ID: uuid.NewString(), OwnerID: ownerID, AssetID: assetID, SourceTradeID: input.SourceTradeID, Revision: count + 1, Operation: input.Operation, Reason: strings.TrimSpace(input.Reason), ActorID: ownerID, CreatedAt: time.Now().UTC()}
		if input.Replacement != nil {
			revision.Side = input.Replacement.Side
			revision.QuantityScaled = input.Replacement.QuantityScaled
			revision.UnitPrice = input.Replacement.UnitPrice
			revision.Fee = input.Replacement.Fee
			revision.OccurredAt = input.Replacement.OccurredAt.UTC()
			revision.Note = input.Replacement.Note
		}
		if err := tx.Create(&revision).Error; err != nil {
			return err
		}
		asset.Version++
		if err := tx.Model(&asset).Update("version", asset.Version).Error; err != nil {
			return err
		}
		preview.RevisionID, preview.AssetVersion = revision.ID, asset.Version
		result = *preview
		return completePortfolioMutation(tx, mutation.ID, result)
	})
	if err != nil {
		return nil, false, err
	}
	return &result, replayed, nil
}

func (r *PortfolioPostgresRepository) previewCorrection(db *gorm.DB, ownerID string, asset entity.PortfolioAsset, input portfoliorepo.PortfolioTradeCorrectionInput) (*portfoliorepo.PortfolioTradeCorrectionResult, error) {
	if err := validatePortfolioCorrectionVersion(input.ExpectedAssetVersion, asset.Version); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.SourceTradeID) == "" || strings.TrimSpace(input.Reason) == "" || (input.Operation != entity.PortfolioTradeRevisionReplace && input.Operation != entity.PortfolioTradeRevisionVoid) {
		return nil, portfoliorepo.ErrPortfolioCorrectionInvalid
	}
	var trades []entity.PortfolioTrade
	if err := db.Where("owner_id = ? AND asset_id = ?", ownerID, asset.ID).Order("occurred_at ASC, created_at ASC, id ASC").Find(&trades).Error; err != nil {
		return nil, err
	}
	var revisions []entity.PortfolioTradeRevision
	if err := db.Where("owner_id = ? AND asset_id = ?", ownerID, asset.ID).Order("source_trade_id ASC, revision ASC").Find(&revisions).Error; err != nil {
		return nil, err
	}
	effective, err := entity.ApplyPortfolioTradeRevisions(trades, revisions)
	if err != nil {
		return nil, err
	}
	before, err := entity.ComputePortfolioPosition(effective)
	if err != nil {
		return nil, err
	}
	var source entity.PortfolioTrade
	found := false
	for _, row := range trades {
		if row.ID == input.SourceTradeID {
			source, found = row, true
			break
		}
	}
	if !found {
		return nil, portfoliorepo.ErrPortfolioTradeNotFound
	}
	proposal := entity.PortfolioTradeRevision{SourceTradeID: input.SourceTradeID, Revision: nextPortfolioRevision(revisions, input.SourceTradeID), Operation: input.Operation, Reason: input.Reason}
	if input.Replacement != nil {
		proposal.Side, proposal.QuantityScaled, proposal.UnitPrice, proposal.Fee, proposal.OccurredAt, proposal.Note = input.Replacement.Side, input.Replacement.QuantityScaled, input.Replacement.UnitPrice, input.Replacement.Fee, input.Replacement.OccurredAt, input.Replacement.Note
	}
	afterTrades, err := entity.ApplyPortfolioTradeRevisions(trades, append(revisions, proposal))
	if err != nil {
		return nil, err
	}
	after, err := entity.ComputePortfolioPosition(afterTrades)
	if err != nil {
		return nil, err
	}
	affected := make([]string, 0)
	for _, row := range trades {
		if row.ID != source.ID && !row.OccurredAt.Before(source.OccurredAt) {
			affected = append(affected, row.ID)
		}
	}
	return &portfoliorepo.PortfolioTradeCorrectionResult{AssetVersion: asset.Version, Before: before, After: after, AffectedTradeIDs: affected}, nil
}

func validatePortfolioCorrectionVersion(expected, actual int64) error {
	if expected <= 0 {
		return portfoliorepo.ErrPortfolioCorrectionInvalid
	}
	if actual != expected {
		return portfoliorepo.ErrPortfolioAssetConflictVersion
	}
	return nil
}

func nextPortfolioRevision(revisions []entity.PortfolioTradeRevision, sourceID string) int64 {
	var max int64
	for _, row := range revisions {
		if row.SourceTradeID == sourceID && row.Revision > max {
			max = row.Revision
		}
	}
	return max + 1
}

func reservePortfolioMutation(tx *gorm.DB, owner, key, requestHash string) (entity.TransactionMutation, bool, error) {
	mutation := entity.TransactionMutation{ID: uuid.NewString(), OwnerID: owner, Operation: "portfolio.trade_correction", IdempotencyKey: key, RequestHash: requestHash, Status: "pending", ResponseJSON: "{}"}
	result := tx.Exec(`INSERT INTO transaction_mutations (id,owner_id,operation,idempotency_key,request_hash,status,response_json) VALUES (?,?,?,?,?,?,'{}'::jsonb) ON CONFLICT (owner_id,operation,idempotency_key) DO NOTHING`, mutation.ID, owner, mutation.Operation, key, requestHash, mutation.Status)
	if result.Error != nil {
		return entity.TransactionMutation{}, false, result.Error
	}
	if result.RowsAffected == 1 {
		return mutation, true, nil
	}
	var persisted entity.TransactionMutation
	if err := tx.Where("owner_id = ? AND operation = ? AND idempotency_key = ?", owner, mutation.Operation, key).First(&persisted).Error; err != nil {
		return entity.TransactionMutation{}, false, err
	}
	if persisted.RequestHash != requestHash {
		return entity.TransactionMutation{}, false, portfoliorepo.ErrPortfolioCorrectionConflict
	}
	return persisted, false, nil
}

func completePortfolioMutation(tx *gorm.DB, id string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return tx.Model(&entity.TransactionMutation{}).Where("id = ?", id).Updates(map[string]any{"status": "completed", "response_json": string(encoded)}).Error
}

func validPortfolioCorrectionKey(key string) bool {
	if strings.TrimSpace(key) == "" || len(key) > 128 {
		return false
	}
	for _, c := range key {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

func portfolioCorrectionMapDBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return portfoliorepo.ErrPortfolioAssetNotFound
	}
	return err
}
