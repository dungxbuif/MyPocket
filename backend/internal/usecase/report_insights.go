package usecase

import "github.com/mypocket/backend/internal/entity"

func BuildReportInsights(report entity.MonthlyReport) []entity.ReportInsight {
	return entity.BuildReportInsights(report)
}
