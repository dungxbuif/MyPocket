package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	marketrepo "github.com/mypocket/backend/internal/repository"
)

type catalogMarketRepositoryStub struct {
	products []entity.MarketProduct
}

func (s catalogMarketRepositoryStub) ListEnabledMarketProducts() ([]entity.MarketProduct, error) {
	return s.products, nil
}
func (catalogMarketRepositoryStub) FindMarketProduct(string) (*entity.MarketProduct, error) {
	return nil, marketrepo.ErrMarketProductNotFound
}
func (catalogMarketRepositoryStub) ClaimMarketPoll(string, time.Time, time.Time, time.Duration) (marketrepo.MarketPollClaim, error) {
	return marketrepo.MarketPollClaim{}, nil
}
func (catalogMarketRepositoryStub) CompleteMarketPoll(entity.MarketPollClaim, []entity.MarketQuote, time.Time) (int, error) {
	return 0, nil
}
func (catalogMarketRepositoryStub) FailMarketPoll(entity.MarketPollClaim, time.Time, error) error {
	return nil
}

func TestCatalogConfigExposesOnlyConfiguredGoldProductsAndSelectableUnits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	market := catalogMarketRepositoryStub{products: []entity.MarketProduct{
		{ID: "gold-1", ProviderKey: "btmc", ExternalProductID: "ring", Name: "Nhẫn", Kind: entity.PortfolioAssetKindGold, Unit: "chỉ", Currency: "VND", ValuationSide: "buyback", Enabled: true},
		{ID: "gold-2", ProviderKey: "btmh", ExternalProductID: "ring", Name: "Nhẫn khác", Kind: entity.PortfolioAssetKindGold, Unit: "lượng", Currency: "VND", ValuationSide: "buyback", Enabled: true},
		{ID: "stock-1", ProviderKey: "btmc", ExternalProductID: "vnm", Name: "VNM", Kind: "stock", Unit: "cổ phiếu", Currency: "VND", ValuationSide: "last", Enabled: true},
	}}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.GET("/catalog-config", (&AssetMarketHandler{Market: market, ConfiguredProviders: map[string]struct{}{"btmc": {}}}).CatalogConfig)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/catalog-config", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Data struct {
			AssetTypes []struct {
				Code  string `json:"code"`
				Units []struct {
					Code string `json:"code"`
				} `json:"units"`
			} `json:"asset_types"`
			Products []entity.MarketProduct `json:"products"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data.AssetTypes) != 1 || envelope.Data.AssetTypes[0].Code != "gold" {
		t.Fatalf("asset type catalog = %+v", envelope.Data.AssetTypes)
	}
	units := envelope.Data.AssetTypes[0].Units
	if len(units) != 3 || units[0].Code != "chỉ" || units[1].Code != "lượng" || units[2].Code != "phân" {
		t.Fatalf("gold units = %+v", units)
	}
	if len(envelope.Data.Products) != 1 || envelope.Data.Products[0].ID != "gold-1" {
		t.Fatalf("configured products = %+v", envelope.Data.Products)
	}
}

var _ marketrepo.MarketRepository = catalogMarketRepositoryStub{}
