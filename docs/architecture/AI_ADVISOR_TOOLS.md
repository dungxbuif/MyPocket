# Finance Assistant V1 — tool catalog

Status: local JWT/API-key pilot, synchronous JSON transport. This catalog describes the implemented read-only tool boundary; public deployment, streaming and audit readiness are still separate gates.

## Security boundary

- The model never receives SQL, database credentials, arbitrary URLs, or a caller-supplied owner ID.
- The backend supplies the authenticated `Principal.OwnerID` and account timezone to every tool.
- Every tool is read-only. No `create_transaction`, `update_transaction`, `delete_transaction`, `execute_sql`, bank, payment, or confirmation tool is registered.
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

## Strict argument rules

Arguments are bounded JSON objects decoded with unknown fields rejected. IDs are trimmed/deduplicated and resolved under the authenticated owner. Calendar ranges are inclusive account-local dates normalized to UTC half-open instants; ranges longer than twelve months are rejected. Money is integer VND and must remain within the JavaScript-safe integer bound.

The provider-facing JSON Schema mirrors the Go decoder. `range` and `current`/`previous` require `{ "from": "YYYY-MM-DD", "to": "YYYY-MM-DD" }`; `month` requires `YYYY-MM`; array filters are bounded to 100 IDs; strings and cursors have explicit maximum lengths. The adapter serializes assistant calls as `{ "id", "type": "function", "function": { "name", "arguments" } }` and preserves `tool_call_id` on tool results. After four tool executions the next provider request omits tools so the model can only produce the final grounded answer.

Credential validity is checked before each provider request, before each tool execution, and again before persisting the assistant answer. Revoked/expired API keys and invalid sessions fail closed. `POST /runs/{id}/cancel` updates the run and interrupts the in-process provider context; repository writes accept assistant messages only while the run is queued/running, so a cancelled run cannot commit a late answer. A multi-process deployment still needs durable cancellation pub/sub or worker ownership before claiming immediate cross-instance interruption.

Known response views are rendered by the shared frontend card registry: `finance_summary`, `spending_comparison`, `transaction_search`, `transaction_detail`, `wallet_balances`, `budget_progress`, `goal_progress` and `jar_progress`. Unknown or malformed views remain a safe muted fallback; no model-provided markup is rendered.

## Implemented local routes

Advisor routes accept either the existing JWT/session or an owner-scoped user API key. API keys must have `finance:read` plus the route-specific scope:

- `GET /api/v1/ai/advisor/capabilities`
- `GET /api/v1/ai/advisor/overview?month=YYYY-MM` (real report summary; no model call)
- `POST /api/v1/ai/advisor/messages`
- `GET /api/v1/ai/advisor/conversation/messages?conversation_id=...&before_seq=...` (last 50 by default; pass the oldest loaded sequence to page backwards)
- `GET /api/v1/ai/advisor/runs/{id}`
- `POST /api/v1/ai/advisor/runs/{id}/cancel`

`GET` capabilities/overview/history/run use `advisor:read`; `POST /messages` and cancel use `advisor:chat`. Create/list/revoke keys through the JWT-only `/api/v1/api-keys` endpoints; the secret is returned only on creation and only its hash is persisted. Advisor auth decisions are written best-effort to Redis stream `mypocket:audit:advisor` with request ID, credential type/ID, route, status, decision and latency; prompts, notes, tokens and monetary rows are excluded. `POST /messages` currently waits for the bounded provider/tool loop and returns JSON. SSE events, lost-POST recovery, conversation clear, fact-bundle drilldown, durable audit delivery/alerting and public deployment are not yet release claims.

The web panel loads the newest 50 messages, exposes “Tải tin cũ” when another page may exist, requests `before_seq` using the oldest loaded message and merges pages by message ID in chronological order. Reloading never clears the current list; duplicate boundary rows are collapsed client-side.
