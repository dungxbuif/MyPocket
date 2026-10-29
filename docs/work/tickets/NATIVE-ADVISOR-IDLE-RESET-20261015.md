# Advisor idle reset

Status: in_progress — owner approved30-minute idle interval with “Chốt”
on2026-10-15. The backend access-time/worker retention foundation and native
stale-history handling and additive lifecycle metadata are implemented locally;
real database proof and production rollout remain open. No production deletion
or runtime rollout occurred.

Parent: [native completion](NATIVE-FULL-COMPLETION-20261014.md).
Written design: [idle reset lifecycle](../../superpowers/specs/2026-10-15-advisor-idle-reset-design.md),
commit9ab9248d. The local implementation uses the additive last-user timestamp,
generation rotation, physical content purge with minimal replay metadata,
access-time enforcement and a bounded worker pass. The 30-minute interval is
approved; production cleanup is not enabled.
Detailed [implementation plan](../../superpowers/plans/2026-10-15-advisor-idle-reset.md)
now written for inline execution; native lifecycle metadata, real-database
proof and production API/schema rollout remain review-gated.
Affected public/agent docs: [iOS guide](../../guides/finance-assistant-ios.md),
[tool catalog](../../architecture/AI_ADVISOR_TOOLS.md). Internal context/lifecycle
and native screen specification require reconciliation when implemented.

## Intent

Stop carrying stale conversation into later finance questions, remove expired
visible history, never delete wallet/ledger/report data. The approved manual
“Chat mới” action is implemented locally alongside the 30-minute TTL. No
production history has been deleted and no installed-owner device was changed.

## Current authoritative behavior

AdvisorService.Submit fetches50 persisted messages; BuildAdvisorContext sorts
and keeps12, taking the first nonempty text part of each. Idle expiry now runs at
the owner-scoped persistence boundary before submit/history/run reads. It has no
aggregate token-budget check. Each text part is limited8192 bytes by
AdvisorMessage.Validate; number-of-messages is not a provider token guarantee.
Financial card payloads are not reinserted as conversation text.

One conversation per owner already has generation, summary and sequence fields.
StartRun locks conversation and checks owner+requestID replay before new work.
Replay loads the old user message; blindly deleting messages breaks this path.
ListMessages now scopes by owner and current conversation generation, expiring
idle content before reading. CreateAssistantMessage and AppendEvent revalidate
the current conversation generation after locking; reset/expiry cannot allow an
old provider response to repopulate the fresh chat.

## Design requirements before implementation

- Choose idle TTL and define activity as accepted user input, not app foreground,
  history reads, assistant responses or background polling.
- Server-authoritative expiry under owner-scoped conversation transaction; native
  displays/reset coherently after foreground/send. Never rely on device clock.
- Clear visible/model history and summary; isolate new generation. Cancel/purge
  old active work with established conversation→run lock ordering and reject old
  writes against current generation immediately before insert.
- Preserve minimal request-idempotency/usage metadata without retaining purged
  prompt/answer content. Explicit replay result after purge must be designed;
  current replay depends on message content and cannot simply remain unchanged.
- Pending native wallet action cards from purged chat must not execute stale
  confirmation. Financial writes already confirmed must remain untouched.
- Decide physical background deletion versus purge on next authenticated access;
  do not describe lazy deletion as deletion exactly at TTL.
- Token budget is a distinct requirement from idle TTL; tool results in a single
  active turn can still be large. No arbitrary truncation of financial facts.

These interactions upgrade a simple time filter into a conversation lifecycle/
idempotency contract change. The local slice now covers additive history
metadata and native stale-card handling; production cleanup still requires its
own review and rollout gates. Approval of the interval does not authorize
deleting production records.

## Verification

Fake-clock exact TTL boundary, continuous active chat, history-only access,
owner isolation, parallel sends/reset, late answer after reset, completed/failed
request replay after purge, summary cleared, restart/native expiry, stale action
card invalidation; assert financial rows unchanged. Real PostgreSQL transaction
tests plus API/native E2E, no production reset. Human and Markdown-agent docs must
describe actual timing/deletion semantics and limits. Release publication waits
for owner UAT.

Links: [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md),
[cancel race](NATIVE-ADVISOR-CANCEL-20261015.md).

## Local implementation checkpoint — 2026-10-15

- `AdvisorConversation` has an inclusive 30-minute server-time policy and
  nullable `last_user_message_at`; additive migration `000033` backfills it
  from user messages only and does not delete data.
- PostgreSQL access paths lock the owner conversation and purge expired advisor
  messages/events/facts, preserve run idempotency/usage tombstones, reject stale
  assistant/event writes, and keep sequence monotonic.
- The worker invokes a bounded 100-row `SKIP LOCKED` expiry pass independently
  of recurring work.
- A purged request is mapped to HTTP `410 advisor_history_expired`.
- Explicit reset is now a local additive `POST /ai/advisor/conversation/reset`
  route guarded by `Idempotency-Key`; the reset ledger contains no chat content,
  and native “Chat mới” clears only assistant state after confirmation. The
  service now also cancels in-process provider contexts for queued/running runs
  purged by the reset transaction; replayed reset keys report no old runs, so a
  retry cannot cancel newer work.
- Pure TTL, worker contract, handler mapping and existing backend advisor tests
  pass locally at `/tmp/mypocket-advisor-idle-foundation-20261015.log`. Reset
  service cancellation and owner/replay behavior pass in the full Go suite and
  focused race gate. Real
  PostgreSQL migration/race proof and production rollout remain pending. Native
  iOS now clears the local conversation
  pointer when a history/run request returns 410 `advisor_history_expired` (or a
  terminal `purged` run), removes the stale messages, and restores the unsent
  composer text so the user can submit it in a new generation without an
  automatic retry.

## Additive history metadata slice — 2026-10-15

The local slice exposes optional server-authoritative lifecycle metadata in
the existing response envelope `meta` while preserving the message array in
`data` for old clients. Metadata contains conversation ID, generation,
last-user-message timestamp, expiry timestamp and server time. The history read
uses the same owner-scoped idle-expiry enforcement as the message query; reads
never extend the deadline. Native treats the metadata as advisory state for
stale-response suppression and keeps the composer draft on expiry. Adapters that
cannot provide metadata continue returning `meta: null`; no API write,
financial data, migration or production rollout is involved in this additive
contract. Backend package tests, native history contract tests and a
warnings-as-errors simulator build are green; production and real-DB evidence
remain open.
