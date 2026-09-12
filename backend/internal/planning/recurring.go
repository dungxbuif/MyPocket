package planning

import (
	"fmt"
	"time"

	"mypocket/internal/finance"
)

func ValidateCreateRecurringSchedule(input CreateRecurringScheduleInput) (CreateRecurringScheduleInput, time.Time, error) {
	input, startsAt, _, err := validateRecurringScheduleFields(input)
	return input, startsAt, err
}

func ValidateUpdateRecurringSchedule(input UpdateRecurringScheduleInput) (CreateRecurringScheduleInput, time.Time, *time.Time, error) {
	if input.BaseVersion <= 0 {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: base version is required", ErrValidation)
	}
	return validateRecurringScheduleFields(CreateRecurringScheduleInput{
		Name:                input.Name,
		Frequency:           input.Frequency,
		Timezone:            input.Timezone,
		StartsAt:            input.StartsAt,
		EndsAt:              input.EndsAt,
		PostingMode:         input.PostingMode,
		Type:                input.Type,
		SourceWalletID:      input.SourceWalletID,
		DestinationWalletID: input.DestinationWalletID,
		CategoryID:          input.CategoryID,
		BudgetID:            input.BudgetID,
		AmountVND:           input.AmountVND,
		Note:                input.Note,
	})
}

func validateRecurringScheduleFields(input CreateRecurringScheduleInput) (CreateRecurringScheduleInput, time.Time, *time.Time, error) {
	input.Name = trimmed(input.Name)
	input.Timezone = trimmed(input.Timezone)
	input.Note = trimmed(input.Note)
	if input.Name == "" {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: schedule name is required", ErrValidation)
	}
	if !validRecurrenceFrequency(input.Frequency) {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: recurrence frequency is invalid", ErrValidation)
	}
	if input.Timezone == "" {
		input.Timezone = "Asia/Ho_Chi_Minh"
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: recurrence timezone is invalid", ErrValidation)
	}
	startsAt, err := time.Parse(time.RFC3339, trimmed(input.StartsAt))
	if err != nil {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: schedule start is invalid", ErrValidation)
	}
	var endsAt *time.Time
	if trimmed(input.EndsAt) != "" {
		parsed, err := time.Parse(time.RFC3339, trimmed(input.EndsAt))
		if err != nil {
			return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: schedule end is invalid", ErrValidation)
		}
		if parsed.Before(startsAt) {
			return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: schedule end must be after start", ErrValidation)
		}
		endsAt = &parsed
	}
	if input.PostingMode == "" {
		input.PostingMode = RecurringPostingDraft
	}
	if input.PostingMode != RecurringPostingDraft && input.PostingMode != RecurringPostingAutoPost {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: recurring posting mode is invalid", ErrValidation)
	}
	if !validRecurringTransactionType(input.Type) {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: schedule transaction type is invalid", ErrValidation)
	}
	input.SourceWalletID = trimmed(input.SourceWalletID)
	input.DestinationWalletID = trimmed(input.DestinationWalletID)
	input.CategoryID = trimmed(input.CategoryID)
	input.BudgetID = trimmed(input.BudgetID)
	if input.SourceWalletID == "" {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: source wallet is required", ErrValidation)
	}
	if input.BudgetID != "" && input.Type != finance.TransactionExpense {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: only expense schedules can have a budget", ErrValidation)
	}
	if input.Type == finance.TransactionTransfer {
		if input.DestinationWalletID == "" {
			return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: destination wallet is required", ErrValidation)
		}
		if input.DestinationWalletID == input.SourceWalletID {
			return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: transfer wallets must differ", ErrValidation)
		}
		if input.CategoryID != "" {
			return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: transfer cannot have category", ErrValidation)
		}
	} else {
		if input.DestinationWalletID != "" {
			return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: income and expense cannot have destination wallet", ErrValidation)
		}
		if input.CategoryID == "" {
			return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: category is required", ErrValidation)
		}
	}
	if input.AmountVND <= 0 {
		return CreateRecurringScheduleInput{}, time.Time{}, nil, fmt.Errorf("%w: schedule amount must be positive", ErrValidation)
	}
	return input, startsAt, endsAt, nil
}

func validRecurrenceFrequency(value RecurrenceFrequency) bool {
	switch value {
	case RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly:
		return true
	default:
		return false
	}
}

func validRecurringTransactionType(value finance.TransactionType) bool {
	switch value {
	case finance.TransactionIncome, finance.TransactionExpense, finance.TransactionTransfer:
		return true
	default:
		return false
	}
}

func NextOccurrence(previous time.Time, frequency RecurrenceFrequency, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = mustHoChiMinh()
	}
	local := previous.In(loc)
	switch frequency {
	case RecurrenceDaily:
		return local.AddDate(0, 0, 1).UTC()
	case RecurrenceWeekly:
		return local.AddDate(0, 0, 7).UTC()
	case RecurrenceMonthly:
		return local.AddDate(0, 1, 0).UTC()
	default:
		return local.UTC()
	}
}
