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
// @Router /api/v1/auth/profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, ok := c.Get(contextUserIDKey)
	if !ok {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: "chưa đăng nhập"})
		return
	}
	profile, err := h.Auth.Profile(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			Fail(c, http.StatusNotFound, Problem{Code: problemCodeUserNotFound, Title: "Not Found", Detail: "không tìm thấy người dùng"})
			return
		}
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeProfileLoadFailed, Title: problemTitleInternalServer, Detail: "không đọc được profile"})
		return
	}
	OK(c, profile)
}
