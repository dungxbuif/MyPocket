package cache

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/mypocket/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	outbox repository.AuditOutboxRepository
}

func NewRedis(redisURL string) (*Redis, error) {
	u, err := url.Parse(redisURL)
	if err != nil {
		return nil, err
	}
	opt, err := redis.ParseURL(u.String())
	if err != nil {
		return nil, err
	}
	// Audit writes are intentionally bounded and retryable. Keep the retry
	// budget short so an unavailable cache cannot hold an HTTP request open
	// indefinitely; Record also uses an event fingerprint to make a retry
	// idempotent at the stream boundary.
	opt.MaxRetries = 3
	opt.MinRetryBackoff = 20 * time.Millisecond
	opt.MaxRetryBackoff = 200 * time.Millisecond
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Redis{client: client}, nil
}

func (r *Redis) Set(key, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Get(key string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return result, true, nil
}

func (r *Redis) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Del(ctx, key).Err()
}

func (r *Redis) GetAndDelete(key string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := r.client.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return result, true, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

// SetAuditOutbox enables durable hand-off while retaining Redis as the fast
// path. It is called during API/worker wiring before serving requests.
func (r *Redis) SetAuditOutbox(outbox repository.AuditOutboxRepository) {
	if r != nil {
		r.outbox = outbox
	}
}

func (r *Redis) String() string {
	return fmt.Sprintf("redis://%s", r.client.Options().Addr)
}
