package httpapi

import (
	"bytes"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPortfolioPriceRequiresExplicitFieldAndPreservesNullClearing(t *testing.T) {
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
	gin.SetMode(gin.TestMode)
	handler := &PortfolioHandler{Portfolio: postgresrepo.NewPortfolioPostgresRepository(db)}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, owner) })
	router.PATCH("/assets/:id/price", handler.UpdatePrice)
	for _, tc := range []struct {
		name, body string
		status     int
		want       *int64
	}{
		{"omitted", `{}`, 400, portfolioTestPrice(125)},
		{"null body", `null`, 400, portfolioTestPrice(125)},
		{"unrelated field", `{"note":"leave price"}`, 400, portfolioTestPrice(125)},
		{"explicit clear", `{"latest_price":null}`, 200, nil},
		{"known zero", `{"latest_price":0}`, 200, portfolioTestPrice(0)},
		{"updated price", `{"latest_price":250}`, 200, portfolioTestPrice(250)},
		{"negative", `{"latest_price":-1}`, 400, portfolioTestPrice(125)},
		{"string", `{"latest_price":"250"}`, 400, portfolioTestPrice(125)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			asset := entity.PortfolioAsset{ID: uuid.NewString(), OwnerID: owner, Symbol: uuid.NewString()[:8], Name: "Synthetic", LatestPrice: portfolioTestPrice(125)}
			if err := db.Create(&asset).Error; err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest("PATCH", "/assets/"+asset.ID+"/price", bytes.NewBufferString(tc.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			var persisted entity.PortfolioAsset
			if err := db.First(&persisted, "id = ?", asset.ID).Error; err != nil {
				t.Fatal(err)
			}
			if response.Code != tc.status {
				t.Errorf("status=%d want=%d body=%s", response.Code, tc.status, response.Body.String())
			}
			if (persisted.LatestPrice == nil) != (tc.want == nil) || (tc.want != nil && persisted.LatestPrice != nil && *persisted.LatestPrice != *tc.want) {
				t.Errorf("saved price=%v want=%v; omitted or invalid input must not clear valuation", persisted.LatestPrice, tc.want)
			}
		})
	}
}

func TestPortfolioAssetMetadataUpdateTrimsAndPreservesUnspecifiedField(t *testing.T) {
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
	asset := entity.PortfolioAsset{ID: uuid.NewString(), OwnerID: owner, Symbol: "VN30", Name: "Quỹ chỉ số"}
	if err = db.Create(&asset).Error; err != nil {
		t.Fatal(err)
	}
	handler := &PortfolioHandler{Portfolio: postgresrepo.NewPortfolioPostgresRepository(db)}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, owner) })
	router.PATCH("/assets/:id", handler.UpdateAsset)

	request := httptest.NewRequest("PATCH", "/assets/"+asset.ID, bytes.NewBufferString(`{"name":"  Quỹ cân bằng  "}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var persisted entity.PortfolioAsset
	if err = db.First(&persisted, "id = ?", asset.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.Symbol != "VN30" || persisted.Name != "Quỹ cân bằng" {
		t.Fatalf("metadata=%q/%q, want preserved symbol and trimmed name", persisted.Symbol, persisted.Name)
	}
}

func TestPortfolioAssetMetadataUpdateRejectsEmptyPatch(t *testing.T) {
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
	asset := entity.PortfolioAsset{ID: uuid.NewString(), OwnerID: owner, Symbol: "VN30", Name: "Quỹ chỉ số"}
	if err = db.Create(&asset).Error; err != nil {
		t.Fatal(err)
	}
	handler := &PortfolioHandler{Portfolio: postgresrepo.NewPortfolioPostgresRepository(db)}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, owner) })
	router.PATCH("/assets/:id", handler.UpdateAsset)
	request := httptest.NewRequest("PATCH", "/assets/"+asset.ID, bytes.NewBufferString(`{"name":"   "}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func portfolioTestPrice(value int64) *int64 { return &value }
