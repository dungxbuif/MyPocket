package finance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
)

func ValidateCreateReceiptObject(input CreateReceiptObjectInput) (CreateReceiptObjectInput, error) {
	input.ObjectKey = trimmed(input.ObjectKey)
	input.ContentType = trimmed(input.ContentType)
	input.ChecksumSHA256 = trimmed(input.ChecksumSHA256)
	input.OriginalFilename = trimmed(input.OriginalFilename)
	if input.ObjectKey == "" || input.ContentType == "" || input.SizeBytes <= 0 || len(input.ChecksumSHA256) != sha256.Size*2 {
		return CreateReceiptObjectInput{}, ErrValidation
	}
	for _, char := range input.ChecksumSHA256 {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return CreateReceiptObjectInput{}, ErrValidation
		}
	}
	return input, nil
}

func ApplyAccountingEffect(input AccountingInput) (AccountingEffect, error) {
	input.SourceWalletID = trimmed(input.SourceWalletID)
	input.DestinationWalletID = trimmed(input.DestinationWalletID)

	if input.SourceWalletID == "" {
		return AccountingEffect{}, fmt.Errorf("%w: source wallet is required", ErrValidation)
	}

	switch input.Type {
	case TransactionIncome:
		if err := validatePositiveAmount(input.AmountVND); err != nil {
			return AccountingEffect{}, err
		}
		if input.DestinationWalletID != "" {
			return AccountingEffect{}, fmt.Errorf("%w: income cannot have destination wallet", ErrValidation)
		}
		sourceBalance, ok := checkedAdd(input.SourceBalanceVND, input.AmountVND)
		if !ok {
			return AccountingEffect{}, fmt.Errorf("%w: income would overflow source balance", ErrValidation)
		}
		return AccountingEffect{
			SourceBalanceVND: sourceBalance,
			SourceDeltaVND:   input.AmountVND,
		}, nil
	case TransactionExpense:
		if err := validatePositiveAmount(input.AmountVND); err != nil {
			return AccountingEffect{}, err
		}
		if input.DestinationWalletID != "" {
			return AccountingEffect{}, fmt.Errorf("%w: expense cannot have destination wallet", ErrValidation)
		}
		sourceBalance, ok := checkedSub(input.SourceBalanceVND, input.AmountVND)
		if !ok {
			return AccountingEffect{}, fmt.Errorf("%w: expense would overflow source balance", ErrValidation)
		}
		return AccountingEffect{
			SourceBalanceVND: sourceBalance,
			SourceDeltaVND:   -input.AmountVND,
		}, nil
	case TransactionTransfer:
		if err := validatePositiveAmount(input.AmountVND); err != nil {
			return AccountingEffect{}, err
		}
		if input.DestinationWalletID == "" {
			return AccountingEffect{}, fmt.Errorf("%w: transfer destination wallet is required", ErrValidation)
		}
		if input.SourceWalletID == input.DestinationWalletID {
			return AccountingEffect{}, fmt.Errorf("%w: transfer wallets must differ", ErrValidation)
		}
		sourceBalance, sourceOK := checkedSub(input.SourceBalanceVND, input.AmountVND)
		destinationBalance, destinationOK := checkedAdd(input.DestinationBalanceVND, input.AmountVND)
		if !sourceOK || !destinationOK {
			return AccountingEffect{}, fmt.Errorf("%w: transfer would overflow wallet balance", ErrValidation)
		}
		return AccountingEffect{
			SourceBalanceVND:      sourceBalance,
			DestinationBalanceVND: destinationBalance,
			SourceDeltaVND:        -input.AmountVND,
			DestinationDeltaVND:   input.AmountVND,
		}, nil
	case TransactionAdjustment:
		if input.TargetBalanceVND == nil {
			return AccountingEffect{}, fmt.Errorf("%w: adjustment target balance is required", ErrValidation)
		}
		if input.DestinationWalletID != "" {
			return AccountingEffect{}, fmt.Errorf("%w: adjustment cannot have destination wallet", ErrValidation)
		}
		delta, ok := checkedSub(*input.TargetBalanceVND, input.SourceBalanceVND)
		if !ok {
			return AccountingEffect{}, fmt.Errorf("%w: adjustment delta would overflow", ErrValidation)
		}
		if delta == 0 {
			return AccountingEffect{}, fmt.Errorf("%w: adjustment must change balance", ErrValidation)
		}
		return AccountingEffect{
			SourceBalanceVND: *input.TargetBalanceVND,
			SourceDeltaVND:   delta,
		}, nil
	default:
		return AccountingEffect{}, fmt.Errorf("%w: unsupported transaction type", ErrValidation)
	}
}

func checkedAdd(left int64, right int64) (int64, bool) {
	if (right > 0 && left > math.MaxInt64-right) || (right < 0 && left < math.MinInt64-right) {
		return 0, false
	}
	return left + right, true
}

func checkedSub(left int64, right int64) (int64, bool) {
	if (right > 0 && left < math.MinInt64+right) || (right < 0 && left > math.MaxInt64+right) {
		return 0, false
	}
	return left - right, true
}

func validatePositiveAmount(amountVND int64) error {
	if amountVND <= 0 {
		return fmt.Errorf("%w: amount must be a positive VND integer", ErrValidation)
	}
	return nil
}

func ValidateCreateTransaction(input CreateTransactionInput) (CreateTransactionInput, error) {
	input.IdempotencyKey = trimmed(input.IdempotencyKey)

	if input.IdempotencyKey == "" {
		return CreateTransactionInput{}, fmt.Errorf("%w: idempotency key is required", ErrValidation)
	}
	return validateTransactionFields(input)
}

func ValidateUpdateTransaction(input UpdateTransactionInput) (CreateTransactionInput, error) {
	return validateTransactionFields(CreateTransactionInput{
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
}

func validateTransactionFields(input CreateTransactionInput) (CreateTransactionInput, error) {
	input.SourceWalletID = trimmed(input.SourceWalletID)
	input.DestinationWalletID = trimmed(input.DestinationWalletID)
	input.CategoryID = trimmed(input.CategoryID)
	input.ReceiptObjectID = trimmed(input.ReceiptObjectID)
	input.Note = trimmed(input.Note)
	input.WithPerson = trimmed(input.WithPerson)
	input.EventRef = trimmed(input.EventRef)

	if input.OccurredAt.IsZero() {
		return CreateTransactionInput{}, fmt.Errorf("%w: occurred_at is required", ErrValidation)
	}

	switch input.Type {
	case TransactionIncome:
		if input.CategoryID == "" {
			return CreateTransactionInput{}, fmt.Errorf("%w: income category is required", ErrValidation)
		}
		if input.TargetBalanceVND != nil {
			return CreateTransactionInput{}, fmt.Errorf("%w: income cannot target balance", ErrValidation)
		}
	case TransactionExpense:
		if input.CategoryID == "" {
			return CreateTransactionInput{}, fmt.Errorf("%w: expense category is required", ErrValidation)
		}
		if input.TargetBalanceVND != nil {
			return CreateTransactionInput{}, fmt.Errorf("%w: expense cannot target balance", ErrValidation)
		}
	case TransactionTransfer:
		if input.CategoryID != "" {
			return CreateTransactionInput{}, fmt.Errorf("%w: transfer cannot have category", ErrValidation)
		}
		if input.TargetBalanceVND != nil {
			return CreateTransactionInput{}, fmt.Errorf("%w: transfer cannot target balance", ErrValidation)
		}
	case TransactionAdjustment:
		if input.CategoryID != "" {
			return CreateTransactionInput{}, fmt.Errorf("%w: adjustment cannot have category", ErrValidation)
		}
		if input.TargetBalanceVND == nil {
			return CreateTransactionInput{}, fmt.Errorf("%w: adjustment target balance is required", ErrValidation)
		}
		if *input.TargetBalanceVND <= 0 {
			return CreateTransactionInput{}, fmt.Errorf("%w: adjustment target balance must be positive", ErrValidation)
		}
		input.AmountVND = *input.TargetBalanceVND
	default:
		return CreateTransactionInput{}, fmt.Errorf("%w: unsupported transaction type", ErrValidation)
	}

	if input.Type != TransactionAdjustment {
		if err := validatePositiveAmount(input.AmountVND); err != nil {
			return CreateTransactionInput{}, err
		}
	}
	if _, err := ApplyAccountingEffect(AccountingInput{
		Type:                input.Type,
		AmountVND:           input.AmountVND,
		SourceWalletID:      input.SourceWalletID,
		DestinationWalletID: input.DestinationWalletID,
		TargetBalanceVND:    input.TargetBalanceVND,
	}); err != nil {
		return CreateTransactionInput{}, err
	}
	return input, nil
}

func transactionRequestHash(input CreateTransactionInput) (string, error) {
	payload := struct {
		Type                TransactionType `json:"type"`
		SourceWalletID      string          `json:"source_wallet_id"`
		DestinationWalletID string          `json:"destination_wallet_id"`
		CategoryID          string          `json:"category_id"`
		ReceiptObjectID     string          `json:"receipt_object_id"`
		AmountVND           int64           `json:"amount_vnd"`
		TargetBalanceVND    *int64          `json:"target_balance_vnd"`
		OccurredAt          string          `json:"occurred_at"`
		Note                string          `json:"note"`
		WithPerson          string          `json:"with_person"`
		EventRef            string          `json:"event_ref"`
		ExcludedFromReports bool            `json:"excluded_from_reports"`
	}{
		Type:                input.Type,
		SourceWalletID:      input.SourceWalletID,
		DestinationWalletID: input.DestinationWalletID,
		CategoryID:          input.CategoryID,
		ReceiptObjectID:     input.ReceiptObjectID,
		AmountVND:           input.AmountVND,
		TargetBalanceVND:    input.TargetBalanceVND,
		OccurredAt:          input.OccurredAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		Note:                input.Note,
		WithPerson:          input.WithPerson,
		EventRef:            input.EventRef,
		ExcludedFromReports: input.ExcludedFromReports,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("hash transaction request: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}
