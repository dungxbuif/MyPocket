package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/usecase"
)

const advisorResponseLimit = 256 * 1024

var ErrAdvisorNotConfigured = errors.New("advisor provider is not configured")

type AdvisorClientConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	StoreUsage bool
}

type AdvisorClient struct {
	config AdvisorClientConfig
	http   *http.Client
}

func NewAdvisorClient(config AdvisorClientConfig) *AdvisorClient {
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	return &AdvisorClient{config: config, http: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (c *AdvisorClient) Available() bool {
	return c != nil && validURL(c.config.BaseURL) && strings.TrimSpace(c.config.APIKey) != "" && strings.TrimSpace(c.config.Model) != ""
}

func (c *AdvisorClient) Chat(ctx context.Context, input usecase.AdvisorRequest) (usecase.AdvisorResponse, error) {
	if !c.Available() {
		return usecase.AdvisorResponse{}, ErrAdvisorNotConfigured
	}
	started := time.Now()
	slog.Info("AI advisor model request started", "model", c.config.Model, "message_count", len(input.Messages), "tool_count", len(input.Tools))
	type wireToolCall struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	}
	type wireMessage struct {
		Role       string         `json:"role"`
		Content    string         `json:"content,omitempty"`
		ToolCallID string         `json:"tool_call_id,omitempty"`
		ToolCalls  []wireToolCall `json:"tool_calls,omitempty"`
	}
	messages := make([]wireMessage, 0, len(input.Messages))
	for _, message := range input.Messages {
		wire := wireMessage{Role: message.Role, Content: message.Content, ToolCallID: message.ToolCallID}
		for _, call := range message.ToolCalls {
			arguments := string(call.Arguments)
			if strings.TrimSpace(arguments) == "" {
				arguments = "{}"
			}
			wireCall := wireToolCall{ID: call.ID, Type: "function"}
			wireCall.Function.Name = call.Name
			wireCall.Function.Arguments = arguments
			wire.ToolCalls = append(wire.ToolCalls, wireCall)
		}
		messages = append(messages, wire)
	}
	tools := make([]map[string]any, 0, len(input.Tools))
	for _, tool := range input.Tools {
		tools = append(tools, map[string]any{"type": "function", "function": map[string]any{"name": tool.Name, "description": tool.Description, "parameters": tool.Parameters}})
	}
	payload := map[string]any{"model": c.config.Model, "messages": messages}
	if len(tools) > 0 {
		payload["tools"] = tools
		payload["tool_choice"] = "auto"
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return usecase.AdvisorResponse{}, ErrAdvisorNotConfigured
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return usecase.AdvisorResponse{}, ErrAdvisorNotConfigured
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		slog.Warn("AI advisor model request failed", "model", c.config.Model, "latency_ms", time.Since(started).Milliseconds(), "error_kind", advisorErrorKind(err))
		if ctx.Err() != nil {
			return usecase.AdvisorResponse{}, ctx.Err()
		}
		return usecase.AdvisorResponse{}, ErrAdvisorNotConfigured
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return usecase.AdvisorResponse{}, fmt.Errorf("advisor provider returned HTTP %d", resp.StatusCode)
	}
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, advisorResponseLimit+1))
	if err != nil || len(responseBody) > advisorResponseLimit {
		return usecase.AdvisorResponse{}, ErrAdvisorNotConfigured
	}
	var envelope struct {
		ID      string          `json:"id"`
		Usage   json.RawMessage `json:"usage"`
		Choices []struct {
			Message struct {
				Content   *string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(responseBody, &envelope); err != nil || len(envelope.Choices) != 1 {
		return usecase.AdvisorResponse{}, ErrAdvisorNotConfigured
	}
	choice := envelope.Choices[0].Message
	response := usecase.AdvisorResponse{}
	if choice.Content != nil {
		response.Text = *choice.Content
	}
	for _, call := range choice.ToolCalls {
		response.ToolCalls = append(response.ToolCalls, usecase.AdvisorToolCall{ID: call.ID, Name: call.Function.Name, Arguments: json.RawMessage(call.Function.Arguments)})
	}
	usage := map[string]any{"provider": "openai_compatible", "model": c.config.Model, "request_id": envelope.ID, "latency_ms": time.Since(started).Milliseconds()}
	if len(envelope.Usage) > 0 && string(envelope.Usage) != "null" {
		var providerUsage map[string]any
		if json.Unmarshal(envelope.Usage, &providerUsage) == nil {
			for key, value := range providerUsage {
				usage[key] = value
			}
		}
	}
	slog.Info("AI advisor model response completed", "model", c.config.Model, "latency_ms", time.Since(started).Milliseconds(), "prompt_tokens", advisorUsageInt(usage, "prompt_tokens"), "completion_tokens", advisorUsageInt(usage, "completion_tokens"), "total_tokens", advisorUsageInt(usage, "total_tokens"), "tool_call_count", len(response.ToolCalls), "response_bytes", len(responseBody))
	if c.config.StoreUsage {
		response.Usage = usage
	}
	return response, nil
}

func advisorUsageInt(usage map[string]any, key string) int64 {
	switch value := usage[key].(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	case int:
		return int64(value)
	default:
		return 0
	}
}

func advisorErrorKind(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	return "transport_error"
}
