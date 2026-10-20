---
artifact_type: detail_design
id: PHASE-007-DETAIL-DESIGN
status: approved
owner: shared
approval: approved
approved_on: 2026-08-30
trace:
  backlog_item: BL-007
  phase: PHASE-007
  requirements: [REQ-F-013, REQ-F-014, REQ-F-015, REQ-NF-004, REQ-NF-006, REQ-NF-007]
  tickets: [TICKET-022, TICKET-023, TICKET-024, TICKET-025]
  validation_matrix: ../VALIDATION_MATRIX.md
  master_docs_touched: [../../architecture/API.md, ../../architecture/ERD.md, ../../operations/DEPLOYMENT.md, ../../architecture/ARCHITECTURE.md]
---

# DETAIL DESIGN: PHASE-007 Audit, Export, Account Lifecycle, and Production Operations

## 1. Context and Boundary

The final release phase makes sensitive operations observable, exports portable, account lifecycle safe, and homelab deployment recoverable.

- In scope: append-only audit events, exact-email hidden viewer, retention purge, CSV/Sheets-compatible exports, reset/delete jobs, production config, backup/restore, restart, rollback, and release proof.
- Out of scope: general admin roles, click analytics, spreadsheet synchronization, and device management.
- Approval: user-delegated decisions on 2026-08-30. Small task exemption: no.

## 2. Audit Design

`audit_events` stores immutable ID, actor user, action, target type/ID, redacted before/after summary, correlation ID, source IP hash where allowed, and timestamp. Application DB role receives insert/select but no update/delete; retention uses a separate worker capability.

- Audit state/security mutations, auth outcomes, provider jobs, webhook validation outcomes, exports, reset/delete lifecycle, and configuration-sensitive operations.
- Do not audit ordinary page views, filters, search queries, secret values, raw provider payloads, notes, image bytes, tokens, or credentials.
- `/api/v1/audit/events` and hidden `/internal/audit` require authenticated verified email exactly equal to normalized `AUDIT_VIEWER_EMAIL`; the route is absent from navigation and backend authorization is authoritative.
- Retention defaults to 180 days and purges bounded batches without generating recursive per-row events.

## 3. Export and Lifecycle Jobs

- `POST /api/v1/exports` creates an immutable user-scoped snapshot job with selected datasets/date range; worker writes UTF-8 CSV files and optional ZIP to private S3.
- Export schema has stable headers, ISO timestamps, integer VND, no secrets, and a Google Sheets-compatible CSV dialect. Downloads use short-lived presigned URLs.
- Reset/delete requires recent authentication, typed confirmation, CSRF, idempotency, and a preview of affected counts.
- Reset removes finance/planning/provider data but preserves the identity account and required lifecycle audit record.
- Delete disables access immediately, queues user data and S3 deletion, retains only legally/operationally required tombstone evidence, then removes identity data according to documented recovery limits.
- Jobs are resumable, scoped by user ID, batch S3 deletes, and record phase progress plus terminal evidence.

## 4. Production Operations

- Validate required env at startup: public URL, database, cookie security, Google OAuth, S3, provider settings, VAPID, audit viewer, and retention.
- Health split: liveness for process, readiness for PostgreSQL and required startup migrations; external providers do not block core readiness.
- Release order: backup, migrate, start API/worker/frontend, readiness, smoke, then expose traffic.
- Backup includes PostgreSQL and S3 inventory; restore proof uses an isolated database/bucket and verifies row/object checksums.
- Worker leases survive restart; rollback never runs destructive down migrations and documents forward-fix rules.

## 5. API and UI

- Audit API supports bounded cursor pagination and action/date/correlation filters; never free-text searches sensitive payloads.
- Account screen includes export, reset, and delete flows with job progress, failure recovery, and precise Vietnamese confirmation copy.
- Hidden audit page is dense and read-only with timestamp, action, actor, target, correlation ID, and redacted diff detail.
- Operations docs include Compose production profile, reverse-proxy/TLS assumptions, migrations, backup, restore, restart, rollback, and incident checks.

## 6. Security and Failure Policy

- Exact backend authorization for audit; hidden routing is not a security boundary.
- Presigned export URLs expire quickly and remain user-scoped.
- Destructive jobs fail closed on ambiguous scope and never use broad object-key prefixes without validated user ownership.
- Backups are required before production schema/destructive release operations; credentials never enter artifacts or logs.
- Rate-limit audit queries, exports, and lifecycle commands.

## 7. Verification and Release Gate

- Unit: redaction, action mapping, CSV formatting, deletion plan, env validation, and retention batching.
- Integration: audit immutability/auth, export isolation, reset/delete exact scope, idempotent job resume, and S3 cleanup fixtures.
- E2E/security: viewer allow/deny, hidden navigation, export/download, reset/delete confirmation and cancellation.
- Platform: production-like Compose, migration, health, API/worker restart, PostgreSQL backup/restore, S3 inventory recovery, rollback drill, and release smoke.
- Reconcile API, ERD, deployment, architecture, security docs, validation matrix, context, backlog, changelog, and create ADRs for any operational divergence.
