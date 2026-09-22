package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	maxImages      = 20
	maxImageBytes  = 5 * 1024 * 1024
	maxImagePixels = 25_000_000
	maxOCRPolls    = 60
	pollInterval   = time.Second
)

func validateImages(images []Image) error {
	if len(images) > maxImages {
		return errors.New("at most 3 receipt images are allowed")
	}
	for _, img := range images {
		if (img.MIMEType != "image/jpeg" && img.MIMEType != "image/png" && img.MIMEType != "application/pdf") || len(img.Name) > 255 {
			return errors.New("receipt file must be JPEG, PNG, or PDF with a bounded filename")
		}
		if img.SourceURL != "" {
			if img.Base64 != "" || !validOCRSourceURL(img.SourceURL) {
				return errors.New("invalid private OCR source")
			}
			continue
		}
		if len(img.Base64) > base64.StdEncoding.EncodedLen(maxImageBytes) {
			return errors.New("receipt image exceeds 5 MiB")
		}
		data, err := base64.StdEncoding.Strict().DecodeString(img.Base64)
		if err != nil || len(data) == 0 || len(data) > maxImageBytes {
			return errors.New("invalid receipt file encoding or size")
		}
		if img.MIMEType == "application/pdf" {
			if len(data) < 8 || string(data[:5]) != "%PDF-" {
				return errors.New("receipt PDF content does not match its MIME type")
			}
			continue
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || "image/"+format != img.MIMEType || config.Width <= 0 || config.Height <= 0 || int64(config.Width) > maxImagePixels/int64(config.Height) {
			return errors.New("receipt image content, MIME type, or dimensions are invalid")
		}
	}
	return nil
}

func (c *Client) ocr(ctx context.Context, img Image) (string, error) {
	var body []byte
	var err error
	if img.SourceURL != "" {
		body, _, err = c.request(ctx, http.MethodPost, c.config.OCRURL+"/v1/documents", c.config.OCRKey, map[string]any{"input": map[string]string{"url": img.SourceURL}}, "OCR")
	} else if img.MIMEType == "application/pdf" {
		body, err = c.uploadPrivatePDF(ctx, img)
	} else {
		body, _, err = c.request(ctx, http.MethodPost, c.config.OCRURL+"/v1/documents", c.config.OCRKey, map[string]any{"input": map[string]string{"base64": img.Base64}}, "OCR")
	}
	if err != nil {
		return "", wrapProviderError("ocr_submit", "ocr_submit_request", err)
	}
	var submitted struct {
		DocumentID string `json:"documentId"`
	}
	if json.Unmarshal(body, &submitted) != nil || !validDocumentID(submitted.DocumentID) {
		return "", wrapProviderError("ocr_submit", "ocr_submission_invalid", errors.New("invalid OCR submission response"))
	}
	for attempt := 0; attempt < maxOCRPolls; attempt++ {
		body, headers, err := c.request(ctx, http.MethodGet, c.config.OCRURL+"/v1/documents/"+submitted.DocumentID, c.config.OCRKey, nil, "OCR")
		if err != nil {
			return "", wrapProviderError("ocr_poll", "ocr_poll_request", err)
		}
		var document struct {
			Status          string `json:"status"`
			ResultExpiresAt string `json:"resultExpiresAt"`
			Result          *struct {
				Text string `json:"text"`
			} `json:"result"`
		}
		if json.Unmarshal(body, &document) != nil {
			return "", wrapProviderError("ocr_poll", "ocr_response_invalid", errors.New("invalid OCR document response"))
		}
		switch document.Status {
		case "completed":
			if document.ResultExpiresAt != "" {
				expires, err := time.Parse(time.RFC3339, document.ResultExpiresAt)
				if err != nil || !expires.After(time.Now()) {
					return "", errors.New("OCR result expired or has an invalid expiry")
				}
			}
			if document.Result == nil || strings.TrimSpace(document.Result.Text) == "" || len(document.Result.Text) > maxSourceBytes {
				return "", errors.New("OCR result text is missing or exceeds limit")
			}
			return document.Result.Text, nil
		case "failed":
			return "", wrapProviderError("ocr_poll", "ocr_failed", errors.New("OCR recognition failed"))
		case "cancelled", "canceled":
			return "", wrapProviderError("ocr_poll", "ocr_cancelled", errors.New("OCR recognition was cancelled"))
		case "expired":
			return "", wrapProviderError("ocr_poll", "ocr_expired", errors.New("OCR result expired"))
		case "queued", "processing":
			if attempt == maxOCRPolls-1 {
				break
			}
			if err := waitForPoll(ctx, headers.Get("Retry-After")); err != nil {
				return "", err
			}
		default:
			return "", wrapProviderError("ocr_poll", "ocr_status_unsupported", errors.New("OCR returned an unsupported document status"))
		}
	}
	return "", wrapProviderError("ocr_poll", "ocr_poll_limit", errors.New("OCR polling limit exceeded"))
}

func validOCRSourceURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && u.Fragment == "" && u.RawQuery != ""
}

func (c *Client) uploadPrivatePDF(ctx context.Context, file Image) ([]byte, error) {
	data, err := base64.StdEncoding.Strict().DecodeString(file.Base64)
	if err != nil {
		return nil, errors.New("invalid PDF encoding")
	}
	presigned, _, err := c.request(ctx, http.MethodPost, c.config.OCRURL+"/v1/uploads/presign", c.config.OCRKey, map[string]any{
		"filename": file.Name, "sizeBytes": len(data), "contentType": "application/pdf",
	}, "OCR")
	if err != nil {
		return nil, err
	}
	var target struct {
		Method    string            `json:"method"`
		UploadURL string            `json:"uploadUrl"`
		SourceURL string            `json:"sourceUrl"`
		Headers   map[string]string `json:"headers"`
	}
	if json.Unmarshal(presigned, &target) != nil || target.Method != http.MethodPut || !validOCRUploadURL(target.UploadURL) || !validURL(target.SourceURL) {
		return nil, errors.New("invalid OCR upload grant")
	}
	contentLength, lengthOK := target.Headers["Content-Length"]
	contentType, typeOK := target.Headers["Content-Type"]
	if !lengthOK || contentLength != fmt.Sprint(len(data)) || !typeOK || contentType != "application/pdf" {
		return nil, errors.New("invalid OCR upload headers")
	}
	upload, err := http.NewRequestWithContext(ctx, http.MethodPut, target.UploadURL, bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("invalid OCR upload URL")
	}
	upload.Header.Set("Content-Type", contentType)
	upload.ContentLength = int64(len(data))
	response, err := c.http.Do(upload)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("OCR file upload failed")
	}
	io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("OCR file upload returned HTTP %d", response.StatusCode)
	}
	body, _, err := c.request(ctx, http.MethodPost, c.config.OCRURL+"/v1/documents", c.config.OCRKey, map[string]any{"input": map[string]string{"url": target.SourceURL}}, "OCR")
	return body, err
}

func validOCRUploadURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.Fragment != "" || (u.Scheme != "https" && u.Scheme != "http") {
		return false
	}
	// Permit plain HTTP only for local test servers; real pre-signed uploads must use TLS.
	return u.Scheme == "https" || u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost" || u.Hostname() == "::1"
}

func validDocumentID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func waitForPoll(ctx context.Context, retryAfter string) error {
	delay := pollInterval
	if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil && seconds > 0 {
		// No poll can fit after the overall deadline if this delay exceeds 90s.
		if seconds > 90 {
			seconds = 90
		}
		delay = time.Duration(seconds) * time.Second
	} else if date, err := http.ParseTime(retryAfter); err == nil && time.Until(date) > delay {
		delay = time.Until(date)
		if delay > extractTimeout {
			delay = extractTimeout
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
