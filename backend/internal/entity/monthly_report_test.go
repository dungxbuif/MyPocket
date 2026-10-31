package entity

import "testing"

func TestNormalizeMonthlyReportGranularityAcceptsNativeWeeklyAlias(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
	}{
		{input: "", want: "week"},
		{input: "week", want: "week"},
		{input: "weekly", want: "week"},
		{input: "day", want: "day"},
	} {
		if got := NormalizeMonthlyReportGranularity(tc.input); got != tc.want {
			t.Fatalf("NormalizeMonthlyReportGranularity(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeMonthlyReportGranularityRejectsUnknownValue(t *testing.T) {
	if got := NormalizeMonthlyReportGranularity("daily"); got != "" {
		t.Fatalf("unknown granularity normalized to %q, want empty", got)
	}
}
