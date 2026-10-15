package repository

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
)

var ErrJarInvalid = errors.New("invalid jar configuration or period")

type JarRepository interface {
	ListMonth(ownerID, month, timezone string) (*entity.JarMonthSummary, error)
	Create(ownerID, month, timezone, name, allocationMode string, allocationAmount *int64, allocationPercentBPS *int) (*entity.JarMonthConfig, error)
	UpdateMonthConfig(ownerID, jarID, month, name, allocationMode string, allocationAmount *int64, allocationPercentBPS *int) (*entity.JarMonthConfig, error)
	RemoveMonthConfig(ownerID, jarID, month string) error
	FindMonthConfig(ownerID, jarID, month string, includeInactive bool) (*entity.JarMonthConfig, error)
	Cumulative(ownerID, jarID, fromMonth, toMonth, timezone string) (*entity.JarCumulativeSummary, error)
}

// JarCreateMutationRepository is the durable replay boundary for a confirmed
// Finance Assistant jar action. Implementations must atomically create the
// month configuration and persist the response under owner/operation/key.
type JarCreateMutationRepository interface {
	CreateIdempotent(ownerID, month, timezone, name, allocationMode string, allocationAmount *int64, allocationPercentBPS *int, idempotencyKey, requestHash string) (*entity.JarMonthConfig, bool, error)
}
