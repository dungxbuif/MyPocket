package httpapi_test

import (
	"testing"

	"mypocket/internal/platform/httpapi"
)

func TestErrorEnvelopeIncludesStableCodeAndCorrelationID(t *testing.T) {
	body := httpapi.ErrorEnvelope("AUTH_REQUIRED", "Authentication required", "req_123")

	if body.CorrelationID != "req_123" {
		t.Fatalf("missing correlation id: %#v", body)
	}
	if body.Error == nil || body.Error.Code != "AUTH_REQUIRED" {
		t.Fatalf("missing stable error code: %#v", body)
	}
}

func TestErrorEnvelopeDoesNotExposeUnsafeMessage(t *testing.T) {
	body := httpapi.ErrorEnvelope("INTERNAL_FAILURE", "pq: password=secret stack trace", "req_456")

	if body.Error == nil {
		t.Fatal("missing error body")
	}
	if body.Error.Message == "pq: password=secret stack trace" {
		t.Fatalf("unsafe internal detail leaked: %#v", body)
	}
}
