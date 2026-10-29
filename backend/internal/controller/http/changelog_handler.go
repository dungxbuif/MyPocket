package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/usecase"
)

type ChangelogHandler struct{ Service *usecase.ChangelogService }

type changelogInput struct {
	FeedbackIDs []string `json:"feedback_ids"`
	Version     string   `json:"version"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
}

// PublishChangelog godoc
// @Summary Publish a changelog and mark referenced in-progress feedback fixed
// @Tags Feedback Agent
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changelog body changelogInput true "Changelog"
// @Success 201 {object} entity.Changelog
// @Router /api/v1/internal/changelog [post]
func (h *ChangelogHandler) Publish(c *gin.Context) {
	if h.Service == nil {
		feedbackFailure(c, errors.New("changelog service unavailable"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	var input changelogInput
	if c.ShouldBindJSON(&input) != nil {
		feedbackBadRequest(c)
		return
	}
	row, err := h.Service.Publish(usecase.WithAuditContext(c.Request.Context(), requestID(c), "agent"), usecase.ChangelogInput{FeedbackIDs: input.FeedbackIDs, Version: input.Version, Title: input.Title, Description: input.Description})
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	Created(c, row)
}

// ListChangelog godoc
// @Summary List published changelog entries
// @Tags Changelog
// @Produce json
// @Param limit query int false "Maximum rows (1-100)"
// @Success 200 {array} entity.Changelog
// @Router /api/v1/changelog [get]
func (h *ChangelogHandler) List(c *gin.Context) {
	if h.Service == nil {
		feedbackFailure(c, errors.New("changelog service unavailable"))
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
	rows, err := h.Service.List(c.Request.Context(), limit)
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	OK(c, rows)
}

// GetChangelog godoc
// @Summary Get one published changelog entry
// @Tags Changelog
// @Produce json
// @Param id path string true "Changelog ID"
// @Success 200 {object} entity.Changelog
// @Router /api/v1/changelog/{id} [get]
func (h *ChangelogHandler) Get(c *gin.Context) {
	if h.Service == nil {
		feedbackFailure(c, errors.New("changelog service unavailable"))
		return
	}
	row, err := h.Service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		feedbackFailure(c, err)
		return
	}
	OK(c, row)
}
