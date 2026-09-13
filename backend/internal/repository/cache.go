package repository

import "time"

type CacheRepository interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, bool, error)
	Delete(key string) error
}
