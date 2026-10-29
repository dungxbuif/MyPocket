---
artifact_type: adr
id: ADR-008
status: accepted
owner: shared
human_fields:
  - decision_approval
  - final_status
ai_fields:
  - context
  - alternatives_considered
  - consequences
  - linked_work
shared_fields:
  - decision
  - trace
trace:
  requirements: [REQ-F-004, REQ-NF-002, REQ-NF-003]
  phase: PHASE-003
  tickets_or_bugs: [BUSINESS-LOGIC-AUDIT-2026-09-10]
  detail_design: ../superpowers/specs/2026-08-23-mypocket-system-design.md
  master_docs:
    - ADR-003-offline-sync-conflict-review.md
    - ../architecture/ARCHITECTURE.md
    - ../architecture/API.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-008: Commit offline mutations and sync evidence atomically

## Status

- Status: accepted
- Date: 2026-09-10
- Decision approval: owner requested “tiếp tục fixx” after the ADR-008 proposal and blocker report.

## Context

The current sync service commits a domain mutation first, appends its change-feed row in a second transaction, and stores its idempotency result in a third transaction. A crash or database error between those commits can leave authoritative financial state without a corresponding change or replay record. Retrying the same create may then be rejected as a duplicate even though the original mutation succeeded. Direct REST finance/portfolio mutation handlers call domain repositories without appending any sync change, so another device cannot rely on the cursor feed to observe those writes.

This violates ADR-003's deterministic retry and no-silent-data-loss intent. It cannot be corrected safely in an HTTP handler because finance, planning, portfolio, cursor reservation, change append, and mutation receipt currently own separate transaction boundaries.

## Decision

Introduce one server-side command transaction per authoritative finance/portfolio mutation, regardless of whether it enters through REST, recurring/draft automation, or `/sync/mutations`. For an offline sync mutation:

1. Lock or reserve `(user_id, mutation_id)` and reject hash reuse.
2. Execute the domain command through repositories bound to the same `*sql.Tx`.
3. Reserve the per-user cursor and append the normalized change in that transaction.
4. Store the final mutation result in that transaction.
5. Commit once; emit the restricted audit event after commit with the mutation ID as correlation ID.

Repositories should accept a small shared query/transaction interface rather than starting nested transactions. Direct HTTP and worker commands use the same coordinator so their committed writes append a change before commit. Sync additionally reserves and stores the idempotency receipt in that transaction. Rejected validation and conflict results may be persisted without a domain change, but still use the mutation reservation transaction so retries are deterministic.

## Alternatives Considered

- Transactional outbox only: protects change delivery but still leaves mutation-result idempotency split unless the receipt is part of the same outbox transaction.
- Reconcile missing changes after restart: eventually repairs the feed but cannot reliably distinguish an applied create from an unrelated duplicate and weakens immediate read-after-sync behavior.
- Retry append/store independently: reduces transient failures but does not close the crash window.
- Leave the current behavior: rejected because money mutations require deterministic, recoverable sync semantics.

## Consequences

- Positive: domain state, cursor/change feed, and idempotency receipt become all-or-nothing.
- Positive: a replay can return the original result without reapplying accounting effects.
- Negative: finance and portfolio repositories need transaction-bound command paths and integration tests with injected failures at each boundary.
- Neutral: the public sync request/response schema does not need to change.

## Acceptance Tests

Implementation: `platform/commandtx` binds repositories to one transaction with nested savepoints. Per-user advisory locks precede entity locks and receipt lookup. `platform/changefeed` appends canonical entity payloads and advances the cursor before commit. Accounting emits affected wallets plus the transaction. Draft confirmation and portfolio workers use these same command paths. Resync reads a repeatable-read snapshot. Audit follows commit and remains best-effort.

Migration `0012_atomic_sync_asset_feed.sql` adds asset entity/operations to existing CHECK constraints. Apply before upgrading API/worker; historical migrations and data are preserved. Old feed omissions cannot be reconstructed, so existing integrations must perform a full resync after upgrade. Rolling back binaries restores the earlier sync limitations.

Browser retries preserve the mutation ID after an ambiguous response. A batch commits per mutation; retrying the unchanged batch replays any committed prefix.

- Failure before domain write leaves no domain row, change, or mutation receipt.
- Failure after domain write but before change append rolls back all three artifacts.
- Failure after change append but before receipt storage rolls back all three artifacts.
- Retrying after an ambiguous client disconnect returns the stored result and does not reapply balances.
- Two concurrent requests with the same mutation ID apply at most once; different hashes produce deterministic rejection.
- A direct REST or worker mutation becomes visible through `/sync/changes` with the committed entity version.

## Linked Work

- Prior decision: [ADR-003](ADR-003-offline-sync-conflict-review.md)
- Audit record: [Business Logic Audit](../work/test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md)
- Detail design: [Atomic Sync](../work/phases/ATOMIC-SYNC-DETAIL_DESIGN.md)
- Reusable audit skill: [SKILL.md](../../.agents/skills/auditing-mypocket-business-logic/SKILL.md)
