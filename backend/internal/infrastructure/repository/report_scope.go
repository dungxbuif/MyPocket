package repository

import (
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/gorm"
)

// applyCanonicalReportEligibility is the one accounting predicate shared by
// the monthly read model and the existing finance/report readers. Adjustments,
// transfers, excluded rows, and legacy system transfer categories are never
// report income or expense.
func applyCanonicalReportEligibility(db *gorm.DB, alias string) *gorm.DB {
	return db.Where(alias+".included_in_reports = TRUE").
		Where(alias+".type IN (?, ?)", entity.TransactionTypeIncome, entity.TransactionTypeExpense).
		Where(alias+".transfer_id IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM categories report_transfer_categories WHERE report_transfer_categories.id = "+alias+".category_id AND report_transfer_categories.system_key IN (?, ?))", "income_transfer_in", "expense_transfer_out")
}
