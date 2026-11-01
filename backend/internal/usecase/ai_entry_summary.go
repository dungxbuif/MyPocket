package usecase

import (
	"fmt"
	"strings"

	"github.com/mypocket/backend/internal/entity"
)

// Model prose is not evidence of row counts, wallet availability or approval.
// Render only facts derived after normalization and validation, without
// claiming that OCR found every source transaction.
func aiEntrySummary(drafts []entity.AIExtractDraft, invalid int) string {
	if len(drafts) == 0 {
		return "Chưa tạo được đề xuất giao dịch. Hãy kiểm tra lại nội dung chứng từ. Chưa lưu giao dịch nào."
	}
	counts := map[string]int{}
	missingWallets := 0
	for _, draft := range drafts {
		counts[draft.Type]++
		if draft.WalletID == "" {
			missingWallets++
		}
	}
	reply := fmt.Sprintf("Đã tạo %d đề xuất để xem lại: %d khoản thu, %d khoản chi, %d chuyển khoản, %d chưa rõ loại. Chưa lưu giao dịch nào.", len(drafts), counts["income"], counts["expense"], counts["transfer"], counts["unknown"])
	if invalid > 0 {
		reply += fmt.Sprintf(" %d đề xuất chưa đủ điều kiện duyệt; hãy sửa các trường được đánh dấu.", invalid)
	}
	if missingWallets > 0 {
		reply += fmt.Sprintf(" %d đề xuất cần chọn ví.", missingWallets)
	}
	return reply + " Hãy kiểm tra từng đề xuất và đối chiếu chứng từ trước khi duyệt."
}

// Only remove known standalone missing-wallet claims after resolving a wallet.
// Keep semantic/financial uncertainty, including compound model questions.
func normalizeAIEntryQuestions(questions []string, hasWallet bool) []string {
	result := make([]string, 0, len(questions))
	seen := make(map[string]bool, len(questions))
	for _, question := range questions {
		question = strings.TrimSpace(question)
		if question == "" || seen[question] {
			continue
		}
		if hasWallet {
			switch question {
			case "Chưa có ví: hãy tạo/chọn ví để gán giao dịch này.", "Chưa có ví: hãy tạo/chọn ví để ghi giao dịch.":
				continue
			}
		}
		seen[question] = true
		result = append(result, question)
	}
	return result
}
