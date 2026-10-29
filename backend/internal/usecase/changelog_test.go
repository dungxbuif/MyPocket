package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

func TestChangelogPublishFinalizesFeedbackAndAuditsOnlyIdentifiers(t *testing.T) {
	feedback := &feedbackRepoStub{rows: map[string]entity.Feedback{"fb-1": {ID: "fb-1", UserID: "owner", Type: entity.FeedbackTypeBug, Title: "Private title", Description: "Private financial detail", Status: entity.FeedbackStatusInProgress}}}
	audit := &auditStub{}
	changelogs := &changelogRepoStub{feedback: feedback, rows: map[string]entity.Changelog{}}
	service := NewChangelogService(feedback, changelogs, audit, func() time.Time { return time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC) })
	row, err := service.Publish(context.Background(), ChangelogInput{FeedbackIDs: []string{"fb-1"}, Version: "1.4.2", Title: "Categorization fix", Description: "Fixed the issue."})
	if err != nil {
		t.Fatal(err)
	}
	if row.Version != "1.4.2" || feedback.rows["fb-1"].Status != entity.FeedbackStatusFixed || feedback.rows["fb-1"].ChangelogID == nil {
		t.Fatalf("finalization mismatch: row=%+v feedback=%+v", row, feedback.rows["fb-1"])
	}
	if len(audit.events) != 1 || audit.events[0].Action != "changelog.published" || audit.events[0].ChangelogID != row.ID {
		t.Fatalf("audit mismatch: %+v", audit.events)
	}
	if strings.Contains(strings.Join([]string{audit.events[0].FeedbackID, audit.events[0].Version}, " "), "Private") {
		t.Fatal("audit must not carry feedback content")
	}
}
