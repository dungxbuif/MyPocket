package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

type reportRetentionCall struct {
	name  string
	now   time.Time
	limit int
}

type reportRetentionDeleteStub struct {
	name    string
	deleted int
	err     error
	calls   *[]reportRetentionCall
}

func (s *reportRetentionDeleteStub) DeleteExpired(_ context.Context, now time.Time, limit int) (int, error) {
	*s.calls = append(*s.calls, reportRetentionCall{name: s.name, now: now, limit: limit})
	return s.deleted, s.err
}

func TestReportRetentionDeletesAnalysesBeforeSnapshots(t *testing.T) {
	now := time.Date(2026, 11, 1, 2, 0, 0, 0, time.UTC)
	calls := make([]reportRetentionCall, 0, 2)
	analyses := &reportRetentionDeleteStub{name: "analyses", deleted: 2, calls: &calls}
	snapshots := &reportRetentionDeleteStub{name: "snapshots", deleted: 3, calls: &calls}

	deleted, err := runReportRetentionPass(context.Background(), analyses, snapshots, now)
	if err != nil {
		t.Fatalf("retention pass failed: %v", err)
	}
	if deleted != 5 {
		t.Fatalf("deleted count = %d, want 5", deleted)
	}
	if len(calls) != 2 || calls[0].name != "analyses" || calls[1].name != "snapshots" {
		t.Fatalf("retention order = %+v, want analyses then snapshots", calls)
	}
	for _, call := range calls {
		if !call.now.Equal(now) || call.limit != 100 {
			t.Fatalf("retention call = %+v, want now=%s limit=100", call, now)
		}
	}
}

func TestReportRetentionStopsBeforeSnapshotsWhenAnalysisDeleteFails(t *testing.T) {
	wantErr := errors.New("analysis store unavailable")
	calls := make([]reportRetentionCall, 0, 2)
	analyses := &reportRetentionDeleteStub{name: "analyses", deleted: 1, err: wantErr, calls: &calls}
	snapshots := &reportRetentionDeleteStub{name: "snapshots", deleted: 9, calls: &calls}

	deleted, err := runReportRetentionPass(context.Background(), analyses, snapshots, time.Now())
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if deleted != 1 {
		t.Fatalf("deleted count = %d, want analysis count 1", deleted)
	}
	if len(calls) != 1 || calls[0].name != "analyses" {
		t.Fatalf("retention should not visit snapshots after analysis failure: %+v", calls)
	}
}
