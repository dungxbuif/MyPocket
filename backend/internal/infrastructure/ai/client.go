// Package ai implements read-only text extraction through an HTTP provider.
package ai

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type Input = entity.AIExtractInput
type Output = entity.AIExtractOutput
type Draft = entity.AIExtractDraft
type Image = entity.AIImage
type Message = entity.AIHistoryMessage

// Keep the provider wire contract independent of shared entity JSON tags.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Config struct{ BaseURL, APIKey, Model, OCRURL, OCRKey string }

const (
	maxTextBytes     = 32 * 1024
	maxSourceBytes   = 64 * 1024
	maxPromptBytes   = 128 * 1024
	maxResponseBytes = 256 * 1024
	maxHistory       = 12
	maxHistoryBytes  = 8192
	extractTimeout   = 90 * time.Second
	modelTimeout     = 30 * time.Second
)

var ErrNotConfigured = errors.New("AI provider is not configured")

//go:embed extract.v1.txt
var extractionPrompt string

type Client struct {
	config Config
	http   *http.Client
}

func NewClient(config Config) *Client {
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.OCRURL = strings.TrimRight(strings.TrimSpace(config.OCRURL), "/")
	return &Client{config: config, http: &http.Client{
		Timeout: modelTimeout,
		// A redirect could disclose credentials or replay a document submission.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

func validURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && u.User == nil && u.RawQuery == "" && u.Fragment == ""
}

func (c *Client) Configured() bool {
	return c != nil && validURL(c.config.BaseURL) && strings.TrimSpace(c.config.APIKey) != "" && strings.TrimSpace(c.config.Model) != ""
}

func (c *Client) OCRConfigured() bool {
	return c != nil && validURL(c.config.OCRURL) && strings.TrimSpace(c.config.OCRKey) != ""
}

func (c *Client) Extract(ctx context.Context, input Input) (Output, error) {
	if !c.Configured() {
		return Output{}, ErrNotConfigured
	}
	ctx, cancel := context.WithTimeout(ctx, extractTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Output{}, err
	}
	if len(input.Text) > maxTextBytes || len(input.Timezone) > 128 || (strings.TrimSpace(input.Text) == "" && len(input.Images) == 0) {
		return Output{}, errors.New("invalid AI input: text is empty or exceeds limits")
	}
	if len(input.History) > maxHistory {
		input.History = input.History[len(input.History)-maxHistory:]
	}
	for _, m := range input.History {
		if (m.Role != "user" && m.Role != "assistant") || len(m.Content) > maxHistoryBytes {
			return Output{}, errors.New("invalid AI history")
		}
	}
	if err := validateImages(input.Images); err != nil {
		return Output{}, err
	}
	if len(input.Images) > 0 && !c.OCRConfigured() {
		return Output{}, fmt.Errorf("OCR provider is not configured: %w", ErrNotConfigured)
	}
	// Check the catalog/history budget before submitting any transient image.
	if _, err := buildMessages(input, input.Text); err != nil {
		return Output{}, err
	}
	source := input.Text
	for _, img := range input.Images {
		extracted, err := c.ocr(ctx, img)
		if err != nil {
			return Output{}, err
		}
		if source != "" {
			source += "\n\n"
		}
		source += extracted
		if len(source) > maxSourceBytes {
			return Output{}, errors.New("OCR source text exceeds limit")
		}
	}
	messages, err := buildMessages(input, source)
	if err != nil {
		return Output{SourceText: source}, err
	}
	request := struct {
		Model          string            `json:"model"`
		Messages       []chatMessage     `json:"messages"`
		ResponseFormat map[string]string `json:"response_format"`
	}{c.config.Model, messages, map[string]string{"type": "json_object"}}
	modelCtx, modelCancel := context.WithTimeout(ctx, modelTimeout)
	defer modelCancel()
	body, _, err := c.request(modelCtx, http.MethodPost, c.config.BaseURL+"/chat/completions", c.config.APIKey, request, "AI")
	if err != nil {
		return Output{SourceText: source}, err
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content      string          `json:"content"`
				ToolCalls    json.RawMessage `json:"tool_calls"`
				FunctionCall json.RawMessage `json:"function_call"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if json.Unmarshal(body, &response) != nil || len(response.Choices) != 1 {
		return Output{SourceText: source}, errors.New("invalid AI response envelope")
	}
	choice := response.Choices[0]
	if (choice.FinishReason != "" && choice.FinishReason != "stop") || (len(choice.Message.ToolCalls) > 0 && string(choice.Message.ToolCalls) != "null" && string(choice.Message.ToolCalls) != "[]") || (len(choice.Message.FunctionCall) > 0 && string(choice.Message.FunctionCall) != "null") {
		return Output{SourceText: source}, errors.New("AI response was incomplete or requested tools")
	}
	out, err := parseOutput([]byte(choice.Message.Content))
	out.SourceText = source
	return out, err
}

func buildMessages(input Input, source string) ([]chatMessage, error) {
	// Only the fields needed to resolve references leave the application.
	type wallet struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Currency string `json:"currency"`
	}
	type category struct {
		ID        string   `json:"id"`
		Name      string   `json:"name"`
		Kind      string   `json:"kind"`
		ParentID  *string  `json:"parent_id"`
		WalletIDs []string `json:"wallet_ids"`
	}
	if len(input.Wallets) > 1000 || len(input.Categories) > 2000 {
		return nil, errors.New("AI reference catalog exceeds limit")
	}
	wallets := make([]wallet, 0, len(input.Wallets))
	for _, w := range input.Wallets {
		wallets = append(wallets, wallet{w.ID, w.Name, w.Type, w.Currency})
	}
	categories := make([]category, 0, len(input.Categories))
	for _, c := range input.Categories {
		categories = append(categories, category{c.ID, c.Name, c.Kind, c.ParentID, c.WalletIDs})
	}
	content, err := json.Marshal(struct {
		Timezone   string     `json:"timezone"`
		Now        time.Time  `json:"now"`
		Wallets    []wallet   `json:"wallets"`
		Categories []category `json:"categories"`
		Source     string     `json:"untrusted_source_text"`
	}{input.Timezone, input.Now, wallets, categories, source})
	if err != nil {
		return nil, errors.New("invalid AI reference context")
	}
	messages := make([]chatMessage, 0, len(input.History)+2)
	messages = append(messages, chatMessage{Role: "system", Content: extractionPrompt})
	for _, message := range input.History {
		messages = append(messages, chatMessage{Role: message.Role, Content: message.Content})
	}
	messages = append(messages, chatMessage{Role: "user", Content: string(content)})
	encoded, err := json.Marshal(messages)
	if err != nil || len(encoded) > maxPromptBytes {
		return nil, errors.New("AI input context exceeds limit")
	}
	return messages, nil
}

// request never returns upstream bodies, URLs, or transport errors to callers.
// POSTs carry no idempotency header and are never retried by this adapter.
func (c *Client) request(ctx context.Context, method, endpoint, key string, payload any, provider string) ([]byte, http.Header, error) {
	var reader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("%s request encoding failed", provider)
		}
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, nil, fmt.Errorf("%s request is invalid", provider)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		return nil, nil, fmt.Errorf("%s provider request failed", provider)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("%s provider returned HTTP %d", provider, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		return nil, nil, fmt.Errorf("%s response could not be read", provider)
	}
	if len(body) > maxResponseBytes {
		return nil, nil, fmt.Errorf("%s response exceeds limit", provider)
	}
	return body, resp.Header, nil
}
