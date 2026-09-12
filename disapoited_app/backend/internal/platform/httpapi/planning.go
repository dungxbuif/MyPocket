package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/planning"
	"mypocket/internal/platform/config"
)

type budgetsResponse struct {
	Status        string                    `json:"status"`
	Budgets       []planning.BudgetProgress `json:"budgets"`
	CorrelationID string                    `json:"correlation_id"`
}

type budgetResponse struct {
	Status        string          `json:"status"`
	Budget        planning.Budget `json:"budget"`
	CorrelationID string          `json:"correlation_id"`
}

type budgetRequest struct {
	BaseVersion int64                     `json:"base_version"`
	Name        string                    `json:"name"`
	PeriodType  planning.BudgetPeriodType `json:"period_type"`
	AmountVND   int64                     `json:"amount_vnd"`
	CategoryIDs []string                  `json:"category_ids"`
	CustomStart string                    `json:"custom_start"`
	CustomEnd   string                    `json:"custom_end"`
}

type eventsResponse struct {
	Status        string                  `json:"status"`
	Events        []planning.EventSummary `json:"events"`
	CorrelationID string                  `json:"correlation_id"`
}

type eventResponse struct {
	Status        string                `json:"status"`
	Event         planning.EventSummary `json:"event"`
	CorrelationID string                `json:"correlation_id"`
}

type eventRequest struct {
	BaseVersion int64  `json:"base_version"`
	Name        string `json:"name"`
	StartsOn    string `json:"starts_on"`
	EndsOn      string `json:"ends_on"`
	Note        string `json:"note"`
}

type obligationsResponse struct {
	Status        string                       `json:"status"`
	Obligations   []planning.ObligationSummary `json:"obligations"`
	CorrelationID string                       `json:"correlation_id"`
}

type obligationResponse struct {
	Status        string                     `json:"status"`
	Obligation    planning.ObligationSummary `json:"obligation"`
	CorrelationID string                     `json:"correlation_id"`
}

type obligationRequest struct {
	BaseVersion  int64                        `json:"base_version"`
	Direction    planning.ObligationDirection `json:"direction"`
	PrincipalVND int64                        `json:"principal_vnd"`
	Counterparty string                       `json:"counterparty"`
	DueOn        string                       `json:"due_on"`
	Note         string                       `json:"note"`
}

type recurringSchedulesResponse struct {
	Status        string                       `json:"status"`
	Schedules     []planning.RecurringSchedule `json:"schedules"`
	CorrelationID string                       `json:"correlation_id"`
}

type recurringScheduleResponse struct {
	Status        string                     `json:"status"`
	Schedule      planning.RecurringSchedule `json:"schedule"`
	CorrelationID string                     `json:"correlation_id"`
}

type recurringScheduleRequest struct {
	BaseVersion         int64                         `json:"base_version"`
	Name                string                        `json:"name"`
	Frequency           planning.RecurrenceFrequency  `json:"frequency"`
	Timezone            string                        `json:"timezone"`
	StartsAt            string                        `json:"starts_at"`
	EndsAt              string                        `json:"ends_at"`
	PostingMode         planning.RecurringPostingMode `json:"posting_mode"`
	Type                finance.TransactionType       `json:"type"`
	SourceWalletID      string                        `json:"source_wallet_id"`
	DestinationWalletID string                        `json:"destination_wallet_id"`
	CategoryID          string                        `json:"category_id"`
	BudgetID            string                        `json:"budget_id"`
	AmountVND           int64                         `json:"amount_vnd"`
	Note                string                        `json:"note"`
}

type planningVersionRequest struct {
	BaseVersion int64 `json:"base_version"`
}

type transactionDraftsResponse struct {
	Status        string                      `json:"status"`
	Drafts        []planning.TransactionDraft `json:"drafts"`
	CorrelationID string                      `json:"correlation_id"`
}

type transactionDraftDecisionResponse struct {
	Status        string                    `json:"status"`
	Draft         planning.TransactionDraft `json:"draft"`
	Transaction   *finance.Transaction      `json:"transaction,omitempty"`
	CorrelationID string                    `json:"correlation_id"`
}

type confirmTransactionDraftRequest struct {
	Version   int64  `json:"version"`
	AmountVND int64  `json:"amount_vnd"`
	Note      string `json:"note"`
}

type rejectTransactionDraftRequest struct {
	Version int64 `json:"version"`
}

func budgets(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}

		switch r.Method {
		case http.MethodGet:
			listBudgets(w, r, repo, userID)
		case http.MethodPost:
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				createBudget(w, r, repo, userID)
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func budgetByID(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		budgetID, action, valid := parseBudgetPath(r.URL.Path)
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Budget not found", correlationID(r.Context())))
			return
		}
		switch {
		case r.Method == http.MethodPatch && action == "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				updateBudget(w, r, repo, userID, budgetID)
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				archiveBudget(w, r, repo, userID, budgetID)
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func events(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		switch r.Method {
		case http.MethodGet:
			items, err := repo.ListEvents(r.Context(), userID)
			if err != nil {
				writePlanningError(w, r, err, "Events unavailable")
				return
			}
			writeJSON(w, http.StatusOK, eventsResponse{Status: "ok", Events: items, CorrelationID: correlationID(r.Context())})
		case http.MethodPost:
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input, _, ok := decodeEventRequest(w, r)
				if !ok {
					return
				}
				event, err := repo.CreateEvent(r.Context(), userID, planning.CreateEventInput(input))
				if err != nil {
					writePlanningError(w, r, err, "Event unavailable")
					return
				}
				writeJSON(w, http.StatusCreated, eventResponse{Status: "ok", Event: event, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func eventByID(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		eventID, action, relatedID, valid := parseLinkedPath(r.URL.Path, "/api/v1/events/", "transactions")
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Event not found", correlationID(r.Context())))
			return
		}
		switch {
		case r.Method == http.MethodPatch && action == "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input, baseVersion, ok := decodeEventRequest(w, r)
				if !ok {
					return
				}
				if baseVersion <= 0 {
					writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "base_version is required", correlationID(r.Context())))
					return
				}
				event, err := repo.UpdateEvent(r.Context(), userID, eventID, planning.UpdateEventInput{BaseVersion: baseVersion, Name: input.Name, StartsOn: input.StartsOn, EndsOn: input.EndsOn, Note: input.Note})
				if err != nil {
					writePlanningError(w, r, err, "Event unavailable")
					return
				}
				writeJSON(w, http.StatusOK, eventResponse{Status: "ok", Event: event, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				baseVersion, ok := decodePlanningBaseVersion(w, r)
				if !ok {
					return
				}
				if err := repo.ArchiveEvent(r.Context(), userID, eventID, baseVersion); err != nil {
					writePlanningError(w, r, err, "Event unavailable")
					return
				}
				writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "transactions":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := repo.LinkEventTransaction(r.Context(), userID, eventID, relatedID); err != nil {
					writePlanningError(w, r, err, "Event unavailable")
					return
				}
				writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func obligations(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		switch r.Method {
		case http.MethodGet:
			items, err := repo.ListObligations(r.Context(), userID)
			if err != nil {
				writePlanningError(w, r, err, "Obligations unavailable")
				return
			}
			writeJSON(w, http.StatusOK, obligationsResponse{Status: "ok", Obligations: items, CorrelationID: correlationID(r.Context())})
		case http.MethodPost:
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input, _, ok := decodeObligationRequest(w, r)
				if !ok {
					return
				}
				obligation, err := repo.CreateObligation(r.Context(), userID, planning.CreateObligationInput(input))
				if err != nil {
					writePlanningError(w, r, err, "Obligation unavailable")
					return
				}
				writeJSON(w, http.StatusCreated, obligationResponse{Status: "ok", Obligation: obligation, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func obligationByID(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		obligationID, action, relatedID, valid := parseLinkedPath(r.URL.Path, "/api/v1/obligations/", "repayments")
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Obligation not found", correlationID(r.Context())))
			return
		}
		switch {
		case r.Method == http.MethodPatch && action == "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input, baseVersion, ok := decodeObligationRequest(w, r)
				if !ok {
					return
				}
				if baseVersion <= 0 {
					writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "base_version is required", correlationID(r.Context())))
					return
				}
				obligation, err := repo.UpdateObligation(r.Context(), userID, obligationID, planning.UpdateObligationInput{BaseVersion: baseVersion, Direction: input.Direction, PrincipalVND: input.PrincipalVND, Counterparty: input.Counterparty, DueOn: input.DueOn, Note: input.Note})
				if err != nil {
					writePlanningError(w, r, err, "Obligation unavailable")
					return
				}
				writeJSON(w, http.StatusOK, obligationResponse{Status: "ok", Obligation: obligation, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				baseVersion, ok := decodePlanningBaseVersion(w, r)
				if !ok {
					return
				}
				if err := repo.ArchiveObligation(r.Context(), userID, obligationID, baseVersion); err != nil {
					writePlanningError(w, r, err, "Obligation unavailable")
					return
				}
				writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "repayments":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := repo.LinkObligationRepayment(r.Context(), userID, obligationID, relatedID); err != nil {
					writePlanningError(w, r, err, "Obligation unavailable")
					return
				}
				writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func recurringSchedules(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		switch r.Method {
		case http.MethodGet:
			schedules, err := repo.ListRecurringSchedules(r.Context(), userID)
			if err != nil {
				writePlanningError(w, r, err, "Recurring schedules unavailable")
				return
			}
			writeJSON(w, http.StatusOK, recurringSchedulesResponse{Status: "ok", Schedules: schedules, CorrelationID: correlationID(r.Context())})
		case http.MethodPost:
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input, _, ok := decodeRecurringScheduleRequest(w, r)
				if !ok {
					return
				}
				schedule, err := repo.CreateRecurringSchedule(r.Context(), userID, input)
				if err != nil {
					writePlanningError(w, r, err, "Recurring schedule unavailable")
					return
				}
				writeJSON(w, http.StatusCreated, recurringScheduleResponse{Status: "ok", Schedule: schedule, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func recurringScheduleByID(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		scheduleID, action, valid := parseBudgetPathWithPrefix(r.URL.Path, "/api/v1/recurring-schedules/")
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Recurring schedule not found", correlationID(r.Context())))
			return
		}
		switch {
		case r.Method == http.MethodPatch && action == "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input, baseVersion, ok := decodeRecurringScheduleRequest(w, r)
				if !ok {
					return
				}
				if baseVersion <= 0 {
					writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "base_version is required", correlationID(r.Context())))
					return
				}
				schedule, err := repo.UpdateRecurringSchedule(r.Context(), userID, scheduleID, planning.UpdateRecurringScheduleInput{
					BaseVersion: baseVersion, Name: input.Name, Frequency: input.Frequency, Timezone: input.Timezone,
					StartsAt: input.StartsAt, EndsAt: input.EndsAt, PostingMode: input.PostingMode, Type: input.Type,
					SourceWalletID: input.SourceWalletID, DestinationWalletID: input.DestinationWalletID, CategoryID: input.CategoryID,
					BudgetID: input.BudgetID, AmountVND: input.AmountVND, Note: input.Note,
				})
				if err != nil {
					writePlanningError(w, r, err, "Recurring schedule unavailable")
					return
				}
				writeJSON(w, http.StatusOK, recurringScheduleResponse{Status: "ok", Schedule: schedule, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
			return
		case r.Method == http.MethodPost && action == "pause":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				baseVersion, ok := decodePlanningBaseVersion(w, r)
				if !ok {
					return
				}
				schedule, err := repo.PauseRecurringSchedule(r.Context(), userID, scheduleID, baseVersion)
				if err != nil {
					writePlanningError(w, r, err, "Recurring schedule unavailable")
					return
				}
				writeJSON(w, http.StatusOK, recurringScheduleResponse{Status: "ok", Schedule: schedule, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
			return
		case r.Method == http.MethodPost && action == "resume":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				baseVersion, ok := decodePlanningBaseVersion(w, r)
				if !ok {
					return
				}
				schedule, err := repo.ResumeRecurringSchedule(r.Context(), userID, scheduleID, baseVersion, time.Now())
				if err != nil {
					writePlanningError(w, r, err, "Recurring schedule unavailable")
					return
				}
				writeJSON(w, http.StatusOK, recurringScheduleResponse{Status: "ok", Schedule: schedule, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
			return
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				baseVersion, ok := decodePlanningBaseVersion(w, r)
				if !ok {
					return
				}
				if err := repo.ArchiveRecurringSchedule(r.Context(), userID, scheduleID, baseVersion); err != nil {
					writePlanningError(w, r, err, "Recurring schedule unavailable")
					return
				}
				writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
	}
}

func transactionDrafts(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		drafts, err := repo.ListTransactionDrafts(r.Context(), userID)
		if err != nil {
			writePlanningError(w, r, err, "Drafts unavailable")
			return
		}
		writeJSON(w, http.StatusOK, transactionDraftsResponse{Status: "ok", Drafts: drafts, CorrelationID: correlationID(r.Context())})
	}
}

func transactionDraftByID(cfg config.Config, repo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Planning repository unavailable", correlationID(r.Context())))
			return
		}
		draftID, action, valid := parseTransactionDraftPath(r.URL.Path)
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Draft not found", correlationID(r.Context())))
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch action {
			case "confirm":
				confirmTransactionDraft(w, r, repo, userID, draftID)
			case "reject":
				rejectTransactionDraft(w, r, repo, userID, draftID)
			}
		})).ServeHTTP(w, r)
	}
}

func confirmTransactionDraft(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string, draftID string) {
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Idempotency-Key is required", correlationID(r.Context())))
		return
	}
	var req confirmTransactionDraftRequest
	if err := decodeDraftDecisionRequest(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return
	}
	if req.Version <= 0 || req.AmountVND <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Version and positive amount are required", correlationID(r.Context())))
		return
	}
	decision, err := repo.ConfirmTransactionDraft(r.Context(), userID, draftID, planning.ConfirmTransactionDraftInput{
		Version:        req.Version,
		AmountVND:      req.AmountVND,
		Note:           req.Note,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		writePlanningError(w, r, err, "Draft confirmation unavailable")
		return
	}
	writeJSON(w, http.StatusOK, transactionDraftDecisionResponse{Status: "ok", Draft: decision.Draft, Transaction: decision.Transaction, CorrelationID: correlationID(r.Context())})
}

func rejectTransactionDraft(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string, draftID string) {
	var req rejectTransactionDraftRequest
	if err := decodeDraftDecisionRequest(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return
	}
	if req.Version <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Version is required", correlationID(r.Context())))
		return
	}
	decision, err := repo.RejectTransactionDraft(r.Context(), userID, draftID, planning.RejectTransactionDraftInput{Version: req.Version})
	if err != nil {
		writePlanningError(w, r, err, "Draft rejection unavailable")
		return
	}
	writeJSON(w, http.StatusOK, transactionDraftDecisionResponse{Status: "ok", Draft: decision.Draft, CorrelationID: correlationID(r.Context())})
}

func decodeDraftDecisionRequest(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func listBudgets(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string) {
	progress, err := repo.ListBudgetProgress(r.Context(), userID, time.Now())
	if err != nil {
		writePlanningError(w, r, err, "Budgets unavailable")
		return
	}
	writeJSON(w, http.StatusOK, budgetsResponse{Status: "ok", Budgets: progress, CorrelationID: correlationID(r.Context())})
}

func createBudget(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string) {
	input, _, ok := decodeBudgetRequest(w, r)
	if !ok {
		return
	}
	budget, err := repo.CreateBudget(r.Context(), userID, planning.CreateBudgetInput(input))
	if err != nil {
		writePlanningError(w, r, err, "Budget unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, budgetResponse{Status: "ok", Budget: budget, CorrelationID: correlationID(r.Context())})
}

func updateBudget(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string, budgetID string) {
	input, baseVersion, ok := decodeBudgetRequest(w, r)
	if !ok {
		return
	}
	if baseVersion <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "base_version is required", correlationID(r.Context())))
		return
	}
	budget, err := repo.UpdateBudget(r.Context(), userID, budgetID, planning.UpdateBudgetInput{BaseVersion: baseVersion, Name: input.Name, PeriodType: input.PeriodType, AmountVND: input.AmountVND, CategoryIDs: input.CategoryIDs, CustomStart: input.CustomStart, CustomEnd: input.CustomEnd})
	if err != nil {
		writePlanningError(w, r, err, "Budget unavailable")
		return
	}
	writeJSON(w, http.StatusOK, budgetResponse{Status: "ok", Budget: budget, CorrelationID: correlationID(r.Context())})
}

func archiveBudget(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string, budgetID string) {
	baseVersion, ok := decodePlanningBaseVersion(w, r)
	if !ok {
		return
	}
	if err := repo.ArchiveBudget(r.Context(), userID, budgetID, baseVersion); err != nil {
		writePlanningError(w, r, err, "Budget unavailable")
		return
	}
	writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
}

func decodeBudgetRequest(w http.ResponseWriter, r *http.Request) (planning.CreateBudgetInput, int64, bool) {
	var req budgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return planning.CreateBudgetInput{}, 0, false
	}
	customStart, ok := parseOptionalDate(w, r, req.CustomStart)
	if !ok {
		return planning.CreateBudgetInput{}, 0, false
	}
	customEnd, ok := parseOptionalDate(w, r, req.CustomEnd)
	if !ok {
		return planning.CreateBudgetInput{}, 0, false
	}
	return planning.CreateBudgetInput{
		Name:        req.Name,
		PeriodType:  req.PeriodType,
		AmountVND:   req.AmountVND,
		CategoryIDs: req.CategoryIDs,
		CustomStart: customStart,
		CustomEnd:   customEnd,
	}, req.BaseVersion, true
}

func decodePlanningBaseVersion(w http.ResponseWriter, r *http.Request) (int64, bool) {
	var req planningVersionRequest
	if err := decodeDraftDecisionRequest(r, &req); err != nil || req.BaseVersion <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "base_version is required", correlationID(r.Context())))
		return 0, false
	}
	return req.BaseVersion, true
}

func decodeEventRequest(w http.ResponseWriter, r *http.Request) (planning.CreateEventInput, int64, bool) {
	var req eventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return planning.CreateEventInput{}, 0, false
	}
	return planning.CreateEventInput{Name: req.Name, StartsOn: req.StartsOn, EndsOn: req.EndsOn, Note: req.Note}, req.BaseVersion, true
}

func decodeObligationRequest(w http.ResponseWriter, r *http.Request) (planning.CreateObligationInput, int64, bool) {
	var req obligationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return planning.CreateObligationInput{}, 0, false
	}
	return planning.CreateObligationInput{Direction: req.Direction, PrincipalVND: req.PrincipalVND, Counterparty: req.Counterparty, DueOn: req.DueOn, Note: req.Note}, req.BaseVersion, true
}

func decodeRecurringScheduleRequest(w http.ResponseWriter, r *http.Request) (planning.CreateRecurringScheduleInput, int64, bool) {
	var req recurringScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return planning.CreateRecurringScheduleInput{}, 0, false
	}
	return planning.CreateRecurringScheduleInput{
		Name:                req.Name,
		Frequency:           req.Frequency,
		Timezone:            req.Timezone,
		StartsAt:            req.StartsAt,
		EndsAt:              req.EndsAt,
		PostingMode:         req.PostingMode,
		Type:                req.Type,
		SourceWalletID:      req.SourceWalletID,
		DestinationWalletID: req.DestinationWalletID,
		CategoryID:          req.CategoryID,
		BudgetID:            req.BudgetID,
		AmountVND:           req.AmountVND,
		Note:                req.Note,
	}, req.BaseVersion, true
}

func parseOptionalDate(w http.ResponseWriter, r *http.Request, value string) (*time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, true
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid date", correlationID(r.Context())))
		return nil, false
	}
	return &parsed, true
}

func parseLinkedPath(path string, prefix string, linkedAction string) (resourceID string, action string, relatedID string, valid bool) {
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", "", "", false
	}
	if len(parts) == 1 {
		return parts[0], "", "", true
	}
	if len(parts) == 2 && parts[1] == "archive" {
		return parts[0], "archive", "", true
	}
	if len(parts) == 3 && parts[1] == linkedAction && parts[2] != "" {
		return parts[0], linkedAction, parts[2], true
	}
	return "", "", "", false
}

func parseBudgetPath(path string) (budgetID string, action string, valid bool) {
	return parseBudgetPathWithPrefix(path, "/api/v1/budgets/")
}

func parseBudgetPathWithPrefix(path string, prefix string) (budgetID string, action string, valid bool) {
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", "", false
	}
	if len(parts) == 1 {
		return parts[0], "", true
	}
	if len(parts) == 2 && (parts[1] == "archive" || parts[1] == "pause" || parts[1] == "resume") {
		return parts[0], parts[1], true
	}
	return "", "", false
}

func parseTransactionDraftPath(path string) (draftID string, action string, valid bool) {
	rest := strings.TrimPrefix(path, "/api/v1/transaction-drafts/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || (parts[1] != "confirm" && parts[1] != "reject") {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func writePlanningError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, planning.ErrValidation):
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
	case errors.Is(err, planning.ErrForbidden):
		writeJSON(w, http.StatusForbidden, ErrorEnvelope("FORBIDDEN", "Planning object is not available", correlationID(r.Context())))
	case errors.Is(err, planning.ErrDraftAlreadyResolved):
		writeJSON(w, http.StatusConflict, ErrorEnvelope("DRAFT_ALREADY_RESOLVED", "Draft is already resolved", correlationID(r.Context())))
	case errors.Is(err, planning.ErrDraftVersionConflict):
		writeJSON(w, http.StatusConflict, ErrorEnvelope("DRAFT_VERSION_CONFLICT", "Draft version has changed", correlationID(r.Context())))
	case errors.Is(err, planning.ErrVersionConflict):
		writeJSON(w, http.StatusConflict, ErrorEnvelope("VERSION_CONFLICT", "Planning object version has changed", correlationID(r.Context())))
	default:
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", fallback, correlationID(r.Context())))
	}
}
