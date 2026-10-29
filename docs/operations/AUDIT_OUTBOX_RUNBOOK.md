---
artifact_type: runbook
id: AUDIT-OUTBOX-RUNBOOK
status: ready
owner: operations
updated: 2026-10-20
---

# Audit outbox runbook

This runbook covers the PostgreSQL `audit_outbox` delivery buffer introduced by
migration `000037`. It contains only redacted access/feedback metadata. Do not
select or print `payload` in production diagnostics. The outbox is a delivery
buffer, not a compliance archive; production retention and dead-letter alert
thresholds still require an operational owner decision.

## Read-only health check

Run against the production database through the approved read-only path. The
queries below expose counts and bounded error metadata, never payload contents:

```sql
SELECT
  count(*) FILTER (WHERE delivered_at IS NULL) AS pending,
  count(*) FILTER (WHERE delivered_at IS NOT NULL) AS delivered,
  count(*) FILTER (WHERE delivered_at IS NULL AND last_error IS NOT NULL) AS failed_pending,
  max(attempts) FILTER (WHERE delivered_at IS NULL) AS max_pending_attempts,
  max(created_at) FILTER (WHERE delivered_at IS NULL) AS newest_pending
FROM audit_outbox;

SELECT stream, count(*) AS pending, max(attempts) AS max_attempts,
       min(available_at) AS next_available
FROM audit_outbox
WHERE delivered_at IS NULL
GROUP BY stream
ORDER BY stream;

SELECT id, stream, event_id, attempts, available_at, lease_until,
       left(last_error, 160) AS bounded_error
FROM audit_outbox
WHERE delivered_at IS NULL AND last_error IS NOT NULL
ORDER BY available_at ASC, created_at ASC
LIMIT 50;
```

Healthy steady state is an empty or short-lived pending queue, no lease that
remains past two worker intervals, and retry timestamps moving forward. A
non-zero queue is not itself permission to delete rows or force a migration.

## Delivery probe

The worker owns claiming and marking rows. Run one bounded pass only in a
controlled window after confirming the matching API/worker image and schema
version:

```sh
cd backend
WORKER_ONCE=true go run ./cmd/worker
```

Inspect worker logs for the `audit outbox worker pass` result, then repeat the
read-only queries. A row is marked delivered only after Redis `XADD` succeeds;
the Redis event fingerprint prevents a response-loss retry from duplicating
the stream event. Do not call repository update/delete SQL manually.

## Failure handling and stop conditions

- Redis outage: leave rows pending; the worker retries with a bounded
  exponential delay capped at five minutes.
- Worker restart: an expired two-minute lease makes an uncompleted row
  claimable again; no payload is lost by process restart.
- Persistent failures: preserve `last_error`, capture the bounded error and
  escalate using the deployment incident path. Do not expose payloads or
  credentials in logs.
- Retention/dead-letter: no automatic purge or dead-letter transition is
  enabled by this runbook. Obtain an approved retention period and alert
  destination before adding cleanup or declaring compliance retention.

## Evidence required for rollout

Record the migration version/dirty flag, two worker intervals of healthy logs,
before/after pending counts, and a disposable Redis outage/recovery test. The
local contract and integration evidence are tracked in
[the audit delivery detail design](../work/tickets/FEEDBACK-20261019-AUDIT-DELIVERY-DETAIL_DESIGN.md)
and [the validation matrix](../work/VALIDATION_MATRIX.md).
