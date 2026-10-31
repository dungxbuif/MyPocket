package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
)

func TestMarketCatalogSeparatesSameNameDifferentSources(t *testing.T) {
	gin.SetMode(gin.TestMode)
	market := catalogMarketRepositoryStub{products: []entity.MarketProduct{
		{ID: "btmc-ring", ProviderKey: "btmc", ExternalProductID: "ring-buy", Name: "Nhẫn tròn trơn", Kind: entity.PortfolioAssetKindGold, Unit: "chỉ", Currency: "VND", ValuationSide: "buyback", Enabled: true},
		{ID: "btmh-ring", ProviderKey: "btmh", ExternalProductID: "ring-buy", Name: "Nhẫn tròn trơn", Kind: entity.PortfolioAssetKindGold, Unit: "chỉ", Currency: "VND", ValuationSide: "buyback", Enabled: true},
	}}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.GET("/catalog", (&AssetMarketHandler{Market: market, ConfiguredProviders: map[string]struct{}{"btmc": {}, "btmh": {}}}).Catalog)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/catalog", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Data []entity.MarketProduct `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data) != 2 || envelope.Data[0].ID == envelope.Data[1].ID {
		t.Fatalf("catalog collapsed same-name sources: %+v", envelope.Data)
	}
}
