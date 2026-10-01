package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	portfoliorepo "github.com/mypocket/backend/internal/repository"
)

type PortfolioHandler struct {
	Portfolio portfoliorepo.PortfolioRepository
}

type portfolioAssetInput struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type portfolioPriceInput struct {
	LatestPrice *int64 `json:"latest_price"`
}

type portfolioTradeInput struct {
	Side       string  `json:"side"`
	Quantity   string  `json:"quantity"`
	UnitPrice  *int64  `json:"unit_price"`
	Fee        *int64  `json:"fee"`
	OccurredAt string  `json:"occurred_at"`
	Note       *string `json:"note"`
}

func (r *Router) RegisterPortfolioRoutes(h *PortfolioHandler) {
	g := r.Engine.Group("/api/v1/portfolio")
	g.Use(r.AuthMiddleware.RequireAuth)
	g.GET("/summary", h.Summary)
	g.GET("/assets", h.ListAssets)
	g.POST("/assets", h.CreateAsset)
	g.PATCH("/assets/:id/price", h.UpdatePrice)
	g.GET("/assets/:id/trades", h.ListTrades)
	g.POST("/assets/:id/trades", h.CreateTrade)
}

// PortfolioSummary godoc
// @Summary Read owner-scoped portfolio positions and valuation
// @Tags Portfolio
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.PortfolioSummary
// @Failure 401 {object} Problem
// @Failure 500 {object} Problem
// @Router /api/v1/portfolio/summary [get]
func (h *PortfolioHandler) Summary(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	result, err := h.Portfolio.Summary(owner)
	if err != nil {
		portfolioFailure(c)
		return
	}
	OK(c, result)
}

// ListPortfolioAssets godoc
// @Summary List owner-scoped portfolio assets
// @Tags Portfolio
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.PortfolioAsset
// @Router /api/v1/portfolio/assets [get]
func (h *PortfolioHandler) ListAssets(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	rows, err := h.Portfolio.ListAssets(owner)
	if err != nil {
		portfolioFailure(c)
		return
	}
	OK(c, rows)
}

// CreatePortfolioAsset godoc
// @Summary Create a portfolio asset
// @Tags Portfolio
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param asset body portfolioAssetInput true "Portfolio asset"
// @Success 201 {object} entity.PortfolioAsset
// @Failure 400 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/portfolio/assets [post]
func (h *PortfolioHandler) CreateAsset(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	var input portfolioAssetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return
	}
	asset, err := h.Portfolio.CreateAsset(owner, portfoliorepo.PortfolioAssetInput{Symbol: strings.TrimSpace(input.Symbol), Name: strings.TrimSpace(input.Name)})
	if err != nil {
		portfolioMapError(c, err)
		return
	}
	Created(c, asset)
}

// UpdatePortfolioPrice godoc
// @Summary Update or clear the latest manual portfolio price
// @Tags Portfolio
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Param price body portfolioPriceInput true "Latest price; null clears it"
// @Success 200 {object} entity.PortfolioAsset
// @Router /api/v1/portfolio/assets/{id}/price [patch]
func (h *PortfolioHandler) UpdatePrice(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	var input portfolioPriceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return
	}
	updated, err := h.Portfolio.UpdateAssetPrice(owner, strings.TrimSpace(c.Param("id")), portfoliorepo.PortfolioPriceUpdate{LatestPrice: input.LatestPrice, LatestPriceAt: portfolioTimePtr(time.Now().UTC())})
	if err != nil {
		portfolioMapError(c, err)
		return
	}
	OK(c, updated)
}

// ListPortfolioTrades godoc
// @Summary List immutable portfolio trades for an asset
// @Tags Portfolio
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Success 200 {array} entity.PortfolioTrade
// @Router /api/v1/portfolio/assets/{id}/trades [get]
func (h *PortfolioHandler) ListTrades(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	rows, err := h.Portfolio.ListTrades(owner, strings.TrimSpace(c.Param("id")))
	if err != nil {
		portfolioMapError(c, err)
		return
	}
	OK(c, rows)
}

// CreatePortfolioTrade godoc
// @Summary Append a buy or sell trade to a portfolio asset
// @Tags Portfolio
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Param trade body portfolioTradeInput true "Portfolio trade"
// @Success 201 {object} entity.PortfolioTrade
// @Failure 400 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/portfolio/assets/{id}/trades [post]
func (h *PortfolioHandler) CreateTrade(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	var input portfolioTradeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return
	}
	quantity, err := entity.ParsePortfolioQuantity(strings.TrimSpace(input.Quantity))
	if err != nil {
		portfolioBadRequest(c, "Khối lượng phải là số thập phân dương, tối đa 8 chữ số sau dấu phẩy.")
		return
	}
	if input.UnitPrice == nil || *input.UnitPrice < 0 || (input.Fee != nil && *input.Fee < 0) {
		portfolioBadRequest(c, "Giá và phí phải là số nguyên VND không âm.")
		return
	}
	occurredAt := time.Now().UTC()
	if strings.TrimSpace(input.OccurredAt) != "" {
		parsed, parseErr := time.Parse(time.RFC3339, input.OccurredAt)
		if parseErr != nil {
			portfolioBadRequest(c, "Thời điểm giao dịch danh mục không hợp lệ.")
			return
		}
		occurredAt = parsed.UTC()
	}
	fee := int64(0)
	if input.Fee != nil {
		fee = *input.Fee
	}
	trade, err := h.Portfolio.CreateTrade(owner, strings.TrimSpace(c.Param("id")), portfoliorepo.PortfolioTradeInput{Side: strings.TrimSpace(input.Side), QuantityScaled: quantity, UnitPrice: *input.UnitPrice, Fee: fee, OccurredAt: occurredAt, Note: normalizeOptional(input.Note)})
	if err != nil {
		portfolioMapError(c, err)
		return
	}
	Created(c, trade)
}

func portfolioMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, portfoliorepo.ErrPortfolioAssetNotFound), errors.Is(err, portfoliorepo.ErrPortfolioTradeNotFound):
		Fail(c, http.StatusNotFound, Problem{Code: "PORTFOLIO_NOT_FOUND", Title: problemTitleNotFound, Detail: "Không tìm thấy tài sản danh mục."})
	case errors.Is(err, portfoliorepo.ErrPortfolioAssetConflict):
		Fail(c, http.StatusConflict, Problem{Code: "PORTFOLIO_ASSET_CONFLICT", Title: problemTitleBadRequest, Detail: "Mã tài sản đã tồn tại trong danh mục."})
	case errors.Is(err, portfoliorepo.ErrPortfolioInsufficientQuantity):
		Fail(c, http.StatusConflict, Problem{Code: "PORTFOLIO_INSUFFICIENT_QUANTITY", Title: problemTitleBadRequest, Detail: "Không thể bán vượt quá khối lượng đang nắm giữ."})
	case errors.Is(err, portfoliorepo.ErrPortfolioAssetInvalid), errors.Is(err, portfoliorepo.ErrPortfolioTradeInvalid):
		portfolioBadRequest(c, "Dữ liệu danh mục chưa hợp lệ.")
	default:
		portfolioFailure(c)
	}
}

func portfolioBadRequest(c *gin.Context, detail string) {
	Fail(c, http.StatusBadRequest, Problem{Code: "PORTFOLIO_BAD_REQUEST", Title: problemTitleBadRequest, Detail: detail})
}

func portfolioFailure(c *gin.Context) {
	Fail(c, http.StatusInternalServerError, Problem{Code: "PORTFOLIO_FAILED", Title: problemTitleInternalServer, Detail: "Không xử lý được danh mục đầu tư."})
}

func portfolioUnavailable(c *gin.Context) {
	Fail(c, http.StatusNotImplemented, Problem{Code: "PORTFOLIO_UNAVAILABLE", Title: problemTitleInternalServer, Detail: "Danh mục đầu tư chưa được cấu hình."})
}

func portfolioTimePtr(value time.Time) *time.Time { return &value }
