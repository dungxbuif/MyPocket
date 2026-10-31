package repository

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	marketrepo "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openAccumulatedAssetsTestDB(t *testing.T) (*gorm.DB, string, func(string)) {
	t.Helper()
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("requires disposable PostgreSQL TEST_DATABASE_URL")
	}
	if os.Getenv("APP_ENV") == "production" {
		t.Skip("refuses to run accumulated-assets integration tests in production")
	}
	database, err := gorm.Open(postgres.Open(os.Getenv("TEST_DATABASE_URL")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	ownerID := uuid.NewString()
	if err := database.Create(&entity.User{ID: ownerID, Email: ownerID + "@accumulated-assets.test", GoogleSubject: ownerID, Timezone: "UTC"}).Error; err != nil {
		t.Fatal(err)
	}
	productIDs := make([]string, 0, 4)
	t.Cleanup(func() {
		if err := database.Where("id = ?", ownerID).Delete(&entity.User{}).Error; err != nil {
			t.Error(err)
		}
		if len(productIDs) > 0 {
			if err := database.Where("id IN ?", productIDs).Delete(&entity.MarketProduct{}).Error; err != nil {
				t.Error(err)
			}
		}
		if sqlDB, err := database.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return database, ownerID, func(id string) { productIDs = append(productIDs, id) }
}

func createAccumulatedTestProduct(t *testing.T, database *gorm.DB, register func(string), provider, name, unit string) entity.MarketProduct {
	t.Helper()
	product := entity.MarketProduct{
		ID: uuid.NewString(), ProviderKey: provider, ExternalProductID: uuid.NewString(),
		Name: name, Kind: entity.PortfolioAssetKindGold, Unit: unit, Currency: "VND",
		ValuationSide: "buyback", Enabled: true,
	}
	if err := database.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	register(product.ID)
	return product
}

func createAccumulatedTestAsset(t *testing.T, database *gorm.DB, ownerID, symbol, unit string, confirmed bool, mode string, productID *string) entity.PortfolioAsset {
	t.Helper()
	asset := entity.PortfolioAsset{
		ID: uuid.NewString(), OwnerID: ownerID, Symbol: symbol, Name: symbol + " asset",
		UnitConfirmed: confirmed, PriceMode: mode, MarketProductID: productID, Version: 1,
	}
	if unit != "" {
		asset.UnitLabel = &unit
	}
	if err := database.Create(&asset).Error; err != nil {
		t.Fatal(err)
	}
	return asset
}

func TestConfirmAssetUnitRejectsForeignOwnerAndStaleVersion(t *testing.T) {
	database, ownerID, _ := openAccumulatedAssetsTestDB(t)
	portfolio := NewPortfolioPostgresRepository(database)
	asset, err := portfolio.CreateAsset(ownerID, marketrepo.PortfolioAssetInput{Symbol: "LEGACY", Name: "Legacy asset"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := portfolio.ConfirmAssetUnit(uuid.NewString(), asset.ID, 1, "chỉ"); !errors.Is(err, marketrepo.ErrPortfolioAssetNotFound) {
		t.Fatalf("foreign owner error=%v, want not found", err)
	}
	if _, err := portfolio.ConfirmAssetUnit(ownerID, asset.ID, 2, "chỉ"); !errors.Is(err, marketrepo.ErrPortfolioAssetConflictVersion) {
		t.Fatalf("stale version error=%v, want conflict", err)
	}
	confirmed, err := portfolio.ConfirmAssetUnit(ownerID, asset.ID, 1, "chỉ")
	if err != nil {
		t.Fatal(err)
	}
	if !confirmed.UnitConfirmed || confirmed.UnitLabel == nil || *confirmed.UnitLabel != "chỉ" || confirmed.Version != 2 {
		t.Fatalf("confirmed asset=%+v", confirmed)
	}
}

func TestCannotChangeConfirmedUnitWithTrades(t *testing.T) {
	database, ownerID, _ := openAccumulatedAssetsTestDB(t)
	portfolio := NewPortfolioPostgresRepository(database)
	unit := "chỉ"
	asset, err := portfolio.CreateAsset(ownerID, marketrepo.PortfolioAssetInput{Symbol: "GOLD", Name: "Gold", UnitLabel: &unit})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := portfolio.CreateTrade(ownerID, asset.ID, marketrepo.PortfolioTradeInput{Side: entity.PortfolioTradeBuy, QuantityScaled: entity.QuantityScale, UnitPrice: 8_000_000, OccurredAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if _, err := portfolio.ConfirmAssetUnit(ownerID, asset.ID, asset.Version, "lượng"); !errors.Is(err, marketrepo.ErrPortfolioAssetUnitConflict) {
		t.Fatalf("unit change error=%v, want trade-protected conflict", err)
	}
}

func TestCannotChangeAutomaticAssetUnitIntoCatalogMismatch(t *testing.T) {
	database, ownerID, register := openAccumulatedAssetsTestDB(t)
	product := createAccumulatedTestProduct(t, database, register, "btmc", "Gold", "chỉ")
	productID := product.ID
	asset := createAccumulatedTestAsset(t, database, ownerID, "AUTO", "chỉ", true, entity.PortfolioPriceModeAuto, &productID)
	portfolio := NewPortfolioPostgresRepository(database)
	_, err := portfolio.ConfirmAssetUnit(ownerID, asset.ID, asset.Version, "lượng")
	if !errors.Is(err, marketrepo.ErrMarketBindingUnitMismatch) {
		t.Fatalf("automatic unit change error=%v, want market binding mismatch", err)
	}
}

func TestMarketBindingRequiresConfirmedMatchingUnit(t *testing.T) {
	database, ownerID, register := openAccumulatedAssetsTestDB(t)
	product := createAccumulatedTestProduct(t, database, register, "btmc", "Gold", "chỉ")
	productID := product.ID
	asset := createAccumulatedTestAsset(t, database, ownerID, "UNCONFIRMED", "chỉ", false, entity.PortfolioPriceModeManual, nil)
	portfolio := NewPortfolioPostgresRepository(database)
	if _, err := portfolio.UpdateAssetPriceSource(ownerID, asset.ID, marketrepo.PortfolioPriceSourceInput{ExpectedVersion: 1, PriceMode: entity.PortfolioPriceModeAuto, MarketProductID: &productID}); !errors.Is(err, marketrepo.ErrMarketBindingInvalid) {
		t.Fatalf("unconfirmed binding error=%v, want invalid binding", err)
	}
	confirmed, err := portfolio.ConfirmAssetUnit(ownerID, asset.ID, 1, "lượng")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := portfolio.UpdateAssetPriceSource(ownerID, asset.ID, marketrepo.PortfolioPriceSourceInput{ExpectedVersion: confirmed.Version, PriceMode: entity.PortfolioPriceModeAuto, MarketProductID: &productID}); !errors.Is(err, marketrepo.ErrMarketBindingUnitMismatch) {
		t.Fatalf("mismatched binding error=%v, want unit mismatch", err)
	}
}

func TestLegacyPriceEndpointCannotOverrideAutomaticAsset(t *testing.T) {
	database, ownerID, register := openAccumulatedAssetsTestDB(t)
	product := createAccumulatedTestProduct(t, database, register, "btmc", "Gold", "chỉ")
	productID := product.ID
	asset := createAccumulatedTestAsset(t, database, ownerID, "AUTO", "chỉ", true, entity.PortfolioPriceModeAuto, &productID)
	price := int64(8_000_000)
	_, err := NewPortfolioPostgresRepository(database).UpdateAssetPrice(ownerID, asset.ID, marketrepo.PortfolioPriceUpdate{LatestPrice: &price, LatestPriceAt: timePtrForTest(time.Now().UTC())})
	if !errors.Is(err, marketrepo.ErrPortfolioAssetAutomaticPrice) {
		t.Fatalf("automatic price update error=%v, want protected error", err)
	}
}

func TestMarketSummaryUnknownAndStaleValuation(t *testing.T) {
	database, ownerID, register := openAccumulatedAssetsTestDB(t)
	unknownProduct := createAccumulatedTestProduct(t, database, register, "btmc", "Unknown", "chỉ")
	staleProduct := createAccumulatedTestProduct(t, database, register, "btmh", "Stale", "chỉ")
	unknownID, staleID := unknownProduct.ID, staleProduct.ID
	unknown := createAccumulatedTestAsset(t, database, ownerID, "UNKNOWN", "chỉ", true, entity.PortfolioPriceModeAuto, &unknownID)
	stale := createAccumulatedTestAsset(t, database, ownerID, "STALE", "chỉ", true, entity.PortfolioPriceModeAuto, &staleID)
	now := time.Now().UTC()
	if err := database.Create(&entity.MarketQuote{ProductID: staleProduct.ID, Price: 10_000_000, ObservedAt: now.Add(-2 * time.Hour), FetchedAt: now.Add(-2 * time.Hour)}).Error; err != nil {
		t.Fatal(err)
	}
	portfolio := NewPortfolioPostgresRepository(database)
	for _, asset := range []entity.PortfolioAsset{unknown, stale} {
		if _, err := portfolio.CreateTrade(ownerID, asset.ID, marketrepo.PortfolioTradeInput{Side: entity.PortfolioTradeBuy, QuantityScaled: entity.QuantityScale, UnitPrice: 1_000_000, OccurredAt: now.Add(-3 * time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	summary, err := portfolio.Summary(ownerID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalMarketValue != nil || summary.TotalUnrealizedPnL != nil || summary.UnknownAssetCount != 1 || summary.StaleAssetCount != 1 {
		t.Fatalf("summary completeness=%+v, want unknown total and one stale asset", summary)
	}
	if summary.KnownMarketValue == nil || *summary.KnownMarketValue != 10_000_000 {
		t.Fatalf("known market value=%v, want 10000000", summary.KnownMarketValue)
	}
}

func TestMarketPollConcurrentWorkersOnlyOneClaimWins(t *testing.T) {
	database, _, _ := openAccumulatedAssetsTestDB(t)
	market := NewAssetMarketPostgresRepository(database)
	now := time.Now().UTC().Truncate(time.Second)
	providerKey := "btmc-" + uuid.NewString()
	var wg sync.WaitGroup
	results := make(chan marketrepo.MarketPollClaim, 2)
	errorsCh := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			claim, err := market.ClaimMarketPoll(providerKey, now, now, time.Minute)
			results <- claim
			errorsCh <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errorsCh)
	wins, sawErrors := 0, 0
	for err := range errorsCh {
		if err != nil {
			sawErrors++
		}
	}
	for claim := range results {
		if claim.Acquired {
			wins++
		}
	}
	if sawErrors != 0 || wins != 1 {
		t.Fatalf("concurrent claims errors=%d wins=%d, want one winner", sawErrors, wins)
	}
}

func TestMarketPollOutOfOrderQuoteKeepsNewest(t *testing.T) {
	database, _, register := openAccumulatedAssetsTestDB(t)
	providerKey := "btmc-" + uuid.NewString()
	product := createAccumulatedTestProduct(t, database, register, providerKey, "Gold", "chỉ")
	market := NewAssetMarketPostgresRepository(database)
	firstNow := time.Now().UTC().Truncate(time.Second)
	firstClaim, err := market.ClaimMarketPoll(providerKey, firstNow, firstNow, time.Minute)
	if err != nil || !firstClaim.Acquired {
		t.Fatalf("first claim=%+v err=%v", firstClaim, err)
	}
	newer := firstNow.Add(-10 * time.Minute)
	if _, err := market.CompleteMarketPoll(firstClaim.Claim, []entity.MarketQuote{{ProductID: product.ID, Price: 10_000_000, ObservedAt: newer, FetchedAt: firstNow}}, firstNow); err != nil {
		t.Fatal(err)
	}
	secondNow := firstNow.Add(10 * time.Minute)
	secondClaim, err := market.ClaimMarketPoll(providerKey, secondNow, secondNow, time.Minute)
	if err != nil || !secondClaim.Acquired {
		t.Fatalf("second claim=%+v err=%v", secondClaim, err)
	}
	if accepted, err := market.CompleteMarketPoll(secondClaim.Claim, []entity.MarketQuote{{ProductID: product.ID, Price: 9_000_000, ObservedAt: newer.Add(-time.Minute), FetchedAt: secondNow}}, secondNow); err != nil || accepted != 0 {
		t.Fatalf("older quote accepted=%d err=%v, want no rollback", accepted, err)
	}
	var quote entity.MarketQuote
	if err := database.First(&quote, "product_id = ?", product.ID).Error; err != nil {
		t.Fatal(err)
	}
	if quote.Price != 10_000_000 || !quote.ObservedAt.Equal(newer) {
		t.Fatalf("stored quote=%+v, want newest quote retained", quote)
	}
}

func TestMarketPollFailureKeepsLastPriceAndAssetVersion(t *testing.T) {
	database, ownerID, register := openAccumulatedAssetsTestDB(t)
	providerKey := "btmc-" + uuid.NewString()
	product := createAccumulatedTestProduct(t, database, register, providerKey, "Gold", "chỉ")
	productID := product.ID
	asset := createAccumulatedTestAsset(t, database, ownerID, "AUTO", "chỉ", true, entity.PortfolioPriceModeAuto, &productID)
	now := time.Now().UTC().Truncate(time.Second)
	if err := database.Create(&entity.MarketQuote{ProductID: product.ID, Price: 10_000_000, ObservedAt: now.Add(-time.Minute), FetchedAt: now.Add(-time.Minute)}).Error; err != nil {
		t.Fatal(err)
	}
	market := NewAssetMarketPostgresRepository(database)
	claim, err := market.ClaimMarketPoll(providerKey, now, now, time.Minute)
	if err != nil || !claim.Acquired {
		t.Fatalf("claim=%+v err=%v", claim, err)
	}
	if err := market.FailMarketPoll(claim.Claim, now, errors.New("provider failure")); err != nil {
		t.Fatal(err)
	}
	var quote entity.MarketQuote
	if err := database.First(&quote, "product_id = ?", product.ID).Error; err != nil {
		t.Fatal(err)
	}
	var persisted entity.PortfolioAsset
	if err := database.First(&persisted, "id = ?", asset.ID).Error; err != nil {
		t.Fatal(err)
	}
	if quote.Price != 10_000_000 || persisted.Version != asset.Version {
		t.Fatalf("failure changed quote/version: quote=%+v assetVersion=%d want %d", quote, persisted.Version, asset.Version)
	}
}

func timePtrForTest(value time.Time) *time.Time { return &value }
