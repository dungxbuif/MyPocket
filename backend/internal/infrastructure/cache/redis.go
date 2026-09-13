package cache

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
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

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) String() string {
	return fmt.Sprintf("redis://%s", r.client.Options().Addr)
}
