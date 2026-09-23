package httpapi

import (
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	aiEntryRoute      = "/api/v1/ai/entry"
	aiInvalidCode     = "ai_entry_invalid"
	aiConflictCode    = "ai_entry_conflict"
	aiUnavailableCode = "ai_entry_unavailable"
	aiNotFoundCode    = "ai_entry_not_found"
	aiFailureCode     = "ai_entry_failed"
	aiMaxBody         = 101 * 1024 * 1024
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
	g.POST("/process", h.Process)
	g.GET("/requests/:request_id", h.Request)
	g.PATCH("/proposals/:id", h.EditProposal)
	g.POST("/proposals/:id/approve", h.ApproveProposal)
	g.POST("/proposals/:id/reject", h.RejectProposal)
	r.Engine.GET("/api/v1/transactions/:id/attachments/:attachmentId/download", r.AuthMiddleware.RequireAuth, h.DownloadAttachment)
}

// DownloadAttachment godoc
// @Summary Download a private attachment linked to an owned transaction
// @Tags Transactions
// @Produce application/octet-stream
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Param attachmentId path string true "Attachment ID"
// @Success 302 "Redirect to a short-lived private object URL"
// @Failure 404 {object} Problem
// @Router /api/v1/transactions/{id}/attachments/{attachmentId}/download [get]
func (h *AIEntryHandler) DownloadAttachment(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	attachment, err := h.Service.Entries.AttachmentForTransaction(c.Request.Context(), owner, c.Param("id"), c.Param("attachmentId"))
	if err != nil {
		aiEntryFail(c, gorm.ErrRecordNotFound)
		return
	}
	url, err := h.Service.Storage.SignedGet(c.Request.Context(), attachment.ObjectKey)
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	c.Redirect(http.StatusFound, url)
}

type aiEntryProcessResponse struct {
	ID         string                   `json:"id"`
	Processing bool                     `json:"processing"`
	Error      string                   `json:"error,omitempty"`
	ErrorCode  string                   `json:"error_code,omitempty"`
	Reply      string                   `json:"reply,omitempty"`
	Proposals  []entity.AIEntryProposal `json:"proposals"`
}

func processResponse(process *entity.AIEntrySession) aiEntryProcessResponse {
	return aiEntryProcessResponse{ID: process.ID, Processing: process.Processing, Error: process.Error, ErrorCode: process.ErrorCode, Reply: process.Reply, Proposals: process.Proposals}
}

// Process godoc
// @Summary Process one text and optional file submission into reviewable transaction proposals
// @Tags AI Entry
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param request_id formData string true "Idempotency UUID"
// @Param text formData string false "Transaction description"
// @Param timezone formData string true "IANA timezone"
// @Param files formData file false "JPEG, PNG, or PDF, up to twenty files"
// @Success 200 {object} Response
// @Failure 400 {object} Problem
// @Failure 503 {object} Problem
// @Router /api/v1/ai/entry/process [post]
func (h *AIEntryHandler) Process(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, aiMaxBody)
	if err := c.Request.ParseMultipartForm(6 * 1024 * 1024); err != nil {
		aiEntryFail(c, port.ErrAIInvalid)
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	in := usecase.AIEntryMessageInput{RequestID: c.PostForm("request_id"), Text: c.PostForm("text"), Timezone: c.PostForm("timezone")}
	parsedImages, err := parseAIEntryFiles(c.Request.MultipartForm.File["files"])
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	in.Images = parsedImages
	process, err := h.Service.Process(c.Request.Context(), owner, in)
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	OK(c, processResponse(process))
}

func parseAIEntryFiles(files []*multipart.FileHeader) ([]entity.AIImage, error) {
	if len(files) > 20 {
		return nil, port.ErrAIInvalid
	}
	images := make([]entity.AIImage, 0, len(files))
	for _, header := range files {
		if header.Size <= 0 || header.Size > 5*1024*1024 || len(header.Filename) > 255 {
			return nil, port.ErrAIInvalid
		}
		file, err := header.Open()
		if err != nil {
			return nil, port.ErrAIInvalid
		}
		data, readErr := io.ReadAll(io.LimitReader(file, 5*1024*1024+1))
		closeErr := file.Close()
		if readErr != nil || closeErr != nil || int64(len(data)) != header.Size || len(data) > 5*1024*1024 {
			return nil, port.ErrAIInvalid
		}
		mime := http.DetectContentType(data)
		if mime != "image/jpeg" && mime != "image/png" && mime != "application/pdf" {
			return nil, port.ErrAIInvalid
		}
		images = append(images, entity.AIImage{Name: header.Filename, MIMEType: mime, Base64: base64.StdEncoding.EncodeToString(data)})
	}
	return images, nil
}

// Request godoc
// @Summary Read one-shot AI process status and proposals by idempotency key
// @Tags AI Entry
// @Produce json
// @Security BearerAuth
// @Param request_id path string true "Idempotency UUID"
// @Success 200 {object} Response
// @Router /api/v1/ai/entry/requests/{request_id} [get]
func (h *AIEntryHandler) Request(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	process, err := h.Service.Entries.Session(c.Request.Context(), owner, c.Param("request_id"))
	if err != nil {
		aiEntryFail(c, err)
		return
	}
	OK(c, processResponse(process))
}

// Capabilities godoc
// @Summary AI entry configuration availability
// @Tags AI Entry
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response
// @Router /api/v1/ai/entry/capabilities [get]
func (h *AIEntryHandler) Capabilities(c *gin.Context) {
	OK(c, gin.H{"ai_configured": h.Service.Extractor.Configured(), "ocr_configured": h.Service.Extractor.OCRConfigured(), "files_configured": h.Service.Extractor.OCRConfigured() && h.Service.Storage != nil})
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
	case errors.Is(err, gorm.ErrRecordNotFound):
		status, code, detail = http.StatusNotFound, aiNotFoundCode, "Không tìm thấy yêu cầu hoặc đề xuất."
	case errors.Is(err, port.ErrAIConflict), errors.Is(err, port.ErrAIBusy):
		status, code, detail = http.StatusConflict, aiConflictCode, err.Error()
	case errors.Is(err, port.ErrAIInvalid):
		status, code, detail = http.StatusBadRequest, aiInvalidCode, err.Error()
	case errors.Is(err, usecase.ErrAIUnavailable), errors.Is(err, usecase.ErrAIProvider):
		status, code, detail = http.StatusServiceUnavailable, aiUnavailableCode, err.Error()
	case errors.Is(err, usecase.ErrAIStorage):
		status, code, detail = http.StatusServiceUnavailable, "ai_entry_storage_failed", err.Error()
	}
	Fail(c, status, Problem{Code: code, Title: http.StatusText(status), Detail: detail})
}
