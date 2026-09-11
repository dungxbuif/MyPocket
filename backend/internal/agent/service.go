package agent

import (
	"context"
	"database/sql"
	"errors"
)

type Store interface {
	CreateRun(context.Context, string, string, Kind, string) (Run, error)
	GetRun(context.Context, string, string) (Run, error)
}

type Service struct{ Store Store }

func (s Service) Submit(ctx context.Context, userID, key string, kind Kind, text string) (Run, error) {
	if err := ValidateCreate(kind, text, key); err != nil {
		return Run{}, err
	}
	run, err := s.Store.CreateRun(ctx, userID, key, kind, text)
	if errors.Is(err, sql.ErrNoRows) {
		return Run{}, ErrConflict
	}
	return run, err
}

func (s Service) Get(ctx context.Context, userID, id string) (Run, error) {
	return s.Store.GetRun(ctx, userID, id)
}
