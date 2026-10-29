package push

import (
	"context"
	"errors"
	"testing"
	"time"
)

type scriptedTransport struct {
	responses []Response
	requests  []Request
}

func (s *scriptedTransport) Send(_ context.Context, request Request) (Response, error) {
	s.requests = append(s.requests, request)
	if len(s.responses) == 0 {
		return Response{}, errors.New("transport script exhausted")
	}
	response := s.responses[0]
	s.responses = s.responses[1:]
	return response, nil
}

func TestDeliverSentOnSuccessAndKeepsPayloadPrivateToTransport(t *testing.T) {
	transport := &scriptedTransport{responses: []Response{{StatusCode: 200}}}
	result, err := Deliver(context.Background(), transport, Request{
		NotificationID: "notice-1", TokenID: "token-1", DeviceToken: "secret-token",
		CollapseID: "notice-1", Payload: []byte(`{"aps":{"alert":{"title":"Budget"}}}`),
	}, Attempt{Number: 1}, time.Unix(100, 0))
	if err != nil || result.State != StateSent || result.RevokeToken {
		t.Fatalf("expected sent result, got result=%+v err=%v", result, err)
	}
	if len(transport.requests) != 1 || string(transport.requests[0].Payload) == "" {
		t.Fatal("transport must receive the push request")
	}
}

func TestDeliverGoneRevokesOnlyMatchingTokenWithoutRetry(t *testing.T) {
	transport := &scriptedTransport{responses: []Response{{StatusCode: 410}}}
	result, err := Deliver(context.Background(), transport, Request{NotificationID: "notice-1", TokenID: "token-1", DeviceToken: "secret-token"}, Attempt{Number: 3}, time.Unix(100, 0))
	if err != nil || result.State != StateRevokeToken || !result.RevokeToken || result.RetryAt != nil {
		t.Fatalf("expected revoke result, got result=%+v err=%v", result, err)
	}
}

func TestDeliverRetryHonorsRetryAfterAndBoundsAttempts(t *testing.T) {
	now := time.Unix(100, 0)
	transport := &scriptedTransport{responses: []Response{{StatusCode: 429, RetryAfter: 7 * time.Second}}}
	result, err := Deliver(context.Background(), transport, Request{NotificationID: "notice-1", TokenID: "token-1", DeviceToken: "secret-token"}, Attempt{Number: 1}, now)
	if err != nil || result.State != StateRetry || result.RetryAt == nil || !result.RetryAt.Equal(now.Add(7*time.Second)) {
		t.Fatalf("expected retry-after result, got result=%+v err=%v", result, err)
	}

	transport.responses = []Response{{StatusCode: 503}}
	result, err = Deliver(context.Background(), transport, Request{NotificationID: "notice-1", TokenID: "token-1", DeviceToken: "secret-token"}, Attempt{Number: maxAttempts}, now)
	if err != nil || result.State != StateFailed || result.RetryAt != nil {
		t.Fatalf("expected bounded permanent failure, got result=%+v err=%v", result, err)
	}
}

func TestDeliverRejectsInvalidCredentialsOrTopicWithoutRetry(t *testing.T) {
	for _, status := range []int{400, 403} {
		transport := &scriptedTransport{responses: []Response{{StatusCode: status}}}
		result, err := Deliver(context.Background(), transport, Request{NotificationID: "notice-1", TokenID: "token-1", DeviceToken: "secret-token"}, Attempt{Number: 1}, time.Unix(100, 0))
		if err != nil || result.State != StateFailed || result.RetryAt != nil {
			t.Fatalf("status %d must fail without retry: result=%+v err=%v", status, result, err)
		}
	}
}

func TestValidateConfigRequiresTopicAndCredentialsOnlyWhenEnabled(t *testing.T) {
	if err := (Config{}).Validate(); err != nil {
		t.Fatalf("disabled APNs must be valid: %v", err)
	}
	invalid := Config{Enabled: true, TeamID: "team", KeyID: "key", Topic: "com.example.app"}
	if err := invalid.Validate(); err == nil {
		t.Fatal("enabled APNs without key material must be rejected")
	}
	valid := Config{Enabled: true, TeamID: "team", KeyID: "key", Topic: "com.example.app", PrivateKey: "redacted-key", Environment: "sandbox"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid APNs config rejected: %v", err)
	}
}
