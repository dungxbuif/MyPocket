package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestAdvisorToolRegistryContainsOnlyReadTools(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1"}}}}
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(reader))
	for _, name := range []string{"search_transactions", "get_transaction", "get_finance_summary", "compare_spending_periods", "get_wallet_balances", "get_budget_progress", "get_goal_progress", "get_jar_progress", "get_product_help", "propose_create_wallet", "propose_create_jar", "propose_update_jar", "propose_create_budget", "propose_update_budget"} {
		if _, ok := registry.Lookup(name); !ok {
			t.Fatalf("missing read tool %s", name)
		}
	}
	for _, forbidden := range []string{"execute_sql", "create_transaction", "create_wallet", "confirm_action"} {
		if _, ok := registry.Lookup(forbidden); ok {
			t.Fatalf("write/escape tool registered: %s", forbidden)
		}
	}
}

func TestAdvisorToolDefinitionsExposeModelParameters(t *testing.T) {
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{}))
	expected := map[string][]string{
		"get_finance_summary":      {"range", "wallet_ids", "category_ids", "include_children"},
		"search_transactions":      {"range", "wallet_ids", "category_ids", "include_children", "type", "report_scope", "note_contains", "min_amount", "max_amount", "limit", "cursor"},
		"compare_spending_periods": {"current", "previous", "wallet_ids", "category_ids", "include_children"},
		"get_wallet_balances":      {"wallet_ids"},
		"get_budget_progress":      {"budget_ids", "active_on"},
		"get_goal_progress":        {"wallet_ids"},
		"get_jar_progress":         {"month", "jar_ids"},
		"get_transaction":          {"transaction_id"},
		"propose_create_wallet":    {"name", "type", "opening_balance", "is_in_total", "description", "target_amount", "target_date", "credit_limit", "last_statement_balance", "statement_day", "payment_due_day"},
		"propose_create_jar":       {"month", "name", "allocation_mode", "allocation_amount", "allocation_percent"},
		"propose_update_jar":       {"jar_id", "month", "name", "allocation_mode", "allocation_amount", "allocation_percent"},
		"propose_create_budget":    {"name", "limit_amount", "wallet_id", "category_id", "start_date", "end_date"},
		"propose_update_budget":    {"budget_id", "name", "limit_amount", "wallet_id", "category_id", "start_date", "end_date"},
		"get_product_help":         {"query", "max_results"},
	}
	for name, properties := range expected {
		tool, ok := registry.Lookup(name)
		if !ok {
			t.Fatalf("missing tool %s", name)
		}
		definition := tool.Definition()
		actual, ok := definition.Parameters["properties"].(map[string]any)
		if !ok {
			t.Fatalf("%s schema must expose properties, got %#v", name, definition.Parameters)
		}
		for _, property := range properties {
			if _, ok := actual[property]; !ok {
				t.Errorf("%s schema missing property %s", name, property)
			}
		}
	}
}

func TestAdvisorToolRejectsOwnerOverrideAndUnknownFields(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1"}}}}
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(reader))
	tool, ok := registry.Lookup("get_finance_summary")
	if !ok {
		t.Fatal("summary tool missing")
	}
	_, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, json.RawMessage(`{"range":{"from":"2026-09-01","to":"2026-09-30"},"owner_id":"other"}`))
	if err == nil {
		t.Fatal("model must not supply owner_id")
	}
}

func TestAdvisorGetTransactionUsesExactOwnerScopedID(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1", ViewKind: "transaction_detail"}}}}
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(reader))
	tool, ok := registry.Lookup("get_transaction")
	if !ok {
		t.Fatal("get_transaction tool missing")
	}
	_, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, json.RawMessage(`{"transaction_id":"tx-1"}`))
	if err != nil {
		t.Fatal(err)
	}
	if reader.owner != "owner-1" || len(reader.queries) != 1 {
		t.Fatalf("unexpected query delegation: owner=%q queries=%+v", reader.owner, reader.queries)
	}
	query := reader.queries[0]
	if query.Kind != "get_transaction" || len(query.ObjectIDs) != 1 || query.ObjectIDs[0] != "tx-1" {
		t.Fatalf("unexpected transaction query: %+v", query)
	}
}

func TestAdvisorGetTransactionRequiresOneID(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1"}}}}
	tool, _ := NewAdvisorToolRegistry(NewFinanceQueryService(reader)).Lookup("get_transaction")
	if _, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1", Timezone: "Asia/Ho_Chi_Minh"}, json.RawMessage(`{"transaction_id":""}`)); err == nil {
		t.Fatal("empty transaction id must be rejected")
	}
}

func TestAdvisorWalletProposalValidatesAndDoesNotWrite(t *testing.T) {
	tool, ok := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup("propose_create_wallet")
	if !ok {
		t.Fatal("wallet proposal tool missing")
	}
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage([]byte("{\"name\":\"Tiền mặt\",\"type\":\"basic\",\"opening_balance\":250000}")))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "proposed" || result.ViewKind != "wallet_creation_proposal" || !strings.Contains(string(result.View), "create_wallet") {
		t.Fatalf("unexpected wallet proposal: %+v", result)
	}
	var proposal struct {
		ActionID string `json:"action_id"`
	}
	if err := json.Unmarshal(result.View, &proposal); err != nil {
		t.Fatal(err)
	}
	if proposal.ActionID == "" {
		t.Fatal("wallet proposal must carry a durable action_id for replay-safe confirmation")
	}
	for _, raw := range []string{
		"{\"name\":\"Goal\",\"type\":\"goal\"}",
		"{\"name\":\"Card\",\"type\":\"credit\"}",
		"{\"name\":\"Cash\",\"type\":\"basic\",\"owner_id\":\"other\"}",
	} {
		if _, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage([]byte(raw))); err == nil {
			t.Fatalf("invalid wallet proposal must be rejected: %s", raw)
		}
	}
}

func TestAdvisorJarProposalValidatesAndDoesNotWrite(t *testing.T) {
	tool, ok := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup("propose_create_jar")
	if !ok {
		t.Fatal("jar proposal tool missing")
	}
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"month":"2026-09","name":"Highlands","allocation_mode":"fixed","allocation_amount":800000}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "proposed" || result.ViewKind != "jar_creation_proposal" || !strings.Contains(string(result.View), "create_jar") {
		t.Fatalf("unexpected jar proposal: %+v", result)
	}
	var proposal struct {
		ActionID string `json:"action_id"`
		Month    string `json:"month"`
	}
	if err := json.Unmarshal(result.View, &proposal); err != nil {
		t.Fatal(err)
	}
	if proposal.ActionID == "" || proposal.Month != "2026-09" {
		t.Fatalf("jar proposal must carry action and month: %+v", proposal)
	}
	for _, raw := range []string{
		`{"month":"2026-09","name":"Bad","allocation_mode":"fixed"}`,
		`{"month":"2026-09","name":"Bad","allocation_mode":"percent","allocation_percent":101}`,
		`{"month":"2026-09","name":"Bad","allocation_mode":"none","allocation_amount":1}`,
		`{"month":"2026-09","name":"Bad","allocation_mode":"fixed","owner_id":"other"}`,
	} {
		if _, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(raw)); err == nil {
			t.Fatalf("invalid jar proposal must be rejected: %s", raw)
		}
	}
}

func TestAdvisorJarUpdateProposalValidatesAndDoesNotWrite(t *testing.T) {
	tool, ok := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup("propose_update_jar")
	if !ok {
		t.Fatal("jar update proposal tool missing")
	}
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"jar_id":"11111111-1111-4111-8111-111111111111","month":"2026-09","name":"Highlands","allocation_mode":"percent","allocation_percent":20}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "proposed" || result.ViewKind != "jar_update_proposal" || !strings.Contains(string(result.View), "update_jar") {
		t.Fatalf("unexpected jar update proposal: %+v", result)
	}
	var proposal struct {
		ActionID string `json:"action_id"`
		JarID    string `json:"jar_id"`
		Month    string `json:"month"`
	}
	if err := json.Unmarshal(result.View, &proposal); err != nil {
		t.Fatal(err)
	}
	if proposal.ActionID == "" || proposal.JarID != "11111111-1111-4111-8111-111111111111" || proposal.Month != "2026-09" {
		t.Fatalf("jar update proposal must carry action, jar and month: %+v", proposal)
	}
	for _, raw := range []string{
		`{"jar_id":"","month":"2026-09","name":"Bad","allocation_mode":"none"}`,
		`{"jar_id":"11111111-1111-4111-8111-111111111111","month":"2026-09","name":"Bad","allocation_mode":"fixed"}`,
		`{"jar_id":"11111111-1111-4111-8111-111111111111","month":"2026-09","name":"Bad","allocation_mode":"percent","allocation_percent":101}`,
	} {
		if _, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(raw)); err == nil {
			t.Fatalf("invalid jar update proposal must be rejected: %s", raw)
		}
	}
}

func TestAdvisorBudgetProposalValidatesAndDoesNotWrite(t *testing.T) {
	tool, ok := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup("propose_create_budget")
	if !ok {
		t.Fatal("budget proposal tool missing")
	}
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"name":"Ăn uống","limit_amount":4000000,"start_date":"2026-09-01","end_date":"2026-09-30"}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "proposed" || result.ViewKind != "budget_creation_proposal" || !strings.Contains(string(result.View), "create_budget") {
		t.Fatalf("unexpected budget proposal: %+v", result)
	}
	for _, raw := range []string{
		`{"name":"Bad","limit_amount":0,"start_date":"2026-09-01","end_date":"2026-09-30"}`,
		`{"name":"Bad","limit_amount":1,"start_date":"2026-09-30","end_date":"2026-09-01"}`,
		`{"name":"Bad","limit_amount":1,"start_date":"2026-09-01","end_date":"2026-09-30","wallet_id":"not-a-uuid"}`,
		`{"name":"Bad","limit_amount":1,"start_date":"2026-09-01","end_date":"2026-09-30","owner_id":"other"}`,
	} {
		if _, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(raw)); err == nil {
			t.Fatalf("invalid budget proposal must be rejected: %s", raw)
		}
	}
}

func TestAdvisorBudgetUpdateProposalValidatesAndDoesNotWrite(t *testing.T) {
	tool, ok := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup("propose_update_budget")
	if !ok {
		t.Fatal("budget update proposal tool missing")
	}
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"budget_id":"11111111-1111-4111-8111-111111111111","name":"Ăn uống","limit_amount":5000000,"start_date":"2026-09-01","end_date":"2026-09-30"}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "proposed" || result.ViewKind != "budget_update_proposal" || !strings.Contains(string(result.View), "update_budget") {
		t.Fatalf("unexpected budget update proposal: %+v", result)
	}
	for _, raw := range []string{
		`{"budget_id":"bad","name":"Bad","limit_amount":1,"start_date":"2026-09-01","end_date":"2026-09-30"}`,
		`{"budget_id":"11111111-1111-4111-8111-111111111111","name":"Bad","limit_amount":1,"start_date":"2026-09-30","end_date":"2026-09-01"}`,
		`{"budget_id":"11111111-1111-4111-8111-111111111111","name":"Bad","limit_amount":1,"start_date":"2026-09-01","end_date":"2026-09-30","owner_id":"other"}`,
	} {
		if _, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(raw)); err == nil {
			t.Fatalf("invalid budget update proposal must be rejected: %s", raw)
		}
	}
}

func TestAdvisorProductHelpIsVersionedAndSourceIdentifiable(t *testing.T) {
	tool, ok := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup("get_product_help")
	if !ok {
		t.Fatal("product help tool missing")
	}
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"query":"làm sao tạo ngân sách","max_results":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" || result.ViewKind != "product_help" {
		t.Fatalf("unexpected product help result: %+v", result)
	}
	var view struct {
		Version string `json:"version"`
		Sources []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Excerpt string `json:"excerpt"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(result.View, &view); err != nil {
		t.Fatal(err)
	}
	if view.Version == "" || len(view.Sources) == 0 || view.Sources[0].ID == "" || view.Sources[0].Title == "" || view.Sources[0].Excerpt == "" {
		t.Fatalf("product help must include versioned sources: %+v", view)
	}
	if len(view.Sources) > 2 {
		t.Fatalf("max_results was not applied: %d", len(view.Sources))
	}
}

func TestAdvisorProductHelpNoMatchDoesNotDumpHandbook(t *testing.T) {
	tool, _ := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup("get_product_help")
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"query":"tính năng không tồn tại xyz"}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "not_found" || len(result.View) == 0 || strings.Contains(string(result.View), "handbook") {
		t.Fatalf("no-match must be explicit and bounded: %+v", result)
	}
}

func TestAdvisorProductHelpRejectsBoundsAndUnknownFields(t *testing.T) {
	tool, _ := NewAdvisorToolRegistry(NewFinanceQueryService(&financeReaderStub{})).Lookup("get_product_help")
	for _, raw := range []string{
		`{"query":""}`,
		`{"query":"x","max_results":0}`,
		`{"query":"x","max_results":6}`,
		`{"query":"x","owner_id":"other"}`,
	} {
		if _, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(raw)); err == nil {
			t.Fatalf("invalid product-help input accepted: %s", raw)
		}
	}
}
