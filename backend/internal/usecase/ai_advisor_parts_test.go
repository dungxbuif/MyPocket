package usecase

import (
	"encoding/json"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestBuildAdvisorPartsKeepsNumbersOutOfFreeTextContract(t *testing.T) {
	answer := AdvisorAnswer{
		Text: "Bạn đã chi 150 ₫.",
		Results: []entity.FinanceResult{{
			Status:   "ok",
			ViewKind: "finance_summary",
			View:     json.RawMessage(`{"income":1000,"expense":150,"net":850,"count":3}`),
		}},
	}
	parts := BuildAdvisorParts(answer)
	if len(parts) != 2 || parts[0].Type != "text" || parts[1].Type != "metric_group" {
		t.Fatalf("unexpected parts: %+v", parts)
	}
	if parts[1].Data == nil {
		t.Fatal("metric part must carry server view data")
	}
}

func TestBuildAdvisorPartsRejectsUnknownAndMalformedViews(t *testing.T) {
	parts := BuildAdvisorParts(AdvisorAnswer{Text: "ok", Results: []entity.FinanceResult{{ViewKind: "future_card", View: json.RawMessage(`{"amount":999}`)}, {ViewKind: "finance_summary", View: json.RawMessage(`not-json`)}}})
	if len(parts) != 1 || parts[0].Type != "text" {
		t.Fatalf("unknown or malformed views must not reach the renderer: %+v", parts)
	}
}
