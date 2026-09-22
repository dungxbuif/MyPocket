---
artifact_type: adr
id: ADR-005
status: accepted
owner: shared
date: 2026-09-21
---

# AI entry: persistent review and explicit approval

## Context and authority

Owner requested long-press Add → AI chat → prefilled transaction list, editing, approval and rejection, then explicitly requested implementation. [AI-ENTRY-01](../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) scopes that instruction. The broader [two-flow plan](../work/tickets/TICKET-09-DETAIL_DESIGN.md) remains only partly authorized; no separate phase.

## Decision

2026-09-22 follow-up: the owner replaced conversation/session behavior with a one-shot process. The legacy `ai_entry_sessions`/`ai_entry_messages` schema remains for process compatibility, but no messages are written or exposed; see [ADR-006](ADR-006-private-ai-attachments.md) and the [stateless contract](../superpowers/specs/2026-09-21-stateless-ai-entry-design.md).

Persist an owner-scoped process/idempotency record, proposals and request identities in PostgreSQL (migration 000010); conversation messages were part of the initial implementation but are now unused. Draft extraction has no write tools. Owner-scoped approval locks the proposal, revalidates wallet/category/date/amount, writes one ordinary income/expense row and updates proposal status in the same transaction. A proposal's transaction ID remains a receipt after ledger deletion, preventing replay from recreating deleted data. Editing uses expected version; approved/rejected proposals are terminal.

Manual Add remains normal click. Long press opens a one-shot AI entry sheet; financial Q&A is not a mode in this entry UI. Image/PDF input goes through backend OCR then a text-only model adapter. This bounded slice uses synchronous requests and a two-minute processing lease; no automatic provider POST retries. Approved attachment retention is defined by [ADR-006](ADR-006-private-ai-attachments.md).

## Alternatives and consequences

- Client-only drafts would be easier, but closing/reloading would lose review state and could bypass confirmation replay protection.
- Letting the model call transaction-create would blur intent and approval; explicit UI approval is the sole ledger entry point.
- Full worker/storage/transfer rollout would widen this owner-directed UI slice substantially. Bounded synchronous extraction ships the review flow first, with visible failure/unknown-submission recovery and unsupported transfer approval.

Tradeoffs: no retained image preview, no durable background OCR recovery, no automatic proposal merge across messages, and no transfer/credit ledger. Unsupported/missing fields require correction or rejection. Account limit is 20 extraction attempts per rolling 24 hours and 20 messages/session. Live AI quality requires actual model credentials and owner samples; fixture success is not a quality benchmark.

## Trace and verification

[Ticket](../work/tickets/TICKET-09-01-nhap-lieu-chung-tu.md) · [Backlog](../work/BACKLOG.md) · [Validation](../work/VALIDATION_MATRIX.md) · [API](../architecture/API.md) · [ERD](../architecture/ERD.md) · [Architecture](../architecture/ARCHITECTURE.md) · [UI](../design/screens/assistant/README.md) · [Release](../releases/CHANGELOG.md).

Proof: Go unit/HTTP/provider tests, real PostgreSQL concurrent approval and cross-owner fixtures, frontend tests/browser fixtures, and live Chrome opening the chat through the real FE proxy. End-to-end extraction against the owner's AI endpoint is pending configuration; human acceptance remains open.
