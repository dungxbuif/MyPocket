package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

const feedbackAuditStream = "mypocket:audit:feedback"
const advisorAuditStream = "mypocket:audit:advisor"

func auditFields(event repository.AuditEvent) map[string]interface{} {
	fields := map[string]interface{}{
		"request_id":    event.RequestID,
		"actor_kind":    event.ActorKind,
		"credential_id": event.CredentialID,
		"action":        event.Action,
		"allowed":       boolFlag(event.Allowed),
		"status":        event.Status,
		"reason":        event.Reason,
		"latency_ms":    strconv.FormatInt(event.LatencyMS, 10),
		"feedback_id":   event.FeedbackID,
		"feedback_ids":  strings.Join(event.FeedbackIDs, ","),
		"changelog_id":  event.ChangelogID,
		"old_status":    event.OldStatus,
		"new_status":    event.NewStatus,
		"version":       event.Version,
		"recorded_at":   strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}
	return fields
}

func (r *Redis) Record(ctx context.Context, event repository.AuditEvent) error {
	if r == nil || r.client == nil {
		return nil
	}
	stream := feedbackAuditStream
	if strings.HasPrefix(event.Action, "advisor.") {
		stream = advisorAuditStream
	}
	return r.client.XAdd(ctx, &redis.XAddArgs{Stream: stream, MaxLen: 10000, Approx: true, Values: auditFields(event)}).Err()
}

func boolFlag(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
