---
artifact_type: system_design
id: SDD-MASTER
status: approved
owner: shared
human_fields: [scope_approval, constraints, architectural_decisions]
ai_fields: [system_decomposition, flows, risks, verification, reconciliation]
shared_fields: [status, trace]
updated: 2026-08-24
trace:
  source_srs: ../../SRS.md
  approved_design: ../superpowers/specs/2026-08-23-mypocket-system-design.md
  requirements: ../requirements/REQUIREMENTS.md
  roadmap: ../work/ROADMAP.md
---

# MyPocket System Design Description

## Approval

The user approved the architecture, scope decomposition, security rules, two distinct image-ingestion flows, testing strategy, hidden audit viewer, and configurable 180-day default audit retention during design review on 2026-08-23 and 2026-08-24.

## Design Sources

- Product source: [`SRS.md`](../../SRS.md)
- Full approved design: [`2026-08-23-mypocket-system-design.md`](../superpowers/specs/2026-08-23-mypocket-system-design.md)
- Stable requirements: [`REQUIREMENTS.md`](../requirements/REQUIREMENTS.md)
- Architecture masters: [`ARCHITECTURE.md`](ARCHITECTURE.md), [`API.md`](API.md), [`ERD.md`](ERD.md), [`INTEGRATIONS.md`](INTEGRATIONS.md)

## Approved System Shape

- React + TypeScript installable PWA with IndexedDB, offline outbox, React Router, and TanStack Query.
- Go modular monolith exposed through an API process plus a separate worker process sharing domain/application packages.
- PostgreSQL as authoritative relational state and coordination mechanism.
- Private S3-compatible object storage for receipts and generated exports.
- Google OAuth with a signed stateless cookie, no server session table, no linked-device management, and no persisted Google access/refresh token.
- Multi-user isolation enforced in every Go query and command.
- Review-first shared transaction drafts for AI text, receipt OCR, AI-chat images, bank webhooks, and recurring occurrences.
- Full offline read/write with idempotent mutations, per-record versions, per-user change cursors, tombstones, and explicit user-reviewed conflicts.
- Hidden append-only audit page authorized by exact verified email equality with `AUDIT_VIEWER_EMAIL`; retention defaults to 180 days.

## Design Invariants

1. Confirmed transactions are the only ingestion artifacts that affect accounting.
2. Transfers, balance effects, sync changes, idempotency, and relevant audit writes are atomic where correctness requires one outcome.
3. Provider output is untrusted input and must satisfy schema and domain validation.
4. Cross-user access is denied even when an ID, hidden URL, or object key is known.
5. Offline conflicts never use silent last-write-wins behavior.
6. Receipt objects remain private and use short-lived authorized access.
7. Audit records exclude secrets and sensitive raw payloads.
8. VND uses integer units and initial reporting uses `Asia/Ho_Chi_Minh` boundaries.

## Delivery Slices

| Phase | Independently testable outcome |
| --- | --- |
| PHASE-001 | Deployable web/API/worker foundation with Google login, user isolation, PostgreSQL, S3 adapter, and operations baseline |
| PHASE-002 | Correct wallets, categories, transactions, transfers, adjustments, Vietnamese seed data, and receipt metadata |
| PHASE-003 | Full offline client mirror, outbox, incremental synchronization, tombstones, and conflict resolution |
| PHASE-004 | Budgets, events, recurring drafts, debts/loans, inbox, and Web Push |
| PHASE-005 | PWA navigation, dashboard, search, wallet views, analytics, comparison, and cumulative trend |
| PHASE-006 | AI text, receipt OCR, AI-chat images, bank webhook, and shared draft confirmation |
| PHASE-007 | Hidden audit viewer, retention, exports, account lifecycle, backup/restore, and production release proof |
| PHASE-008 | Deferred voice transcription feeding the established draft contract |

## Verification Contract

- Pure finance, analytics, parsing, redaction, and scheduling rules receive Go unit proof.
- Persistence, migrations, isolation, atomicity, sync, webhooks, jobs, and audit authorization receive PostgreSQL integration proof.
- React forms, drafts, conflict resolution, charts, and hidden-route behavior receive component proof.
- IndexedDB hydration, outbox replay, reconnect, tombstones, and conflicts receive browser-storage proof.
- User-visible workflows receive Playwright-style E2E and UAT proof.
- Homelab startup, migrations, worker restart, object access, push fallback, and backup/restore receive platform proof.
- Every behavior remains `planned` in the validation matrix until linked evidence exists.

## Change Control

Any implementation divergence affecting authentication, ownership, money representation, transaction atomicity, offline conflict policy, provider confirmation, audit authorization/retention, API versioning, data deletion, or deployment boundaries requires an updated detail design and ADR review before code changes continue.
