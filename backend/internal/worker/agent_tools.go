package worker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"mypocket/internal/agent"
)

const maxOCRImageBytes int64 = 15 << 20

type AgentToolRepository interface {
	ClaimToolDue(context.Context, string, time.Time, time.Duration) (agent.ToolRun, bool, error)
	MarkToolSubmitted(context.Context, agent.ToolRun, agent.ToolSubmission, time.Time) error
	CompleteTool(context.Context, agent.ToolRun, agent.ToolResult) error
	UpdateTool(context.Context, agent.ToolRun, agent.ToolResult, time.Time) error
}
type AgentToolObjectStore interface {
	GetObject(context.Context, string, int64) (io.ReadCloser, int64, error)
}

type AgentToolRunner struct {
	Repo                            AgentToolRepository
	Store                           AgentToolObjectStore
	Tool                            agent.ImageTool
	Owner                           string
	PollInterval, MaxProcessingTime time.Duration
	Now                             func() time.Time
}

func (r AgentToolRunner) RunOnce(ctx context.Context) (bool, error) {
	if r.Repo == nil || r.Store == nil || r.Tool == nil {
		return false, nil
	}
	now := time.Now().UTC()
	if r.Now != nil {
		now = r.Now().UTC()
	}
	poll := r.PollInterval
	if poll <= 0 {
		poll = 5 * time.Second
	}
	maxTime := r.MaxProcessingTime
	if maxTime <= 0 {
		maxTime = 10 * time.Minute
	}
	run, ok, err := r.Repo.ClaimToolDue(ctx, r.Owner, now, time.Minute)
	if err != nil || !ok {
		return ok, err
	}
	if now.Sub(run.CreatedAt) > maxTime {
		return true, r.Repo.UpdateTool(ctx, run, agent.ToolResult{Status: agent.ToolExpired, ErrorCode: "PROCESSING_TIMEOUT"}, now)
	}
	if run.ProviderDocumentID == "" {
		body, size, err := r.Store.GetObject(ctx, run.ObjectKey, maxOCRImageBytes)
		if err != nil {
			return true, r.fail(ctx, run, "IMAGE_UNAVAILABLE", now)
		}
		defer body.Close()
		data, err := io.ReadAll(io.LimitReader(body, maxOCRImageBytes+1))
		if err != nil || int64(len(data)) > maxOCRImageBytes || int64(len(data)) != size {
			return true, r.fail(ctx, run, "IMAGE_INVALID", now)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != run.ChecksumSHA256 {
			return true, r.fail(ctx, run, "CHECKSUM_MISMATCH", now)
		}
		sub, err := r.Tool.Submit(ctx, agent.ImageInput{ContentType: run.ContentType, Bytes: data, Languages: []string{"vi", "en"}})
		if err != nil {
			return true, r.retry(ctx, run, "OCR_UNAVAILABLE", now, poll)
		}
		run.ProviderDocumentID = sub.ProviderID
		if sub.Status != agent.ToolCompleted {
			return true, r.Repo.MarkToolSubmitted(ctx, run, sub, now.Add(maxDuration(sub.RetryAfter, poll)))
		}
	}
	result, err := r.Tool.Read(ctx, run.ProviderDocumentID)
	if err != nil {
		if errors.Is(err, agent.ErrToolExpired) {
			return true, r.Repo.UpdateTool(ctx, run, agent.ToolResult{Status: agent.ToolExpired, ErrorCode: "PROVIDER_EXPIRED"}, now)
		}
		if errors.Is(err, agent.ErrToolNotFound) {
			return true, r.fail(ctx, run, "PROVIDER_NOT_FOUND", now)
		}
		return true, r.retry(ctx, run, "OCR_UNAVAILABLE", now, poll)
	}
	switch result.Status {
	case agent.ToolCompleted:
		return true, r.Repo.CompleteTool(ctx, run, result)
	case agent.ToolProcessing:
		return true, r.Repo.UpdateTool(ctx, run, result, now.Add(maxDuration(result.RetryAfter, poll)))
	case agent.ToolFailed, agent.ToolCancelled, agent.ToolExpired:
		return true, r.Repo.UpdateTool(ctx, run, result, now)
	default:
		return true, errors.New("unsupported OCR tool status")
	}
}
func (r AgentToolRunner) fail(ctx context.Context, run agent.ToolRun, code string, now time.Time) error {
	return r.Repo.UpdateTool(ctx, run, agent.ToolResult{Status: agent.ToolFailed, ErrorCode: code}, now)
}
func (r AgentToolRunner) retry(ctx context.Context, run agent.ToolRun, code string, now time.Time, poll time.Duration) error {
	if run.Attempts >= 3 {
		return r.fail(ctx, run, code, now)
	}
	return r.Repo.UpdateTool(ctx, run, agent.ToolResult{Status: agent.ToolProcessing, ErrorCode: code}, now.Add(poll))
}
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
