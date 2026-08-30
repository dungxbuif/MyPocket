package finance_test

import (
	"errors"
	"testing"

	"mypocket/internal/finance"
)

func TestCreateWalletRejectsCreditMetadataForCashWallet(t *testing.T) {
	_, err := finance.ValidateCreateWallet(finance.CreateWalletInput{
		Name:           "Tiền mặt",
		Type:           finance.WalletCash,
		CreditLimitVND: ptrMoney(10_000_000),
	})

	if !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateCreditWalletAcceptsCreditMetadata(t *testing.T) {
	input, err := finance.ValidateCreateWallet(finance.CreateWalletInput{
		Name:           "Thẻ tín dụng",
		Type:           finance.WalletCredit,
		CreditLimitVND: ptrMoney(20_000_000),
		StatementDay:   ptrInt(20),
		PaymentDueDay:  ptrInt(5),
	})

	if err != nil {
		t.Fatalf("expected credit wallet metadata to be accepted, got %v", err)
	}
	if input.Name != "Thẻ tín dụng" {
		t.Fatalf("expected wallet name to survive validation, got %q", input.Name)
	}
}

func ptrMoney(value int64) *int64 {
	return &value
}

func ptrInt(value int) *int {
	return &value
}
