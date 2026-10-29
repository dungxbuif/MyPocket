package usecase

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestAdvisorResourceSnapshotsRequireReadVersionAndEditableCategory(t *testing.T) {
	snapshots := NewAdvisorResourceSnapshots()
	snapshots.Record(entity.FinanceResult{Status: "ok", ViewKind: "transaction_detail", View: json.RawMessage(`{"id":"tx-1","version":7}`)})
	snapshots.Record(entity.FinanceResult{Status: "ok", ViewKind: "category_catalog", View: json.RawMessage(`{"items":[{"id":"cat-system","version":3,"is_system":true,"editable":false},{"id":"cat-personal","version":6,"is_system":false,"editable":true}]}`)})

	valid := entity.FinanceResult{View: json.RawMessage(`{"action":"update_transaction","transaction_id":"tx-1","expected_version":7}`)}
	if err := snapshots.ValidateProposal(valid); err != nil {
		t.Fatalf("matching version rejected: %v", err)
	}
	stale := valid
	stale.View = json.RawMessage(`{"action":"update_transaction","transaction_id":"tx-1","expected_version":6}`)
	if err := snapshots.ValidateProposal(stale); !errors.Is(err, ErrAdvisorSnapshotMismatch) {
		t.Fatalf("stale version error=%v", err)
	}
	unknown := valid
	unknown.View = json.RawMessage(`{"action":"update_transaction","transaction_id":"tx-unknown","expected_version":7}`)
	if err := snapshots.ValidateProposal(unknown); !errors.Is(err, ErrAdvisorSnapshotRequired) {
		t.Fatalf("unknown resource error=%v", err)
	}
	system := entity.FinanceResult{View: json.RawMessage(`{"action":"update_category","category_id":"cat-system","expected_version":3}`)}
	if err := snapshots.ValidateProposal(system); !errors.Is(err, ErrAdvisorSnapshotNotEditable) {
		t.Fatalf("system category error=%v", err)
	}
}

func TestAdvisorResourceSnapshotsValidateCreateReferences(t *testing.T) {
	snapshots := NewAdvisorResourceSnapshots()
	snapshots.Record(entity.FinanceResult{Status: "ok", ViewKind: "wallet_balances", View: json.RawMessage(`{"items":[{"id":"wallet-1","name":"Cash"}]}`)})
	snapshots.Record(entity.FinanceResult{Status: "ok", ViewKind: "category_detail", View: json.RawMessage(`{"id":"cat-1","version":1,"is_system":false,"editable":true}`)})
	proposal := entity.FinanceResult{View: json.RawMessage(`{"action":"create_transaction","wallet_id":"wallet-1","category_id":"cat-1"}`)}
	if err := snapshots.ValidateProposal(proposal); err != nil {
		t.Fatalf("owned references rejected: %v", err)
	}
	proposal.View = json.RawMessage(`{"action":"create_transaction","wallet_id":"wallet-foreign","category_id":"cat-1"}`)
	if err := snapshots.ValidateProposal(proposal); !errors.Is(err, ErrAdvisorSnapshotRequired) {
		t.Fatalf("foreign wallet reference error=%v", err)
	}
}
