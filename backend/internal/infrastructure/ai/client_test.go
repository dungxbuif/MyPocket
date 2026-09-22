package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

const validDraft = `{"type":"transfer","amount":9007199254740991,"wallet_id":"unverified-id","category_id":null,"occurred_at":"","note":"","included_in_reports":false,"questions":["Which destination wallet?"]}`

func modelResponse(w http.ResponseWriter, content string) {
	_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": content}, "finish_reason": "stop"}}})
}

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)
	return NewClient(Config{BaseURL: s.URL + "/v1/", APIKey: "test-model-key", Model: "test-model", OCRURL: s.URL, OCRKey: "test-ocr-key"})
}

func pngImage(t *testing.T) Image {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return Image{Name: "receipt.png", MIMEType: "image/png", Base64: base64.StdEncoding.EncodeToString(buf.Bytes())}
}

func TestConfigured(t *testing.T) {
	for _, cfg := range []Config{{}, {BaseURL: "http://localhost/v1", Model: "m"}, {BaseURL: "ftp://host/v1", APIKey: "test", Model: "m"}, {BaseURL: "https://user:pass@host/v1", APIKey: "test", Model: "m"}} {
		c := NewClient(cfg)
		if c.Configured() {
			t.Fatal("invalid configuration accepted")
		}
		if _, err := c.Extract(context.Background(), Input{Text: "coffee"}); !errors.Is(err, ErrNotConfigured) {
			t.Fatalf("error = %v", err)
		}
	}
	c := NewClient(Config{BaseURL: "https://example.invalid/v1", APIKey: "test", Model: "m"})
	if !c.Configured() || c.OCRConfigured() {
		t.Fatal("incorrect capabilities")
	}
	c = NewClient(Config{OCRURL: "https://example.invalid", OCRKey: "test"})
	if !c.OCRConfigured() || c.Configured() {
		t.Fatal("OCR independent of model")
	}
}

func TestTextOnlyRequestAndUnverifiedIDs(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-model-key" {
			t.Errorf("incorrect model request")
		}
		var req map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		if _, ok := req["tools"]; ok {
			t.Error("must not supply tools")
		}
		var responseFormat struct {
			Type       string `json:"type"`
			JSONSchema struct {
				Name   string         `json:"name"`
				Strict bool           `json:"strict"`
				Schema map[string]any `json:"schema"`
			} `json:"json_schema"`
		}
		if err := json.Unmarshal(req["response_format"], &responseFormat); err != nil {
			t.Errorf("decode response format: %v", err)
		}
		if responseFormat.Type != "json_schema" || responseFormat.JSONSchema.Name != "transaction_proposal" || !responseFormat.JSONSchema.Strict {
			t.Errorf("response format must use strict OpenAI-compatible JSON Schema: %+v", responseFormat)
		}
		if properties, ok := responseFormat.JSONSchema.Schema["properties"].(map[string]any); !ok || properties["drafts"] == nil {
			t.Errorf("schema missing drafts property: %+v", responseFormat.JSONSchema.Schema)
		}
		var messages []chatMessage
		var wireMessages []map[string]json.RawMessage
		if err := json.Unmarshal(req["messages"], &wireMessages); err != nil {
			t.Error(err)
		}
		for _, message := range wireMessages {
			if _, ok := message["role"]; !ok {
				t.Error("provider message missing lowercase role")
			}
			if _, ok := message["content"]; !ok {
				t.Error("provider message missing lowercase content")
			}
		}
		if err := json.Unmarshal(req["messages"], &messages); err != nil {
			t.Error(err)
		}
		if len(messages) != 2 || messages[0].Role != "system" {
			t.Errorf("one-shot request must have only system and input messages: %d", len(messages))
		}
		last := messages[len(messages)-1].Content
		if !strings.Contains(last, "coffee") || !strings.Contains(last, "wallet-1") || !strings.Contains(last, `"description":"Daily cash for food and transit"`) || strings.Contains(last, "private-owner") {
			t.Errorf("incorrect catalog projection")
		}
		if !strings.Contains(messages[0].Content, "descriptions") || !strings.Contains(messages[0].Content, "untrusted data") {
			t.Errorf("system prompt must treat wallet descriptions as untrusted reference data")
		}
		modelResponse(w, `{"reply":"Please clarify","drafts":[`+validDraft+`]}`)
	})
	description := "Daily cash for food and transit"
	out, err := c.Extract(context.Background(), Input{Text: "coffee", Timezone: "Asia/Ho_Chi_Minh", Now: time.Now(), Wallets: []entity.Wallet{{ID: "wallet-1", OwnerID: "private-owner", Name: "Cash", Description: &description}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Drafts) != 1 || out.Drafts[0].WalletID != "unverified-id" || out.Drafts[0].Type != "transfer" || out.Drafts[0].Amount != 9007199254740991 || out.SourceText != "coffee" {
		t.Fatalf("wrong output: %+v", out)
	}
}

func TestBuildMessagesRejectsOversizedWalletDescription(t *testing.T) {
	description := strings.Repeat("x", 2049)
	_, err := buildMessages(Input{Text: "test", Wallets: []entity.Wallet{{ID: "wallet-1", Name: "Cash", Description: &description}}}, "test")
	if err == nil {
		t.Fatal("oversized wallet descriptions must be rejected before provider submission")
	}
}

func TestSchemaValidation(t *testing.T) {
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	cases := []string{`null`, `[]`, `{}`, `{"reply":null,"drafts":[]}`, `{"reply":"?","drafts":null}`, `{"reply":"?","drafts":[],"tool":"write"}`, `{"reply":"?","drafts":[]} trailing`, `{"reply":"?","reply":"duplicate","drafts":[]}`, `{"reply":"?","drafts":[{}]}`,
		`{"reply":"?","drafts":[` + strings.Replace(validDraft, "9007199254740991", "9007199254740992", 1) + `]}`,
		`{"reply":"?","drafts":[` + strings.Replace(validDraft, "9007199254740991", "-1", 1) + `]}`,
		`{"reply":"?","drafts":[` + strings.Replace(validDraft, "9007199254740991", "1.5", 1) + `]}`,
		`{"reply":"?","drafts":[` + strings.Replace(validDraft, `"included_in_reports":false`, `"included_in_reports":null`, 1) + `]}`,
		`{"reply":"?","drafts":[` + strings.TrimSuffix(strings.Repeat(validDraft+",", 31), ",") + `]}`,
	}
	for i, content := range cases {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) { modelResponse(w, content) })
			if _, err := c.Extract(context.Background(), Input{Text: "test"}); err == nil {
				t.Fatal("malformed output accepted")
			}
		})
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		modelResponse(w, `{"reply":"Which wallet?","drafts":[]}`)
	})
	out, err := c.Extract(context.Background(), Input{Text: "test"})
	if err != nil || out.Drafts == nil || len(out.Drafts) != 0 {
		t.Fatalf("clarification rejected: %v", err)
	}
}

func TestSchemaMismatchLogReportsStructureWithoutLoggingModelContent(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	privateText := "private transaction narrative sentinel"
	privateValue := "private amount value sentinel"
	content, err := json.Marshal(map[string]any{
		"reply": privateText,
		"drafts": []any{map[string]any{
			"type":                   "expense",
			"amount":                 privateValue,
			"private_account_123456": "must not log values or model-provided identifier keys",
			privateText:              "must not log model-provided key names",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { modelResponse(w, string(content)) })
	if _, err := c.Extract(context.Background(), Input{Text: "not-for-logs", Timezone: "Asia/Ho_Chi_Minh"}); err == nil {
		t.Fatal("schema-invalid output was accepted")
	} else if err.Error() != errSchema.Error() {
		t.Fatalf("client-facing schema error must stay generic, got %q", err)
	}
	output := logs.String()
	for _, want := range []string{"AI provider response rejected", "schema_mismatch", "issue_0_code", "object_fields", "issue_0_path", "$.drafts[0]", "issue_0_expected", "issue_0_actual", "amount", "<unrecognized_field>", "test-model", "finish_reason", "stop", "response_bytes", "latency_ms"} {
		if !strings.Contains(output, want) {
			t.Errorf("diagnostic log missing %q: %s", want, output)
		}
	}
	for _, secret := range []string{privateText, privateValue, "private_account_123456", "must not log values", "not-for-logs"} {
		if strings.Contains(output, secret) {
			t.Errorf("diagnostic log exposed model/input content %q", secret)
		}
	}
}

func TestSchemaDiagnosticShowsArrayElementShapeWithoutValues(t *testing.T) {
	_, err := parseOutput([]byte(`[{"type":"expense","amount":"private amount","note":"private note","category":"food"}]`))
	var mismatch *SchemaMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("expected a typed schema mismatch, got %v", err)
	}
	serialized, marshalErr := json.Marshal(mismatch.Issues)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	shape := string(serialized)
	for _, want := range []string{"root_type", "array_item_shape", "unrecognized_field", "amount,note,type"} {
		if !strings.Contains(shape, want) {
			t.Errorf("schema diagnostic missing %q: %s", want, shape)
		}
	}
	for _, private := range []string{"private amount", "private note"} {
		if strings.Contains(shape, private) {
			t.Errorf("schema diagnostic exposed value %q", private)
		}
	}
}

func TestOCRBeforeModel(t *testing.T) {
	var submits, polls, models atomic.Int32
	img := pngImage(t)
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/documents":
			submits.Add(1)
			if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-ocr-key" {
				t.Error("incorrect OCR submission")
			}
			var body struct {
				Input struct {
					Base64 string `json:"base64"`
				} `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Input.Base64 != img.Base64 {
				t.Error("missing OCR bytes")
			}
			w.WriteHeader(http.StatusAccepted)
			_, _ = io.WriteString(w, `{"documentId":"doc_123","status":"queued"}`)
		case "/v1/documents/doc_123":
			polls.Add(1)
			if r.Method != "GET" {
				t.Error("incorrect OCR poll")
			}
			_, _ = io.WriteString(w, `{"status":"completed","result":{"text":"Receipt total 42000"}}`)
		case "/v1/chat/completions":
			models.Add(1)
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), img.Base64) || strings.Contains(string(body), "image_url") || !strings.Contains(string(body), "Receipt total 42000") {
				t.Error("model must receive only extracted text")
			}
			modelResponse(w, `{"reply":"Which wallet?","drafts":[]}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	})
	out, err := c.Extract(context.Background(), Input{Text: "lunch", Images: []Image{img}})
	if err != nil || !strings.Contains(out.SourceText, "Receipt total 42000") || submits.Load() != 1 || polls.Load() != 1 || models.Load() != 1 {
		t.Fatalf("OCR flow failed: %v %+v", err, out)
	}
}

func TestPDFIsUploadedToOCRPrivatelyBeforeTextOnlyModel(t *testing.T) {
	pdf := []byte("%PDF-1.7\nexample receipt bytes\n%%EOF")
	var ocrSubmissions, uploads, models atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/uploads/presign":
			if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer test-ocr-key" {
				t.Error("invalid OCR presign request")
			}
			var request struct {
				Filename    string `json:"filename"`
				SizeBytes   int    `json:"sizeBytes"`
				ContentType string `json:"contentType"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Filename != "statement.pdf" || request.SizeBytes != len(pdf) || request.ContentType != "application/pdf" {
				t.Errorf("invalid presign payload: %+v (%v)", request, err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"method": "PUT", "uploadUrl": "http://" + r.Host + "/upload-target", "sourceUrl": "https://ocr-source.example/doc.pdf", "headers": map[string]string{"Content-Length": fmt.Sprint(len(pdf)), "Content-Type": "application/pdf"}})
		case "/upload-target":
			uploads.Add(1)
			body, _ := io.ReadAll(r.Body)
			if r.Method != http.MethodPut || r.Header.Get("Content-Type") != "application/pdf" || !bytes.Equal(body, pdf) {
				t.Error("PDF bytes were not uploaded with the signed content type")
			}
			w.WriteHeader(http.StatusOK)
		case "/v1/documents":
			ocrSubmissions.Add(1)
			var request struct {
				Input struct {
					URL string `json:"url"`
				} `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Input.URL != "https://ocr-source.example/doc.pdf" {
				t.Error("OCR source URL was not submitted")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"documentId": "doc_pdf_1", "status": "queued"})
		case "/v1/documents/doc_pdf_1":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "completed", "result": map[string]string{"text": "Transfer 35000 VND"}})
		case "/v1/chat/completions":
			models.Add(1)
			body, _ := io.ReadAll(r.Body)
			if bytes.Contains(body, pdf) || strings.Contains(string(body), base64.StdEncoding.EncodeToString(pdf)) || strings.Contains(string(body), "ocr-source.example") || bytes.Contains(body, []byte("image_url")) || !bytes.Contains(body, []byte("Transfer 35000 VND")) {
				t.Error("LLM must receive OCR text only, never a file or its URL")
			}
			modelResponse(w, `{"reply":"Review","drafts":[]}`)
		default:
			t.Errorf("unexpected provider path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	c := NewClient(Config{BaseURL: server.URL + "/v1/", APIKey: "test-model-key", Model: "test-model", OCRURL: server.URL, OCRKey: "test-ocr-key"})
	out, err := c.Extract(context.Background(), Input{Images: []Image{{Name: "statement.pdf", MIMEType: "application/pdf", Base64: base64.StdEncoding.EncodeToString(pdf)}}})
	if err != nil || out.SourceText != "Transfer 35000 VND" || uploads.Load() != 1 || ocrSubmissions.Load() != 1 || models.Load() != 1 {
		t.Fatalf("PDF OCR flow failed: out=%+v err=%v upload=%d submit=%d model=%d", out, err, uploads.Load(), ocrSubmissions.Load(), models.Load())
	}
}

func TestPrivateStoredImageUsesSignedURLForOCRAndTextOnlyForModel(t *testing.T) {
	const signedURL = "https://private-storage.invalid/receipt.png?signature=temporary"
	var models atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/documents":
			var request struct {
				Input struct {
					URL string `json:"url"`
				} `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Input.URL != signedURL {
				t.Errorf("OCR must receive the private signed URL: %+v %v", request, err)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"documentId": "private_1", "status": "queued"})
		case "/v1/documents/private_1":
			_, _ = io.WriteString(w, `{"status":"completed","result":{"text":"OCR amount 42000"}}`)
		case "/v1/chat/completions":
			models.Add(1)
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "private-storage.invalid") || strings.Contains(string(body), "base64") || !strings.Contains(string(body), "OCR amount 42000") {
				t.Error("LLM boundary leaked file reference or omitted OCR text")
			}
			modelResponse(w, `{"reply":"review","drafts":[]}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	})
	_, err := c.Extract(context.Background(), Input{Text: "receipt", Timezone: "UTC", Images: []Image{{Name: "receipt.png", MIMEType: "image/png", SourceURL: signedURL}}})
	if err != nil || models.Load() != 1 {
		t.Fatalf("private OCR flow failed: %v models=%d", err, models.Load())
	}
}

func TestInvalidImagesFailBeforeSubmission(t *testing.T) {
	good := pngImage(t)
	wrong := good
	wrong.MIMEType = "image/jpeg"
	wrongPDF := good
	wrongPDF.MIMEType = "application/pdf"
	bad := good
	bad.Base64 = base64.StdEncoding.EncodeToString([]byte("not an image"))
	badPDF := Image{Name: "spoof.pdf", MIMEType: "application/pdf", Base64: base64.StdEncoding.EncodeToString([]byte("not a PDF"))}
	large := good
	large.Base64 = strings.Repeat("A", 7*1024*1024)
	largePDF := Image{Name: "large.pdf", MIMEType: "application/pdf", Base64: base64.StdEncoding.EncodeToString(append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), maxImageBytes)...))}
	pixels := good
	data, _ := base64.StdEncoding.DecodeString(good.Base64)
	binary.BigEndian.PutUint32(data[16:20], 100000)
	binary.BigEndian.PutUint32(data[20:24], 100000)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	pixels.Base64 = base64.StdEncoding.EncodeToString(data)
	for _, images := range [][]Image{{wrong}, {wrongPDF}, {bad}, {badPDF}, {large}, {largePDF}, {pixels}, {good, wrong}, {good, good, good, good}} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid image triggered provider call") })
		if _, err := c.Extract(context.Background(), Input{Images: images}); err == nil {
			t.Fatal("invalid image accepted")
		}
	}
}

func TestJPEGAndThreeImages(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatal(err)
	}
	img := Image{Name: "photo.jpg", MIMEType: "image/jpeg", Base64: base64.StdEncoding.EncodeToString(buf.Bytes())}
	var submissions atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/documents":
			submissions.Add(1)
			w.WriteHeader(202)
			_, _ = io.WriteString(w, `{"documentId":"doc_1"}`)
		case "/v1/documents/doc_1":
			_, _ = io.WriteString(w, `{"status":"completed","result":{"text":"receipt"}}`)
		case "/v1/chat/completions":
			modelResponse(w, `{"reply":"Review","drafts":[]}`)
		default:
			t.Error("unexpected path")
		}
	})
	out, err := c.Extract(context.Background(), Input{Images: []Image{img, img, img}})
	if err != nil || submissions.Load() != 3 || out.SourceText != "receipt\n\nreceipt\n\nreceipt" {
		t.Fatalf("three JPEG flow failed: %v", err)
	}
}

func TestProviderProtocolFailures(t *testing.T) {
	for _, response := range []string{`{}`, `{"choices":[]}`, `{"choices":[{"message":{"content":"{}"},"finish_reason":"length"}]}`, `{"choices":[{"message":{"content":"{}","tool_calls":[{}]},"finish_reason":"tool_calls"}]}`, strings.Repeat("x", 262145)} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, response) })
		if _, err := c.Extract(context.Background(), Input{Text: "test"}); err == nil {
			t.Fatal("invalid envelope accepted")
		}
	}
}

func TestOCRSubmissionNeverRetriedOrRedirected(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusBadGateway} {
		var calls atomic.Int32
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.Header().Set("Location", "/must-not-follow")
			http.Error(w, "sensitive synthetic failure", status)
		})
		if _, err := c.Extract(context.Background(), Input{Images: []Image{pngImage(t)}}); err == nil || calls.Load() != 1 {
			t.Fatalf("unexpected retry or redirect: %v calls=%d", err, calls.Load())
		}
	}
}

func TestModelFailurePreservesOCRSource(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/documents":
			w.WriteHeader(202)
			_, _ = io.WriteString(w, `{"documentId":"doc_1"}`)
		case "/v1/documents/doc_1":
			_, _ = io.WriteString(w, `{"status":"completed","result":{"text":"source evidence"}}`)
		default:
			http.Error(w, "private provider diagnostic", 500)
		}
	})
	out, err := c.Extract(context.Background(), Input{Images: []Image{pngImage(t)}})
	if err == nil || out.SourceText != "source evidence" || len(out.Drafts) != 0 {
		t.Fatal("lost OCR evidence or produced drafts on model failure")
	}
}

func TestOCRTerminalFailuresAndPartialFailure(t *testing.T) {
	for _, status := range []string{"failed", "cancelled", "canceled", "expired", "unknown", "http410", "past-expiry", "partial"} {
		t.Run(status, func(t *testing.T) {
			var submits atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" && r.URL.Path == "/v1/documents" {
					submits.Add(1)
					w.WriteHeader(202)
					_, _ = io.WriteString(w, `{"documentId":"doc_1","status":"queued"}`)
					return
				}
				if r.Method == "GET" {
					if status == "http410" {
						http.Error(w, "private upstream details", 410)
						return
					}
					if status == "past-expiry" {
						_, _ = io.WriteString(w, `{"status":"completed","resultExpiresAt":"2000-01-01T00:00:00Z","result":{"text":"old"}}`)
						return
					}
					state := status
					if status == "partial" && submits.Load() == 1 {
						_, _ = io.WriteString(w, `{"status":"completed","result":{"text":"first receipt"}}`)
						return
					}
					if status == "partial" {
						state = "failed"
					}
					_ = json.NewEncoder(w).Encode(map[string]string{"status": state})
					return
				}
				t.Error("model must not run on failed OCR")
			})
			images := []Image{pngImage(t)}
			want := int32(1)
			if status == "partial" {
				images = append(images, pngImage(t))
				want = 2
			}
			out, err := c.Extract(context.Background(), Input{Images: images})
			if err == nil || len(out.Drafts) != 0 || submits.Load() != want || strings.Contains(err.Error(), "private") {
				t.Fatalf("unsafe OCR failure: %v", err)
			}
		})
	}
}

func TestErrorsAreRedactedAndNotRetried(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "test-model-key private content", 500)
	})
	_, err := c.Extract(context.Background(), Input{Text: "test"})
	if err == nil || strings.Contains(err.Error(), "test-model-key") || strings.Contains(err.Error(), "private content") || calls.Load() != 1 {
		t.Fatalf("unsafe upstream error: %v", err)
	}
}

func TestContextTimeoutAndCancellation(t *testing.T) {
	for _, ocr := range []bool{false, true} {
		t.Run(map[bool]string{false: "model", true: "OCR"}[ocr], func(t *testing.T) {
			release := make(chan struct{})
			defer close(release)
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if ocr {
					if r.Method == "POST" {
						w.WriteHeader(202)
						_, _ = io.WriteString(w, `{"documentId":"doc_1"}`)
					} else {
						w.Header().Set("Retry-After", "120")
						_, _ = io.WriteString(w, `{"status":"processing"}`)
					}
					return
				}
				_, _ = io.Copy(io.Discard, r.Body)
				select {
				case <-r.Context().Done():
				case <-release:
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
			defer cancel()
			in := Input{Text: "test"}
			if ocr {
				in.Images = []Image{pngImage(t)}
			}
			start := time.Now()
			if _, err := c.Extract(ctx, in); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("expected deadline: %v", err)
			}
			if time.Since(start) > time.Second {
				t.Fatal("deadline not respected")
			}
			ctx2, cancel2 := context.WithCancel(context.Background())
			cancel2()
			if _, err := c.Extract(ctx2, in); !errors.Is(err, context.Canceled) {
				t.Fatalf("expected cancellation: %v", err)
			}
		})
	}
}

func TestInputBounds(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid input sent upstream") })
	for _, in := range []Input{{Text: strings.Repeat("x", 32769)}, {}} {
		if _, err := c.Extract(context.Background(), in); err == nil {
			t.Fatal("invalid input accepted")
		}
	}
}
