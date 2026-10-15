# Advisor: fresh chat after30 idle minutes

Status: implemented_locally. Owner approved30-minute behavior; the lifecycle
contract is now reflected in the local backend reset/expiry and native Chat mới
slice. Production rollout and real PostgreSQL expiry proof remain open.
[Work item](../../work/tickets/NATIVE-ADVISOR-IDLE-RESET-20261015.md).

## Outcome and boundaries

After30 minutes without accepted user input, old conversation content is no
longer visible or included in model requests. Next question starts fresh. Never
delete financial transactions, wallets, categories, budgets or confirmed actions.
Use existing native bases, green/neutral palette and Manrope. No streaming,
provider replacement, banking integration or global user usage limits.

A message accepted at time T expires at T+30min, inclusive, using server UTC.
History reads, foregrounding, polling and assistant output do not extend TTL.
A new accepted user question before expiry extends TTL; rejected/replayed input
does not. An empty chat has no deadline. Native only displays the server deadline,
and revalidates after foreground/network recovery. It cannot extend retention.

## Selected approach

Use server-authoritative generation rotation and content purge under the existing
conversation→run lock order. Enforce expiry on authenticated chat/history/run
access and through the existing worker loop, not a phone-only timer. Lazy-only
purge was rejected because data otherwise survives indefinitely when the user
does not return; deleting all rows was rejected because it destroys idempotency.

Add nullable last_user_message_at to conversations and index its expiry scan.
Backfill from latest persisted user message; never use assistant updated_at.
Keep one conversation ID per owner; generation increments and sequence remains
monotonic across reset. Clear summary and summary cursor/version on reset.

Within one transaction: lock conversation; re-evaluate idle condition; lock old
runs; mark old runs purged, clearing leases and setting terminal time; physically
delete old messages, events and fact bundles; clear summary; increment generation
and clear last_user_message_at. Preserve minimal run IDs, owner, requestID,
payload hash, generation, status and usage metadata for replay/accounting. No
raw prompts/answers/tool facts may remain in those retained fields. Failed purge
rolls back entirely; never expose partial reset or run the model on stale content.

Answer/event/fact writes must lock/read the current conversation and reject
generation mismatch or terminal purged run. An explicit reset also cancels
queued/running provider contexts in the current process; work already executing
in another process/provider may finish remotely, but cannot repopulate history or
become a native action card. Use existing lease-lost semantics for stale writes,
without logging private content.

Worker processes at most100 expired conversations per bounded transaction batch,
using SKIP LOCKED and a deadline. Continue batches in subsequent passes. Expiry
is enforced immediately on access; physical deletion occurs at a successful
worker pass, normally within its configured interval after TTL. Worker downtime/
backlog delays physical deletion, not access-time visibility enforcement. Document
this limitation; no claim of deletion exactly at30:00 or erasure from backups.
Purge is independent from recurring-work success; one failed job cannot suppress
the other. Metrics record counts, latency, IDs/errors only, not content.

## API and native contract

Owner JWT and full-owner API keys share unchanged authority and owner isolation.
Chat history metadata adds generation, last_user_message_at, expires_at and
server_time; current-generation history only. New native decoder fields are
optional for compatibility with old production until coordinated rollout.

`POST /api/v1/ai/advisor/conversation/reset` resets this owner's chat immediately;
requires Idempotency-Key, records only minimal replay metadata. Repeating a reset
must return the original reset outcome (including its server timestamp), never
erase messages accepted afterward.
Do not invent a fake finance run/user message to represent reset. Use a dedicated
minimal reset-request ledger with owner+idempotency key unique, request hash,
result generation and timestamps; no content. Reset records are not purged by
chat expiry. Different body with same key returns409.

Submitting an old purged chat request ID returns410 `advisor_history_expired`;
does not rerun provider or require a deleted user message. Changed payload under
same request ID remains409. Run polling returns status purged without old content;
old-generation history cursors return current empty/new history and generation.

Native adds Chat mới in existing header menu with destructive confirmation and
shared controls. Server-confirmed reset/expiry clears messages, stale wallet
confirmation state and active-run state, cancels old polling, retains unsent
composer draft, and ignores late old-generation responses. Expiry while offline
hides expired chat locally using last server deadline plus elapsed time; cannot
claim server deletion until online verification. Existing already queued financial
outbox remains untouched. Do not automatically submit the retained draft.

## Verification and rollout

TDD: fake-clock TTL just-before/equal/after; user versus assistant/read/replay
activity; empty chat; owner isolation with JWT/API key; same request409/410;
idempotent reset after new message; late answer/event/fact writes; parallel
reset/send and expiry/send with real PostgreSQL; worker partial failure/restart;
physical content deletion and preserved minimal usage/replay rows; unchanged
financial tables. Native API/UI tests for expired history, old response ignored,
unsent draft preserved, foreground/restart/offline, stale action disabled.

Use additive migration, isolated synthetic fixtures and backup before deployment.
Do not reset production database or purge historic chat during migration itself.
Deploy APIs/migration first, verify health/auth/replay/worker, then compatible
native build. Enabling expiry can delete existing chat older than30min; this is
irreversible for live history and requires explicit rollout acknowledgement.
Rollback can stop future cleanup but cannot restore purged chat automatically.
Owner UAT and release publication remain separate from test/install success.

Reconcile internal lifecycle/runbook/validation and public iOS guide, endpoint
catalog/OpenAPI plus Markdown AI-agent surface. Idle TTL does not alone prevent
an oversized current prompt/tool result; aggregate context-budget policy is a
separate change and must not be falsely reported as solved by retention.

Self-review: timing and purge-vs-retention semantics explicit; replay/generation
and stale actions covered; no unresolved placeholders; finance data excluded.
