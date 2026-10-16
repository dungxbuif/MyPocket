# Finance Assistant V1 — tool catalog

Status: JWT/API-key pilot, synchronous and async JSON transport. This catalog
describes the implemented non-mutating tool boundary (finance reads,
product-help retrieval and prepared wallet/jar proposals). Production runs
API/worker image `release-20261016-7846945`; owner-authenticated proposal and
provider UAT are still required before closing the release gate. Streaming,
durable audit delivery and cross-process cancellation remain separate gates.

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

- The model never receives SQL, database credentials, arbitrary URLs, or a caller-supplied owner ID.
- The backend supplies the authenticated `Principal.OwnerID` and account timezone to every tool.
- No model tool mutates financial data. Finance tools read owner data; `propose_create_wallet`, `propose_create_jar`, `propose_update_jar`, `propose_create_budget` and `propose_update_budget` validate and return proposals only. Explicit confirmation is a separate owner/generation-bound HTTP action. No `create_transaction`, `update_transaction`, `delete_transaction`, `execute_sql`, bank or payment tool is registered.
- User API-key authentication and best-effort Redis advisor access audit are implemented locally. Durable audit delivery/alerting, SSE replay, configured-provider proof and public third-party deployment remain release prerequisites.

## Registered tools

| Name | Input | Result |
| --- | --- | --- |
| `search_transactions` | `range`, optional wallet/category IDs, type, report scope, note substring, min/max, limit ≤50, cursor | owner-scoped transaction page; `total_count` is calculated before pagination |
| `get_transaction` | `{ "transaction_id": "..." }` | one owner-scoped detail, including report flag and transfer ID; report exclusion does not hide detail |
| `get_finance_summary` | `range`, optional wallet/category IDs and descendants | report-included income, expense, net and count; transfers/system transfer categories excluded |
| `compare_spending_periods` | `current`, `previous`, optional wallet/category filters | current/previous expense, delta and nullable basis-point change when baseline is zero |
| `get_wallet_balances` | optional `wallet_ids` | current owner-wallet balances from opening balance plus full ledger |
| `get_budget_progress` | optional `budget_ids`, optional `active_on` | existing budget calculation and scope |
| `get_goal_progress` | optional goal wallet IDs | goal wallet current/target/remaining view |
| `get_jar_progress` | `month`, optional `jar_ids` | SELECT-only jar allocation/spend; does not initialize month/config rows |
| `get_product_help` | `{ "query": "...", "max_results": 1..5 }` | versioned handbook excerpts with stable source IDs; no owner data or external URLs |
| `propose_create_wallet` | required `name`, `type` (`basic`, `goal`, `credit`); optional balance/description/total flag and type-specific goal/credit fields | validated `wallet_creation_proposal`, status `proposed`; creates no wallet |
| `propose_create_jar` | required `month` (`YYYY-MM`), `name`, `allocation_mode` (`none`, `fixed`, `percent`); mode-specific amount/percent | validated `jar_creation_proposal`, status `proposed`; creates no jar |
| `propose_update_jar` | required UUID `jar_id`, `month`, `name`, `allocation_mode`; mode-specific amount/percent | validated `jar_update_proposal`, status `proposed`; replaces no config until explicit confirmation |
| `propose_create_budget` | required `name`, positive `limit_amount`, `start_date`, `end_date`; optional owner-scoped wallet/category UUIDs | validated `budget_creation_proposal`, status `proposed`; creates no budget |
| `propose_update_budget` | required UUID `budget_id`, `name`, positive `limit_amount`, `start_date`, `end_date`; optional owner-scoped wallet/category UUIDs | validated `budget_update_proposal`, status `proposed`; replaces no budget until explicit confirmation |

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
Capabilities advertise `write:false` and `actions:["create_wallet","create_jar","update_jar","create_budget","update_budget"]`: this means
client-confirmable proposal, not a write tool. A mounted native confirmation card
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

The provider-facing JSON Schema mirrors the Go decoder. `range` and `current`/`previous` require `{ "from": "YYYY-MM-DD", "to": "YYYY-MM-DD" }`; `month` requires `YYYY-MM`; array filters are bounded to 100 IDs; strings and cursors have explicit maximum lengths. The adapter serializes assistant calls as `{ "id", "type": "function", "function": { "name", "arguments" } }` and preserves `tool_call_id` on tool results. Current loop permits at most eight tool executions and six provider requests per run. After eight tools, the next permitted provider request omits tools; a further tool request or exhausted provider-call budget returns a loop-limit error. These are run-safety bounds, not per-user usage quotas.

Credential validity is checked before each provider request, before each tool execution, and again before persisting the assistant answer. Revoked/expired API keys and invalid sessions fail closed. `POST /runs/{id}/cancel` updates the run and interrupts the in-process provider context. The explicit reset service also interrupts queued/running contexts purged by its transaction; a replayed reset never cancels newer work. The approved 2026-10-15 conversation→run lock/recheck and idle-cleanup worker are deployed in the current API/worker image; the controlled-lock late-insert regression passes (realDB gate4/4). This does not establish universal cross-process cancellation atomicity. Synchronous submit still returns `run_id` only after provider work, leaving the native initial wait without its Stop action. [Cancellation audit](../work/tickets/NATIVE-ADVISOR-CANCEL-20261015.md). A multi-process deployment still needs durable cancellation pub/sub before claiming immediate cross-instance interruption.

Known response views include `finance_summary`, `spending_comparison`, `transaction_search`, `transaction_detail`, `wallet_balances`, `budget_progress`, `goal_progress`, `jar_progress`, `product_help`, `wallet_creation_proposal`, `jar_creation_proposal`, `jar_update_proposal`, `budget_creation_proposal` and `budget_update_proposal`. Unknown or malformed views use a textual fallback, masked before formatting in the local corrected native renderer. Device/full-app acceptance remains separate. No model-provided markup is rendered.

## Implemented local routes

Advisor routes accept JWT/session or a full-owner user API key. There are no per-key scopes; the tool allowlist and confirmation boundary remain independent of bearer authority. The routes are mounted in production behind auth; owner-authenticated UAT is still required.

- `GET /api/v1/ai/advisor/capabilities` (advertises `get_product_help` as a read-only tool)
- `GET /api/v1/ai/advisor/overview?month=YYYY-MM` (real report summary; no model call)
- `POST /api/v1/ai/advisor/messages`
- `GET /api/v1/ai/advisor/conversation/messages?conversation_id=...&before_seq=...` (last 50 by default; pass the oldest loaded sequence to page backwards). The existing `data` array may include additive `meta` lifecycle fields: `conversation_id`, `generation`, `last_user_message_at`, `expires_at` and server `server_time`; reads never extend the deadline.
- `POST /api/v1/ai/advisor/conversation/reset` with a required `Idempotency-Key` (clears only advisor content, rotates generation, and returns a content-free reset result)
- `GET /api/v1/ai/advisor/runs/{id}`
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

All listed routes use the same owner credential authority. Create/list/revoke keys through `/api/v1/api-keys` with JWT or owner key; secret is shown once and only its hash persists. See [full-key guide](../guides/api-keys.md). Advisor auth still audits best-effort to Redis stream `mypocket:audit:advisor` with request/credential ID, route, status, decision and latency; no prompts, notes, tokens or monetary rows. `POST /messages` waits for the bounded provider/tool loop and returns JSON. SSE, lost-POST recovery, conversation clear, fact-bundle drilldown and durable audit delivery remain separate gates.

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

The orchestrator prepends an embedded versioned product handbook (`v1.3`,
2026-10-15) with navigation, explicit jar-proposal capability and
non-mutating tool boundaries. The
`get_product_help` read-only tool now searches a small source-controlled section
catalog and returns version/source IDs plus bounded excerpts; no external
retrieval, owner data or arbitrary URLs are accepted. The handbook describes
`Trợ lý` as the primary native tab and keeps only secondary destinations under
`Tài khoản`; the provider regression rejects the former Assistant-in-More
wording. A no-match query returns `not_found` rather than dumping the handbook.
This is deterministic local retrieval, not a general semantic index or live
RAG service; real-provider answer quality and production gates remain separate.
