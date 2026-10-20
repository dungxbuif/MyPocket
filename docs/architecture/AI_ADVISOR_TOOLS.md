# Finance Assistant V1 — tool catalog

Status: JWT/API-key pilot, synchronous and async JSON transport plus bounded
owner-scoped SSE event replay. This catalog
describes the implemented non-mutating tool boundary (finance reads,
product-help retrieval and prepared wallet/jar/budget/transaction/category
proposals). The current
production API/worker rollout is `release-20261019-uat2`; its route-parity,
schema and authenticated UAT evidence is tracked in
[`ROLLOUT_READINESS.md`](../operations/ROLLOUT_READINESS.md). Streaming,
durable audit delivery and physical-device acceptance remain separate gates.

## External MCP operations are a separate catalog

The local [MCP/API adapter](../ai/mcp-api-parity.md) exposes named authenticated
user operations at `/api/v1/mcp` using the official Go SDK v1.8.0. Its dynamic
`tools/list` catalog currently has 101 tools, including ordinary direct REST
writes, and is not the provider-facing registry described below. A trusted
owner agent may call `transactions.create` or `wallets.create` directly under
the same canonical rules as JWT/API-key REST. No MCP scopes or universal
confirmation step are added.

Calling `advisor.messages` or `advisor.messages_async` still runs this
provider-facing registry, where `direct_writes:false` and
`confirmation_required:true` remain authoritative. `advisor.action.confirm`
maps to the persisted action endpoint with `action_id`, `conversation_id` and
`generation`; payload replacement and forged/stale generation are rejected.
Credential revalidation, idempotency receipts and `expected_version` checks
remain canonical. API-key expiry/revocation applies throughout long runs.

MCP async responses add polling handles. `advisor.run.events` returns bounded
persisted replay `{events, after_seq, replay_only:true}` (maximum 100 events),
not live model-token streaming. The two-minute default synchronous MCP tool
deadline is separate from the async run lifecycle. See
[ADR-025](../decisions/ADR-025-mcp-api-parity.md) for transport boundaries and
the [MCP contract](../ai/mcp-api-parity.md) for limits, errors, file/resource
behavior, explicit route exclusions and redacted audit metadata. These local
transport changes do not claim a production rollout.

## Deterministic AI/OCR evaluation fixtures

The checked-in fixture pack at [`evals/ai/fixtures.json`](../../evals/ai/fixtures.json)
is the local contract for overlapping OCR rows, missing dates, wallet inference,
owner-safe advisor arguments and proposal review state. It contains synthetic
IDs and amounts only; it does not contain provider credentials, prompts from
real users, OCR source text, or financial rows. The Go test
`TestAIEvalFixtures` loads these cases and exercises the real draft normalizer,
deduplicator, wallet fallback, argument decoder and advisor registry.

The fixture clock is fixed to `2026-09-22T10:00:00+07:00`, so missing-date
assertions never depend on the host clock. Exact duplicate OCR rows are merged
with their review question retained, while distinct rows (including same-day
rows with different amounts or notes) remain separate. A missing date is filled
only as a disclosed review default. A model-supplied unknown or foreign wallet
falls back to the supplied owner wallet and adds a review question; it is never
silently accepted. Advisor tool arguments are decoded with unknown fields
rejected and the authenticated principal supplies owner scope; `owner_id` is
never a model-controlled argument.

All generated AI-entry proposals remain `pending` until an explicit owner
approval request. The fixture suite asserts that no auto-approval or ledger
write occurs. `AI_EVAL_LIVE=1` remains the only opt-in path for provider/OCR
execution; the deterministic fixture test never reads provider environment
variables or contacts a live service.

Native presentation correction2026-10-15 (local, not installed update): exact
signed finance-safe money; invalid values show unavailable, not fabricated zero.
Count/JSON fallback conversion no longer uses trapping casts; unknown/malformed
fallback applies existing masking. Range/goal-percent/privacy-sensitivity RED→GREEN;
final287/287PASS (279unit+8native chatUI) includes numeric/count/JSON and hosted
card checks. Device/full privacy acceptance remains
separate; no backend owner/API/tool contract change.
[Work](../work/tickets/NATIVE-AI-CARD-NUMERIC-SAFETY-20261015.md).

## Security boundary

Portfolio trade correction is an explicit owner-confirmed HTTP contract, not an
Advisor tool; its preview/commit routes are documented in the [portfolio
guide](../guides/portfolio-ios.md). The remaining decision-gated AI residuals
are tracked as explicit contracts rather than implied capabilities: [AI action
expansion](../work/tickets/FEEDBACK-20261019-AI-ACTIONS-CONTRACT.md) and
[product-help/RAG refresh](../work/tickets/FEEDBACK-20261019-PRODUCT-HELP-RAG-CONTRACT.md).
Until their contracts are approved, the implementation below remains the
authoritative boundary.

- The model never receives SQL, database credentials, arbitrary URLs, or a caller-supplied owner ID.
- The backend supplies the authenticated `Principal.OwnerID` and account timezone to every tool.
- No model tool mutates financial data. Finance tools read owner data; all `propose_*` tools validate and return proposals only. Explicit confirmation is a separate owner/generation-bound HTTP action. Credit/payment/external transfer and SQL tools are not registered.
- A provider tool-call response is validated as a complete batch before any tool executes. Every call must have a non-empty correlation ID and IDs must be unique within that response; malformed or ambiguous batches fail closed with `ErrAdvisorToolInvalid`, avoiding partial finance queries. This transport check is additive to tool argument/schema validation.
- User API-key authentication and redacted advisor access audit are implemented locally. Audit payloads are durably buffered in PostgreSQL migration 037 before the Redis fast path; the worker leases and retries delivery. Production migration/worker health, alerting, configured-provider proof and public third-party deployment remain release prerequisites. SSE replay is available for persisted events; it does not yet provide token-by-token model streaming.

## Registered tools

| Name | Input | Result |
| --- | --- | --- |
| `search_transactions` | `range`, optional wallet/category IDs, type, report scope, note substring, min/max, limit ≤50, cursor | owner-scoped transaction page; `total_count` is calculated before pagination |
| `get_transaction` | `{ "transaction_id": "..." }` | one owner-scoped detail, including report flag, transfer ID and persisted `version`; report exclusion does not hide detail |
| `get_finance_summary` | `range`, optional wallet/category IDs and descendants | report-included income, expense, net and count; transfers/system transfer categories excluded |
| `compare_spending_periods` | `current`, `previous`, optional wallet/category filters | current/previous expense, delta and nullable basis-point change when baseline is zero |
| `get_wallet_balances` | optional `wallet_ids` | current owner-wallet balances from opening balance plus full ledger |
| `get_budget_progress` | optional `budget_ids`, optional `active_on` | existing budget calculation and scope |
| `get_goal_progress` | optional goal wallet IDs | goal wallet current/target/remaining view |
| `get_jar_progress` | `month`, optional `jar_ids` | SELECT-only jar allocation/spend; does not initialize month/config rows |
| `get_product_help` | `{ "query": "...", "max_results": 1..5 }` | versioned handbook excerpts with stable source IDs; no owner data or external URLs |
| `list_categories` | optional `kind`, up to 20 owner wallet IDs, `include_archived`, `limit` 1..100 and opaque cursor | owner-visible category catalog sorted by ID, with duplicate names preserved, `editable`, `is_system`, `is_active` and `version`; cursor is bound to owner and filters |
| `get_category` | `{ "category_id": "..." }` | one owner-visible or system category with parent/wallet scope and persisted `version`; system categories are readable but not editable |
| `propose_create_wallet` | required `name`, `type` (`basic`, `goal`, `credit`); optional balance/description/total flag and type-specific goal/credit fields | validated `wallet_creation_proposal`, status `proposed`; creates no wallet |
| `propose_create_jar` | required `month` (`YYYY-MM`), `name`, `allocation_mode` (`none`, `fixed`, `percent`); mode-specific amount/percent | validated `jar_creation_proposal`, status `proposed`; creates no jar |
| `propose_update_jar` | required UUID `jar_id`, `month`, `name`, `allocation_mode`; mode-specific amount/percent | validated `jar_update_proposal`, status `proposed`; replaces no config until explicit confirmation |
| `propose_create_budget` | required `name`, positive `limit_amount`, `start_date`, `end_date`; optional owner-scoped wallet/category UUIDs | validated `budget_creation_proposal`, status `proposed`; creates no budget |
| `propose_update_budget` | required UUID `budget_id`, `name`, positive `limit_amount`, `start_date`, `end_date`; optional owner-scoped wallet/category UUIDs | validated `budget_update_proposal`, status `proposed`; replaces no budget until explicit confirmation |
| `propose_create_transaction` | required wallet UUID, income/expense type and positive amount; optional category/date/note/report flag | validated `transaction_creation_proposal`; no ledger write |
| `propose_update_transaction` | transaction UUID, replacement ordinary transaction fields, required positive `expected_version` snapshot | validated `transaction_update_proposal`; the ID/version must have been read in the same advisor run; linked transfer/credit rows remain on dedicated flows; stale confirmation returns `409` |
| `propose_delete_transaction` | transaction UUID and required positive `expected_version` snapshot | validated `transaction_delete_proposal`; no delete until confirmation; stale confirmation returns `409` |
| `propose_transfer_wallet` | source/destination wallet UUIDs, positive amount, optional date/note | validated atomic `wallet_transfer_proposal`; no paired rows until confirmation |
| `propose_create_category` | name, category kind; optional parent/wallet/icon fields | validated `category_creation_proposal`; no category write |
| `propose_update_category` | category UUID, replacement category fields, and required positive `expected_version` snapshot | validated `category_update_proposal`; the category/version must have been read in the same advisor run; system/foreign categories rejected by domain handler; stale confirmation returns `409` |
| `propose_recategorize_transaction` | ordinary transaction UUID, category UUID, and required positive `expected_version` snapshot | validated `transaction_recategorize_proposal`; canonical transaction handler rejects transfer/credit/adjustment rows and stale snapshots |

Wallet proposals require positive target for goal wallets, positive credit limit
for credit wallets, valid optional calendar dates and statement/due days1–31.
Native action cards with persisted context require explicit confirmation through
`POST /api/v1/ai/advisor/actions/{action_id}/confirm`. The client sends only
`conversation_id` and `generation`; the server resolves the owner-scoped,
persisted proposal and derives all financial fields before delegating to the
durable wallet mutation ledger with `action_id` as the idempotency key. A lost
confirmation response can therefore be replayed without a second wallet.
Legacy proposal messages without server action context retain the ordinary
`POST /api/v1/wallets` fallback with a mounted-card key. The model never
executes the action.
Confirmation lookup follows owner-scoped history cursors, so a card loaded
through older-message pagination remains confirmable in its current generation.
It does not expand the 12-message model context. Stale generations still return
a conflict, foreign-owner cards remain unavailable, and replay uses the same
action ID. This lookup repair is locally verified on 2026-10-19; production
rollout and owner UAT remain separate.
Capabilities advertise `write:false` and the full bounded action set (wallet,
jar, budget, ordinary transaction create/update/delete, internal transfer,
category create/update and recategorization): this means client-confirmable
proposal, not a write tool. A mounted native confirmation card
now keeps one idempotency key across same-session transport retries, so a timeout
does not silently create a new request identity. Legacy proposal messages
without `action_id` use the mounted-card fallback key; new messages keep the
same replay boundary after app restart without adding a separate prepared
action row or confirm endpoint.

Native proposal renderer now validates the existing schema before offering
confirmation: supported type/nonempty bounded name, exact nonnegative safe money,
required goal/credit amounts, valid goal date and credit days. It does not turn
unknown type into basic or invalid present amount into0. Documented missing
opening-balance/total defaults remain; optional null fields remain supported.
This renderer fix is local; durable action/replay and production gates above
remain open. [Verification](../work/tickets/NATIVE-AI-WALLET-PROPOSAL-VALIDATION-20261014.md).

## Strict argument rules

Arguments are bounded JSON objects decoded with unknown fields rejected. IDs are trimmed/deduplicated and resolved under the authenticated owner. Calendar ranges are inclusive account-local dates normalized to UTC half-open instants; ranges longer than twelve months are rejected. Money is integer VND and must remain within the JavaScript-safe integer bound.

The provider-facing JSON Schema mirrors the Go decoder. `range` and `current`/`previous` require `{ "from": "YYYY-MM-DD", "to": "YYYY-MM-DD" }`; `month` requires `YYYY-MM`; array filters are bounded to 100 IDs; strings and cursors have explicit maximum lengths. The adapter serializes assistant calls as `{ "id", "type": "function", "function": { "name", "arguments" } }` and preserves `tool_call_id` on tool results. Current loop permits at most eight tool executions and six provider requests per run. After eight tools, the next permitted provider request omits tools; a further tool request or exhausted provider-call budget returns a loop-limit error. These are run-safety bounds, not per-user usage quotas. Transaction/category updates now use persisted positive row versions and atomic compare-and-swap; AI proposal tools require `expected_version`, and confirmation rejects missing or stale snapshots with a 409 before any write.

Credential validity is checked before each provider request, before each tool execution, and again before persisting the assistant answer. Revoked/expired API keys and invalid sessions fail closed. `POST /runs/{id}/cancel` updates the owner-scoped run and interrupts an in-process provider context immediately; each running process also polls the persisted run status through the non-locking status probe, so cancellation handled by another API process interrupts the provider within the bounded poll interval (250ms nominal, 1s probe timeout). The explicit reset service also interrupts queued/running contexts purged by its transaction; a replayed reset never cancels newer work. The approved 2026-10-15 conversation→run lock/recheck and idle-cleanup worker are deployed in the current API/worker image; the controlled-lock late-insert regression passes (realDB gate4/4). The synchronous endpoint still returns `run_id` after provider work; the native composer uses the separate async endpoint described below to expose Stop during the initial request. [Cancellation audit](../work/tickets/NATIVE-ADVISOR-CANCEL-20261015.md). Cancellation remains fail-closed at persistence boundaries: a late provider result cannot create an assistant message after the run is cancelled or its generation is purged.

Known response views include `finance_summary`, `spending_comparison`, `transaction_search`, `transaction_detail`, `wallet_balances`, `budget_progress`, `goal_progress`, `jar_progress`, `product_help`, `wallet_creation_proposal`, `jar_creation_proposal`, `jar_update_proposal`, `budget_creation_proposal` and `budget_update_proposal`. Unknown or malformed views use a textual fallback, masked before formatting in the local corrected native renderer. Device/full-app acceptance remains separate. No model-provided markup is rendered.

## Implemented local routes

Advisor routes accept JWT/session or a full-owner user API key. There are no per-key scopes; the tool allowlist and confirmation boundary remain independent of bearer authority. The routes are mounted in production behind auth; owner-authenticated UAT is still required.

- `GET /api/v1/ai/advisor/capabilities` (advertises the read-only tools plus the
  bounded `propose_*` actions; `direct_writes:false` and
  `confirmation_required:true` are authoritative)
- `GET /api/v1/ai/advisor/overview?month=YYYY-MM` (real report summary; no model call)
- `POST /api/v1/ai/advisor/messages`
- `POST /api/v1/ai/advisor/messages/async`
- `GET /api/v1/ai/advisor/conversation/messages?conversation_id=...&before_seq=...` (last 50 by default; pass the oldest loaded sequence to page backwards). The existing `data` array may include additive `meta` lifecycle fields: `conversation_id`, `generation`, `last_user_message_at`, `expires_at` and server `server_time`; reads never extend the deadline.
- `POST /api/v1/ai/advisor/conversation/reset` with a required `Idempotency-Key` (clears only advisor content, rotates generation, and returns a content-free reset result)
- `GET /api/v1/ai/advisor/runs/{id}`
- `GET /api/v1/ai/advisor/runs/{id}/events` (owner-scoped replay-only SSE;
  it never calls the model and is not token streaming)
- `POST /api/v1/ai/advisor/runs/{id}/cancel`
- `POST /api/v1/ai/advisor/actions/{action_id}/confirm` with
  `{"conversation_id":"...","generation":1}`; persisted wallet or jar
  proposal fields are server-owned, stale generations return `409`, unknown
  actions return `404`, and malformed/tampered request fields return `400`.
  Jar creation confirmations delegate to the owner-scoped jar route and reuse
  the `transaction_mutations` ledger under `jar.create`. Update confirmations
  delegate to the owner-scoped `PUT /api/v1/jars/{jar_id}/months/{month}` route;
  replacement semantics are replay-safe and do not create a second jar.
  Budget confirmations delegate to `POST /api/v1/budgets` and reuse the same
  ledger under `budget.create`; the existing wallet/category/date validation
  remains authoritative. Budget update confirmations dispatch to the
  owner-scoped `PATCH /api/v1/budgets/{id}` route with replacement semantics;
  no second budget is created.

### Idle history retention

The server records the timestamp of the last accepted user question. After an
inclusive 30-minute idle boundary, an authenticated history/run read or the
next submit rotates the conversation generation and purges advisor messages,
events and fact bundles only; wallets, transactions, reports and already
confirmed financial actions are never touched. Run rows remain as minimal
owner-scoped idempotency/usage tombstones, so retrying the same purged request
does not call the provider again and returns `410 advisor_history_expired`.
History reads, polling and assistant output do not extend the deadline. The
worker also performs bounded `SKIP LOCKED` cleanup for owners who do not return;
worker downtime can delay physical deletion, while access-time enforcement
still hides expired content. The local native header now exposes an explicit
idempotent “Chat mới” reset; history metadata is additive, and the reset/expiry
routes are present in the current production OpenAPI; owner-authenticated production UAT remains open.

### Async initial-run cancellation (deployed 2026-10-16)

The synchronous `POST /messages` contract remains unchanged. A separate
`POST /messages/async` contract reserves and returns the owner-scoped
`run_id` before provider work starts (`202 Accepted`), so the client can call
`POST /runs/{id}/cancel` during the first request. The server will execute the
same bounded tool loop in the background, persist the same message/event/usage
records, and keeps idempotent `client_request_id` replay semantics. The native
composer uses this endpoint and polls the existing run resource. Tests prove
immediate run visibility, cancellation before provider completion, and
unchanged synchronous behavior. Production Swagger exposes the route; an
owner-authenticated provider run is still required for end-to-end UAT.

All listed routes use the same owner credential authority. Create/list/revoke keys through `/api/v1/api-keys` with JWT or owner key; secret is shown once and only its hash persists. See [full-key guide](../guides/api-keys.md). Advisor auth audits to Redis stream `mypocket:audit:advisor` with request/credential ID, route, status, decision and latency; no prompts, notes, tokens or monetary rows. The sink fingerprints logical events, suppresses identical retries for 24 hours, trims the stream approximately at 10,000 entries, and uses a bounded three-attempt Redis client retry. Audit failure remains best-effort at the request boundary. Redis persistence/outage recovery beyond that retry window, alerting and a cross-process outbox remain release gates; see the [audit delivery contract](../work/tickets/FEEDBACK-20261019-AUDIT-DELIVERY-DETAIL_DESIGN.md). `POST /messages` waits for the bounded provider/tool loop and returns JSON. `GET /runs/{id}/events` replays at most 100 persisted events after `after_seq` (or `Last-Event-ID`) as authenticated SSE and never calls the model again. It returns `410`/status through the ordinary run endpoint when history is compacted; the current implementation is replay-only and does not claim token streaming. Lost-POST recovery and fact-bundle drilldown remain separate gates.

The native Assistant loads recent messages, exposes “Tải tin cũ” when another
page may exist and requests `before_seq` using the oldest loaded sequence.

The native Assistant also exposes “Chat mới” in the overflow menu. It calls the
owner-scoped reset route with a persisted idempotency key; a replay returns the
original generation and `server_time`, cannot delete a newer question, and does
not return old run IDs for cancellation. The reset ledger stores no prompt,
answer, tool fact or financial row.
Current native chat behavior and local verification limits are described in the
[human and agent guide](../guides/finance-assistant-ios.md). Archived web code is
not the maintained product surface.

Before each provider request, the backend builds a deterministic chronological context from at most the 12 most recent persisted user-visible text messages. Older card payloads are not treated as current facts; the model must call the owner-scoped finance tools again for current numbers.

## Product-help evidence and limitations

The orchestrator prepends an embedded versioned product handbook (`v1.4`,
2026-10-20) with the native primary/secondary navigation map, supported
proposal actions, explicit jar-proposal capability and non-mutating tool
boundaries. The
`get_product_help` read-only tool now searches a source-controlled section
catalog (`v1.3`) covering every native secondary destination and returns
version/source IDs plus bounded excerpts; no external retrieval, owner data or
arbitrary URLs are accepted. The handbook describes
`Trợ lý` as the primary native tab and keeps only secondary destinations under
`Tài khoản`; the provider regression rejects the former Assistant-in-More
wording. A no-match query returns `not_found` rather than dumping the handbook.
This is deterministic local retrieval, not a general semantic index or live
RAG service; real-provider answer quality and production gates remain separate.
The catalog accurately describes wallet/jar/budget/transaction/category
proposals, including the confirmation and `expected_version` boundaries, while
explicitly marking unsupported portfolio, credit-payment, recurring-schedule,
Travel Mode and feedback writes. Feature title matches outrank incidental
mentions in navigation help. The system message also injects the account-local
calendar date through a testable clock helper; invalid or missing timezones do
not produce a guessed date. Context remains bounded to the latest 12 visible
text messages, and no hidden financial memory or card payload is retained. The
[feedback regression design](../work/tickets/FEEDBACK-20261019-AI-DETAIL_DESIGN.md)
records offline lookup, older-card confirmation and receipt-recovery evidence.

## Reliability candidate flags

When `AI_ADVISOR_GROUNDING_ENABLED=true`, the provider-facing registry adds the
terminal `finish_answer` protocol. Finance claims must reference facts returned
by tools in the same run; the server renders typed values and rejects free-text
financial finals. When `AI_ADVISOR_RUN_LEASES_ENABLED=true`, migration 000041
lease/heartbeat/atomic completion and the independent worker reaper are active.
Both flags default false until the configured model and migrated database pass
the release gates.

If the provider returns free text or invalid claims while grounding is enabled,
the orchestrator permits one repair call within the existing six-call budget.
If repair still fails after at least one validated finance result, it returns a
non-quantitative fallback while preserving the verified cards; it never
publishes the ungrounded model prose.

### Review closure: snapshots, context and grounded finals

The execution loop records successful resource reads and validates proposal
references before adding a proposal to the persisted answer. An unread ID,
stale version or non-editable system category returns a correlated tool error
(`snapshot_required`, `snapshot_mismatch`, `snapshot_not_editable`) so the model
can read again. Rejected proposals never become confirmation cards. Canonical
confirmation still performs owner checks, CAS and idempotency independently.

Each fact in the provider tool result includes its run-local `source_ref`, for
example `lookup.q1.expense`; `finish_answer.claims[].fact_ref` uses that exact
value. Labels come from server fact identifiers rather than model `slot` text.
For an accepted proposal use `mode=finance, template_id=proposal, claims=[]`;
the server renders review/confirmation wording. Product help can use
`source_id` returned by `get_product_help`; the server returns that handbook
excerpt. Clarification selects one server-owned question with `question_id`
(`intent`, `wallet`, `category`, `period`, `transaction`, `amount`). Model `explanation` is
not published by these paths. Eight finance calls do not consume the terminal
`finish_answer` allowance; all provider calls, including repairs, remain within
the six-call budget. Repair transcripts always include a matching tool result.

Requests retain at most 12 visible text messages and 24 KiB including compact
continuation hints. Tool transcripts stop at 96 KiB and return the sourced
cards already obtained. With runtime leases enabled, previous context is
loaded from the current owner's current generation; wallet/category/transaction
references are revalidated through finance reads before use. Amounts, notes and
CAS versions are never recovered from memory. Explicit broad queries clear the
previous selected wallet/category scope. Date boundaries use the account's
calendar and `AddDate`, including daylight-saving transitions.

The 31-case JSON file is an acceptance dataset. Its schema/registry test checks
dataset validity only; it is not a 31-prompt model-quality run. See the
[review closure](../work/tickets/AI-ADVISOR-REVIEW-FIX-20261020.md) for behavioral
regressions and the separate live-provider release gate.

## Audit delivery

Advisor access events remain redacted and owner-credential scoped. The API
persists each event in PostgreSQL `audit_outbox` (migration 037) before the
Redis fast path; the worker uses a leased `SKIP LOCKED` batch and bounded
backoff to retry delivery. Production migration, worker health and alerting
are still rollout gates; this buffer is not a compliance archive.
