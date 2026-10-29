---
artifact_type: detail_design
id: FEEDBACK-20261019-AUDIT-DELIVERY
status: implemented_local
owner: shared
updated: 2026-10-19
---

# Redis access-audit delivery contract

Trace: [remaining work map](FEEDBACK-20261019-REMAINING.md),
[advisor architecture](../../architecture/AI_ADVISOR_TOOLS.md),
[validation matrix](../VALIDATION_MATRIX.md),
[operations runbook](../../operations/AUDIT_OUTBOX_RUNBOOK.md).

## Scope

Harden the existing privacy-safe Redis audit sink without changing financial
semantics, API responses, owner authorization, or native behavior. The sink
must make a repeated `Record` call for the same logical event idempotent and
must keep bounded stream retention. The event contains only the existing
redacted metadata; prompts, OCR, notes, amounts, tokens and credentials are
not added.

## Contract

- The logical event fingerprint includes stream, request ID, actor, credential
  ID, action, decision/status metadata, latency, feedback/changelog IDs and
  version. The generated timestamp is deliberately excluded.
- Repeating the same event within 24 hours produces one Redis stream entry.
  A distinct request or action remains a distinct event.
- Each audit stream is trimmed approximately at 10,000 entries. This is a
  bounded safety limit, not a compliance archive or a time-based retention
  guarantee.
- The Redis client retries a failed command at most three times with bounded
  backoff. A final failure remains best-effort at the request boundary; the
  caller does not fail a financial or authentication request because audit
  delivery is unavailable.
- API requests now persist the same redacted payload in `audit_outbox` before
  attempting the Redis fast path. A successful fast-path write is safe to
  replay because the existing event fingerprint deduplicates it.
- The worker claims at most 100 rows with `FOR UPDATE SKIP LOCKED` and a
  two-minute lease, publishes to Redis, marks delivered, or schedules a
  bounded exponential retry (up to five minutes). A Redis outage does not
  fail the financial/auth request when the outbox row was persisted.
- The outbox is a delivery buffer, not a compliance archive; retention and
  dead-letter alerting remain an operational rollout gate.

## Implementation and verification

`backend/internal/infrastructure/cache/audit.go` uses an atomic Lua script to
claim the fingerprint key and append to the selected stream. If `XADD` fails,
the claim is removed so a later retry can deliver the event. The Redis client
has an explicit bounded retry policy in `redis.go`.

`backend/internal/infrastructure/repository/audit_outbox_postgres.go` uses an
idempotent `(stream,event_id)` key and leased `SKIP LOCKED` claims. The worker
implementation is in `backend/cmd/worker/audit.go`.

The regression test runs against `TEST_REDIS_URL` and calls `Record` twice with
the same event, asserting exactly one stream entry. The test also retains the
existing stream-selection and metadata-redaction checks. A disposable Redis
fixture is required; no production Redis or database is used.

The PostgreSQL regression covers enqueue idempotency, lease exclusion,
failure retry and terminal delivery on a disposable database after migration
`000037_audit_outbox`.

## Out of scope / follow-up gate

Production rollout still requires migration backup/version checks, worker
health monitoring, a retention/dead-letter alert policy and deployment/rollback
evidence. Until those gates are completed, this is durable local delivery
buffering, not a claim of production audit compliance.
