package usecase

import (
	"github.com/mypocket/backend/internal/entity"
	"testing"
)

func TestAIEntryValidationPreventsUnsupportedOrIncompleteLedgerWrites(t *testing.T) {
	wallet := entity.Wallet{ID: "w", OwnerID: "owner", Type: "basic"}
	valid := entity.AIEntryDraft{Type: "expense", Amount: 35000, WalletID: "w", OccurredAt: "2026-09-21T00:30:00+07:00", IncludedInReports: true}
	if err := ValidateAIEntryDraft(valid, wallet, nil); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*entity.AIEntryDraft){
		func(d *entity.AIEntryDraft) { d.Type = "transfer" }, func(d *entity.AIEntryDraft) { d.Amount = 0 },
		func(d *entity.AIEntryDraft) { d.Amount = 9007199254740992 }, func(d *entity.AIEntryDraft) { d.OccurredAt = "" },
		func(d *entity.AIEntryDraft) { d.OccurredAt = "2026-09-21" }, func(d *entity.AIEntryDraft) { d.WalletID = "other" },
	} {
		d := valid
		change(&d)
		if ValidateAIEntryDraft(d, wallet, nil) == nil {
			t.Fatalf("invalid draft accepted: %+v", d)
		}
	}
	wallet.Type = "credit"
	if ValidateAIEntryDraft(valid, wallet, nil) == nil {
		t.Fatal("credit treated as ordinary expense")
	}
}

func TestAIEntryValidationUsesCategoryWalletAndSavingsRules(t *testing.T) {
	id := "cat"
	d := entity.AIEntryDraft{Type: "income", Amount: 10000, WalletID: "goal", CategoryID: &id, OccurredAt: "2026-09-20T17:30:00Z"}
	w := entity.Wallet{ID: "goal", Type: "goal"}
	key := "income_interest"
	c := entity.Category{ID: id, Kind: "income", SystemKey: &key, WalletIDs: []string{"goal"}}
	if err := ValidateAIEntryDraft(d, w, &c); err != nil {
		t.Fatal(err)
	}
	c.WalletIDs = []string{"different"}
	if ValidateAIEntryDraft(d, w, &c) == nil {
		t.Fatal("wallet scope bypass")
	}
	c.WalletIDs = nil
	c.SystemKey = nil
	if ValidateAIEntryDraft(d, w, &c) == nil {
		t.Fatal("savings catalog bypass")
	}
	c.SystemKey = &key
	c.Kind = "expense"
	if ValidateAIEntryDraft(d, w, &c) == nil {
		t.Fatal("category kind bypass")
	}
	if ValidateAIEntryDraft(d, w, nil) == nil {
		t.Fatal("missing category accepted")
	}
}
