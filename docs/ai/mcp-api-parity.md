---
artifact_type: integration_contract
id: MCP-API-PARITY
status: active
owner: shared
updated: 2026-10-20
---

# Connect a personal agent through MCP

The local candidate exposes `/api/v1/mcp` using the official Go SDK
`github.com/modelcontextprotocol/go-sdk/mcp` v1.8.0. A validated owner JWT and an
`mpk_` API key have the same user authority through REST and MCP. There are no
MCP scopes, read-only keys, or extra MCP confirmation steps. Give a key only to
an integration you trust to change your account, create/revoke keys and log out.
This document describes checked-in code; deployment and owner-authenticated
staging/provider/device acceptance remain separate gates.

## Connection and discovery

Configure a Streamable HTTP client with your server origin plus `/api/v1/mcp`
and the header `Authorization: Bearer <JWT-or-mpk-secret>`. Never place a secret
in the URL. Use the SDK's negotiation/initialize flow, then `tools/list` and
`tools/call`. The SDK owns JSON-RPC framing and protocol negotiation.

The adapter uses `Stateless: true` and `JSONResponse: true`. POST handles the
protocol; GET and DELETE are mounted behind authentication for SDK transport
semantics and normally return `405 Allow: POST`, not a live event stream or a
persistent server session. Clients can disable standalone SSE. Calls need
`Content-Type: application/json` and an Accept header supporting both
`application/json` and `text/event-stream`; let the client SDK set negotiated
protocol headers. An Origin header must exactly match a configured allowed
origin; non-browser clients without Origin are accepted.

The machine-readable catalog is the authenticated `tools/list` result, not a
second static list of schemas in Swagger. Each entry publishes its stable
domain-qualified `name`, a description containing its canonical `METHOD /path`,
and an `inputSchema` with a closed object (`additionalProperties: false`). Path
parameters are required fields; handler business validation still applies.
The current tested inventory is 101 named tools and 19 explicitly excluded
method/path entries. New routes must update the catalog or exclusion manifest;
route parity tests fail on unclassified growth.

This JSON locator is consumed by the documentation contract tests. Exclusions
are route classifications, not permission scopes:

```json
{
  "endpoint": "/api/v1/mcp",
  "transport": "streamable_http_stateless_json",
  "authentication": "Authorization: Bearer <JWT-or-mpk-secret>",
  "authority": "same_owner_user_operations_no_scopes",
  "catalog_method": "tools/list",
  "schema_field": "inputSchema",
  "tool_count": 101,
  "excluded_operations": [
    "POST /api/v1/login",
    "POST /api/v1/auth/refresh",
    "POST /api/v1/auth/device-login",
    "GET /api/v1/auth/google",
    "GET /api/v1/auth/google/callback",
    "POST /api/v1/auth/mobile/google/exchange",
    "GET /api/v1/health",
    "GET /api/v1/changelog",
    "GET /api/v1/changelog/{id}",
    "GET /api/v1/agent/feedback",
    "GET /api/v1/agent/feedback/{id}/screenshot",
    "PATCH /api/v1/internal/feedback/{id}/status",
    "POST /api/v1/internal/changelog",
    "GET /api/v1/mcp",
    "POST /api/v1/mcp",
    "DELETE /api/v1/mcp",
    "GET /api/v1/docs/*any",
    "GET /docs",
    "GET /docs/"
  ]
}
```

Login/refresh/device-login/Google exchange establish credentials outside MCP.
The `/agent/*` and `/internal/*` feedback routes require their dedicated service
token and are never granted by user credentials. Health, changelog and docs are
public surfaces; the MCP endpoint itself never becomes a recursive tool.

## Calls, owner binding and writes

Named descriptors encode JSON, query, path or multipart fields and dispatch
in-process through the canonical Gin router. The original bearer is validated
again on the inner REST request. Owner identity comes from the validated JWT
session or persisted key row; nested or top-level `owner_id`/`user_id` selectors
(including spelling variants) are rejected. Clients cannot choose a method,
host or arbitrary path; there is no `call_rest` tool.

Ordinary tools such as `wallets.create`, `transactions.create`,
`transactions.transfer.create` and `budgets.create` write directly under the
ordinary endpoint contract. The Finance Assistant's `propose_*` model registry
is a different interface: `advisor.messages`/`advisor.messages_async` produce
persisted proposals, and `advisor.action.confirm` receives only the server action
ID plus conversation ID and generation. It resolves the server-owned payload;
a client cannot replace it. AI-entry edit/review/approve/reject contracts are
also preserved. Device enrollment still requires the account password.

Example MCP request after negotiation (all credentials and IDs are placeholders):

```http
POST /api/v1/mcp
Authorization: Bearer <JWT-or-mpk-secret>
Content-Type: application/json
Accept: application/json, text/event-stream

{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"wallets.create","arguments":{"name":"Example wallet","type":"basic","opening_balance":0},"_meta":{"idempotency_key":"example-wallet-create-1"}}}
```

`params._meta.idempotency_key` maps to `Idempotency-Key`; it is not a business
argument. Application metadata accepts only that key (1–128 printable ASCII
characters); SDK protocol metadata remains SDK-owned. An `idempotency_key` or
`_meta` business field is rejected. The adapter never generates keys, coalesces
requests or automatically retries writes. Keep the same key and payload on a
retry only where the canonical operation supports receipts. Supplying a key
does not add receipt support to an operation that lacks it.

Required-key operations currently include transfer update/delete, budget-series
create/update/stop, portfolio correction commit and advisor reset. Optional
receipt support exists for wallet/category/budget/jar/transaction/transfer
creates. Unkeyed legacy writes retain their canonical behavior and are unsafe
to automatically retry. Domain validation and required-key rejection remain
the REST handler's responsibility. See [offline receipts](../guides/offline-sync-api.md).
Updates preserve `expected_version` and existing compare-and-swap rules; a stale
snapshot returns `409` without a write. Do not invent a version after conflict:
fetch current state and explicitly prepare a new change. Transfer pairs remain
atomic and portfolio corrections retain their append-only revision contract.

## Results, errors and limits

Successful and dispatched error results preserve the REST body:

```javascript
// result.structuredContent; also present as JSON in text content
{status: 200, body: {data: []}, content_type: "application/json", request_id: "<request-id>"}
// result._meta repeats status, content_type and request_id
```

`isError` is true for failed operations. Inspect structured `status` and
`body.code`, not only the outer HTTP status of the MCP envelope. A canonical
Problem retains `status`, `code`, `detail` and `request_id`; successful JSON
numbers are preserved without float conversion. Invalid bearer credentials
fail outer HTTP authentication with `401` before protocol handling. Unknown
tools and protocol errors are SDK errors. Dispatched owner/not-found stays
`404`, validation `400`, stale CAS/idempotency `409`, and provider/runtime
failure retains its canonical status (commonly `503`).

| Adapter condition | Structured status / code |
| --- | --- |
| Invalid adapter arguments | `400 MCP_INVALID_ARGUMENTS` |
| Output exceeds bound | `502 MCP_RESPONSE_TOO_LARGE` |
| Invalid attachment redirect | `502 MCP_ATTACHMENT_UNAVAILABLE` |
| Invalid/oversized event replay | `502 MCP_INVALID_EVENT_REPLAY` |
| Operation deadline | `504 MCP_DEADLINE_EXCEEDED` |
| Caller cancellation, if a result can still be delivered | `408 MCP_REQUEST_CANCELLED` |

Default bounds are 140 MiB outer request bytes (base64 overhead included),
8 MiB canonical response bytes and two minutes per tool operation. Options can
set tighter/different bounds. Oversized outer bodies return HTTP `413`;
disallowed Origin returns HTTP `403`. Inner route limits remain authoritative.
The output cap rejects the result without returning a partial financial body.
Synchronous dispatch follows the caller's cancellation context, including older
negotiated SDK protocols. A timeout can leave acceptance uncertain: consult the
canonical receipt/recovery route before considering another write. Transport
bounds are not per-owner AI quotas and do not expand authorization.

## AI entry, async work and attachments

`ai_entry.process` accepts `request_id`, `text`, `timezone` and an ordered
`files` array. Each file has `filename`, `mime_type` and `base64` (standard
base64 bytes). The adapter converts it to the existing multipart endpoint.
Accepted detected MIME types are JPEG, PNG and PDF; the declared type must match
the bytes. Filenames are bounded and cannot contain path separators. Limits
are 20 files, 5 MiB decoded bytes per file and 101 MiB encoded multipart body;
the 140 MiB outer JSON limit also applies. No provider secret is sent to the
client. After ambiguous processing acceptance, retain the original request ID
and poll `ai_entry.request.get`; do not automatically resubmit or generate a
new request ID.

`feedback.create` keeps JSON text-only feedback and also accepts an optional
`screenshot: {filename, mime_type, base64}` object. All three screenshot fields
are required when supplied; `mime_type` must be `image/png` and match detected
bytes. The screenshot must be nonempty and at most 1 MiB decoded (exactly
1 MiB is accepted), the filename is 1–255 bytes without path separators, and
the entire canonical multipart body is at most 2 MiB. The adapter sends the
existing `type`, `title`, `description` and `screenshot` fields. Text validation,
private storage and authenticated owner binding remain canonical. REST OpenAPI
publishes both accepted content types and its `x-multipart-form` file variant;
MCP `tools/list` publishes the closed screenshot input schema.

Accepted asynchronous results can add `structuredContent.async`:

| Flow | Handle and next tools |
| --- | --- |
| `advisor.messages_async` | `kind: advisor_run`, opaque `id`, `status`, `conversation_id`, `poll_tool: advisor.run.get`, `events_tool: advisor.run.events`, `cancel_tool: advisor.run.cancel` |
| AI entry processing | `kind: ai_entry_request`, opaque `id`, `processing`, `poll_tool: ai_entry.request.get` |

`advisor.run.events` takes `id` and optional integer `after_seq`. It normalizes
finite persisted SSE replay into `{events, after_seq, replay_only: true}`, with
at most 100 strictly increasing sequenced events per response. Continue using
the returned cursor; an invalid or oversized page fails atomically. This is not
live model-token streaming. Cancelling/resetting runs and confirming actions
keep their owner/conversation/generation semantics; advisor credentials are
revalidated at provider/tool/persistence boundaries.

`transactions.attachment.download` performs the same transaction-owner and
approval-link checks as REST. Its successful canonical `302` becomes a non-error
MCP result with a resource link to the short-lived private URL (currently five
minutes). Signed redirect HTML is discarded. The body supplies
`rest_fallback: {method: "GET", path: "/api/v1/transactions/<id>/attachments/<attachment-id>/download"}`
so clients without resource-link support can use authenticated REST and renew
access. Do not persist or log the signed link; MCP does not proxy binary bytes.

`feedback.screenshot` applies the same success conversion after canonical
feedback-owner checks: `302` is a non-error ResourceLink, with
`rest_fallback: {method: "GET", path: "/api/v1/feedback/<id>/screenshot"}`.
Missing/foreign screenshots retain `404`. Only the ResourceLink carries the
signed URL; structured/text content never includes the redirect HTML or URL.

## Credential lifecycle and audit privacy

API keys are checked for digest, valid owner, expiry and revocation on every
request; a key-looking invalid bearer never falls back to JWT. JWTs retain
Redis session/family validation. Keys do not refresh into sessions. Through
`account.logout`, an API key revokes itself while JWT logout requires its
refresh token and revokes its session family. See [API keys](../guides/api-keys.md)
and [authentication](../architecture/AUTH_API.md).

With an audit sink configured, tool calls enqueue dedicated `mcp.<operation>`
events through a nonblocking best-effort queue (64 pending events and at most
one drain worker per handler). Full queues drop events; a stalled sink cannot
stall these audit deliveries' business results. In-process MCP advisor calls
still authenticate normally, but skip the duplicate synchronous REST access
audit because the outer queue records `mcp.advisor.*`. A private request context
marker controls this; headers or tool arguments cannot suppress REST auditing.
The application factory copies the feedback handler/service and uses a separate
64-event, one-worker queue for MCP `feedback.created` and
`feedback.owner_resolution` domain events. Both events and the outer `mcp.*`
event are preserved while capacity remains; full queues drop new events.
Direct REST advisor access and feedback domain auditing stay synchronous.
Queue acceptance is not durable persistence: durability
starts when the worker reaches the existing PostgreSQL audit outbox path.
Records use request ID, credential kind/ID, an allowlisted operation name, HTTP
status, allowed/denied reason, latency and bounded byte counts. Request bytes
count decoded argument JSON up to the outer request cap; response bytes count
captured canonical body bytes up to the response cap, including overflow.
Identifiers are copied and capped at 128 bytes. No caller context or payload is
retained in the queue. Unknown names become `mcp.unknown`. Records do
not contain arguments, response bodies, signed URLs, tokens, prompts, OCR,
base64, notes, balances or provider usage. Sink failure does not change the
business result. This is an audit metadata policy, not a promise that tool
results lack financial or one-time secret data requested by the owner.

Feedback domain events retain their existing redacted identifiers and
resolution reason; the metadata/byte-count policy above describes the dedicated
`mcp.*` operation events. No request context is retained by either queue.

## Maintainer verification

The production and test composition share `NewApplicationRouter`.
`mcp_operation_catalog_test.go` independently classifies authenticated and
excluded routes; `mcp_docs_contract_test.go` checks OpenAPI auth, this locator
and real SDK discovery. Generate Swagger from handler annotations:

```bash
cd backend
rtk go generate ./cmd/api
rtk go test ./internal/controller/http -run 'TestMCP.*Docs|TestMCPAuthenticatedRouteParity|TestMCPRouteExclusions|TestCandidateRouteContract' -count=1
```

Local tests do not establish production rollout. Before deployment, run
owner-authenticated JWT/key initialize/list/call probes, revoked-key and foreign
owner probes, write replay/CAS, multipart recovery, resource download and advisor
async/replay/cancel/confirmation checks against staging. Use synthetic data and
keep credentials outside recorded logs.

References: [API index](../architecture/API.md),
[Advisor registry](../architecture/AI_ADVISOR_TOOLS.md),
[ADR-025](../decisions/ADR-025-mcp-api-parity.md),
[approved design](../work/tickets/MCP-API-PARITY-20261020-DETAIL_DESIGN.md),
[implementation plan](../superpowers/plans/2026-10-20-mcp-api-parity.md),
[validation matrix](../work/VALIDATION_MATRIX.md),
[release notes](../releases/CHANGELOG.md).
