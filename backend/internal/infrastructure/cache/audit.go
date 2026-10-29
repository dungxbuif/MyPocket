package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

const feedbackAuditStream = "mypocket:audit:feedback"
const advisorAuditStream = "mypocket:audit:advisor"

const (
	auditStreamMaxLen = int64(10000)
	auditDedupeTTL    = 24 * time.Hour
)

var auditRecordScript = redis.NewScript(`
local claimed = redis.call('SET', KEYS[1], '1', 'NX', 'EX', ARGV[1])
if not claimed then
  return 0
end
local ok, result = pcall(redis.call, 'XADD', KEYS[2], 'MAXLEN', '~', ARGV[2], '*', unpack(ARGV, 3))
if not ok then
  redis.call('DEL', KEYS[1])
  error(result)
end
return 1
`)

func auditFields(event repository.AuditEvent) map[string]interface{} {
	fields := map[string]interface{}{
		"request_id":     event.RequestID,
		"actor_kind":     event.ActorKind,
		"credential_id":  event.CredentialID,
		"action":         event.Action,
		"allowed":        boolFlag(event.Allowed),
		"status":         event.Status,
		"reason":         event.Reason,
		"latency_ms":     strconv.FormatInt(event.LatencyMS, 10),
		"request_bytes":  strconv.FormatInt(max(event.RequestBytes, 0), 10),
		"response_bytes": strconv.FormatInt(max(event.ResponseBytes, 0), 10),
		"feedback_id":    event.FeedbackID,
		"feedback_ids":   strings.Join(event.FeedbackIDs, ","),
		"changelog_id":   event.ChangelogID,
		"old_status":     event.OldStatus,
		"new_status":     event.NewStatus,
		"version":        event.Version,
		"recorded_at":    strconv.FormatInt(time.Now().UTC().Unix(), 10),
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
	fields := auditFields(event)
	eventID := auditEventID(stream, event)
	fields["event_id"] = eventID
	if r.outbox != nil {
		payload, marshalErr := json.Marshal(fields)
		if marshalErr == nil {
			// Persist before attempting the fast path. If Redis is unavailable,
			// the worker can retry from this row without losing the audit event.
			if err := r.outbox.EnqueueAudit(ctx, stream, eventID, payload, time.Now().UTC()); err == nil {
				if err := r.recordFields(ctx, stream, eventID, fields); err == nil {
				}
				return nil
			}
		}
	}
	return r.recordFields(ctx, stream, eventID, fields)
}

func (r *Redis) recordFields(ctx context.Context, stream, eventID string, fields map[string]interface{}) error {
	// Older persisted outbox payloads predate byte-count fields. Supply scalar
	// defaults so replay remains compatible without a migration or nil Redis args.
	for _, key := range []string{"request_bytes", "response_bytes"} {
		if _, ok := fields[key]; !ok {
			fields[key] = "0"
		}
	}
	values := make([]interface{}, 0, len(fields)*2)
	// Stable field ordering makes this command inspectable and avoids changing
	// the event fingerprint when the map iteration order changes.
	for _, key := range []string{
		"request_id", "actor_kind", "credential_id", "action", "allowed", "status", "reason",
		"latency_ms", "request_bytes", "response_bytes", "feedback_id", "feedback_ids", "changelog_id", "old_status", "new_status",
		"version", "recorded_at", "event_id",
	} {
		values = append(values, key, fields[key])
	}
	dedupeKey := "mypocket:audit:dedupe:" + stream + ":" + eventID
	args := make([]interface{}, 0, len(values)+2)
	args = append(args, int(auditDedupeTTL/time.Second), auditStreamMaxLen)
	args = append(args, values...)
	_, err := auditRecordScript.Run(ctx, r.client, []string{dedupeKey, stream}, args...).Result()
	return err
}

// PublishAudit is used only by the durable audit worker. The payload is the
// already-redacted JSON object persisted in audit_outbox.
func (r *Redis) PublishAudit(ctx context.Context, stream, eventID string, payload []byte) error {
	var fields map[string]interface{}
	if err := json.Unmarshal(payload, &fields); err != nil {
		return fmt.Errorf("decode audit payload: %w", err)
	}
	fields["event_id"] = eventID
	return r.recordFields(ctx, stream, eventID, fields)
}

func auditEventID(stream string, event repository.AuditEvent) string {
	feedbackIDs := append([]string(nil), event.FeedbackIDs...)
	// Feedback service already sorts IDs, but sorting here makes the sink
	// idempotent even when a caller constructs an equivalent event differently.
	// Do not include recorded_at: a retry is the same audit event.
	sort.Strings(feedbackIDs)
	data, _ := json.Marshal(struct {
		Stream        string   `json:"stream"`
		RequestID     string   `json:"request_id"`
		ActorKind     string   `json:"actor_kind"`
		CredentialID  string   `json:"credential_id"`
		Action        string   `json:"action"`
		Allowed       bool     `json:"allowed"`
		Status        string   `json:"status"`
		Reason        string   `json:"reason"`
		LatencyMS     int64    `json:"latency_ms"`
		RequestBytes  int64    `json:"request_bytes,omitempty"`
		ResponseBytes int64    `json:"response_bytes,omitempty"`
		FeedbackIDs   []string `json:"feedback_ids"`
		FeedbackID    string   `json:"feedback_id"`
		ChangelogID   string   `json:"changelog_id"`
		OldStatus     string   `json:"old_status"`
		NewStatus     string   `json:"new_status"`
		Version       string   `json:"version"`
	}{
		Stream: stream, RequestID: event.RequestID, ActorKind: event.ActorKind, CredentialID: event.CredentialID,
		Action: event.Action, Allowed: event.Allowed, Status: event.Status, Reason: event.Reason,
		LatencyMS: event.LatencyMS, RequestBytes: event.RequestBytes, ResponseBytes: event.ResponseBytes, FeedbackIDs: feedbackIDs, FeedbackID: event.FeedbackID,
		ChangelogID: event.ChangelogID, OldStatus: event.OldStatus, NewStatus: event.NewStatus,
		Version: event.Version,
	})
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func boolFlag(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
