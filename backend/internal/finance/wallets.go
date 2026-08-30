package finance

import "fmt"

func ValidateCreateWallet(input CreateWalletInput) (CreateWalletInput, error) {
	input.Name = trimmed(input.Name)
	if input.Name == "" {
		return CreateWalletInput{}, fmt.Errorf("%w: wallet name is required", ErrValidation)
	}
	if !validWalletType(input.Type) {
		return CreateWalletInput{}, fmt.Errorf("%w: unsupported wallet type", ErrValidation)
	}
	if input.Type != WalletCredit && hasCreditMetadata(input) {
		return CreateWalletInput{}, fmt.Errorf("%w: credit metadata requires credit wallet type", ErrValidation)
	}
	if input.CreditLimitVND != nil && *input.CreditLimitVND < 0 {
		return CreateWalletInput{}, fmt.Errorf("%w: credit limit cannot be negative", ErrValidation)
	}
	if !validDay(input.StatementDay) || !validDay(input.PaymentDueDay) {
		return CreateWalletInput{}, fmt.Errorf("%w: credit day must be between 1 and 31", ErrValidation)
	}
	return input, nil
}

func ValidateUpdateWallet(input UpdateWalletInput) (UpdateWalletInput, error) {
	input.Name = trimmed(input.Name)
	if input.Name == "" {
		return UpdateWalletInput{}, fmt.Errorf("%w: wallet name is required", ErrValidation)
	}
	return input, nil
}

func validWalletType(value WalletType) bool {
	switch value {
	case WalletCash, WalletBank, WalletCredit, WalletEWallet, WalletSavings, WalletDebt:
		return true
	default:
		return false
	}
}

func hasCreditMetadata(input CreateWalletInput) bool {
	return input.CreditLimitVND != nil || input.StatementDay != nil || input.PaymentDueDay != nil
}

func validDay(value *int) bool {
	return value == nil || (*value >= 1 && *value <= 31)
}
