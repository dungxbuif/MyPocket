package usecase

import (
	"strings"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestAdvisorFactsHaveUniqueRunCallReferences(t *testing.T) {
	index := NewAdvisorEvidenceIndex()
	result := entity.FinanceResult{QueryKey: "q1", Status: "ok", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: int64Ptr(3000000), SourceRef: "q1"}}}
	if err := index.Add("call-1", result); err != nil {
		t.Fatal(err)
	}
	if err := index.Add("call-2", result); err != nil {
		t.Fatal(err)
	}
	if _, err := index.Resolve("call-1.q1.expense"); err != nil {
		t.Fatal(err)
	}
	if _, err := index.Resolve("call-2.q1.expense"); err != nil {
		t.Fatal(err)
	}
}

func TestAdvisorGroundingRejectsForgedStaleAndWrongScopeFacts(t *testing.T) {
	index := NewAdvisorEvidenceIndex()
	if err := index.Add("call-1", entity.FinanceResult{QueryKey: "q1", Status: "ok", Facts: []entity.Fact{{ID: "q1.expense", Kind: "money", Integer: int64Ptr(3000000)}}}); err != nil {
		t.Fatal(err)
	}
	for _, ref := range []string{"call-1.q1.fake", "call-2.q1.expense", "q1.expense"} {
		answer := entity.AdvisorFinalAnswer{Mode: "finance", TemplateID: "summary", Claims: []entity.AdvisorClaim{{Slot: "expense", FactRef: ref}}}
		if err := ValidateAdvisorFinal(answer, index); err == nil {
			t.Fatalf("forged evidence %q was accepted", ref)
		}
	}
}

func TestAdvisorGroundingRendersSignedAmountsAndZeroBaseline(t *testing.T) {
	index := NewAdvisorEvidenceIndex()
	zero := int64(0)
	if err := index.Add("call-1", entity.FinanceResult{QueryKey: "q1", Status: "ok", Facts: []entity.Fact{{ID: "q1.delta", Kind: "money", Integer: &zero, Currency: stringPtr(entity.WalletCurrencyVND)}}}); err != nil {
		t.Fatal(err)
	}
	answer := entity.AdvisorFinalAnswer{Mode: "finance", TemplateID: "summary", Claims: []entity.AdvisorClaim{{Slot: "delta", FactRef: "call-1.q1.delta"}}}
	rendered, err := RenderAdvisorFinal(answer, index)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "0") || strings.Contains(rendered, "<nil>") {
		t.Fatalf("unexpected rendered answer: %q", rendered)
	}
}

func TestAdvisorLegacyFinancialTextUsesSafeFallback(t *testing.T) {
	index := NewAdvisorEvidenceIndex()
	answer := entity.AdvisorFinalAnswer{Mode: "finance", TemplateID: "summary", Explanation: "Bạn đã chi ba triệu đồng"}
	if err := ValidateAdvisorFinal(answer, index); err == nil {
		t.Fatal("financial answer without typed claim must be rejected")
	}
}

func TestAdvisorGroundingUsesTemplateAllowlistAndFactKinds(t *testing.T) {
	index := NewAdvisorEvidenceIndex()
	if err := index.Add("call-1", entity.FinanceResult{QueryKey: "q1", Status: "ok", Facts: []entity.Fact{{ID: "q1.raw", Kind: "arbitrary_json", Text: stringPtr("unsafe")}}}); err != nil {
		t.Fatal(err)
	}
	for _, answer := range []entity.AdvisorFinalAnswer{
		{Mode: "finance", TemplateID: "model_invented", Claims: []entity.AdvisorClaim{{Slot: "x", FactRef: "call-1.q1.raw"}}},
		{Mode: "finance", TemplateID: "summary", Claims: []entity.AdvisorClaim{{Slot: "x", FactRef: "call-1.q1.raw"}}},
	} {
		if err := ValidateAdvisorFinal(answer, index); err == nil {
			t.Fatalf("unsafe final accepted: %+v", answer)
		}
	}
}

func int64Ptr(v int64) *int64    { return &v }
func stringPtr(v string) *string { return &v }
