package repository

import "time"

type CacheRepository interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, bool, error)
	Delete(key string) error
}

// AtomicCacheRepository is optional. Consumers use it when a value must be
// consumed exactly once (for example, rotating refresh tokens).
type AtomicCacheRepository interface {
	CacheRepository
	GetAndDelete(key string) (string, bool, error)
}
