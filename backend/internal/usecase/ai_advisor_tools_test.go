package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestAdvisorToolRegistryContainsOnlyReadTools(t *testing.T) {
	reader := &financeReaderStub{bundle: entity.FactBundle{Results: []entity.FinanceResult{{QueryKey: "q1"}}}}
	registry := NewAdvisorToolRegistry(NewFinanceQueryService(reader))
	for _, name := range []string{"search_transactions", "get_transaction", "get_finance_summary", "compare_spending_periods", "get_wallet_balances", "get_budget_progress", "get_goal_progress", "get_jar_progress"} {
		if _, ok := registry.Lookup(name); !ok {
			t.Fatalf("missing read tool %s", name)
		}
	}
	for _, forbidden := range []string{"execute_sql", "create_transaction", "confirm_action"} {
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
