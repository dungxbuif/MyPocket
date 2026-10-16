package entity

import (
	"errors"
	"math"
	"math/big"
	"sort"
	"time"
)

var (
	ErrSavingsInvalidQuery   = errors.New("invalid savings report query")
	ErrSavingsFutureMonth    = errors.New("savings month is in the future")
	ErrSavingsWalletNotFound = errors.New("goal wallet not found")
	ErrSavingsOverflow       = errors.New("savings arithmetic overflow")
	ErrSavingsInvalidLedger  = errors.New("invalid savings ledger effect")
)

type SavingsMetrics struct {
	OpeningBalance      int64 `json:"opening_balance"`
	ClosingBalance      int64 `json:"closing_balance"`
	Deposits            int64 `json:"deposits"`
	Withdrawals         int64 `json:"withdrawals"`
	Adjustments         int64 `json:"adjustments"`
	NetChange           int64 `json:"net_change"`
	InternalTransfers   int64 `json:"internal_transfers"`
	InitialBalanceAdded int64 `json:"initial_balance_added"`
}

type SavingsGoalItem struct {
	SavingsMetrics
	WalletID         string        `json:"wallet_id"`
	Name             string        `json:"name"`
	TargetAmount     *int64        `json:"target_amount"`
	TargetDate       *CalendarDate `json:"target_date"`
	Remaining        int64         `json:"remaining"`
	ProgressBPS      int64         `json:"progress_bps"`
	Reached          bool          `json:"reached"`
	TransactionCount int           `json:"transaction_count"`
}

type SavingsGoals struct {
	SavingsMetrics
	Items []SavingsGoalItem `json:"items"`
}

type SavingsSurplus struct {
	Scope            string `json:"scope"`
	Income           int64  `json:"income"`
	Expense          int64  `json:"expense"`
	Net              int64  `json:"net"`
	SavingsRateBPS   *int64 `json:"savings_rate_bps"`
	TransactionCount int    `json:"transaction_count"`
}

type SavingsReport struct {
	Month    string         `json:"month"`
	Timezone string         `json:"timezone"`
	AsOf     time.Time      `json:"as_of"`
	Goals    SavingsGoals   `json:"goals"`
	Surplus  SavingsSurplus `json:"surplus"`
}

type SavingsTransactionPage struct {
	Items    []Transaction `json:"items"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
	Month    string        `json:"month"`
	Timezone string        `json:"timezone"`
	AsOf     time.Time     `json:"as_of"`
}

type savingsScope struct {
	month           string
	start, end, now time.Time
	location        *time.Location
	wallets         map[string]Wallet
	goals           map[string]Wallet
}

func newSavingsScope(owner string, wallets []Wallet, month, walletID string, now time.Time, location *time.Location) (savingsScope, error) {
	if owner == "" || now.IsZero() {
		return savingsScope{}, ErrSavingsInvalidQuery
	}
	if location == nil {
		location = time.UTC
	}
	if month == "" {
		month = now.In(location).Format("2006-01")
	}
	parsed, err := time.ParseInLocation("2006-01", month, location)
	if err != nil || parsed.Year() < 1 || parsed.Format("2006-01") != month {
		return savingsScope{}, ErrSavingsInvalidQuery
	}
	if parsed.After(now) {
		return savingsScope{}, ErrSavingsFutureMonth
	}
	s := savingsScope{month: month, start: parsed, end: parsed.AddDate(0, 1, 0), now: now, location: location, wallets: map[string]Wallet{}, goals: map[string]Wallet{}}
	for _, w := range wallets {
		if w.OwnerID != owner {
			continue
		}
		s.wallets[w.ID] = w
		if w.Type == WalletTypeGoal && (walletID == "" || walletID == w.ID) {
			s.goals[w.ID] = w
		}
	}
	if walletID != "" {
		if _, ok := s.goals[walletID]; !ok {
			return savingsScope{}, ErrSavingsWalletNotFound
		}
	}
	return s, nil
}

func (s savingsScope) includes(owner string, tx Transaction) bool {
	_, owned := s.wallets[tx.WalletID]
	return owned && tx.OwnerID == owner && !tx.OccurredAt.IsZero() && tx.OccurredAt.Before(s.end) && !tx.OccurredAt.After(s.now)
}

// savingsGoalEffect intentionally ignores IncludedInReports: balances must
// reconcile transfers and adjustments as well as excluded ordinary entries.
func savingsGoalEffect(tx Transaction) (kind string, signed int64, err error) {
	if tx.Amount < 0 {
		return "", 0, ErrSavingsInvalidLedger
	}
	switch tx.Type {
	case TransactionTypeIncome:
		kind, signed = "deposit", tx.Amount
	case TransactionTypeExpense:
		kind, signed = "withdrawal", -tx.Amount
	case TransactionTypeAdjustment:
		kind = "adjustment"
		if tx.AdjustmentDirection == nil {
			return "", 0, ErrSavingsInvalidLedger
		}
		switch *tx.AdjustmentDirection {
		case AdjustmentDirectionIncrease:
			signed = tx.Amount
		case AdjustmentDirectionDecrease:
			signed = -tx.Amount
		default:
			return "", 0, ErrSavingsInvalidLedger
		}
	default:
		return "", 0, ErrSavingsInvalidLedger
	}
	if tx.TransferID != nil {
		kind = "transfer"
	}
	return kind, signed, nil
}

func savingsSurplusKind(tx Transaction) string {
	if !tx.IncludedInReports || tx.TransferID != nil || tx.CreditPaymentID != nil || (tx.CreditKind != nil && *tx.CreditKind == CreditKindPayment) {
		return ""
	}
	if tx.Type == TransactionTypeIncome || tx.Type == TransactionTypeExpense {
		return tx.Type
	}
	return ""
}

func savingsAdd(target *int64, value int64) error {
	if (value > 0 && *target > math.MaxInt64-value) || (value < 0 && *target < math.MinInt64-value) {
		return ErrSavingsOverflow
	}
	*target += value
	return nil
}

func savingsDifference(a, b int64) (int64, error) {
	v := new(big.Int).Sub(big.NewInt(a), big.NewInt(b))
	if !v.IsInt64() {
		return 0, ErrSavingsOverflow
	}
	return v.Int64(), nil
}

func savingsRatio(numerator, denominator int64) (int64, error) {
	v := new(big.Int).Mul(big.NewInt(numerator), big.NewInt(10000))
	v.Quo(v, big.NewInt(denominator))
	if !v.IsInt64() {
		return 0, ErrSavingsOverflow
	}
	return v.Int64(), nil
}

// CalculateSavingsReport derives both sections from the same owner/calendar
// scope used by SavingsReportTransactions. No persisted wallet balance is read.
func CalculateSavingsReport(owner string, wallets []Wallet, rows []Transaction, month, walletID string, now time.Time, location *time.Location) (SavingsReport, error) {
	s, err := newSavingsScope(owner, wallets, month, walletID, now, location)
	if err != nil {
		return SavingsReport{}, err
	}
	r := SavingsReport{Month: s.month, Timezone: s.location.String(), AsOf: now.UTC(), Goals: SavingsGoals{Items: []SavingsGoalItem{}}, Surplus: SavingsSurplus{Scope: "account"}}
	items := map[string]*SavingsGoalItem{}
	for id, w := range s.goals {
		metrics := SavingsMetrics{}
		// The baseline is configuration, not a fabricated ledger transaction.
		// Legacy data without creation metadata retains the historical baseline.
		if w.CreatedAt.IsZero() || w.CreatedAt.Before(s.start) {
			metrics.OpeningBalance, metrics.ClosingBalance = w.OpeningBalance, w.OpeningBalance
		} else if w.CreatedAt.Before(s.end) && !w.CreatedAt.After(s.now) {
			metrics.ClosingBalance, metrics.InitialBalanceAdded = w.OpeningBalance, w.OpeningBalance
		}
		items[id] = &SavingsGoalItem{WalletID: id, Name: w.Name, TargetAmount: w.TargetAmount, TargetDate: w.TargetDate, SavingsMetrics: metrics}
	}
	for _, tx := range rows {
		if !s.includes(owner, tx) {
			continue
		}
		inMonth := !tx.OccurredAt.Before(s.start)
		if item := items[tx.WalletID]; item != nil {
			kind, signed, effectErr := savingsGoalEffect(tx)
			if effectErr != nil {
				return SavingsReport{}, effectErr
			}
			if err = savingsAdd(&item.ClosingBalance, signed); err != nil {
				return SavingsReport{}, err
			}
			if !inMonth {
				if err = savingsAdd(&item.OpeningBalance, signed); err != nil {
					return SavingsReport{}, err
				}
			} else {
				item.TransactionCount++
				var field *int64
				switch kind {
				case "deposit":
					field = &item.Deposits
				case "withdrawal":
					field = &item.Withdrawals
					signed = tx.Amount
				case "adjustment":
					field = &item.Adjustments
				case "transfer":
					field = &item.InternalTransfers
				}
				if err = savingsAdd(field, signed); err != nil {
					return SavingsReport{}, err
				}
			}
		}
		if !inMonth {
			continue
		}
		kind := savingsSurplusKind(tx)
		if kind == "" {
			continue
		}
		if tx.Amount < 0 {
			return SavingsReport{}, ErrSavingsInvalidLedger
		}
		r.Surplus.TransactionCount++
		field := &r.Surplus.Income
		if kind == TransactionTypeExpense {
			field = &r.Surplus.Expense
		}
		if err = savingsAdd(field, tx.Amount); err != nil {
			return SavingsReport{}, err
		}
	}
	for _, item := range items {
		item.NetChange, err = savingsDifference(item.ClosingBalance, item.OpeningBalance)
		if err != nil {
			return SavingsReport{}, err
		}
		if item.TargetAmount != nil && *item.TargetAmount > 0 {
			item.Reached = item.ClosingBalance >= *item.TargetAmount
			if !item.Reached {
				item.Remaining, err = savingsDifference(*item.TargetAmount, item.ClosingBalance)
				if err != nil {
					return SavingsReport{}, err
				}
			}
			if item.ClosingBalance > 0 {
				item.ProgressBPS, err = savingsRatio(item.ClosingBalance, *item.TargetAmount)
				if err != nil {
					return SavingsReport{}, err
				}
			}
		}
		for _, pair := range [][2]*int64{{&r.Goals.OpeningBalance, &item.OpeningBalance}, {&r.Goals.ClosingBalance, &item.ClosingBalance}, {&r.Goals.Deposits, &item.Deposits}, {&r.Goals.Withdrawals, &item.Withdrawals}, {&r.Goals.Adjustments, &item.Adjustments}, {&r.Goals.NetChange, &item.NetChange}, {&r.Goals.InternalTransfers, &item.InternalTransfers}, {&r.Goals.InitialBalanceAdded, &item.InitialBalanceAdded}} {
			if err = savingsAdd(pair[0], *pair[1]); err != nil {
				return SavingsReport{}, err
			}
		}
		r.Goals.Items = append(r.Goals.Items, *item)
	}
	sort.Slice(r.Goals.Items, func(i, j int) bool {
		a, b := r.Goals.Items[i], r.Goals.Items[j]
		if a.Name == b.Name {
			return a.WalletID < b.WalletID
		}
		return a.Name < b.Name
	})
	r.Surplus.Net, err = savingsDifference(r.Surplus.Income, r.Surplus.Expense)
	if err != nil {
		return SavingsReport{}, err
	}
	if r.Surplus.Income > 0 {
		rate, ratioErr := savingsRatio(r.Surplus.Net, r.Surplus.Income)
		if ratioErr != nil {
			return SavingsReport{}, ratioErr
		}
		r.Surplus.SavingsRateBPS = &rate
	}
	return r, nil
}

func SavingsReportTransactions(owner string, wallets []Wallet, rows []Transaction, month, walletID, section, kind string, limit, offset int, now time.Time, location *time.Location) (SavingsTransactionPage, error) {
	if kind == "" {
		kind = "all"
	}
	validKind := kind == "all"
	switch section {
	case "goals":
		validKind = validKind || kind == "deposit" || kind == "withdrawal" || kind == "adjustment" || kind == "transfer"
	case "surplus":
		validKind = walletID == "" && (validKind || kind == "income" || kind == "expense")
	default:
		validKind = false
	}
	if !validKind || limit < 1 || limit > 100 || offset < 0 {
		return SavingsTransactionPage{}, ErrSavingsInvalidQuery
	}
	s, err := newSavingsScope(owner, wallets, month, walletID, now, location)
	if err != nil {
		return SavingsTransactionPage{}, err
	}
	r := SavingsTransactionPage{Items: []Transaction{}, Limit: limit, Offset: offset, Month: s.month, Timezone: s.location.String(), AsOf: now.UTC()}
	matched := []Transaction{}
	for _, tx := range rows {
		if !s.includes(owner, tx) || tx.OccurredAt.Before(s.start) {
			continue
		}
		rowKind := ""
		if section == "goals" {
			if _, ok := s.goals[tx.WalletID]; !ok {
				continue
			}
			rowKind, _, err = savingsGoalEffect(tx)
			if err != nil {
				return SavingsTransactionPage{}, err
			}
		} else {
			rowKind = savingsSurplusKind(tx)
			if rowKind != "" && tx.Amount < 0 {
				return SavingsTransactionPage{}, ErrSavingsInvalidLedger
			}
		}
		if rowKind != "" && (kind == "all" || kind == rowKind) {
			matched = append(matched, tx)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].OccurredAt.Equal(matched[j].OccurredAt) {
			return matched[i].ID > matched[j].ID
		}
		return matched[i].OccurredAt.After(matched[j].OccurredAt)
	})
	r.Total = len(matched)
	if offset < r.Total {
		end := offset + min(limit, r.Total-offset)
		r.Items = matched[offset:end]
	}
	return r, nil
}
