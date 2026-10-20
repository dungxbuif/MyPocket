package entity

import "time"

// AdvisorContextState is a bounded, server-derived continuation hint. It is
// intentionally reference-only: amounts, notes and optimistic-lock versions
// are always fetched again from the finance reader.
type AdvisorContextState struct {
	Version         int       `json:"version"`
	Generation      int64     `json:"generation"`
	ThroughSeq      int64     `json:"through_seq"`
	LastQueryKind   string    `json:"last_query_kind,omitempty"`
	DateFrom        string    `json:"date_from,omitempty"`
	DateTo          string    `json:"date_to,omitempty"`
	Timezone        string    `json:"timezone,omitempty"`
	WalletRefs      []string  `json:"wallet_refs,omitempty"`
	CategoryRefs    []string  `json:"category_refs,omitempty"`
	TransactionRefs []string  `json:"transaction_refs,omitempty"`
	SelectedTopic   string    `json:"selected_topic,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
	// These fields are never populated; they make accidental persistence of
	// financial snapshots fail code review/searches loudly.
	LastAmount  int64 `json:"-"`
	LastVersion int64 `json:"-"`
}
