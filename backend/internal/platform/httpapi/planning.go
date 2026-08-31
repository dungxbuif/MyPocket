package httpapi

import (
	"encoding/json"
	"errors"
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
	Name     string `json:"name"`
	StartsOn string `json:"starts_on"`
	EndsOn   string `json:"ends_on"`
	Note     string `json:"note"`
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
	Name                string                       `json:"name"`
	Frequency           planning.RecurrenceFrequency `json:"frequency"`
	Timezone            string                       `json:"timezone"`
	StartsAt            string                       `json:"starts_at"`
	Type                finance.TransactionType      `json:"type"`
	SourceWalletID      string                       `json:"source_wallet_id"`
	DestinationWalletID string                       `json:"destination_wallet_id"`
	CategoryID          string                       `json:"category_id"`
	AmountVND           int64                        `json:"amount_vnd"`
	Note                string                       `json:"note"`
}

type transactionDraftsResponse struct {
	Status        string                      `json:"status"`
	Drafts        []planning.TransactionDraft `json:"drafts"`
	CorrelationID string                      `json:"correlation_id"`
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
				input, ok := decodeEventRequest(w, r)
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
				input, ok := decodeEventRequest(w, r)
				if !ok {
					return
				}
				event, err := repo.UpdateEvent(r.Context(), userID, eventID, planning.UpdateEventInput(input))
				if err != nil {
					writePlanningError(w, r, err, "Event unavailable")
					return
				}
				writeJSON(w, http.StatusOK, eventResponse{Status: "ok", Event: event, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := repo.ArchiveEvent(r.Context(), userID, eventID); err != nil {
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
				input, ok := decodeObligationRequest(w, r)
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
				input, ok := decodeObligationRequest(w, r)
				if !ok {
					return
				}
				obligation, err := repo.UpdateObligation(r.Context(), userID, obligationID, planning.UpdateObligationInput(input))
				if err != nil {
					writePlanningError(w, r, err, "Obligation unavailable")
					return
				}
				writeJSON(w, http.StatusOK, obligationResponse{Status: "ok", Obligation: obligation, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := repo.ArchiveObligation(r.Context(), userID, obligationID); err != nil {
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
				input, ok := decodeRecurringScheduleRequest(w, r)
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
		if r.Method == http.MethodPost && action == "archive" {
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := repo.ArchiveRecurringSchedule(r.Context(), userID, scheduleID); err != nil {
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

func listBudgets(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string) {
	progress, err := repo.ListBudgetProgress(r.Context(), userID, time.Now())
	if err != nil {
		writePlanningError(w, r, err, "Budgets unavailable")
		return
	}
	writeJSON(w, http.StatusOK, budgetsResponse{Status: "ok", Budgets: progress, CorrelationID: correlationID(r.Context())})
}

func createBudget(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string) {
	input, ok := decodeBudgetRequest(w, r)
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
	input, ok := decodeBudgetRequest(w, r)
	if !ok {
		return
	}
	budget, err := repo.UpdateBudget(r.Context(), userID, budgetID, planning.UpdateBudgetInput(input))
	if err != nil {
		writePlanningError(w, r, err, "Budget unavailable")
		return
	}
	writeJSON(w, http.StatusOK, budgetResponse{Status: "ok", Budget: budget, CorrelationID: correlationID(r.Context())})
}

func archiveBudget(w http.ResponseWriter, r *http.Request, repo PlanningRepository, userID string, budgetID string) {
	if err := repo.ArchiveBudget(r.Context(), userID, budgetID); err != nil {
		writePlanningError(w, r, err, "Budget unavailable")
		return
	}
	writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
}

func decodeBudgetRequest(w http.ResponseWriter, r *http.Request) (planning.CreateBudgetInput, bool) {
	var req budgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return planning.CreateBudgetInput{}, false
	}
	customStart, ok := parseOptionalDate(w, r, req.CustomStart)
	if !ok {
		return planning.CreateBudgetInput{}, false
	}
	customEnd, ok := parseOptionalDate(w, r, req.CustomEnd)
	if !ok {
		return planning.CreateBudgetInput{}, false
	}
	return planning.CreateBudgetInput{
		Name:        req.Name,
		PeriodType:  req.PeriodType,
		AmountVND:   req.AmountVND,
		CategoryIDs: req.CategoryIDs,
		CustomStart: customStart,
		CustomEnd:   customEnd,
	}, true
}

func decodeEventRequest(w http.ResponseWriter, r *http.Request) (planning.CreateEventInput, bool) {
	var req eventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return planning.CreateEventInput{}, false
	}
	return planning.CreateEventInput{Name: req.Name, StartsOn: req.StartsOn, EndsOn: req.EndsOn, Note: req.Note}, true
}

func decodeObligationRequest(w http.ResponseWriter, r *http.Request) (planning.CreateObligationInput, bool) {
	var req obligationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return planning.CreateObligationInput{}, false
	}
	return planning.CreateObligationInput{Direction: req.Direction, PrincipalVND: req.PrincipalVND, Counterparty: req.Counterparty, DueOn: req.DueOn, Note: req.Note}, true
}

func decodeRecurringScheduleRequest(w http.ResponseWriter, r *http.Request) (planning.CreateRecurringScheduleInput, bool) {
	var req recurringScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return planning.CreateRecurringScheduleInput{}, false
	}
	return planning.CreateRecurringScheduleInput{
		Name:                req.Name,
		Frequency:           req.Frequency,
		Timezone:            req.Timezone,
		StartsAt:            req.StartsAt,
		Type:                req.Type,
		SourceWalletID:      req.SourceWalletID,
		DestinationWalletID: req.DestinationWalletID,
		CategoryID:          req.CategoryID,
		AmountVND:           req.AmountVND,
		Note:                req.Note,
	}, true
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
	if len(parts) == 2 && parts[1] == "archive" {
		return parts[0], "archive", true
	}
	return "", "", false
}

func writePlanningError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, planning.ErrValidation):
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
	case errors.Is(err, planning.ErrForbidden):
		writeJSON(w, http.StatusForbidden, ErrorEnvelope("FORBIDDEN", "Budget is not available", correlationID(r.Context())))
	default:
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", fallback, correlationID(r.Context())))
	}
}
