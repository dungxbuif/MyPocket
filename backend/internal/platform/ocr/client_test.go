package ocr

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mypocket/internal/agent"
)

func TestOCRClientContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer hidden-key" {
			t.Fatalf("bad auth")
		}
		switch r.URL.Path {
		case "/v1/ocr/capabilities":
			io.WriteString(w, `{"api":"v1","input":["base64"]}`)
		case "/v1/documents":
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(202)
			io.WriteString(w, `{"document_id":"doc-1","status":"processing"}`)
		case "/v1/documents/doc-1":
			io.WriteString(w, `{"status":"completed","text":"Total 120000","fields":{"total":120000}}`)
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
	sub, err := client.Submit(context.Background(), agent.ImageInput{ContentType: "image/jpeg", Bytes: []byte("image")})
	if err != nil || sub.ProviderID != "doc-1" || sub.RetryAfter != 2*time.Second {
		t.Fatalf("sub=%#v err=%v", sub, err)
	}
	result, err := client.Read(context.Background(), sub.ProviderID)
	if err != nil || result.Status != agent.ToolCompleted || result.Text == "" {
		t.Fatalf("result=%#v err=%v", result, err)
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, `{"api":"v2","input":["url"]}`) }))
	defer server.Close()
	client, _ := New(server.URL, "key", time.Second)
	if client.Capabilities(context.Background()) == nil {
		t.Fatal("expected incompatible capabilities")
	}
}
