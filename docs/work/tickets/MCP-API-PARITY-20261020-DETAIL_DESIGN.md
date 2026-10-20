---
artifact_type: detail_design
id: MCP-API-PARITY-20261020
status: verified
owner: shared
approval: owner approved full implementation on 2026-10-20 and selected the official Go SDK
updated: 2026-10-20
---

# MCP/API parity adapter — detail design

The [final-review correction wave](../bugs/MCP-FINAL-REVIEW-20261020.md)
reconciles optional feedback PNG multipart/resource-link variants, canonical
audit delivery through the application router, and the committed category
receipt dependency. Its isolated candidate verification supersedes reliance on
uncommitted shared-worktree dependencies in the original aggregate gate.

## Approved transport decision

The implementation uses the official Go SDK package
`github.com/modelcontextprotocol/go-sdk/mcp`. The SDK owns protocol
negotiation, JSON-RPC envelopes, sessions and Streamable HTTP semantics. The
MyPocket code owns authentication, the named operation catalog, canonical REST
dispatch, owner isolation, audit redaction and response mapping. Hand-written
MCP/JSON-RPC transport code is out of scope.

## Intent

The product decision for this work is strict capability parity: every
operation available to an authenticated user through JWT must also be
available to the same owner through a validated `mpk_` API key and through the
MCP adapter. The adapter must not create a second permission model. It must
preserve the same owner derivation, validation, compare-and-swap, idempotency,
confirmation, and error behavior as the canonical REST operation.

This is an integration boundary for personal agents and automation. It is not
a new autonomous-finance policy. The in-app Assistant's proposal cards remain
an existing product UX: ordinary MCP calls may invoke ordinary user writes
directly, while Assistant prepared actions still require the existing
conversation/generation-bound confirmation contract.

## Repository context inspected

- `backend/internal/controller/http/middleware.go`: `RequireAuth`,
  `RequireAdvisorAuth`, API-key/session principal context, advisor audit.
- `backend/internal/usecase/user_api_key.go`: API-key digest, expiry,
  revocation and owner validation.
- `backend/internal/controller/http/router.go` and every `routes_*.go` file:
  authenticated route inventory.
- `backend/internal/controller/http/ai_advisor_handler.go`: Assistant routes,
  persisted action confirmation and SSE replay.
- `backend/internal/usecase/ai_advisor_tools.go`: current model tool registry;
  proposal tools are not direct write tools.
- `backend/internal/usecase/finance_query.go`: owner-scoped finance query
  boundary.
- `backend/cmd/api/main.go`: production route composition and principal
  revalidation wiring.
- `backend/docs/swagger.json`: checked-in 2026-10-20 route/method catalog.
- `docs/architecture/API.md`, `docs/architecture/AUTH_API.md`,
  `docs/guides/api-keys.md`, and `docs/architecture/AI_ADVISOR_TOOLS.md`.

## Capability boundary

### In scope: authenticated user operations

These REST groups are user operations and therefore must remain reachable with
JWT, an equivalent same-owner API key, and MCP:

| Domain | REST groups and operation families |
| --- | --- |
| Account | `/api/v1/auth/profile`, `/api/v1/home`, API-key create/list/revoke, device enrollment/revoke, logout |
| Categories | list/create/update/status/wallet applicability/delete |
| Wallets | list/create/update/delete |
| Transactions | list/create/update/delete, adjustment, bulk-delete, paired transfer create/update/delete, travel link/unlink, attachment download |
| Budgets | CRUD, budget report and scoped transactions |
| Budget recurrence | series list/create/update/stop |
| Jars and months | jar CRUD/month configuration/report, month summary/note save/delete |
| Recurring transactions | list/create/update/delete/run-due |
| Credit | credit entry, credit payment, credit statement |
| Travel | event CRUD and activate/deactivate |
| Portfolio | summary, asset CRUD/price, trades, correction preview/commit |
| Reports | savings, savings transactions, insider |
| Notifications | list/read, push-token register/revoke |
| Feedback | authenticated owner list/create/get/screenshot/resolution |
| AI entry | capabilities/process/request recovery/proposal edit/approve/reject |
| Finance Assistant | capabilities/overview/history/reset/messages/async run/status/events/cancel/action confirmation |
| Sync | authenticated sync capabilities |

The complete method/path inventory is the checked-in `backend/docs/swagger.json`
and must be machine-tested against the catalog. The adapter must not silently
drop a newly authenticated route when the REST surface grows.

### Explicitly out of user parity

These are not authenticated JWT user operations and must not be granted to an
MCP bearer merely because it has an API key:

- `POST /api/v1/login`, refresh, device-login, Google start/callback/mobile
  exchange: these are credential-establishment or exchange operations, not
  operations performed by an already authenticated user. MCP uses an existing
  JWT or `mpk_` key.
- `/api/v1/agent/*` and `/api/v1/internal/*`: dedicated feedback-agent service
  token boundary; never accept user API keys or MCP user credentials.
- Public changelog reads and health/docs are public, not owner capabilities.

These exclusions are protocol/security classifications, not extra user API-key
scopes. If a later requirement asks an agent to bootstrap login, that must be a
separate credential-exchange design with an explicit threat review.

## Decision and alternatives

### Recommended: canonical operation descriptors with in-process REST dispatch

Add a small operation catalog that describes every user REST operation once:

```go
type ExternalOperation struct {
    Name             string
    Method           string
    PathTemplate     string
    InputSchema      map[string]any
    BodyEncoding     string // json | query | multipart | empty
    Idempotency      string // none | optional | required
    Async            bool
    SensitiveResult  bool
}
```

Each descriptor has a typed request encoder and invokes the already-registered
Gin route in-process with the authenticated bearer credential. The existing
handler remains the authority for domain validation, owner lookup, CAS,
idempotency and mutation semantics. The adapter only translates SDK-decoded
tool arguments into an HTTP request and translates the HTTP response into an
SDK `CallToolResult`. It must use an allowlisted descriptor, never a
caller-supplied path.

This avoids duplicating dozens of handlers while preserving ordinary REST
writes. The in-process dispatcher uses the same `Router.Engine.ServeHTTP`
composition as production, so API-key and JWT auth execute again on the inner
request. The outer MCP request still authenticates first for audit and early
failure.

### Rejected: direct calls to repositories/usecases from MCP

This would be faster for the first few tools but would duplicate handler
validation and response mapping. It would create a high-risk split where a
wallet or transaction write behaves differently through MCP than REST. Direct
usecase adapters may be introduced later only for a measured hot path after
contract parity tests exist.

### Rejected: generic `call_rest(method, path, body)` MCP tool

It exposes an arbitrary HTTP surface, complicates path/body authorization,
allows accidental access to internal routes, and makes tool schemas unusable
for agents. Named descriptors are required for auditability and fail-closed
route growth.

### Rejected: separate MCP scopes or MCP-only confirmation

The user decision explicitly forbids new scopes. A user API key has the same
owner authority as that owner's JWT. MCP must not add a second approval policy.
The one exception is an existing operation's own contract: Assistant prepared
actions remain server-persisted and confirmation-bound because that is the
contract of that Assistant endpoint, not an MCP-specific restriction.

## MCP transport contract

Phase 1 mounts the official SDK's Streamable HTTP handler at `/api/v1/mcp`
with `Authorization: Bearer <JWT-or-mpk-secret>`. The SDK implements
`initialize`, session/protocol negotiation, `notifications/initialized`,
`tools/list`, and `tools/call`. Compatibility is verified with the matching SDK
client and conformance-focused HTTP tests. MyPocket must not hand-parse or
hand-emit MCP JSON-RPC envelopes.

`tools/list` returns named user operations from the descriptor catalog. Every
write is listed; it may carry descriptive annotations such as `readOnlyHint`
or `destructiveHint`, but annotations do not change authorization or require a
new confirmation step.

`tools/call` arguments are a JSON object. Reserved adapter metadata is kept out
of business JSON:

```json
{
  "name": "transactions.create",
  "arguments": {
    "wallet_id": "...",
    "category_id": "...",
    "type": "expense",
    "amount": 120000,
    "occurred_at": "2026-10-20T10:00:00+07:00"
  },
  "_meta": {
    "idempotency_key": "stable-client-key"
  }
}
```

The exact MCP wire envelope follows the selected protocol version; the
business mapping above is stable. The adapter maps `_meta.idempotency_key` to
the HTTP `Idempotency-Key` header and rejects a duplicate key supplied in the
business body. It never invents a retry key or retries a failed write.

The adapter returns structured content with the original HTTP status, response
body, and request ID in metadata. A domain `Problem` remains machine-readable:
`status`, `code`, `detail`, and `request_id` are retained. Provider prompts,
OCR text, SQL, credentials and raw financial payloads are never added to MCP
errors or audit records.

## Owner and credential flow

```text
MCP client
   |
   | Bearer JWT or mpk_ API key
   v
MCP outer auth + audit
   |
   v
descriptor allowlist -> in-process REST request
   |
   v
RequireAuth -> owner Principal -> canonical handler/usecase/repository
```

- `RequireAuth` remains the only owner resolver for ordinary operations.
- `OwnerID` comes from the validated JWT session or persisted API-key row; any
  `owner_id`/`user_id` argument is rejected as an unknown field or ignored only
  where the canonical handler already treats it as unrelated metadata. The MCP
  encoder must explicitly reject it before dispatch to make impersonation
  attempts obvious.
- A user API key is checked for digest, owner, expiry and revocation on every
  inner request. API keys do not refresh into JWTs.
- Assistant async runs revalidate credentials through the existing
  `AdvisorPrincipalValidator` before provider calls, tool calls and answer
  persistence. The adapter must not call `AdvisorTool.Execute` directly for a
  conversational run without that validator.
- Device credentials, password confirmation and Face ID remain separate local
  client concerns; MCP receives an already authenticated bearer.

## Operation mapping rules

Descriptor names are stable, lowercase, domain-qualified names. They map to
existing REST routes as follows (the implementation plan expands this into a
machine-tested table):

| MCP namespace | Examples | Encoding/notes |
| --- | --- | --- |
| `account.*` | `account.profile.get`, `account.profile.update_timezone`, `account.api_key.create`, `account.api_key.revoke`, `account.logout` | JSON or empty; logout keeps current key/JWT semantics |
| `categories.*` | `list`, `create`, `update`, `set_status`, `set_wallets`, `delete` | JSON/path; preserve category validation |
| `wallets.*` | `list`, `create`, `update`, `delete` | JSON/path; idempotency header passed through |
| `transactions.*` | list/create/update/delete/adjustment/bulk-delete/transfer/travel | JSON/query/path; preserve pair atomicity and idempotency |
| `budgets.*` | CRUD/report/transactions | JSON/query/path |
| `budget_series.*` | list/create/update/stop | JSON/path; version/idempotency unchanged |
| `jars.*` and `months.*` | list/create/month update/report/remove, summary/note | JSON/query/path |
| `recurring.*` | CRUD/run_due | JSON/path; existing occurrence cap/idempotency remains |
| `credit.*` | entry/payment/statement | JSON/query/path; credit-specific pairing remains |
| `travel.*` | event CRUD/activate/deactivate | JSON/path |
| `portfolio.*` | summary/assets/price/trades/correction preview/commit | JSON/path; asset-version guards remain |
| `reports.*` | savings/savings transactions/insider | query; owner filters remain server-validated |
| `notifications.*` | list/read/device register/revoke | JSON/path |
| `feedback.*` | owner list/create/get/screenshot/resolution | JSON/path; never agent/internal feedback |
| `ai_entry.*` | capabilities/process/request/proposal edit/approve/reject | multipart for process, JSON/path otherwise |
| `advisor.*` | capabilities/overview/history/reset/messages/async/run/events/cancel/confirm | JSON/query/path; see async/confirmation rules |
| `sync.*` | capabilities | query/empty |

## Special flows

### Ordinary writes, idempotency and CAS

The adapter passes `Idempotency-Key` exactly when supplied. It does not change
legacy endpoint behavior: a route that currently permits an unkeyed online
write remains callable unkeyed; an endpoint that requires a key rejects the
MCP call with the same 400/409/503 behavior. The client, not the server, owns
retry-key stability.

The adapter must not coalesce calls. Transfer create/update/delete, budget
series changes, wallet/category durable creates, AI-entry approval, and any
other receipt-backed operation retain the canonical ledger key and payload
hash. Transaction/category/portfolio updates retain `expected_version` CAS;
stale writes return the existing 409 without a mutation.

### Finance Assistant

`advisor.messages` and `advisor.messages_async` call the existing endpoints,
not the registry directly. The returned run/conversation IDs are opaque. MCP
clients use `advisor.run.get`, `advisor.run.events`, and `advisor.run.cancel`.
Events are finite persisted replay; the current server does not claim token
streaming. `advisor.action.confirm` accepts only `action_id`, conversation ID
and generation, preserving server-owned proposal data and existing idempotency.

### AI entry and files

`ai_entry.process` accepts an adapter-defined `files` array with filename,
MIME type and base64 bytes. The encoder builds the existing multipart request
and enforces the same 20-file, 5 MiB/file and 101 MiB request limits before
dispatch. No OCR/model API key is exposed to the MCP client. The adapter does
not resubmit after an ambiguous response; clients poll
`ai_entry.request.get` using the same request ID.

`transactions.attachment.download` follows the canonical 302 result and
returns an MCP resource link with the same short-lived private URL. It must
redact the URL from logs and never persist it. If the selected MCP client does
not support resource links, the response includes the documented direct REST
fallback rather than silently claiming binary support.

### Async and SSE compatibility

The initial operation adapter treats business operations as request/response.
Async operations return a stable handle and status. The replay-only advisor
event endpoint is normalized to a bounded array of events with `after_seq`;
this does not promise model-token streaming. SDK transport sessions or SSE do
not change the advisor endpoint's replay-only business semantics.

## Audit and privacy

Add a dedicated MCP audit wrapper using the existing `repository.AuditSink`
and Redis/PostgreSQL outbox path. Event fields are: request ID, credential ID,
credential kind, operation name, HTTP status, allowed/denied decision, reason,
latency and bounded request/response byte counts. Do not record MCP arguments,
multipart bytes, base64, signed attachment URLs, prompts, OCR, notes, balances,
tokens or provider usage.

`RequireAdvisorAuth` currently recognizes only the `/api/v1/ai/advisor` prefix
for advisor audit. The MCP wrapper must therefore emit `mcp.*` events itself;
it must not assume an advisor route prefix will audit MCP calls.

## Error, limits and compatibility contract

- Protocol parse/invalid-request/unknown-tool errors are emitted by the SDK;
  MyPocket tests their observable MCP behavior without duplicating its logic.
- A dispatched REST response keeps the original HTTP status and MyPocket
  `Problem` code in structured MCP metadata.
- Authentication and revocation failures are 401; domain ownership/not-found
  behavior remains the route's current 404 behavior; validation remains 400;
  conflicts/CAS/idempotency remain 409; provider/runtime failures preserve
  503. The adapter must not normalize all failures to a generic MCP error.
- Bound outer HTTP request bytes, catalog argument schemas, base64 file bytes
  and tool output bytes around the SDK handler. Inner route limits remain
  authoritative. Use context deadlines; never detach a caller's cancellation
  for synchronous writes.
- No automatic retries at the adapter layer. Client retry rules come from the
  endpoint's existing contract and must preserve request IDs/Idempotency-Keys.

## Risks and decisions requiring explicit review

1. In-process dispatch preserves behavior but couples MCP descriptors to Gin
   route composition. The implementation must fail startup/tests if a
   descriptor points at a missing route.
2. A named catalog must be updated when a new authenticated REST route ships;
   parity tests prevent silent omission.
3. MCP resource links expose the same short-lived attachment URL as REST; logs
   and tracing must redact it, and TTL/owner checks remain storage-owned.
4. Full-owner API keys are intentionally powerful. No MCP-only scope, quota or
   approval policy is introduced in this design.
5. Login/refresh/device bootstrap and service-token agent routes remain
   explicit non-parity gaps. They cannot safely be made MCP operations without
   a credential-bootstrap threat model.

## Verification and reconciliation

### Local implementation checkpoint — 2026-10-20

Tasks 1–9 are implemented and locally verified. Official SDK v1.8.0 owns
stateless JSON Streamable HTTP; `NewApplicationRouter` is the canonical route
assembly used by production and parity tests. Runtime discovery, independent
inventory and generated OpenAPI agree on 101 tools and 19 exclusions. Existing
JWT/API-key owner authority, direct writes, CAS/idempotency and Assistant
confirmation remain canonical-handler contracts.

Fresh disposable PostgreSQL 17/Redis 7 verification uses schema 40 clean:
full backend and full backend race each pass 727 tests/subtests (11 opt-in live
provider/storage skips); focused MCP/router/audit race passes 178 with zero
skips. Docs/route gate passes 9, regeneration produces no artifact drift, vet,
module verification and format/diff checks pass. The final injection regression
requires exact invalid-arguments status/code, zero canonical dispatch and
unchanged persisted wallet/transaction state. Audit byte counts survive real
Redis/outbox replay. Transfer tests prove observed pair state and replay, not
a forced rollback after a second-insert storage failure.

Human API/auth/key/advisor docs, agent-readable Markdown/runtime schemas,
generated Swagger, ADR-025, changelog, architecture, context, backlog, validation
and rollout readiness are reconciled. No new database schema or native/web UI
flow was introduced; ERD and screen docs need no adapter-specific change.
`docs/architecture/SDD.md` is absent; ARCHITECTURE/API and ADR-025 are the
maintained architecture surfaces. See [validation](../VALIDATION_MATRIX.md)
and [rollout readiness](../../operations/ROLLOUT_READINESS.md).

Status means local implementation verified. Owner-authenticated staging
JWT/API-key initialize/list/read/write, revoked-key, replay/CAS, multipart and
advisor async/cancel acceptance remain open. No deployment, production access,
provider quality or physical-device acceptance occurred in Task 9.

Required proof is listed in the companion plan:
`docs/superpowers/plans/2026-10-20-mcp-api-parity.md`.

The implementation reconciles the human API/auth/API-key/AI-entry
guides, the machine-readable endpoint catalog/OpenAPI or `llms` surface, and
the validation matrix. ADR-025 records the authenticated transport and durable
parity boundary; shared documentation reconciliation is recorded above.
