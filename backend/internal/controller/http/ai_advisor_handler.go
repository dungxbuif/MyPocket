package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
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
	Service      *usecase.AdvisorService
	Users        repository.UserRepository
	Finance      *usecase.FinanceQueryService
	Wallets      repository.WalletRepository
	Jars         repository.JarRepository
	Budgets      repository.BudgetRepository
	Categories   repository.CategoryRepository
	Transactions repository.TransactionRepository
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

type advisorConfirmActionInput struct {
	ConversationID string `json:"conversation_id"`
	Generation     int64  `json:"generation"`
}

func (r *Router) RegisterAdvisorRoutes(h *AdvisorHandler) {
	g := r.Engine.Group(advisorRoute)
	g.Use(r.AuthMiddleware.RequireAdvisorAuth)
	g.GET("/capabilities", h.Capabilities)
	g.GET("/overview", h.Overview)
	g.GET("/conversation/messages", h.Messages)
	g.POST("/conversation/reset", h.ResetConversation)
	g.POST("/messages", h.MessagesPost)
	g.POST("/actions/:action_id/confirm", h.ConfirmAction)
	g.GET("/runs/:id", h.Run)
	g.POST("/runs/:id/cancel", h.Cancel)
}

// ConfirmAction godoc
// @Summary Confirm a persisted Finance Assistant action
// @Tags AI Advisor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param action_id path string true "Persisted action identity"
// @Param action body advisorConfirmActionInput true "Conversation binding"
// @Success 201 {object} map[string]any
// @Success 200 {object} map[string]any
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Failure 503 {object} Problem
// @Router /api/v1/ai/advisor/actions/{action_id}/confirm [post]
func (h *AdvisorHandler) ConfirmAction(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil || h.Service.Store == nil {
		advisorFailure(c, errors.New("advisor action unavailable"))
		return
	}
	actionID := strings.TrimSpace(c.Param("action_id"))
	if !validAdvisorActionID(actionID) {
		advisorBadRequest(c)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8*1024)
	var input advisorConfirmActionInput
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || strings.TrimSpace(input.ConversationID) == "" || input.Generation <= 0 || len(input.ConversationID) > 128 {
		advisorBadRequest(c)
		return
	}
	messages, err := h.Service.Store.ListMessages(c.Request.Context(), owner, input.ConversationID, 0, 50)
	if err != nil {
		advisorFailure(c, err)
		return
	}
	view, found, stale := findPreparedProposal(messages, actionID, input.ConversationID, input.Generation)
	if stale {
		Fail(c, http.StatusConflict, Problem{Code: "advisor_action_stale", Title: problemTitleBadRequest, Detail: "Đề xuất trợ lý đã hết hạn hoặc thuộc thế hệ hội thoại khác."})
		return
	}
	if !found {
		Fail(c, http.StatusNotFound, Problem{Code: "advisor_action_not_found", Title: problemTitleNotFound, Detail: "Không tìm thấy đề xuất trợ lý."})
		return
	}
	// Reuse the ordinary domain handler so validation and the durable mutation
	// ledger remain in one place. The proposal view is the source of truth;
	// request fields cannot override it.
	payload, err := json.Marshal(view)
	if err != nil {
		advisorFailure(c, err)
		return
	}
	request := c.Request.Clone(c.Request.Context())
	request.Body = io.NopCloser(bytes.NewReader(payload))
	request.ContentLength = int64(len(payload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", actionID)
	c.Request = request
	switch stringValue(view["action"]) {
	case "create_wallet":
		if h.Wallets == nil {
			advisorFailure(c, errors.New("wallet action unavailable"))
			return
		}
		(&WalletHandler{Wallets: h.Wallets}).CreateWallet(c)
	case "create_jar":
		if h.Jars == nil {
			advisorFailure(c, errors.New("jar action unavailable"))
			return
		}
		(&JarHandler{Jars: h.Jars, Users: h.Users}).Create(c)
	case "update_jar":
		if h.Jars == nil {
			advisorFailure(c, errors.New("jar action unavailable"))
			return
		}
		jarID := strings.TrimSpace(stringValue(view["jar_id"]))
		month := strings.TrimSpace(stringValue(view["month"]))
		if _, err := uuid.Parse(jarID); err != nil || month == "" {
			advisorBadRequest(c)
			return
		}
		// UpdateMonth is the canonical owner-scoped mutation path. Populate the
		// route params directly because this confirmation endpoint dispatches
		// without issuing a second HTTP request.
		c.Params = gin.Params{{Key: "jar_id", Value: jarID}, {Key: "month", Value: month}}
		(&JarHandler{Jars: h.Jars, Users: h.Users}).UpdateMonth(c)
	case "create_budget":
		if h.Budgets == nil {
			advisorFailure(c, errors.New("budget action unavailable"))
			return
		}
		(&BudgetHandler{Budgets: h.Budgets, Wallets: h.Wallets, Categories: h.Categories, Transactions: h.Transactions, Users: h.Users}).CreateBudget(c)
	case "update_budget":
		if h.Budgets == nil {
			advisorFailure(c, errors.New("budget action unavailable"))
			return
		}
		budgetID := strings.TrimSpace(stringValue(view["budget_id"]))
		if _, err := uuid.Parse(budgetID); err != nil {
			advisorBadRequest(c)
			return
		}
		c.Params = gin.Params{{Key: "id", Value: budgetID}}
		(&BudgetHandler{Budgets: h.Budgets, Wallets: h.Wallets, Categories: h.Categories, Transactions: h.Transactions, Users: h.Users}).UpdateBudget(c)
	default:
		advisorBadRequest(c)
	}
}

func validAdvisorActionID(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if character < 32 || character > 126 {
			return false
		}
	}
	return true
}

func findPreparedProposal(messages []entity.AdvisorMessage, actionID, conversationID string, generation int64) (map[string]any, bool, bool) {
	for _, message := range messages {
		if message.Role != entity.AdvisorRoleAssistant || message.ConversationID != conversationID {
			continue
		}
		for _, part := range message.Parts {
			if part.Type != "wallet_confirmation" && part.Type != "jar_confirmation" && part.Type != "budget_confirmation" {
				continue
			}
			encoded, err := json.Marshal(part.Data)
			if err != nil {
				continue
			}
			var envelope map[string]any
			if json.Unmarshal(encoded, &envelope) != nil {
				continue
			}
			view, ok := envelope["view"].(map[string]any)
			if !ok || strings.TrimSpace(stringValue(view["action_id"])) != actionID || stringValue(view["conversation_id"]) != conversationID {
				continue
			}
			if stringValue(view["action"]) != "create_wallet" && stringValue(view["action"]) != "create_jar" && stringValue(view["action"]) != "update_jar" && stringValue(view["action"]) != "create_budget" && stringValue(view["action"]) != "update_budget" {
				continue
			}
			// Bind the action to both the persisted message generation and the
			// generation echoed in its view. A stale generation is a conflict,
			// rather than an ambiguous not-found that invites a retry with new state.
			if message.Generation != generation || numberValue(view["generation"]) != generation {
				return nil, false, true
			}
			return view, true, false
		}
	}
	return nil, false, false
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func numberValue(value any) int64 {
	switch number := value.(type) {
	case int64:
		return number
	case float64:
		return int64(number)
	case int:
		return int64(number)
	default:
		return 0
	}
}

// ResetConversation godoc
// @Summary Explicitly clear the owner's Finance Assistant conversation
// @Tags AI Advisor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string true "Stable reset request key"
// @Success 200 {object} entity.AdvisorResetResult
// @Router /api/v1/ai/advisor/conversation/reset [post]
func (h *AdvisorHandler) ResetConversation(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	requestID := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if requestID == "" || len(requestID) > 128 {
		advisorBadRequest(c)
		return
	}
	if h == nil || h.Service == nil || h.Service.Store == nil {
		advisorFailure(c, errors.New("advisor unavailable"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4*1024)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		advisorBadRequest(c)
		return
	}
	hash := sha256.Sum256(body)
	result, err := h.Service.ResetConversation(c.Request.Context(), owner, requestID, hex.EncodeToString(hash[:]))
	if err != nil {
		advisorFailure(c, err)
		return
	}
	OK(c, result)
}

// AdvisorCapabilities godoc
// @Summary Read Finance Assistant capabilities
// @Tags AI Advisor
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]any
// @Router /api/v1/ai/advisor/capabilities [get]
func (h *AdvisorHandler) Capabilities(c *gin.Context) {
	OK(c, map[string]any{"enabled": h != nil && h.Service != nil && h.Service.Enabled(), "write": false, "parts_version": 1, "tools": []string{"search_transactions", "get_transaction", "get_finance_summary", "compare_spending_periods", "get_wallet_balances", "get_budget_progress", "get_goal_progress", "get_jar_progress", "get_product_help", "propose_create_wallet", "propose_create_jar", "propose_update_jar", "propose_create_budget", "propose_update_budget"}, "actions": []string{"create_wallet", "create_jar", "update_jar", "create_budget", "update_budget"}})
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
// @Description The response keeps the message array in data. The envelope meta may include server-authoritative conversation_id, generation, last_user_message_at, expires_at and server_time.
// @Success 200 {array} entity.AdvisorMessage
// @Router /api/v1/ai/advisor/conversation/messages [get]
func (h *AdvisorHandler) Messages(c *gin.Context) {
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
	var meta any
	if metadataStore, ok := h.Service.Store.(repository.AdvisorHistoryMetadataRepository); ok {
		historyMeta, metaErr := metadataStore.AdvisorHistoryMeta(c.Request.Context(), owner, conversationID)
		if metaErr != nil {
			advisorFailure(c, metaErr)
			return
		}
		meta = historyMeta
	}
	OKWithMeta(c, rows, meta)
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
	} else if errors.Is(err, repository.ErrAdvisorPurged) {
		status, code = http.StatusGone, "advisor_history_expired"
	} else if errors.Is(err, repository.ErrNotFound) {
		status, code = http.StatusNotFound, "advisor_not_found"
	} else if errors.Is(err, usecase.ErrAdvisorToolInvalid) {
		status, code = http.StatusBadRequest, "advisor_invalid"
	}
	Fail(c, status, Problem{Code: code, Title: "Finance Assistant request failed", Detail: "không thể xử lý yêu cầu trợ lý tài chính"})
}
