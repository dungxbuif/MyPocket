package httpapi

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"mypocket/internal/audit"
	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

func appendAudit(ctx context.Context, repo AuditRepository, event audit.Event) {
	if repo == nil {
		return
	}
	if err := repo.Append(ctx, event); err != nil {
		slog.Warn("audit append failed", "correlation_id", event.CorrelationID, "action", event.Action)
	}
}

func authenticatedUserIDFromCookie(r *http.Request, cfg config.Config) string {
	cookie, err := r.Cookie(identity.AuthCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	claims, err := identity.NewCookieSigner([]byte(cfg.CookieSecret)).Verify(cookie.Value)
	if err != nil {
		return ""
	}
	return claims.UserID
}

func authenticatedUserIDFromRequest(r *http.Request, cfg config.Config) string {
	if userID := authenticatedUserIDFromContext(r.Context()); userID != "" {
		return userID
	}
	return authenticatedUserIDFromCookie(r, cfg)
}

func shouldAuditRequest(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return false
	}
	path := r.URL.Path
	if strings.HasPrefix(path, "/api/v1/audit/") || strings.HasPrefix(path, "/api/v1/health/") {
		return false
	}
	return strings.HasPrefix(path, "/api/v1/")
}

func requestAction(r *http.Request) string {
	return "http." + strings.ToLower(r.Method) + "." + strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/"), "/")
}

func outcomeFromStatus(status int) audit.Outcome {
	switch {
	case status == http.StatusConflict:
		return audit.OutcomeConflict
	case status == http.StatusForbidden || status == http.StatusUnauthorized:
		return audit.OutcomeDenied
	case status >= 500:
		return audit.OutcomeFailure
	case status >= 400:
		return audit.OutcomeFailure
	default:
		return audit.OutcomeSuccess
	}
}

func severityFromStatus(status int) audit.Severity {
	switch {
	case status == http.StatusForbidden || status == http.StatusUnauthorized:
		return audit.SeveritySecurity
	case status >= 500:
		return audit.SeverityError
	case status >= 400:
		return audit.SeverityWarn
	default:
		return audit.SeverityInfo
	}
}

func entityTypeFromPath(path string) string {
	parts := apiPathParts(path)
	if len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "wallets":
		return "wallet"
	case "categories":
		return "category"
	case "transactions":
		return "transaction"
	case "budgets":
		return "budget"
	case "events":
		return "event"
	case "obligations":
		return "obligation"
	case "recurring-schedules":
		return "recurring_schedule"
	case "notifications":
		return "notification"
	case "push-subscriptions":
		return "push_subscription"
	case "assets":
		return "asset"
	case "sync":
		return "sync"
	case "auth":
		return "auth"
	default:
		return parts[0]
	}
}

func entityIDFromPath(path string) string {
	parts := apiPathParts(path)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func apiPathParts(path string) []string {
	path = strings.Trim(strings.TrimPrefix(path, "/api/v1/"), "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func safeRequestPath(r *http.Request) string {
	return r.URL.EscapedPath()
}

func clientIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
