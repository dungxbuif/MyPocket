package finance

import "fmt"

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
		return AccountingEffect{
			SourceBalanceVND: input.SourceBalanceVND + input.AmountVND,
			SourceDeltaVND:   input.AmountVND,
		}, nil
	case TransactionExpense:
		if err := validatePositiveAmount(input.AmountVND); err != nil {
			return AccountingEffect{}, err
		}
		if input.DestinationWalletID != "" {
			return AccountingEffect{}, fmt.Errorf("%w: expense cannot have destination wallet", ErrValidation)
		}
		return AccountingEffect{
			SourceBalanceVND: input.SourceBalanceVND - input.AmountVND,
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
		return AccountingEffect{
			SourceBalanceVND:      input.SourceBalanceVND - input.AmountVND,
			DestinationBalanceVND: input.DestinationBalanceVND + input.AmountVND,
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
		delta := *input.TargetBalanceVND - input.SourceBalanceVND
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

func validatePositiveAmount(amountVND int64) error {
	if amountVND <= 0 {
		return fmt.Errorf("%w: amount must be a positive VND integer", ErrValidation)
	}
	return nil
}
