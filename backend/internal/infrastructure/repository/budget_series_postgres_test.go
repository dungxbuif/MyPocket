package repository

import (
	"errors"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"sync"
	"testing"
	"time"
)

func TestBudgetSeriesConcurrentGenerationReplayAndProtection(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	user := entity.User{ID: owner, GoogleSubject: owner, Email: owner + "@test.invalid", Timezone: "UTC"}
	if err = db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	defer db.Where("id = ?", owner).Delete(&entity.User{})
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	start, _ := entity.ParseCalendarDate("2026-10-01")
	series := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
	cfg := entity.BudgetSeriesConfig{Name: "monthly", LimitAmount: 100, CarryoverEnabled: true}
	repo := NewBudgetSeriesPostgresRepository(db)
	made, replay, err := repo.Mutate(owner, &series, &cfg, "create", "key", "hash", 0, now)
	if err != nil || replay {
		t.Fatalf("create %v %v", replay, err)
	}
	again, replay, err := repo.Mutate(owner, &series, &cfg, "create", "key", "hash", 0, now)
	if err != nil || !replay || again.ID != made.ID {
		t.Fatalf("replay %v %v", replay, err)
	}
	if _, replay, err = repo.Mutate(owner, &series, &cfg, "create", "key", "hash", 0, now.AddDate(0, 6, 0)); err != nil || !replay {
		t.Fatalf("late replay %v %v", replay, err)
	}
	if _, _, err = repo.Mutate(owner, &series, &cfg, "create", "key", "different", 0, now); !errors.Is(err, contract.ErrMutationConflict) {
		t.Fatalf("conflict %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, e := repo.Ensure(owner, now.AddDate(0, 3, 0)); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	rows, e := NewBudgetPostgresRepository(db).List(owner)
	if e != nil || len(rows) != 4 {
		t.Fatalf("rows %d %v", len(rows), e)
	}
	if err = NewBudgetPostgresRepository(db).Delete(owner, rows[0].ID); !errors.Is(err, contract.ErrBudgetSeriesManaged) {
		t.Fatalf("delete %v", err)
	}
	edit := entity.BudgetSeries{ID: made.ID}
	cfg.Name = "new"
	if _, _, err = repo.Mutate(owner, &edit, &cfg, "update", "edit", "edit-hash", 99, now); !errors.Is(err, contract.ErrBudgetSeriesVersion) {
		t.Fatalf("version %v", err)
	}
	foreign := uuid.NewString()
	if _, _, err = repo.Mutate(foreign, &edit, nil, "stop", "stop", "stop-hash", 1, now); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("foreign %v", err)
	}
}

func budgetSeriesFixture(t *testing.T) (*gorm.DB, string, contract.BudgetSeriesRepository, time.Time) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	if err = db.Create(&entity.User{ID: owner, GoogleSubject: owner, Email: owner + "@test.invalid", Timezone: "UTC"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ?", owner).Delete(&entity.TransactionMutation{})
		if e := db.Where("id = ?", owner).Delete(&entity.User{}).Error; e != nil {
			t.Error(e)
		}
	})
	return db, owner, NewBudgetSeriesPostgresRepository(db), time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
}

func TestBudgetSeriesPendingRevisionReplacesAndPreservesSnapshots(t *testing.T) {
	db, owner, repo, now := budgetSeriesFixture(t)
	start, _ := entity.ParseCalendarDate("2026-10-01")
	s := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
	cfg := entity.BudgetSeriesConfig{Name: "October", LimitAmount: 100}
	result, _, err := repo.Mutate(owner, &s, &cfg, "create", "c", "c", 0, now)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Name = "November"
	cfg.LimitAmount = 200
	cfg.CarryoverEnabled = true
	result, _, err = repo.Mutate(owner, &s, &cfg, "update", "e1", "e1", result.Version, now)
	if err != nil {
		t.Fatal(err)
	}
	cfg.LimitAmount = 300
	result, _, err = repo.Mutate(owner, &s, &cfg, "update", "e2", "e2", result.Version, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.CurrentConfig.LimitAmount != 100 || result.PendingConfig == nil || result.PendingConfig.LimitAmount != 300 || result.PendingConfig.EffectiveStartDate.Format("2006-01-02") != "2026-11-01" {
		t.Fatalf("revisions %+v", result)
	}
	var configs int64
	if err = db.Model(&entity.BudgetSeriesConfig{}).Where("series_id = ?", s.ID).Count(&configs).Error; err != nil || configs != 2 {
		t.Fatalf("configs %d %v", configs, err)
	}
	if _, err = repo.Ensure(owner, now.AddDate(0, 1, 0)); err != nil {
		t.Fatal(err)
	}
	rows, err := NewBudgetPostgresRepository(db).List(owner)
	if err != nil || len(rows) != 2 || rows[0].LimitAmount != 300 || !rows[0].CarryoverEnabled || rows[1].LimitAmount != 100 {
		t.Fatalf("snapshots %+v %v", rows, err)
	}
	stopped, _, err := repo.Mutate(owner, &s, nil, "stop", "stop", "stop", result.Version, now.AddDate(0, 1, 0))
	if err != nil || stopped.Active {
		t.Fatalf("stop %v %+v", err, stopped)
	}
	if _, err = repo.Ensure(owner, now.AddDate(0, 5, 0)); err != nil {
		t.Fatal(err)
	}
	rows, err = NewBudgetPostgresRepository(db).List(owner)
	if err != nil || len(rows) != 2 {
		t.Fatalf("stop generated new rows %v %+v", err, rows)
	}
	replayed, replay, err := repo.Mutate(owner, &s, nil, "stop", "stop", "stop", result.Version, now.AddDate(0, 2, 0))
	if err != nil || !replay || replayed.Version != stopped.Version {
		t.Fatalf("stop replay %v %v", replay, err)
	}
}

func TestBudgetSeriesOverlapIsVisibleAndDoesNotSkipPeriod(t *testing.T) {
	db, owner, repo, now := budgetSeriesFixture(t)
	start, _ := entity.ParseCalendarDate("2026-10-01")
	end, _ := entity.ParseCalendarDate("2026-10-31")
	explicit := entity.Budget{ID: uuid.NewString(), Name: "explicit", LimitAmount: 50, StartDate: start, EndDate: end}
	budgets := NewBudgetPostgresRepository(db)
	if err := budgets.Save(owner, &explicit, true); err != nil {
		t.Fatal(err)
	}
	s := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
	cfg := entity.BudgetSeriesConfig{Name: "series", LimitAmount: 100}
	result, _, err := repo.Mutate(owner, &s, &cfg, "create", "c", "c", 0, now)
	if err != nil || result.GenerationError == nil || *result.GenerationError != "BUDGET_OVERLAP" {
		t.Fatalf("overlap %v %+v", err, result)
	}
	if err = budgets.Delete(owner, explicit.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Ensure(owner, now.AddDate(0, 1, 0)); err != nil {
		t.Fatal(err)
	}
	rows, err := budgets.List(owner)
	if err != nil || len(rows) != 2 || rows[1].StartDate.Format("2006-01-02") != "2026-10-01" {
		t.Fatalf("skipped %+v %v", rows, err)
	}
}

func TestBudgetSeriesDeletedScopeStopsBeforeBudgetCascade(t *testing.T) {
	for _, scope := range []string{"wallet", "category"} {
		t.Run(scope, func(t *testing.T) {
			db, owner, repo, now := budgetSeriesFixture(t)
			id := uuid.NewString()
			cfg := entity.BudgetSeriesConfig{Name: "scoped", LimitAmount: 100}
			if scope == "wallet" {
				if err := db.Create(&entity.Wallet{ID: id, OwnerID: owner, Name: "scope", Type: entity.WalletTypeBasic, Currency: "VND"}).Error; err != nil {
					t.Fatal(err)
				}
				cfg.WalletID = &id
			} else {
				if err := db.Create(&entity.Category{ID: id, OwnerID: &owner, Name: "scope", Kind: entity.TransactionTypeExpense, IsActive: true}).Error; err != nil {
					t.Fatal(err)
				}
				cfg.CategoryID = &id
			}
			start, _ := entity.ParseCalendarDate("2026-10-01")
			s := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
			if _, _, err := repo.Mutate(owner, &s, &cfg, "create", "c", "c", 0, now); err != nil {
				t.Fatal(err)
			}
			var err error
			if scope == "wallet" {
				err = NewWalletPostgresRepository(db).Delete(owner, id)
			} else {
				err = NewCategoryPostgresRepository(db).Delete(owner, id)
			}
			if err != nil {
				t.Fatal(err)
			}
			listed, err := repo.List(owner, now)
			if err != nil || len(listed) != 1 || listed[0].Active || listed[0].GenerationError == nil || *listed[0].GenerationError != "scope_deleted" {
				t.Fatalf("deleted %+v %v", listed, err)
			}
			if _, err = repo.Ensure(owner, now.AddDate(0, 2, 0)); err != nil {
				t.Fatal(err)
			}
			rows, err := NewBudgetPostgresRepository(db).List(owner)
			if err != nil || len(rows) != 0 {
				t.Fatalf("scope widened %+v %v", rows, err)
			}
		})
	}
}

func TestBudgetSeriesForeignScopesCannotBeSelected(t *testing.T) {
	db, owner, repo, now := budgetSeriesFixture(t)
	_, foreign, _, _ := budgetSeriesFixture(t)
	wid := uuid.NewString()
	if err := db.Create(&entity.Wallet{ID: wid, OwnerID: foreign, Name: "foreign", Type: entity.WalletTypeBasic, Currency: "VND"}).Error; err != nil {
		t.Fatal(err)
	}
	start, _ := entity.ParseCalendarDate("2026-10-01")
	s := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
	cfg := entity.BudgetSeriesConfig{Name: "bad", LimitAmount: 100, WalletID: &wid}
	if _, _, err := repo.Mutate(owner, &s, &cfg, "create", "c", "c", 0, now); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("foreign scope %v", err)
	}
}

func TestBudgetSeriesCatchUpBoundAndTemporalValidation(t *testing.T) {
	db, owner, repo, now := budgetSeriesFixture(t)
	past, _ := entity.ParseCalendarDate("2026-09-01")
	s := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: past}
	cfg := entity.BudgetSeriesConfig{Name: "bounded", LimitAmount: 100}
	if _, _, err := repo.Mutate(owner, &s, &cfg, "create", "past", "past", 0, now); !errors.Is(err, gorm.ErrInvalidData) {
		t.Fatalf("past start %v", err)
	}
	start, _ := entity.ParseCalendarDate("2026-10-12")
	s.Cadence = "weekly"
	s.StartDate = start
	if _, _, err := repo.Mutate(owner, &s, &cfg, "create", "weekly", "weekly", 0, now); err != nil {
		t.Fatal(err)
	}
	later := now.AddDate(0, 0, 707)
	count, err := repo.Ensure(owner, later)
	if err != nil || count != 100 {
		t.Fatalf("bounded %d %v", count, err)
	}
	count, err = repo.Ensure(owner, later)
	if err != nil || count != 1 {
		t.Fatalf("continued %d %v", count, err)
	}
	var rows int64
	if err = db.Model(&entity.Budget{}).Where("owner_id = ?", owner).Count(&rows).Error; err != nil || rows != 102 {
		t.Fatalf("total %d %v", rows, err)
	}
}

func TestBudgetSeriesConcurrentMutationReplaysOneReceipt(t *testing.T) {
	db, owner, repo, now := budgetSeriesFixture(t)
	start, _ := entity.ParseCalendarDate("2026-10-01")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
			cfg := entity.BudgetSeriesConfig{Name: "unique", LimitAmount: 100}
			if _, _, err := repo.Mutate(owner, &s, &cfg, "create", "shared", "shared-hash", 0, now); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	for _, model := range []any{&entity.BudgetSeries{}, &entity.Budget{}, &entity.TransactionMutation{}} {
		var count int64
		if err := db.Model(model).Where("owner_id = ?", owner).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("%T count %d %v", model, count, err)
		}
	}
}

func TestBudgetSeriesArchivedSelectionsAndPendingReferences(t *testing.T) {
	db, owner, repo, now := budgetSeriesFixture(t)
	cid := uuid.NewString()
	category := entity.Category{ID: cid, OwnerID: &owner, Name: "archive", Kind: entity.TransactionTypeExpense, IsActive: true}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	start, _ := entity.ParseCalendarDate("2026-10-01")
	s := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
	cfg := entity.BudgetSeriesConfig{Name: "unscoped", LimitAmount: 100}
	result, _, err := repo.Mutate(owner, &s, &cfg, "create", "c", "c", 0, now)
	if err != nil {
		t.Fatal(err)
	}
	cfg.CategoryID = &cid
	result, _, err = repo.Mutate(owner, &s, &cfg, "update", "pending", "pending", result.Version, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Model(&entity.Category{}).Where("id = ?", cid).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	cfg.LimitAmount = 200
	result, _, err = repo.Mutate(owner, &s, &cfg, "update", "keep", "keep", result.Version, now)
	if err != nil || result.PendingConfig == nil || result.PendingConfig.LimitAmount != 200 {
		t.Fatalf("keep archived pending %v %+v", err, result)
	}
	another := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
	if _, _, err = repo.Mutate(owner, &another, &cfg, "create", "new-archived", "new-archived", 0, now); !errors.Is(err, contract.ErrCategoryInactive) {
		t.Fatalf("new archived %v", err)
	}
	if _, err = repo.Ensure(owner, now.AddDate(0, 1, 0)); err != nil {
		t.Fatal(err)
	}
}

func TestBudgetSeriesScopeDeletionUsesOutstandingConfigApplicability(t *testing.T) {
	for _, scope := range []string{"wallet", "category"} {
		for _, scenario := range []string{"superseded", "outstanding", "future"} {
			t.Run(scope+"/"+scenario, func(t *testing.T) {
				db, owner, repo, now := budgetSeriesFixture(t)
				a, b := uuid.NewString(), uuid.NewString()
				for _, id := range []string{a, b} {
					var row any
					if scope == "wallet" {
						row = &entity.Wallet{ID: id, OwnerID: owner, Name: id, Type: entity.WalletTypeBasic, Currency: "VND"}
					} else {
						row = &entity.Category{ID: id, OwnerID: &owner, Name: id, Kind: entity.TransactionTypeExpense, IsActive: true}
					}
					if err := db.Create(row).Error; err != nil {
						t.Fatal(err)
					}
				}
				start, _ := entity.ParseCalendarDate("2026-10-01")
				end, _ := entity.ParseCalendarDate("2026-10-31")
				cfg := entity.BudgetSeriesConfig{Name: "A", LimitAmount: 100}
				if scope == "wallet" {
					cfg.WalletID = &a
				} else {
					cfg.CategoryID = &a
				}
				if scenario == "outstanding" {
					explicit := entity.Budget{ID: uuid.NewString(), Name: "catch-up blocker", LimitAmount: 100, StartDate: start, EndDate: end, WalletID: cfg.WalletID, CategoryID: cfg.CategoryID}
					if err := NewBudgetPostgresRepository(db).Save(owner, &explicit, true); err != nil {
						t.Fatal(err)
					}
				}
				s := entity.BudgetSeries{ID: uuid.NewString(), Cadence: "monthly", StartDate: start}
				created, _, err := repo.Mutate(owner, &s, &cfg, "create", "create", "create", 0, now)
				if err != nil {
					t.Fatal(err)
				}
				cfg.Name = "B"
				if scope == "wallet" {
					cfg.WalletID = &b
				} else {
					cfg.CategoryID = &b
				}
				edited, _, err := repo.Mutate(owner, &s, &cfg, "update", "edit", "edit", created.Version, now)
				if err != nil {
					t.Fatal(err)
				}
				if scenario == "superseded" {
					if _, err = repo.Ensure(owner, now.AddDate(0, 1, 0)); err != nil {
						t.Fatal(err)
					}
				}
				deleted := a
				if scenario == "future" {
					deleted = b
				}
				if scope == "wallet" {
					err = NewWalletPostgresRepository(db).Delete(owner, deleted)
				} else {
					err = NewCategoryPostgresRepository(db).Delete(owner, deleted)
				}
				if err != nil {
					t.Fatal(err)
				}
				var persisted entity.BudgetSeries
				if err = db.Where("id = ?", s.ID).First(&persisted).Error; err != nil {
					t.Fatal(err)
				}
				if scenario != "superseded" {
					if persisted.Active || persisted.GenerationError == nil || *persisted.GenerationError != "scope_deleted" {
						t.Fatalf("required scope did not stop: %+v", persisted)
					}
					return
				}
				if !persisted.Active || persisted.GenerationError != nil || persisted.Version != edited.Version {
					t.Fatalf("superseded scope stopped current series: %+v", persisted)
				}
				if _, err = repo.Ensure(owner, now.AddDate(0, 2, 0)); err != nil {
					t.Fatal(err)
				}
				var december entity.Budget
				if err = db.Where("series_id = ? AND start_date = ?", s.ID, "2026-12-01").First(&december).Error; err != nil {
					t.Fatal(err)
				}
				if (scope == "wallet" && (december.WalletID == nil || *december.WalletID != b)) || (scope == "category" && (december.CategoryID == nil || *december.CategoryID != b)) {
					t.Fatalf("December did not retain new scope: %+v", december)
				}
			})
		}
	}
}
