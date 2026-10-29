package httpapi

import (
	"encoding/json"
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

type portfolioAssetUpdateInput struct {
	Symbol *string `json:"symbol"`
	Name   *string `json:"name"`
}

type portfolioPriceInput struct {
	LatestPrice  *int64 `json:"latest_price" validate:"required" extensions:"x-nullable"`
	pricePresent bool
}

func (input *portfolioPriceInput) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	price, present := fields["latest_price"]
	input.pricePresent = present
	if !present {
		return nil
	}
	return json.Unmarshal(price, &input.LatestPrice)
}

type portfolioTradeInput struct {
	Side       string  `json:"side"`
	Quantity   string  `json:"quantity"`
	UnitPrice  *int64  `json:"unit_price"`
	Fee        *int64  `json:"fee"`
	OccurredAt string  `json:"occurred_at"`
	Note       *string `json:"note"`
}

type portfolioTradeCorrectionInput struct {
	SourceTradeID        string                          `json:"source_trade_id"`
	ExpectedAssetVersion int64                           `json:"expected_asset_version"`
	Operation            string                          `json:"operation"`
	Reason               string                          `json:"reason"`
	Replacement          *portfolioTradeCorrectionFields `json:"replacement,omitempty"`
}

type portfolioTradeCorrectionFields struct {
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
	g.PATCH("/assets/:id", h.UpdateAsset)
	g.DELETE("/assets/:id", h.DeleteAsset)
	g.PATCH("/assets/:id/price", h.UpdatePrice)
	g.GET("/assets/:id/trades", h.ListTrades)
	g.POST("/assets/:id/trades", h.CreateTrade)
	g.POST("/assets/:id/trade-corrections/preview", h.PreviewTradeCorrection)
	g.POST("/assets/:id/trade-corrections", h.CorrectTrade)
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

// UpdatePortfolioAsset godoc
// @Summary Update portfolio asset metadata
// @Description Updates symbol and/or name only. Portfolio trades remain immutable and are not rewritten.
// @Tags Portfolio
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Param asset body portfolioAssetUpdateInput true "Metadata fields to update"
// @Success 200 {object} entity.PortfolioAsset
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/portfolio/assets/{id} [patch]
func (h *PortfolioHandler) UpdateAsset(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	var input portfolioAssetUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return
	}
	if input.Symbol == nil && input.Name == nil {
		portfolioBadRequest(c, "Cần cung cấp symbol hoặc name để cập nhật tài sản.")
		return
	}
	updated, err := h.Portfolio.UpdateAsset(owner, strings.TrimSpace(c.Param("id")), portfoliorepo.PortfolioAssetUpdate{Symbol: input.Symbol, Name: input.Name})
	if err != nil {
		portfolioMapError(c, err)
		return
	}
	OK(c, updated)
}

// DeletePortfolioAsset godoc
// @Summary Delete an unused portfolio asset
// @Tags Portfolio
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Success 204
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/portfolio/assets/{id} [delete]
func (h *PortfolioHandler) DeleteAsset(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	if err := h.Portfolio.DeleteAsset(owner, strings.TrimSpace(c.Param("id"))); err != nil {
		switch {
		case errors.Is(err, portfoliorepo.ErrPortfolioAssetNotFound):
			portfolioMapError(c, err)
		case errors.Is(err, portfoliorepo.ErrPortfolioAssetHasTrades):
			Fail(c, http.StatusConflict, Problem{Code: "portfolio_asset_has_trades", Title: "Conflict", Detail: "Không thể xóa tài sản đã có giao dịch; hãy dùng giao dịch bù trừ."})
		case errors.Is(err, portfoliorepo.ErrPortfolioAssetInvalid):
			transactionBadRequest(c, "tài sản đầu tư không hợp lệ")
		default:
			portfolioFailure(c)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdatePortfolioPrice godoc
// @Summary Update or clear the latest manual portfolio price
// @Description latest_price must be present: null clears the price, zero records a known zero. Omitting the field or sending a null body returns 400 without modifying valuation.
// @Tags Portfolio
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Param price body portfolioPriceInput true "Latest price; null clears it"
// @Success 200 {object} entity.PortfolioAsset
// @Failure 400 {object} Problem
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
	if !input.pricePresent {
		portfolioBadRequest(c, "Cần cung cấp latest_price; dùng null để xóa giá hiện tại.")
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

// PreviewPortfolioTradeCorrection validates a revision and returns the
// complete before/after projection without writing a trade or revision.
// @Summary Preview an immutable portfolio trade correction
// @Tags Portfolio
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Param correction body portfolioTradeCorrectionInput true "Correction"
// @Success 200 {object} repository.PortfolioTradeCorrectionResult
// @Router /api/v1/portfolio/assets/{id}/trade-corrections/preview [post]
func (h *PortfolioHandler) PreviewTradeCorrection(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	input, err := decodePortfolioCorrection(c)
	if err != nil {
		portfolioBadRequest(c, err.Error())
		return
	}
	result, err := h.Portfolio.PreviewTradeCorrection(owner, strings.TrimSpace(c.Param("id")), input)
	if err != nil {
		portfolioMapError(c, err)
		return
	}
	OK(c, result)
}

// CorrectPortfolioTrade appends an immutable replace/void revision. The
// expected asset version and Idempotency-Key make retries and stale previews
// deterministic.
// @Summary Commit a portfolio trade correction
// @Tags Portfolio
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string true "Stable correction key"
// @Param id path string true "Asset ID"
// @Param correction body portfolioTradeCorrectionInput true "Correction"
// @Success 200 {object} repository.PortfolioTradeCorrectionResult
// @Router /api/v1/portfolio/assets/{id}/trade-corrections [post]
func (h *PortfolioHandler) CorrectTrade(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if key == "" {
		portfolioBadRequest(c, "Cần cung cấp Idempotency-Key cho thao tác sửa giao dịch.")
		return
	}
	input, err := decodePortfolioCorrection(c)
	if err != nil {
		portfolioBadRequest(c, err.Error())
		return
	}
	result, replayed, err := h.Portfolio.CorrectTrade(owner, strings.TrimSpace(c.Param("id")), input, key, postgresPortfolioCorrectionHash(input))
	if err != nil {
		portfolioMapError(c, err)
		return
	}
	if replayed {
		c.Header("X-Idempotent-Replay", "true")
	}
	OK(c, result)
}

func decodePortfolioCorrection(c *gin.Context) (portfoliorepo.PortfolioTradeCorrectionInput, error) {
	var raw portfolioTradeCorrectionInput
	if err := c.ShouldBindJSON(&raw); err != nil {
		return portfoliorepo.PortfolioTradeCorrectionInput{}, errors.New(problemDetailInvalidJSON)
	}
	input := portfoliorepo.PortfolioTradeCorrectionInput{SourceTradeID: strings.TrimSpace(raw.SourceTradeID), ExpectedAssetVersion: raw.ExpectedAssetVersion, Operation: strings.TrimSpace(raw.Operation), Reason: strings.TrimSpace(raw.Reason)}
	if raw.Replacement == nil {
		return input, nil
	}
	quantity, err := entity.ParsePortfolioQuantity(strings.TrimSpace(raw.Replacement.Quantity))
	if err != nil {
		return input, errors.New("khối lượng thay thế không hợp lệ")
	}
	if raw.Replacement.UnitPrice == nil || *raw.Replacement.UnitPrice < 0 || (raw.Replacement.Fee != nil && *raw.Replacement.Fee < 0) {
		return input, errors.New("giá và phí thay thế không hợp lệ")
	}
	occurredAt := time.Now().UTC()
	if strings.TrimSpace(raw.Replacement.OccurredAt) != "" {
		parsed, parseErr := time.Parse(time.RFC3339, raw.Replacement.OccurredAt)
		if parseErr != nil {
			return input, errors.New("thời điểm thay thế không hợp lệ")
		}
		occurredAt = parsed.UTC()
	}
	fee := int64(0)
	if raw.Replacement.Fee != nil {
		fee = *raw.Replacement.Fee
	}
	input.Replacement = &portfoliorepo.PortfolioTradeInput{Side: strings.TrimSpace(raw.Replacement.Side), QuantityScaled: quantity, UnitPrice: *raw.Replacement.UnitPrice, Fee: fee, OccurredAt: occurredAt, Note: normalizeOptional(raw.Replacement.Note)}
	return input, nil
}

func postgresPortfolioCorrectionHash(input portfoliorepo.PortfolioTradeCorrectionInput) string {
	return portfoliorepo.PortfolioCorrectionHash(input)
}

func portfolioMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, portfoliorepo.ErrPortfolioAssetNotFound), errors.Is(err, portfoliorepo.ErrPortfolioTradeNotFound):
		Fail(c, http.StatusNotFound, Problem{Code: "PORTFOLIO_NOT_FOUND", Title: problemTitleNotFound, Detail: "Không tìm thấy tài sản danh mục."})
	case errors.Is(err, portfoliorepo.ErrPortfolioAssetConflict):
		Fail(c, http.StatusConflict, Problem{Code: "PORTFOLIO_ASSET_CONFLICT", Title: problemTitleBadRequest, Detail: "Mã tài sản đã tồn tại trong danh mục."})
	case errors.Is(err, portfoliorepo.ErrPortfolioAssetHasTrades):
		Fail(c, http.StatusConflict, Problem{Code: "PORTFOLIO_ASSET_HAS_TRADES", Title: problemTitleBadRequest, Detail: "Không thể xóa tài sản đã có giao dịch; hãy dùng giao dịch bù trừ."})
	case errors.Is(err, portfoliorepo.ErrPortfolioInsufficientQuantity):
		Fail(c, http.StatusConflict, Problem{Code: "PORTFOLIO_INSUFFICIENT_QUANTITY", Title: problemTitleBadRequest, Detail: "Không thể bán vượt quá khối lượng đang nắm giữ."})
	case errors.Is(err, portfoliorepo.ErrPortfolioAssetConflictVersion), errors.Is(err, portfoliorepo.ErrPortfolioCorrectionConflict):
		Fail(c, http.StatusConflict, Problem{Code: "PORTFOLIO_CORRECTION_CONFLICT", Title: problemTitleBadRequest, Detail: "Dữ liệu danh mục đã thay đổi hoặc yêu cầu sửa bị trùng; hãy tải lại và xem trước lại."})
	case errors.Is(err, portfoliorepo.ErrPortfolioCorrectionPending):
		Fail(c, http.StatusConflict, Problem{Code: "PORTFOLIO_CORRECTION_PENDING", Title: problemTitleBadRequest, Detail: "Yêu cầu sửa giao dịch trước đó vẫn đang xử lý; không tự gửi lại."})
	case errors.Is(err, portfoliorepo.ErrPortfolioCorrectionInvalid), errors.Is(err, entity.ErrPortfolioRevisionInvalid), errors.Is(err, entity.ErrPortfolioRevisionConflict):
		portfolioBadRequest(c, "Dữ liệu sửa giao dịch danh mục chưa hợp lệ.")
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
