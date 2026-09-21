package httpapi

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/gorm"
	"net/http"
)

const (
	aiEntryRoute      = "/api/v1/ai/entry"
	aiInvalidCode     = "ai_entry_invalid"
	aiConflictCode    = "ai_entry_conflict"
	aiUnavailableCode = "ai_entry_unavailable"
	aiNotFoundCode    = "ai_entry_not_found"
	aiFailureCode     = "ai_entry_failed"
	aiMaxBody         = 22 * 1024 * 1024
)

type AIEntryHandler struct{ Service *usecase.AIEntryService }
type aiProposalEdit struct {
	Version int                 `json:"version"`
	Draft   entity.AIEntryDraft `json:"draft"`
}
type aiProposalDecision struct {
	Version int `json:"version"`
}

func (r *Router) RegisterAIEntryRoutes(h *AIEntryHandler) {
	g := r.Engine.Group(aiEntryRoute)
	g.Use(r.AuthMiddleware.RequireAuth)
	g.GET("/capabilities", h.Capabilities)
	g.POST("/sessions", h.CreateSession)
	g.GET("/sessions/latest", h.LatestSession)
	g.GET("/sessions/:id", h.Session)
	g.POST("/sessions/:id/messages", h.Message)
	g.PATCH("/proposals/:id", h.EditProposal)
	g.POST("/proposals/:id/approve", h.ApproveProposal)
	g.POST("/proposals/:id/reject", h.RejectProposal)
}

// Capabilities godoc
// @Summary AI entry configuration availability
// @Tags AI Entry
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response
// @Router /api/v1/ai/entry/capabilities [get]
func (h *AIEntryHandler) Capabilities(c *gin.Context) {
	OK(c, gin.H{"ai_configured": h.Service.Extractor.Configured(), "ocr_configured": h.Service.Extractor.OCRConfigured()})
}

// CreateSession godoc
// @Summary Create an owner-scoped AI entry chat
// @Tags AI Entry
// @Produce json
// @Security BearerAuth
// @Success 201 {object} entity.AIEntrySession
// @Router /api/v1/ai/entry/sessions [post]
func (h *AIEntryHandler) CreateSession(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	s, err := h.Service.Entries.CreateSession(c.Request.Context(), owner)
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	Created(c, s)
}

// LatestSession godoc
// @Summary Resume latest AI entry chat
// @Tags AI Entry
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.AIEntrySession
// @Router /api/v1/ai/entry/sessions/latest [get]
func (h *AIEntryHandler) LatestSession(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	s, err := h.Service.Entries.LatestSession(c.Request.Context(), owner)
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	OK(c, s)
}

// Session godoc
// @Summary Read AI entry chat and proposals
// @Tags AI Entry
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} entity.AIEntrySession
// @Router /api/v1/ai/entry/sessions/{id} [get]
func (h *AIEntryHandler) Session(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	s, err := h.Service.Entries.Session(c.Request.Context(), owner, c.Param("id"))
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	OK(c, s)
}

// Message godoc
// @Summary Extract reviewable drafts from text and OCR-first images
// @Tags AI Entry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param message body usecase.AIEntryMessageInput true "Message"
// @Success 200 {object} entity.AIEntrySession
// @Failure 400 {object} Problem
// @Failure 409 {object} Problem
// @Failure 503 {object} Problem
// @Router /api/v1/ai/entry/sessions/{id}/messages [post]
func (h *AIEntryHandler) Message(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, aiMaxBody)
	var in usecase.AIEntryMessageInput
	if err := c.ShouldBindJSON(&in); err != nil {
		aiEntryFail(c, port.ErrAIInvalid)
		return
	}
	s, err := h.Service.Send(c.Request.Context(), owner, c.Param("id"), in)
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	OK(c, s)
}

// EditProposal godoc
// @Summary Edit a pending AI proposal without changing ledger
// @Tags AI Entry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Proposal ID"
// @Param proposal body aiProposalEdit true "Draft and expected version"
// @Success 200 {object} entity.AIEntryProposal
// @Router /api/v1/ai/entry/proposals/{id} [patch]
func (h *AIEntryHandler) EditProposal(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	var in aiProposalEdit
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 65536)
	if c.ShouldBindJSON(&in) != nil || in.Version < 1 {
		aiEntryFail(c, port.ErrAIInvalid)
		return
	}
	p, err := h.Service.Entries.EditProposal(c.Request.Context(), owner, c.Param("id"), in.Version, in.Draft)
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	OK(c, p)
}

// ApproveProposal godoc
// @Summary Approve a proposal and atomically create its ledger transaction once
// @Tags AI Entry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Proposal ID"
// @Param decision body aiProposalDecision true "Expected version"
// @Success 200 {object} entity.AIEntryProposal
// @Router /api/v1/ai/entry/proposals/{id}/approve [post]
func (h *AIEntryHandler) ApproveProposal(c *gin.Context) { h.decide(c, true) }

// RejectProposal godoc
// @Summary Reject a proposal without changing ledger
// @Tags AI Entry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Proposal ID"
// @Param decision body aiProposalDecision true "Expected version"
// @Success 200 {object} entity.AIEntryProposal
// @Router /api/v1/ai/entry/proposals/{id}/reject [post]
func (h *AIEntryHandler) RejectProposal(c *gin.Context) { h.decide(c, false) }
func (h *AIEntryHandler) decide(c *gin.Context, approve bool) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	var in aiProposalDecision
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
	if c.ShouldBindJSON(&in) != nil || in.Version < 1 {
		aiEntryFail(c, port.ErrAIInvalid)
		return
	}
	p, err := h.Service.Entries.DecideProposal(c.Request.Context(), owner, c.Param("id"), in.Version, approve)
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	OK(c, p)
}
func aiEntryFail(c *gin.Context, err error) {
	status, code, detail := http.StatusInternalServerError, aiFailureCode, "Không xử lý được yêu cầu. Vui lòng thử lại."
	switch {
	case errors.Is(err, port.ErrAIRateLimited):
		status, code, detail = http.StatusTooManyRequests, "ai_entry_rate_limited", err.Error()
	case errors.Is(err, gorm.ErrRecordNotFound):
		status, code, detail = http.StatusNotFound, aiNotFoundCode, "Không tìm thấy phiên hoặc đề xuất."
	case errors.Is(err, port.ErrAIConflict), errors.Is(err, port.ErrAIBusy):
		status, code, detail = http.StatusConflict, aiConflictCode, err.Error()
	case errors.Is(err, port.ErrAIInvalid), errors.Is(err, port.ErrAISessionFull):
		status, code, detail = http.StatusBadRequest, aiInvalidCode, err.Error()
	case errors.Is(err, usecase.ErrAIUnavailable), errors.Is(err, usecase.ErrAIProvider):
		status, code, detail = http.StatusServiceUnavailable, aiUnavailableCode, err.Error()
	}
	Fail(c, status, Problem{Code: code, Title: http.StatusText(status), Detail: detail})
}
