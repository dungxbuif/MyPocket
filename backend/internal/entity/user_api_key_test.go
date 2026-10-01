package entity

import (
	"testing"
	"time"
)

func TestNormalizeAPIKeyScopesAddsAdvisorPrerequisites(t *testing.T) {
	scopes, err := NormalizeAPIKeyScopes([]string{APIKeyScopeAdvisorChat, APIKeyScopeAdvisorChat})
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{APIKeyScopeAdvisorChat, APIKeyScopeAdvisorRead, APIKeyScopeFinanceRead} {
		found := false
		for _, scope := range scopes {
			if scope == required {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing implied scope %q: %v", required, scopes)
		}
	}
}

func TestNormalizeAPIKeyScopesAddsFeedbackReadForWrite(t *testing.T) {
	scopes, err := NormalizeAPIKeyScopes([]string{APIKeyScopeFeedbackWrite})
	if err != nil {
		t.Fatal(err)
	}
	if !containsScope(scopes, APIKeyScopeFeedbackWrite) || !containsScope(scopes, APIKeyScopeFeedbackRead) {
		t.Fatalf("feedback write key must include feedback read: %v", scopes)
	}
}

func TestNormalizeAPIKeyScopesRejectsUnknownAndEmpty(t *testing.T) {
	if _, err := NormalizeAPIKeyScopes(nil); err == nil {
		t.Fatal("empty scopes must be rejected")
	}
	if _, err := NormalizeAPIKeyScopes([]string{"admin"}); err == nil {
		t.Fatal("unknown scope must be rejected")
	}
}

func TestUserAPIKeyActiveAtHonorsExpiryAndRevocation(t *testing.T) {
	now := time.Date(2026, 9, 22, 8, 0, 0, 0, time.UTC)
	expired := now.Add(-time.Second)
	key := UserAPIKey{ExpiresAt: &expired}
	if key.ActiveAt(now) {
		t.Fatal("expired key must be inactive")
	}
	revoked := now.Add(-time.Second)
	key = UserAPIKey{RevokedAt: &revoked}
	if key.ActiveAt(now) {
		t.Fatal("revoked key must be inactive")
	}
}
