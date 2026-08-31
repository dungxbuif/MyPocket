package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mypocket/internal/platform/config"
	"mypocket/internal/portfolio"
)

type assetsResponse struct {
	Status        string               `json:"status"`
	Assets        []portfolio.Position `json:"assets"`
	CorrelationID string               `json:"correlation_id"`
}

type assetResponse struct {
	Status        string             `json:"status"`
	Asset         portfolio.Position `json:"asset"`
	CorrelationID string             `json:"correlation_id"`
}

type portfolioSummaryResponse struct {
	Status        string                     `json:"status"`
	Summary       portfolio.PortfolioSummary `json:"summary"`
	CorrelationID string                     `json:"correlation_id"`
}

type createAssetRequest struct {
	ID                string                `json:"id"`
	Type              portfolio.AssetType   `json:"type"`
	Symbol            string                `json:"symbol"`
	Exchange          string                `json:"exchange"`
	Name              string                `json:"name"`
	Unit              string                `json:"unit"`
	PricingMode       portfolio.PricingMode `json:"pricing_mode"`
	ProviderKey       string                `json:"provider_key"`
	ProviderSymbol    string                `json:"provider_symbol"`
	IncludeInNetWorth *bool                 `json:"include_in_net_worth"`
}

type tradeRequest struct {
	ID           string              `json:"id"`
	Side         portfolio.TradeSide `json:"side"`
	Quantity     string              `json:"quantity"`
	UnitPriceVND int64               `json:"unit_price_vnd"`
	FeeVND       int64               `json:"fee_vnd"`
	OccurredAt   time.Time           `json:"occurred_at"`
	Note         string              `json:"note"`
	BaseVersion  int64               `json:"base_version"`
}

type priceRequest struct {
	ID              string    `json:"id"`
	UnitPriceVND    int64     `json:"unit_price_vnd"`
	PricedAt        time.Time `json:"priced_at"`
	Source          string    `json:"source"`
	ProviderQuoteID string    `json:"provider_quote_id"`
	BaseVersion     int64     `json:"base_version"`
}

type versionRequest struct {
	BaseVersion int64 `json:"base_version"`
}

func assets(cfg config.Config, repo PortfolioRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Portfolio repository unavailable", correlationID(r.Context())))
			return
		}
		switch r.Method {
		case http.MethodGet:
			includeArchived := r.URL.Query().Get("include_archived") == "true"
			assets, err := repo.ListPositions(r.Context(), userID, includeArchived)
			if err != nil {
				writePortfolioError(w, r, err, "Assets unavailable")
				return
			}
			writeJSON(w, http.StatusOK, assetsResponse{Status: "ok", Assets: assets, CorrelationID: correlationID(r.Context())})
		case http.MethodPost:
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req createAssetRequest
				if !decodeJSON(w, r, &req) {
					return
				}
				asset, err := repo.CreatePosition(r.Context(), userID, portfolio.CreatePositionInput{
					ID: req.ID, Type: req.Type, Symbol: req.Symbol, Exchange: req.Exchange,
					Name: req.Name, Unit: req.Unit, PricingMode: req.PricingMode,
					ProviderKey: req.ProviderKey, ProviderSymbol: req.ProviderSymbol,
					IncludeInNetWorth: req.IncludeInNetWorth,
				})
				if err != nil {
					writePortfolioError(w, r, err, "Create asset failed")
					return
				}
				writeJSON(w, http.StatusCreated, assetResponse{Status: "ok", Asset: asset, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func assetByID(cfg config.Config, repo PortfolioRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Portfolio repository unavailable", correlationID(r.Context())))
			return
		}
		assetID, action, relatedID, valid := parseAssetPath(r.URL.Path)
		if !valid {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Asset not found", correlationID(r.Context())))
			return
		}
		switch {
		case r.Method == http.MethodGet && action == "":
			asset, err := repo.GetPosition(r.Context(), userID, assetID)
			if err != nil {
				writePortfolioError(w, r, err, "Asset unavailable")
				return
			}
			writeJSON(w, http.StatusOK, assetResponse{Status: "ok", Asset: asset, CorrelationID: correlationID(r.Context())})
		case r.Method == http.MethodPost && action == "archive":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req versionRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				if err := repo.ArchivePosition(r.Context(), userID, assetID, req.BaseVersion); err != nil {
					writePortfolioError(w, r, err, "Archive asset failed")
					return
				}
				writeJSON(w, http.StatusOK, commandResponse{Status: "ok", CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "trades" && relatedID == "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req tradeRequest
				if !decodeJSON(w, r, &req) {
					return
				}
				asset, err := repo.AddTrade(r.Context(), userID, assetID, portfolio.AddTradeInput{
					ID: req.ID, Side: req.Side, Quantity: req.Quantity, UnitPriceVND: req.UnitPriceVND,
					FeeVND: req.FeeVND, OccurredAt: req.OccurredAt, Note: req.Note, BaseVersion: req.BaseVersion,
				})
				if err != nil {
					writePortfolioError(w, r, err, "Add asset trade failed")
					return
				}
				writeJSON(w, http.StatusCreated, assetResponse{Status: "ok", Asset: asset, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPatch && action == "trades" && relatedID != "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req tradeRequest
				if !decodeJSON(w, r, &req) {
					return
				}
				asset, err := repo.UpdateTrade(r.Context(), userID, assetID, relatedID, portfolio.UpdateTradeInput{
					Side: req.Side, Quantity: req.Quantity, UnitPriceVND: req.UnitPriceVND,
					FeeVND: req.FeeVND, OccurredAt: req.OccurredAt, Note: req.Note, BaseVersion: req.BaseVersion,
				})
				if err != nil {
					writePortfolioError(w, r, err, "Update asset trade failed")
					return
				}
				writeJSON(w, http.StatusOK, assetResponse{Status: "ok", Asset: asset, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "archive-trade" && relatedID != "":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req versionRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				asset, err := repo.ArchiveTrade(r.Context(), userID, assetID, relatedID, req.BaseVersion)
				if err != nil {
					writePortfolioError(w, r, err, "Archive asset trade failed")
					return
				}
				writeJSON(w, http.StatusOK, assetResponse{Status: "ok", Asset: asset, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		case r.Method == http.MethodPost && action == "prices":
			requireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req priceRequest
				if !decodeJSON(w, r, &req) {
					return
				}
				source := strings.TrimSpace(req.Source)
				if source == "" {
					source = "manual"
				}
				asset, err := repo.AddPrice(r.Context(), userID, assetID, portfolio.AddPriceInput{
					ID: req.ID, UnitPriceVND: req.UnitPriceVND, PricedAt: req.PricedAt,
					Source: source, ProviderQuoteID: req.ProviderQuoteID, BaseVersion: req.BaseVersion,
				})
				if err != nil {
					writePortfolioError(w, r, err, "Add asset price failed")
					return
				}
				writeJSON(w, http.StatusCreated, assetResponse{Status: "ok", Asset: asset, CorrelationID: correlationID(r.Context())})
			})).ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		}
	}
}

func portfolioSummary(cfg config.Config, repo PortfolioRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Portfolio repository unavailable", correlationID(r.Context())))
			return
		}
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		summary, err := repo.Summary(r.Context(), userID)
		if err != nil {
			writePortfolioError(w, r, err, "Portfolio summary unavailable")
			return
		}
		writeJSON(w, http.StatusOK, portfolioSummaryResponse{Status: "ok", Summary: summary, CorrelationID: correlationID(r.Context())})
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid JSON body", correlationID(r.Context())))
		return false
	}
	return true
}

func parseAssetPath(path string) (assetID string, action string, relatedID string, valid bool) {
	rest := strings.TrimPrefix(path, "/api/v1/assets/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", "", true
	}
	if len(parts) == 2 && parts[0] != "" {
		switch parts[1] {
		case "archive", "trades", "prices":
			return parts[0], parts[1], "", true
		}
	}
	if len(parts) == 3 && parts[0] != "" && parts[1] == "trades" && parts[2] != "" {
		return parts[0], "trades", parts[2], true
	}
	if len(parts) == 4 && parts[0] != "" && parts[1] == "trades" && parts[2] != "" && parts[3] == "archive" {
		return parts[0], "archive-trade", parts[2], true
	}
	return "", "", "", false
}

func writePortfolioError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, portfolio.ErrValidation):
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
	case errors.Is(err, portfolio.ErrOversell):
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope("ASSET_OVERSELL", "Asset quantity is not enough to sell", correlationID(r.Context())))
	case errors.Is(err, portfolio.ErrConflict):
		writeJSON(w, http.StatusConflict, ErrorEnvelope("ASSET_VERSION_CONFLICT", "Asset version conflict", correlationID(r.Context())))
	case errors.Is(err, portfolio.ErrForbidden):
		writeJSON(w, http.StatusForbidden, ErrorEnvelope("FORBIDDEN", "Forbidden", correlationID(r.Context())))
	default:
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", fallback, correlationID(r.Context())))
	}
}

func baseVersionFromQuery(r *http.Request) int64 {
	value, _ := strconv.ParseInt(r.URL.Query().Get("base_version"), 10, 64)
	return value
}
