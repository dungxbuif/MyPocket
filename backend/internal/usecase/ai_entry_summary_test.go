package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestAIEntrySummaryUsesNormalizedProposals(t *testing.T) {
	r := &entryRepoStub{}
	drafts := []entity.AIExtractDraft{
		{AIEntryDraft: entity.AIEntryDraft{Type: "expense", Amount: 10, OccurredAt: "2026-11-01T00:00:00+07:00"}, Questions: []string{"Chưa có ví: hãy tạo/chọn ví để gán giao dịch này.", "Xác nhận chiều chuyển tiền.", "Xác nhận chiều chuyển tiền."}},
		{AIEntryDraft: entity.AIEntryDraft{Type: "transfer", Amount: 20}},
		{AIEntryDraft: entity.AIEntryDraft{Type: "invalid", Amount: 30}},
	}
	s := &AIEntryService{Entries: r, Wallets: entryWallets{}, Categories: entryCategories{}, Extractor: entryExtractor{configured: true, out: entity.AIExtractOutput{Reply: "18 giao dịch. Chưa có ví.", Drafts: drafts}}}
	process, err := s.Send(context.Background(), "owner", "session", AIEntryMessageInput{RequestID: "00000000-0000-4000-8000-000000000001", Text: "statement", Timezone: "Asia/Ho_Chi_Minh"})
	if err != nil {
		t.Fatal(err)
	}
	if process.Reply != r.output.Reply || !strings.Contains(process.Reply, "3 đề xuất") || !strings.Contains(process.Reply, "1 khoản chi") || !strings.Contains(process.Reply, "1 chuyển khoản") || !strings.Contains(process.Reply, "1 chưa rõ loại") || strings.Contains(process.Reply, "18") || strings.Contains(process.Reply, "Chưa có ví") {
		t.Fatalf("summary is not grounded in normalized drafts: %q", process.Reply)
	}
	questions := strings.Join(r.output.Drafts[0].Questions, "\n")
	if strings.Contains(questions, "Chưa có ví") || strings.Count(questions, "Xác nhận chiều chuyển tiền.") != 1 {
		t.Fatalf("contradicted/duplicate questions survived: %q", questions)
	}
	if len(r.output.Drafts) != 3 || r.output.Drafts[0].WalletID != "w" {
		t.Fatal("proposals or wallet fallback changed")
	}
}

func TestAIEntryQuestionCleanupPreservesUncertainty(t *testing.T) {
	questions := []string{"Chưa có ví: hãy tạo/chọn ví để ghi giao dịch.", "Số tiền ngoại tệ cần xác nhận.", "Số tiền ngoại tệ cần xác nhận.", ""}
	withoutWallet := normalizeAIEntryQuestions(questions, false)
	withWallet := normalizeAIEntryQuestions(questions, true)
	if len(withoutWallet) != 2 || len(withWallet) != 1 || withWallet[0] != questions[1] {
		t.Fatalf("lost safety evidence: without=%v with=%v", withoutWallet, withWallet)
	}
}

func TestAIEntrySummaryReportsValidationAndMissingWalletCounts(t *testing.T) {
	drafts := []entity.AIExtractDraft{
		{AIEntryDraft: entity.AIEntryDraft{Type: "income", WalletID: "w"}},
		{AIEntryDraft: entity.AIEntryDraft{Type: "expense"}},
	}
	reply := aiEntrySummary(drafts, 1)
	for _, want := range []string{"2 đề xuất", "1 khoản thu", "1 khoản chi", "1 đề xuất chưa đủ điều kiện", "1 đề xuất cần chọn ví", "Chưa lưu giao dịch nào"} {
		if !strings.Contains(reply, want) {
			t.Fatalf("missing %q in %q", want, reply)
		}
	}
	if empty := aiEntrySummary(nil, 0); !strings.Contains(empty, "Chưa tạo được đề xuất") || strings.Contains(empty, "Chưa có ví") {
		t.Fatalf("ungrounded empty summary: %q", empty)
	}
}
