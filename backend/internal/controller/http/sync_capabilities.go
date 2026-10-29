package httpapi

import (
	"github.com/gin-gonic/gin"
	transactionrepo "github.com/mypocket/backend/internal/repository"
)

type SyncCapabilitiesResponse struct {
	ContractVersion                  int      `json:"contract_version"`
	Enabled                          bool     `json:"enabled"`
	Operations                       []string `json:"operations"`
	ReceiptMetadata                  string   `json:"receipt_metadata"`
	IdempotencyKeyRequiredForOffline bool     `json:"idempotency_key_required_for_offline"`
	LegacyUnkeyedOnline              bool     `json:"legacy_unkeyed_online"`
	TransactionCreateContext         string   `json:"transaction_create_context"`
}

// SyncCapabilities godoc
// @Summary Inspect durable offline-create contract support
// @Description JWT/session or full-owner API key required. This read-only endpoint advertises configured implementation, not database health. Old-server404/disabled/unsupported version must block outbox sends. No-header online creates remain legacy and unsafe to retry.
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=SyncCapabilitiesResponse}
// @Failure 401 {object} Problem
// @Router /api/v1/sync/capabilities [get]
func (h *TransactionHandler) SyncCapabilities(c *gin.Context) {
	// Capability state is deployment-sensitive: caching a previous enabled
	// response could authorize an offline drain after the server has rolled the
	// contract back. Keep this read fresh at every foreground sync attempt.
	c.Header("Cache-Control", "no-store")
	response := SyncCapabilitiesResponse{
		ContractVersion: 1, Operations: []string{}, ReceiptMetadata: "/meta/mutation",
		IdempotencyKeyRequiredForOffline: true, LegacyUnkeyedOnline: true,
	}
	if h != nil && h.Transactions != nil && h.Transfers != nil {
		if _, ok := h.Transactions.(transactionrepo.TransactionCreateMutationRepository); ok {
			response.Enabled = true
			response.TransactionCreateContext = "explicit_travel_v1"
			response.Operations = []string{"transaction.create", "transfer.create"}
		}
	}
	OK(c, response)
}
