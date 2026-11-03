package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mypocket/backend/internal/entity"
)

// advisorLoopLimitReply is shown when the tool loop ends without any result.
const advisorLoopLimitReply = "Mình chưa gom đủ dữ liệu trong giới hạn số bước cho phép. Hãy thử hỏi cụ thể hơn (một tháng hoặc một ví)."

// jarContextSuggestions extracts the server-built suggester output carried by a
// message's hidden jar_context part. ok is false when there is nothing to propose.
func jarContextSuggestions(message *entity.AdvisorMessage) (month string, result JarSuggestResult, ok bool) {
	if message == nil {
		return "", JarSuggestResult{}, false
	}
	for _, part := range message.Parts {
		if part.Type != "jar_context" || part.Data == nil {
			continue
		}
		encoded, err := json.Marshal(part.Data)
		if err != nil {
			continue
		}
		var data struct {
			Month     string          `json:"month"`
			Suggested []JarSuggestion `json:"suggested_assignments"`
			Needs     []string        `json:"needs_clarification"`
		}
		if json.Unmarshal(encoded, &data) != nil || len(data.Suggested) == 0 {
			continue
		}
		return data.Month, JarSuggestResult{Suggested: data.Suggested, NeedsClarification: data.Needs}, true
	}
	return "", JarSuggestResult{}, false
}

func hasJarContext(message *entity.AdvisorMessage) bool {
	if message == nil {
		return false
	}
	for _, part := range message.Parts {
		if part.Type == "jar_context" {
			return true
		}
	}
	return false
}

// jarNoEvidenceReply ends a classification request cleanly when the model failed and the deterministic
// suggester found no evidence either, instead of leaving a silent failed run.
const jarNoEvidenceReply = "Mình chưa thấy ghi chú, hashtag (#ten-hu) hoặc khoản trước đó đủ rõ để gợi ý hũ cho các khoản chưa phân bổ, nên không đoán. Bạn có thể thêm hashtag tên hũ vào ghi chú hoặc gắn hũ cho vài khoản mẫu, rồi nhờ mình phân loại lại."

func hasJarProposal(results []entity.FinanceResult) bool {
	for _, result := range results {
		if result.ViewKind == "jar_assignment_proposal" {
			return true
		}
	}
	return false
}

// answerWithJarFallback runs the orchestrator. For a message that carried a
// non-empty jar_context.suggested_assignments, a run that ends without a
// jar_assignment_proposal (finished, loop limit, or provider failure) gets the
// card built by the server from the deterministic suggestions instead.
func (s *AdvisorService) answerWithJarFallback(ctx context.Context, principal Principal, chat []AdvisorChatMessage, userMessage *entity.AdvisorMessage) (AdvisorAnswer, error) {
	answer, err := s.Orchestrator.Answer(ctx, principal, chat)
	month, suggestions, ok := jarContextSuggestions(userMessage)
	if err != nil && (errors.Is(err, ErrAdvisorPrincipalInvalid) || ctx.Err() != nil) {
		return answer, err
	}
	if ok && (err != nil || !hasJarProposal(answer.Results)) {
		if built, buildErr := s.buildJarFallback(ctx, principal, month, suggestions); buildErr == nil {
			built.Usage = answer.Usage
			built.Evidence = answer.Evidence
			built.Results = append(answer.Results, built.Results...)
			return built, nil
		}
	}
	if err != nil && !ok && hasJarContext(userMessage) {
		return AdvisorAnswer{Text: jarNoEvidenceReply, Usage: answer.Usage, Evidence: answer.Evidence}, nil
	}
	if err != nil && errors.Is(err, ErrAdvisorLoopLimit) {
		return AdvisorAnswer{Text: advisorLoopLimitReply, Usage: answer.Usage, Evidence: answer.Evidence}, nil
	}
	return answer, err
}

func (s *AdvisorService) buildJarFallback(ctx context.Context, principal Principal, month string, suggestions JarSuggestResult) (AdvisorAnswer, error) {
	if s.Orchestrator == nil {
		return AdvisorAnswer{}, ErrAdvisorProviderUnavailable
	}
	tool, found := s.Orchestrator.Tools.Lookup("propose_assign_jar")
	if !found {
		return AdvisorAnswer{}, ErrAdvisorToolUnsupported
	}
	rows := make([]map[string]string, 0, len(suggestions.Suggested))
	for _, item := range suggestions.Suggested {
		rows = append(rows, map[string]string{"transaction_id": item.TransactionID, "jar_id": item.JarID, "reason": item.Reason})
	}
	raw, err := json.Marshal(map[string]any{"month": month, "assignments": rows})
	if err != nil {
		return AdvisorAnswer{}, err
	}
	result, err := tool.Execute(ctx, principal, raw)
	if err != nil {
		return AdvisorAnswer{}, err
	}
	text := fmt.Sprintf("Mình đề xuất %d khoản dựa trên ghi chú; %d khoản cần bạn xác nhận thêm.", len(suggestions.Suggested), len(suggestions.NeedsClarification))
	return AdvisorAnswer{Text: text, Results: []entity.FinanceResult{result}}, nil
}
