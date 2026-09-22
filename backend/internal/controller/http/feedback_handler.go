package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/gorm"
)

type FeedbackHandler struct{ Service *usecase.FeedbackService }

type feedbackInput struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type feedbackStatusInput struct {
	Status string `json:"status"`
}

// ListFeedback godoc
// @Summary List the authenticated user's feedback
// @Tags Feedback
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Feedback
// @Router /api/v1/feedback [get]
func (h *FeedbackHandler) List(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		feedbackUnauthorized(c)
		return
	}
	if h.Service == nil {
		feedbackFailure(c, errors.New("feedback service unavailable"))
		return
	}
	rows, err := h.Service.List(feedbackContext(c, "user"), owner)
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	OK(c, rows)
}

// CreateFeedback godoc
// @Summary Create owner-scoped feedback
// @Tags Feedback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param feedback body feedbackInput true "Feedback"
// @Success 201 {object} entity.Feedback
// @Router /api/v1/feedback [post]
func (h *FeedbackHandler) Create(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		feedbackUnauthorized(c)
		return
	}
	if h.Service == nil {
		feedbackFailure(c, errors.New("feedback service unavailable"))
		return
	}
	input, err := parseFeedbackInput(c)
	if err != nil {
		feedbackBadRequest(c)
		return
	}
	row, err := h.Service.Create(feedbackContext(c, "user"), owner, input)
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	Created(c, row)
}

func parseFeedbackInput(c *gin.Context) (usecase.FeedbackInput, error) {
	if !strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "multipart/form-data") {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		var input feedbackInput
		if err := c.ShouldBindJSON(&input); err != nil {
			return usecase.FeedbackInput{}, err
		}
		return usecase.FeedbackInput{Type: input.Type, Title: input.Title, Description: input.Description}, nil
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
		return usecase.FeedbackInput{}, err
	}
	input := usecase.FeedbackInput{Type: c.PostForm("type"), Title: c.PostForm("title"), Description: c.PostForm("description")}
	file, header, err := c.Request.FormFile("screenshot")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return input, nil
		}
		return usecase.FeedbackInput{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 1<<20+1))
	if err != nil || len(data) == 0 || len(data) > 1<<20 {
		return usecase.FeedbackInput{}, errors.New("invalid screenshot size")
	}
	detected := http.DetectContentType(data)
	if detected != "image/png" || (header != nil && header.Size > 1<<20) {
		return usecase.FeedbackInput{}, errors.New("invalid screenshot type")
	}
	input.Screenshot, input.ScreenshotMIME, input.ScreenshotSize = bytes.NewReader(data), detected, int64(len(data))
	return input, nil
}

// GetFeedback godoc
// @Summary Get one feedback item owned by the authenticated user
// @Tags Feedback
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feedback ID"
// @Success 200 {object} entity.Feedback
// @Router /api/v1/feedback/{id} [get]
func (h *FeedbackHandler) Get(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		feedbackUnauthorized(c)
		return
	}
	if h.Service == nil {
		feedbackFailure(c, errors.New("feedback service unavailable"))
		return
	}
	row, err := h.Service.Get(feedbackContext(c, "user"), owner, c.Param("id"))
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	OK(c, row)
}

// GetFeedbackScreenshot godoc
// @Summary Get a short-lived private feedback screenshot URL
// @Tags Feedback
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feedback ID"
// @Success 302
// @Router /api/v1/feedback/{id}/screenshot [get]
func (h *FeedbackHandler) Screenshot(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		feedbackUnauthorized(c)
		return
	}
	if h.Service == nil {
		feedbackFailure(c, errors.New("feedback service unavailable"))
		return
	}
	url, err := h.Service.Screenshot(feedbackContext(c, "user"), owner, c.Param("id"))
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	c.Redirect(http.StatusFound, url)
}

// ListAgentFeedback godoc
// @Summary List feedback for the local development agent
// @Tags Feedback Agent
// @Produce json
// @Security BearerAuth
// @Param status query string false "open, triaged, in_progress, fixed, rejected"
// @Param limit query int false "Maximum rows (1-100)"
// @Success 200 {array} entity.Feedback
// @Router /api/v1/agent/feedback [get]
func (h *FeedbackHandler) AgentList(c *gin.Context) {
	if h.Service == nil {
		feedbackFailure(c, errors.New("feedback service unavailable"))
		return
	}
	limit := 0
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			feedbackBadRequest(c)
			return
		}
		if parsed < 1 || parsed > 100 {
			feedbackBadRequest(c)
			return
		}
		limit = parsed
	}
	rows, err := h.Service.AgentList(feedbackContext(c, "agent"), c.Query("status"), limit)
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	OK(c, rows)
}

// GetAgentFeedbackScreenshot godoc
// @Summary Get a short-lived feedback screenshot URL for the local agent
// @Tags Feedback Agent
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feedback ID"
// @Success 302
// @Router /api/v1/agent/feedback/{id}/screenshot [get]
func (h *FeedbackHandler) AgentScreenshot(c *gin.Context) {
	if h.Service == nil {
		feedbackFailure(c, errors.New("feedback service unavailable"))
		return
	}
	url, err := h.Service.AgentScreenshot(feedbackContext(c, "agent"), c.Param("id"))
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	c.Redirect(http.StatusFound, url)
}

// SetFeedbackStatus godoc
// @Summary Advance feedback status through the agent lifecycle
// @Tags Feedback Agent
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Feedback ID"
// @Param status body feedbackStatusInput true "Next status"
// @Success 200 {object} entity.Feedback
// @Router /api/v1/internal/feedback/{id}/status [patch]
func (h *FeedbackHandler) SetStatus(c *gin.Context) {
	if h.Service == nil {
		feedbackFailure(c, errors.New("feedback service unavailable"))
		return
	}
	var input feedbackStatusInput
	if c.ShouldBindJSON(&input) != nil {
		feedbackBadRequest(c)
		return
	}
	row, err := h.Service.SetStatus(feedbackContext(c, "agent"), c.Param("id"), input.Status)
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	OK(c, row)
}

func feedbackUnauthorized(c *gin.Context) {
	Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: "chưa đăng nhập"})
}

func feedbackBadRequest(c *gin.Context) {
	Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: "Dữ liệu phản hồi không hợp lệ."})
}

func feedbackFailure(c *gin.Context, err error) {
	status, code, detail := http.StatusInternalServerError, "FEEDBACK_FAILED", "Không thể xử lý phản hồi. Vui lòng thử lại."
	switch {
	case errors.Is(err, repository.ErrNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		status, code, detail = http.StatusNotFound, "FEEDBACK_NOT_FOUND", "Không tìm thấy phản hồi."
	case errors.Is(err, repository.ErrFeedbackConflict), errors.Is(err, repository.ErrChangelogConflict), errors.Is(err, usecase.ErrFeedbackTransition):
		status, code, detail = http.StatusConflict, "FEEDBACK_CONFLICT", "Trạng thái phản hồi không thể chuyển tiếp."
	case errors.Is(err, repository.ErrFeedbackInvalid), errors.Is(err, usecase.ErrFeedbackAgentInput), errors.Is(err, entity.ErrFeedbackTypeInvalid), errors.Is(err, entity.ErrFeedbackTitleRequired), errors.Is(err, entity.ErrFeedbackDescription), errors.Is(err, entity.ErrFeedbackStatusInvalid):
		status, code, detail = http.StatusBadRequest, problemCodeBadRequest, "Dữ liệu phản hồi không hợp lệ."
	case errors.Is(err, usecase.ErrFeedbackScreenshotInvalid):
		status, code, detail = http.StatusBadRequest, "FEEDBACK_SCREENSHOT_INVALID", "Ảnh chụp màn hình không hợp lệ hoặc vượt quá 1 MiB."
	case errors.Is(err, usecase.ErrFeedbackStorage):
		status, code, detail = http.StatusServiceUnavailable, "FEEDBACK_SCREENSHOT_STORAGE_UNAVAILABLE", "Không thể lưu ảnh chụp màn hình lúc này."
	case errors.Is(err, entity.ErrChangelogVersionRequired), errors.Is(err, entity.ErrChangelogTitleRequired), errors.Is(err, entity.ErrChangelogDescription):
		status, code, detail = http.StatusBadRequest, problemCodeBadRequest, "Thông tin changelog không hợp lệ."
	}
	Fail(c, status, Problem{Code: code, Title: http.StatusText(status), Detail: detail})
}

func feedbackContext(c *gin.Context, actor string) context.Context {
	return usecase.WithAuditContext(c.Request.Context(), requestID(c), actor)
}
