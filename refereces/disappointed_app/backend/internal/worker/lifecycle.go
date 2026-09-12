package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"mypocket/internal/lifecycle"
)

type LifecycleObjectStore interface {
	PutObject(context.Context, string, string, io.Reader, int64) error
	GetObject(context.Context, string, int64) (io.ReadCloser, int64, error)
	DeleteObject(context.Context, string) error
	DeletePrefix(context.Context, string) error
}

type LifecycleRunner struct {
	Repo      *lifecycle.Repository
	Store     LifecycleObjectStore
	Formatter lifecycle.ExportFormatter
}

func (r LifecycleRunner) RunOnce(ctx context.Context) (int, error) {
	job, err := r.Repo.ClaimDue(ctx)
	if err == lifecycle.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if r.Formatter == nil {
		r.Formatter = lifecycle.CSVExportFormatter{}
	}
	if err = r.process(ctx, job); err != nil {
		_ = r.Repo.Retry(ctx, job, "JOB_FAILED")
		return 0, err
	}
	return 1, nil
}

func (r LifecycleRunner) process(ctx context.Context, job lifecycle.Job) error {
	switch job.Kind {
	case lifecycle.KindExport:
		var request lifecycle.ExportRequest
		if err := json.Unmarshal(job.Request, &request); err != nil {
			return lifecycle.ErrInvalid
		}
		snapshot, err := r.Repo.ExportSnapshot(ctx, job.UserID, request)
		if err != nil {
			return err
		}
		var out bytes.Buffer
		if err = r.Formatter.Write(ctx, snapshot, &out); err != nil {
			return err
		}
		key := fmt.Sprintf("users/%s/exports/%s.csv", job.UserID, job.ID)
		if r.Store == nil {
			return fmt.Errorf("lifecycle object store unavailable")
		}
		if err = r.Store.PutObject(ctx, key, "text/csv; charset=utf-8", bytes.NewReader(out.Bytes()), int64(out.Len())); err != nil {
			return err
		}
		return r.Repo.Complete(ctx, job, map[string]any{"rows": len(snapshot.Rows)}, key)
	case lifecycle.KindImport:
		if len(job.Result) > 0 {
			count, err := r.Repo.ApplyImport(ctx, job)
			if err != nil {
				return err
			}
			return r.Repo.Complete(ctx, job, map[string]any{"imported_rows": count}, "")
		}
		var request lifecycle.ImportRequest
		if err := json.Unmarshal(job.Request, &request); err != nil {
			return lifecycle.ErrInvalid
		}
		prefix := "users/" + job.UserID + "/imports/"
		if !strings.HasPrefix(request.ObjectKey, prefix) {
			return lifecycle.ErrInvalid
		}
		if r.Store == nil {
			return fmt.Errorf("lifecycle object store unavailable")
		}
		body, _, err := r.Store.GetObject(ctx, request.ObjectKey, lifecycle.MaxImportBytes)
		if err != nil {
			return err
		}
		defer body.Close()
		wallets, categories, err := r.Repo.OwnedReferences(ctx, job.UserID)
		if err != nil {
			return err
		}
		preview := (lifecycle.ImportParser{}).Parse(job.ID, job.Version+1, body, wallets, categories)
		return r.Repo.AwaitConfirmation(ctx, job, preview)
	case lifecycle.KindReset:
		if r.Store != nil {
			if err := r.Store.DeletePrefix(ctx, "users/"+job.UserID+"/"); err != nil {
				return err
			}
		}
		if err := r.Repo.ResetUserData(ctx, job.UserID); err != nil {
			return err
		}
		return r.Repo.Complete(ctx, job, map[string]any{"reset": true}, "")
	case lifecycle.KindDelete:
		// The API disables access before this cleanup job is queued.
		if r.Store != nil {
			if err := r.Store.DeletePrefix(ctx, "users/"+job.UserID+"/"); err != nil {
				return err
			}
		}
		if err := r.Repo.ResetUserData(ctx, job.UserID); err != nil {
			return err
		}
		return r.Repo.DeleteUserData(ctx, job.UserID)
	default:
		return lifecycle.ErrInvalid
	}
}
