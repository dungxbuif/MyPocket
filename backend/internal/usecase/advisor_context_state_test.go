package usecase

import (
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

func TestAdvisorContextBoundsUTF8AndCurrentMessage(t *testing.T) {
	old := []entity.AdvisorMessage{{Seq: 1, Role: entity.AdvisorRoleUser, Parts: []entity.AdvisorPart{{Type: "text", Text: "cũ " + string(make([]byte, 24000))}}}}
	current := entity.AdvisorMessage{Seq: 2, Role: entity.AdvisorRoleUser, Parts: []entity.AdvisorPart{{Type: "text", Text: "xin chào 👋"}}}
	chat, err := BuildAdvisorRequestContext(append(old, current), entity.AdvisorContextState{Generation: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(chat) == 0 || chat[len(chat)-1].Content != "xin chào 👋" {
		t.Fatalf("current message must remain whole: %+v", chat)
	}
	var bytes int
	for _, item := range chat {
		bytes += len([]byte(item.Content))
	}
	if bytes > advisorContextTextBudget {
		t.Fatalf("context exceeded byte budget: %d", bytes)
	}
}

func TestAdvisorContextInheritsExplicitFiltersOnly(t *testing.T) {
	state, err := BuildAdvisorContextState(entity.AdvisorContextState{}, 4, 8, []entity.FinanceResult{{QueryKey: "call-1", ViewKind: "transaction_search", Source: entity.SourceScope{Timezone: "Asia/Ho_Chi_Minh", Filter: entity.FinanceFilter{WalletIDs: []string{"w1"}, CategoryIDs: []string{"c1"}, Range: entity.DateRange{From: "2026-09-01", To: "2026-09-30"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if state.Generation != 4 || state.ThroughSeq != 8 || state.Timezone != "Asia/Ho_Chi_Minh" || len(state.WalletRefs) != 1 || state.WalletRefs[0] != "w1" {
		t.Fatalf("unexpected state: %+v", state)
	}
	if state.LastAmount != 0 || state.LastVersion != 0 {
		t.Fatalf("context state must not retain amount/version: %+v", state)
	}
}

func TestAdvisorContextResetPurgesReferences(t *testing.T) {
	state := entity.AdvisorContextState{Generation: 2, ThroughSeq: 9, WalletRefs: []string{"w1"}, LastQueryKind: "get_wallet_balances"}
	reset, err := BuildAdvisorContextState(state, 3, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if reset.Generation != 3 || len(reset.WalletRefs) != 0 || reset.LastQueryKind != "" {
		t.Fatalf("stale references survived generation reset: %+v", reset)
	}
}

func TestAdvisorContextStateHasBoundedExpiryMetadata(t *testing.T) {
	state, err := BuildAdvisorContextState(entity.AdvisorContextState{}, 1, 1, []entity.FinanceResult{{QueryKey: "q", Source: entity.SourceScope{Timezone: "UTC", Filter: entity.FinanceFilter{Range: entity.DateRange{From: "2026-01-01", To: "2026-12-31"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if state.UpdatedAt.IsZero() || state.UpdatedAt.After(time.Now().UTC().Add(time.Second)) {
		t.Fatalf("invalid state timestamp: %v", state.UpdatedAt)
	}
}
