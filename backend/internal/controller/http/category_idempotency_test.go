package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	categoryrepo "github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type categoryCreateRepositoryStub struct {
	categoryrepo.CategoryRepository
	created        []entity.Category
	mutationOwner  string
	mutationKey    string
	mutationHash   string
	mutationResult *entity.Category
}

func (s *categoryCreateRepositoryStub) ValidateWallets(string, []string) error { return nil }

func (s *categoryCreateRepositoryStub) Create(ownerID string, category *entity.Category) error {
	category.OwnerID = &ownerID
	s.created = append(s.created, *category)
	return nil
}

func (s *categoryCreateRepositoryStub) CreateIdempotent(ownerID string, category *entity.Category, key, hash string) (*entity.Category, bool, error) {
	if s.mutationResult != nil {
		if ownerID == s.mutationOwner && key == s.mutationKey && hash != s.mutationHash {
			return nil, false, categoryrepo.ErrMutationConflict
		}
		if ownerID == s.mutationOwner && key == s.mutationKey {
			return s.mutationResult, true, nil
		}
	}
	if err := s.Create(ownerID, category); err != nil {
		return nil, false, err
	}
	s.mutationOwner, s.mutationKey, s.mutationHash = ownerID, key, hash
	s.mutationResult = &s.created[len(s.created)-1]
	return s.mutationResult, false, nil
}

func TestCreateCategoryIdempotencyKeyReplaysPersistedResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &categoryCreateRepositoryStub{}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	router.POST("/categories", NewCategoryHandler(usecase.NewCategoryInteractor(repository)).CreateCategory)

	requestBody := `{"name":"Coffee","kind":"expense","icon_key":"tag"}`
	request := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "advisor-action-category-1")
	first := httptest.NewRecorder()
	router.ServeHTTP(first, request)
	if first.Code != http.StatusCreated {
		t.Fatalf("first status = %d, body = %s", first.Code, first.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "advisor-action-category-1")
	second := httptest.NewRecorder()
	router.ServeHTTP(second, request)
	if second.Code != http.StatusOK {
		t.Fatalf("replay status = %d, want 200, body = %s", second.Code, second.Body.String())
	}
	var firstPayload, secondPayload struct {
		Data entity.Category `json:"data"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &firstPayload); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(second.Body.Bytes(), &secondPayload); err != nil {
		t.Fatal(err)
	}
	if firstPayload.Data.ID == "" || secondPayload.Data.ID != firstPayload.Data.ID {
		t.Fatalf("replay category = %q, first category = %q", secondPayload.Data.ID, firstPayload.Data.ID)
	}
	if len(repository.created) != 1 {
		t.Fatalf("category writes = %d, want exactly one", len(repository.created))
	}
}

func TestAdvisorConfirmCategoryActionReplaysUsingActionID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &categoryCreateRepositoryStub{}
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "category_confirmation", Data: map[string]any{"view": map[string]any{
			"action_id": "category-action-1", "action": "create_category", "conversation_id": "conversation-1", "generation": int64(4),
			"name": "Ăn uống", "kind": "expense", "icon_key": "tag",
		}}}},
	}}}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{
		Service: &usecase.AdvisorService{Store: store}, Categories: repository,
	}).ConfirmAction)

	confirm := func() *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/actions/category-action-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
		request.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(recorder, request)
		return recorder
	}
	first := confirm()
	if first.Code != http.StatusCreated {
		t.Fatalf("first confirmation status=%d body=%s", first.Code, first.Body.String())
	}
	replay := confirm()
	if replay.Code != http.StatusOK {
		t.Fatalf("replayed confirmation status=%d body=%s", replay.Code, replay.Body.String())
	}
	if len(repository.created) != 1 {
		t.Fatalf("category writes=%d, want one", len(repository.created))
	}
}

func TestCreateCategoryIdempotencyConflictOwnerAndInvalidKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &categoryCreateRepositoryStub{}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, c.GetHeader("Synthetic-Owner")); c.Next() })
	router.POST("/categories", NewCategoryHandler(usecase.NewCategoryInteractor(repository)).CreateCategory)
	for _, tc := range []struct {
		owner, body string
		keys        []string
		want        int
	}{
		{"owner-1", `{"name":"Coffee","kind":"expense"}`, []string{"stable"}, 201},
		{"owner-1", `{"name":"Changed","kind":"expense"}`, []string{"stable"}, 409},
		{"owner-2", `{"name":"Coffee","kind":"expense"}`, []string{"stable"}, 201},
		{"owner-1", `{"name":"Invalid"}`, []string{""}, 400},
		{"owner-1", `{"name":"Invalid"}`, []string{"a", "b"}, 400},
		{"owner-1", `{"name":"Invalid"}`, []string{strings.Repeat("a", 129)}, 400},
		{"owner-1", `{"name":"Invalid"}`, []string{"non-ascii-\u00e9"}, 400},
	} {
		r := httptest.NewRequest("POST", "/categories", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Synthetic-Owner", tc.owner)
		for _, key := range tc.keys {
			r.Header.Add("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("owner=%s keys=%v status=%d want=%d body=%s", tc.owner, tc.keys, w.Code, tc.want, w.Body.String())
		}
	}
	if len(repository.created) != 2 {
		t.Fatalf("writes=%d want two isolated owners", len(repository.created))
	}
}
