package worker

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"
	"time"

	"mypocket/internal/agent"
)

type toolRepoStub struct {
	run       agent.ToolRun
	completed int
	updated   agent.ToolResult
	submitted agent.ToolSubmission
}

func (s *toolRepoStub) ClaimToolDue(context.Context, string, time.Time, time.Duration) (agent.ToolRun, bool, error) {
	return s.run, true, nil
}
func (s *toolRepoStub) MarkToolSubmitted(_ context.Context, _ agent.ToolRun, v agent.ToolSubmission, _ time.Time) error {
	s.submitted = v
	return nil
}
func (s *toolRepoStub) CompleteTool(_ context.Context, _ agent.ToolRun, v agent.ToolResult) error {
	s.completed++
	return nil
}
func (s *toolRepoStub) UpdateTool(_ context.Context, _ agent.ToolRun, v agent.ToolResult, _ time.Time) error {
	s.updated = v
	return nil
}

type toolStoreStub struct{ data []byte }

func (s toolStoreStub) GetObject(context.Context, string, int64) (io.ReadCloser, int64, error) {
	return io.NopCloser(bytes.NewReader(s.data)), int64(len(s.data)), nil
}

type imageToolStub struct {
	sub     agent.ToolSubmission
	result  agent.ToolResult
	submits int
}

func (s *imageToolStub) Submit(context.Context, agent.ImageInput) (agent.ToolSubmission, error) {
	s.submits++
	return s.sub, nil
}
func (s *imageToolStub) Read(context.Context, string) (agent.ToolResult, error) { return s.result, nil }

func TestAgentToolRunnerVerifiesPrivateBytesThenStoresResult(t *testing.T) {
	data := []byte("receipt")
	sum := sha256.Sum256(data)
	repo := &toolRepoStub{run: agent.ToolRun{ID: "t", UserID: "u", CreatedAt: time.Now(), ObjectKey: "users/u/receipt", ContentType: "image/jpeg", ChecksumSHA256: hex.EncodeToString(sum[:])}}
	tool := &imageToolStub{sub: agent.ToolSubmission{ProviderID: "doc", Status: agent.ToolCompleted}, result: agent.ToolResult{Status: agent.ToolCompleted, Text: "120000"}}
	processed, err := (AgentToolRunner{Repo: repo, Store: toolStoreStub{data}, Tool: tool}).RunOnce(context.Background())
	if err != nil || !processed || repo.completed != 1 || tool.submits != 1 {
		t.Fatalf("processed=%v completed=%d submits=%d err=%v", processed, repo.completed, tool.submits, err)
	}
}
func TestAgentToolRunnerRejectsChecksumBeforeSubmission(t *testing.T) {
	repo := &toolRepoStub{run: agent.ToolRun{CreatedAt: time.Now(), ChecksumSHA256: "bad"}}
	tool := &imageToolStub{}
	_, _ = (AgentToolRunner{Repo: repo, Store: toolStoreStub{[]byte("receipt")}, Tool: tool}).RunOnce(context.Background())
	if tool.submits != 0 || repo.updated.ErrorCode != "CHECKSUM_MISMATCH" {
		t.Fatalf("unsafe submit=%d result=%#v", tool.submits, repo.updated)
	}
}
