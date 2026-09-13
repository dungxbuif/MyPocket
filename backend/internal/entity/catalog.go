package entity

type SystemCategorySeed struct {
	ID        string
	ParentKey string
	Kind      string
	Name      string
	SystemKey string
}

// SystemCategorySeeds is the stable catalog ported from the reference app.
func SystemCategorySeeds() []SystemCategorySeed {
	return []SystemCategorySeed{
		{ID: "00000000-0000-4000-8000-000000000301", Kind: "expense", Name: "Ăn uống", SystemKey: "expense_food"},
		{ID: "00000000-0000-4000-8000-000000000302", Kind: "expense", Name: "Hoá đơn & Tiện ích", SystemKey: "expense_bills"},
		{ID: "00000000-0000-4000-8000-000000000303", Kind: "expense", Name: "Mua sắm", SystemKey: "expense_shopping"},
		{ID: "00000000-0000-4000-8000-000000000304", Kind: "expense", Name: "Gia đình", SystemKey: "expense_family"},
		{ID: "00000000-0000-4000-8000-000000000305", Kind: "expense", Name: "Di chuyển", SystemKey: "expense_transport"},
		{ID: "00000000-0000-4000-8000-000000000306", Kind: "expense", Name: "Sức khỏe", SystemKey: "expense_health"},
		{ID: "00000000-0000-4000-8000-000000000307", Kind: "expense", Name: "Giáo dục", SystemKey: "expense_education"},
		{ID: "00000000-0000-4000-8000-000000000308", Kind: "expense", Name: "Giải trí", SystemKey: "expense_entertainment"},
		{ID: "00000000-0000-4000-8000-000000000309", Kind: "expense", Name: "Quà tặng & Quyên góp", SystemKey: "expense_gifts"},
		{ID: "00000000-0000-4000-8000-000000000310", Kind: "expense", Name: "Bảo hiểm", SystemKey: "expense_insurance"},
		{ID: "00000000-0000-4000-8000-000000000311", Kind: "expense", Name: "Đầu tư", SystemKey: "expense_investment"},
		{ID: "00000000-0000-4000-8000-000000000312", Kind: "expense", Name: "Du lịch", SystemKey: "expense_travel"},
		{ID: "00000000-0000-4000-8000-000000000313", Kind: "expense", Name: "Tiết kiệm", SystemKey: "expense_savings"},
		{ID: "00000000-0000-4000-8000-000000000314", Kind: "expense", Name: "Kinh doanh", SystemKey: "expense_business"},
		{ID: "00000000-0000-4000-8000-000000000315", Kind: "expense", Name: "Các chi phí chung", SystemKey: "expense_general"},
		{ID: "00000000-0000-4000-8000-000000000316", Kind: "income", Name: "Thu nhập", SystemKey: "income_root"},
		{ID: "00000000-0000-4000-8000-000000000317", Kind: "debt", Name: "Vay/Nợ", SystemKey: "debt_root"},
		{ID: "00000000-0000-4000-8000-000000000321", ParentKey: "expense_food", Kind: "expense", Name: "Ăn vặt", SystemKey: "expense_food_snacks"},
		{ID: "00000000-0000-4000-8000-000000000322", ParentKey: "expense_food", Kind: "expense", Name: "Cà phê", SystemKey: "expense_food_coffee"},
		{ID: "00000000-0000-4000-8000-000000000323", ParentKey: "expense_food", Kind: "expense", Name: "Cơm bữa", SystemKey: "expense_food_meals"},
		{ID: "00000000-0000-4000-8000-000000000324", ParentKey: "expense_food", Kind: "expense", Name: "Nhà hàng", SystemKey: "expense_food_restaurant"},
		{ID: "00000000-0000-4000-8000-000000000366", ParentKey: "income_root", Kind: "income", Name: "Lương", SystemKey: "income_salary"},
		{ID: "00000000-0000-4000-8000-000000000367", ParentKey: "income_root", Kind: "income", Name: "Thu nhập khác", SystemKey: "income_other"},
		{ID: "00000000-0000-4000-8000-000000000372", ParentKey: "income_root", Kind: "income", Name: "Thưởng", SystemKey: "income_bonus"},
		{ID: "00000000-0000-4000-8000-000000000374", ParentKey: "debt_root", Kind: "debt", Name: "Cho vay", SystemKey: "debt_lend"},
		{ID: "00000000-0000-4000-8000-000000000375", ParentKey: "debt_root", Kind: "debt", Name: "Trả nợ", SystemKey: "debt_repay"},
		{ID: "00000000-0000-4000-8000-000000000376", ParentKey: "debt_root", Kind: "debt", Name: "Đi vay", SystemKey: "debt_loan"},
		{ID: "00000000-0000-4000-8000-000000000377", ParentKey: "debt_root", Kind: "debt", Name: "Thu nợ", SystemKey: "debt_collect"},
	}
}
