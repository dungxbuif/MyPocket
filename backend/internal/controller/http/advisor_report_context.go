package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
)

func (h *AdvisorHandler) messageParts(ctx context.Context, principal usecase.Principal, input advisorMessageInput) ([]entity.AdvisorPart, string, error) {
	parts := []entity.AdvisorPart{{Type: "text", Text: input.Text}}
	payload := []byte(input.Text) // Preserve hashes for legacy requests/replays.
	if input.ReportContext != nil {
		if h.Reports == nil {
			return nil, "", repository.ErrNotFound
		}
		reportContext, err := h.Reports.FollowUpContext(ctx, principal, *input.ReportContext)
		if err != nil {
			return nil, "", err
		}
		parts = append(parts, entity.AdvisorPart{Type: "report_context", Data: reportContext})
		payload, err = json.Marshal(struct {
			Text    string
			Context *usecase.ReportFollowUpRequest
		}{input.Text, input.ReportContext})
		if err != nil {
			return nil, "", err
		}
	}
	if input.JarContext != nil {
		jarContext, err := h.jarContext(ctx, principal, *input.JarContext)
		if err != nil {
			return nil, "", err
		}
		parts = append(parts, entity.AdvisorPart{Type: "jar_context", Data: jarContext})
		payload, err = json.Marshal(struct {
			Payload []byte
			Jar     *advisorJarContextInput
		}{payload, input.JarContext})
		if err != nil {
			return nil, "", err
		}
	}
	hash := sha256.Sum256(payload)
	return parts, hex.EncodeToString(hash[:]), nil
}
