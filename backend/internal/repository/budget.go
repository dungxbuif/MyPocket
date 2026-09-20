package repository

import (
	"errors"
	"github.com/mypocket/backend/internal/entity"
)

var ErrBudgetConflict = errors.New("budget scope overlaps")
var ErrBudgetEnded = errors.New("budget has ended")

type BudgetRepository interface {
	List(owner string) ([]entity.Budget, error)
	Save(owner string, budget *entity.Budget, creating bool) error
	Delete(owner, id string) error
}
