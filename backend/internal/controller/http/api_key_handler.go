package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/usecase"
)

const apiKeysRoute = "/api/v1/api-keys"

type APIKeyHandler struct{ Service *usecase.UserAPIKeyService }

type apiKeyCreateInput struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expires_at,omitempty"`
}

type apiKeyCreatedResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Secret    string     `json:"secret"`
}

func (r *Router) RegisterAPIKeyRoutes(h *APIKeyHandler) {
	g := r.Engine.Group(apiKeysRoute)
	g.Use(r.AuthMiddleware.RequireAuth)
	g.POST("", h.Create)
	g.GET("", h.List)
	g.DELETE("/:id", h.Revoke)
}

// APIKeyCreate godoc
// @Summary Create a user API key
// @Tags API Keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body apiKeyCreateInput true "API key configuration"
// @Success 201 {object} apiKeyCreatedResponse
// @Router /api/v1/api-keys [post]
func (h *APIKeyHandler) Create(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil {
		apiKeyFailure(c, usecase.ErrAPIKeyUnavailable)
		return
	}
	var input apiKeyCreateInput
	if c.ShouldBindJSON(&input) != nil {
		apiKeyBadRequest(c)
		return
	}
	var expiresAt *time.Time
	if strings.TrimSpace(input.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(input.ExpiresAt))
		if err != nil {
			apiKeyBadRequest(c)
			return
		}
		expiresAt = &parsed
	}
	created, err := h.Service.Create(c.Request.Context(), usecase.APIKeyCreateInput{OwnerID: owner, Name: input.Name, Scopes: input.Scopes, ExpiresAt: expiresAt})
	if err != nil {
		apiKeyFailure(c, err)
		return
	}
	Created(c, apiKeyCreatedResponse{ID: created.Key.ID, Name: created.Key.Name, Scopes: created.Key.Scopes, ExpiresAt: created.Key.ExpiresAt, Secret: created.Secret})
}

// APIKeyList godoc
// @Summary List current user's API keys
// @Tags API Keys
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.UserAPIKey
// @Router /api/v1/api-keys [get]
func (h *APIKeyHandler) List(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil {
		apiKeyFailure(c, usecase.ErrAPIKeyUnavailable)
		return
	}
	keys, err := h.Service.List(c.Request.Context(), owner)
	if err != nil {
		apiKeyFailure(c, err)
		return
	}
	OK(c, keys)
}

// APIKeyRevoke godoc
// @Summary Revoke a current user's API key
// @Tags API Keys
// @Produce json
// @Security BearerAuth
// @Param id path string true "API key ID"
// @Success 204
// @Router /api/v1/api-keys/{id} [delete]
func (h *APIKeyHandler) Revoke(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h == nil || h.Service == nil {
		apiKeyFailure(c, usecase.ErrAPIKeyUnavailable)
		return
	}
	if err := h.Service.Revoke(c.Request.Context(), owner, c.Param("id")); err != nil {
		apiKeyFailure(c, err)
		return
	}
	NoContent(c)
}

func apiKeyBadRequest(c *gin.Context) {
	Fail(c, http.StatusBadRequest, Problem{Code: "api_key_invalid", Title: "Invalid API key request", Detail: "cấu hình API key không hợp lệ"})
}

func apiKeyFailure(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "api_key_failed"
	if errors.Is(err, usecase.ErrAPIKeyNameInvalid) || errors.Is(err, usecase.ErrAPIKeyExpiryInvalid) || errors.Is(err, entity.ErrAPIKeyInvalid) {
		status, code = http.StatusBadRequest, "api_key_invalid"
	} else if errors.Is(err, usecase.ErrAPIKeyUnavailable) {
		status, code = http.StatusServiceUnavailable, "api_key_unavailable"
	} else if errors.Is(err, usecase.ErrAPIKeyInvalid) {
		status, code = http.StatusUnauthorized, "api_key_invalid"
	}
	Fail(c, status, Problem{Code: code, Title: "API key request failed", Detail: "không thể xử lý API key"})
}
