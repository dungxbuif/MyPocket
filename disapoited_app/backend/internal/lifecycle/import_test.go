package lifecycle

import (
	"strings"
	"testing"
)

func TestImportParserRequiresOwnedReferencesAndAllValidRows(t *testing.T) {
	csv := "occurred_at,type,amount_vnd,source_wallet_id,destination_wallet_id,category_id,note,excluded_from_reports\n" +
		"2026-09-11T10:00:00+07:00,expense,12000,wallet-a,,category-a,Cafe,false\n"
	preview := (ImportParser{}).Parse("job", 2, strings.NewReader(csv), map[string]bool{"wallet-a": true}, map[string]bool{"category-a": true})
	if !preview.Confirmable || len(preview.ValidRows) != 1 || len(preview.Errors) != 0 {
		t.Fatalf("unexpected preview: %+v", preview)
	}
	foreign := (ImportParser{}).Parse("job", 2, strings.NewReader(csv), map[string]bool{}, map[string]bool{"category-a": true})
	if foreign.Confirmable || len(foreign.Errors) != 1 || foreign.Errors[0].Row != 2 {
		t.Fatalf("foreign wallet was accepted: %+v", foreign)
	}
}

func TestImportParserRejectsFractionalMoneyAndPartialConfirmation(t *testing.T) {
	csv := "occurred_at,type,amount_vnd,source_wallet_id,destination_wallet_id,category_id,note,excluded_from_reports\n" +
		"2026-09-11T10:00:00+07:00,expense,12.5,wallet-a,,category-a,Cafe,false\n"
	preview := (ImportParser{}).Parse("job", 2, strings.NewReader(csv), map[string]bool{"wallet-a": true}, map[string]bool{"category-a": true})
	if preview.Confirmable || len(preview.ValidRows) != 0 || preview.Errors[0].Code != "INVALID_AMOUNT" {
		t.Fatalf("fractional money was accepted: %+v", preview)
	}
}
