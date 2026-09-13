package entity

import "testing"

func TestSystemCategorySeedsContainStableRootsAndChildren(t *testing.T) {
	seeds := SystemCategorySeeds()
	if len(seeds) < 20 {
		t.Fatalf("expected full system catalog, got %d", len(seeds))
	}
	seen := make(map[string]bool, len(seeds))
	for _, seed := range seeds {
		if seed.SystemKey == "" || seed.Name == "" || seed.Kind == "" {
			t.Fatalf("invalid seed: %+v", seed)
		}
		if seen[seed.SystemKey] {
			t.Fatalf("duplicate system key %q", seed.SystemKey)
		}
		seen[seed.SystemKey] = true
	}
	for _, key := range []string{"expense_food", "income_root", "debt_root", "income_salary", "debt_loan"} {
		if !seen[key] {
			t.Fatalf("missing required system key %q", key)
		}
	}
}

func TestWalletConstantsKeepSupportedContract(t *testing.T) {
	if WalletCurrencyVND != "VND" {
		t.Fatalf("currency must be VND")
	}
	if WalletTypeBasic == WalletTypeGoal || WalletTypeGoal == WalletTypeCredit {
		t.Fatal("wallet types must be distinct")
	}
}
