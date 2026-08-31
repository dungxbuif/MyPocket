package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"mypocket/internal/finance"
)

type FinanceCommands interface {
	ListWallets(ctx context.Context, userID string) ([]finance.Wallet, error)
	CreateWallet(ctx context.Context, userID string, input finance.CreateWalletInput) (finance.Wallet, error)
	UpdateWallet(ctx context.Context, userID string, walletID string, input finance.UpdateWalletInput) (finance.Wallet, error)
	ArchiveWallet(ctx context.Context, userID string, walletID string) error
	SetDefaultAIWallet(ctx context.Context, userID string, walletID string) error
	ListCategories(ctx context.Context, userID string) ([]finance.Category, error)
	CreateCategory(ctx context.Context, userID string, input finance.CreateCategoryInput) (finance.Category, error)
	UpdateCategory(ctx context.Context, userID string, categoryID string, input finance.UpdateCategoryInput) (finance.Category, error)
	ArchiveCategory(ctx context.Context, userID string, categoryID string) error
	SetWalletCategoryActive(ctx context.Context, userID string, walletID string, categoryID string, active bool) error
	ListTransactions(ctx context.Context, userID string, filters finance.TransactionFilters) ([]finance.Transaction, error)
	CreateTransaction(ctx context.Context, userID string, input finance.CreateTransactionInput) (finance.Transaction, error)
	UpdateTransaction(ctx context.Context, userID string, transactionID string, input finance.UpdateTransactionInput) (finance.Transaction, error)
	ArchiveTransaction(ctx context.Context, userID string, transactionID string) error
}

type Store interface {
	LoadMutationResult(ctx context.Context, userID string, mutationID string) (requestHash string, result MutationResult, found bool, err error)
	StoreMutationResult(ctx context.Context, userID string, mutationID string, requestHash string, result MutationResult) error
	AppendChange(ctx context.Context, userID string, entityType EntityType, entityID string, operation Operation, version int64, payload json.RawMessage) (Change, error)
	ListChanges(ctx context.Context, userID string, after int64, limit int) (ChangesResult, error)
	CurrentCursor(ctx context.Context, userID string) (int64, error)
	EntityVersionAndPayload(ctx context.Context, userID string, entityType EntityType, entityID string) (int64, json.RawMessage, bool, error)
}

type Service struct {
	store   Store
	finance FinanceCommands
}

func NewService(store Store, financeCommands FinanceCommands) *Service {
	return &Service{store: store, finance: financeCommands}
}

func (s *Service) ApplyMutations(ctx context.Context, userID string, mutations []Mutation) ([]MutationResult, error) {
	if err := validateBatch(userID, mutations); err != nil {
		return nil, err
	}
	results := make([]MutationResult, 0, len(mutations))
	for _, mutation := range mutations {
		result, err := s.applyOne(ctx, userID, mutation)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) Changes(ctx context.Context, userID string, after int64, limit int) (ChangesResult, error) {
	if strings.TrimSpace(userID) == "" || after < 0 {
		return ChangesResult{}, fmt.Errorf("%w: invalid change cursor", ErrValidation)
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.store.ListChanges(ctx, userID, after, limit)
}

func (s *Service) Resync(ctx context.Context, userID string) (Snapshot, error) {
	if strings.TrimSpace(userID) == "" {
		return Snapshot{}, fmt.Errorf("%w: user is required", ErrValidation)
	}
	wallets, err := s.finance.ListWallets(ctx, userID)
	if err != nil {
		return Snapshot{}, fmt.Errorf("list resync wallets: %w", err)
	}
	categories, err := s.finance.ListCategories(ctx, userID)
	if err != nil {
		return Snapshot{}, fmt.Errorf("list resync categories: %w", err)
	}
	transactions, err := s.finance.ListTransactions(ctx, userID, finance.TransactionFilters{})
	if err != nil {
		return Snapshot{}, fmt.Errorf("list resync transactions: %w", err)
	}
	cursor, err := s.store.CurrentCursor(ctx, userID)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Wallets: wallets, Categories: categories, Transactions: transactions, NextCursor: cursor}, nil
}

func (s *Service) applyOne(ctx context.Context, userID string, mutation Mutation) (MutationResult, error) {
	hash, err := RequestHash(mutation)
	if err != nil {
		return MutationResult{}, err
	}
	storedHash, stored, found, err := s.store.LoadMutationResult(ctx, userID, mutation.MutationID)
	if err != nil {
		return MutationResult{}, err
	}
	if found {
		if storedHash != hash {
			return rejectedResult(mutation, "mutation_id reused with different request"), nil
		}
		if stored.State == ResultApplied {
			stored.State = ResultReplayed
		}
		return stored, nil
	}

	result, err := s.executeMutation(ctx, userID, mutation)
	if err != nil {
		if errors.Is(err, finance.ErrValidation) || errors.Is(err, finance.ErrForbidden) || errors.Is(err, ErrValidation) {
			result = rejectedResult(mutation, safeReason(err))
		} else {
			return MutationResult{}, err
		}
	}
	if err := s.store.StoreMutationResult(ctx, userID, mutation.MutationID, hash, result); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}

func (s *Service) executeMutation(ctx context.Context, userID string, mutation Mutation) (MutationResult, error) {
	if mutation.Operation != OperationCreate {
		conflict, ok, err := s.conflictForStaleBase(ctx, userID, mutation)
		if err != nil {
			return MutationResult{}, err
		}
		if ok {
			return MutationResult{
				MutationID: mutation.MutationID,
				EntityType: mutation.EntityType,
				EntityID:   mutation.EntityID,
				Operation:  mutation.Operation,
				State:      ResultConflict,
				Conflict:   &conflict,
			}, nil
		}
	}

	switch mutation.EntityType {
	case EntityWallet:
		return s.executeWalletMutation(ctx, userID, mutation)
	case EntityCategory:
		return s.executeCategoryMutation(ctx, userID, mutation)
	case EntityTransaction:
		return s.executeTransactionMutation(ctx, userID, mutation)
	default:
		return MutationResult{}, fmt.Errorf("%w: unsupported entity_type", ErrValidation)
	}
}

func (s *Service) executeWalletMutation(ctx context.Context, userID string, mutation Mutation) (MutationResult, error) {
	switch mutation.Operation {
	case OperationCreate:
		var input struct {
			Name           string             `json:"name"`
			Type           finance.WalletType `json:"type"`
			BalanceVND     int64              `json:"balance_vnd"`
			IncludeInTotal *bool              `json:"include_in_total"`
		}
		if err := decodePayload(mutation.Payload, &input); err != nil {
			return MutationResult{}, err
		}
		wallet, err := s.finance.CreateWallet(ctx, userID, finance.CreateWalletInput{
			ID:             mutation.EntityID,
			Name:           input.Name,
			Type:           input.Type,
			BalanceVND:     input.BalanceVND,
			IncludeInTotal: input.IncludeInTotal,
		})
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, wallet.Version, mustJSON(wallet))
	case OperationUpdate:
		var input struct {
			Name           string `json:"name"`
			IncludeInTotal bool   `json:"include_in_total"`
		}
		if err := decodePayload(mutation.Payload, &input); err != nil {
			return MutationResult{}, err
		}
		wallet, err := s.finance.UpdateWallet(ctx, userID, mutation.EntityID, finance.UpdateWalletInput{Name: input.Name, IncludeInTotal: &input.IncludeInTotal})
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, wallet.Version, mustJSON(wallet))
	case OperationArchive:
		if err := s.finance.ArchiveWallet(ctx, userID, mutation.EntityID); err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, mutation.BaseVersion+1, json.RawMessage(`{}`))
	case OperationSetDefaultAI:
		if err := s.finance.SetDefaultAIWallet(ctx, userID, mutation.EntityID); err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, mutation.BaseVersion+1, json.RawMessage(`{}`))
	default:
		return MutationResult{}, fmt.Errorf("%w: unsupported wallet operation", ErrValidation)
	}
}

func (s *Service) executeCategoryMutation(ctx context.Context, userID string, mutation Mutation) (MutationResult, error) {
	switch mutation.Operation {
	case OperationCreate:
		var input struct {
			Kind finance.CategoryKind `json:"kind"`
			Name string               `json:"name"`
		}
		if err := decodePayload(mutation.Payload, &input); err != nil {
			return MutationResult{}, err
		}
		category, err := s.finance.CreateCategory(ctx, userID, finance.CreateCategoryInput{ID: mutation.EntityID, Kind: input.Kind, Name: input.Name})
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, category.Version, mustJSON(category))
	case OperationUpdate:
		var input struct {
			Name string `json:"name"`
		}
		if err := decodePayload(mutation.Payload, &input); err != nil {
			return MutationResult{}, err
		}
		category, err := s.finance.UpdateCategory(ctx, userID, mutation.EntityID, finance.UpdateCategoryInput{Name: input.Name})
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, category.Version, mustJSON(category))
	case OperationArchive:
		if err := s.finance.ArchiveCategory(ctx, userID, mutation.EntityID); err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, mutation.BaseVersion+1, json.RawMessage(`{}`))
	case OperationSetCategoryActive:
		var input struct {
			WalletID string `json:"wallet_id"`
			Active   bool   `json:"active"`
		}
		if err := decodePayload(mutation.Payload, &input); err != nil {
			return MutationResult{}, err
		}
		if err := s.finance.SetWalletCategoryActive(ctx, userID, input.WalletID, mutation.EntityID, input.Active); err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, mutation.BaseVersion+1, mutation.Payload)
	default:
		return MutationResult{}, fmt.Errorf("%w: unsupported category operation", ErrValidation)
	}
}

func (s *Service) executeTransactionMutation(ctx context.Context, userID string, mutation Mutation) (MutationResult, error) {
	switch mutation.Operation {
	case OperationCreate:
		input, err := transactionInputFromPayload(mutation.Payload)
		if err != nil {
			return MutationResult{}, err
		}
		input.ID = mutation.EntityID
		input.IdempotencyKey = mutation.MutationID
		transaction, err := s.finance.CreateTransaction(ctx, userID, input)
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, transaction.Version, mustJSON(transaction))
	case OperationUpdate:
		input, err := transactionInputFromPayload(mutation.Payload)
		if err != nil {
			return MutationResult{}, err
		}
		transaction, err := s.finance.UpdateTransaction(ctx, userID, mutation.EntityID, finance.UpdateTransactionInput{
			Type:                input.Type,
			SourceWalletID:      input.SourceWalletID,
			DestinationWalletID: input.DestinationWalletID,
			CategoryID:          input.CategoryID,
			AmountVND:           input.AmountVND,
			TargetBalanceVND:    input.TargetBalanceVND,
			OccurredAt:          input.OccurredAt,
			Note:                input.Note,
			WithPerson:          input.WithPerson,
			EventRef:            input.EventRef,
			ExcludedFromReports: input.ExcludedFromReports,
		})
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, transaction.Version, mustJSON(transaction))
	case OperationArchive:
		if err := s.finance.ArchiveTransaction(ctx, userID, mutation.EntityID); err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, mutation.BaseVersion+1, json.RawMessage(`{}`))
	default:
		return MutationResult{}, fmt.Errorf("%w: unsupported transaction operation", ErrValidation)
	}
}

func (s *Service) conflictForStaleBase(ctx context.Context, userID string, mutation Mutation) (Conflict, bool, error) {
	serverVersion, serverPayload, found, err := s.store.EntityVersionAndPayload(ctx, userID, mutation.EntityType, mutation.EntityID)
	if err != nil {
		return Conflict{}, false, err
	}
	if !found {
		return Conflict{}, false, nil
	}
	if mutation.BaseVersion <= 0 || serverVersion == mutation.BaseVersion {
		return Conflict{}, false, nil
	}
	return Conflict{
		EntityType:    mutation.EntityType,
		EntityID:      mutation.EntityID,
		Operation:     mutation.Operation,
		BaseVersion:   mutation.BaseVersion,
		ServerVersion: serverVersion,
		LocalPayload:  normalizedPayload(mutation.Payload),
		ServerPayload: serverPayload,
	}, true, nil
}

func (s *Service) appliedWithChange(ctx context.Context, userID string, mutation Mutation, version int64, payload json.RawMessage) (MutationResult, error) {
	payload = normalizedPayload(payload)
	change, err := s.store.AppendChange(ctx, userID, mutation.EntityType, mutation.EntityID, mutation.Operation, version, payload)
	if err != nil {
		return MutationResult{}, err
	}
	return MutationResult{
		MutationID: mutation.MutationID,
		EntityType: mutation.EntityType,
		EntityID:   mutation.EntityID,
		Operation:  mutation.Operation,
		State:      ResultApplied,
		Version:    version,
		Payload:    change.Payload,
	}, nil
}

func validateBatch(userID string, mutations []Mutation) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("%w: user is required", ErrValidation)
	}
	if len(mutations) == 0 || len(mutations) > 100 {
		return fmt.Errorf("%w: mutation batch must contain 1 to 100 items", ErrValidation)
	}
	var previous int64
	seen := map[string]struct{}{}
	for _, mutation := range mutations {
		if strings.TrimSpace(mutation.MutationID) == "" || strings.TrimSpace(mutation.DeviceID) == "" || strings.TrimSpace(mutation.EntityID) == "" {
			return fmt.Errorf("%w: mutation identity is required", ErrValidation)
		}
		if _, ok := seen[mutation.MutationID]; ok {
			return fmt.Errorf("%w: duplicate mutation_id in batch", ErrValidation)
		}
		seen[mutation.MutationID] = struct{}{}
		if mutation.Sequence <= previous {
			return fmt.Errorf("%w: mutations must be ordered by increasing sequence", ErrValidation)
		}
		previous = mutation.Sequence
	}
	return nil
}

func decodePayload(payload json.RawMessage, target any) error {
	if err := json.Unmarshal(normalizedPayload(payload), target); err != nil {
		return fmt.Errorf("%w: invalid mutation payload", ErrValidation)
	}
	return nil
}

func transactionInputFromPayload(payload json.RawMessage) (finance.CreateTransactionInput, error) {
	var input struct {
		Type                finance.TransactionType `json:"type"`
		SourceWalletID      string                  `json:"source_wallet_id"`
		DestinationWalletID string                  `json:"destination_wallet_id"`
		CategoryID          string                  `json:"category_id"`
		AmountVND           int64                   `json:"amount_vnd"`
		TargetBalanceVND    *int64                  `json:"target_balance_vnd"`
		OccurredAt          time.Time               `json:"occurred_at"`
		Note                string                  `json:"note"`
		WithPerson          string                  `json:"with_person"`
		EventRef            string                  `json:"event_ref"`
		ExcludedFromReports bool                    `json:"excluded_from_reports"`
	}
	if err := decodePayload(payload, &input); err != nil {
		return finance.CreateTransactionInput{}, err
	}
	return finance.CreateTransactionInput{
		Type:                input.Type,
		SourceWalletID:      input.SourceWalletID,
		DestinationWalletID: input.DestinationWalletID,
		CategoryID:          input.CategoryID,
		AmountVND:           input.AmountVND,
		TargetBalanceVND:    input.TargetBalanceVND,
		OccurredAt:          input.OccurredAt,
		Note:                input.Note,
		WithPerson:          input.WithPerson,
		EventRef:            input.EventRef,
		ExcludedFromReports: input.ExcludedFromReports,
	}, nil
}

func rejectedResult(mutation Mutation, reason string) MutationResult {
	return MutationResult{
		MutationID: mutation.MutationID,
		EntityType: mutation.EntityType,
		EntityID:   mutation.EntityID,
		Operation:  mutation.Operation,
		State:      ResultRejected,
		Reason:     reason,
	}
}

func safeReason(err error) string {
	message := err.Error()
	if strings.Contains(message, ": ") {
		return message[strings.LastIndex(message, ": ")+2:]
	}
	return message
}

func mustJSON(value any) json.RawMessage {
	body, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return body
}
