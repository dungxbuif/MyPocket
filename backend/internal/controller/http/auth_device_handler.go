package httpapi

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"net/http"
)

type deviceLoginRequest struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

// EnrollDevice godoc
// @Summary Remember a device for biometric sign-in (password confirmation required)
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body usecase.DeviceEnrollmentInput true "Device enrollment"
// @Success 200 {object} usecase.DeviceCredentialOutput
// @Failure 400 {object} Problem
// @Failure 401 {object} Problem
// @Failure 403 {object} Problem
// @Router /api/v1/auth/devices [post]
func (h *AuthHandler) EnrollDevice(c *gin.Context) {
	var in usecase.DeviceEnrollmentInput
	if c.ShouldBindJSON(&in) != nil {
		Fail(c, http.StatusBadRequest, Problem{Code: "device_enrollment_invalid", Title: "Bad Request", Detail: "Thông tin đăng ký thiết bị không hợp lệ."})
		return
	}
	out, err := h.Auth.EnrollDevice(c.Request.Context(), c.GetString(contextUserIDKey), in)
	if errors.Is(err, usecase.ErrInvalidCredentials) {
		Fail(c, http.StatusForbidden, Problem{Code: "device_enrollment_password_invalid", Title: "Forbidden", Detail: "Mật khẩu xác nhận không đúng."})
		return
	}
	if err != nil {
		deviceAuthError(c, err)
		return
	}
	OK(c, out)
}

// DeviceLogin godoc
// @Summary Exchange a remembered device credential for a new access/refresh session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body deviceLoginRequest true "Device credential"
// @Success 200 {object} usecase.LoginOutput
// @Failure 401 {object} Problem
// @Router /api/v1/auth/device-login [post]
func (h *AuthHandler) DeviceLogin(c *gin.Context) {
	var in deviceLoginRequest
	if c.ShouldBindJSON(&in) != nil {
		deviceAuthError(c, usecase.ErrInvalidDeviceCredential)
		return
	}
	if _, err := uuid.Parse(in.ID); err != nil {
		deviceAuthError(c, usecase.ErrInvalidDeviceCredential)
		return
	}
	out, err := h.Auth.LoginWithDevice(c.Request.Context(), in.ID, in.Secret)
	if err != nil {
		deviceAuthError(c, err)
		return
	}
	OK(c, out)
}

// RevokeDevice godoc
// @Summary Forget a remembered device owned by the authenticated user
// @Tags Authentication
// @Security BearerAuth
// @Param id path string true "Device credential ID"
// @Success 204
// @Failure 401 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/auth/devices/{id} [delete]
func (h *AuthHandler) RevokeDevice(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		deviceAuthError(c, repository.ErrNotFound)
		return
	}
	if err := h.Auth.RevokeDevice(c.Request.Context(), c.GetString(contextUserIDKey), id); err != nil {
		deviceAuthError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Logout godoc
// @Summary Revoke the current access/refresh session family; keep remembered device enrollment
// @Description JWT logout revokes its access/refresh family. API-key logout revokes only that user key; no refresh token is needed for API keys.
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Param request body refreshRequest true "Current refresh token"
// @Success 204
// @Failure 401 {object} Problem
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	value, _ := c.Get(contextAdvisorPrincipalKey)
	principal, ok := value.(usecase.Principal)
	if !ok || principal.CredentialKind != "session" {
		deviceAuthError(c, usecase.ErrInvalidCredentials)
		return
	}
	var in refreshRequest
	if c.ShouldBindJSON(&in) != nil {
		Fail(c, http.StatusBadRequest, Problem{Code: "logout_invalid", Title: "Bad Request", Detail: "Yêu cầu đăng xuất không hợp lệ."})
		return
	}
	if err := h.Auth.Logout(c.Request.Context(), principal.OwnerID, principal.CredentialID, in.RefreshToken); err != nil {
		deviceAuthError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func deviceAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidDeviceEnrollment):
		Fail(c, 400, Problem{Code: "device_enrollment_invalid", Title: "Bad Request", Detail: err.Error()})
	case errors.Is(err, usecase.ErrInvalidDeviceCredential):
		Fail(c, 401, Problem{Code: "device_credential_invalid", Title: "Unauthorized", Detail: err.Error()})
	case errors.Is(err, usecase.ErrInvalidCredentials), errors.Is(err, usecase.ErrInvalidRefreshToken):
		Fail(c, 401, Problem{Code: "auth_invalid", Title: "Unauthorized", Detail: "Thông tin xác thực không hợp lệ."})
	case errors.Is(err, repository.ErrNotFound):
		Fail(c, 404, Problem{Code: "device_not_found", Title: "Not Found", Detail: "Không tìm thấy thiết bị."})
	default:
		Fail(c, 503, Problem{Code: "auth_unavailable", Title: "Service Unavailable", Detail: "Chưa thể xử lý xác thực. Vui lòng thử lại."})
	}
}

func registerDeviceAuthRoutes(protected *gin.RouterGroup, h *AuthHandler, keys *usecase.UserAPIKeyService) {
	protected.POST("/auth/devices", h.EnrollDevice)
	protected.DELETE("/auth/devices/:id", h.RevokeDevice)
	protected.POST("/auth/logout", func(c *gin.Context) {
		principal, ok := advisorPrincipalFromContext(c)
		if ok && principal.CredentialKind == "user_api_key" {
			if keys == nil {
				apiKeyFailure(c, usecase.ErrAPIKeyUnavailable)
				return
			}
			if err := keys.Revoke(c.Request.Context(), principal.OwnerID, principal.CredentialID); err != nil {
				apiKeyFailure(c, err)
				return
			}
			NoContent(c)
			return
		}
		h.Logout(c)
	})
}
