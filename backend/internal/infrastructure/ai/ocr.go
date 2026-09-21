package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxImages      = 3
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
		if (img.MIMEType != "image/jpeg" && img.MIMEType != "image/png") || len(img.Name) > 255 {
			return errors.New("receipt image must be JPEG or PNG with a bounded filename")
		}
		if len(img.Base64) > base64.StdEncoding.EncodedLen(maxImageBytes) {
			return errors.New("receipt image exceeds 5 MiB")
		}
		data, err := base64.StdEncoding.Strict().DecodeString(img.Base64)
		if err != nil || len(data) == 0 || len(data) > maxImageBytes {
			return errors.New("invalid receipt image encoding or size")
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || "image/"+format != img.MIMEType || config.Width <= 0 || config.Height <= 0 || int64(config.Width) > maxImagePixels/int64(config.Height) {
			return errors.New("receipt image content, MIME type, or dimensions are invalid")
		}
	}
	return nil
}

func (c *Client) ocr(ctx context.Context, img Image) (string, error) {
	body, _, err := c.request(ctx, http.MethodPost, c.config.OCRURL+"/v1/documents", c.config.OCRKey, map[string]any{"input": map[string]string{"base64": img.Base64}}, "OCR")
	if err != nil {
		return "", err
	}
	var submitted struct {
		DocumentID string `json:"documentId"`
	}
	if json.Unmarshal(body, &submitted) != nil || !validDocumentID(submitted.DocumentID) {
		return "", errors.New("invalid OCR submission response")
	}
	for attempt := 0; attempt < maxOCRPolls; attempt++ {
		body, headers, err := c.request(ctx, http.MethodGet, c.config.OCRURL+"/v1/documents/"+submitted.DocumentID, c.config.OCRKey, nil, "OCR")
		if err != nil {
			return "", err
		}
		var document struct {
			Status          string `json:"status"`
			ResultExpiresAt string `json:"resultExpiresAt"`
			Result          *struct {
				Text string `json:"text"`
			} `json:"result"`
		}
		if json.Unmarshal(body, &document) != nil {
			return "", errors.New("invalid OCR document response")
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
			return "", errors.New("OCR recognition failed")
		case "cancelled", "canceled":
			return "", errors.New("OCR recognition was cancelled")
		case "expired":
			return "", errors.New("OCR result expired")
		case "queued", "processing":
			if attempt == maxOCRPolls-1 {
				break
			}
			if err := waitForPoll(ctx, headers.Get("Retry-After")); err != nil {
				return "", err
			}
		default:
			return "", errors.New("OCR returned an unsupported document status")
		}
	}
	return "", errors.New("OCR polling limit exceeded")
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
