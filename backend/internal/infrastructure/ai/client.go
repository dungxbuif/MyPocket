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

type Config struct {
	BaseURL, APIKey, Model, OCRURL, OCRKey string
	StoreUsage                             bool
}

const (
	maxTextBytes              = 32 * 1024
	maxSourceBytes            = 64 * 1024
	maxPromptBytes            = 128 * 1024
	maxResponseBytes          = 256 * 1024
	maxWalletDescriptionBytes = 2048
	// A cold local Qwen model may need over 30 seconds just to prefill a
	// 5k-token catalog prompt. Keep enough headroom for first-use startup while
	// still bounding a stuck provider request.
	extractTimeout = 210 * time.Second
	modelTimeout   = 180 * time.Second
	// The application has no draft-count cap. This is only a per-response
	// transport budget; schema parsing and persistence accept every draft.
	// A batch may contain up to twenty attachments and has no application
	// draft-count cap. Keep a generous provider envelope so complete JSON is
	// not rejected merely because a large review batch reaches the old 8k cap.
	modelMaxTokens = 32768
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

func messageBytes(messages []chatMessage) int {
	encoded, err := json.Marshal(messages)
	if err != nil {
		return 0
	}
	return len(encoded)
}

func providerErrorKind(err error) string {
	if err == nil {
		return "none"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	var requestErr *providerRequestError
	if errors.As(err, &requestErr) && requestErr.Status > 0 {
		return fmt.Sprintf("http_%d", requestErr.Status)
	}
	return "transport_error"
}

func modelUsage(requestID string, raw json.RawMessage, finish string, latency time.Duration) map[string]any {
	usage := map[string]any{
		"provider":          "openai_compatible",
		"request_id":        requestID,
		"finish_reason":     finish,
		"latency_ms":        latency.Milliseconds(),
		"prompt_tokens":     0,
		"completion_tokens": 0,
		"total_tokens":      0,
	}
	if len(raw) > 0 && string(raw) != "null" {
		var providerUsage map[string]any
		if json.Unmarshal(raw, &providerUsage) == nil {
			usage["provider_usage"] = providerUsage
			for _, field := range []string{"prompt_tokens", "completion_tokens", "total_tokens"} {
				if value, ok := providerUsage[field]; ok {
					usage[field] = value
				}
			}
		}
	}
	return usage
}

func modelUsageInt(usage map[string]any, key string) int64 {
	if usage == nil {
		return 0
	}
	switch value := usage[key].(type) {
	case float64:
		return int64(value)
	case json.Number:
		parsed, _ := value.Int64()
		return parsed
	case int:
		return int64(value)
	case int64:
		return value
	default:
		return 0
	}
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
	instruction := strings.TrimSpace(input.Instruction)
	if instruction == "" {
		instruction = strings.TrimSpace(input.Text)
	}
	slog.Info("AI extraction started", "model", c.config.Model, "instruction_bytes", len(instruction), "image_count", len(input.Images), "wallet_count", len(input.Wallets), "category_count", len(input.Categories))
	if !c.Configured() {
		slog.Warn("AI extraction unavailable", "reason_code", "provider_not_configured")
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
		ocrStarted := time.Now()
		slog.Info("AI OCR started", "attachment_index", len(attachmentTexts), "mime_type", img.MIMEType, "bytes", len(img.Base64))
		extracted, err := c.ocr(ctx, img)
		if err != nil {
			stage, code, status := "ocr", "ocr_error", 0
			var diagnostic *AIProviderError
			if errors.As(err, &diagnostic) {
				stage, code, status = diagnostic.Diagnostic()
			}
			slog.Warn("AI OCR failed", "attachment_index", len(attachmentTexts), "stage", stage, "code", code, "provider_status", status, "latency_ms", time.Since(ocrStarted).Milliseconds())
			return Output{AttachmentTexts: attachmentTexts}, err
		}
		slog.Info("AI OCR completed", "attachment_index", len(attachmentTexts), "text_bytes", len(extracted), "latency_ms", time.Since(ocrStarted).Milliseconds())
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
	slog.Info("AI model request started", "model", c.config.Model, "source_bytes", len(source), "prompt_bytes", messageBytes(messages), "max_tokens", modelMaxTokens)
	body, _, err := c.request(modelCtx, http.MethodPost, c.config.BaseURL+"/chat/completions", c.config.APIKey, request, "AI")
	if err != nil {
		slog.Warn("AI model request failed", "model", c.config.Model, "code", "model_request", "latency_ms", time.Since(modelStarted).Milliseconds(), "error_kind", providerErrorKind(err))
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
		ID    string          `json:"id"`
		Usage json.RawMessage `json:"usage"`
	}
	if json.Unmarshal(body, &response) != nil || len(response.Choices) != 1 {
		return Output{SourceText: source}, errors.New("invalid AI response envelope")
	}
	choice := response.Choices[0]
	modelUsage := modelUsage(response.ID, response.Usage, choice.FinishReason, time.Since(modelStarted))
	modelUsage["model"] = c.config.Model
	if choice.FinishReason == "length" {
		slog.Warn("AI provider response truncated", "reason_code", "output_token_limit", "max_tokens", modelMaxTokens, "response_bytes", len(choice.Message.Content), "prompt_tokens", modelUsageInt(modelUsage, "prompt_tokens"), "completion_tokens", modelUsageInt(modelUsage, "completion_tokens"))
		return Output{SourceText: source, AttachmentTexts: attachmentTexts, OCRComplete: ocrComplete}, errors.New("AI response exceeded output token limit")
	}
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
	if c.config.StoreUsage {
		out.ModelUsage = modelUsage
	}
	if err != nil {
		var mismatch *SchemaMismatchError
		if errors.As(err, &mismatch) {
			return out, wrapProviderError("schema", "schema_mismatch", err)
		}
		return out, wrapProviderError("schema", "response_invalid", err)
	}
	slog.Info("AI provider extraction completed",
		"model", c.config.Model,
		"latency_ms", time.Since(modelStarted).Milliseconds(),
		"ocr_attachments", len(attachmentTexts),
		"draft_count", len(out.Drafts),
		"reply_bytes", len(out.Reply),
		"prompt_tokens", modelUsageInt(modelUsage, "prompt_tokens"),
		"completion_tokens", modelUsageInt(modelUsage, "completion_tokens"),
		"total_tokens", modelUsageInt(modelUsage, "total_tokens"),
	)
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
	walletAssignmentPolicy := "no_wallets; leave wallet_id empty and return drafts with a question to create/select a wallet"
	if len(wallets) > 1 {
		walletAssignmentPolicy = "choose_the_best_supplied_wallet; when uncertain default to wallet_id " + wallets[0].ID + "; return drafts with a review question instead of asking first"
	}
	if len(wallets) == 1 {
		walletAssignmentPolicy = "single_wallet; use wallet_id " + wallets[0].ID + " for every recognizable entry; do not ask for account-to-wallet mapping"
	}
	instruction := strings.TrimSpace(input.Instruction)
	if instruction == "" {
		instruction = strings.TrimSpace(input.Text)
	}
	content, err := json.Marshal(struct {
		Timezone               string     `json:"timezone"`
		Now                    time.Time  `json:"now"`
		UserInstruction        string     `json:"user_instruction,omitempty"`
		WalletAssignmentPolicy string     `json:"wallet_assignment_policy"`
		Wallets                []wallet   `json:"wallets"`
		Categories             []category `json:"categories"`
		Source                 string     `json:"untrusted_source_text"`
	}{Timezone: input.Timezone, Now: input.Now, UserInstruction: instruction, WalletAssignmentPolicy: walletAssignmentPolicy, Wallets: wallets, Categories: categories, Source: source})
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
