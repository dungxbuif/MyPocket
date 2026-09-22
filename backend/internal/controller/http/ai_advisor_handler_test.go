package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type advisorOverviewUsers struct{}

func (advisorOverviewUsers) FindByEmail(string) (*entity.User, error) {
	return nil, repository.ErrNotFound
}
func (advisorOverviewUsers) FindByID(string) (*entity.User, error) {
	return &entity.User{ID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, nil
}
func (advisorOverviewUsers) FindOrCreateGoogleUser(entity.GoogleProfile) (*entity.User, error) {
	return nil, repository.ErrNotFound
}
func (advisorOverviewUsers) SetTimezone(string, string, bool) (*entity.User, error) {
	return nil, repository.ErrNotFound
}
func (advisorOverviewUsers) Create(*entity.User) error { return nil }

type advisorOverviewReader struct{}

func (advisorOverviewReader) ReadBundle(context.Context, string, []entity.NormalizedQuery) (entity.FactBundle, error) {
	return entity.FactBundle{Results: []entity.FinanceResult{{Status: "ok", ViewKind: "finance_summary", View: []byte(`{"income":1000,"expense":150,"net":850,"count":3}`), Source: entity.SourceScope{Timezone: "Asia/Ho_Chi_Minh"}}}}, nil
}

func TestAdvisorRoutesExposeReadOnlyConversationSurface(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := &Router{Engine: gin.New(), AuthMiddleware: &AuthMiddleware{}}
	router.RegisterAdvisorRoutes(&AdvisorHandler{})
	routes := map[string]bool{}
	for _, route := range router.Engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, expected := range []string{
		"GET /api/v1/ai/advisor/capabilities",
		"GET /api/v1/ai/advisor/overview",
		"POST /api/v1/ai/advisor/messages",
		"GET /api/v1/ai/advisor/conversation/messages",
		"GET /api/v1/ai/advisor/runs/:id",
		"POST /api/v1/ai/advisor/runs/:id/cancel",
	} {
		if !routes[expected] {
			t.Fatalf("missing route %s", expected)
		}
	}
	for route := range routes {
		if strings.Contains(route, "confirm") {
			t.Fatalf("V1 must not expose write confirmation: %s", route)
		}
	}
}

func TestAdvisorOverviewUsesFinanceReaderWithoutProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	service := usecase.NewFinanceQueryService(advisorOverviewReader{})
	engine.GET("/overview", (&AdvisorHandler{Users: advisorOverviewUsers{}, Finance: service}).Overview)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/overview?month=2026-09", nil)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"expense":150`) {
		t.Fatalf("unexpected overview response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
