package agent

import (
	"context"
	"database/sql"
	"errors"
)

type Store interface {
	CreateRun(context.Context, string, string, Kind, string) (Run, error)
	CreateRunWithTool(context.Context, string, string, Kind, string, string) (Run, error)
	GetRun(context.Context, string, string) (Run, error)
	CreateToolRun(context.Context, string, string, string) (ToolRun, error)
}

type Service struct{ Store Store }

func (s Service) Submit(ctx context.Context, userID, key string, kind Kind, text, receiptID string) (Run, error) {
	if err := ValidateCreate(kind, text, key); err != nil {
		return Run{}, err
	}
	var run Run
	var err error
	if receiptID != "" {
		run, err = s.Store.CreateRunWithTool(ctx, userID, key, kind, text, receiptID)
	} else {
		run, err = s.Store.CreateRun(ctx, userID, key, kind, text)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return Run{}, ErrConflict
	}
	return run, err
}

func (s Service) Get(ctx context.Context, userID, id string) (Run, error) {
	return s.Store.GetRun(ctx, userID, id)
}
