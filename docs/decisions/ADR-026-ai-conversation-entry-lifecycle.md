---
artifact_type: architecture_decision
id: ADR-026
status: accepted
owner: shared
updated: 2026-10-31
---

# ADR-026 — Unified Assistant entry lifecycle contract

## Context

[Task 1 brief](../../.superpowers/sdd/2026-10-31-unified-ai-assistant/task-1-brief.md) requires a durable, owner-scoped handoff from an Assistant conversation to entry review. Existing advisor runs own conversation messages; the entry service owns proposal drafts and approval. The two authorities must remain separate.

## Decision

`advisor_entry_tasks` records an accepted request bound to owner, conversation, generation, client request ID, payload hash, and entry session. A request ID is unique within `(owner_id, conversation_id, generation)`. A repeated request with the same hash may return its task; a different hash is a conflict. The payload hash is private and excluded from task JSON. Status advances `accepted → processing → review_ready`, with `failed` or `cancelled` terminal alternatives. Proposal status vocabulary is `pending`, `applying`, `applied`, `rejected`, `expired`, `conflict`, `superseded`; proposal mutation remains with the entry service.

An Assistant `entry_batch` part carries only `task_id`, `batch_id`, positive `version`, `item_ids`, `item_count`, and `summary`. It never embeds drafts or authorizes ledger writes. Clients read current proposal data and versions from the entry service before review or approval. Unknown part kinds remain decodable so older clients can preserve conversation history. The capability matrix advertises `entry_integration`, `supported_part_kinds`, `action_kinds`, `attachment_formats`, `attachment_max_count`, `attachment_max_bytes`, and `parts_version`. Missing optional fields mean no advertised new entry action or attachment capability.

Migration `000042` adds tasks, sequenced events, and attachment binding rows. Composite foreign keys bind tasks to the same owner as their conversation and entry session, and bindings to the same owner as their task and private attachment. Attachment object keys stay in `transaction_attachments`; task/event/part payloads do not contain object URLs, OCR text, or signed URLs. A binding does not transfer object ownership. Entry attachment cleanup continues to use the private object retention policy; deleting a task does not extend an object's expiry.

Worker claims, heartbeats, completion, failure, and event append must match the task owner, conversation generation, and opaque lease token in one database transaction. Claims use database time, bounded `lease_until` and `deadline_at`, and the worker index. A reset increments conversation generation; stale workers and stale callbacks cannot publish a batch or append events after that boundary. Cancellation also fences the lease. Reaping expired tasks is terminal and must not replay a provider call. Event sequence increments atomically per task and is stable for bounded replay.

Review-ready/failed/cancelled task and event rows are retained for 30 days for idempotent recovery, then may be purged by an owner-scoped retention job once implemented. Until such a job exists, the rows remain; migration `000042` itself performs no cleanup. Existing private attachment deletion deadlines still apply independently. The down migration is for disposable fixtures only, never a production rollback.

## Consequences and verification

This task defines entity, repository interface, schema, and client decoding contract. Repository implementation, API wiring, worker execution, retention job, and actual capability advertisement belong to later tasks; no entry task starts solely because this schema exists. Go entity/migration tests and an iOS JSON fixture cover the contract. A migrated disposable PostgreSQL test is still needed to prove database execution and reset/lease races when the repository is implemented.

Trace: [migration guide](../../backend/migrations/README.md), [entry design](../work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [advisor runtime](../work/tickets/AI-ADVISOR-RUNTIME-20261020.md).
