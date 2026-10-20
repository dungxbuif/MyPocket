package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	Notifications repository.NotificationRepository
	PushTokens    *usecase.PushTokenInteractor
}

// RegisterPushToken godoc
// @Summary Register an iOS push token for the authenticated owner
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body usecase.PushTokenInput true "Provider token registration"
// @Success 200 {object} entity.PushToken
// @Failure 400 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/notifications/devices [post]
func (h *NotificationHandler) RegisterPushToken(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok || h == nil || h.PushTokens == nil {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: "chưa đăng nhập"})
		return
	}
	var in usecase.PushTokenInput
	if c.ShouldBindJSON(&in) != nil {
		Fail(c, http.StatusBadRequest, Problem{Code: "PUSH_TOKEN_INVALID", Title: "Bad Request", Detail: "thông tin thiết bị không hợp lệ"})
		return
	}
	item, err := h.PushTokens.Register(c.Request.Context(), owner, in)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidPushToken) {
			Fail(c, http.StatusBadRequest, Problem{Code: "PUSH_TOKEN_INVALID", Title: "Bad Request", Detail: err.Error()})
			return
		}
		if errors.Is(err, repository.ErrPushTokenConflict) {
			Fail(c, http.StatusConflict, Problem{Code: "PUSH_TOKEN_CONFLICT", Title: "Conflict", Detail: "thiết bị đã được đăng ký cho tài khoản khác"})
			return
		}
		Fail(c, http.StatusInternalServerError, Problem{Code: "PUSH_TOKEN_REGISTER_FAILED", Title: problemTitleInternalServer, Detail: "không thể đăng ký thiết bị"})
		return
	}
	OK(c, item)
}

// RevokePushToken godoc
// @Summary Revoke an iOS push token owned by the authenticated user
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Push token registration ID"
// @Success 204
// @Failure 404 {object} Problem
// @Router /api/v1/notifications/devices/{id} [delete]
func (h *NotificationHandler) RevokePushToken(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok || h == nil || h.PushTokens == nil {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: "chưa đăng nhập"})
		return
	}
	if err := h.PushTokens.Revoke(c.Request.Context(), owner, c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Fail(c, http.StatusNotFound, Problem{Code: "PUSH_TOKEN_NOT_FOUND", Title: problemTitleNotFound, Detail: "không tìm thấy thiết bị"})
			return
		}
		Fail(c, http.StatusInternalServerError, Problem{Code: "PUSH_TOKEN_REVOKE_FAILED", Title: problemTitleInternalServer, Detail: "không thể thu hồi thiết bị"})
		return
	}
	NoContent(c)
}

// ListNotifications godoc
// @Summary List in-app notifications for the authenticated owner
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Notification
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok || h == nil || h.Notifications == nil {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: "chưa đăng nhập"})
		return
	}
	items, err := h.Notifications.List(c.Request.Context(), owner)
	if err != nil {
		Fail(c, http.StatusInternalServerError, Problem{Code: "NOTIFICATION_LOAD_FAILED", Title: problemTitleInternalServer, Detail: "không đọc được thông báo"})
		return
	}
	OK(c, items)
}

// MarkNotificationRead godoc
// @Summary Mark one in-app notification as read
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 204
// @Failure 404 {object} Problem
// @Router /api/v1/notifications/{id}/read [post]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok || h == nil || h.Notifications == nil {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: "chưa đăng nhập"})
		return
	}
	if err := h.Notifications.MarkRead(c.Request.Context(), owner, c.Param("id")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Fail(c, http.StatusNotFound, Problem{Code: "NOTIFICATION_NOT_FOUND", Title: problemTitleNotFound, Detail: "không tìm thấy thông báo"})
			return
		}
		Fail(c, http.StatusInternalServerError, Problem{Code: "NOTIFICATION_UPDATE_FAILED", Title: problemTitleInternalServer, Detail: "không cập nhật được thông báo"})
		return
	}
	NoContent(c)
}
