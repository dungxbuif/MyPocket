package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"mypocket/internal/audit"
	"mypocket/internal/finance"
	"mypocket/internal/portfolio"
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

type PortfolioCommands interface {
	ListPositions(ctx context.Context, userID string, includeArchived bool) ([]portfolio.Position, error)
	CreatePosition(ctx context.Context, userID string, input portfolio.CreatePositionInput) (portfolio.Position, error)
	GetPosition(ctx context.Context, userID, assetID string) (portfolio.Position, error)
	ArchivePosition(ctx context.Context, userID, assetID string, baseVersion int64) error
	AddTrade(ctx context.Context, userID, assetID string, input portfolio.AddTradeInput) (portfolio.Position, error)
	UpdateTrade(ctx context.Context, userID, assetID, tradeID string, input portfolio.UpdateTradeInput) (portfolio.Position, error)
	ArchiveTrade(ctx context.Context, userID, assetID, tradeID string, baseVersion int64) (portfolio.Position, error)
	AddPrice(ctx context.Context, userID, assetID string, input portfolio.AddPriceInput) (portfolio.Position, error)
}

type AuditSink interface {
	Append(ctx context.Context, event audit.Event) error
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
	store     Store
	finance   FinanceCommands
	portfolio PortfolioCommands
	audit     AuditSink
}

func NewService(store Store, financeCommands FinanceCommands, portfolioCommands ...PortfolioCommands) *Service {
	var portfolioCommand PortfolioCommands
	if len(portfolioCommands) > 0 {
		portfolioCommand = portfolioCommands[0]
	}
	return &Service{store: store, finance: financeCommands, portfolio: portfolioCommand}
}

func (s *Service) SetAuditSink(sink AuditSink) {
	s.audit = sink
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
	var assets []portfolio.Position
	if s.portfolio != nil {
		assets, err = s.portfolio.ListPositions(ctx, userID, false)
		if err != nil {
			return Snapshot{}, fmt.Errorf("list resync assets: %w", err)
		}
	}
	cursor, err := s.store.CurrentCursor(ctx, userID)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Wallets: wallets, Categories: categories, Transactions: transactions, Assets: assets, NextCursor: cursor}, nil
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
		s.appendMutationAudit(ctx, userID, stored)
		return stored, nil
	}

	result, err := s.executeMutation(ctx, userID, mutation)
	if err != nil {
		if errors.Is(err, finance.ErrValidation) || errors.Is(err, finance.ErrForbidden) || errors.Is(err, portfolio.ErrValidation) || errors.Is(err, portfolio.ErrForbidden) || errors.Is(err, portfolio.ErrOversell) || errors.Is(err, portfolio.ErrConflict) || errors.Is(err, ErrValidation) {
			result = rejectedResult(mutation, safeReason(err))
		} else {
			return MutationResult{}, err
		}
	}
	if err := s.store.StoreMutationResult(ctx, userID, mutation.MutationID, hash, result); err != nil {
		return MutationResult{}, err
	}
	s.appendMutationAudit(ctx, userID, result)
	return result, nil
}

func (s *Service) appendMutationAudit(ctx context.Context, userID string, result MutationResult) {
	if s.audit == nil {
		return
	}
	outcome := audit.OutcomeSuccess
	severity := audit.SeverityInfo
	if result.State == ResultRejected {
		outcome = audit.OutcomeDenied
		severity = audit.SeverityWarn
	}
	if result.State == ResultConflict {
		outcome = audit.OutcomeConflict
		severity = audit.SeverityWarn
	}
	if result.State == ResultReplayed {
		outcome = audit.OutcomeReplayed
	}
	_ = s.audit.Append(ctx, audit.Event{
		CorrelationID: "sync:" + result.MutationID,
		ActorUserID:   userID,
		Action:        "sync." + string(result.Operation),
		EntityType:    string(result.EntityType),
		EntityID:      result.EntityID,
		Outcome:       outcome,
		Severity:      severity,
		Source:        audit.SourceSync,
		Metadata:      audit.SafeMetadata(map[string]any{"mutation_id": result.MutationID, "entity_type": result.EntityType, "operation": result.Operation, "version": result.Version, "reason": result.Reason}),
	})
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
	case EntityAsset:
		return s.executeAssetMutation(ctx, userID, mutation)
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
			ReceiptObjectID:     input.ReceiptObjectID,
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

func (s *Service) executeAssetMutation(ctx context.Context, userID string, mutation Mutation) (MutationResult, error) {
	if s.portfolio == nil {
		return MutationResult{}, fmt.Errorf("%w: portfolio sync unavailable", ErrValidation)
	}
	switch mutation.Operation {
	case OperationCreate:
		var input struct {
			Type              portfolio.AssetType   `json:"type"`
			Symbol            string                `json:"symbol"`
			Exchange          string                `json:"exchange"`
			Name              string                `json:"name"`
			Unit              string                `json:"unit"`
			PricingMode       portfolio.PricingMode `json:"pricing_mode"`
			ProviderKey       string                `json:"provider_key"`
			ProviderSymbol    string                `json:"provider_symbol"`
			IncludeInNetWorth *bool                 `json:"include_in_net_worth"`
		}
		if err := decodePayload(mutation.Payload, &input); err != nil {
			return MutationResult{}, err
		}
		asset, err := s.portfolio.CreatePosition(ctx, userID, portfolio.CreatePositionInput{
			ID:                mutation.EntityID,
			Type:              input.Type,
			Symbol:            input.Symbol,
			Exchange:          input.Exchange,
			Name:              input.Name,
			Unit:              input.Unit,
			PricingMode:       input.PricingMode,
			ProviderKey:       input.ProviderKey,
			ProviderSymbol:    input.ProviderSymbol,
			IncludeInNetWorth: input.IncludeInNetWorth,
		})
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, asset.Version, mustJSON(asset))
	case OperationArchive:
		if err := s.portfolio.ArchivePosition(ctx, userID, mutation.EntityID, mutation.BaseVersion); err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, mutation.BaseVersion+1, json.RawMessage(`{}`))
	case OperationAddTrade:
		input, err := assetTradeInputFromPayload(mutation.Payload)
		if err != nil {
			return MutationResult{}, err
		}
		asset, err := s.portfolio.AddTrade(ctx, userID, mutation.EntityID, input)
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, asset.Version, mustJSON(asset))
	case OperationUpdateTrade:
		input, err := assetUpdateTradeInputFromPayload(mutation.Payload)
		if err != nil {
			return MutationResult{}, err
		}
		tradeID := payloadString(mutation.Payload, "trade_id")
		if tradeID == "" {
			return MutationResult{}, fmt.Errorf("%w: asset trade_id is required", ErrValidation)
		}
		asset, err := s.portfolio.UpdateTrade(ctx, userID, mutation.EntityID, tradeID, input)
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, asset.Version, mustJSON(asset))
	case OperationArchiveTrade:
		tradeID := payloadString(mutation.Payload, "trade_id")
		if tradeID == "" {
			return MutationResult{}, fmt.Errorf("%w: asset trade_id is required", ErrValidation)
		}
		asset, err := s.portfolio.ArchiveTrade(ctx, userID, mutation.EntityID, tradeID, mutation.BaseVersion)
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, asset.Version, mustJSON(asset))
	case OperationAddPrice:
		var input struct {
			ID              string    `json:"id"`
			UnitPriceVND    int64     `json:"unit_price_vnd"`
			PricedAt        time.Time `json:"priced_at"`
			Source          string    `json:"source"`
			ProviderQuoteID string    `json:"provider_quote_id"`
			BaseVersion     int64     `json:"base_version"`
		}
		if err := decodePayload(mutation.Payload, &input); err != nil {
			return MutationResult{}, err
		}
		if input.Source == "" {
			input.Source = "manual"
		}
		asset, err := s.portfolio.AddPrice(ctx, userID, mutation.EntityID, portfolio.AddPriceInput(input))
		if err != nil {
			return MutationResult{}, err
		}
		return s.appliedWithChange(ctx, userID, mutation, asset.Version, mustJSON(asset))
	default:
		return MutationResult{}, fmt.Errorf("%w: unsupported asset operation", ErrValidation)
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
		ReceiptObjectID     string                  `json:"receipt_object_id"`
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
		ReceiptObjectID:     input.ReceiptObjectID,
		AmountVND:           input.AmountVND,
		TargetBalanceVND:    input.TargetBalanceVND,
		OccurredAt:          input.OccurredAt,
		Note:                input.Note,
		WithPerson:          input.WithPerson,
		EventRef:            input.EventRef,
		ExcludedFromReports: input.ExcludedFromReports,
	}, nil
}

func assetTradeInputFromPayload(payload json.RawMessage) (portfolio.AddTradeInput, error) {
	var input struct {
		ID           string              `json:"id"`
		Side         portfolio.TradeSide `json:"side"`
		Quantity     string              `json:"quantity"`
		UnitPriceVND int64               `json:"unit_price_vnd"`
		FeeVND       int64               `json:"fee_vnd"`
		OccurredAt   time.Time           `json:"occurred_at"`
		Note         string              `json:"note"`
		BaseVersion  int64               `json:"base_version"`
	}
	if err := decodePayload(payload, &input); err != nil {
		return portfolio.AddTradeInput{}, err
	}
	return portfolio.AddTradeInput(input), nil
}

func assetUpdateTradeInputFromPayload(payload json.RawMessage) (portfolio.UpdateTradeInput, error) {
	input, err := assetTradeInputFromPayload(payload)
	if err != nil {
		return portfolio.UpdateTradeInput{}, err
	}
	return portfolio.UpdateTradeInput{
		Side:         input.Side,
		Quantity:     input.Quantity,
		UnitPriceVND: input.UnitPriceVND,
		FeeVND:       input.FeeVND,
		OccurredAt:   input.OccurredAt,
		Note:         input.Note,
		BaseVersion:  input.BaseVersion,
	}, nil
}

func payloadString(payload json.RawMessage, key string) string {
	var values map[string]any
	if err := json.Unmarshal(normalizedPayload(payload), &values); err != nil {
		return ""
	}
	value, _ := values[key].(string)
	return strings.TrimSpace(value)
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
