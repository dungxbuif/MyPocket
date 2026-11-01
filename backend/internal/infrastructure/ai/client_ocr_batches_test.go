package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestOCRDocumentBatchesRejectOversizedDocumentBeforeModel(t *testing.T) {
	var models atomic.Int32
	text := strings.Repeat("x", maxSourceBytes+1)
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/documents":
			_ = json.NewEncoder(w).Encode(map[string]any{"documentId": "doc", "status": "completed", "result": map[string]any{"text": text}})
		case "/v1/chat/completions":
			models.Add(1)
			modelResponse(w, `{"reply":"review","drafts":[]}`)
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	})
	out, err := c.Extract(context.Background(), Input{Images: []Image{pngImage(t)}})
	var diagnostic *AIProviderError
	if !errors.As(err, &diagnostic) || diagnostic.Code != "model_source_limit" || models.Load() != 0 || len(out.Drafts) != 0 {
		t.Fatalf("oversized document was split or misdiagnosed: calls=%d err=%v", models.Load(), err)
	}
	if !out.OCRComplete || len(out.AttachmentTexts) != 1 || out.AttachmentTexts[0] != text || !strings.Contains(out.SourceText, text) {
		t.Fatal("oversized evidence was truncated")
	}
}

func TestOCRDocumentBatchesPreserveEvidenceAndFailAtomically(t *testing.T) {
	for _, failLast := range []bool{false, true} {
		t.Run(map[bool]string{false: "complete", true: "later_failure"}[failLast], func(t *testing.T) {
			var ocrCalls atomic.Int32
			var sources []string
			text := strings.Repeat("transaction row\n", 2800)
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v1/documents":
					ocrCalls.Add(1)
					_ = json.NewEncoder(w).Encode(map[string]any{"documentId": "doc", "status": "completed", "result": map[string]any{"text": text}})
				case "/v1/chat/completions":
					var request struct {
						Messages []chatMessage `json:"messages"`
					}
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						t.Error(err)
						return
					}
					var payload struct {
						Source      string `json:"untrusted_source_text"`
						Instruction string `json:"user_instruction"`
					}
					if err := json.Unmarshal([]byte(request.Messages[1].Content), &payload); err != nil {
						t.Error(err)
						return
					}
					if payload.Instruction != "extract all" || len(payload.Source) > maxSourceBytes {
						t.Error("chunk lost instruction or exceeded budget")
					}
					sources = append(sources, payload.Source)
					if failLast && len(sources) == 2 {
						w.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					modelResponse(w, `{"reply":"review","drafts":[{"type":"expense","amount":10,"wallet_id":"w","category_id":null,"occurred_at":"","note":"same row","included_in_reports":true,"questions":[],"source_currency":"VND"}]}`)
				default:
					t.Errorf("unexpected request %s", r.URL.Path)
				}
			})
			first, second := pngImage(t), pngImage(t)
			first.Name, second.Name = "first.png", "second.png"
			out, err := c.Extract(context.Background(), Input{Text: "extract all", Images: []Image{first, second}})
			if ocrCalls.Load() != 2 || len(sources) != 2 || len(out.AttachmentTexts) != 2 || !out.OCRComplete {
				t.Fatalf("incomplete evidence: OCR=%d models=%d err=%v", ocrCalls.Load(), len(sources), err)
			}
			if strings.Count(strings.Join(sources, ""), text) != 2 || !strings.Contains(sources[0], "OCR document 1") || !strings.Contains(sources[1], "OCR document 2") || strings.Count(out.SourceText, text) != 2 {
				t.Fatal("source lost, repeated or reordered")
			}
			if strings.Count(strings.Join(sources, ""), "extract all") != 1 {
				t.Fatal("user source evidence repeated between chunks")
			}
			if failLast {
				if err == nil || len(out.Drafts) != 0 {
					t.Fatal("partial proposals escaped failed batch")
				}
			} else if err != nil || len(out.Drafts) != 2 {
				t.Fatalf("identical rows lost: count=%d err=%v", len(out.Drafts), err)
			}
		})
	}
}
