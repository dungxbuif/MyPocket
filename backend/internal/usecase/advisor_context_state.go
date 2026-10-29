package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

const (
	advisorContextTextBudget = 24 * 1024
	advisorContextRefLimit   = 8
	advisorContextStateLimit = 4 * 1024
)

var ErrAdvisorContextStateInvalid = errors.New("advisor context state is invalid")

func BuildAdvisorContextState(previous entity.AdvisorContextState, generation, throughSeq int64, results []entity.FinanceResult) (entity.AdvisorContextState, error) {
	if generation < 1 || throughSeq < 0 {
		return entity.AdvisorContextState{}, ErrAdvisorContextStateInvalid
	}
	state := entity.AdvisorContextState{Version: 1, Generation: generation, ThroughSeq: throughSeq, UpdatedAt: time.Now().UTC()}
	if previous.Generation == generation {
		state = previous
		state.Version = 1
		state.Generation = generation
		state.ThroughSeq = throughSeq
		state.UpdatedAt = time.Now().UTC()
	}
	for _, result := range results {
		if result.Status != "" && result.Status != "ok" {
			continue
		}
		state.LastQueryKind = result.ViewKind
		if result.Source.Timezone != "" {
			state.Timezone = result.Source.Timezone
		}
		if result.Source.Filter.Range.From != "" && result.Source.Filter.Range.To != "" {
			state.DateFrom, state.DateTo = result.Source.Filter.Range.From, result.Source.Filter.Range.To
		} else if result.Source.StartAt != nil && result.Source.EndExclusive != nil {
			location, err := time.LoadLocation(state.Timezone)
			if err != nil {
				return entity.AdvisorContextState{}, ErrAdvisorContextStateInvalid
			}
			state.DateFrom = result.Source.StartAt.In(location).Format("2006-01-02")
			state.DateTo = result.Source.EndExclusive.In(location).AddDate(0, 0, -1).Format("2006-01-02")
		}
		explicitQuery := result.ViewKind == "finance_summary" || result.ViewKind == "transaction_search" || result.ViewKind == "spending_comparison"
		if explicitQuery || len(result.Source.Filter.WalletIDs) > 0 {
			state.WalletRefs = appendUniqueBounded(nil, result.Source.Filter.WalletIDs, advisorContextRefLimit)
		}
		if explicitQuery || len(result.Source.Filter.CategoryIDs) > 0 {
			state.CategoryRefs = appendUniqueBounded(nil, result.Source.Filter.CategoryIDs, advisorContextRefLimit)
		}
		if result.ViewKind == "transaction_detail" {
			var view struct {
				ID string `json:"id"`
			}
			if json.Unmarshal(result.View, &view) == nil && strings.TrimSpace(view.ID) != "" {
				state.TransactionRefs = appendUniqueBounded(state.TransactionRefs, []string{view.ID}, advisorContextRefLimit)
			}
		}
	}
	state.LastAmount, state.LastVersion = 0, 0
	encoded, err := json.Marshal(state)
	if err != nil || len(encoded) > advisorContextStateLimit {
		return entity.AdvisorContextState{}, ErrAdvisorContextStateInvalid
	}
	return state, nil
}

func appendUniqueBounded(existing, values []string, limit int) []string {
	seen := make(map[string]struct{}, len(existing)+len(values))
	out := make([]string, 0, limit)
	for _, value := range append(existing, values...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
		if len(out) == limit {
			break
		}
	}
	return out
}

// BuildAdvisorRequestContext preserves whole UTF-8 messages and removes the
// oldest content first. The current (last) message is never truncated.
func BuildAdvisorRequestContext(history []entity.AdvisorMessage, state entity.AdvisorContextState) ([]AdvisorChatMessage, error) {
	ordered := append([]entity.AdvisorMessage(nil), history...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Seq < ordered[j].Seq })
	chat := make([]AdvisorChatMessage, 0, len(ordered)+1)
	for _, message := range ordered {
		if content := messageText(message); content != "" {
			chat = append(chat, AdvisorChatMessage{Role: message.Role, Content: content})
		}
	}
	if len(chat) == 0 {
		return nil, ErrAdvisorContextStateInvalid
	}
	if len(chat) > advisorContextMessageLimit {
		chat = chat[len(chat)-advisorContextMessageLimit:]
	}
	for len(chat) > 1 && advisorChatBytes(chat) > advisorContextTextBudget {
		chat = chat[1:]
	}
	if advisorChatBytes(chat) > advisorContextTextBudget {
		// A single current message is an API validation error rather than a
		// malformed UTF-8 truncation; the caller can ask the user to shorten it.
		return nil, ErrAdvisorContextStateInvalid
	}
	if state.Version == 1 && state.Generation > 0 {
		// Only reference/filter fields enter the model; no persisted narrative,
		// note, amount or optimistic-lock version is reused.
		hint := map[string]any{"date_from": state.DateFrom, "date_to": state.DateTo, "timezone": state.Timezone, "wallet_refs": state.WalletRefs, "category_refs": state.CategoryRefs, "transaction_refs": state.TransactionRefs}
		payload, _ := json.Marshal(hint)
		if len(payload) > advisorContextStateLimit {
			return nil, ErrAdvisorContextStateInvalid
		}
		content := "Server-validated continuation hints, not current financial facts. Re-query amounts and versions; ask for clarification if multiple references are plausible: " + string(payload)
		for len(chat) > 1 && advisorChatBytes(chat)+len(content) > advisorContextTextBudget {
			chat = chat[1:]
		}
		if advisorChatBytes(chat)+len(content) > advisorContextTextBudget {
			return nil, ErrAdvisorContextStateInvalid
		}
		chat = append([]AdvisorChatMessage{{Role: "system", Content: content}}, chat...)
	}
	return chat, nil
}

func revalidateAdvisorContext(ctx context.Context, principal Principal, registry *AdvisorToolRegistry, state entity.AdvisorContextState) (entity.AdvisorContextState, error) {
	state.LastAmount, state.LastVersion = 0, 0
	state.SelectedTopic = ""
	state.Timezone = principal.Timezone
	validate := func(name string, args any, id string) bool {
		tool, ok := registry.Lookup(name)
		if !ok {
			return false
		}
		raw, _ := json.Marshal(args)
		result, err := tool.Execute(ctx, principal, raw)
		if err != nil || result.Status != "ok" {
			return false
		}
		var view struct {
			ID    string `json:"id"`
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			IsActive *bool `json:"is_active"`
		}
		if json.Unmarshal(result.View, &view) != nil || view.IsActive != nil && !*view.IsActive {
			return false
		}
		if view.ID == id {
			return true
		}
		for _, item := range view.Items {
			if item.ID == id {
				return true
			}
		}
		return false
	}
	for _, field := range []struct {
		refs      *[]string
		tool, key string
	}{{&state.WalletRefs, "get_wallet_balances", "wallet_ids"}, {&state.CategoryRefs, "get_category", "category_id"}, {&state.TransactionRefs, "get_transaction", "transaction_id"}} {
		refs := appendUniqueBounded(nil, *field.refs, advisorContextRefLimit)
		*field.refs = nil
		for _, id := range refs {
			var value any = id
			if field.key == "wallet_ids" {
				value = []string{id}
			}
			if validate(field.tool, map[string]any{field.key: value}, id) {
				*field.refs = append(*field.refs, id)
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return entity.AdvisorContextState{}, err
	}
	return state, nil
}

func advisorChatBytes(chat []AdvisorChatMessage) int {
	total := 0
	for _, message := range chat {
		total += len([]byte(message.Content))
	}
	return total
}
