package repository

import "github.com/mypocket/backend/internal/entity"

type InsiderReader interface {
	ReadInsider(ownerID, month, walletID, categoryID string) (*entity.InsiderSummary, error)
}
