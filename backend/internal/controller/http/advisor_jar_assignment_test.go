package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

type jarAssignTransactions struct {
	advisorTransactionRepositoryStub
	rows    map[string]*entity.Transaction
	updates []string
}

func (s *jarAssignTransactions) List(owner string) ([]entity.Transaction, error) {
	var out []entity.Transaction
	for _, row := range s.rows {
		if row.OwnerID == owner {
			out = append(out, *row)
		}
	}
	return out, nil
}
func (s *jarAssignTransactions) Find(owner, id string) (*entity.Transaction, error) {
	if row := s.rows[id]; row != nil && row.OwnerID == owner {
		return row, nil
	}
	return nil, repository.ErrNotFound
}
func (s *jarAssignTransactions) Update(owner, id string, updates map[string]any) (*entity.Transaction, error) {
	row, err := s.Find(owner, id)
	if err != nil {
		return nil, err
	}
	s.updates = append(s.updates, id)
	row.JarID, _ = updates["jar_id"].(*string)
	row.Version++
	return row, nil
}

type jarContextJars struct{ transactionJarRepositoryStub }

func (jarContextJars) FindMonthConfig(owner, jarID, _ string, _ bool) (*entity.JarMonthConfig, error) {
	if owner == "owner-1" && jarID == "44444444-4444-4444-8444-444444444444" {
		return &entity.JarMonthConfig{JarID: jarID, Name: "Ăn uống", Active: true}, nil
	}
	return nil, repository.ErrNotFound
}

const (
	jaTx1 = "11111111-1111-4111-8111-111111111111"
	jaTx2 = "11111111-1111-4111-8111-222222222222"
	jaTx3 = "11111111-1111-4111-8111-333333333333"
	jaJar = "44444444-4444-4444-8444-444444444444"
	jaWal = "22222222-2222-4222-8222-222222222222"
)

func jarAssignFixture(view map[string]any) (*gin.Engine, *jarAssignTransactions, *advisorActionStore) {
	gin.SetMode(gin.TestMode)
	note := "Phở bò"
	rows := map[string]*entity.Transaction{}
	for _, id := range []string{jaTx1, jaTx2} {
		category := "55555555-5555-4555-8555-555555555555"
		rows[id] = &entity.Transaction{CategoryID: &category, ID: id, OwnerID: "owner-1", WalletID: jaWal, Type: entity.TransactionTypeExpense, Amount: 50000, OccurredAt: time.Date(2026, 9, 10, 3, 0, 0, 0, time.UTC), Note: &note, IncludedInReports: true, Version: 1}
	}
	rows[jaTx2].Version = 2 // changed after the card was prepared
	transactions := &jarAssignTransactions{rows: rows}
	store := &advisorActionStore{messages: []entity.AdvisorMessage{{
		ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
		Parts: []entity.AdvisorPart{{Type: "jar_assignment_confirmation", Data: map[string]any{"view": view}}},
	}}}
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner-1"); c.Next() })
	engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{
		Service: &usecase.AdvisorService{Store: store}, Transactions: transactions,
		Wallets:    &transactionWalletRepositoryStub{wallets: map[string]entity.Wallet{jaWal: {ID: jaWal, OwnerID: "owner-1", Type: entity.WalletTypeBasic}}},
		Categories: &transactionCategoryRepositoryStub{categories: []entity.Category{{ID: "55555555-5555-4555-8555-555555555555", Name: "Ăn", Kind: entity.TransactionTypeExpense, IsActive: true}}}, Users: &transactionUserRepositoryStub{user: entity.User{ID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}}, Jars: &jarContextJars{},
	}).ConfirmAction)
	return engine, transactions, store
}

func confirmJarAssignment(engine *gin.Engine) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions/assign-1/confirm", strings.NewReader(`{"conversation_id":"conversation-1","generation":4}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	return recorder
}

func TestAdvisorConfirmJarAssignmentReportsPerRowConflictAndAppliesOthers(t *testing.T) {
	view := map[string]any{"action_id": "assign-1", "action": "assign_jar", "conversation_id": "conversation-1", "generation": int64(4), "assignments": []any{
		map[string]any{"transaction_id": jaTx1, "jar_id": jaJar, "status": "proposed", "expected_version": int64(1)},
		map[string]any{"transaction_id": jaTx2, "jar_id": jaJar, "status": "proposed", "expected_version": int64(1)},
		map[string]any{"transaction_id": jaTx3, "jar_id": jaJar, "status": "rejected", "reject_reason": "Khoản thu không được gắn hũ."},
	}}
	engine, transactions, store := jarAssignFixture(view)
	recorder := confirmJarAssignment(engine)
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, `"applied_count":1`) || !strings.Contains(body, `"status":"conflict"`) || !strings.Contains(body, `"status":"skipped"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, body)
	}
	if len(transactions.updates) != 1 || transactions.updates[0] != jaTx1 || transactions.rows[jaTx2].JarID != nil {
		t.Fatalf("updates=%v", transactions.updates)
	}
	if len(store.marked) == 0 || store.marked[len(store.marked)-1] != entity.AIProposalApplied {
		t.Fatalf("marked=%v", store.marked)
	}
}

func TestAdvisorConfirmJarAssignmentIsIdempotentWhenAlreadyApplied(t *testing.T) {
	view := map[string]any{"action_id": "assign-1", "action": "assign_jar", "status": entity.AIProposalApplied, "conversation_id": "conversation-1", "generation": int64(4), "assignments": []any{
		map[string]any{"transaction_id": jaTx1, "jar_id": jaJar, "status": "proposed", "expected_version": int64(1)},
	}}
	engine, transactions, _ := jarAssignFixture(view)
	if recorder := confirmJarAssignment(engine); recorder.Code != http.StatusOK || len(transactions.updates) != 0 {
		t.Fatalf("status=%d updates=%v", recorder.Code, transactions.updates)
	}
}

func TestAdvisorConfirmJarAssignmentRowAlreadyInJarIsUnchanged(t *testing.T) {
	view := map[string]any{"action_id": "assign-1", "action": "assign_jar", "conversation_id": "conversation-1", "generation": int64(4), "assignments": []any{
		map[string]any{"transaction_id": jaTx1, "jar_id": "", "status": "proposed", "expected_version": int64(1)},
	}}
	engine, transactions, _ := jarAssignFixture(view)
	if recorder := confirmJarAssignment(engine); !strings.Contains(recorder.Body.String(), "unchanged") || len(transactions.updates) != 0 {
		t.Fatalf("body=%s updates=%v", recorder.Body.String(), transactions.updates)
	}
}

func TestAdvisorJarContextIsOwnerScopedReloadedAndHashed(t *testing.T) {
	_, transactions, _ := jarAssignFixture(map[string]any{})
	handler := &AdvisorHandler{Transactions: transactions, Categories: &transactionCategoryRepositoryStub{}, Users: &transactionUserRepositoryStub{user: entity.User{ID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}}, Jars: &jarContextJars{}}
	principal := usecase.Principal{OwnerID: "owner-1"}
	input := advisorMessageInput{Text: "gợi ý hũ", JarContext: &advisorJarContextInput{JarID: jaJar, Month: "2026-09"}}
	parts, hash, err := handler.messageParts(t.Context(), principal, input)
	if err != nil || len(parts) != 2 || parts[1].Type != "jar_context" {
		t.Fatalf("parts=%v err=%v", parts, err)
	}
	data := parts[1].Data.(map[string]any)
	if data["unassigned_expense_count"] != 2 {
		t.Fatalf("data=%v", data)
	}
	_, plain, _ := handler.messageParts(t.Context(), principal, advisorMessageInput{Text: "gợi ý hũ"})
	if plain == hash {
		t.Fatal("jar_context must change the payload hash")
	}
	if _, _, err := handler.messageParts(t.Context(), usecase.Principal{OwnerID: "owner-2"}, input); err == nil {
		t.Fatal("another owner's jar must not load")
	}
}
