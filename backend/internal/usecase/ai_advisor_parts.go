package usecase

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

// BuildAdvisorParts keeps quantitative UI data on the server. The model's
// prose remains separate from the query views, so the frontend never has to
// parse amounts out of free text.
func BuildAdvisorParts(answer AdvisorAnswer) []entity.AdvisorPart {
	return BuildAdvisorPartsForRun(answer, "", 0, "")
}

// BuildAdvisorPartsForRun is the persisted-message variant used by the
// orchestrator. Prepared action cards carry server-owned conversation/run
// context so a later confirmation can be bound to the exact owner generation.
// The context is additive and only attached to proposal views.
func BuildAdvisorPartsForRun(answer AdvisorAnswer, conversationID string, generation int64, runID string) []entity.AdvisorPart {
	parts := make([]entity.AdvisorPart, 0, 1+len(answer.Results))
	if strings.TrimSpace(answer.Text) != "" {
		parts = append(parts, entity.AdvisorPart{Type: "text", Text: answer.Text})
	}
	for _, result := range answer.Results {
		partType, ok := advisorPartType(result.ViewKind)
		if !ok || len(result.View) == 0 || result.Status == "not_found" || result.Status == "unsupported" {
			continue
		}
		var data any
		if err := json.Unmarshal(result.View, &data); err != nil {
			continue
		}
		if isAdvisorProposalView(result.ViewKind) && conversationID != "" && generation > 0 && runID != "" {
			if view, ok := data.(map[string]any); ok {
				view["conversation_id"] = conversationID
				view["generation"] = generation
				view["run_id"] = runID
				if _, exists := view["status"]; !exists {
					view["status"] = entity.AIProposalPending
				}
				if _, exists := view["revision"]; !exists {
					view["revision"] = int64(1)
				}
				if _, exists := view["supported_edit_fields"]; !exists {
					view["supported_edit_fields"] = advisorSupportedEditFields(stringValue(view["action"]))
				}
			}
		}
		parts = append(parts, entity.AdvisorPart{Type: partType, Data: map[string]any{
			"view":     data,
			"source":   result.Source,
			"status":   result.Status,
			"warnings": result.WarningCodes,
		}})
	}
	if len(parts) == 0 {
		return []entity.AdvisorPart{{Type: "text", Text: "Chưa có dữ liệu để hiển thị."}}
	}
	return parts
}

func advisorSupportedEditFields(action string) []string {
	switch action {
	case "create_wallet":
		return []string{"name", "type", "opening_balance", "is_in_total", "description", "target_amount", "target_date", "credit_limit", "last_statement_balance", "statement_day", "payment_due_day"}
	case "create_jar", "update_jar":
		return []string{"month", "name", "allocation_mode", "allocation_amount", "allocation_percent"}
	case "create_budget", "update_budget":
		return []string{"name", "limit_amount", "wallet_id", "category_id", "start_date", "end_date"}
	case "create_transaction", "update_transaction":
		return []string{"wallet_id", "category_id", "type", "amount", "occurred_at", "note", "included_in_reports"}
	case "transfer_wallet":
		return []string{"source_wallet_id", "destination_wallet_id", "amount", "occurred_at", "note"}
	case "create_category", "update_category":
		return []string{"name", "kind", "parent_id", "wallet_ids", "icon_key"}
	case "recategorize_transaction":
		return []string{"category_id"}
	default:
		return []string{}
	}
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func advisorPartType(viewKind string) (string, bool) {
	switch viewKind {
	case "finance_summary":
		return "metric_group", true
	case "transaction_search":
		return "transaction_list", true
	case "transaction_detail":
		return "transaction_detail", true
	case "spending_comparison":
		return "period_comparison", true
	case "wallet_balances":
		return "wallet_balances", true
	case "budget_progress":
		return "budget_status", true
	case "goal_progress":
		return "goal_progress", true
	case "jar_progress":
		return "jar_progress", true
	case "wallet_creation_proposal":
		return "wallet_confirmation", true
	case "jar_creation_proposal":
		return "jar_confirmation", true
	case "jar_update_proposal":
		return "jar_confirmation", true
	case "budget_creation_proposal":
		return "budget_confirmation", true
	case "budget_update_proposal":
		return "budget_confirmation", true
	case "transaction_creation_proposal", "transaction_update_proposal", "transaction_delete_proposal", "transaction_recategorize_proposal":
		return "transaction_confirmation", true
	case "wallet_transfer_proposal":
		return "transfer_confirmation", true
	case "category_creation_proposal", "category_update_proposal":
		return "category_confirmation", true
	case "product_help":
		return "product_help", true
	default:
		return "", false
	}
}

func isAdvisorProposalView(viewKind string) bool {
	partType, ok := advisorPartType(viewKind)
	return ok && (partType == "wallet_confirmation" || partType == "jar_confirmation" || partType == "budget_confirmation" || partType == "transaction_confirmation" || partType == "transfer_confirmation" || partType == "category_confirmation")
}

// narrowReportFollowUpCards keeps a follow-up about one saved report finding focused on that
// finding's evidence. The model may run a broad exploratory search; transaction rows outside the
// finding's category/date/type scope are removed from the card (the card is dropped when nothing
// is left) and, when several lists remain, only the most specific one (fewest rows) is kept.
// Runs without report context keep every card because several lists can be a legitimate answer.
func narrowReportFollowUpCards(parts []entity.AdvisorPart, userMessage *entity.AdvisorMessage) []entity.AdvisorPart {
	if userMessage == nil {
		return parts
	}
	var contextPart *entity.AdvisorPart
	for i := range userMessage.Parts {
		if userMessage.Parts[i].Type == "report_context" {
			contextPart = &userMessage.Parts[i]
			break
		}
	}
	if contextPart == nil {
		return parts
	}
	scope := followUpFindingScope(contextPart.Data)
	narrowed := make([]entity.AdvisorPart, 0, len(parts))
	for _, part := range parts {
		if part.Type == "transaction_list" && scope != nil {
			if filtered, ok := scope.filter(part); ok {
				part = filtered
			} else {
				continue
			}
		}
		narrowed = append(narrowed, part)
	}
	rowCount := func(part entity.AdvisorPart) int {
		data, _ := part.Data.(map[string]any)
		view, _ := data["view"].(map[string]any)
		items, _ := view["items"].([]any)
		return len(items)
	}
	best, lists := -1, 0
	for i, part := range narrowed {
		if part.Type != "transaction_list" {
			continue
		}
		lists++
		if n := rowCount(part); n > 0 && (best < 0 || n <= rowCount(narrowed[best])) {
			best = i
		}
	}
	if lists < 2 || best < 0 {
		return narrowed
	}
	result := make([]entity.AdvisorPart, 0, len(narrowed))
	for i, part := range narrowed {
		if part.Type == "transaction_list" && i != best {
			continue
		}
		result = append(result, part)
	}
	return result
}

type followUpScope struct {
	categories map[string]bool
	from, end  time.Time
	kind       string
}

// followUpFindingScope returns the evidence scope of the single finding carried by a follow-up,
// or nil when the context covers several findings or has no category-scoped drilldown.
func followUpFindingScope(data any) *followUpScope {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var context struct {
		Findings []struct {
			Drilldown struct {
				CategoryIDs  []string  `json:"category_ids"`
				From         time.Time `json:"from"`
				EndExclusive time.Time `json:"end_exclusive"`
				Type         string    `json:"type"`
			} `json:"drilldown"`
		} `json:"findings"`
	}
	if json.Unmarshal(raw, &context) != nil || len(context.Findings) != 1 || len(context.Findings[0].Drilldown.CategoryIDs) == 0 {
		return nil
	}
	d := context.Findings[0].Drilldown
	scope := &followUpScope{categories: map[string]bool{}, from: d.From, end: d.EndExclusive, kind: d.Type}
	for _, id := range d.CategoryIDs {
		scope.categories[id] = true
	}
	return scope
}

func (s *followUpScope) filter(part entity.AdvisorPart) (entity.AdvisorPart, bool) {
	data, _ := part.Data.(map[string]any)
	view, _ := data["view"].(map[string]any)
	items, _ := view["items"].([]any)
	kept := make([]any, 0, len(items))
	for _, item := range items {
		row, _ := item.(map[string]any)
		category, _ := row["category_id"].(string)
		kind, _ := row["type"].(string)
		at, err := time.Parse(time.RFC3339, stringValue(row["occurred_at"]))
		if !s.categories[category] || (s.kind != "" && kind != s.kind) || err != nil ||
			(!s.from.IsZero() && at.Before(s.from)) || (!s.end.IsZero() && !at.Before(s.end)) {
			continue
		}
		kept = append(kept, item)
	}
	if len(kept) == 0 {
		return part, false
	}
	narrowedView := map[string]any{}
	for k, v := range view {
		narrowedView[k] = v
	}
	narrowedView["items"] = kept
	narrowedView["total_count"] = len(kept)
	delete(narrowedView, "next_cursor") // the cursor belongs to the unfiltered search
	narrowedData := map[string]any{}
	for k, v := range data {
		narrowedData[k] = v
	}
	narrowedData["view"] = narrowedView
	part.Data = narrowedData
	return part, true
}
