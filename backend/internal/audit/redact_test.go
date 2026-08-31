package audit_test

import (
	"strings"
	"testing"

	"mypocket/internal/audit"
)

func TestHashValueIsStableAndDoesNotExposeInput(t *testing.T) {
	hash := audit.HashValue("01234567890123456789012345678901", "Owner@Example.com")
	again := audit.HashValue("01234567890123456789012345678901", "owner@example.com")
	if hash == "" || hash != again {
		t.Fatalf("expected stable lower-cased hash, got %q and %q", hash, again)
	}
	if strings.Contains(hash, "owner") || strings.Contains(hash, "example") {
		t.Fatalf("hash leaked input: %s", hash)
	}
}

func TestSafeMetadataAllowsOnlyDebugSafeKeys(t *testing.T) {
	body := string(audit.SafeMetadata(map[string]any{
		"status":        200,
		"mutation_id":   "mut_1",
		"authorization": "Bearer secret",
		"cookie":        "secret",
		"note":          "private note",
	}))
	if !strings.Contains(body, "status") || !strings.Contains(body, "mutation_id") {
		t.Fatalf("safe keys missing: %s", body)
	}
	if strings.Contains(body, "secret") || strings.Contains(body, "private note") || strings.Contains(body, "authorization") {
		t.Fatalf("metadata leaked unsafe data: %s", body)
	}
}
