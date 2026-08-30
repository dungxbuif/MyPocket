package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

type walletsResponse struct {
	Status        string       `json:"status"`
	Wallets       []walletBody `json:"wallets"`
	CorrelationID string       `json:"correlation_id"`
}

type walletResponse struct {
	Status        string     `json:"status"`
	Wallet        walletBody `json:"wallet"`
	CorrelationID string     `json:"correlation_id"`
}

type walletBody struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Type           finance.WalletType `json:"type"`
	BalanceVND     int64              `json:"balance_vnd"`
	IncludeInTotal bool               `json:"include_in_total"`
	IsDefaultAI    bool               `json:"is_default_ai"`
	Version        int64              `json:"version"`
}

type createWalletRequest struct {
	Name           string             `json:"name"`
	Type           finance.WalletType `json:"type"`
	CreditLimitVND *int64             `json:"credit_limit_vnd"`
	StatementDay   *int               `json:"statement_day"`
	PaymentDueDay  *int               `json:"payment_due_day"`
}

type updateWalletRequest struct {
	Name           string `json:"name"`
	IncludeInTotal *bool  `json:"include_in_total"`
}

type commandResponse struct {
	Status        string `json:"status"`
	CorrelationID string `json:"correlation_id"`
}

type categoriesResponse struct {
	Status        string         `json:"status"`
	Categories    []categoryBody `json:"categories"`
	CorrelationID string         `json:"correlation_id"`
}

type categoryResponse struct {
	Status        string       `json:"status"`
	Category      categoryBody `json:"category"`
	CorrelationID string       `json:"correlation_id"`
}

type categoryBody struct {
	ID        string               `json:"id"`
	ParentID  string               `json:"parent_id,omitempty"`
	Kind      finance.CategoryKind `json:"kind"`
	Name      string               `json:"name"`
	SystemKey string               `json:"system_key,omitempty"`
	IsSystem  bool                 `json:"is_system"`
}

type createCategoryRequest struct {
	Kind finance.CategoryKind `json:"kind"`
	Name string               `json:"name"`
}

type updateCategoryRequest struct {
	Name string `json:"name"`
}

type setWalletCategoryRequest struct {
	Active bool `json:"active"`
}

type transactionsResponse struct {
	Status        string            `json:"status"`
	Transactions  []transactionBody `json:"transactions"`
	CorrelationID string            `json:"correlation_id"`
}

type transactionResponse struct {
	Status        string          `json:"status"`
	Transaction   transactionBody `json:"transaction"`
	CorrelationID string          `json:"correlation_id"`
}

type transactionBody struct {
	ID                  string                  `json:"id"`
	Type                finance.TransactionType `json:"type"`
	SourceWalletID      string                  `json:"source_wallet_id"`
	DestinationWalletID string                  `json:"destination_wallet_id,omitempty"`
	CategoryID          string                  `json:"category_id,omitempty"`
	AmountVND           int64                   `json:"amount_vnd"`
	BalanceAfterVND     int64                   `json:"balance_after_vnd"`
	OccurredAt          time.Time               `json:"occurred_at"`
	Note                string                  `json:"note"`
	WithPerson          string                  `json:"with_person"`
	EventRef            string                  `json:"event_ref"`
	ExcludedFromReports bool                    `json:"excluded_from_reports"`
	Version             int64                   `json:"version"`
}

type transactionRequest struct {
	Type                finance.TransactionType `json:"type"`
	SourceWalletID      string                  `json:"source_wallet_id"`
	DestinationWalletID string                  `json:"destination_wallet_id"`
	CategoryID          string                  `json:"category_id"`
	AmountVND           int64                   `json:"amount_vnd"`
	TargetBalanceVND    *int64                  `json:"target_balance_vnd"`
	OccurredAt          time.Time               `json:"occurred_at"`
	Note                string                  `json:"note"`
	WithPerson          string                  `json:"with_person"`
	EventRef            string                  `json:"event_ref"`
	ExcludedFromReports bool                    `json:"excluded_from_reports"`
}

func wallets(cfg config.Config, repo FinanceRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Finance repository unavailable", correlationID(r.Context())))
			return
		}

		switch r.Method {
		case http.MethodGet:
			listWallets(w, r, repo, userID)
		case http.MethodPost:
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				createWallet(w, r, repo, userID)
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func walletByID(cfg config.Config, repo FinanceRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Finance repository unavailable", correlationID(r.Context())))
			return
		}

		walletID, categoryID, action, valid := parseWalletPath(r.URL.Path)
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Wallet not found", correlationID(r.Context())))
			return
		}

		switch {
		case r.Method == http.MethodPatch && action == "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				updateWallet(w, r, repo, userID, walletID)
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				archiveWallet(w, r, repo, userID, walletID)
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "default-ai":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				setDefaultAIWallet(w, r, repo, userID, walletID)
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPut && action == "category":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				setWalletCategoryActive(w, r, repo, userID, walletID, categoryID)
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func listWallets(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string) {
	wallets, err := repo.ListWallets(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Wallets unavailable", correlationID(r.Context())))
		return
	}
	body := make([]walletBody, 0, len(wallets))
	for _, wallet := range wallets {
		body = append(body, toWalletBody(wallet))
	}
	writeJSON(w, http.StatusOK, walletsResponse{
		Status:        "ok",
		Wallets:       body,
		CorrelationID: correlationID(r.Context()),
	})
}

func createWallet(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string) {
	var req createWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return
	}
	wallet, err := repo.CreateWallet(r.Context(), userID, finance.CreateWalletInput{
		Name:           req.Name,
		Type:           req.Type,
		CreditLimitVND: req.CreditLimitVND,
		StatementDay:   req.StatementDay,
		PaymentDueDay:  req.PaymentDueDay,
	})
	if errors.Is(err, finance.ErrValidation) {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
		return
	}
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Wallet unavailable", correlationID(r.Context())))
		return
	}
	writeJSON(w, http.StatusCreated, walletResponse{
		Status:        "ok",
		Wallet:        toWalletBody(wallet),
		CorrelationID: correlationID(r.Context()),
	})
}

func updateWallet(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string, walletID string) {
	var req updateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return
	}
	wallet, err := repo.UpdateWallet(r.Context(), userID, walletID, finance.UpdateWalletInput{
		Name:           req.Name,
		IncludeInTotal: req.IncludeInTotal,
	})
	if err != nil {
		writeFinanceError(w, r, err, "Wallet unavailable")
		return
	}
	writeJSON(w, http.StatusOK, walletResponse{
		Status:        "ok",
		Wallet:        toWalletBody(wallet),
		CorrelationID: correlationID(r.Context()),
	})
}

func archiveWallet(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string, walletID string) {
	if err := repo.ArchiveWallet(r.Context(), userID, walletID); err != nil {
		writeFinanceError(w, r, err, "Wallet unavailable")
		return
	}
	writeJSON(w, http.StatusOK, commandResponse{
		Status:        "ok",
		CorrelationID: correlationID(r.Context()),
	})
}

func setDefaultAIWallet(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string, walletID string) {
	if err := repo.SetDefaultAIWallet(r.Context(), userID, walletID); err != nil {
		writeFinanceError(w, r, err, "Wallet unavailable")
		return
	}
	writeJSON(w, http.StatusOK, commandResponse{
		Status:        "ok",
		CorrelationID: correlationID(r.Context()),
	})
}

func setWalletCategoryActive(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string, walletID string, categoryID string) {
	var req setWalletCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return
	}
	if err := repo.SetWalletCategoryActive(r.Context(), userID, walletID, categoryID, req.Active); err != nil {
		writeFinanceError(w, r, err, "Wallet category unavailable")
		return
	}
	writeJSON(w, http.StatusOK, commandResponse{
		Status:        "ok",
		CorrelationID: correlationID(r.Context()),
	})
}

func categories(cfg config.Config, repo FinanceRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Finance repository unavailable", correlationID(r.Context())))
			return
		}
		switch r.Method {
		case http.MethodGet:
			listCategories(w, r, repo, userID)
		case http.MethodPost:
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				createCategory(w, r, repo, userID)
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func categoryByID(cfg config.Config, repo FinanceRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Finance repository unavailable", correlationID(r.Context())))
			return
		}

		categoryID, action, valid := parseCategoryPath(r.URL.Path)
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Category not found", correlationID(r.Context())))
			return
		}

		switch {
		case r.Method == http.MethodPatch && action == "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				updateCategory(w, r, repo, userID, categoryID)
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				archiveCategory(w, r, repo, userID, categoryID)
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func listCategories(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string) {
	categories, err := repo.ListCategories(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Categories unavailable", correlationID(r.Context())))
		return
	}
	body := make([]categoryBody, 0, len(categories))
	for _, category := range categories {
		body = append(body, toCategoryBody(category))
	}
	writeJSON(w, http.StatusOK, categoriesResponse{
		Status:        "ok",
		Categories:    body,
		CorrelationID: correlationID(r.Context()),
	})
}

func createCategory(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string) {
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return
	}
	category, err := repo.CreateCategory(r.Context(), userID, finance.CreateCategoryInput{
		Kind: req.Kind,
		Name: req.Name,
	})
	if err != nil {
		writeFinanceError(w, r, err, "Category unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, categoryResponse{
		Status:        "ok",
		Category:      toCategoryBody(category),
		CorrelationID: correlationID(r.Context()),
	})
}

func updateCategory(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string, categoryID string) {
	var req updateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return
	}
	category, err := repo.UpdateCategory(r.Context(), userID, categoryID, finance.UpdateCategoryInput{Name: req.Name})
	if err != nil {
		writeFinanceError(w, r, err, "Category unavailable")
		return
	}
	writeJSON(w, http.StatusOK, categoryResponse{
		Status:        "ok",
		Category:      toCategoryBody(category),
		CorrelationID: correlationID(r.Context()),
	})
}

func archiveCategory(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string, categoryID string) {
	if err := repo.ArchiveCategory(r.Context(), userID, categoryID); err != nil {
		writeFinanceError(w, r, err, "Category unavailable")
		return
	}
	writeJSON(w, http.StatusOK, commandResponse{
		Status:        "ok",
		CorrelationID: correlationID(r.Context()),
	})
}

func transactions(cfg config.Config, repo FinanceRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Finance repository unavailable", correlationID(r.Context())))
			return
		}

		switch r.Method {
		case http.MethodGet:
			listTransactions(w, r, repo, userID)
		case http.MethodPost:
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				createTransaction(w, r, repo, userID)
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func transactionByID(cfg config.Config, repo FinanceRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Finance repository unavailable", correlationID(r.Context())))
			return
		}

		transactionID, action, valid := parseTransactionPath(r.URL.Path)
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Transaction not found", correlationID(r.Context())))
			return
		}

		switch {
		case r.Method == http.MethodPatch && action == "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				updateTransaction(w, r, repo, userID, transactionID)
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				archiveTransaction(w, r, repo, userID, transactionID)
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func listTransactions(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string) {
	filters, err := transactionFiltersFromQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
		return
	}
	transactions, err := repo.ListTransactions(r.Context(), userID, filters)
	if err != nil {
		writeFinanceError(w, r, err, "Transactions unavailable")
		return
	}
	body := make([]transactionBody, 0, len(transactions))
	for _, transaction := range transactions {
		body = append(body, toTransactionBody(transaction))
	}
	writeJSON(w, http.StatusOK, transactionsResponse{
		Status:        "ok",
		Transactions:  body,
		CorrelationID: correlationID(r.Context()),
	})
}

func createTransaction(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string) {
	req, ok := decodeTransactionRequest(w, r)
	if !ok {
		return
	}
	transaction, err := repo.CreateTransaction(r.Context(), userID, finance.CreateTransactionInput{
		IdempotencyKey:      strings.TrimSpace(r.Header.Get("Idempotency-Key")),
		Type:                req.Type,
		SourceWalletID:      req.SourceWalletID,
		DestinationWalletID: req.DestinationWalletID,
		CategoryID:          req.CategoryID,
		AmountVND:           req.AmountVND,
		TargetBalanceVND:    req.TargetBalanceVND,
		OccurredAt:          req.OccurredAt,
		Note:                req.Note,
		WithPerson:          req.WithPerson,
		EventRef:            req.EventRef,
		ExcludedFromReports: req.ExcludedFromReports,
	})
	if err != nil {
		writeFinanceError(w, r, err, "Transaction unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, transactionResponse{
		Status:        "ok",
		Transaction:   toTransactionBody(transaction),
		CorrelationID: correlationID(r.Context()),
	})
}

func updateTransaction(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string, transactionID string) {
	req, ok := decodeTransactionRequest(w, r)
	if !ok {
		return
	}
	transaction, err := repo.UpdateTransaction(r.Context(), userID, transactionID, finance.UpdateTransactionInput{
		Type:                req.Type,
		SourceWalletID:      req.SourceWalletID,
		DestinationWalletID: req.DestinationWalletID,
		CategoryID:          req.CategoryID,
		AmountVND:           req.AmountVND,
		TargetBalanceVND:    req.TargetBalanceVND,
		OccurredAt:          req.OccurredAt,
		Note:                req.Note,
		WithPerson:          req.WithPerson,
		EventRef:            req.EventRef,
		ExcludedFromReports: req.ExcludedFromReports,
	})
	if err != nil {
		writeFinanceError(w, r, err, "Transaction unavailable")
		return
	}
	writeJSON(w, http.StatusOK, transactionResponse{
		Status:        "ok",
		Transaction:   toTransactionBody(transaction),
		CorrelationID: correlationID(r.Context()),
	})
}

func archiveTransaction(w http.ResponseWriter, r *http.Request, repo FinanceRepository, userID string, transactionID string) {
	if err := repo.ArchiveTransaction(r.Context(), userID, transactionID); err != nil {
		writeFinanceError(w, r, err, "Transaction unavailable")
		return
	}
	writeJSON(w, http.StatusOK, commandResponse{
		Status:        "ok",
		CorrelationID: correlationID(r.Context()),
	})
}

func authenticatedUserID(w http.ResponseWriter, r *http.Request, cfg config.Config) (string, bool) {
	cookie, err := r.Cookie(identity.AuthCookieName)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Authentication required", correlationID(r.Context())))
		return "", false
	}
	claims, err := identity.NewCookieSigner([]byte(cfg.CookieSecret)).Verify(cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Authentication required", correlationID(r.Context())))
		return "", false
	}
	return claims.UserID, true
}

func toWalletBody(wallet finance.Wallet) walletBody {
	return walletBody{
		ID:             wallet.ID,
		Name:           wallet.Name,
		Type:           wallet.Type,
		BalanceVND:     wallet.BalanceVND,
		IncludeInTotal: wallet.IncludeInTotal,
		IsDefaultAI:    wallet.IsDefaultAI,
		Version:        wallet.Version,
	}
}

func toCategoryBody(category finance.Category) categoryBody {
	return categoryBody{
		ID:        category.ID,
		ParentID:  category.ParentID,
		Kind:      category.Kind,
		Name:      category.Name,
		SystemKey: category.SystemKey,
		IsSystem:  category.IsSystem,
	}
}

func decodeTransactionRequest(w http.ResponseWriter, r *http.Request) (transactionRequest, bool) {
	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return transactionRequest{}, false
	}
	return req, true
}

func transactionFiltersFromQuery(r *http.Request) (finance.TransactionFilters, error) {
	query := r.URL.Query()
	filters := finance.TransactionFilters{
		WalletID:   query.Get("wallet_id"),
		CategoryID: query.Get("category_id"),
		Type:       finance.TransactionType(query.Get("type")),
		Query:      query.Get("q"),
	}
	if value := query.Get("date_from"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return finance.TransactionFilters{}, errors.New("Invalid date_from")
		}
		filters.DateFrom = &parsed
	}
	if value := query.Get("date_to"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return finance.TransactionFilters{}, errors.New("Invalid date_to")
		}
		filters.DateTo = &parsed
	}
	if value := query.Get("excluded_from_reports"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return finance.TransactionFilters{}, errors.New("Invalid excluded_from_reports")
		}
		filters.ExcludedFromReports = &parsed
	}
	if value := query.Get("include_archived"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return finance.TransactionFilters{}, errors.New("Invalid include_archived")
		}
		filters.IncludeArchivedItems = parsed
	}
	return filters, nil
}

func parseTransactionPath(path string) (string, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/transactions/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", true
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "archive" {
		return parts[0], "archive", true
	}
	return "", "", false
}

func parseWalletPath(path string) (string, string, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/wallets/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", "", true
	}
	if len(parts) == 2 && parts[0] != "" && (parts[1] == "archive" || parts[1] == "default-ai") {
		return parts[0], "", parts[1], true
	}
	if len(parts) == 3 && parts[0] != "" && parts[1] == "categories" && parts[2] != "" {
		return parts[0], parts[2], "category", true
	}
	return "", "", "", false
}

func parseCategoryPath(path string) (string, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/categories/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", true
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "archive" {
		return parts[0], "archive", true
	}
	return "", "", false
}

func toTransactionBody(transaction finance.Transaction) transactionBody {
	return transactionBody{
		ID:                  transaction.ID,
		Type:                transaction.Type,
		SourceWalletID:      transaction.SourceWalletID,
		DestinationWalletID: transaction.DestinationWalletID,
		CategoryID:          transaction.CategoryID,
		AmountVND:           transaction.AmountVND,
		BalanceAfterVND:     transaction.BalanceAfterVND,
		OccurredAt:          transaction.OccurredAt,
		Note:                transaction.Note,
		WithPerson:          transaction.WithPerson,
		EventRef:            transaction.EventRef,
		ExcludedFromReports: transaction.ExcludedFromReports,
		Version:             transaction.Version,
	}
}

func writeFinanceError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, finance.ErrValidation):
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
	case errors.Is(err, finance.ErrForbidden):
		writeJSON(w, http.StatusForbidden, ErrorEnvelope("FORBIDDEN", "Forbidden", correlationID(r.Context())))
	case errors.Is(err, finance.ErrSystemCategoryLocked):
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
	default:
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", fallback, correlationID(r.Context())))
	}
}
