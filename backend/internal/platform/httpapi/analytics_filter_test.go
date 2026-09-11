package httpapi

import (
	"net/http/httptest"
	"testing"
)

func TestParseAnalyticsFilterRejectsReversedDateRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/reports/cash-flow?from=2026-09-11&to=2026-09-10", nil)
	if _, err := parseAnalyticsFilter(req); err == nil {
		t.Fatal("expected reversed analytics date range to be rejected")
	}
}
