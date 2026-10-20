package finance_test

import (
	"errors"
	"testing"

	"mypocket/internal/finance"
)

func TestCreateWalletRejectsCreditMetadataForCashWallet(t *testing.T) {
	_, err := finance.ValidateCreateWallet(finance.CreateWalletInput{
		Name:           "Tiền mặt",
		Type:           finance.WalletBasic,
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

func TestCreateWalletAcceptsOnlyBehaviorTypes(t *testing.T) {
	for _, value := range []finance.WalletType{finance.WalletBasic, finance.WalletGoal, finance.WalletCredit} {
		if _, err := finance.ValidateCreateWallet(finance.CreateWalletInput{Name: "Ví", Type: value}); err != nil {
			t.Fatalf("expected behavior type %q to be accepted, got %v", value, err)
		}
	}
	for _, legacy := range []finance.WalletType{"cash", "bank", "e_wallet", "savings", "debt"} {
		if _, err := finance.ValidateCreateWallet(finance.CreateWalletInput{Name: "Ví cũ", Type: legacy}); !errors.Is(err, finance.ErrValidation) {
			t.Fatalf("expected legacy wallet type %q to be rejected, got %v", legacy, err)
		}
	}
}

func TestCreateGoalWalletAcceptsOnlyGoalMetadata(t *testing.T) {
	deadline := "2027-12-31"
	input, err := finance.ValidateCreateWallet(finance.CreateWalletInput{
		Name:           "Quỹ tự do",
		Type:           finance.WalletGoal,
		GoalTargetVND:  ptrMoney(50_000_000),
		GoalDeadlineOn: &deadline,
	})
	if err != nil {
		t.Fatalf("expected goal wallet metadata to be accepted, got %v", err)
	}
	if input.GoalTargetVND == nil || *input.GoalTargetVND != 50_000_000 || input.GoalDeadlineOn == nil || *input.GoalDeadlineOn != deadline {
		t.Fatalf("goal metadata was not preserved: %#v", input)
	}

	if _, err := finance.ValidateCreateWallet(finance.CreateWalletInput{Name: "Ví thường", Type: finance.WalletBasic, GoalTargetVND: ptrMoney(1_000_000)}); !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected basic wallet with goal metadata to be rejected, got %v", err)
	}
	if _, err := finance.ValidateCreateWallet(finance.CreateWalletInput{Name: "Quỹ lỗi", Type: finance.WalletGoal, GoalTargetVND: ptrMoney(0)}); !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected non-positive goal target to be rejected, got %v", err)
	}
}

func ptrMoney(value int64) *int64 {
	return &value
}

func ptrInt(value int) *int {
	return &value
}
