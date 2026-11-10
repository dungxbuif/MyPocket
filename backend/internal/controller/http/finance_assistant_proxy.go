package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/usecase"
)

// FinanceAssistantProxy preserves the native advisor facade while moving
// execution and conversation ownership to the private Python service.
type FinanceAssistantProxy struct {
	BaseURL     string
	Secret      string
	Client      *http.Client
	Now         func() time.Time
	Delegations *usecase.AssistantDelegationService
}

type financeAssistantClaims struct {
	Actions      []string `json:"actions"`
	CredentialID string   `json:"credential_id,omitempty"`
	jwt.RegisteredClaims
}

func (p *FinanceAssistantProxy) Enabled() bool {
	return p != nil && strings.TrimSpace(p.BaseURL) != "" && len(p.Secret) >= 32
}

func (p *FinanceAssistantProxy) now() time.Time {
	if p.Now != nil {
		return p.Now().UTC()
	}
	return time.Now().UTC()
}

func (p *FinanceAssistantProxy) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return &http.Client{Timeout: 250 * time.Second}
}

func (p *FinanceAssistantProxy) assertion(owner, credentialID string) (string, error) {
	now := p.now()
	claims := financeAssistantClaims{
		Actions:      []string{"assistant:chat", "assistant:channels"},
		CredentialID: credentialID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "mypocket", Subject: owner, Audience: jwt.ClaimStrings{"finance-assistant"},
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)), ID: uuid.NewString(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(p.Secret))
}

func (r *Router) RegisterFinanceAssistantProxyRoutes(proxy *FinanceAssistantProxy, legacy *AdvisorHandler) {
	g := r.Engine.Group(advisorRoute)
	g.Use(r.AuthMiddleware.RequireAdvisorAuth)
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/capabilities"},
		{http.MethodGet, "/conversation/messages"}, {http.MethodPost, "/conversation/reset"},
		{http.MethodPost, "/messages"}, {http.MethodPost, "/messages/async"},
		{http.MethodGet, "/runs/:id"}, {http.MethodGet, "/runs/:id/events"}, {http.MethodPost, "/runs/:id/cancel"},
		{http.MethodGet, "/channels"}, {http.MethodPost, "/channels/telegram"}, {http.MethodPost, "/channels/:adapter_id/pairings"}, {http.MethodDelete, "/channels/:adapter_id"},
	} {
		g.Handle(route.method, route.path, proxy.Forward)
	}
	// Deterministic overview and legacy proposal confirmation remain ledger-side
	// until an external draft exists; neither invokes the in-process model.
	g.GET("/overview", legacy.Overview)
	g.POST("/actions/:action_id/confirm", legacy.ConfirmAction)
	g.PATCH("/actions/:action_id/replace", legacy.ReplaceAction)
	g.POST("/entries", legacy.EntryPost)
	g.GET("/entries/:id", legacy.EntryGet)
	g.POST("/entries/:id/cancel", legacy.EntryCancel)
	g.POST("/entries/:id/decisions", legacy.EntryDecision)
}

func (p *FinanceAssistantProxy) Forward(c *gin.Context) {
	principal, ok := advisorPrincipalFromContext(c)
	if !ok || strings.TrimSpace(principal.OwnerID) == "" || !p.Enabled() {
		Fail(c, http.StatusServiceUnavailable, Problem{Code: "finance_assistant_unavailable", Title: problemTitleInternalServer, Detail: "Finance Assistant chưa sẵn sàng."})
		return
	}
	assertion, err := p.assertion(principal.OwnerID, principal.CredentialID)
	if err != nil {
		advisorFailure(c, err)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 20*1024*1024))
	if err != nil {
		advisorBadRequest(c)
		return
	}
	base, err := url.Parse(strings.TrimRight(p.BaseURL, "/"))
	if err != nil {
		advisorFailure(c, err)
		return
	}
	base.Path = c.Request.URL.Path
	if strings.HasPrefix(base.Path, advisorRoute+"/channels") {
		base.Path = "/v1" + strings.TrimPrefix(base.Path, advisorRoute)
	}
	base.RawQuery = c.Request.URL.RawQuery
	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, base.String(), bytes.NewReader(body))
	if err != nil {
		advisorFailure(c, err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+assertion)
	for _, name := range []string{"Content-Type", "Accept", "Idempotency-Key", "Last-Event-ID"} {
		if value := c.GetHeader(name); value != "" {
			req.Header.Set(name, value)
		}
	}
	req.Header.Set("X-Request-ID", requestID(c))
	var issued usecase.AssistantDelegationCreated
	persistDelegation := c.Request.Method == http.MethodPost && c.FullPath() == advisorRoute+"/channels/telegram"
	ephemeralDelegation := c.Request.Method == http.MethodPost && (c.FullPath() == advisorRoute+"/messages" || c.FullPath() == advisorRoute+"/messages/async")
	if persistDelegation || ephemeralDelegation {
		if p.Delegations == nil {
			Fail(c, http.StatusServiceUnavailable, Problem{Code: "finance_assistant_delegation_unavailable", Title: problemTitleInternalServer, Detail: "Finance Assistant chưa thể truy cập dữ liệu tài chính."})
			return
		}
		ttl := 10 * time.Minute
		if persistDelegation {
			ttl = 90 * 24 * time.Hour
		}
		issued, err = p.Delegations.Create(c.Request.Context(), principal.OwnerID, ttl)
		if err != nil {
			advisorFailure(c, err)
			return
		}
		req.Header.Set("X-MyPocket-Delegation", issued.Secret)
		req.Header.Set("X-MyPocket-Delegation-ID", issued.Delegation.ID)
	}
	response, err := p.client().Do(req)
	if err != nil {
		if issued.Delegation.ID != "" {
			_ = p.Delegations.Revoke(c.Request.Context(), principal.OwnerID, issued.Delegation.ID)
		}
		advisorFailure(c, err)
		return
	}
	defer response.Body.Close()
	if issued.Delegation.ID != "" && (response.StatusCode >= http.StatusBadRequest || ephemeralDelegation) {
		_ = p.Delegations.Revoke(c.Request.Context(), principal.OwnerID, issued.Delegation.ID)
	}
	if response.StatusCode < http.StatusBadRequest {
		for _, header := range []string{"X-MyPocket-Previous-Delegation-ID", "X-MyPocket-Revoke-Delegation-ID"} {
			if credentialID := strings.TrimSpace(response.Header.Get(header)); credentialID != "" {
				_ = p.Delegations.Revoke(c.Request.Context(), principal.OwnerID, credentialID)
			}
		}
	}
	for _, name := range []string{"Content-Type", "Cache-Control"} {
		if value := response.Header.Get(name); value != "" {
			c.Header(name, value)
		}
	}
	if response.StatusCode != http.StatusNoContent && strings.HasPrefix(c.FullPath(), advisorRoute+"/channels") {
		payload, readErr := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
		if readErr != nil || len(payload) > 2*1024*1024 || (response.StatusCode < http.StatusBadRequest && !json.Valid(payload)) {
			if readErr == nil {
				readErr = errors.New("finance assistant returned malformed channel payload")
			}
			advisorFailure(c, readErr)
			return
		}
		if response.StatusCode < http.StatusBadRequest {
			wrapped := make([]byte, 0, len(payload)+10)
			wrapped = append(wrapped, []byte(`{"data":`)...)
			wrapped = append(wrapped, payload...)
			wrapped = append(wrapped, '}')
			c.Data(response.StatusCode, "application/json", wrapped)
			return
		}
		c.Data(response.StatusCode, response.Header.Get("Content-Type"), payload)
		return
	}
	c.Status(response.StatusCode)
	_, _ = io.Copy(c.Writer, response.Body)
}
