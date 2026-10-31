package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/push"
	"github.com/mypocket/backend/internal/repository"
)

type pushDeliveryRepoFake struct {
	enqueueLimit int
	claimLimit   int
	claimLease   time.Duration
	claimEnv     string
	jobs         []repository.PushDeliveryJob
	sent         []string
	retries      []pushDeliveryMark
	revoked      []string
	failed       []pushDeliveryMark
}

type pushDeliveryMark struct {
	id         string
	leaseUntil time.Time
	available  time.Time
	reason     string
}

func (f *pushDeliveryRepoFake) EnqueueNotificationDeliveries(_ context.Context, _ time.Time, limit int) (int, error) {
	f.enqueueLimit = limit
	return 1, nil
}

func (f *pushDeliveryRepoFake) ClaimPushDeliveries(_ context.Context, _ time.Time, limit int, lease time.Duration, environment string) ([]repository.PushDeliveryJob, error) {
	f.claimLimit, f.claimLease, f.claimEnv = limit, lease, environment
	return f.jobs, nil
}

func (f *pushDeliveryRepoFake) MarkPushDeliverySent(_ context.Context, id string, leaseUntil, _ time.Time) error {
	f.sent = append(f.sent, id)
	if leaseUntil.IsZero() {
		return errors.New("missing lease")
	}
	return nil
}

func (f *pushDeliveryRepoFake) MarkPushDeliveryRetry(_ context.Context, id string, leaseUntil, availableAt time.Time, reason string) error {
	f.retries = append(f.retries, pushDeliveryMark{id: id, leaseUntil: leaseUntil, available: availableAt, reason: reason})
	return nil
}

func (f *pushDeliveryRepoFake) MarkPushDeliveryRevoked(_ context.Context, id string, leaseUntil, _ time.Time) error {
	f.revoked = append(f.revoked, id)
	return nil
}

func (f *pushDeliveryRepoFake) MarkPushDeliveryFailed(_ context.Context, id string, leaseUntil, _ time.Time, reason string) error {
	f.failed = append(f.failed, pushDeliveryMark{id: id, leaseUntil: leaseUntil, reason: reason})
	return nil
}

type pushTokenDeliveryRepoFake struct {
	token        entity.PushToken
	revokeCalls  []string
	rotateOnSend func()
}

func (f *pushTokenDeliveryRepoFake) LoadForDelivery(_ context.Context, owner, id string) (entity.PushToken, error) {
	if f.token.OwnerID != owner || f.token.ID != id {
		return entity.PushToken{}, repository.ErrNotFound
	}
	return f.token, nil
}

func (f *pushTokenDeliveryRepoFake) RevokeIfCurrent(_ context.Context, owner, id, token string, _ time.Time) (bool, error) {
	if f.token.OwnerID == owner && f.token.ID == id && f.token.Token == token {
		f.revokeCalls = append(f.revokeCalls, token)
		return true, nil
	}
	return false, nil
}

type pushTransportFake struct {
	response     push.Response
	err          error
	requests     []push.Request
	beforeReturn func()
}

func (f *pushTransportFake) Send(_ context.Context, request push.Request) (push.Response, error) {
	f.requests = append(f.requests, request)
	if f.beforeReturn != nil {
		f.beforeReturn()
	}
	return f.response, f.err
}

func TestRunPushDeliveryPassHandoffsAndSendsSuccess(t *testing.T) {
	now := time.Date(2026, 11, 1, 3, 4, 5, 0, time.UTC)
	deliveryRepo := &pushDeliveryRepoFake{jobs: []repository.PushDeliveryJob{{
		Delivery:   entity.PushDelivery{ID: "delivery-1", OwnerID: "owner-1", NotificationID: "notice-1", PushTokenID: "token-1", Environment: "production", CollapseID: "notice-1", Payload: []byte(`{"aps":{}}`), Attempts: 1},
		LeaseUntil: now.Add(2 * time.Minute),
	}}}
	tokenRepo := &pushTokenDeliveryRepoFake{token: entity.PushToken{ID: "token-1", OwnerID: "owner-1", Token: "device-token", Environment: "production"}}
	transport := &pushTransportFake{response: push.Response{StatusCode: 200}}

	count, err := runPushDeliveryPass(context.Background(), deliveryRepo, tokenRepo, transport, "production", now)
	if err != nil || count != 1 {
		t.Fatalf("push pass count=%d err=%v", count, err)
	}
	if deliveryRepo.enqueueLimit != 100 || deliveryRepo.claimLimit != 100 || deliveryRepo.claimLease != 2*time.Minute || deliveryRepo.claimEnv != "production" {
		t.Fatalf("unexpected queue contract: %+v", deliveryRepo)
	}
	if len(transport.requests) != 1 || transport.requests[0].DeviceToken != "device-token" || string(transport.requests[0].Payload) != `{"aps":{}}` {
		t.Fatalf("unexpected APNs request: %+v", transport.requests)
	}
	if len(deliveryRepo.sent) != 1 || deliveryRepo.sent[0] != "delivery-1" {
		t.Fatalf("delivery was not marked sent: %+v", deliveryRepo.sent)
	}
}

func TestRunPushDeliveryPass410RevokesMatchingTokenAndDelivery(t *testing.T) {
	now := time.Date(2026, 11, 1, 3, 4, 5, 0, time.UTC)
	deliveryRepo := pushDeliveryRepoForTest(now, 1, "sandbox")
	tokenRepo := &pushTokenDeliveryRepoFake{token: entity.PushToken{ID: "token-1", OwnerID: "owner-1", Token: "old-token", Environment: "sandbox"}}
	transport := &pushTransportFake{response: push.Response{StatusCode: 410}}

	count, err := runPushDeliveryPass(context.Background(), deliveryRepo, tokenRepo, transport, "sandbox", now)
	if err != nil || count != 1 || len(tokenRepo.revokeCalls) != 1 || tokenRepo.revokeCalls[0] != "old-token" || len(deliveryRepo.revoked) != 1 {
		t.Fatalf("410 handling count=%d err=%v revoked=%v delivery=%v", count, err, tokenRepo.revokeCalls, deliveryRepo.revoked)
	}
}

func TestRunPushDeliveryPassDoesNotRevokeTokenRotatedAfterSend(t *testing.T) {
	now := time.Date(2026, 11, 1, 3, 4, 5, 0, time.UTC)
	deliveryRepo := pushDeliveryRepoForTest(now, 1, "sandbox")
	tokenRepo := &pushTokenDeliveryRepoFake{token: entity.PushToken{ID: "token-1", OwnerID: "owner-1", Token: "old-token", Environment: "sandbox"}}
	transport := &pushTransportFake{response: push.Response{StatusCode: 410}, beforeReturn: func() {
		tokenRepo.token.Token = "new-token"
	}}

	count, err := runPushDeliveryPass(context.Background(), deliveryRepo, tokenRepo, transport, "sandbox", now)
	if err != nil || count != 1 || len(tokenRepo.revokeCalls) != 0 || len(deliveryRepo.revoked) != 1 {
		t.Fatalf("rotated token was revoked count=%d err=%v revoke=%v delivery=%v", count, err, tokenRepo.revokeCalls, deliveryRepo.revoked)
	}
}

func TestRunPushDeliveryPassRetriesTransientAndFailsPermanentWithoutRawProviderData(t *testing.T) {
	now := time.Date(2026, 11, 1, 3, 4, 5, 0, time.UTC)
	for _, test := range []struct {
		name       string
		attempts   int
		response   push.Response
		wantRetry  bool
		wantFailed bool
		wantReason string
	}{
		{name: "retry-after", attempts: 1, response: push.Response{StatusCode: 429, RetryAfter: 7 * time.Second}, wantRetry: true, wantReason: "apns_http_429"},
		{name: "retry-limit", attempts: 5, response: push.Response{StatusCode: 503}, wantFailed: true, wantReason: "apns_http_503"},
		{name: "bad-request", attempts: 1, response: push.Response{StatusCode: 400, FailureCode: "http_400"}, wantFailed: true, wantReason: "http_400"},
		{name: "forbidden", attempts: 1, response: push.Response{StatusCode: 403, FailureCode: "http_403"}, wantFailed: true, wantReason: "http_403"},
	} {
		t.Run(test.name, func(t *testing.T) {
			deliveryRepo := pushDeliveryRepoForTest(now, test.attempts)
			tokenRepo := &pushTokenDeliveryRepoFake{token: entity.PushToken{ID: "token-1", OwnerID: "owner-1", Token: "device-token", Environment: "production"}}
			transport := &pushTransportFake{response: test.response}
			if _, err := runPushDeliveryPass(context.Background(), deliveryRepo, tokenRepo, transport, "production", now); err != nil {
				t.Fatal(err)
			}
			if test.wantRetry != (len(deliveryRepo.retries) == 1) || test.wantFailed != (len(deliveryRepo.failed) == 1) {
				t.Fatalf("unexpected terminal state retries=%v failed=%v", deliveryRepo.retries, deliveryRepo.failed)
			}
			if test.wantRetry && !deliveryRepo.retries[0].available.Equal(now.Add(7*time.Second)) {
				t.Fatalf("retry-after was not preserved: %+v", deliveryRepo.retries[0])
			}
			gotReason := ""
			if len(deliveryRepo.retries) == 1 {
				gotReason = deliveryRepo.retries[0].reason
			}
			if len(deliveryRepo.failed) == 1 {
				gotReason = deliveryRepo.failed[0].reason
			}
			if gotReason != test.wantReason || strings.Contains(gotReason, "device-token") {
				t.Fatalf("unexpected safe reason: %q", gotReason)
			}
		})
	}
}

func TestRunPushDeliveryPassRevokesAlreadyRevokedTokenWithoutSending(t *testing.T) {
	now := time.Date(2026, 11, 1, 3, 4, 5, 0, time.UTC)
	deliveryRepo := pushDeliveryRepoForTest(now, 1)
	revokedAt := now.Add(-time.Minute)
	tokenRepo := &pushTokenDeliveryRepoFake{token: entity.PushToken{ID: "token-1", OwnerID: "owner-1", Token: "device-token", Environment: "production", RevokedAt: &revokedAt}}
	transport := &pushTransportFake{response: push.Response{StatusCode: 200}}
	if count, err := runPushDeliveryPass(context.Background(), deliveryRepo, tokenRepo, transport, "production", now); err != nil || count != 1 {
		t.Fatalf("revoked token count=%d err=%v", count, err)
	}
	if len(transport.requests) != 0 || len(deliveryRepo.revoked) != 1 {
		t.Fatalf("revoked token was sent or not marked: requests=%v revoked=%v", transport.requests, deliveryRepo.revoked)
	}
}

func pushDeliveryRepoForTest(now time.Time, attempts int, environments ...string) *pushDeliveryRepoFake {
	environment := "production"
	if len(environments) > 0 && environments[0] != "" {
		environment = environments[0]
	}
	return &pushDeliveryRepoFake{jobs: []repository.PushDeliveryJob{{
		Delivery:   entity.PushDelivery{ID: "delivery-1", OwnerID: "owner-1", NotificationID: "notice-1", PushTokenID: "token-1", Environment: environment, CollapseID: "notice-1", Payload: []byte(`{"aps":{}}`), Attempts: attempts},
		LeaseUntil: now.Add(2 * time.Minute),
	}}}
}
