package repository

import (
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrMarketProductNotFound     = errors.New("market product not found")
	ErrMarketProductInvalid      = errors.New("market product is invalid")
	ErrMarketBindingInvalid      = errors.New("market price source binding is invalid")
	ErrMarketBindingUnitMismatch = errors.New("market product unit does not match asset unit")
	ErrMarketBindingHasTrades    = errors.New("market product cannot change after trades")
	ErrMarketQuoteUnavailable    = errors.New("market quote is unavailable")
	ErrMarketPollLeaseLost       = errors.New("market poll lease lost")
)

type MarketPollClaim struct {
	Claim    entity.MarketPollClaim
	Acquired bool
}

type MarketRepository interface {
	ListEnabledMarketProducts() ([]entity.MarketProduct, error)
	FindMarketProduct(productID string) (*entity.MarketProduct, error)
	ClaimMarketPoll(providerKey string, slotAt, now time.Time, lease time.Duration) (MarketPollClaim, error)
	CompleteMarketPoll(claim entity.MarketPollClaim, quotes []entity.MarketQuote, now time.Time) (int, error)
	FailMarketPoll(claim entity.MarketPollClaim, now time.Time, cause error) error
}
