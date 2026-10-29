package repository

import (
	"errors"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrTravelInvalid     = errors.New("invalid travel event")
	ErrTravelNotFound    = errors.New("travel event not found")
	ErrTravelTransaction = errors.New("invalid travel transaction link")
)

type TravelUpdate struct {
	Name     string
	Context  *string
	StartsOn *entity.CalendarDate
	EndsOn   *entity.CalendarDate
}

type TravelRepository interface {
	List(ownerID string) ([]entity.TravelEvent, error)
	Find(ownerID, id string) (*entity.TravelEvent, error)
	Active(ownerID string) (*entity.TravelEvent, error)
	Create(ownerID string, event *entity.TravelEvent) error
	Update(ownerID, id string, updates TravelUpdate) (*entity.TravelEvent, error)
	Delete(ownerID, id string) error
	SetActive(ownerID, id string, active bool) (*entity.TravelEvent, error)
	LinkTransaction(ownerID, transactionID string, eventID *string) (*entity.Transaction, error)
}
