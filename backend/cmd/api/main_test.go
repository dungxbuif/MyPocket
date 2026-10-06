package main

import "testing"

func TestOptionalAttachmentStoresNormalizeUnconfiguredStorage(t *testing.T) {
	aiStorage, feedbackStorage := optionalAttachmentStores(nil)
	if aiStorage != nil || feedbackStorage != nil {
		t.Fatalf("unconfigured storage must stay nil in both interfaces: ai=%T feedback=%T", aiStorage, feedbackStorage)
	}
}
