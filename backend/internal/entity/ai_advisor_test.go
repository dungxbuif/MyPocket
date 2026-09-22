package entity

import "testing"

func TestAdvisorMessageRejectsOversizedUserPart(t *testing.T) {
	message := AdvisorMessage{Role: AdvisorRoleUser, Parts: []AdvisorPart{{Type: "text", Text: string(make([]byte, 8193))}}}
	if err := message.Validate(); err == nil {
		t.Fatal("oversized user message must be rejected")
	}
}

func TestAdvisorRunAllowsOnlyKnownTransitions(t *testing.T) {
	if !CanTransitionAdvisorRun(AdvisorStatusQueued, AdvisorStatusRunning) {
		t.Fatal("queued run should be claimable")
	}
	if CanTransitionAdvisorRun(AdvisorStatusCompleted, AdvisorStatusRunning) {
		t.Fatal("terminal run must not be restarted")
	}
}

func TestAdvisorMessageRejectsOversizedPartData(t *testing.T) {
	message := AdvisorMessage{Role: AdvisorRoleAssistant, Parts: []AdvisorPart{{Type: "metric_group", Data: string(make([]byte, 64*1024+1))}}}
	if err := message.Validate(); err == nil {
		t.Fatal("oversized structured part must be rejected")
	}
}
