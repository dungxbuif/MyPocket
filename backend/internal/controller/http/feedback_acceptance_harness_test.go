package httpapi

// This is a traceability/acceptance guard for the remaining UAT queue. It is
// deliberately local-only: it validates the synthetic fixture manifest and
// its deterministic source evidence, but never calls a provider or production.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type feedbackHarnessManifest struct {
	Scope              string                `json:"scope"`
	ProductionMutation bool                  `json:"production_mutation"`
	Cases              []feedbackHarnessCase `json:"cases"`
}

type feedbackHarnessCase struct {
	ID                string                    `json:"id"`
	Area              string                    `json:"area"`
	LocalEvidence     []feedbackHarnessEvidence `json:"local_evidence"`
	ExternalOnlyGates []string                  `json:"external_only_gates"`
}

type feedbackHarnessEvidence struct {
	File  string   `json:"file"`
	Tests []string `json:"tests"`
}

func TestFeedbackAcceptanceHarnessManifestIsSyntheticAndTraceable(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../../.."))
	manifestPath := filepath.Join(root, "docs/work/tickets/FEEDBACK-20261019-UAT-HARNESS.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read harness manifest: %v", err)
	}
	lower := strings.ToLower(string(data))
	for _, forbidden := range []string{"money.dungxbuif.com", "production_secret", "live_api_key", "authorization: bearer"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("manifest contains forbidden production material %q", forbidden)
		}
	}
	var manifest feedbackHarnessManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode harness manifest: %v", err)
	}
	if manifest.Scope != "synthetic-local-acceptance" || manifest.ProductionMutation {
		t.Fatalf("manifest must be synthetic and non-mutating: scope=%q production_mutation=%v", manifest.Scope, manifest.ProductionMutation)
	}
	expected := map[string]bool{
		"62e31193": true, "726943be": true, "1db32af3": true,
		"31ef1877": true, "d098ff72": true, "2d0811c0": true,
		"26fcad87": true, "06fb8f5b": true, "4ca27075": true,
		"fc8cae65": true, "824dea29": true, "7ddf77f3": true,
		"019b2e9c": true,
	}
	if len(manifest.Cases) != len(expected) {
		t.Fatalf("manifest has %d cases, want %d", len(manifest.Cases), len(expected))
	}
	seen := make(map[string]bool, len(manifest.Cases))
	for _, item := range manifest.Cases {
		if !expected[item.ID] || seen[item.ID] {
			t.Fatalf("unexpected or duplicate feedback id %q", item.ID)
		}
		seen[item.ID] = true
		if strings.TrimSpace(item.Area) == "" || len(item.ExternalOnlyGates) == 0 || len(item.LocalEvidence) == 0 {
			t.Fatalf("case %s needs area, local evidence and external-only gates", item.ID)
		}
		for _, evidence := range item.LocalEvidence {
			if filepath.IsAbs(evidence.File) || strings.TrimSpace(evidence.File) == "" || len(evidence.Tests) == 0 {
				t.Fatalf("case %s has invalid evidence declaration: %+v", item.ID, evidence)
			}
			path := filepath.Join(root, filepath.FromSlash(evidence.File))
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("case %s evidence file %s: %v", item.ID, evidence.File, err)
			}
			text := string(source)
			for _, symbol := range evidence.Tests {
				if !strings.Contains(text, symbol) {
					t.Fatalf("case %s evidence symbol %q missing from %s", item.ID, symbol, evidence.File)
				}
			}
		}
	}
	for id := range expected {
		if !seen[id] {
			t.Fatalf("expected feedback id %s is missing", id)
		}
	}
}
