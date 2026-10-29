package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type reviewProviderFunc func(context.Context, AdvisorRequest) (AdvisorResponse, error)

type reviewContextStore struct {
	*advisorRuntimeStub
	state entity.AdvisorContextState
}

func (s *reviewContextStore) LoadAdvisorContextState(_ context.Context, owner, conversation string, generation int64) (entity.AdvisorContextState, error) {
	if owner != "owner-1" || conversation != "conversation-1" || generation != 1 {
		return entity.AdvisorContextState{}, repository.ErrNotFound
	}
	return s.state, nil
}

type reviewContextReader struct{ t *testing.T }

func (r *reviewContextReader) ReadBundle(_ context.Context, owner string, queries []entity.NormalizedQuery) (entity.FactBundle, error) {
	if owner != "owner-1" {
		r.t.Fatalf("reference validation escaped owner: %s", owner)
	}
	q := queries[0]
	result := entity.FinanceResult{QueryKey: q.Key, Status: "not_found"}
	if q.Kind == "get_wallet_balances" {
		result.Status = "ok"
		result.ViewKind = "wallet_balances"
		result.View = json.RawMessage(`{"items":[{"id":"11111111-1111-4111-8111-111111111111","balance":999999}]}`)
	}
	return entity.FactBundle{Results: []entity.FinanceResult{result}}, nil
}

func TestAdvisorReviewServiceReusesOnlyRevalidatedContinuation(t *testing.T) {
	store := &reviewContextStore{advisorRuntimeStub: &advisorRuntimeStub{advisorStoreStub: &advisorStoreStub{
		run:     &entity.AdvisorRun{ID: "run-context", OwnerID: "owner-1", ConversationID: "conversation-1", Generation: 1, Status: entity.AdvisorStatusQueued},
		message: &entity.AdvisorMessage{Seq: 9, Role: "user", Parts: []entity.AdvisorPart{{Type: "text", Text: "So với tháng trước?"}}},
	}}, state: entity.AdvisorContextState{Version: 1, Generation: 1, DateFrom: "2026-09-01", DateTo: "2026-09-30", Timezone: "Asia/Ho_Chi_Minh", WalletRefs: []string{"11111111-1111-4111-8111-111111111111"}, CategoryRefs: []string{"22222222-2222-4222-8222-222222222222"}}}
	p := &capturingAdvisorProvider{response: AdvisorResponse{Text: "answer"}}
	service := NewAdvisorServiceWithRuntime(store, NewAdvisorOrchestrator(p, NewAdvisorToolRegistry(NewFinanceQueryService(&reviewContextReader{t}))), true)
	if _, err := service.Submit(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, repository.AdvisorRunInput{RequestID: "ctx", Parts: store.message.Parts}); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(p.request.Messages)
	if !strings.Contains(string(payload), "2026-09-01") || !strings.Contains(string(payload), "11111111-1111-4111-8111-111111111111") {
		t.Fatalf("continuation state never reached provider: %s", payload)
	}
	if strings.Contains(string(payload), "22222222-2222-4222-8222-222222222222") || strings.Contains(string(payload), "999999") {
		t.Fatalf("unvalidated refs or financial snapshots leaked into memory: %s", payload)
	}
}

func TestAdvisorReviewContextUsesAccountCalendar(t *testing.T) {
	start := time.Date(2026, 8, 31, 17, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 30, 17, 0, 0, 0, time.UTC)
	state, err := BuildAdvisorContextState(entity.AdvisorContextState{}, 1, 2, []entity.FinanceResult{{Status: "ok", ViewKind: "finance_summary", Source: entity.SourceScope{Timezone: "Asia/Ho_Chi_Minh", StartAt: &start, EndExclusive: &end}}})
	if err != nil || state.DateFrom != "2026-09-01" || state.DateTo != "2026-09-30" {
		t.Fatalf("wrong account-local month: %+v %v", state, err)
	}
}

func TestAdvisorReviewBroadQueryClearsPriorWalletScope(t *testing.T) {
	previous := entity.AdvisorContextState{Version: 1, Generation: 1, WalletRefs: []string{"old-wallet"}, CategoryRefs: []string{"old-category"}}
	state, err := BuildAdvisorContextState(previous, 1, 3, []entity.FinanceResult{{Status: "ok", ViewKind: "finance_summary", Source: entity.SourceScope{Timezone: "Asia/Ho_Chi_Minh", Filter: entity.FinanceFilter{Range: entity.DateRange{From: "2026-10-01", To: "2026-10-31"}}}}})
	if err != nil || len(state.WalletRefs) != 0 || len(state.CategoryRefs) != 0 {
		t.Fatalf("all-wallet query retained old selected scope: %+v %v", state, err)
	}
}

func TestAdvisorReviewGroundedHelpUsesCatalogSource(t *testing.T) {
	provider := &advisorProviderStub{responses: []AdvisorResponse{
		{ToolCalls: []AdvisorToolCall{{ID: "help", Name: "get_product_help", Arguments: json.RawMessage(`{"query":"ngân sách","max_results":1}`)}}},
		{ToolCalls: []AdvisorToolCall{{ID: "done", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"product_help","source_id":"handbook.budgets.v1.3","explanation":"Bạn có chín tỷ đồng"}`)}}},
	}}
	o := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	o.GroundingEnabled = true
	answer, err := o.Answer(context.Background(), Principal{OwnerID: "owner-1"}, []AdvisorChatMessage{{Role: "user", Content: "Tạo ngân sách ở đâu?"}})
	if err != nil || !strings.Contains(answer.Text, "Tài khoản > Ngân sách") || strings.Contains(answer.Text, "chín tỷ") {
		t.Fatalf("help did not render verified catalog: %q %v", answer.Text, err)
	}
}

func TestAdvisorReviewClarificationAsksOneSelectedQuestion(t *testing.T) {
	var final entity.AdvisorFinalAnswer
	if err := json.Unmarshal([]byte(`{"mode":"clarification","question_id":"wallet","explanation":"Bạn có chín tỷ đồng"}`), &final); err != nil {
		t.Fatal(err)
	}
	text, err := RenderAdvisorFinal(final, NewAdvisorEvidenceIndex())
	if err != nil || text != "Bạn muốn chọn ví nào?" {
		t.Fatalf("clarification must ask selected server-owned question: %q %v", text, err)
	}
}

func TestAdvisorReviewGroundedProposalFinishesWithoutInventedFacts(t *testing.T) {
	p := &advisorProviderStub{responses: []AdvisorResponse{
		{ToolCalls: []AdvisorToolCall{{ID: "p", Name: "propose_create_wallet", Arguments: json.RawMessage(`{"name":"Ví mới","type":"basic"}`)}}},
		{ToolCalls: []AdvisorToolCall{{ID: "done", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"finance","template_id":"proposal"}`)}}},
	}}
	o := NewAdvisorOrchestrator(p, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	o.GroundingEnabled = true
	answer, err := o.Answer(context.Background(), Principal{OwnerID: "owner-1"}, []AdvisorChatMessage{{Role: "user", Content: "Tạo ví mới"}})
	if err != nil || len(answer.Results) != 1 || answer.Results[0].Status != "proposed" || !strings.Contains(answer.Text, "xác nhận") {
		t.Fatalf("proposal cannot terminate: %+v %v", answer, err)
	}
}

func (f reviewProviderFunc) Chat(ctx context.Context, req AdvisorRequest) (AdvisorResponse, error) {
	return f(ctx, req)
}

func TestAdvisorReviewContextCapsShortMessages(t *testing.T) {
	var history []entity.AdvisorMessage
	for i := 1; i <= 40; i++ {
		history = append(history, entity.AdvisorMessage{Seq: int64(i), Role: "user", Parts: []entity.AdvisorPart{{Type: "text", Text: fmt.Sprintf("message-%d", i)}}})
	}
	got, err := BuildAdvisorRequestContext(history, entity.AdvisorContextState{})
	if err != nil || len(got) != 12 || got[0].Content != "message-29" || got[11].Content != "message-40" {
		t.Fatalf("expected latest 12 complete messages, got %d: %v", len(got), err)
	}
}

func TestAdvisorReviewGroundingCannotPublishModelNarrative(t *testing.T) {
	idx := NewAdvisorEvidenceIndex()
	if err := idx.Add("lookup", entity.FinanceResult{Status: "ok", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: int64Ptr(150)}}}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"product_help", "clarification"} {
		text, err := RenderAdvisorFinal(entity.AdvisorFinalAnswer{Mode: mode, Explanation: "Số dư của bạn là chín tỷ đồng"}, idx)
		if err == nil && strings.Contains(text, "chín tỷ") {
			t.Fatalf("%s bypassed grounding: %s", mode, text)
		}
	}
	text, err := RenderAdvisorFinal(entity.AdvisorFinalAnswer{Mode: "finance", TemplateID: "summary", Claims: []entity.AdvisorClaim{{Slot: "Thu nhập chín tỷ", FactRef: "lookup.q1.expense"}}}, idx)
	if err == nil && (strings.Contains(text, "Thu nhập") || strings.Contains(text, "chín tỷ")) {
		t.Fatalf("model relabeled expense fact: %s", text)
	}
}

func TestAdvisorReviewFinishRemainsAvailableAfterEightReads(t *testing.T) {
	result := entity.FinanceResult{Status: "ok", QueryKey: "q1", ViewKind: "finance_summary", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: int64Ptr(150)}}}
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{result}}}
	calls := 0
	provider := reviewProviderFunc(func(_ context.Context, req AdvisorRequest) (AdvisorResponse, error) {
		calls++
		if calls == 1 {
			var batch []AdvisorToolCall
			for i := 0; i < 8; i++ {
				batch = append(batch, AdvisorToolCall{ID: fmt.Sprintf("r%d", i), Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)})
			}
			return AdvisorResponse{ToolCalls: batch}, nil
		}
		found := false
		for _, tool := range req.Tools {
			if tool.Name == "finish_answer" {
				found = true
			} else {
				t.Fatalf("finance tool still advertised after budget: %s", tool.Name)
			}
		}
		if !found {
			t.Fatal("finish_answer removed after eight reads")
		}
		return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "done", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"finance","template_id":"summary","claims":[{"slot":"expense","fact_ref":"r7.q1.expense"}]}`)}}}, nil
	})
	o := NewAdvisorOrchestrator(provider, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	o.GroundingEnabled = true
	if _, err := o.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}}); err != nil {
		t.Fatal(err)
	}
}

func TestAdvisorReviewRepairPairsTerminalToolResult(t *testing.T) {
	calls := 0
	p := reviewProviderFunc(func(_ context.Context, req AdvisorRequest) (AdvisorResponse, error) {
		calls++
		if calls == 1 {
			return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "bad", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"invalid"}`)}}}, nil
		}
		for i, m := range req.Messages {
			if len(m.ToolCalls) > 0 {
				if i+1 >= len(req.Messages) || req.Messages[i+1].Role != "tool" || req.Messages[i+1].ToolCallID != "bad" {
					t.Fatal("repair transcript contains unmatched assistant tool call")
				}
			}
		}
		return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "fixed", Name: "finish_answer", Arguments: json.RawMessage(`{"mode":"clarification","explanation":"Bạn muốn xem ví nào?"}`)}}}, nil
	})
	o := NewAdvisorOrchestrator(p, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	o.GroundingEnabled = true
	if _, err := o.Answer(context.Background(), Principal{OwnerID: "owner-1"}, []AdvisorChatMessage{{Role: "user", Content: "hello"}}); err != nil {
		t.Fatal(err)
	}
}

func TestAdvisorReviewProposalRequiresSnapshotInRealLoop(t *testing.T) {
	calls := 0
	p := reviewProviderFunc(func(_ context.Context, req AdvisorRequest) (AdvisorResponse, error) {
		calls++
		if calls == 1 {
			return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "proposal", Name: "propose_delete_transaction", Arguments: json.RawMessage(`{"transaction_id":"11111111-1111-4111-8111-111111111111","expected_version":1}`)}}}, nil
		}
		last := req.Messages[len(req.Messages)-1]
		if last.Role != "tool" || !strings.Contains(last.Content, "snapshot_required") {
			t.Fatalf("missing repairable snapshot error: %+v", last)
		}
		return AdvisorResponse{Text: "Cần đọc giao dịch trước."}, nil
	})
	o := NewAdvisorOrchestrator(p, NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})))
	answer, err := o.Answer(context.Background(), Principal{OwnerID: "owner-1"}, []AdvisorChatMessage{{Role: "user", Content: "delete"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(answer.Results) != 0 {
		t.Fatalf("unread proposal escaped into cards: %+v", answer.Results)
	}
}

func TestAdvisorReviewTranscriptStopsOversizedResult(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{Status: "ok", QueryKey: "q1", ViewKind: "transaction_search", View: json.RawMessage(`{"items":[],"text":"` + strings.Repeat("x", 100*1024) + `"}`)}}}}
	calls := 0
	p := reviewProviderFunc(func(_ context.Context, req AdvisorRequest) (AdvisorResponse, error) {
		calls++
		if calls > 1 {
			t.Fatal("oversized result was sent to provider")
		}
		return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "large", Name: "search_transactions", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}}}, nil
	})
	o := NewAdvisorOrchestrator(p, NewAdvisorToolRegistry(NewFinanceQueryService(reader)))
	answer, err := o.Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "list"}})
	if err != nil || len(answer.Results) != 1 {
		t.Fatalf("must terminate with sourced cards: %+v %v", answer, err)
	}
}

func TestAdvisorReviewEvidenceReferencesAreSentToModel(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1", Status: "ok", ViewKind: "finance_summary", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: int64Ptr(150), SourceRef: "q1"}}}}}}
	calls := 0
	p := reviewProviderFunc(func(_ context.Context, req AdvisorRequest) (AdvisorResponse, error) {
		calls++
		if calls == 1 {
			return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "lookup", Name: "get_finance_summary", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"}}`)}}}, nil
		}
		last := req.Messages[len(req.Messages)-1]
		if !strings.Contains(last.Content, `"source_ref":"lookup.q1.expense"`) {
			t.Fatalf("model never receives resolvable reference: %s", last.Content)
		}
		return AdvisorResponse{Text: "ok"}, nil
	})
	_, err := NewAdvisorOrchestrator(p, NewAdvisorToolRegistry(NewFinanceQueryService(reader))).Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "summary"}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdvisorReviewSnapshotValidatesChangedReferences(t *testing.T) {
	s := NewAdvisorResourceSnapshots()
	s.Record(entity.FinanceResult{Status: "ok", ViewKind: "transaction_detail", View: json.RawMessage(`{"id":"tx-1","version":7}`)})
	for _, raw := range []string{
		`{"action":"update_transaction","transaction_id":"tx-1","expected_version":7,"wallet_id":"foreign-wallet"}`,
		`{"action":"update_transaction","transaction_id":"tx-1","expected_version":7,"category_id":"foreign-category"}`,
	} {
		if err := s.ValidateProposal(entity.FinanceResult{View: json.RawMessage(raw)}); err == nil {
			t.Fatalf("changed reference bypassed snapshot: %s", raw)
		}
	}
}

func TestAdvisorReviewEvidenceRejectsDuplicateCallID(t *testing.T) {
	index := NewAdvisorEvidenceIndex()
	result := entity.FinanceResult{Status: "ok", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: int64Ptr(100)}}}
	if err := index.Add("duplicate", result); err != nil {
		t.Fatal(err)
	}
	if err := index.Add("duplicate", result); err == nil {
		t.Fatal("duplicate call identity accepted")
	}
}

func TestAdvisorReviewSnapshotRejectsArchivedCategoryForNewTransaction(t *testing.T) {
	s := NewAdvisorResourceSnapshots()
	s.Record(entity.FinanceResult{Status: "ok", ViewKind: "wallet_balances", View: json.RawMessage(`{"items":[{"id":"w1"}]}`)})
	s.Record(entity.FinanceResult{Status: "ok", ViewKind: "category_detail", View: json.RawMessage(`{"id":"archived","version":2,"is_active":false,"editable":true}`)})
	err := s.ValidateProposal(entity.FinanceResult{View: json.RawMessage(`{"action":"create_transaction","wallet_id":"w1","category_id":"archived"}`)})
	if err == nil {
		t.Fatal("new transaction can select archived category")
	}
}

func TestAdvisorReviewLoopRejectsStaleAndSystemCategoryProposals(t *testing.T) {
	for _, tc := range []struct{ name, view, args, code string }{
		{"stale", `{"id":"11111111-1111-4111-8111-111111111111","version":7,"editable":true,"is_active":true}`, `{"category_id":"11111111-1111-4111-8111-111111111111","expected_version":6,"name":"New","kind":"expense"}`, "snapshot_mismatch"},
		{"system", `{"id":"11111111-1111-4111-8111-111111111111","version":7,"editable":false,"is_system":true,"is_active":true}`, `{"category_id":"11111111-1111-4111-8111-111111111111","expected_version":7,"name":"New","kind":"expense"}`, "snapshot_not_editable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{Status: "ok", QueryKey: "q1", ViewKind: "category_detail", View: json.RawMessage(tc.view)}}}}
			calls := 0
			p := reviewProviderFunc(func(_ context.Context, req AdvisorRequest) (AdvisorResponse, error) {
				calls++
				if calls == 1 {
					return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "read", Name: "get_category", Arguments: json.RawMessage(`{"category_id":"11111111-1111-4111-8111-111111111111"}`)}}}, nil
				}
				if calls == 2 {
					return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "update", Name: "propose_update_category", Arguments: json.RawMessage(tc.args)}}}, nil
				}
				last := req.Messages[len(req.Messages)-1]
				if last.Role != "tool" || last.ToolCallID != "update" || !strings.Contains(last.Content, tc.code) {
					t.Fatalf("expected correlated %s: %+v", tc.code, last)
				}
				return AdvisorResponse{Text: "Cần đọc lại hoặc chọn danh mục khác."}, nil
			})
			answer, err := NewAdvisorOrchestrator(p, NewAdvisorToolRegistry(NewFinanceQueryService(reader))).Answer(context.Background(), Principal{OwnerID: "owner-1"}, []AdvisorChatMessage{{Role: "user", Content: "Đổi tên danh mục"}})
			if err != nil || len(answer.Results) != 1 || answer.Results[0].Status == "proposed" {
				t.Fatalf("invalid proposal escaped: %+v %v", answer, err)
			}
		})
	}
}

func TestAdvisorReviewInvalidSearchLimitCanBeRepaired(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{Status: "ok", QueryKey: "q1", ViewKind: "transaction_search", View: json.RawMessage(`{"items":[]}`)}}}}
	calls := 0
	p := reviewProviderFunc(func(_ context.Context, req AdvisorRequest) (AdvisorResponse, error) {
		calls++
		if calls == 1 {
			return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "bad", Name: "search_transactions", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"},"limit":100}`)}}}, nil
		}
		if calls == 2 {
			last := req.Messages[len(req.Messages)-1]
			if last.Role != "tool" || last.ToolCallID != "bad" || !strings.Contains(last.Content, "finance_arguments_invalid") {
				t.Fatalf("missing bounded argument error: %+v", last)
			}
			return AdvisorResponse{ToolCalls: []AdvisorToolCall{{ID: "fixed", Name: "search_transactions", Arguments: json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"},"limit":50}`)}}}, nil
		}
		return AdvisorResponse{Text: "Không có giao dịch."}, nil
	})
	answer, err := NewAdvisorOrchestrator(p, NewAdvisorToolRegistry(NewFinanceQueryService(reader))).Answer(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, []AdvisorChatMessage{{Role: "user", Content: "Tháng 9 liệt kê mọi giao dịch"}})
	if err != nil || len(answer.Results) != 1 || len(reader.queries) != 1 || reader.queries[0].Filter.Limit != 50 {
		t.Fatalf("invalid arguments aborted run or reached repository: %+v %v", answer, err)
	}
}
