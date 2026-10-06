package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	Notifications repository.NotificationRepository
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
