package httpapi

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	marketrepo "github.com/mypocket/backend/internal/repository"
)

type AssetMarketHandler struct {
	Market              marketrepo.MarketRepository
	Portfolio           marketrepo.PortfolioRepository
	ConfiguredProviders map[string]struct{}
}

type portfolioPriceSourceInput struct {
	ExpectedVersion int64   `json:"expected_version"`
	PriceMode       string  `json:"price_mode"`
	ProductID       *string `json:"product_id,omitempty"`
}

func (input *portfolioPriceSourceInput) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for key := range fields {
		if key != "expected_version" && key != "price_mode" && key != "product_id" {
			return errors.New("unknown portfolio price source field")
		}
	}
	type plain portfolioPriceSourceInput
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*input = portfolioPriceSourceInput(decoded)
	return nil
}

func (r *Router) RegisterAssetMarketRoutes(h *AssetMarketHandler) {
	g := r.Engine.Group("/api/v1/portfolio")
	g.Use(r.AuthMiddleware.RequireAuth)
	g.GET("/catalog", h.Catalog)
	g.PATCH("/assets/:id/price-source", h.UpdatePriceSource)
}

// Catalog lists only enabled, backend-configured products. It contains no
// owner or holding data.
// @Summary List enabled accumulated-asset market products
// @Tags Portfolio
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.MarketProduct
// @Router /api/v1/portfolio/catalog [get]
func (h *AssetMarketHandler) Catalog(c *gin.Context) {
	if _, ok := transactionOwner(c); !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Market == nil {
		portfolioUnavailable(c)
		return
	}
	products, err := h.Market.ListEnabledMarketProducts()
	if err != nil {
		portfolioFailure(c)
		return
	}
	configured := make([]entity.MarketProduct, 0, len(products))
	for _, product := range products {
		if _, ok := h.ConfiguredProviders[product.ProviderKey]; !ok {
			continue
		}
		configured = append(configured, product)
	}
	OK(c, configured)
}

// UpdatePriceSource binds one holding to one exact catalog product or
// explicitly unlinks it back to manual mode. Product names are never used as
// identity and the asset version fences concurrent edits.
// @Summary Bind or unlink an accumulated-asset price source
// @Tags Portfolio
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Asset ID"
// @Param source body portfolioPriceSourceInput true "Price source binding"
// @Success 200 {object} entity.PortfolioAsset
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/portfolio/assets/{id}/price-source [patch]
func (h *AssetMarketHandler) UpdatePriceSource(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Portfolio == nil {
		portfolioUnavailable(c)
		return
	}
	var input portfolioPriceSourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return
	}
	if input.ProductID != nil {
		value := strings.TrimSpace(*input.ProductID)
		input.ProductID = &value
	}
	mode := strings.TrimSpace(input.PriceMode)
	if mode == "" && input.ProductID != nil {
		mode = "automatic"
	}
	if mode == "automatic" {
		if h.Market == nil || input.ProductID == nil {
			portfolioBadRequest(c, "Nguồn giá tài sản không hợp lệ.")
			return
		}
		product, findErr := h.Market.FindMarketProduct(*input.ProductID)
		if findErr != nil {
			portfolioMapError(c, findErr)
			return
		}
		if _, configured := h.ConfiguredProviders[product.ProviderKey]; !configured {
			portfolioBadRequest(c, "Nguồn giá tài sản chưa được cấu hình.")
			return
		}
	}
	updated, err := h.Portfolio.UpdateAssetPriceSource(owner, strings.TrimSpace(c.Param("id")), marketrepo.PortfolioPriceSourceInput{ExpectedVersion: input.ExpectedVersion, PriceMode: mode, MarketProductID: input.ProductID})
	if err != nil {
		portfolioMapError(c, err)
		return
	}
	OK(c, updated)
}
