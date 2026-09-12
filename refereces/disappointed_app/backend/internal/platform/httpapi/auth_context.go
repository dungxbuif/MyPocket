package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mypocket/internal/audit"
	"mypocket/internal/identity"
	"mypocket/internal/platform/authcache"
	"mypocket/internal/platform/config"
)

func authContextMiddleware(cfg config.Config, identityRepo IdentityRepository, apiKeyRepo APIKeyRepository, cache AuthCache, limiter APIKeyLimiter, auditRepo AuditRepository, next http.Handler) http.Handler {
	if cache == nil {
		cache = authcache.Noop{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(r.Header.Get("Authorization")) != "" {
			if userID, ok := resolveBearerUserID(r.Context(), cfg, apiKeyRepo, cache, bearerToken(r)); ok {
				if limiter == nil {
					if cfg.AppEnv == "production" {
						writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "API rate limiter unavailable", correlationID(r.Context())))
						return
					}
				} else {
					tokenHash, _ := identity.HashAPIKey(cfg.APIKeyHashSecret, bearerToken(r))
					allowed, retryAfter, err := limiter.Allow(r.Context(), userID+":"+tokenHash, cfg.APIRateLimitPerMinute, time.Minute)
					if err != nil {
						writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "API rate limiter unavailable", correlationID(r.Context())))
						return
					}
					if !allowed {
						seconds := int(retryAfter.Round(time.Second) / time.Second)
						if seconds < 1 {
							seconds = 1
						}
						w.Header().Set("Retry-After", fmt.Sprint(seconds))
						writeJSON(w, http.StatusTooManyRequests, ErrorEnvelope("RATE_LIMITED", "API rate limit exceeded", correlationID(r.Context())))
						return
					}
				}
				appendAudit(r.Context(), auditRepo, audit.Event{CorrelationID: correlationID(r.Context()), ActorUserID: userID, Action: "api_key.authenticate", EntityType: "api_key", Outcome: audit.OutcomeSuccess, Severity: audit.SeveritySecurity, Source: audit.SourceAPI, RequestMethod: r.Method, RequestPath: safeRequestPath(r)})
				next.ServeHTTP(w, r.WithContext(withAuthenticatedUser(r.Context(), userID, "api_key")))
				return
			}
			appendAudit(r.Context(), auditRepo, audit.Event{CorrelationID: correlationID(r.Context()), Action: "api_key.authenticate", EntityType: "api_key", Outcome: audit.OutcomeDenied, Severity: audit.SeveritySecurity, Source: audit.SourceAPI, ErrorCode: "AUTH_REQUIRED", RequestMethod: r.Method, RequestPath: safeRequestPath(r)})
			writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Authentication required", correlationID(r.Context())))
			return
		}
		if userID, ok := resolveCookieUserID(r.Context(), cfg, identityRepo, cache, r); ok {
			next.ServeHTTP(w, r.WithContext(withAuthenticatedUser(r.Context(), userID, "cookie")))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func resolveBearerUserID(ctx context.Context, cfg config.Config, repo APIKeyRepository, cache AuthCache, token string) (string, bool) {
	if token == "" || repo == nil {
		return "", false
	}
	keyHash, err := identity.HashAPIKey(cfg.APIKeyHashSecret, token)
	if err != nil {
		return "", false
	}
	cacheKey := apiKeyCacheKey(keyHash)
	cachedUserID, cached, _ := cache.Get(ctx, cacheKey)
	user, _, err := repo.AuthenticateAPIKey(ctx, token, cfg.APIKeyHashSecret)
	if err != nil {
		return "", false
	}
	if cached && cachedUserID != "" && cachedUserID != user.ID {
		_ = cache.Delete(ctx, cacheKey)
	}
	_ = cache.Set(ctx, cacheKey, user.ID, 5*time.Minute)
	return user.ID, true
}

func apiKeyCacheKey(keyHash string) string {
	return "auth:api_key_hash:" + keyHash
}

func resolveCookieUserID(ctx context.Context, cfg config.Config, repo IdentityRepository, cache AuthCache, r *http.Request) (string, bool) {
	cookie, err := r.Cookie(identity.AuthCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	cacheKey := "auth:session:" + authcache.DigestToken(cookie.Value)
	if userID, ok, err := cache.Get(ctx, cacheKey); err == nil && ok && userID != "" {
		if repo == nil {
			return userID, true
		}
		if _, findErr := repo.FindByID(ctx, userID); findErr == nil {
			return userID, true
		}
		_ = cache.Delete(ctx, cacheKey)
		return "", false
	}
	claims, err := identity.NewCookieSigner([]byte(cfg.CookieSecret)).Verify(cookie.Value)
	if err != nil {
		return "", false
	}
	if repo != nil {
		if _, err := repo.FindByID(ctx, claims.UserID); err != nil && !errors.Is(err, identity.ErrUserNotFound) {
			return "", false
		} else if errors.Is(err, identity.ErrUserNotFound) {
			return "", false
		}
	}
	_ = cache.Set(ctx, cacheKey, claims.UserID, time.Hour)
	return claims.UserID, true
}

func bearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return ""
	}
	value, ok := strings.CutPrefix(header, "Bearer ")
	if !ok {
		return ""
	}
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, identity.APIKeyPrefix) {
		return ""
	}
	return value
}
