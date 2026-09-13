package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/usecase"
)

type HomeHandler struct {
	Auth *usecase.AuthInteractor
}

func NewHomeHandler(auth *usecase.AuthInteractor) *HomeHandler {
	return &HomeHandler{Auth: auth}
}

// GetHome godoc
// @Summary Get authenticated home summary
// @Tags Account
// @Produce json
// @Security BearerAuth
// @Success 200 {object} usecase.HomeOutput
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/home [get]
func (h *HomeHandler) GetHome(c *gin.Context) {
	userID, ok := c.Get(contextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "chưa đăng nhập"})
		return
	}
	home, err := h.Auth.Home(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy người dùng"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không đọc được trang chủ"})
		return
	}
	c.JSON(http.StatusOK, home)
}
