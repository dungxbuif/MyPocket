package repository

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestPortfolioDeleteCannotCascadeConcurrentTrade(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	if err = db.Create(&entity.User{ID: owner, Email: owner + "@portfolio.test", GoogleSubject: owner, Timezone: "UTC"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Where("id = ?", owner).Delete(&entity.User{}).Error; err != nil {
			t.Error(err)
		}
	})
	repo := NewPortfolioPostgresRepository(db)
	asset, err := repo.CreateAsset(owner, contract.PortfolioAssetInput{Symbol: "LOCK", Name: "Concurrent synthetic asset"})
	if err != nil {
		t.Fatal(err)
	}
	tradeTx := db.Begin()
	if tradeTx.Error != nil {
		t.Fatal(tradeTx.Error)
	}
	defer tradeTx.Rollback()
	var tradePID int
	if err = tradeTx.Raw("SELECT pg_backend_pid()").Scan(&tradePID).Error; err != nil {
		t.Fatal(err)
	}
	var locked entity.PortfolioAsset
	if err = tradeTx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "id = ?", asset.ID).Error; err != nil {
		t.Fatal(err)
	}
	deleted := make(chan error, 1)
	go func() { deleted <- repo.DeleteAsset(owner, asset.ID) }()
	// Observe actual database blocking, not an assumed scheduler delay. Before
	// the fix DELETE blocks after its stale count; after the fix SELECT blocks
	// before counting. Both paths resume after this trade commits.
	deadline := time.Now().Add(5 * time.Second)
	for {
		var waiting int64
		err = db.Raw("SELECT count(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock' AND ? = ANY(pg_blocking_pids(pid))", tradePID).Scan(&waiting).Error
		if err != nil {
			t.Fatal(err)
		}
		if waiting > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("delete never reached the held asset lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
	trade, err := NewPortfolioPostgresRepository(tradeTx).CreateTrade(owner, asset.ID, contract.PortfolioTradeInput{Side: entity.PortfolioTradeBuy, QuantityScaled: entity.QuantityScale, UnitPrice: 100, OccurredAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	if err = tradeTx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	select {
	case err = <-deleted:
		if !errors.Is(err, contract.ErrPortfolioAssetHasTrades) {
			t.Errorf("delete must retain a committed trade and return conflict, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("delete did not finish after trade commit")
	}
	var assets, trades int64
	if err = db.Model(&entity.PortfolioAsset{}).Where("id = ?", asset.ID).Count(&assets).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Model(&entity.PortfolioTrade{}).Where("id = ?", trade.ID).Count(&trades).Error; err != nil {
		t.Fatal(err)
	}
	if assets != 1 || trades != 1 {
		t.Fatalf("concurrent delete lost committed history: assets=%d trades=%d", assets, trades)
	}
}

func TestPortfolioUpdateAssetMetadataDoesNotRewriteTrades(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	if err = db.Create(&entity.User{ID: owner, Email: owner + "@portfolio.test", GoogleSubject: owner, Timezone: "UTC"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Where("id = ?", owner).Delete(&entity.User{}).Error })
	repo := NewPortfolioPostgresRepository(db)
	asset, err := repo.CreateAsset(owner, contract.PortfolioAssetInput{Symbol: "META", Name: "Before"})
	if err != nil {
		t.Fatal(err)
	}
	trade, err := repo.CreateTrade(owner, asset.ID, contract.PortfolioTradeInput{Side: entity.PortfolioTradeBuy, QuantityScaled: entity.QuantityScale, UnitPrice: 100, OccurredAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	name := "After"
	updated, err := repo.UpdateAsset(owner, asset.ID, contract.PortfolioAssetUpdate{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Symbol != "META" || updated.Name != "After" {
		t.Fatalf("updated asset=%q/%q, want META/After", updated.Symbol, updated.Name)
	}
	var persisted entity.PortfolioTrade
	if err = db.First(&persisted, "id = ?", trade.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.AssetID != asset.ID || persisted.UnitPrice != 100 {
		t.Fatalf("metadata update changed immutable trade: %+v", persisted)
	}
}

func TestPortfolioUpdateAssetRejectsDuplicateSymbolAndForeignOwner(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated local TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	foreign := uuid.NewString()
	for _, id := range []string{owner, foreign} {
		if err = db.Create(&entity.User{ID: id, Email: id + "@portfolio.test", GoogleSubject: id, Timezone: "UTC"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { _ = db.Where("id IN ?", []string{owner, foreign}).Delete(&entity.User{}).Error })
	repo := NewPortfolioPostgresRepository(db)
	first, err := repo.CreateAsset(owner, contract.PortfolioAssetInput{Symbol: "ONE", Name: "One"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.CreateAsset(owner, contract.PortfolioAssetInput{Symbol: "TWO", Name: "Two"})
	if err != nil {
		t.Fatal(err)
	}
	symbol := "ONE"
	if _, err = repo.UpdateAsset(owner, second.ID, contract.PortfolioAssetUpdate{Symbol: &symbol}); !errors.Is(err, contract.ErrPortfolioAssetConflict) {
		t.Fatalf("duplicate symbol error=%v, want conflict", err)
	}
	name := "No cross-owner write"
	if _, err = repo.UpdateAsset(foreign, first.ID, contract.PortfolioAssetUpdate{Name: &name}); !errors.Is(err, contract.ErrPortfolioAssetNotFound) {
		t.Fatalf("foreign owner error=%v, want not found", err)
	}
}

func TestPortfolioCorrectionVersionGuard(t *testing.T) {
	tests := []struct {
		name     string
		expected int64
		actual   int64
		wantErr  error
	}{
		{name: "missing version", expected: 0, actual: 1, wantErr: contract.ErrPortfolioCorrectionInvalid},
		{name: "stale version", expected: 1, actual: 2, wantErr: contract.ErrPortfolioAssetConflictVersion},
		{name: "current version", expected: 2, actual: 2, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePortfolioCorrectionVersion(tt.expected, tt.actual); !errors.Is(err, tt.wantErr) {
				t.Fatalf("validatePortfolioCorrectionVersion(%d, %d)=%v, want %v", tt.expected, tt.actual, err, tt.wantErr)
			}
		})
	}
}
