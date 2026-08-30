package finance_test

import (
	"errors"
	"testing"

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
