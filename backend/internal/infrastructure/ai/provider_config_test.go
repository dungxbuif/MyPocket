package ai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mypocket/backend/internal/usecase"
)

func TestProviderConfigurationValidationRejectsUnsafeOrIncompleteValues(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{name: "missing endpoint", cfg: Config{APIKey: "key", Model: "model"}},
		{name: "missing key", cfg: Config{BaseURL: "https://provider.example/v1", Model: "model"}},
		{name: "missing model", cfg: Config{BaseURL: "https://provider.example/v1", APIKey: "key"}},
		{name: "non-http endpoint", cfg: Config{BaseURL: "ftp://provider.example/v1", APIKey: "key", Model: "model"}},
		{name: "endpoint credentials", cfg: Config{BaseURL: "https://user:password@provider.example/v1", APIKey: "key", Model: "model"}},
		{name: "endpoint query", cfg: Config{BaseURL: "https://provider.example/v1?key=secret", APIKey: "key", Model: "model"}},
		{name: "endpoint fragment", cfg: Config{BaseURL: "https://provider.example/v1#secret", APIKey: "key", Model: "model"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err == nil {
				t.Fatal("unsafe or incomplete provider configuration was accepted")
			}
			if (tt.cfg.APIKey != "" && strings.Contains(err.Error(), tt.cfg.APIKey)) || strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "secret") {
				t.Fatalf("validation error leaked credential material: %v", err)
			}
			if NewClient(tt.cfg).Configured() {
				t.Fatal("invalid provider configuration reported as configured")
			}
		})
	}

	valid := Config{BaseURL: "https://provider.example/v1", APIKey: "key", Model: "model"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid provider configuration rejected: %v", err)
	}
}

func TestAdvisorProviderConfigurationUsesTheSameFailClosedContract(t *testing.T) {
	for _, cfg := range []AdvisorClientConfig{
		{},
		{BaseURL: "https://provider.example/v1", APIKey: "key"},
		{BaseURL: "https://user:password@provider.example/v1", APIKey: "key", Model: "model"},
	} {
		if err := cfg.Validate(); err == nil {
			t.Fatal("invalid advisor provider configuration was accepted")
		}
		if NewAdvisorClient(cfg).Available() {
			t.Fatal("invalid advisor provider configuration reported as available")
		}
	}
	if err := (AdvisorClientConfig{BaseURL: "https://provider.example/v1", APIKey: "key", Model: "model"}).Validate(); err != nil {
		t.Fatalf("valid advisor provider configuration rejected: %v", err)
	}
}

func TestExtractionCredentialIsHeaderOnlyAndNeverFallsBack(t *testing.T) {
	const secret = "model-secret-that-must-not-leak"
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		if strings.Contains(string(body), secret) || strings.Contains(r.URL.String(), secret) {
			t.Fatalf("provider secret leaked into request body or URL: %s", body)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+secret {
			t.Fatalf("provider secret must be sent only as bearer header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		modelResponse(w, `{"reply":"ok","drafts":[]}`)
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL + "/v1", APIKey: secret, Model: "model-a"})
	if _, err := client.Extract(context.Background(), Input{Text: "coffee"}); err != nil {
		t.Fatalf("configured provider request failed: %v", err)
	}
	if requests != 1 {
		t.Fatalf("provider failure/selection must not trigger a hidden fallback; requests=%d", requests)
	}
}

func TestAdvisorCredentialIsHeaderOnly(t *testing.T) {
	const secret = "advisor-secret-that-must-not-leak"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		if strings.Contains(string(body), secret) || strings.Contains(r.URL.String(), secret) {
			t.Fatalf("advisor secret leaked into request body or URL")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+secret {
			t.Fatalf("advisor secret must be sent only as bearer header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()

	client := NewAdvisorClient(AdvisorClientConfig{BaseURL: server.URL + "/v1", APIKey: secret, Model: "model-a"})
	if _, err := client.Chat(context.Background(), usecase.AdvisorRequest{Messages: []usecase.AdvisorChatMessage{{Role: "user", Content: "hello"}}}); err != nil {
		t.Fatalf("configured advisor request failed: %v", err)
	}
}
