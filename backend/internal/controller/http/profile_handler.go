package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/usecase"
)

type ProfileHandler struct {
	Auth *usecase.AuthInteractor
}

func NewProfileHandler(auth *usecase.AuthInteractor) *ProfileHandler {
	return &ProfileHandler{Auth: auth}
}

// GetProfile godoc
// @Summary Get authenticated user profile
// @Tags Account
// @Produce json
// @Security BearerAuth
// @Success 200 {object} usecase.UserProfile
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, ok := c.Get(contextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "chưa đăng nhập"})
		return
	}
	profile, err := h.Auth.Profile(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy người dùng"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không đọc được profile"})
		return
	}
	c.JSON(http.StatusOK, profile)
}
