package agent

import "testing"

func TestParseModelResultRejectsUnsafeShapes(t *testing.T) {
	bad := []string{
		`{"transaction":{"type":"expense","amount_vnd":1,"source_wallet_id":"w","category_id":"c","occurred_at":"bad","note":""}}`,
		`{"transaction":{"type":"transfer","amount_vnd":1,"source_wallet_id":"w","destination_wallet_id":"w","occurred_at":"2026-09-11T00:00:00Z","note":""}}`,
		`{"transaction":{"type":"expense","amount_vnd":1,"source_wallet_id":"w","category_id":"c","occurred_at":"2026-09-11T00:00:00Z","note":"","surprise":true}}`,
	}
	for _, raw := range bad {
		if _, err := ParseModelResult(raw, KindTransactionDraft); err == nil {
			t.Fatalf("expected rejection for %s", raw)
		}
	}
}

func TestParseModelResultAcceptsTypedDraft(t *testing.T) {
	r, err := ParseModelResult(`{"transaction":{"type":"expense","amount_vnd":120000,"source_wallet_id":"w","category_id":"c","occurred_at":"2026-09-11T00:00:00Z","note":"Lunch"}}`, KindTransactionDraft)
	if err != nil || r.Transaction == nil || r.Transaction.AmountVND != 120000 {
		t.Fatalf("unexpected result %#v, %v", r, err)
	}
}
