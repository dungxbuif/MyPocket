package httpapi

import (
	"context"
	"errors"
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	var input feedbackInput
	if c.ShouldBindJSON(&input) != nil {
		feedbackBadRequest(c)
		return
	}
	row, err := h.Service.Create(feedbackContext(c, "user"), owner, usecase.FeedbackInput{Type: input.Type, Title: input.Title, Description: input.Description})
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	Created(c, row)
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
	case errors.Is(err, entity.ErrChangelogVersionRequired), errors.Is(err, entity.ErrChangelogTitleRequired), errors.Is(err, entity.ErrChangelogDescription):
		status, code, detail = http.StatusBadRequest, problemCodeBadRequest, "Thông tin changelog không hợp lệ."
	}
	Fail(c, status, Problem{Code: code, Title: http.StatusText(status), Detail: detail})
}

func feedbackContext(c *gin.Context, actor string) context.Context {
	return usecase.WithAuditContext(c.Request.Context(), requestID(c), actor)
}
