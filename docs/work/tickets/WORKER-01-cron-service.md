---
artifact_type: ticket
id: WORKER-01
status: draft
owner: human
priority: TBD
lane: high-risk
human_fields:
  - priority
  - approval
  - month_report_semantics
ai_fields:
  - impacted_areas
  - test_expectations
  - next_artifact
shared_fields:
  - status
  - trace
trace:
  backlog_item: ../BACKLOG.md
  recurring_transactions: TICKET-05-02-giao-dich-den-ky.md
  monthly_summary: TICKET-07-04-tong-ket-thang-ai.md
  current_month_design: CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# WORKER-01 — Dịch vụ worker chạy cronjob

## Context and intent

The owner requested a dedicated backend worker process sharing MyPocket's configuration, database, repositories, and use cases. Named examples are generating due recurring transactions and a month-end job to close/generate a report. This is intake only; implementation and dependency choices are not approved yet. The month-end request conflicts with [CORE-03](CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), which currently derives completion and keeps reports live/recalculable without a scheduled close.

## Candidate scope

- A separately runnable worker executable/process, independently startable from the API.
- Scheduled-job orchestration and lifecycle only; domain rules remain in shared use cases/repositories.
- First candidate consumers: [TICKET-05-02](TICKET-05-02-giao-dich-den-ky.md) recurring transaction generation and [TICKET-07-04](TICKET-07-04-tong-ket-thang-ai.md) month-end report generation/close. The latter's persisted-artifact semantics are undecided.
- Retry/restart/parallel-run safety, account-timezone boundaries, structured logs, graceful shutdown, and operational health evidence.
- No AI narrative, notification delivery, new provider, or copied business logic in the worker's first slice unless separately approved.

## Decisions required before detail design

- [ ] Define what “close/generate the monthly report” means: (A) worker only triggers/report-builds a non-authoritative cache/artifact that can be rebuilt after edits (recommended to preserve CORE-03 recalculability), (B) replaceable month-end snapshot, or (C) immutable close with later adjustments. If B/C, explicitly revise CORE-03 rules for late edits/deletes and user-visible status.
- [ ] Confirm recurring schedule catch-up, duplicate prevention, month-end dates, timezone-change behavior, and what happens when a referenced wallet or jar is removed, in line with [TICKET-05-02](TICKET-05-02-giao-dich-den-ky.md).
- [ ] Decide single-instance versus multi-instance deployment assumptions and the lock/claim strategy.

## Proposed acceptance criteria — pending owner review

- API and worker are separate processes but use the same shared domain/use-case and persistence code.
- Retrying, restarting, or running multiple worker instances cannot create duplicate recurring transactions or duplicate month artifacts.
- Jobs honor each account's saved timezone and do not overwrite user-authored month notes.
- Job failures are observable and retryable without silently skipping due work; logs exclude secrets and personal transaction content.
- Unit/integration tests cover retries, duplicate claims, timezone boundaries, month-end behavior, and recovery after process restart.
- Runtime/deployment instructions document configuration, health checks, graceful shutdown, and safe rollout/rollback.

## Impact and next step

- Code: new `backend/cmd/worker` candidate plus shared `internal/usecase`, repositories, config, and database wiring; exact files remain for detail design.
- Data/API: likely no public API change; job claims/idempotency may require schema changes and an additive migration.
- Runtime/operations: new independently deployed process and environment configuration.
- Docs: `API.md` only if contracts change; architecture/ERD/requirements, validation matrix, changelog, and an ADR as design requires.
- Approval: detail design and implementation plan are required before code because this changes architecture, persistence, and deployment/runtime.

Status: draft. Next artifact: clarify month-report semantics and recurring edge rules, then prepare a written detail design for owner review.
