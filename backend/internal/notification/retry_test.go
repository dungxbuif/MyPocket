package notification

import (
	"errors"
	"testing"
)

func TestRetryClassificationAndRedaction(t *testing.T) {
	if !IsExpiredDeliveryError(errors.New("push returned 410 Gone")) {
		t.Fatal("expected expired delivery")
	}
	if IsExpiredDeliveryError(errors.New("temporary timeout")) {
		t.Fatal("timeout must be retryable")
	}
	if got := RedactEndpoint("https://push.example/subscription/abcdef1234567890"); got == "https://push.example/subscription/abcdef1234567890" || got == "" {
		t.Fatalf("endpoint was not redacted: %q", got)
	}
}
