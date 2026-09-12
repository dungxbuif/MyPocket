package ocr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/agent"
)

func TestOCRClientContract(t *testing.T) {
	expiresAt := "2026-09-12T10:00:00Z"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer hidden-key" {
			t.Fatalf("bad auth")
		}
		switch r.URL.Path {
		case "/v1/ocr/capabilities":
			io.WriteString(w, `{"engine":"OCR","capabilityVersion":"ocr-v1.0","defaultProfile":{"recognitionLevel":"accurate","languages":["vi-VN","en-US"],"automaticallyDetectsLanguage":true,"usesLanguageCorrection":true},"supportedLevels":["accurate","fast"],"supportedRevisions":[1,2,3],"supportedLanguages":["vi-VN","en-US"],"limits":{"maxBase64Bytes":26214400,"maxBatchItems":100,"maxLanguagesPerDoc":10,"maxCustomWords":100}}`)
		case "/v1/documents":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method %s", r.Method)
			}
			var payload struct {
				Input struct {
					Base64 string `json:"base64"`
				} `json:"input"`
				Options struct {
					Languages []string `json:"languages"`
				} `json:"options"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("invalid submit payload: %v", err)
			}
			if payload.Input.Base64 != base64.StdEncoding.EncodeToString([]byte("image")) {
				t.Fatalf("unexpected base64 payload %#v", payload.Input.Base64)
			}
			if strings.Join(payload.Options.Languages, ",") != "vi-VN,en-US" {
				t.Fatalf("unexpected languages %#v", payload.Options.Languages)
			}
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(202)
			io.WriteString(w, `{"documentId":"doc_1","status":"queued","createdAt":"2026-09-12T09:59:00Z"}`)
		case "/v1/documents/doc_1":
			io.WriteString(w, `{"documentId":"doc_1","status":"completed","result":{"text":"Total 120000","pageCount":1,"pages":[{"pageNumber":1,"text":"Total 120000"}]},"resultExpiresAt":"`+expiresAt+`"}`)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	client, err := New(server.URL, "hidden-key", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.Capabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
	sub, err := client.Submit(context.Background(), agent.ImageInput{ContentType: "image/jpeg", Bytes: []byte("image"), Languages: []string{"vi", "en"}})
	if err != nil || sub.ProviderID != "doc_1" || sub.Status != agent.ToolQueued || sub.RetryAfter != 2*time.Second {
		t.Fatalf("sub=%#v err=%v", sub, err)
	}
	result, err := client.Read(context.Background(), sub.ProviderID)
	if err != nil || result.Status != agent.ToolCompleted || result.Text == "" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if result.ExpiresAt.Format(time.RFC3339) != expiresAt {
		t.Fatalf("expires_at=%s", result.ExpiresAt.Format(time.RFC3339))
	}
	if result.Fields["page_count"] != 1 {
		t.Fatalf("fields=%#v", result.Fields)
	}
}

func TestOCRClientErrorsNeverLeakSecretOrBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(429)
		io.WriteString(w, "hidden-key private body")
	}))
	defer server.Close()
	client, _ := New(server.URL, "hidden-key", time.Second)
	_, err := client.Read(context.Background(), "x")
	if err == nil || strings.Contains(err.Error(), "hidden-key") || strings.Contains(err.Error(), "private body") {
		t.Fatalf("unsafe error %v", err)
	}
}

func TestOCRCapabilitiesDriftFailsClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, `{"engine":"OCR","capabilityVersion":"ocr-v1.0","supportedLanguages":["en-US"],"limits":{"maxBase64Bytes":0}}`)
	}))
	defer server.Close()
	client, _ := New(server.URL, "key", time.Second)
	if client.Capabilities(context.Background()) == nil {
		t.Fatal("expected incompatible capabilities")
	}
}
