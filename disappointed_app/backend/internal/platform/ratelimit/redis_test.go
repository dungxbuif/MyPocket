package ratelimit

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRedisLimiterEnforcesIndependentFixedWindows(t *testing.T) {
	raw := os.Getenv("REDIS_URL")
	if raw == "" {
		t.Skip("REDIS_URL is not set")
	}
	limiter, err := NewRedis(raw)
	if err != nil {
		t.Fatal(err)
	}
	key := time.Now().Format("150405.000000000")
	for i := 0; i < 2; i++ {
		allowed, _, err := limiter.Allow(context.Background(), key, 2, time.Minute)
		if err != nil || !allowed {
			t.Fatalf("request %d allowed=%v err=%v", i, allowed, err)
		}
	}
	allowed, retry, err := limiter.Allow(context.Background(), key, 2, time.Minute)
	if err != nil || allowed || retry <= 0 {
		t.Fatalf("limit allowed=%v retry=%v err=%v", allowed, retry, err)
	}
	other, _, err := limiter.Allow(context.Background(), key+"-other", 2, time.Minute)
	if err != nil || !other {
		t.Fatalf("independent key allowed=%v err=%v", other, err)
	}
}
func TestRedisLimiterFailsClosedWhenUnavailable(t *testing.T) {
	limiter, err := NewRedis("redis://127.0.0.1:1/0")
	if err != nil {
		t.Fatal(err)
	}
	if allowed, _, err := limiter.Allow(context.Background(), "key", 1, time.Minute); err == nil || allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
}
