package ai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOCRCompletedPollInvalidResultKeepsDiagnostic(t *testing.T) {
	for _, result := range []string{
		`{"status":"completed","result":{"text":""}}`,
		`{"status":"completed","resultExpiresAt":"2000-01-01T00:00:00Z","result":{"text":"expired evidence"}}`,
	} {
		t.Run(result, func(t *testing.T) {
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					_, _ = io.WriteString(w, `{"status":"queued","documentId":"doc_1"}`)
					return
				}
				_, _ = io.WriteString(w, result)
			})
			_, err := client.Extract(context.Background(), Input{Images: []Image{pngImage(t)}})
			var diagnostic *AIProviderError
			if !errors.As(err, &diagnostic) || diagnostic.Stage != "ocr_poll" || diagnostic.Code != "ocr_result_invalid" {
				t.Fatalf("completed OCR poll lost stage/code: %T %v", err, err)
			}
		})
	}
}

func TestOCRPartialBatchPreservesCompletedSourceForRecovery(t *testing.T) {
	submissions := 0
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/documents" {
			t.Error("incomplete OCR batch must not call the model")
			http.Error(w, "unexpected model request", http.StatusInternalServerError)
			return
		}
		submissions++
		if submissions == 1 {
			_, _ = io.WriteString(w, `{"status":"completed","result":{"text":"first receipt amount 42000"}}`)
			return
		}
		_, _ = io.WriteString(w, `{"status":"failed"}`)
	})
	out, err := client.Extract(context.Background(), Input{Text: "review my receipts", Images: []Image{pngImage(t), pngImage(t)}})
	if err == nil || out.OCRComplete || len(out.Drafts) != 0 || len(out.AttachmentTexts) != 1 {
		t.Fatalf("partial batch must remain incomplete without drafts: %+v err=%v", out, err)
	}
	if !strings.Contains(out.SourceText, "first receipt amount 42000") || !strings.Contains(out.SourceText, "OCR document 1") || !strings.Contains(out.SourceText, "review my receipts") {
		t.Fatalf("completed document evidence was lost: %q", out.SourceText)
	}
	if strings.Contains(out.SourceText, "OCR document 2") {
		t.Fatalf("failed document must not appear as completed evidence: %q", out.SourceText)
	}
}

func TestOCRSourceLimitDoesNotCompleteUnprocessedImages(t *testing.T) {
	submissions := 0
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/documents" {
			t.Error("oversized source must not call the model")
			return
		}
		submissions++
		_, _ = io.WriteString(w, `{"status":"completed","result":{"text":"`+strings.Repeat("x", maxSourceBytes-1)+`"}}`)
	})
	out, err := client.Extract(context.Background(), Input{Images: []Image{pngImage(t), pngImage(t)}})
	if err == nil || !strings.Contains(err.Error(), "source text exceeds limit") {
		t.Fatalf("expected aggregate source limit, got %v", err)
	}
	if out.OCRComplete || submissions != 1 || len(out.AttachmentTexts) != 1 || len(out.Drafts) != 0 {
		t.Fatalf("unprocessed second image incorrectly marked completed: complete=%v submissions=%d texts=%d drafts=%d", out.OCRComplete, submissions, len(out.AttachmentTexts), len(out.Drafts))
	}
}

func TestModelRejectedResponseRetainsCompletedOCRMetadata(t *testing.T) {
	for name, response := range map[string]string{
		"invalid JSON":        "{",
		"no choices":          `{"choices":[]}`,
		"filtered finish":     `{"choices":[{"finish_reason":"content_filter","message":{"content":""}}]}`,
		"unexpected tools":    `{"choices":[{"finish_reason":"stop","message":{"content":"","tool_calls":[{"id":"unexpected"}]}}]}`,
		"unexpected function": `{"choices":[{"finish_reason":"stop","message":{"content":"","function_call":{"name":"unexpected"}}}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			modelCalls := 0
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1/documents" {
					_, _ = io.WriteString(w, `{"status":"completed","result":{"text":"receipt total 42000"}}`)
					return
				}
				modelCalls++
				_, _ = io.WriteString(w, response)
			})
			out, err := client.Extract(context.Background(), Input{Images: []Image{pngImage(t)}})
			if err == nil || len(out.Drafts) != 0 || modelCalls != 1 {
				t.Fatalf("rejected response must fail without drafts/retry: err=%v drafts=%d calls=%d", err, len(out.Drafts), modelCalls)
			}
			if !out.OCRComplete || len(out.AttachmentTexts) != 1 || out.AttachmentTexts[0] != "receipt total 42000" || !strings.Contains(out.SourceText, "receipt total 42000") {
				t.Fatalf("model rejection discarded completed OCR evidence: %+v", out)
			}
		})
	}
}

func TestModelEnvelopeFailureCarriesSafeModelDiagnostic(t *testing.T) {
	for name, response := range map[string]string{
		"malformed envelope": `{`,
		"missing choices":    `{"id":"provider-request","choices":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/chat/completions" {
					t.Fatalf("unexpected provider path: %s", r.URL.Path)
				}
				_, _ = io.WriteString(w, response)
			})
			_, err := client.Extract(context.Background(), Input{Text: "receipt"})
			var diagnostic *AIProviderError
			if !errors.As(err, &diagnostic) || diagnostic.Stage != "model" || diagnostic.Code != "model_response_invalid" {
				t.Fatalf("invalid model envelope must carry a safe model diagnostic: %T %+v", err, err)
			}
			if strings.Contains(err.Error(), "provider-request") || strings.Contains(err.Error(), response) {
				t.Fatalf("model response details leaked through diagnostic: %v", err)
			}
		})
	}
}

func TestModelIncompleteResponseCarriesSafeModelDiagnostic(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[{"finish_reason":"tool_calls","message":{"content":"","tool_calls":[{"id":"call-1"}]}}]}`)
	})
	_, err := client.Extract(context.Background(), Input{Text: "receipt"})
	var diagnostic *AIProviderError
	if !errors.As(err, &diagnostic) || diagnostic.Stage != "model" || diagnostic.Code != "model_response_incomplete" {
		t.Fatalf("incomplete/tool response must carry a safe model diagnostic: %T %+v", err, err)
	}
}

func TestEmptyModelContentCarriesSchemaDiagnosticWithoutDrafts(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[{"finish_reason":"stop","message":{"content":""}}]}`)
	})
	out, err := client.Extract(context.Background(), Input{Text: "receipt"})
	var diagnostic *AIProviderError
	if !errors.As(err, &diagnostic) || diagnostic.Stage != "schema" || diagnostic.Code != "schema_mismatch" {
		t.Fatalf("empty model content must be a safe schema diagnostic: %T %+v", err, err)
	}
	if len(out.Drafts) != 0 {
		t.Fatalf("empty provider content must not produce drafts: %#v", out.Drafts)
	}
}
