package sync

import (
	"encoding/json"
	"time"

	"mypocket/internal/finance"
	"mypocket/internal/portfolio"
)

type EntityType string
type Operation string
type ResultState string

const (
	ServerEpoch = "atomic-sync-v1"

	EntityWallet      EntityType = "wallet"
	EntityCategory    EntityType = "category"
	EntityTransaction EntityType = "transaction"
	EntityAsset       EntityType = "asset"

	OperationCreate            Operation = "create"
	OperationUpdate            Operation = "update"
	OperationArchive           Operation = "archive"
	OperationSetDefaultAI      Operation = "set_default_ai"
	OperationSetCategoryActive Operation = "set_category_active"
	OperationAddTrade          Operation = "add_trade"
	OperationUpdateTrade       Operation = "update_trade"
	OperationArchiveTrade      Operation = "archive_trade"
	OperationAddPrice          Operation = "add_price"

	ResultApplied  ResultState = "applied"
	ResultReplayed ResultState = "replayed"
	ResultRejected ResultState = "rejected"
	ResultConflict ResultState = "conflict"
)

type Mutation struct {
	MutationID  string          `json:"mutation_id"`
	DeviceID    string          `json:"device_id"`
	Sequence    int64           `json:"sequence"`
	EntityType  EntityType      `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	Operation   Operation       `json:"operation"`
	BaseVersion int64           `json:"base_version"`
	Payload     json.RawMessage `json:"payload"`
}

type MutationResult struct {
	MutationID string          `json:"mutation_id"`
	EntityType EntityType      `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Operation  Operation       `json:"operation"`
	State      ResultState     `json:"state"`
	Version    int64           `json:"version,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Reason     string          `json:"reason,omitempty"`
	Conflict   *Conflict       `json:"conflict,omitempty"`
}

type Conflict struct {
	EntityType    EntityType      `json:"entity_type"`
	EntityID      string          `json:"entity_id"`
	Operation     Operation       `json:"operation"`
	BaseVersion   int64           `json:"base_version"`
	ServerVersion int64           `json:"server_version"`
	LocalPayload  json.RawMessage `json:"local_payload"`
	ServerPayload json.RawMessage `json:"server_payload"`
}

type Change struct {
	Cursor     int64           `json:"cursor"`
	EntityType EntityType      `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Operation  Operation       `json:"operation"`
	Version    int64           `json:"version"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"created_at"`
}

type ChangesResult struct {
	Changes    []Change `json:"changes"`
	NextCursor int64    `json:"next_cursor"`
}

type Snapshot struct {
	ServerEpoch  string                `json:"server_epoch"`
	Wallets      []finance.Wallet      `json:"wallets"`
	Categories   []finance.Category    `json:"categories"`
	Transactions []finance.Transaction `json:"transactions"`
	Assets       []portfolio.Position  `json:"assets"`
	NextCursor   int64                 `json:"next_cursor"`
}
