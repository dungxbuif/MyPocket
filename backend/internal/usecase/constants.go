package usecase

const (
	sessionCachePrefix = "session:"
	refreshTokenPrefix = "refresh-token:"
	profileCachePrefix = "profile:"
	homeCachePrefix    = "home:"
	cacheProfileTTL    = 60 // seconds
	cacheHomeTTL       = 30 // seconds
)
