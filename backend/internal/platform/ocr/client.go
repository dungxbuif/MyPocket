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
		API   string   `json:"api"`
		Input []string `json:"input"`
	}
	if _, err := c.do(req, &response); err != nil {
		return err
	}
	if response.API != "v1" || !contains(response.Input, "base64") {
		return errors.New("OCR provider capabilities are incompatible")
	}
	return nil
}

func (c *Client) Submit(ctx context.Context, input agent.ImageInput) (agent.ToolSubmission, error) {
	body, _ := json.Marshal(map[string]any{"content_type": input.ContentType, "content_base64": base64.StdEncoding.EncodeToString(input.Bytes), "languages": input.Languages})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/documents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	var response struct {
		ID     string              `json:"document_id"`
		Status agent.ToolRunStatus `json:"status"`
	}
	retry, err := c.do(req, &response)
	if err != nil {
		return agent.ToolSubmission{}, err
	}
	if response.ID == "" || (response.Status != agent.ToolProcessing && response.Status != agent.ToolCompleted) {
		return agent.ToolSubmission{}, errors.New("invalid OCR submission response")
	}
	return agent.ToolSubmission{ProviderID: response.ID, Status: response.Status, RetryAfter: retry}, nil
}

func (c *Client) Read(ctx context.Context, id string) (agent.ToolResult, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/documents/"+url.PathEscape(id), nil)
	var response struct {
		Status    agent.ToolRunStatus `json:"status"`
		Text      string              `json:"text"`
		Fields    map[string]any      `json:"fields"`
		ExpiresAt time.Time           `json:"expires_at"`
		ErrorCode string              `json:"error_code"`
	}
	retry, err := c.do(req, &response)
	if err != nil {
		return agent.ToolResult{}, err
	}
	switch response.Status {
	case agent.ToolProcessing, agent.ToolCompleted, agent.ToolFailed, agent.ToolCancelled, agent.ToolExpired:
	default:
		return agent.ToolResult{}, errors.New("invalid OCR status")
	}
	return agent.ToolResult{Status: response.Status, Text: response.Text, Fields: response.Fields, ExpiresAt: response.ExpiresAt, RetryAfter: retry, ErrorCode: response.ErrorCode}, nil
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
