package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"mypocket/internal/agent"
)

const maxResponseBytes = 2 << 20

type Client struct {
	baseURL, apiKey string
	httpClient      *http.Client
}

func New(baseURL, apiKey string, timeout time.Duration) (*Client, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || u.Scheme == "" || u.Host == "" || strings.TrimSpace(apiKey) == "" || timeout <= 0 {
		return nil, errors.New("invalid OCR provider configuration")
	}
	return &Client{baseURL: u.String(), apiKey: apiKey, httpClient: &http.Client{Timeout: timeout}}, nil
}

func (c *Client) Capabilities(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/ocr/capabilities", nil)
	var response struct {
		Engine             string   `json:"engine"`
		CapabilityVersion  string   `json:"capabilityVersion"`
		SupportedLanguages []string `json:"supportedLanguages"`
		Limits             struct {
			MaxBase64Bytes     int64 `json:"maxBase64Bytes"`
			MaxLanguagesPerDoc int   `json:"maxLanguagesPerDoc"`
		} `json:"limits"`
	}
	if _, err := c.do(req, &response); err != nil {
		return err
	}
	if response.Engine != "OCR" || !strings.HasPrefix(response.CapabilityVersion, "ocr-v1.") || response.Limits.MaxBase64Bytes <= 0 || response.Limits.MaxLanguagesPerDoc <= 0 || !contains(response.SupportedLanguages, "vi-VN") {
		return errors.New("OCR provider capabilities are incompatible")
	}
	return nil
}

func (c *Client) Submit(ctx context.Context, input agent.ImageInput) (agent.ToolSubmission, error) {
	request := map[string]any{"input": map[string]string{"base64": base64.StdEncoding.EncodeToString(input.Bytes)}}
	if languages := normalizeLanguages(input.Languages); len(languages) > 0 {
		request["options"] = map[string]any{"languages": languages}
	}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/documents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	var response struct {
		ID     string              `json:"documentId"`
		Status agent.ToolRunStatus `json:"status"`
	}
	retry, err := c.do(req, &response)
	if err != nil {
		return agent.ToolSubmission{}, err
	}
	if response.ID == "" || !validOCRStatus(response.Status, true) {
		return agent.ToolSubmission{}, errors.New("invalid OCR submission response")
	}
	return agent.ToolSubmission{ProviderID: response.ID, Status: response.Status, RetryAfter: retry}, nil
}

func (c *Client) Read(ctx context.Context, id string) (agent.ToolResult, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/documents/"+url.PathEscape(id), nil)
	var response struct {
		DocumentID    string              `json:"documentId"`
		Status        agent.ToolRunStatus `json:"status"`
		Result        *ocrResult          `json:"result"`
		ExpiresAt     time.Time           `json:"resultExpiresAt"`
		ResultExpired bool                `json:"resultExpired"`
		ErrorDetail   string              `json:"errorDetail"`
	}
	retry, err := c.do(req, &response)
	if err != nil {
		return agent.ToolResult{}, err
	}
	if !validOCRStatus(response.Status, false) {
		return agent.ToolResult{}, errors.New("invalid OCR status")
	}
	if response.ResultExpired {
		response.Status = agent.ToolExpired
	}
	text := ""
	fields := map[string]any{"document_id": response.DocumentID}
	if response.Result != nil {
		text = response.Result.Text
		fields["page_count"] = response.Result.PageCount
		if len(response.Result.Pages) > 0 {
			fields["pages"] = response.Result.Pages
		}
	}
	return agent.ToolResult{Status: response.Status, Text: text, Fields: fields, ExpiresAt: response.ExpiresAt, RetryAfter: retry, ErrorCode: response.ErrorDetail}, nil
}

type ocrResult struct {
	Text      string           `json:"text"`
	PageCount int              `json:"pageCount"`
	Pages     []map[string]any `json:"pages"`
}

func (c *Client) do(req *http.Request, target any) (time.Duration, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, errors.New("OCR provider request failed")
	}
	defer resp.Body.Close()
	retry := parseRetry(resp.Header.Get("Retry-After"))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 404 {
			return retry, agent.ErrToolNotFound
		}
		if resp.StatusCode == 410 {
			return retry, agent.ErrToolExpired
		}
		if resp.StatusCode == 429 {
			return retry, agent.ErrToolQuota
		}
		return retry, fmt.Errorf("OCR provider returned status %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil || len(raw) > maxResponseBytes {
		return retry, errors.New("OCR provider response unavailable")
	}
	if json.Unmarshal(raw, target) != nil {
		return retry, errors.New("invalid OCR provider response")
	}
	return retry, nil
}
func parseRetry(v string) time.Duration {
	n, _ := strconv.Atoi(strings.TrimSpace(v))
	if n <= 0 {
		return 5 * time.Second
	}
	return time.Duration(n) * time.Second
}
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func normalizeLanguages(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		switch strings.TrimSpace(v) {
		case "vi":
			out = append(out, "vi-VN")
		case "en":
			out = append(out, "en-US")
		case "":
			continue
		default:
			out = append(out, strings.TrimSpace(v))
		}
	}
	return out
}

func validOCRStatus(status agent.ToolRunStatus, submission bool) bool {
	switch status {
	case agent.ToolQueued, agent.ToolProcessing, agent.ToolCompleted:
		return true
	case agent.ToolFailed, agent.ToolCancelled, agent.ToolExpired:
		return !submission
	default:
		return false
	}
}
