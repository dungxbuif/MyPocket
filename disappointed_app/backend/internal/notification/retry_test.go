package notification

import (
	"errors"
	"net/http"
	"testing"
)

func TestRetryClassificationAndRedaction(t *testing.T) {
	if !IsExpiredDeliveryError(&DeliveryError{StatusCode: http.StatusGone}) {
		t.Fatal("expected expired delivery")
	}
	if IsExpiredDeliveryError(errors.New("untyped error containing 410 must not delete a subscription")) {
		t.Fatal("only typed provider statuses may expire a subscription")
	}
	if IsExpiredDeliveryError(errors.New("temporary timeout")) {
		t.Fatal("timeout must be retryable")
	}
	if got := RedactEndpoint("https://push.example/subscription/abcdef1234567890"); got == "https://push.example/subscription/abcdef1234567890" || got == "" {
		t.Fatalf("endpoint was not redacted: %q", got)
	}
}
