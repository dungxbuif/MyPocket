package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mypocket/internal/agent"
)

const maxResponseBytes = 2 << 20

type Client struct {
	baseURL, apiKey, model string
	httpClient             *http.Client
	maxRetries             int
}

func New(baseURL, apiKey, model string, timeout time.Duration, maxRetries int) (*Client, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || strings.TrimSpace(apiKey) == "" || strings.TrimSpace(model) == "" || timeout <= 0 || maxRetries < 0 || maxRetries > 3 {
		return nil, errors.New("invalid OpenAI-compatible provider configuration")
	}
	return &Client{baseURL: parsed.String(), apiKey: apiKey, model: model, httpClient: &http.Client{Timeout: timeout}, maxRetries: maxRetries}, nil
}

func (c *Client) Generate(ctx context.Context, input agent.ModelRequest) (agent.ModelResponse, error) {
	payload := map[string]any{
		"model":           c.model,
		"messages":        []map[string]string{{"role": "system", "content": input.System}, {"role": "user", "content": input.Input}},
		"response_format": map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "mypocket_result", "strict": true, "schema": json.RawMessage(input.Schema)}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return agent.ModelResponse{}, errors.New("encode provider request")
	}
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		result, retry, err := c.do(ctx, body)
		if err == nil || !retry || attempt == c.maxRetries {
			return result, err
		}
		select {
		case <-ctx.Done():
			return agent.ModelResponse{}, errors.New("provider request cancelled")
		case <-time.After(time.Duration(attempt+1) * 50 * time.Millisecond):
		}
	}
	return agent.ModelResponse{}, errors.New("provider unavailable")
}

func (c *Client) do(ctx context.Context, body []byte) (agent.ModelResponse, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return agent.ModelResponse{}, false, errors.New("create provider request")
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return agent.ModelResponse{}, true, errors.New("provider request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retry := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		return agent.ModelResponse{}, retry, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil || len(raw) > maxResponseBytes {
		return agent.ModelResponse{}, false, errors.New("provider response unavailable")
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
				Refusal string `json:"refusal"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(raw, &decoded) != nil || len(decoded.Choices) != 1 {
		return agent.ModelResponse{}, false, errors.New("invalid provider response")
	}
	choice := decoded.Choices[0].Message
	if strings.TrimSpace(choice.Refusal) != "" {
		return agent.ModelResponse{Refusal: "model refused request"}, false, nil
	}
	if strings.TrimSpace(choice.Content) == "" {
		return agent.ModelResponse{}, false, errors.New("empty provider response")
	}
	return agent.ModelResponse{Text: choice.Content}, false, nil
}
