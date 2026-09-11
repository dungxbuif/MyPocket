package finance_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"mypocket/internal/finance"
)

func TestApplyAccountingEffect(t *testing.T) {
	tests := []struct {
		name        string
		input       finance.AccountingInput
		wantSource  int64
		wantDest    int64
		wantSrcDiff int64
		wantDstDiff int64
	}{
		{
			name: "income increases source",
			input: finance.AccountingInput{
				Type:             finance.TransactionIncome,
				AmountVND:        100_000,
				SourceWalletID:   "wallet_cash",
				SourceBalanceVND: 1_000_000,
			},
			wantSource:  1_100_000,
			wantSrcDiff: 100_000,
		},
		{
			name: "expense decreases source",
			input: finance.AccountingInput{
				Type:             finance.TransactionExpense,
				AmountVND:        40_000,
				SourceWalletID:   "wallet_cash",
				SourceBalanceVND: 1_000_000,
			},
			wantSource:  960_000,
			wantSrcDiff: -40_000,
		},
		{
			name: "transfer moves money between wallets",
			input: finance.AccountingInput{
				Type:                  finance.TransactionTransfer,
				AmountVND:             250_000,
				SourceWalletID:        "wallet_cash",
				DestinationWalletID:   "wallet_bank",
				SourceBalanceVND:      1_000_000,
				DestinationBalanceVND: 100_000,
			},
			wantSource:  750_000,
			wantDest:    350_000,
			wantSrcDiff: -250_000,
			wantDstDiff: 250_000,
		},
		{
			name: "adjustment sets source target balance",
			input: finance.AccountingInput{
				Type:             finance.TransactionAdjustment,
				SourceWalletID:   "wallet_cash",
				SourceBalanceVND: 1_000_000,
				TargetBalanceVND: ptrMoney(840_000),
			},
			wantSource:  840_000,
			wantSrcDiff: -160_000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := finance.ApplyAccountingEffect(tt.input)
			if err != nil {
				t.Fatalf("expected accounting effect, got %v", err)
			}
			if got.SourceBalanceVND != tt.wantSource {
				t.Fatalf("source balance = %d, want %d", got.SourceBalanceVND, tt.wantSource)
			}
			if got.DestinationBalanceVND != tt.wantDest {
				t.Fatalf("destination balance = %d, want %d", got.DestinationBalanceVND, tt.wantDest)
			}
			if got.SourceDeltaVND != tt.wantSrcDiff {
				t.Fatalf("source delta = %d, want %d", got.SourceDeltaVND, tt.wantSrcDiff)
			}
			if got.DestinationDeltaVND != tt.wantDstDiff {
				t.Fatalf("destination delta = %d, want %d", got.DestinationDeltaVND, tt.wantDstDiff)
			}
		})
	}
}

func TestApplyAccountingEffectRejectsInvalidTransfers(t *testing.T) {
	tests := []struct {
		name  string
		input finance.AccountingInput
	}{
		{
			name: "missing destination",
			input: finance.AccountingInput{
				Type:             finance.TransactionTransfer,
				AmountVND:        100_000,
				SourceWalletID:   "wallet_cash",
				SourceBalanceVND: 1_000_000,
			},
		},
		{
			name: "same wallet",
			input: finance.AccountingInput{
				Type:                finance.TransactionTransfer,
				AmountVND:           100_000,
				SourceWalletID:      "wallet_cash",
				DestinationWalletID: "wallet_cash",
				SourceBalanceVND:    1_000_000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := finance.ApplyAccountingEffect(tt.input)
			if !errors.Is(err, finance.ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestApplyAccountingEffectRejectsInvalidAmounts(t *testing.T) {
	_, err := finance.ApplyAccountingEffect(finance.AccountingInput{
		Type:             finance.TransactionExpense,
		AmountVND:        0,
		SourceWalletID:   "wallet_cash",
		SourceBalanceVND: 1_000_000,
	})

	if !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestApplyAccountingEffectRejectsAdjustmentWithoutTarget(t *testing.T) {
	_, err := finance.ApplyAccountingEffect(finance.AccountingInput{
		Type:             finance.TransactionAdjustment,
		SourceWalletID:   "wallet_cash",
		SourceBalanceVND: 1_000_000,
	})

	if !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestApplyAccountingEffectRejectsBalanceOverflow(t *testing.T) {
	tests := []struct {
		name  string
		input finance.AccountingInput
	}{
		{
			name: "income cannot overflow the source balance",
			input: finance.AccountingInput{Type: finance.TransactionIncome, AmountVND: 1, SourceWalletID: "source", SourceBalanceVND: math.MaxInt64},
		},
		{
			name: "expense cannot underflow the source balance",
			input: finance.AccountingInput{Type: finance.TransactionExpense, AmountVND: 1, SourceWalletID: "source", SourceBalanceVND: math.MinInt64},
		},
		{
			name: "transfer cannot overflow the destination balance",
			input: finance.AccountingInput{Type: finance.TransactionTransfer, AmountVND: 1, SourceWalletID: "source", DestinationWalletID: "destination", DestinationBalanceVND: math.MaxInt64},
		},
		{
			name: "adjustment cannot overflow its recorded delta",
			input: finance.AccountingInput{Type: finance.TransactionAdjustment, SourceWalletID: "source", SourceBalanceVND: math.MinInt64, TargetBalanceVND: ptrMoney(math.MaxInt64)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := finance.ApplyAccountingEffect(tt.input)
			if !errors.Is(err, finance.ErrValidation) {
				t.Fatalf("expected overflow validation error, got %v", err)
			}
		})
	}
}

func TestValidateUpdateTransactionPreservesReceiptReference(t *testing.T) {
	normalized, err := finance.ValidateUpdateTransaction(finance.UpdateTransactionInput{
		Type:            finance.TransactionExpense,
		SourceWalletID:  "wallet_cash",
		CategoryID:      "category_food",
		ReceiptObjectID: "receipt_123",
		AmountVND:       50_000,
		OccurredAt:      time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("validate update: %v", err)
	}
	if normalized.ReceiptObjectID != "receipt_123" {
		t.Fatalf("receipt reference = %q, want receipt_123", normalized.ReceiptObjectID)
	}
}
