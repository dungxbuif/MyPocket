package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/usecase"
)

const advisorResponseLimit = 256 * 1024

var ErrAdvisorNotConfigured = errors.New("advisor provider is not configured")

type AdvisorClientConfig struct {
	BaseURL string
	APIKey  string
	Model   string
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
	return response, nil
}
