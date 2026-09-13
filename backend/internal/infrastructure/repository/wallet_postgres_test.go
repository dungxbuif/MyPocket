package repository

import (
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestCalculateCurrentBalancesAppliesIncomeAndExpensePerWallet(t *testing.T) {
	wallets := []entity.Wallet{
		{ID: "wallet-1", OpeningBalance: 100000},
		{ID: "wallet-2", OpeningBalance: -50000},
	}
	transactions := []entity.Transaction{
		{WalletID: "wallet-1", Type: entity.TransactionTypeIncome, Amount: 25000},
		{WalletID: "wallet-1", Type: entity.TransactionTypeExpense, Amount: 10000},
		{WalletID: "wallet-2", Type: entity.TransactionTypeIncome, Amount: 5000},
	}

	calculateCurrentBalances(wallets, transactions)

	if wallets[0].CurrentBalance != 115000 {
		t.Fatalf("wallet-1 current balance = %d", wallets[0].CurrentBalance)
	}
	if wallets[1].CurrentBalance != -45000 {
		t.Fatalf("wallet-2 current balance = %d", wallets[1].CurrentBalance)
	}
}

func TestCalculateCurrentBalancesIgnoresUnsupportedLedgerKinds(t *testing.T) {
	wallets := []entity.Wallet{{ID: "wallet-1", OpeningBalance: 100000}}
	transactions := []entity.Transaction{{WalletID: "wallet-1", Type: "adjustment", Amount: 90000}}

	calculateCurrentBalances(wallets, transactions)

	if wallets[0].CurrentBalance != 100000 {
		t.Fatalf("current balance = %d", wallets[0].CurrentBalance)
	}
}
