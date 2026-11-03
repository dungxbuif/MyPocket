package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

func TestNormalizeJarText(t *testing.T) {
	cases := map[string]string{"Đổ Xăng": "do-xang", "  Ăn  uống!! ": "an-uong", "Di chuyển": "di-chuyen", "ĐI CHỢ": "di-cho", "Quà_tặng 2": "qua-tang-2"}
	for in, want := range cases {
		if got := NormalizeJarText(in); got != want {
			t.Errorf("%q -> %q want %q", in, got, want)
		}
	}
	if got := normalizeJarNote("Đổ xăng 50k ngày 12/10 #di-chuyen"); got != "do-xang-ngay" {
		t.Errorf("note normalisation = %q", got)
	}
}

func suggesterFixture() (JarSuggester, string, string) {
	move, food := uuid.NewString(), uuid.NewString()
	s := JarSuggester{ActiveJars: map[string][]JarSuggestJar{"2026-10": {{move, "Di chuyển"}, {food, "Ăn uống"}}}}
	return s, move, food
}

func TestJarSuggesterRules(t *testing.T) {
	s, move, food := suggesterFixture()
	inactive := uuid.NewString() // not in the month's active jars
	s.History = []JarSuggestHistory{
		{"Đổ xăng 50k", move}, {"do xang 70000", move}, {"Đổ xăng", move},
		{"Cơm trưa", food}, {"Cơm trưa 40k", food},
		{"Cà phê", food}, {"Cà phê", move}, {"cà phê 30k", food},
		{"Gửi xe", inactive}, {"Gửi xe", inactive},
		{"Phở", food},
	}
	exp := func(id, note string) JarSuggestExpense { return JarSuggestExpense{ID: id, Note: note, Month: "2026-10"} }
	got := s.Suggest([]JarSuggestExpense{
		exp("hash", "mua đồ #An-Uống"), exp("high", "Đổ xăng 60k"), exp("medium", "Cơm trưa 45k"),
		exp("conflict", "Cà phê"), exp("inactive", "Gửi xe"), exp("single", "Phở"), exp("none", "Mua sách"), exp("empty", ""), exp("badtag", "#khong-co"),
	})
	by := map[string]JarSuggestion{}
	for _, item := range got.Suggested {
		by[item.TransactionID] = item
	}
	if by["hash"].JarID != food || by["hash"].Confidence != "high" {
		t.Errorf("hash: %+v", by["hash"])
	}
	if by["high"].JarID != move || by["high"].Confidence != "high" || by["high"].Reason != "3 khoản 'Đổ xăng 60k' trước đây gắn hũ Di chuyển" {
		t.Errorf("high: %+v", by["high"])
	}
	if by["medium"].JarID != food || by["medium"].Confidence != "medium" {
		t.Errorf("medium: %+v", by["medium"])
	}
	want := map[string]bool{"conflict": true, "inactive": true, "single": true, "none": true, "empty": true, "badtag": true}
	if len(got.NeedsClarification) != len(want) {
		t.Fatalf("needs=%v", got.NeedsClarification)
	}
	for _, id := range got.NeedsClarification {
		if !want[id] {
			t.Errorf("unexpected clarification %s", id)
		}
	}
}

func TestJarSuggesterHashtagRequiresActiveJarInMonth(t *testing.T) {
	s, _, _ := suggesterFixture()
	got := s.Suggest([]JarSuggestExpense{{ID: "x", Note: "#di-chuyen", Month: "2026-09"}})
	if len(got.Suggested) != 0 || len(got.NeedsClarification) != 1 {
		t.Fatalf("%+v", got)
	}
}

func TestJarSuggesterCapsAt50Rows(t *testing.T) {
	s, _, _ := suggesterFixture()
	rows := make([]JarSuggestExpense, 60)
	for i := range rows {
		rows[i] = JarSuggestExpense{ID: fmt.Sprint(i), Note: "x", Month: "2026-10"}
	}
	if got := s.Suggest(rows); len(got.NeedsClarification) != MaxJarAssignmentRows {
		t.Fatalf("rows=%d", len(got.NeedsClarification))
	}
}

// Another owner's history never reaches this owner's suggester: callers pass
// owner-scoped data only, so foreign notes in a different suggester have no effect.
func TestJarSuggesterIsolatedPerInstance(t *testing.T) {
	a, move, _ := suggesterFixture()
	a.History = []JarSuggestHistory{{"Đổ xăng", move}, {"Đổ xăng", move}}
	b, _, _ := suggesterFixture()
	if got := b.Suggest([]JarSuggestExpense{{ID: "1", Note: "Đổ xăng", Month: "2026-10"}}); len(got.Suggested) != 0 {
		t.Fatalf("leak: %+v", got)
	}
}

func TestAdvisorServiceBuildsJarCardWhenModelProducesNone(t *testing.T) {
	jar, tx1, tx2 := uuid.NewString(), uuid.NewString(), uuid.NewString()
	source := &jarAssignSourceStub{rows: map[string]*JarAssignmentTransaction{}, jars: map[string]string{jar + "|2026-10": "Di chuyển"}}
	for _, id := range []string{tx1, tx2} {
		source.rows["owner-1|"+id] = &JarAssignmentTransaction{ID: id, Amount: 1000, Note: "Đổ xăng", LocalMonth: "2026-10", Version: 2}
	}
	data := map[string]any{"month": "2026-10", "suggested_assignments": []any{map[string]any{"transaction_id": tx1, "jar_id": jar, "reason": "3 khoản 'Đổ xăng' trước đây gắn hũ Di chuyển", "confidence": "high"}}, "needs_clarification": []any{tx2}}
	for name, script := range map[string][]AdvisorResponse{"plain text": {{Text: "xin chào"}}, "loop limit": nil} {
		t.Run(name, func(t *testing.T) {
			store := &advisorStoreStub{run: &entity.AdvisorRun{ID: "run-1", OwnerID: "owner-1", ConversationID: "c", Generation: 1, Status: entity.AdvisorStatusQueued}, message: &entity.AdvisorMessage{ID: "m", Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}, {Type: "jar_context", Data: data}}}}
			registry := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))
			registry.SetJarAssignmentSource(source)
			service := NewAdvisorService(store, NewAdvisorOrchestrator(&advisorProviderStub{responses: script}, registry))
			result, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{OwnerID: "owner-1", RequestID: "r", PayloadHash: "h", Parts: store.message.Parts})
			if err != nil {
				t.Fatal(err)
			}
			if result.Text != "Mình đề xuất 1 khoản dựa trên ghi chú; 1 khoản cần bạn xác nhận thêm." || len(result.Results) != 1 || result.Results[0].ViewKind != "jar_assignment_proposal" {
				t.Fatalf("result=%+v", result)
			}
			found := false
			for _, part := range store.assistant.Parts {
				found = found || part.Type == "jar_assignment_confirmation"
			}
			if !found || store.status != entity.AdvisorStatusCompleted {
				t.Fatalf("parts=%+v status=%s", store.assistant.Parts, store.status)
			}
		})
	}
}

func TestAdvisorServiceWithoutJarContextKeepsFailure(t *testing.T) {
	store := &advisorStoreStub{run: &entity.AdvisorRun{ID: "run-1", OwnerID: "owner-1", ConversationID: "c", Generation: 1, Status: entity.AdvisorStatusQueued}, message: &entity.AdvisorMessage{ID: "m", Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}}}}
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))
	service := NewAdvisorService(store, NewAdvisorOrchestrator(&advisorProviderStub{}, registry))
	if _, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1"}, repository.AdvisorRunInput{OwnerID: "owner-1", RequestID: "r", PayloadHash: "h", Parts: store.message.Parts}); err == nil {
		t.Fatal("expected failure without jar_context")
	}
}
