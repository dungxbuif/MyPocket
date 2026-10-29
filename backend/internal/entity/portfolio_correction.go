package entity

import (
	"errors"
	"sort"
	"time"
)

const (
	PortfolioTradeRevisionReplace = "replace"
	PortfolioTradeRevisionVoid    = "void"
)

var (
	ErrPortfolioRevisionInvalid  = errors.New("portfolio trade revision is invalid")
	ErrPortfolioRevisionConflict = errors.New("portfolio trade revision conflicts with existing revision")
)

// PortfolioTradeRevision is an immutable audit record describing the effective
// replacement of one original trade. The original trade row is never changed.
// Replacement fields are ignored for a void operation.
type PortfolioTradeRevision struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OwnerID        string    `json:"owner_id" gorm:"index;not null"`
	AssetID        string    `json:"asset_id" gorm:"index;not null"`
	SourceTradeID  string    `json:"source_trade_id" gorm:"index;not null"`
	Revision       int64     `json:"revision" gorm:"not null"`
	Operation      string    `json:"operation" gorm:"not null"`
	Side           string    `json:"side,omitempty"`
	QuantityScaled int64     `json:"-"`
	UnitPrice      int64     `json:"unit_price,omitempty"`
	Fee            int64     `json:"fee,omitempty"`
	OccurredAt     time.Time `json:"occurred_at,omitempty"`
	Note           *string   `json:"note,omitempty"`
	Reason         string    `json:"reason" gorm:"not null"`
	ActorID        string    `json:"actor_id" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at"`
}

// ApplyPortfolioTradeRevisions returns the effective immutable trade history.
// Revisions are applied per source trade and the latest revision wins. The
// original ID/CreatedAt are retained as deterministic tie-breakers; an
// intentionally changed OccurredAt changes financial chronology.
func ApplyPortfolioTradeRevisions(trades []PortfolioTrade, revisions []PortfolioTradeRevision) ([]PortfolioTrade, error) {
	byID := make(map[string]PortfolioTrade, len(trades))
	for _, trade := range trades {
		if trade.ID == "" || trade.AssetID == "" {
			return nil, ErrPortfolioTradeInvalid
		}
		byID[trade.ID] = trade
	}
	latest := make(map[string]PortfolioTradeRevision, len(revisions))
	for _, revision := range revisions {
		if revision.SourceTradeID == "" || revision.Revision <= 0 || revision.Reason == "" ||
			(revision.Operation != PortfolioTradeRevisionReplace && revision.Operation != PortfolioTradeRevisionVoid) {
			return nil, ErrPortfolioRevisionInvalid
		}
		original, ok := byID[revision.SourceTradeID]
		if !ok || (revision.AssetID != "" && revision.AssetID != original.AssetID) {
			return nil, ErrPortfolioRevisionInvalid
		}
		prior, exists := latest[revision.SourceTradeID]
		if exists && revision.Revision <= prior.Revision {
			return nil, ErrPortfolioRevisionConflict
		}
		if revision.Operation == PortfolioTradeRevisionReplace &&
			(revision.QuantityScaled <= 0 || revision.UnitPrice < 0 || revision.Fee < 0 ||
				(revision.Side != PortfolioTradeBuy && revision.Side != PortfolioTradeSell)) {
			return nil, ErrPortfolioTradeInvalid
		}
		latest[revision.SourceTradeID] = revision
	}
	effective := make([]PortfolioTrade, 0, len(trades))
	for _, original := range trades {
		revision, ok := latest[original.ID]
		if !ok {
			effective = append(effective, original)
			continue
		}
		if revision.Operation == PortfolioTradeRevisionVoid {
			continue
		}
		replacement := original
		replacement.Side = revision.Side
		replacement.QuantityScaled = revision.QuantityScaled
		replacement.UnitPrice = revision.UnitPrice
		replacement.Fee = revision.Fee
		if !revision.OccurredAt.IsZero() {
			replacement.OccurredAt = revision.OccurredAt.UTC()
		}
		replacement.Note = revision.Note
		replacement.Quantity = FormatPortfolioQuantity(replacement.QuantityScaled)
		effective = append(effective, replacement)
	}
	sort.SliceStable(effective, func(i, j int) bool {
		if !effective[i].OccurredAt.Equal(effective[j].OccurredAt) {
			return effective[i].OccurredAt.Before(effective[j].OccurredAt)
		}
		if !effective[i].CreatedAt.Equal(effective[j].CreatedAt) {
			return effective[i].CreatedAt.Before(effective[j].CreatedAt)
		}
		return effective[i].ID < effective[j].ID
	})
	if _, err := ComputePortfolioPosition(effective); err != nil {
		return nil, err
	}
	return effective, nil
}
