package ai

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/usecase"
)

type ReportAnalysisProvider struct{ client *AdvisorClient }

func NewReportAnalysisProvider(client *AdvisorClient) *ReportAnalysisProvider {
	return &ReportAnalysisProvider{client: client}
}

func (p *ReportAnalysisProvider) Available() bool {
	return p != nil && p.client != nil && p.client.Available()
}

func (p *ReportAnalysisProvider) Analyze(ctx context.Context, facts entity.ReportFactBundle) (entity.ReportProviderResult, error) {
	if !p.Available() {
		return entity.ReportProviderResult{}, usecase.ErrAdvisorProviderUnavailable
	}
	// The provider receives bounded, immutable facts only. It never receives
	// transaction rows, notes, images, or a full ledger.
	payload, err := json.Marshal(facts)
	if err != nil {
		return entity.ReportProviderResult{}, err
	}
	response, err := p.client.Chat(ctx, usecase.AdvisorRequest{Messages: []usecase.AdvisorChatMessage{
		{Role: "system", Content: "Bạn là trợ lý phân tích chi tiêu. Chỉ dùng các fact_ref trong JSON. Trả về JSON đúng schema {observations:[{fact_ref,direction,text}],suggestions:[{fact_ref,text}]}. Không đưa số, phần trăm, tên giao dịch, ghi chú, nguyên nhân, hay khẳng định nhân quả vào text; text chỉ nên là nhận xét định tính ngắn gọn bằng tiếng Việt. Tối đa 3 observations và 3 suggestions."},
		{Role: "user", Content: string(payload)},
	}})
	if err != nil {
		return entity.ReportProviderResult{}, err
	}
	var result entity.ReportProviderResult
	if err := json.Unmarshal([]byte(response.Text), &result); err != nil {
		return entity.ReportProviderResult{}, errors.Join(entity.ErrReportAnalysisOutputInvalid, err)
	}
	result.Usage = response.Usage
	return result, nil
}
