package usecase

import (
	"encoding/json"
	"strings"

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
		if (result.ViewKind == "wallet_creation_proposal" || result.ViewKind == "jar_creation_proposal" || result.ViewKind == "jar_update_proposal" || result.ViewKind == "budget_creation_proposal" || result.ViewKind == "budget_update_proposal") && conversationID != "" && generation > 0 && runID != "" {
			if view, ok := data.(map[string]any); ok {
				view["conversation_id"] = conversationID
				view["generation"] = generation
				view["run_id"] = runID
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
	case "product_help":
		return "product_help", true
	default:
		return "", false
	}
}
