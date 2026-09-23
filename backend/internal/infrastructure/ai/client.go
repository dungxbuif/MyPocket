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
	"log/slog"
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

// Keep the provider wire contract independent of shared entity JSON tags.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Config struct{ BaseURL, APIKey, Model, OCRURL, OCRKey string }

const (
	maxTextBytes              = 32 * 1024
	maxSourceBytes            = 64 * 1024
	maxPromptBytes            = 128 * 1024
	maxResponseBytes          = 256 * 1024
	maxWalletDescriptionBytes = 2048
	// A cold local Qwen model may need over 30 seconds just to prefill a
	// 5k-token catalog prompt. Keep enough headroom for first-use startup while
	// still bounding a stuck provider request.
	extractTimeout = 180 * time.Second
	modelTimeout   = 90 * time.Second
	modelMaxTokens = 2048
)

var ErrNotConfigured = errors.New("AI provider is not configured")

// AIProviderError carries only a safe stage/code and optional HTTP status. Its
// message never includes upstream response bodies, URLs, credentials, or OCR
// text; callers may still inspect the underlying error with errors.Is/As.
type AIProviderError struct {
	Stage      string
	Code       string
	HTTPStatus int
	Err        error
}

func (e *AIProviderError) Error() string {
	if e == nil {
		return "AI provider request failed"
	}
	switch e.Stage {
	case "ocr_submit":
		return "OCR provider submission failed"
	case "ocr_poll":
		return "OCR provider processing failed"
	case "model":
		return "AI provider request failed"
	case "schema":
		return errSchema.Error()
	default:
		return "AI provider request failed"
	}
}

func (e *AIProviderError) Unwrap() error { return e.Err }

// Diagnostic is intentionally small so the usecase can audit the failure
// without importing this infrastructure package.
func (e *AIProviderError) Diagnostic() (string, string, int) {
	if e == nil {
		return "provider", "provider_error", 0
	}
	return e.Stage, e.Code, e.HTTPStatus
}

type providerRequestError struct {
	Provider    string
	Status      int
	Err         error
	UnknownMode bool
}

func (e *providerRequestError) Error() string {
	if e == nil {
		return "provider request failed"
	}
	if e.Status > 0 {
		return fmt.Sprintf("%s provider returned HTTP %d", e.Provider, e.Status)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s provider request failed", e.Provider)
	}
	return "provider request failed"
}

func (e *providerRequestError) Unwrap() error { return e.Err }

func wrapProviderError(stage, code string, err error) error {
	if err == nil {
		return nil
	}
	var requestErr *providerRequestError
	status := 0
	if errors.As(err, &requestErr) {
		status = requestErr.Status
	}
	return &AIProviderError{Stage: stage, Code: code, HTTPStatus: status, Err: err}
}

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

func (c *Client) ValidateFiles(files []Image) error { return validateImages(files) }

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
	attachmentTexts := make([]string, 0, len(input.Images))
	for _, img := range input.Images {
		extracted, err := c.ocr(ctx, img)
		if err != nil {
			return Output{AttachmentTexts: attachmentTexts}, err
		}
		attachmentTexts = append(attachmentTexts, extracted)
		if source != "" {
			source += "\n\n"
		}
		source += extracted
		if len(source) > maxSourceBytes {
			return Output{SourceText: source, AttachmentTexts: attachmentTexts, OCRComplete: true}, errors.New("OCR source text exceeds limit")
		}
	}
	ocrComplete := true
	messages, err := buildMessages(input, source)
	if err != nil {
		return Output{SourceText: source, AttachmentTexts: attachmentTexts, OCRComplete: ocrComplete}, err
	}
	request := struct {
		Model              string         `json:"model"`
		Messages           []chatMessage  `json:"messages"`
		ResponseFormat     map[string]any `json:"response_format"`
		MaxTokens          int            `json:"max_tokens"`
		Temperature        float64        `json:"temperature"`
		ChatTemplateKwargs map[string]any `json:"chat_template_kwargs"`
	}{
		Model:              c.config.Model,
		Messages:           messages,
		ResponseFormat:     transactionResponseFormat(),
		MaxTokens:          modelMaxTokens,
		Temperature:        0,
		ChatTemplateKwargs: map[string]any{"enable_thinking": false},
	}
	modelCtx, modelCancel := context.WithTimeout(ctx, modelTimeout)
	defer modelCancel()
	modelStarted := time.Now()
	body, _, err := c.request(modelCtx, http.MethodPost, c.config.BaseURL+"/chat/completions", c.config.APIKey, request, "AI")
	if err != nil {
		return Output{SourceText: source, AttachmentTexts: attachmentTexts, OCRComplete: ocrComplete}, wrapProviderError("model", "model_request", err)
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
	if mismatch := new(SchemaMismatchError); errors.As(err, &mismatch) {
		attributes := []any{
			"reason_code", "schema_mismatch",
			"model", c.config.Model,
			"finish_reason", choice.FinishReason,
			"latency_ms", time.Since(modelStarted).Milliseconds(),
			"response_bytes", len(choice.Message.Content),
			"issue_count", len(mismatch.Issues),
		}
		for i, issue := range mismatch.Issues {
			prefix := fmt.Sprintf("issue_%d_", i)
			attributes = append(attributes,
				prefix+"code", issue.Code,
				prefix+"path", issue.Path,
				prefix+"expected", issue.Expected,
				prefix+"actual", issue.Actual,
			)
		}
		slog.Warn("AI provider response rejected", attributes...)
	}
	out.SourceText = source
	out.AttachmentTexts = attachmentTexts
	out.OCRComplete = ocrComplete
	if err != nil {
		var mismatch *SchemaMismatchError
		if errors.As(err, &mismatch) {
			return out, wrapProviderError("schema", "schema_mismatch", err)
		}
		return out, wrapProviderError("schema", "response_invalid", err)
	}
	return out, err
}

func buildMessages(input Input, source string) ([]chatMessage, error) {
	// Only the fields needed to resolve references leave the application.
	type wallet struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Type        string `json:"type"`
		Currency    string `json:"currency"`
		Description string `json:"description,omitempty"`
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
		description := ""
		if w.Description != nil {
			if len(*w.Description) > maxWalletDescriptionBytes {
				return nil, errors.New("AI wallet description exceeds limit")
			}
			description = *w.Description
		}
		wallets = append(wallets, wallet{w.ID, w.Name, w.Type, w.Currency, description})
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
	messages := make([]chatMessage, 0, 2)
	messages = append(messages, chatMessage{Role: "system", Content: extractionPrompt})
	messages = append(messages, chatMessage{Role: "user", Content: string(content)})
	encoded, err := json.Marshal(messages)
	if err != nil || len(encoded) > maxPromptBytes {
		return nil, errors.New("AI input context exceeds limit")
	}
	return messages, nil
}

func transactionResponseFormat() map[string]any {
	draftSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type":                map[string]any{"type": "string", "enum": []string{"income", "expense", "transfer", "unknown"}},
			"amount":              map[string]any{"type": "integer", "minimum": 0},
			"wallet_id":           map[string]any{"type": "string"},
			"category_id":         map[string]any{"type": []string{"string", "null"}},
			"occurred_at":         map[string]any{"type": "string"},
			"note":                map[string]any{"type": "string"},
			"included_in_reports": map[string]any{"type": "boolean"},
			"questions":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required":             []string{"type", "amount", "wallet_id", "category_id", "occurred_at", "note", "included_in_reports", "questions"},
		"additionalProperties": false,
	}
	resultSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"reply":  map[string]any{"type": "string"},
			"drafts": map[string]any{"type": "array", "items": draftSchema},
		},
		"required":             []string{"reply", "drafts"},
		"additionalProperties": false,
	}
	return map[string]any{
		"type":        "json_schema",
		"json_schema": map[string]any{"name": "transaction_proposal", "strict": true, "schema": resultSchema},
	}
}

// request never returns upstream bodies, URLs, or transport errors to callers.
// Ordinary POSTs carry no idempotency header and are never retried by this
// adapter. The OCR wait-mode caller may inspect the bounded, boolean-only
// UnknownMode marker to negotiate with a provider deployment that predates the
// wait fields; it then sends one legacy queue request.
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
		errorBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return nil, nil, &providerRequestError{
			Provider:    provider,
			Status:      resp.StatusCode,
			UnknownMode: provider == "OCR" && resp.StatusCode == http.StatusBadRequest && containsUnknownMode(errorBody),
		}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		return nil, nil, &providerRequestError{Provider: provider, Err: err}
	}
	if len(body) > maxResponseBytes {
		return nil, nil, &providerRequestError{Provider: provider, Err: errors.New("response exceeds limit")}
	}
	return body, resp.Header, nil
}

func containsUnknownMode(body []byte) bool {
	normalized := strings.ToLower(string(body))
	return strings.Contains(normalized, `unknown field "mode"`) || strings.Contains(normalized, `unknown field \"mode\"`) || strings.Contains(normalized, "unknown field 'mode'")
}
