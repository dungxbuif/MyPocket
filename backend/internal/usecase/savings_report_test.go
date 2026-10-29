package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

// Embedding unrelated repository methods keeps test doubles read-only: the
// savings service must never need write or lookup methods on ledger adapters.
type savingsWallets struct {
	repository.WalletRepository
	rows []entity.Wallet
	err  error
}

func (s savingsWallets) List(owner string) ([]entity.Wallet, error) {
	if owner != "owner" {
		panic("wrong owner scope")
	}
	return s.rows, s.err
}

type savingsTransactions struct {
	repository.TransactionRepository
	rows []entity.Transaction
	err  error
}

func (s savingsTransactions) List(owner string) ([]entity.Transaction, error) {
	if owner != "owner" {
		panic("wrong owner scope")
	}
	return s.rows, s.err
}

type savingsUsers struct {
	repository.UserRepository
	user *entity.User
	err  error
}

func (s savingsUsers) FindByID(owner string) (*entity.User, error) {
	if owner != "owner" {
		panic("wrong owner scope")
	}
	return s.user, s.err
}

func TestSavingsServiceUsesAccountTimezoneAndOwnedLedger(t *testing.T) {
	now := time.Date(2026, 9, 30, 18, 0, 0, 0, time.UTC)
	s := &SavingsReports{Users: savingsUsers{user: &entity.User{ID: "owner", Timezone: "Asia/Ho_Chi_Minh"}}, Wallets: savingsWallets{rows: []entity.Wallet{{ID: "g", OwnerID: "owner", Type: entity.WalletTypeGoal}}}, Transactions: savingsTransactions{rows: []entity.Transaction{{ID: "tx", OwnerID: "owner", WalletID: "g", Type: entity.TransactionTypeIncome, Amount: 100, OccurredAt: now, IncludedInReports: true}, {ID: "foreign", OwnerID: "foreign", WalletID: "g", Type: entity.TransactionTypeIncome, Amount: 900, OccurredAt: now, IncludedInReports: true}}}, Now: func() time.Time { return now }}
	r, err := s.Report("owner", "", "")
	if err != nil || r.Month != "2026-10" || r.Timezone != "Asia/Ho_Chi_Minh" || r.Goals.ClosingBalance != 100 || r.Surplus.Income != 100 {
		t.Fatalf("report=%+v err=%v", r, err)
	}
	p, err := s.TransactionsPage("owner", "", "", "goals", "all", 50, 0)
	if err != nil || p.Total != 1 || p.Items[0].ID != "tx" {
		t.Fatalf("page=%+v err=%v", p, err)
	}
}

func TestSavingsServiceFailsClosedWhenDependenciesFail(t *testing.T) {
	failed := errors.New("unavailable")
	for _, s := range []*SavingsReports{
		{},
		{Users: savingsUsers{err: failed}, Wallets: savingsWallets{}, Transactions: savingsTransactions{}},
		{Users: savingsUsers{user: nil}, Wallets: savingsWallets{}, Transactions: savingsTransactions{}},
		{Users: savingsUsers{user: &entity.User{ID: "owner", Timezone: "invalid/timezone"}}, Wallets: savingsWallets{}, Transactions: savingsTransactions{}},
		{Users: savingsUsers{user: &entity.User{ID: "foreign", Timezone: "UTC"}}, Wallets: savingsWallets{}, Transactions: savingsTransactions{}},
		{Users: savingsUsers{user: &entity.User{ID: "owner", Timezone: "UTC"}}, Wallets: savingsWallets{err: failed}, Transactions: savingsTransactions{}},
		{Users: savingsUsers{user: &entity.User{ID: "owner", Timezone: "UTC"}}, Wallets: savingsWallets{}, Transactions: savingsTransactions{err: failed}},
	} {
		if _, err := s.Report("owner", "2026-01", ""); err == nil {
			t.Fatalf("dependency failure fabricated report: %+v", s)
		}
		if _, err := s.TransactionsPage("owner", "2026-01", "", "goals", "all", 50, 0); err == nil {
			t.Fatalf("dependency failure fabricated page: %+v", s)
		}
	}
}
