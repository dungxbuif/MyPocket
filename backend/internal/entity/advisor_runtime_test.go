package entity

import (
	"testing"
	"time"
)

func TestDefaultAdvisorRuntimePolicyMatchesReliabilityContract(t *testing.T) {
	policy := DefaultAdvisorRuntimePolicy()
	if policy.RunDeadline != 240*time.Second || policy.LeaseDuration != 45*time.Second || policy.HeartbeatEvery != 10*time.Second || policy.ReaperEvery != 15*time.Second || policy.DBOperation != 2*time.Second {
		t.Fatalf("unexpected policy: %+v", policy)
	}
}
