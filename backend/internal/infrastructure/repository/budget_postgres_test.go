package repository

import (
	"errors"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

func TestBudgetUpdateDoesNotResurrectMissingRow(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner, id := uuid.NewString(), uuid.NewString()
	user := entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner}
	if err = db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	defer db.Where("id = ?", owner).Delete(&entity.User{})
	repo := NewBudgetPostgresRepository(db)
	budget := entity.Budget{ID: id, Name: "race fixture", LimitAmount: 100, StartAt: time.Now(), EndAt: time.Now().Add(time.Hour)}
	if err = repo.Save(owner, &budget, true); err != nil {
		t.Fatal(err)
	}
	// A disappearing row after the read must never trigger GORM Save's insert fallback.
	db.Callback().Update().Before("gorm:update").Register("test:disappear_budget", func(tx *gorm.DB) {
		if tx.Statement.Table == "budgets" {
			if e := tx.Exec("DELETE FROM budgets WHERE id = ? AND owner_id = ?", id, owner).Error; e != nil {
				tx.AddError(e)
			}
		}
	})
	defer db.Callback().Update().Remove("test:disappear_budget")
	budget.Name = "updated"
	err = repo.Save(owner, &budget, false)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing-row update must return not found, got %v", err)
	}
}
