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
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: "chưa đăng nhập"})
		return
	}
	home, err := h.Auth.Home(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			Fail(c, http.StatusNotFound, Problem{Code: problemCodeUserNotFound, Title: "Not Found", Detail: "không tìm thấy người dùng"})
			return
		}
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeHomeLoadFailed, Title: problemTitleInternalServer, Detail: "không đọc được trang chủ"})
		return
	}
	OK(c, home)
}
