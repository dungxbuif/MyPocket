package entity

import (
	"testing"
	"time"
)

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
