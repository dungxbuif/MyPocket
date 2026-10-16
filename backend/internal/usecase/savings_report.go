package usecase

import (
	"errors"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

// SavingsReports reads existing ledger repositories without persisting derived
// balances. Now is injectable so both sections use one captured report instant.
type SavingsReports struct {
	Wallets      repository.WalletRepository
	Transactions repository.TransactionRepository
	Users        repository.UserRepository
	Now          func() time.Time
}

func (s *SavingsReports) ledger(owner string) ([]entity.Wallet, []entity.Transaction, time.Time, *time.Location, error) {
	if owner == "" || s.Users == nil || s.Wallets == nil || s.Transactions == nil {
		return nil, nil, time.Time{}, nil, errors.New("savings repositories unavailable")
	}
	u, err := s.Users.FindByID(owner)
	if err != nil {
		return nil, nil, time.Time{}, nil, err
	}
	if u == nil || u.ID != owner || u.Timezone == "" {
		return nil, nil, time.Time{}, nil, errors.New("savings account unavailable")
	}
	loc, err := time.LoadLocation(u.Timezone)
	if err != nil {
		return nil, nil, time.Time{}, nil, err
	}
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	wallets, err := s.Wallets.List(owner)
	if err != nil {
		return nil, nil, time.Time{}, nil, err
	}
	rows, err := s.Transactions.List(owner)
	if err != nil {
		return nil, nil, time.Time{}, nil, err
	}
	return wallets, rows, now, loc, nil
}

func (s *SavingsReports) Report(owner, month, walletID string) (entity.SavingsReport, error) {
	wallets, rows, now, loc, err := s.ledger(owner)
	if err != nil {
		return entity.SavingsReport{}, err
	}
	return entity.CalculateSavingsReport(owner, wallets, rows, month, walletID, now, loc)
}

func (s *SavingsReports) TransactionsPage(owner, month, walletID, section, kind string, limit, offset int) (entity.SavingsTransactionPage, error) {
	wallets, rows, now, loc, err := s.ledger(owner)
	if err != nil {
		return entity.SavingsTransactionPage{}, err
	}
	return entity.SavingsReportTransactions(owner, wallets, rows, month, walletID, section, kind, limit, offset, now, loc)
}
