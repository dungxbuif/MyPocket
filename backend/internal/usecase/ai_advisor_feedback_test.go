package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestAdvisorProductHelpFindsJarConfirmationInstructions(t *testing.T) {
	tool, _ := NewAdvisorToolRegistry(nil).Lookup("get_product_help")
	result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"query":"hũ chi tiêu","max_results":1}`))
	if err != nil {
		t.Fatal(err)
	}
	var view productHelpView
	if err := json.Unmarshal(result.View, &view); err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" || len(view.Sources) != 1 || !strings.Contains(view.Sources[0].ID, ".jars.") {
		t.Fatalf("supported jar feature has no relevant help source: %+v", view)
	}
	if !strings.Contains(view.Sources[0].Excerpt, "xác nhận") {
		t.Fatalf("jar help must disclose explicit confirmation: %+v", view.Sources[0])
	}
}

func TestAdvisorProductHelpCoversNativeSecondaryDestinations(t *testing.T) {
	tool, ok := NewAdvisorToolRegistry(nil).Lookup("get_product_help")
	if !ok {
		t.Fatal("product help tool missing")
	}
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{name: "reports", query: "báo cáo xu hướng chi tiêu", expected: ".reports."},
		{name: "categories", query: "danh mục cha con", expected: ".categories."},
		{name: "months", query: "tổng hợp theo tháng", expected: ".months."},
		{name: "recurring", query: "giao dịch định kỳ", expected: ".recurring."},
		{name: "travel", query: "Travel Mode chuyến đi", expected: ".travel."},
		{name: "portfolio", query: "danh mục đầu tư lãi lỗ", expected: ".portfolio."},
		{name: "credit", query: "sổ tín dụng hạn mức", expected: ".credit."},
		{name: "notifications", query: "thông báo đã đọc", expected: ".notifications."},
		{name: "ai entry", query: "nhập giao dịch bằng AI ảnh", expected: ".ai_entry."},
		{name: "feedback", query: "phản hồi changelog", expected: ".feedback."},
		{name: "account", query: "API key Face ID", expected: ".account."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tool.Execute(context.Background(), Principal{OwnerID: "owner-1"}, json.RawMessage(`{"query":"`+tc.query+`","max_results":1}`))
			if err != nil {
				t.Fatal(err)
			}
			var view productHelpView
			if err := json.Unmarshal(result.View, &view); err != nil {
				t.Fatal(err)
			}
			if result.Status != "ok" || len(view.Sources) != 1 || !strings.Contains(view.Sources[0].ID, tc.expected) {
				t.Fatalf("query %q returned the wrong handbook source: status=%s view=%+v", tc.query, result.Status, view)
			}
		})
	}
}
