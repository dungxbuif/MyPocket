---
artifact_type: detail_design
id: FINANCE-ASSISTANT-SERVICE-20261110
status: implementation_extended_for_durable_runs
owner: shared
approval: approved_by_owner_2026_11_10
---

# Finance Assistant service: authenticated app proxy and owner isolation

## Problem and scope

Move the native Trợ lý runtime into a separate Python service using LangGraph and LangMem. The native app and Telegram are channel adapters with separate timelines and one shared owner-memory namespace. Each owner receives an opaque, durable workspace under `/data/workspaces`; MyPocket remains the identity and ledger authority.

## Approved implementation

- Keep `/api/v1/ai/advisor/*` as the native compatibility facade. MyPocket authenticates the existing session and forwards only allow-listed requests over the private network with a short-lived `finance-assistant` assertion.
- The service rejects client owner headers and derives workspace scope only from a verified assertion or Telegram binding.
- MyPocket issues revocable `mpa_` delegated credentials whose allowed HTTP operations are checked by MyPocket. Full owner `mpk_` keys are not stored by the service.
- Telegram private chats pair through an expiring, single-use code. App and Telegram checkpoints are distinct; durable memories are owner-scoped.
- The assistant returns channel-neutral text and typed parts. App keeps rich rendering; Telegram renders text. The first release is read-only for financial data.
- Existing proposal confirmation and ledger mutations remain in the Go backend. The external service does not execute financial writes; its `action_drafts` table is reserved schema and is not an advertised capability.
- PostgreSQL is authoritative. Markdown files are atomic, reviewable projections (`USER.md`, `MEMORY.md`, `DREAMS.md`). Database and `/data` are both backed up; if their timestamps differ, database state wins and projections must be regenerated/reconciled rather than imported as authority.

## Rejected alternatives

- Direct app-to-assistant auth, because it creates a second public credential boundary.
- Forwarding app JWTs or full-owner `mpk_` keys, because authority is not bounded to assistant operations.
- Sharing one app/Telegram transcript, because delivery noise and platform formatting would pollute native history.
- Copying GoClaw source, because the inspected repository is non-commercially licensed; it remains a behavioral reference. OpenClaw MIT mechanisms may be ported only when LangGraph/LangMem leave a measured gap.

## Verification

- Assertion: valid, expired, wrong audience, spoofed owner header.
- Isolation: two owners cannot share checkpoint, memory, pairing, draft or filesystem path.
- Channels: app and Telegram timelines differ while explicit memory recall crosses channels.
- Telegram: expired/replayed pairing and duplicate update IDs are idempotent.
- Memory: correction supersedes old value; untrusted tool/source content is not promoted; dream provenance is recorded.
- Writes: the external service reports `write=false` and `direct_writes=false`; existing Go-owned proposal confirmation remains owner-bound and idempotent.
- Runtime: restart and backup/restore preserve memories and receipts; revocation stops access.

## Reconciliation surfaces

Architecture, auth/API reference, OpenAPI, ADR, operations/backup runbook, validation matrix, context, backlog and changelog must match the shipped implementation. Production release remains a separate owner-created `prod-<semver>` operation under the frozen homelab contract.

## Implemented boundary (reconciled 2026-11-10)

- Native iOS never receives model credentials and never calls a model endpoint. It calls the existing `/api/v1/ai/advisor/*` facade; MyPocket authenticates the owner and proxies the allow-listed routes to the private Python service.
- `/messages/async` durably creates one owner-scoped run and encrypted work item,
  then returns `queued` without waiting for the provider. Replays of the same
  `client_request_id` return the same run. The worker leases jobs with
  `SKIP LOCKED`, exposes `queued`/`running`/terminal state through the existing
  run route, retries bounded transient failures, and cooperatively cancels both
  queued and running work. Completed messages remain the source for rich native
  history.
- The per-run `mpa_` delegation is encrypted at rest and expires after ten
  minutes. The proxy does not revoke it at acknowledgement because the worker
  still needs it; the assistant deletes the encrypted copy on every terminal
  outcome and the backend credential expires independently. Synchronous
  requests retain immediate revocation.
- The service uses LangGraph 1.0.1, LangMem 0.0.30 and PostgreSQL-backed checkpoints/store. OpenAI-compatible model calls are bounded to 240 seconds, 4096 output tokens and zero library retries/fallbacks.
- Synchronous app submissions receive a ten-minute `mpa_` delegation that is
  revoked after the response. Asynchronous submissions retain only an encrypted
  per-run copy until a terminal outcome; the backend credential independently
  expires after ten minutes. Telegram adapters share a rotated owner delegation
  with a 90-day maximum lifetime; removing the final adapter revokes it.
- Each owner may register multiple Telegram bots under distinct labels. Bot tokens are accepted once, validated with Bot API, encrypted at rest, and never returned. Pairing codes are adapter-scoped, single-use and expire after ten minutes.
- Telegram webhooks authenticate, persist a unique inbound job and return `202` before model work. The worker leases jobs, retries transient failures and keeps update/message idempotency.
- Explicit memories are owner-scoped and can be recalled across channels. App and Telegram transcripts/checkpoints remain separate. Dream candidates are derived only from user messages, keep provenance, and are not auto-promoted into trusted memory.
- PostgreSQL is authoritative. `/data/workspaces/<opaque-owner>/` contains mode-0600 atomic Markdown projections. Production backup covers both the assistant database and the external workspace volume.
