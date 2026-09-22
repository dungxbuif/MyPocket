package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

const advisorRoute = "/api/v1/ai/advisor"

type AdvisorHandler struct {
	Service *usecase.AdvisorService
	Users   repository.UserRepository
	Finance *usecase.FinanceQueryService
}

type advisorMessageInput struct {
	ClientRequestID string `json:"client_request_id"`
	Text            string `json:"text"`
}

type advisorSubmitResponse struct {
	RunID          string `json:"run_id"`
	ConversationID string `json:"conversation_id"`
	Text           string `json:"text"`
	Status         string `json:"status"`
}

func (r *Router) RegisterAdvisorRoutes(h *AdvisorHandler) {
	g := r.Engine.Group(advisorRoute)
	g.Use(r.AuthMiddleware.RequireAdvisorAuth)
	g.GET("/capabilities", h.Capabilities)
	g.GET("/overview", h.Overview)
	g.GET("/conversation/messages", h.Messages)
	g.POST("/messages", h.MessagesPost)
	g.GET("/runs/:id", h.Run)
	g.POST("/runs/:id/cancel", h.Cancel)
}

// AdvisorCapabilities godoc
// @Summary Read Finance Assistant capabilities
// @Tags AI Advisor
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any
// @Router /api/v1/ai/advisor/capabilities [get]
func (h *AdvisorHandler) Capabilities(c *gin.Context) {
	if !requireAdvisorScope(c, entity.APIKeyScopeAdvisorRead) {
		return
	}
	OK(c, map[string]any{"enabled": h != nil && h.Service != nil && h.Service.Enabled(), "write": false, "parts_version": 1, "tools": []string{"search_transactions", "get_transaction", "get_finance_summary", "compare_spending_periods", "get_wallet_balances", "get_budget_progress", "get_goal_progress", "get_jar_progress"}})
}

// AdvisorOverview godoc
// @Summary Read Finance Assistant overview without calling the model
// @Tags AI Advisor
// @Produce json
// @Security BearerAuth
// @Param month query string false "Account-local calendar month YYYY-MM"
// @Success 200 {object} map[string]any
// @Router /api/v1/ai/advisor/overview [get]
func (h *AdvisorHandler) Overview(c *gin.Context) {
	if !requireAdvisorScope(c, entity.APIKeyScopeAdvisorRead) {
		return
	}
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Finance == nil || h.Users == nil {
		advisorFailure(c, errors.New("advisor overview unavailable"))
		return
	}
	user, err := h.Users.FindByID(owner)
	if err != nil {
		advisorFailure(c, err)
		return
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil || strings.TrimSpace(user.Timezone) == "" || user.Timezone == "Local" {
		advisorBadRequest(c)
		return
	}
	month := strings.TrimSpace(c.Query("month"))
	if month == "" {
		month = time.Now().In(location).Format("2006-01")
	}
	start, err := time.ParseInLocation("2006-01", month, location)
	if err != nil || start.Format("2006-01") != month {
		advisorBadRequest(c)
		return
	}
	end := start.AddDate(0, 1, -1)
	filter, err := entity.NormalizeFinanceFilter(entity.FinanceFilter{Range: entity.DateRange{From: start.Format("2006-01-02"), To: end.Format("2006-01-02")}}, location)
	if err != nil {
		advisorBadRequest(c)
		return
	}
	result, err := h.Finance.Execute(c.Request.Context(), owner, entity.NormalizedQuery{Key: "overview", Kind: "get_finance_summary", Filter: filter})
	if err != nil {
		advisorFailure(c, err)
		return
	}
	var view any
	if len(result.View) > 0 && json.Unmarshal(result.View, &view) != nil {
		advisorFailure(c, errors.New("advisor overview malformed"))
		return
	}
	OK(c, map[string]any{"view_kind": result.ViewKind, "view": view, "source": result.Source, "status": result.Status})
}

// AdvisorMessages godoc
// @Summary Read persisted Finance Assistant messages
// @Tags AI Advisor
// @Produce json
// @Security BearerAuth
// @Param before_seq query int false "Read messages before sequence"
// @Success 200 {array} entity.AdvisorMessage
// @Router /api/v1/ai/advisor/conversation/messages [get]
func (h *AdvisorHandler) Messages(c *gin.Context) {
	if !requireAdvisorScope(c, entity.APIKeyScopeAdvisorRead) {
		return
	}
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil || h.Service.Store == nil {
		advisorFailure(c, errors.New("advisor unavailable"))
		return
	}
	conversationID := strings.TrimSpace(c.Query("conversation_id"))
	if conversationID == "" {
		advisorBadRequest(c)
		return
	}
	beforeSeq := int64(0)
	if rawBefore := strings.TrimSpace(c.Query("before_seq")); rawBefore != "" {
		parsed, parseErr := strconv.ParseInt(rawBefore, 10, 64)
		if parseErr != nil || parsed <= 0 {
			advisorBadRequest(c)
			return
		}
		beforeSeq = parsed
	}
	rows, err := h.Service.Store.ListMessages(c.Request.Context(), owner, conversationID, beforeSeq, 50)
	if err != nil {
		advisorFailure(c, err)
		return
	}
	OK(c, rows)
}

// AdvisorMessagesPost godoc
// @Summary Submit a Finance Assistant read-only question
// @Tags AI Advisor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param message body advisorMessageInput true "Question"
// @Success 200 {object} advisorSubmitResponse
// @Router /api/v1/ai/advisor/messages [post]
func (h *AdvisorHandler) MessagesPost(c *gin.Context) {
	if !requireAdvisorScope(c, entity.APIKeyScopeAdvisorChat) {
		return
	}
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil || h.Users == nil {
		advisorFailure(c, errors.New("advisor unavailable"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16*1024)
	var input advisorMessageInput
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Text) == "" || len(input.Text) > 8192 {
		advisorBadRequest(c)
		return
	}
	if strings.TrimSpace(input.ClientRequestID) == "" {
		input.ClientRequestID = uuid.NewString()
	}
	if len(input.ClientRequestID) > 128 {
		advisorBadRequest(c)
		return
	}
	user, err := h.Users.FindByID(owner)
	if err != nil {
		advisorFailure(c, err)
		return
	}
	hash := sha256.Sum256([]byte(input.Text))
	principal := advisorPrincipal(c, owner)
	principal.Timezone = user.Timezone
	answer, err := h.Service.Submit(c.Request.Context(), principal, repository.AdvisorRunInput{OwnerID: owner, CredentialKind: principal.CredentialKind, CredentialID: principal.CredentialID, ExpiresAt: timePtr(principal.ExpiresAt), RequestID: input.ClientRequestID, PayloadHash: hex.EncodeToString(hash[:]), Parts: []entity.AdvisorPart{{Type: "text", Text: input.Text}}})
	if err != nil {
		advisorFailure(c, err)
		return
	}
	run, err := h.Service.Store.GetRunByRequest(c.Request.Context(), owner, input.ClientRequestID)
	if err != nil {
		advisorFailure(c, err)
		return
	}
	OK(c, advisorSubmitResponse{RunID: run.ID, ConversationID: run.ConversationID, Text: answer.Text, Status: run.Status})
}

// AdvisorRun godoc
// @Summary Read one Finance Assistant run
// @Tags AI Advisor
// @Produce json
// @Security BearerAuth
// @Param id path string true "Run ID"
// @Success 200 {object} entity.AdvisorRun
// @Router /api/v1/ai/advisor/runs/{id} [get]
func (h *AdvisorHandler) Run(c *gin.Context) {
	if !requireAdvisorScope(c, entity.APIKeyScopeAdvisorRead) {
		return
	}
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil || h.Service.Store == nil {
		advisorFailure(c, errors.New("advisor unavailable"))
		return
	}
	run, err := h.Service.Store.GetRun(c.Request.Context(), owner, c.Param("id"))
	if err != nil {
		advisorFailure(c, err)
		return
	}
	OK(c, run)
}

// AdvisorCancel godoc
// @Summary Cancel a queued or running Finance Assistant run
// @Tags AI Advisor
// @Produce json
// @Security BearerAuth
// @Param id path string true "Run ID"
// @Success 204
// @Router /api/v1/ai/advisor/runs/{id}/cancel [post]
func (h *AdvisorHandler) Cancel(c *gin.Context) {
	if !requireAdvisorScope(c, entity.APIKeyScopeAdvisorChat) {
		return
	}
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil || h.Service.Store == nil {
		advisorFailure(c, errors.New("advisor unavailable"))
		return
	}
	if err := h.Service.Cancel(c.Request.Context(), owner, c.Param("id")); err != nil {
		advisorFailure(c, err)
		return
	}
	NoContent(c)
}

func advisorBadRequest(c *gin.Context) {
	Fail(c, http.StatusBadRequest, Problem{Code: "advisor_invalid", Title: "Invalid advisor request", Detail: "câu hỏi hoặc tham số không hợp lệ"})
}

func requireAdvisorScope(c *gin.Context, scope string) bool {
	value, exists := c.Get(contextAdvisorPrincipalKey)
	principal, ok := value.(usecase.Principal)
	if !exists || !ok || principal.CredentialKind != "user_api_key" || principal.HasScope(scope) {
		return true
	}
	Fail(c, http.StatusForbidden, Problem{Code: "api_key_scope_denied", Title: "Forbidden", Detail: "api key không có quyền cho thao tác này"})
	c.Abort()
	return false
}

func advisorPrincipal(c *gin.Context, ownerID string) usecase.Principal {
	if value, ok := c.Get(contextAdvisorPrincipalKey); ok {
		if principal, valid := value.(usecase.Principal); valid {
			return principal
		}
	}
	return usecase.Principal{OwnerID: ownerID, CredentialKind: "session", CredentialID: requestID(c)}
}

func timePtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func advisorFailure(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "advisor_failed"
	if errors.Is(err, usecase.ErrAdvisorPrincipalInvalid) {
		status, code = http.StatusUnauthorized, "advisor_auth_invalid"
	} else if errors.Is(err, usecase.ErrAdvisorProviderUnavailable) {
		status, code = http.StatusServiceUnavailable, "advisor_unavailable"
	} else if errors.Is(err, repository.ErrAdvisorBusy) {
		status, code = http.StatusConflict, "advisor_busy"
	} else if errors.Is(err, repository.ErrAdvisorRequestConflict) {
		status, code = http.StatusConflict, "advisor_request_conflict"
	} else if errors.Is(err, repository.ErrNotFound) {
		status, code = http.StatusNotFound, "advisor_not_found"
	} else if errors.Is(err, usecase.ErrAdvisorToolInvalid) {
		status, code = http.StatusBadRequest, "advisor_invalid"
	}
	Fail(c, status, Problem{Code: code, Title: "Finance Assistant request failed", Detail: "không thể xử lý yêu cầu trợ lý tài chính"})
}
