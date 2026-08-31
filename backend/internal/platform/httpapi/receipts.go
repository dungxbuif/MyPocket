package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/platform/config"
)

const maxReceiptBytes int64 = 15 * 1024 * 1024

type filePresignRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	ChecksumSHA string `json:"checksum_sha256"`
}

type filePresignResponse struct {
	Status        string                `json:"status"`
	File          finance.ReceiptObject `json:"file"`
	UploadURL     string                `json:"upload_url"`
	ExpiresInSecs int                   `json:"expires_in_seconds"`
	CorrelationID string                `json:"correlation_id"`
}

type fileDownloadResponse struct {
	Status        string                `json:"status"`
	File          finance.ReceiptObject `json:"file"`
	DownloadURL   string                `json:"download_url"`
	ExpiresInSecs int                   `json:"expires_in_seconds"`
	CorrelationID string                `json:"correlation_id"`
}

func filePresign(cfg config.Config, repo ReceiptRepository, store ObjectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if repo == nil || store == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Receipt storage unavailable", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var input filePresignRequest
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024)).Decode(&input); err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
				return
			}
			input.Filename = filepath.Base(strings.TrimSpace(input.Filename))
			input.ContentType = strings.ToLower(strings.TrimSpace(input.ContentType))
			if !allowedReceiptContentType(input.ContentType) || input.SizeBytes <= 0 || input.SizeBytes > maxReceiptBytes || input.Filename == "." || input.Filename == "" {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("UPLOAD_REJECTED", "Unsupported receipt file", correlationID(r.Context())))
				return
			}
			key, err := receiptObjectKey(userID, input.Filename)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, ErrorEnvelope("INTERNAL_FAILURE", "Could not create receipt", correlationID(r.Context())))
				return
			}
			receipt, err := repo.CreateReceiptObject(r.Context(), userID, finance.CreateReceiptObjectInput{ObjectKey: key, ContentType: input.ContentType, SizeBytes: input.SizeBytes, ChecksumSHA256: strings.TrimSpace(input.ChecksumSHA), OriginalFilename: input.Filename})
			if err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("UPLOAD_REJECTED", "Invalid receipt metadata", correlationID(r.Context())))
				return
			}
			url, err := store.PresignPut(r.Context(), key, input.ContentType, 15*time.Minute)
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Receipt storage unavailable", correlationID(r.Context())))
				return
			}
			writeJSON(w, http.StatusCreated, filePresignResponse{Status: "ok", File: receipt, UploadURL: url, ExpiresInSecs: 900, CorrelationID: correlationID(r.Context())})
		})).ServeHTTP(w, r)
	}
}

func fileDownload(cfg config.Config, repo ReceiptRepository, store ObjectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if r.Method != http.MethodGet || repo == nil || store == nil {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")
		id = strings.TrimSuffix(id, "/download")
		if id == "" || strings.Contains(id, "/") {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Receipt not found", correlationID(r.Context())))
			return
		}
		receipt, err := repo.GetReceiptObject(r.Context(), userID, id)
		if err != nil {
			status := http.StatusServiceUnavailable
			if errors.Is(err, finance.ErrForbidden) {
				status = http.StatusNotFound
			}
			writeJSON(w, status, ErrorEnvelope("NOT_FOUND", "Receipt not found", correlationID(r.Context())))
			return
		}
		url, err := store.PresignGet(r.Context(), receipt.ObjectKey, 10*time.Minute)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Receipt storage unavailable", correlationID(r.Context())))
			return
		}
		writeJSON(w, http.StatusOK, fileDownloadResponse{Status: "ok", File: receipt, DownloadURL: url, ExpiresInSecs: 600, CorrelationID: correlationID(r.Context())})
	}
}

func allowedReceiptContentType(value string) bool {
	switch value {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func receiptObjectKey(userID, filename string) (string, error) {
	var suffix [12]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return "users/" + userID + "/receipts/" + hex.EncodeToString(suffix[:]) + "-" + filename, nil
}
