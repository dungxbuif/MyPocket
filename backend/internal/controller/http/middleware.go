package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

const (
	contextUserIDKey           = "userID"
	contextAdvisorPrincipalKey = "advisorPrincipal"
	missingAuthHeaderMsg       = "thiếu Authorization"
	invalidAuthFormatMsg       = "định dạng Authorization không hợp lệ"
	invalidTokenMsg            = "token không hợp lệ hoặc đã hết hạn"
	sessionExpiredMsg          = "phiên đăng nhập đã hết hiệu lực"
	verifySessionErrorMsg      = "lỗi kiểm tra phiên"
	authHeader                 = "Authorization"
	authScheme                 = "Bearer"
)

type AuthMiddleware struct {
	TokenSvc      usecase.TokenService
	VerifySession func(string) (string, error)
	APIKeys       *usecase.UserAPIKeyService
	Audit         repository.AuditSink
}

func NewAuthMiddleware(tokenSvc usecase.TokenService, verifySession func(string) (string, error)) *AuthMiddleware {
	return &AuthMiddleware{
		TokenSvc:      tokenSvc,
		VerifySession: verifySession,
	}
}

func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: missingAuthHeaderMsg})
		c.Abort()
		return
	}
	content := strings.TrimSpace(authHeader)
	parts := strings.SplitN(content, " ", 2)
	if len(parts) != 2 || parts[0] != authScheme {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthInvalidFormat, Title: problemTitleUnauthorized, Detail: invalidAuthFormatMsg})
		c.Abort()
		return
	}
	claims, err := m.TokenSvc.Verify(parts[1])
	if err != nil {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeTokenInvalid, Title: problemTitleUnauthorized, Detail: invalidTokenMsg})
		c.Abort()
		return
	}
	userID, err := m.VerifySession(claims.SessionID)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeSessionExpired, Title: problemTitleUnauthorized, Detail: sessionExpiredMsg})
		} else {
			Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeSessionVerifyFailed, Title: problemTitleInternalServer, Detail: verifySessionErrorMsg})
		}
		c.Abort()
		return
	}
	c.Set(contextUserIDKey, userID)
	c.Set(contextAdvisorPrincipalKey, usecase.Principal{OwnerID: userID, CredentialKind: "session", CredentialID: claims.SessionID, ExpiresAt: claims.ExpiresAt})
	c.Next()
}

// RequireAdvisorAuth accepts the normal JWT/session or a user API key. A
// value that looks like an API key never falls back to JWT verification.
// The first gate is finance:read; endpoint handlers add advisor:read/chat.
func (m *AuthMiddleware) RequireAdvisorAuth(c *gin.Context) {
	startedAt := time.Now()
	defer func() { m.recordAdvisorAudit(c, startedAt) }()
	header := strings.TrimSpace(c.GetHeader(authHeader))
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != authScheme || strings.TrimSpace(parts[1]) == "" {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthInvalidFormat, Title: problemTitleUnauthorized, Detail: invalidAuthFormatMsg})
		c.Abort()
		return
	}
	token := strings.TrimSpace(parts[1])
	if strings.HasPrefix(token, "mpk_") {
		if m == nil || m.APIKeys == nil {
			Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeTokenInvalid, Title: problemTitleUnauthorized, Detail: invalidTokenMsg})
			c.Abort()
			return
		}
		principal, err := m.APIKeys.Authenticate(c.Request.Context(), token, []string{entity.APIKeyScopeFinanceRead})
		if err != nil {
			status := http.StatusUnauthorized
			code := problemCodeTokenInvalid
			if errors.Is(err, usecase.ErrAPIKeyScopeDenied) {
				status = http.StatusForbidden
				code = "api_key_scope_denied"
			}
			Fail(c, status, Problem{Code: code, Title: problemTitleUnauthorized, Detail: invalidTokenMsg})
			c.Abort()
			return
		}
		c.Set(contextUserIDKey, principal.OwnerID)
		c.Set(contextAdvisorPrincipalKey, principal)
		c.Next()
		return
	}
	m.RequireAuth(c)
}

func (m *AuthMiddleware) recordAdvisorAudit(c *gin.Context, startedAt time.Time) {
	if m == nil || m.Audit == nil || c == nil || !strings.HasPrefix(c.FullPath(), advisorRoute) {
		return
	}
	status := c.Writer.Status()
	reason := "allowed"
	if status >= http.StatusInternalServerError {
		reason = "error"
	} else if status >= http.StatusBadRequest {
		reason = "denied"
	}
	principal, _ := advisorPrincipalFromContext(c)
	actorKind := principal.CredentialKind
	if actorKind == "" {
		actorKind = "unknown"
	}
	action := "advisor." + strings.ToLower(c.Request.Method) + "." + strings.ReplaceAll(strings.TrimPrefix(strings.TrimPrefix(c.FullPath(), advisorRoute), "/"), "/", ".")
	_ = m.Audit.Record(c.Request.Context(), repository.AuditEvent{
		RequestID: requestID(c), ActorKind: actorKind, CredentialID: principal.CredentialID,
		Action: action, Allowed: status >= http.StatusOK && status < http.StatusBadRequest,
		Status: http.StatusText(status), Reason: reason, LatencyMS: time.Since(startedAt).Milliseconds(),
	})
}

func advisorPrincipalFromContext(c *gin.Context) (usecase.Principal, bool) {
	if c == nil {
		return usecase.Principal{}, false
	}
	value, exists := c.Get(contextAdvisorPrincipalKey)
	principal, ok := value.(usecase.Principal)
	return principal, exists && ok
}
