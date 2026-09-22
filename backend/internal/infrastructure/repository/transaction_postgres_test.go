package repository

import (
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

func TestValidateJarAssignmentPreservesUnchangedHistoricalLink(t *testing.T) {
	instant := time.Date(2026, time.February, 1, 0, 30, 0, 0, time.UTC)
	jarID := "jar-stable-id"
	existing := &entity.Transaction{JarID: &jarID, OccurredAt: instant}
	// A timezone change can re-bucket the same UTC instant into a month with no
	// config. Saving unrelated fields must not need a database month lookup.
	if err := validateJarAssignment(nil, "owner", &jarID, instant, existing); err != nil {
		t.Fatalf("unchanged stable-jar link should be preserved without a new-month config: %v", err)
	}
}
