package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mypocket/backend/internal/entity"
)

// Proposal resource validation is deliberately run-local. A model cannot use
// an ID or version merely because it guessed a valid UUID; the resource must
// have been returned by one of this run's owner-scoped read tools first.
var (
	ErrAdvisorSnapshotRequired    = errors.New("snapshot_required")
	ErrAdvisorSnapshotMismatch    = errors.New("snapshot_mismatch")
	ErrAdvisorSnapshotNotEditable = errors.New("snapshot_not_editable")
)

type advisorResourceSnapshot struct {
	Kind     string
	ID       string
	Version  int64
	Editable bool
	Active   bool
}

// AdvisorResourceSnapshots is a bounded in-memory index for one advisor run.
// It stores identifiers and optimistic versions only, never money, notes or
// raw view payloads.
type AdvisorResourceSnapshots struct {
	resources map[string]advisorResourceSnapshot
}

func NewAdvisorResourceSnapshots() *AdvisorResourceSnapshots {
	return &AdvisorResourceSnapshots{resources: make(map[string]advisorResourceSnapshot)}
}

func (s *AdvisorResourceSnapshots) ensure() {
	if s.resources == nil {
		s.resources = make(map[string]advisorResourceSnapshot)
	}
}

func snapshotKey(kind, id string) string {
	return strings.TrimSpace(kind) + ":" + strings.TrimSpace(id)
}

func (s *AdvisorResourceSnapshots) put(kind, id string, version int64, editable bool) {
	if s == nil || strings.TrimSpace(kind) == "" || strings.TrimSpace(id) == "" {
		return
	}
	s.ensure()
	s.resources[snapshotKey(kind, id)] = advisorResourceSnapshot{Kind: kind, ID: id, Version: version, Editable: editable, Active: true}
}

// Record indexes only typed resource fields emitted by finance tools. Invalid
// or not-found results intentionally contribute nothing.
func (s *AdvisorResourceSnapshots) Record(result entity.FinanceResult) {
	if s == nil || result.Status != "ok" || len(result.View) == 0 {
		return
	}
	var envelope any
	if json.Unmarshal(result.View, &envelope) != nil {
		return
	}
	var values []map[string]any
	switch result.ViewKind {
	case "transaction_detail", "category_detail":
		if item, ok := envelope.(map[string]any); ok {
			values = append(values, item)
		}
	case "transaction_search", "category_catalog", "wallet_balances", "goal_progress":
		if view, ok := envelope.(map[string]any); ok {
			if items, ok := view["items"].([]any); ok {
				for _, raw := range items {
					if item, ok := raw.(map[string]any); ok {
						values = append(values, item)
					}
				}
			}
		}
	default:
		return
	}
	for _, item := range values {
		id, _ := item["id"].(string)
		if id == "" {
			continue
		}
		version := numberValue(item["version"])
		switch result.ViewKind {
		case "category_detail", "category_catalog":
			isSystem, _ := item["is_system"].(bool)
			editable, hasEditable := item["editable"].(bool)
			if !hasEditable {
				editable = !isSystem
			}
			s.put("category", id, version, editable)
			if active, ok := item["is_active"].(bool); ok && !active {
				key := snapshotKey("category", id)
				entry := s.resources[key]
				entry.Active = false
				s.resources[key] = entry
			}
		case "transaction_detail", "transaction_search":
			s.put("transaction", id, version, false)
		default:
			s.put("wallet", id, version, true)
		}
	}
}

// ValidateProposal verifies resource references in a proposal result against
// the read snapshots from this run. It is intentionally separate from the
// canonical confirmation handler, which performs the final DB CAS.
func (s *AdvisorResourceSnapshots) ValidateProposal(result entity.FinanceResult) error {
	if s == nil || len(result.View) == 0 {
		return fmt.Errorf("%w: no resource snapshot", ErrAdvisorSnapshotRequired)
	}
	var proposal map[string]any
	if json.Unmarshal(result.View, &proposal) != nil {
		return fmt.Errorf("%w: malformed proposal", ErrAdvisorSnapshotRequired)
	}
	action, _ := proposal["action"].(string)
	require := func(kind, id string, expected *int64, editable bool) error {
		id = strings.TrimSpace(id)
		item, ok := s.resources[snapshotKey(kind, id)]
		if !ok {
			return fmt.Errorf("%w: %s %s", ErrAdvisorSnapshotRequired, kind, id)
		}
		if expected != nil && item.Version != *expected {
			return fmt.Errorf("%w: %s %s", ErrAdvisorSnapshotMismatch, kind, id)
		}
		if editable && !item.Editable {
			return fmt.Errorf("%w: %s %s", ErrAdvisorSnapshotNotEditable, kind, id)
		}
		return nil
	}
	version, _ := optionalNumber(proposal["expected_version"])
	ref := func(kind, field string, editable bool) error {
		id, _ := proposal[field].(string)
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("%w: missing %s", ErrAdvisorSnapshotRequired, field)
		}
		return require(kind, id, version, editable)
	}
	optionalRef := func(kind, field string) error {
		id, _ := proposal[field].(string)
		if strings.TrimSpace(id) == "" {
			return nil
		}
		if err := require(kind, id, nil, false); err != nil {
			return err
		}
		if !s.resources[snapshotKey(kind, id)].Active {
			return ErrAdvisorSnapshotNotEditable
		}
		return nil
	}
	switch action {
	case "update_transaction":
		if err := ref("transaction", "transaction_id", false); err != nil {
			return err
		}
		if err := optionalRef("wallet", "wallet_id"); err != nil {
			return err
		}
		return optionalRef("category", "category_id")
	case "delete_transaction":
		return ref("transaction", "transaction_id", false)
	case "recategorize_transaction":
		if err := ref("transaction", "transaction_id", false); err != nil {
			return err
		}
		return optionalRef("category", "category_id")
	case "update_category":
		if err := ref("category", "category_id", true); err != nil {
			return err
		}
		if err := optionalRef("category", "parent_id"); err != nil {
			return err
		}
		return validateArrayRefs(s, proposal["wallet_ids"], "wallet")
	case "create_transaction":
		if err := optionalRef("category", "category_id"); err != nil {
			return err
		}
		return ref("wallet", "wallet_id", false)
	case "transfer_wallet":
		if err := ref("wallet", "source_wallet_id", false); err != nil {
			return err
		}
		return ref("wallet", "destination_wallet_id", false)
	case "create_category":
		if err := optionalRef("category", "parent_id"); err != nil {
			return err
		}
		return validateArrayRefs(s, proposal["wallet_ids"], "wallet")
	case "create_budget", "update_budget":
		if err := optionalRef("category", "category_id"); err != nil {
			return err
		}
		return optionalRef("wallet", "wallet_id")
	default:
		// Read-only or wallet-creation proposals carry no existing resource
		// references and need no snapshot check here.
		return nil
	}
}

func validateArrayRefs(s *AdvisorResourceSnapshots, raw any, kind string) error {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	for _, item := range items {
		id, _ := item.(string)
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("%w: empty %s reference", ErrAdvisorSnapshotRequired, kind)
		}
		if _, exists := s.resources[snapshotKey(kind, id)]; !exists {
			return fmt.Errorf("%w: %s %s", ErrAdvisorSnapshotRequired, kind, id)
		}
	}
	return nil
}

func numberValue(value any) int64 {
	if parsed, ok := optionalNumber(value); ok {
		return *parsed
	}
	return 0
}

func optionalNumber(value any) (*int64, bool) {
	switch parsed := value.(type) {
	case float64:
		value := int64(parsed)
		if parsed != float64(value) {
			return nil, false
		}
		return &value, true
	case json.Number:
		value, err := parsed.Int64()
		if err != nil {
			return nil, false
		}
		return &value, true
	case int64:
		return &parsed, true
	case int:
		value := int64(parsed)
		return &value, true
	default:
		return nil, false
	}
}
