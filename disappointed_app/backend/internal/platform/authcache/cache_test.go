package authcache_test

import (
	"context"
	"os"
	"testing"
	"time"

	"mypocket/internal/platform/authcache"
)

func TestDigestTokenIsDeterministicAndNonPlaintext(t *testing.T) {
	const token = "session-secret"
	first := authcache.DigestToken(token)
	if first == token || first != authcache.DigestToken(token) {
		t.Fatalf("digest should be deterministic and opaque: %q", first)
	}
}

func TestRedisRoundTripWhenConfigured(t *testing.T) {
	rawURL := os.Getenv("MYPOCKET_TEST_REDIS_URL")
	if rawURL == "" {
		t.Skip("MYPOCKET_TEST_REDIS_URL is not set")
	}

	cache, err := authcache.NewRedis(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	if cache == nil {
		t.Fatal("expected redis cache")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := "mypocket:test:authcache:" + authcache.DigestToken(t.Name())
	defer cache.Delete(context.Background(), key)

	if err := cache.Set(ctx, key, "user_test", time.Minute); err != nil {
		t.Fatal(err)
	}
	value, ok, err := cache.Get(ctx, key)
	if err != nil || !ok || value != "user_test" {
		t.Fatalf("unexpected cached value: value=%q ok=%v err=%v", value, ok, err)
	}
	if err := cache.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := cache.Get(ctx, key); err != nil || ok {
		t.Fatalf("expected deleted key, ok=%v err=%v", ok, err)
	}
}
