package repository

import (
	"errors"
	"github.com/mypocket/backend/internal/entity"
	"time"
)

var ErrBudgetSeriesManaged = errors.New("generated budget is managed by a series")
var ErrBudgetSeriesVersion = errors.New("budget series version conflict")
var ErrBudgetSeriesStopped = errors.New("budget series is stopped")

type BudgetSeriesRepository interface {
	List(owner string, now time.Time) ([]entity.BudgetSeries, error)
	Mutate(owner string, series *entity.BudgetSeries, config *entity.BudgetSeriesConfig, operation, key, hash string, expectedVersion int64, now time.Time) (*entity.BudgetSeries, bool, error)
	Ensure(owner string, now time.Time) (int, error)
}
