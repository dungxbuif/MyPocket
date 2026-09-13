package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/usecase"
)

const (
	contextUserIDKey      = "userID"
	missingAuthHeaderMsg  = "thiếu Authorization"
	invalidAuthFormatMsg  = "định dạng Authorization không hợp lệ"
	invalidTokenMsg       = "token không hợp lệ hoặc đã hết hạn"
	sessionExpiredMsg     = "phiên đăng nhập đã hết hiệu lực"
	verifySessionErrorMsg = "lỗi kiểm tra phiên"
	authHeader            = "Authorization"
	authScheme            = "Bearer"
)

type AuthMiddleware struct {
	TokenSvc      usecase.TokenService
	VerifySession func(string) (string, error)
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": missingAuthHeaderMsg})
		c.Abort()
		return
	}
	content := strings.TrimSpace(authHeader)
	parts := strings.SplitN(content, " ", 2)
	if len(parts) != 2 || parts[0] != authScheme {
		c.JSON(http.StatusUnauthorized, gin.H{"error": invalidAuthFormatMsg})
		c.Abort()
		return
	}
	claims, err := m.TokenSvc.Verify(parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": invalidTokenMsg})
		c.Abort()
		return
	}
	userID, err := m.VerifySession(claims.SessionID)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": sessionExpiredMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": verifySessionErrorMsg})
		}
		c.Abort()
		return
	}
	c.Set(contextUserIDKey, userID)
	c.Next()
}
