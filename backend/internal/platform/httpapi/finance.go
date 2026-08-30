package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

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

type categoriesResponse struct {
	Status        string         `json:"status"`
	Categories    []categoryBody `json:"categories"`
	CorrelationID string         `json:"correlation_id"`
}

type categoryBody struct {
	ID        string               `json:"id"`
	ParentID  string               `json:"parent_id,omitempty"`
	Kind      finance.CategoryKind `json:"kind"`
	Name      string               `json:"name"`
	SystemKey string               `json:"system_key,omitempty"`
	IsSystem  bool                 `json:"is_system"`
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
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}

		categories, err := repo.ListCategories(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Categories unavailable", correlationID(r.Context())))
			return
		}
		body := make([]categoryBody, 0, len(categories))
		for _, category := range categories {
			body = append(body, categoryBody{
				ID:        category.ID,
				ParentID:  category.ParentID,
				Kind:      category.Kind,
				Name:      category.Name,
				SystemKey: category.SystemKey,
				IsSystem:  category.IsSystem,
			})
		}
		writeJSON(w, http.StatusOK, categoriesResponse{
			Status:        "ok",
			Categories:    body,
			CorrelationID: correlationID(r.Context()),
		})
	}
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
