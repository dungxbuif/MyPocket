package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrAdvisorProviderUnavailable = errors.New("advisor provider unavailable")
	ErrAdvisorLoopLimit           = errors.New("advisor tool loop limit reached")
	ErrAdvisorPrincipalInvalid    = errors.New("advisor principal is no longer valid")
	ErrAdvisorGroundingRequired   = errors.New("advisor grounded final answer required")
)

const advisorMaxToolCalls = 8

const (
	advisorMaxProviderCalls = 6
	advisorUsageStringLimit = 256
	advisorUsageValueLimit  = int64(1_000_000_000)
)

type AdvisorPrincipalValidator interface {
	Validate(context.Context, Principal) error
}

type AdvisorPrincipalValidatorFunc func(context.Context, Principal) error

func (f AdvisorPrincipalValidatorFunc) Validate(ctx context.Context, principal Principal) error {
	return f(ctx, principal)
}

type AdvisorToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type AdvisorChatMessage struct {
	Role       string            `json:"role"`
	Content    string            `json:"content,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	ToolCalls  []AdvisorToolCall `json:"tool_calls,omitempty"`
}

type AdvisorRequest struct {
	Messages []AdvisorChatMessage
	Tools    []ToolDefinition
}

type AdvisorResponse struct {
	Text      string
	ToolCalls []AdvisorToolCall
	Usage     map[string]any
}

type AdvisorProvider interface {
	Chat(context.Context, AdvisorRequest) (AdvisorResponse, error)
}

type AdvisorAnswer struct {
	Text     string
	Results  []entity.FinanceResult
	Usage    map[string]any
	Evidence *AdvisorEvidenceIndex
}

type AdvisorOrchestrator struct {
	Provider         AdvisorProvider
	Tools            *AdvisorToolRegistry
	Validator        AdvisorPrincipalValidator
	GroundingEnabled bool
}

func NewAdvisorOrchestrator(provider AdvisorProvider, tools *AdvisorToolRegistry) *AdvisorOrchestrator {
	return &AdvisorOrchestrator{Provider: provider, Tools: tools}
}

func mergeUsage(total, current map[string]any) {
	normalized := normalizeAdvisorUsage(current)
	if normalized == nil {
		return
	}
	for _, key := range []string{"prompt_tokens", "completion_tokens", "total_tokens"} {
		value, ok := advisorUsageNumber(normalized[key])
		if !ok {
			continue
		}
		currentTotal, _ := advisorUsageNumber(total[key])
		total[key] = currentTotal + value
	}
	calls, _ := total["calls"].([]map[string]any)
	if len(calls) < advisorMaxProviderCalls {
		total["calls"] = append(calls, normalized)
	}
}

func normalizeAdvisorUsage(current map[string]any) map[string]any {
	if current == nil {
		return nil
	}
	normalized := make(map[string]any, 7)
	for _, key := range []string{"provider", "model", "request_id", "finish_reason"} {
		if text, ok := current[key].(string); ok {
			text = strings.TrimSpace(text)
			if len([]rune(text)) > advisorUsageStringLimit {
				text = string([]rune(text)[:advisorUsageStringLimit])
			}
			if text != "" {
				normalized[key] = text
			}
		}
	}
	if value, ok := advisorUsageNumber(current["latency_ms"]); ok && value <= int64(24*time.Hour/time.Millisecond) {
		normalized["latency_ms"] = value
	}
	for _, key := range []string{"prompt_tokens", "completion_tokens", "total_tokens"} {
		if value, ok := advisorUsageNumber(current[key]); ok {
			normalized[key] = value
		}
	}
	return normalized
}

func advisorUsageNumber(value any) (int64, bool) {
	var number float64
	switch typed := value.(type) {
	case float64:
		number = typed
	case float32:
		number = float64(typed)
	case int:
		return advisorUsageNumber(int64(typed))
	case int64:
		if typed < 0 || typed > advisorUsageValueLimit {
			return 0, false
		}
		return typed, true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return advisorUsageNumber(parsed)
	default:
		return 0, false
	}
	if math.IsNaN(number) || math.IsInf(number, 0) || number < 0 || number > float64(advisorUsageValueLimit) || math.Trunc(number) != number {
		return 0, false
	}
	return int64(number), true
}

func (o *AdvisorOrchestrator) Answer(ctx context.Context, principal Principal, messages []AdvisorChatMessage) (AdvisorAnswer, error) {
	if o == nil || o.Provider == nil || o.Tools == nil || strings.TrimSpace(principal.OwnerID) == "" || len(messages) == 0 {
		return AdvisorAnswer{}, ErrAdvisorProviderUnavailable
	}
	working := make([]AdvisorChatMessage, 0, len(messages)+1)
	system := BuildAdvisorSystemMessage(principal)
	if o.GroundingEnabled {
		system.Content += "\nFor every final answer, call finish_answer. Finance answers must use mode=finance with claims bound to the exact source_ref values in tool facts from this run. For prepared actions use template_id=proposal with no claims. For product_help, read get_product_help then provide source_id from its sources. Server renders labels/help/clarification; explanation and slot text are not displayed. Never put unverified amounts in free text."
		system.Content += " For clarification choose one question_id: intent, wallet, category, period, transaction, amount. Ask only for the next missing input."
		system.Content += " Finance mode REQUIRES template_id: choose summary, period_comparison, transaction, wallet, budget, goal, jar, or proposal. Do not omit template_id. The fact_ref must equal the full source_ref string from the tool result."
	}
	working = append(working, system)
	working = append(working, messages...)
	definitions := make([]ToolDefinition, 0, len(o.Tools.tools))
	for _, tool := range o.Tools.tools {
		definitions = append(definitions, tool.Definition())
	}
	if o.GroundingEnabled {
		definitions = append(definitions, advisorFinishAnswerDefinition())
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	results := make([]entity.FinanceResult, 0, 4)
	evidence := NewAdvisorEvidenceIndex()
	usage := map[string]any{"provider_calls": int64(0), "prompt_tokens": int64(0), "completion_tokens": int64(0), "total_tokens": int64(0)}
	usageSeen := false
	providerCalls := 0
	toolCalls := 0
	groundingRepairs := 0
	snapshots := NewAdvisorResourceSnapshots()
	transcriptBytes := 0
	seenFinanceCalls := map[string]bool{}
	var lastArgumentError error
	for {
		if providerCalls >= advisorMaxProviderCalls {
			return AdvisorAnswer{}, ErrAdvisorLoopLimit
		}
		if o.Validator != nil {
			if err := o.Validator.Validate(ctx, principal); err != nil {
				return AdvisorAnswer{}, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
			}
		}
		availableTools := definitions
		if toolCalls >= advisorMaxToolCalls {
			availableTools = nil
			if o.GroundingEnabled {
				availableTools = []ToolDefinition{advisorFinishAnswerDefinition()}
			}
		}
		response, err := o.Provider.Chat(ctx, AdvisorRequest{Messages: working, Tools: availableTools})
		providerCalls++
		usage["provider_calls"] = int64(providerCalls)
		if err != nil {
			return AdvisorAnswer{}, errors.Join(err, lastArgumentError)
		}
		if len(response.ToolCalls) == 0 {
			if o.GroundingEnabled {
				mergeUsage(usage, response.Usage)
				if response.Usage != nil {
					usageSeen = true
				}
				if groundingRepairs < 1 && providerCalls < advisorMaxProviderCalls {
					groundingRepairs++
					working = append(working,
						AdvisorChatMessage{Role: "assistant", Content: response.Text},
						AdvisorChatMessage{Role: "user", Content: "Hãy chuyển câu trả lời thành một lệnh finish_answer hợp lệ; không trả về văn bản tự do."},
					)
					continue
				}
				if len(results) > 0 {
					return groundedFallbackAnswer(results, usage, usageSeen, evidence), nil
				}
				return AdvisorAnswer{}, ErrAdvisorGroundingRequired
			}
			if strings.TrimSpace(response.Text) == "" {
				return AdvisorAnswer{}, ErrAdvisorProviderUnavailable
			}
			mergeUsage(usage, response.Usage)
			if response.Usage != nil {
				usageSeen = true
			}
			var storedUsage map[string]any
			if usageSeen {
				storedUsage = usage
			}
			return AdvisorAnswer{Text: response.Text, Results: results, Usage: storedUsage, Evidence: evidence}, nil
		}
		// Validate the complete provider batch before executing any call. Tool
		// results are correlated by ID in the next provider request; accepting
		// an empty or duplicate ID would make that request ambiguous and could
		// execute only part of a malformed finance query batch.
		if err := validateAdvisorToolCalls(response.ToolCalls); err != nil {
			return AdvisorAnswer{}, err
		}
		mergeUsage(usage, response.Usage)
		if response.Usage != nil {
			usageSeen = true
		}
		if o.GroundingEnabled {
			if len(response.ToolCalls) != 1 || response.ToolCalls[0].Name != "finish_answer" {
				// A final answer is only valid after the model has finished its
				// read/tool phase; mixed batches are ambiguous and rejected.
				for _, call := range response.ToolCalls {
					if call.Name == "finish_answer" {
						return AdvisorAnswer{}, ErrAdvisorToolInvalid
					}
				}
			} else {
				var final entity.AdvisorFinalAnswer
				if err := DecodeAdvisorToolArgs(response.ToolCalls[0].Arguments, &final); err != nil {
					if groundingRepairs < 1 && providerCalls < advisorMaxProviderCalls {
						groundingRepairs++
						working = appendAdvisorRepair(working, response.ToolCalls, "invalid_final_schema")
						continue
					}
					if len(results) > 0 {
						return groundedFallbackAnswer(results, usage, usageSeen, evidence), nil
					}
					return AdvisorAnswer{}, err
				}
				text, err := RenderAdvisorFinal(final, evidence)
				if err != nil {
					if groundingRepairs < 1 && providerCalls < advisorMaxProviderCalls {
						groundingRepairs++
						working = appendAdvisorRepair(working, response.ToolCalls, "invalid_final_evidence")
						continue
					}
					if len(results) > 0 {
						return groundedFallbackAnswer(results, usage, usageSeen, evidence), nil
					}
					return AdvisorAnswer{}, err
				}
				var storedUsage map[string]any
				if usageSeen {
					storedUsage = usage
				}
				return AdvisorAnswer{Text: text, Results: results, Usage: storedUsage, Evidence: evidence}, nil
			}
		}
		working = append(working, AdvisorChatMessage{Role: "assistant", ToolCalls: response.ToolCalls})
		for _, call := range response.ToolCalls {
			if toolCalls >= advisorMaxToolCalls || strings.TrimSpace(call.Name) == "" {
				return AdvisorAnswer{}, ErrAdvisorLoopLimit
			}
			tool, ok := o.Tools.Lookup(call.Name)
			if !ok {
				return AdvisorAnswer{}, fmt.Errorf("%w: %s", ErrAdvisorToolUnsupported, call.Name)
			}
			if seenFinanceCalls[call.ID] {
				return AdvisorAnswer{}, ErrAdvisorToolInvalid
			}
			seenFinanceCalls[call.ID] = true
			if o.Validator != nil {
				if err := o.Validator.Validate(ctx, principal); err != nil {
					return AdvisorAnswer{}, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
				}
			}
			toolCalls++
			result, err := tool.Execute(ctx, principal, call.Arguments)
			if err != nil {
				if errors.Is(err, entity.ErrFinanceRangeInvalid) || errors.Is(err, entity.ErrFinanceDateInvalid) || errors.Is(err, entity.ErrFinanceRangeTooLarge) || errors.Is(err, entity.ErrFinanceAmountInvalid) {
					lastArgumentError = err
					working = append(working, AdvisorChatMessage{Role: "tool", ToolCallID: call.ID, Content: `{"error":"finance_arguments_invalid","next_action":"Correct arguments using the tool schema: limit 1..50, calendar date range at most 12 months, type all|income|expense, report_scope all|included|excluded. Do not retry unchanged arguments."}`})
					continue
				}
				return AdvisorAnswer{}, err
			}
			if result.Status == "proposed" {
				if err := snapshots.ValidateProposal(result); err != nil {
					code := "snapshot_required"
					if errors.Is(err, ErrAdvisorSnapshotMismatch) {
						code = "snapshot_mismatch"
					}
					if errors.Is(err, ErrAdvisorSnapshotNotEditable) {
						code = "snapshot_not_editable"
					}
					encoded, _ := json.Marshal(map[string]string{"error": code, "next_action": "Read the referenced owner resources and current versions before proposing again."})
					working = append(working, AdvisorChatMessage{Role: "tool", ToolCallID: call.ID, Content: string(encoded)})
					continue
				}
			} else {
				snapshots.Record(result)
			}
			results = append(results, result)
			if err := evidence.Add(call.ID, result); err != nil {
				return AdvisorAnswer{}, err
			}
			wireResult := result
			wireResult.Facts = append([]entity.Fact(nil), result.Facts...)
			for i := range wireResult.Facts {
				wireResult.Facts[i].SourceRef = call.ID + "." + wireResult.Facts[i].ID
			}
			encoded, err := json.Marshal(wireResult)
			if err != nil {
				return AdvisorAnswer{}, err
			}
			transcriptBytes += len(encoded) + len(call.Arguments)
			if transcriptBytes > 96*1024 {
				return groundedFallbackAnswer(results, usage, usageSeen, evidence), nil
			}
			working = append(working, AdvisorChatMessage{Role: "tool", ToolCallID: call.ID, Content: string(encoded)})
		}
	}
}

// Every assistant tool call must have a correlated result, including rejected
// terminal calls. Strict OpenAI-compatible providers reject orphan calls.
func appendAdvisorRepair(messages []AdvisorChatMessage, calls []AdvisorToolCall, code string) []AdvisorChatMessage {
	messages = append(messages, AdvisorChatMessage{Role: "assistant", ToolCalls: calls})
	for _, call := range calls {
		payload, _ := json.Marshal(map[string]string{"error": code, "next_action": "For finance provide template_id (summary|period_comparison|transaction|wallet|budget|goal|jar|proposal), and claims with slot and the full fact source_ref as fact_ref; proposal uses claims=[]. For product_help choose source_id from get_product_help. For clarification choose a supported question_id. Do not include other fields."})
		messages = append(messages, AdvisorChatMessage{Role: "tool", ToolCallID: call.ID, Content: string(payload)})
	}
	return messages
}

func groundedFallbackAnswer(results []entity.FinanceResult, usage map[string]any, usageSeen bool, evidence *AdvisorEvidenceIndex) AdvisorAnswer {
	var storedUsage map[string]any
	if usageSeen {
		storedUsage = usage
	}
	return AdvisorAnswer{Text: "Mình đã lấy được dữ liệu từ hệ thống; hãy xem các thẻ kết quả bên dưới.", Results: results, Usage: storedUsage, Evidence: evidence}
}

func validateAdvisorToolCalls(calls []AdvisorToolCall) error {
	seen := make(map[string]struct{}, len(calls))
	for _, call := range calls {
		id := strings.TrimSpace(call.ID)
		if id == "" {
			return ErrAdvisorToolInvalid
		}
		if _, exists := seen[id]; exists {
			return ErrAdvisorToolInvalid
		}
		seen[id] = struct{}{}
	}
	return nil
}

func advisorFinishAnswerDefinition() ToolDefinition {
	return ToolDefinition{
		Name:        "finish_answer",
		Description: "Finish a grounded answer using only facts returned by finance tools in this run.",
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"mode"},
			"properties": map[string]any{
				"question_id": map[string]any{"type": "string", "enum": []string{"intent", "wallet", "category", "period", "transaction", "amount"}},
				"source_id":   map[string]any{"type": "string", "maxLength": 200},
				"mode":        map[string]any{"type": "string", "enum": []string{"finance", "product_help", "clarification"}},
				"template_id": map[string]any{"type": "string", "description": "Required in finance mode; choose a server template.", "enum": []string{"summary", "period_comparison", "transaction", "wallet", "budget", "goal", "jar", "proposal"}},
				"claims": map[string]any{"type": "array", "maxItems": 32, "items": map[string]any{
					"type": "object", "additionalProperties": false,
					"required":   []string{"slot", "fact_ref"},
					"properties": map[string]any{"slot": map[string]any{"type": "string", "maxLength": 100}, "fact_ref": map[string]any{"type": "string", "maxLength": 200}},
				}},
				"explanation": map[string]any{"type": "string", "maxLength": 2000},
			},
		},
	}
}
