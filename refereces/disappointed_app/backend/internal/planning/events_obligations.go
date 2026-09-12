package planning

import (
	"fmt"
	"time"
)

func ValidateCreateEvent(input CreateEventInput) (CreateEventInput, error) {
	input.Name = trimmed(input.Name)
	input.Note = trimmed(input.Note)
	if input.Name == "" {
		return CreateEventInput{}, fmt.Errorf("%w: event name is required", ErrValidation)
	}
	start, err := parseDateValue(input.StartsOn)
	if err != nil {
		return CreateEventInput{}, fmt.Errorf("%w: event start date is invalid", ErrValidation)
	}
	endValue := input.EndsOn
	if trimmed(endValue) == "" {
		endValue = input.StartsOn
	}
	end, err := parseDateValue(endValue)
	if err != nil {
		return CreateEventInput{}, fmt.Errorf("%w: event end date is invalid", ErrValidation)
	}
	if end.Before(start) {
		return CreateEventInput{}, fmt.Errorf("%w: event end must be after start", ErrValidation)
	}
	input.StartsOn = start.Format("2006-01-02")
	input.EndsOn = end.Format("2006-01-02")
	return input, nil
}

func ValidateUpdateEvent(input UpdateEventInput) (UpdateEventInput, error) {
	if input.BaseVersion <= 0 {
		return UpdateEventInput{}, fmt.Errorf("%w: base version is required", ErrValidation)
	}
	created, err := ValidateCreateEvent(CreateEventInput{Name: input.Name, StartsOn: input.StartsOn, EndsOn: input.EndsOn, Note: input.Note})
	if err != nil {
		return UpdateEventInput{}, err
	}
	return UpdateEventInput{BaseVersion: input.BaseVersion, Name: created.Name, StartsOn: created.StartsOn, EndsOn: created.EndsOn, Note: created.Note}, nil
}

func ValidateCreateObligation(input CreateObligationInput) (CreateObligationInput, error) {
	input.Counterparty = trimmed(input.Counterparty)
	input.Note = trimmed(input.Note)
	if !validObligationDirection(input.Direction) {
		return CreateObligationInput{}, fmt.Errorf("%w: obligation direction is invalid", ErrValidation)
	}
	if input.PrincipalVND <= 0 {
		return CreateObligationInput{}, fmt.Errorf("%w: obligation principal must be positive", ErrValidation)
	}
	if input.Counterparty == "" {
		return CreateObligationInput{}, fmt.Errorf("%w: obligation counterparty is required", ErrValidation)
	}
	due, err := parseDateValue(input.DueOn)
	if err != nil {
		return CreateObligationInput{}, fmt.Errorf("%w: obligation due date is invalid", ErrValidation)
	}
	input.DueOn = due.Format("2006-01-02")
	return input, nil
}

func ValidateUpdateObligation(input UpdateObligationInput) (UpdateObligationInput, error) {
	if input.BaseVersion <= 0 {
		return UpdateObligationInput{}, fmt.Errorf("%w: base version is required", ErrValidation)
	}
	created, err := ValidateCreateObligation(CreateObligationInput{Direction: input.Direction, PrincipalVND: input.PrincipalVND, Counterparty: input.Counterparty, DueOn: input.DueOn, Note: input.Note})
	if err != nil {
		return UpdateObligationInput{}, err
	}
	return UpdateObligationInput{BaseVersion: input.BaseVersion, Direction: created.Direction, PrincipalVND: created.PrincipalVND, Counterparty: created.Counterparty, DueOn: created.DueOn, Note: created.Note}, nil
}

func validObligationDirection(value ObligationDirection) bool {
	switch value {
	case ObligationBorrowed, ObligationLent:
		return true
	default:
		return false
	}
}

func parseDateValue(value string) (time.Time, error) {
	return time.Parse("2006-01-02", trimmed(value))
}
