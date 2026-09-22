package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrAdvisorProviderUnavailable = errors.New("advisor provider unavailable")
	ErrAdvisorLoopLimit           = errors.New("advisor tool loop limit reached")
	ErrAdvisorPrincipalInvalid    = errors.New("advisor principal is no longer valid")
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
}

type AdvisorProvider interface {
	Chat(context.Context, AdvisorRequest) (AdvisorResponse, error)
}

type AdvisorAnswer struct {
	Text    string
	Results []entity.FinanceResult
}

type AdvisorOrchestrator struct {
	Provider  AdvisorProvider
	Tools     *AdvisorToolRegistry
	Validator AdvisorPrincipalValidator
}

func NewAdvisorOrchestrator(provider AdvisorProvider, tools *AdvisorToolRegistry) *AdvisorOrchestrator {
	return &AdvisorOrchestrator{Provider: provider, Tools: tools}
}

func (o *AdvisorOrchestrator) Answer(ctx context.Context, principal Principal, messages []AdvisorChatMessage) (AdvisorAnswer, error) {
	if o == nil || o.Provider == nil || o.Tools == nil || strings.TrimSpace(principal.OwnerID) == "" || len(messages) == 0 {
		return AdvisorAnswer{}, ErrAdvisorProviderUnavailable
	}
	working := append([]AdvisorChatMessage(nil), messages...)
	definitions := make([]ToolDefinition, 0, len(o.Tools.tools))
	for _, tool := range o.Tools.tools {
		definitions = append(definitions, tool.Definition())
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	results := make([]entity.FinanceResult, 0, 4)
	providerCalls := 0
	toolCalls := 0
	for {
		if providerCalls >= 6 {
			return AdvisorAnswer{}, ErrAdvisorLoopLimit
		}
		if o.Validator != nil {
			if err := o.Validator.Validate(ctx, principal); err != nil {
				return AdvisorAnswer{}, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
			}
		}
		availableTools := definitions
		if toolCalls >= 4 {
			availableTools = nil
		}
		response, err := o.Provider.Chat(ctx, AdvisorRequest{Messages: working, Tools: availableTools})
		providerCalls++
		if err != nil {
			return AdvisorAnswer{}, err
		}
		if len(response.ToolCalls) == 0 {
			if strings.TrimSpace(response.Text) == "" {
				return AdvisorAnswer{}, ErrAdvisorProviderUnavailable
			}
			return AdvisorAnswer{Text: response.Text, Results: results}, nil
		}
		if toolCalls >= 4 {
			return AdvisorAnswer{}, ErrAdvisorLoopLimit
		}
		working = append(working, AdvisorChatMessage{Role: "assistant", ToolCalls: response.ToolCalls})
		for _, call := range response.ToolCalls {
			if toolCalls >= 4 || strings.TrimSpace(call.Name) == "" {
				return AdvisorAnswer{}, ErrAdvisorLoopLimit
			}
			tool, ok := o.Tools.Lookup(call.Name)
			if !ok {
				return AdvisorAnswer{}, fmt.Errorf("%w: %s", ErrAdvisorToolUnsupported, call.Name)
			}
			if o.Validator != nil {
				if err := o.Validator.Validate(ctx, principal); err != nil {
					return AdvisorAnswer{}, fmt.Errorf("%w: %v", ErrAdvisorPrincipalInvalid, err)
				}
			}
			result, err := tool.Execute(ctx, principal, call.Arguments)
			if err != nil {
				return AdvisorAnswer{}, err
			}
			toolCalls++
			results = append(results, result)
			encoded, err := json.Marshal(result)
			if err != nil {
				return AdvisorAnswer{}, err
			}
			working = append(working, AdvisorChatMessage{Role: "tool", ToolCallID: call.ID, Content: string(encoded)})
		}
	}
}
