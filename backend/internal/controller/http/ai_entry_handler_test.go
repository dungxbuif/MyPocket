package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/ai"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAIEntryHTTPReviewEditApproveRejectWithRealDatabase(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "review test", Type: "basic", Currency: "VND"}
	if err = db.Create(&entity.User{ID: owner, GoogleSubject: owner, Email: owner + "@test.invalid"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("id = ?", wallet.ID).Delete(&entity.Wallet{})
		db.Where("id = ?", wallet.OwnerID).Delete(&entity.User{})
	})
	if err = db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}
	providerCalls := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls++
		if r.URL.Path != "/chat/completions" {
			t.Errorf("wrong path %s", r.URL.Path)
		}
		draft := map[string]any{"type": "expense", "amount": 35000, "wallet_id": wallet.ID, "category_id": nil, "occurred_at": "2026-09-21T10:00:00+07:00", "note": "ăn sáng", "included_in_reports": true, "questions": []string{}}
		content, _ := json.Marshal(map[string]any{"reply": "Bạn kiểm tra các đề xuất trước khi duyệt nhé.", "drafts": []any{draft, draft}})
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": string(content)}, "finish_reason": "stop"}}})
	}))
	defer provider.Close()
	service := &usecase.AIEntryService{Entries: repo.NewAIEntryPostgresRepository(db), Wallets: repo.NewWalletPostgresRepository(db), Categories: repo.NewCategoryPostgresRepository(db), Extractor: ai.NewClient(ai.Config{BaseURL: provider.URL, APIKey: "test-key", Model: "test-model"})}
	h := &AIEntryHandler{Service: service}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, c.GetHeader("X-Test-Owner")); c.Next() })
	engine.POST("/sessions", h.CreateSession)
	engine.GET("/sessions/latest", h.LatestSession)
	engine.GET("/sessions/:id", h.Session)
	engine.POST("/sessions/:id/messages", h.Message)
	engine.PATCH("/proposals/:id", h.EditProposal)
	engine.POST("/proposals/:id/approve", h.ApproveProposal)
	engine.POST("/proposals/:id/reject", h.RejectProposal)
	request := func(method, path string, body any, want int, out any) {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-Owner", owner)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
		}
		if out != nil {
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(envelope.Data, out); err != nil {
				t.Fatal(err)
			}
		}
	}
	var s entity.AIEntrySession
	request("POST", "/sessions", map[string]any{}, 201, &s)
	input := map[string]any{"request_id": uuid.NewString(), "text": "ăn sáng 35k, ăn trưa 35k", "timezone": "Asia/Ho_Chi_Minh"}
	request("POST", "/sessions/"+s.ID+"/messages", input, 200, &s)
	if len(s.Proposals) != 2 || len(s.Messages) != 2 {
		t.Fatalf("missing review list: %+v", s)
	}
	var count int64
	db.Model(&entity.Transaction{}).Where("owner_id = ?", owner).Count(&count)
	if count != 0 {
		t.Fatal("extraction wrote ledger")
	}
	request("POST", "/sessions/"+s.ID+"/messages", input, 200, &s)
	if providerCalls != 1 || len(s.Proposals) != 2 {
		t.Fatal("message retry repeated extraction")
	}
	p := s.Proposals[0]
	p.Draft.Amount = 45000
	p.Draft.IncludedInReports = false
	request("PATCH", "/proposals/"+p.ID, map[string]any{"version": p.Version, "draft": p.Draft}, 200, &p)
	request("POST", "/proposals/"+p.ID+"/approve", map[string]any{"version": 1}, 409, nil)
	request("POST", "/proposals/"+p.ID+"/approve", map[string]any{"version": p.Version}, 200, &p)
	if p.Status != "approved" || p.TransactionID == nil {
		t.Fatal("approval not persisted")
	}
	request("POST", "/proposals/"+p.ID+"/approve", map[string]any{"version": 2}, 200, &p)
	request("POST", "/proposals/"+s.Proposals[1].ID+"/reject", map[string]any{"version": 1}, 200, nil)
	var rows []entity.Transaction
	db.Where("owner_id = ?", owner).Find(&rows)
	if len(rows) != 1 || rows[0].Amount != 45000 || rows[0].IncludedInReports {
		t.Fatalf("wrong approved ledger: %+v", rows)
	}
	request("GET", "/sessions/latest", nil, 200, &s)
	if s.Proposals[0].Status == "pending" || s.Proposals[1].Status == "pending" {
		t.Fatal("review state not restored")
	}
	owner = "other"
	request("GET", "/sessions/"+s.ID, nil, 404, nil)
	owner = wallet.OwnerID // Cleanup always uses the test fixture's owner.
}
