package httpapi

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"net/http/httptest"
	"testing"
	"time"
)

type budgetSeriesStub struct {
	calls     int
	owner     string
	operation string
	config    *entity.BudgetSeriesConfig
	version   int64
	err       error
	replay    bool
}

func (s *budgetSeriesStub) List(string, time.Time) ([]entity.BudgetSeries, error) {
	return []entity.BudgetSeries{}, s.err
}
func (s *budgetSeriesStub) Ensure(string, time.Time) (int, error) { return 0, s.err }
func (s *budgetSeriesStub) Mutate(owner string, series *entity.BudgetSeries, cfg *entity.BudgetSeriesConfig, op, key, hash string, version int64, now time.Time) (*entity.BudgetSeries, bool, error) {
	s.calls++
	s.owner = owner
	s.operation = op
	s.config = cfg
	s.version = version
	return series, s.replay, s.err
}

func TestBudgetSeriesInputAndOwnerGuard(t *testing.T) {
	for _, tc := range []struct {
		name, body, key string
		authorized      bool
		want            int
	}{
		{"valid", `{"name":"Food","limit_amount":100,"cadence":"monthly","start_date":"2026-10-01"}`, "key", true, 201},
		{"no key", `{"name":"Food","limit_amount":100,"cadence":"monthly","start_date":"2026-10-01"}`, "", true, 400},
		{"noncanonical", `{"name":"Food","limit_amount":100,"cadence":"monthly","start_date":"2026-10-02"}`, "key", true, 400},
		{"unauthorized", `{}`, "key", false, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &budgetSeriesStub{}
			h := &BudgetSeriesHandler{Series: repo, Users: &transactionUserRepositoryStub{user: entity.User{Timezone: "UTC"}}, Now: func() time.Time { return time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC) }}
			gin.SetMode(gin.TestMode)
			router := gin.New()
			if tc.authorized {
				router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner") })
			}
			router.POST("/series", h.Create)
			req := httptest.NewRequest("POST", "/series", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			if tc.key != "" {
				req.Header.Set("Idempotency-Key", tc.key)
			}
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.want {
				t.Fatalf("%d %s", res.Code, res.Body.String())
			}
			if tc.want == 201 && (repo.calls != 1 || repo.owner != "owner") {
				t.Fatal("missing owned write")
			}
		})
	}
}

func TestBudgetSeriesUpdateRequiresVersionAndImmutableCadence(t *testing.T) {
	for _, body := range []string{`{"name":"Food","limit_amount":100}`, `{"version":1,"name":"Food","limit_amount":100,"cadence":"yearly"}`} {
		repo := &budgetSeriesStub{}
		h := &BudgetSeriesHandler{Series: repo, Users: &transactionUserRepositoryStub{user: entity.User{Timezone: "UTC"}}}
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner") })
		router.PATCH("/series/:id", h.Update)
		req := httptest.NewRequest("PATCH", "/series/"+uuid.NewString(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "edit")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != 400 || repo.calls != 0 {
			t.Fatalf("%d %s", res.Code, res.Body.String())
		}
	}
}
