package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mypocket/backend/internal/entity"
)

var (
	ErrAdvisorEvidenceInvalid = errors.New("advisor evidence is invalid")
	ErrAdvisorFinalInvalid    = errors.New("advisor final answer is invalid")
)

var advisorFinalTemplates = map[string]struct{}{
	"summary": {}, "period_comparison": {}, "transaction": {},
	"wallet": {}, "budget": {}, "goal": {}, "jar": {}, "proposal": {},
}

var advisorClarifications = map[string]string{
	"":            "Bạn muốn trợ lý hỗ trợ việc gì?",
	"intent":      "Bạn muốn trợ lý hỗ trợ việc gì?",
	"wallet":      "Bạn muốn chọn ví nào?",
	"category":    "Bạn muốn chọn danh mục nào?",
	"period":      "Bạn muốn xem khoảng thời gian nào?",
	"transaction": "Bạn muốn chọn giao dịch nào?",
	"amount":      "Bạn muốn dùng số tiền bao nhiêu?",
}

type AdvisorEvidenceIndex struct {
	facts       map[string]entity.Fact
	help        map[string]productHelpSource
	hasProposal bool
	calls       map[string]bool
}

func NewAdvisorEvidenceIndex() *AdvisorEvidenceIndex {
	return &AdvisorEvidenceIndex{facts: make(map[string]entity.Fact)}
}

func (i *AdvisorEvidenceIndex) Add(callID string, result entity.FinanceResult) error {
	if i == nil || strings.TrimSpace(callID) == "" {
		return ErrAdvisorEvidenceInvalid
	}
	callID = strings.TrimSpace(callID)
	if strings.ContainsAny(callID, ". \t\n") || i.calls[callID] {
		return ErrAdvisorEvidenceInvalid
	}
	if i.calls == nil {
		i.calls = make(map[string]bool)
	}
	i.calls[callID] = true
	if result.Status == "proposed" && isAdvisorProposalView(result.ViewKind) {
		i.hasProposal = true
		return nil
	}
	if result.Status != "" && result.Status != "ok" {
		return nil
	}
	if result.ViewKind == "product_help" {
		var view productHelpView
		if json.Unmarshal(result.View, &view) != nil {
			return ErrAdvisorEvidenceInvalid
		}
		if i.help == nil {
			i.help = make(map[string]productHelpSource)
		}
		for _, source := range view.Sources {
			// Resolve source IDs against the code-owned handbook rather than
			// publishing text supplied by an arbitrary result payload.
			for _, section := range advisorHelpSections {
				if source.ID == section.ID {
					i.help[source.ID] = productHelpSource{ID: section.ID, Title: section.Title, Excerpt: section.Excerpt}
				}
			}
		}
	}
	if i.facts == nil {
		i.facts = make(map[string]entity.Fact)
	}
	for _, fact := range result.Facts {
		if strings.TrimSpace(fact.ID) == "" || fact.Integer == nil && fact.Text == nil {
			continue
		}
		ref := callID + "." + strings.TrimSpace(fact.ID)
		if _, exists := i.facts[ref]; exists {
			return ErrAdvisorEvidenceInvalid
		}
		fact.SourceRef = ref
		i.facts[ref] = fact
	}
	return nil
}

func (i *AdvisorEvidenceIndex) Resolve(ref string) (entity.Fact, error) {
	if i == nil || i.facts == nil {
		return entity.Fact{}, ErrAdvisorEvidenceInvalid
	}
	fact, ok := i.facts[strings.TrimSpace(ref)]
	if !ok {
		return entity.Fact{}, ErrAdvisorEvidenceInvalid
	}
	return fact, nil
}

func ValidateAdvisorFinal(answer entity.AdvisorFinalAnswer, index *AdvisorEvidenceIndex) error {
	switch answer.Mode {
	case "finance":
		if _, ok := advisorFinalTemplates[strings.TrimSpace(answer.TemplateID)]; !ok {
			return ErrAdvisorFinalInvalid
		}
		if answer.TemplateID == "proposal" {
			if index == nil || !index.hasProposal || len(answer.Claims) != 0 {
				return ErrAdvisorFinalInvalid
			}
		} else if len(answer.Claims) == 0 {
			return ErrAdvisorFinalInvalid
		}
	case "product_help", "clarification":
		if len(answer.Claims) != 0 {
			return ErrAdvisorFinalInvalid
		}
		if answer.Mode == "clarification" {
			if _, ok := advisorClarifications[answer.QuestionID]; !ok {
				return ErrAdvisorFinalInvalid
			}
		}
		if answer.Mode == "product_help" && answer.SourceID != "" {
			if index == nil {
				return ErrAdvisorFinalInvalid
			}
			if _, ok := index.help[answer.SourceID]; !ok {
				return ErrAdvisorFinalInvalid
			}
		}
	default:
		return ErrAdvisorFinalInvalid
	}
	seen := make(map[string]struct{}, len(answer.Claims))
	for _, claim := range answer.Claims {
		claim.Slot = strings.TrimSpace(claim.Slot)
		if claim.Slot == "" || claim.FactRef == "" {
			return ErrAdvisorFinalInvalid
		}
		if strings.IndexFunc(claim.Slot, func(r rune) bool { return r < 0x20 || r == '\u007f' }) >= 0 || len([]rune(claim.Slot)) > 100 {
			return ErrAdvisorFinalInvalid
		}
		if _, ok := seen[claim.Slot]; ok {
			return ErrAdvisorFinalInvalid
		}
		seen[claim.Slot] = struct{}{}
		fact, err := index.Resolve(claim.FactRef)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrAdvisorFinalInvalid, claim.FactRef)
		}
		switch fact.Kind {
		case "money", "count", "version", "basis_points", "text":
		default:
			return ErrAdvisorFinalInvalid
		}
	}
	return nil
}

func RenderAdvisorFinal(answer entity.AdvisorFinalAnswer, index *AdvisorEvidenceIndex) (string, error) {
	if err := ValidateAdvisorFinal(answer, index); err != nil {
		return "", err
	}
	if answer.Mode != "finance" {
		if answer.Mode == "clarification" {
			return advisorClarifications[answer.QuestionID], nil
		}
		if answer.SourceID != "" {
			source := index.help[answer.SourceID]
			return source.Title + "\n" + source.Excerpt, nil
		}
		return "Bạn có thể xem số liệu tại Tổng quan, tìm giao dịch tại Giao dịch, hoặc mở Tài khoản để quản lý ví, danh mục, ngân sách và hũ chi tiêu. Trợ lý có thể chuẩn bị đề xuất; bạn kiểm tra và xác nhận trên thẻ trước khi lưu.", nil
	}
	if answer.TemplateID == "proposal" {
		return "Đã chuẩn bị đề xuất. Bạn kiểm tra và xác nhận trên thẻ để lưu.", nil
	}
	parts := make([]string, 0, len(answer.Claims))
	for _, claim := range answer.Claims {
		fact, _ := index.Resolve(claim.FactRef)
		value := ""
		if fact.Integer != nil {
			value = strconv.FormatInt(*fact.Integer, 10)
			if fact.Currency != nil && strings.TrimSpace(*fact.Currency) != "" {
				value += " " + strings.TrimSpace(*fact.Currency)
			}
		} else if fact.Text != nil {
			value = *fact.Text
		}
		parts = append(parts, advisorFactLabel(fact)+": "+value)
	}
	return strings.Join(parts, "\n"), nil
}

// Labels are derived from trusted fact identifiers; model text cannot relabel
// an expense as income or smuggle unsupported claims into a rendered slot.
func advisorFactLabel(fact entity.Fact) string {
	fields := strings.Split(fact.ID, ".")
	key := fields[len(fields)-1]
	labels := map[string]string{
		"expense": "Chi tiêu", "income": "Thu nhập", "net": "Chênh lệch",
		"delta": "Thay đổi", "current_expense": "Chi tiêu kỳ hiện tại", "previous_expense": "Chi tiêu kỳ trước",
		"change_bps": "Thay đổi (điểm cơ bản)", "total_count": "Số giao dịch", "count": "Số lượng",
		"total_amount": "Tổng tiền", "amount": "Số tiền", "balance": "Số dư",
		"available_credit": "Hạn mức còn lại", "target_remaining": "Còn thiếu so với mục tiêu",
		"remaining": "Còn lại", "limit": "Hạn mức", "spent": "Đã chi", "allocated": "Đã phân bổ",
		"total_spent": "Tổng đã chi", "total_allocated": "Tổng đã phân bổ", "version": "Phiên bản",
	}
	if label, ok := labels[key]; ok {
		return label
	}
	return "Giá trị"
}
