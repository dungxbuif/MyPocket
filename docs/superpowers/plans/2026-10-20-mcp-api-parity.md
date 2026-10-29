# MCP/API Parity Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose every authenticated JWT user operation through a named MCP
adapter and same-owner API-key path without changing business permissions,
owner isolation, validation, CAS, idempotency, or existing Assistant
confirmation semantics.

**Architecture:** A descriptor-backed operation catalog maps stable MCP tool
names to allowlisted existing REST routes. The official
`github.com/modelcontextprotocol/go-sdk/mcp` server owns protocol negotiation,
JSON-RPC envelopes and Streamable HTTP transport. A thin MyPocket adapter
authenticates the outer request, then dispatches an in-process HTTP request
through the existing Gin router so canonical middleware and handlers remain
the source of truth. The adapter normalizes multipart, async handles and REST
errors but does not reimplement finance mutations.

**Tech Stack:** Go 1.26, Gin, the official Model Context Protocol Go SDK
(`github.com/modelcontextprotocol/go-sdk/mcp`), existing
`repository.AuditSink`, PostgreSQL/Redis test fixtures, checked-in
Swagger/OpenAPI.

**Spec:** `docs/work/tickets/MCP-API-PARITY-20261020-DETAIL_DESIGN.md`

## Global Constraints

- No new API-key scopes, MCP-only scopes, MCP-only approval rules, quotas or
  owner selectors.
- `OwnerID` is derived only from the validated JWT/session or persisted `mpk_`
  API-key principal; request arguments cannot impersonate an owner.
- Existing REST handlers remain authoritative for validation, ownership,
  idempotency, CAS, mutation ledgers, password confirmation and status codes.
- Assistant prepared actions still require persisted action provenance and
  confirmation; ordinary REST writes invoked through MCP do not gain a new
  approval requirement.
- MCP adapter never logs/secrets financial payloads, prompts, OCR, base64,
  signed URLs, notes, balances, raw provider responses or model usage.
- No production deployment, dependency install, migration, or source mutation
  occurs until the design/plan approval gate is completed.
- Authenticated REST route additions must fail parity tests if not cataloged;
  public, credential-exchange and service-token routes are explicitly outside
  user parity.

## Review Focus

1. A revoked/expired API key must fail both outer and in-process inner auth;
   test `RequireAuth`, MCP dispatch and long-running advisor revalidation.
2. A caller-supplied `owner_id`/`user_id` or foreign resource ID must not cross
   owner boundaries; test body/query/path and nested filter variants.
3. A lost response must be safely replayable only where the canonical route
   supports the same Idempotency-Key; changed payload and stale CAS must remain
   409 with one mutation.
4. Multipart AI entry must preserve file order and exact limits without
   exposing OCR/model credentials; test 20 files, 21 files, 5 MiB boundaries,
   malformed MIME and ambiguous request recovery.
5. Async Assistant/OCR handles must be pollable/cancellable without an
   adapter retry or false claim of token streaming; test 202, replay, cancel,
   expired history and bounded event pagination.

---

### Task 1: Freeze the authenticated REST parity inventory

**Files:**
- Create: `backend/internal/controller/http/mcp_operation_catalog_test.go`
- Modify: `backend/internal/controller/http/route_contract_parity_test.go`
- Test data source: `backend/docs/swagger.json`

**Interfaces:**
- Consumes: `Router.Register*Routes`, checked-in OpenAPI paths, `RequireAuth`
  route registrations.
- Produces: a deterministic expected set of authenticated user method/path
  pairs and an explicit exclusion set for public and service-token routes.

- [ ] **Step 1: Write failing inventory tests**

  Assert every authenticated user route in the descriptor manifest has a
  matching Swagger path/method/security entry and every currently registered
  `RequireAuth`/`RequireAdvisorAuth` user route appears in the manifest.
  Assert `/login`, refresh, device-login, Google exchange, `/agent/*`, and
  `/internal/*` are excluded for the stated reasons.

- [ ] **Step 2: Run the focused test to verify the catalog is absent/incomplete**

  Run: `rtk go test ./internal/controller/http -run 'TestMCP.*Parity|TestMCP.*Exclusion' -count=1`
  Expected: FAIL because the MCP manifest/registration does not exist.

- [ ] **Step 3: Add the explicit route manifest**

  Define the full method/path list from the current Swagger and route files in
  a test-owned manifest. Do not use a broad path prefix as proof; each method
  must be named so a future route drift is visible.

- [ ] **Step 4: Run the inventory tests**

  Run: `rtk go test ./internal/controller/http -run 'TestMCP.*Parity|TestMCP.*Exclusion' -count=1`
  Expected: PASS for the current route set.

- [ ] **Step 5: Commit**

  `git add backend/internal/controller/http/mcp_operation_catalog_test.go backend/internal/controller/http/route_contract_parity_test.go && git commit -m "test: freeze authenticated MCP parity inventory"`

### Task 2: Add descriptor types and named operation catalog

**Files:**
- Create: `backend/internal/controller/http/mcp_operation_catalog.go`
- Test: `backend/internal/controller/http/mcp_operation_catalog_test.go`

**Interfaces:**
- Consumes: Task 1 method/path inventory.
- Produces: `ExternalOperation`, `ExternalOperationCatalog`, stable
  `List()`/`Lookup(name)` methods, argument schemas and body encoders.

- [ ] **Step 1: Write RED catalog tests**

  Test stable names for all domains, duplicate-name rejection, path-template
  coverage, explicit body encoding, idempotency mode, async marker and
  `owner_id`/`user_id` rejection. Test the catalog exposes ordinary writes,
  not only read tools.

- [ ] **Step 2: Implement descriptor metadata**

  Add fields from the design (`Name`, method/path, schema, encoding,
  idempotency, async, sensitivity) plus a typed `Encode(arguments, meta)`
  function. Keep descriptors declarative; no repository/usecase calls here.

- [ ] **Step 3: Register every authenticated route**

  Add descriptors for account, categories, wallets, transactions, budgets,
  budget series, jars/months, recurring, credit, travel, portfolio, reports,
  notifications, feedback, AI entry, advisor and sync. Explicitly omit
  credential-exchange and service-token routes.

- [ ] **Step 4: Verify schemas and manifest parity**

  Run: `rtk go test ./internal/controller/http -run 'TestMCP.*Catalog|TestMCP.*Parity' -count=1`
  Expected: PASS, with all catalog names unique and all current paths covered.

- [ ] **Step 5: Commit**

  `git add backend/internal/controller/http/mcp_operation_catalog.go backend/internal/controller/http/mcp_operation_catalog_test.go && git commit -m "feat: define named authenticated operation catalog"`

### Task 3: Integrate the official MCP SDK transport and in-process dispatcher

**Files:**
- Create: `backend/internal/controller/http/mcp_handler.go`
- Create: `backend/internal/controller/http/mcp_handler_test.go`
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`
- Modify: `backend/internal/controller/http/router.go`
- Modify: `backend/cmd/api/main.go`

**Interfaces:**
- Consumes: `ExternalOperationCatalog`, `Router.Engine`, `AuthMiddleware`.
- Produces: authenticated `/api/v1/mcp` Streamable HTTP endpoint backed by
  `mcp.StreamableHTTPHandler`, with SDK-owned `initialize`, `tools/list`,
  `tools/call`, session negotiation and notifications.

- [ ] **Step 1: Pin the official SDK and write RED integration tests**

  Pin one reviewed stable SDK version in `go.mod`. Exercise the endpoint with
  an SDK client rather than hand-crafted JSON-RPC for initialize, tools/list,
  tools/call, session negotiation and cancellation. Keep negative HTTP tests
  for authentication and body limits. Assert no arbitrary path or method can
  be supplied.

- [ ] **Step 2: Configure the SDK server and Streamable HTTP handler**

  Create `mcp.Server` with explicit MyPocket implementation name/version,
  register only catalog-backed tools and mount `mcp.StreamableHTTPHandler`.
  Wrap it with `http.MaxBytesReader`, request deadline, origin validation and
  existing authentication/audit middleware. Let the SDK own protocol parsing
  and error envelopes. Do not fork, copy or partially reimplement SDK JSON-RPC.

- [ ] **Step 3: Implement authenticated outer middleware and dispatcher**

  Register the route behind `RequireAuth` and a dedicated MCP audit wrapper.
  For a named descriptor, create an in-process request with the original
  bearer credential, propagated request ID, query/path/body/headers, and
  context. Call `Router.Engine.ServeHTTP`; never call a handler by arbitrary
  string or invoke a repository directly.

- [ ] **Step 4: Map HTTP response to SDK structured tool results**

  Return `mcp.CallToolResult` content plus bounded `structuredContent`.
  Preserve HTTP status, response JSON/body, content type and request ID in
  bounded metadata. Set `IsError` for non-2xx while retaining the original
  MyPocket `Problem` code/status/detail.

- [ ] **Step 5: Run focused transport tests**

  Run: `rtk go test ./internal/controller/http -run 'TestMCP(Initialize|Tools|Call|Protocol|Dispatch)' -count=1`
  Expected: PASS.

- [ ] **Step 6: Commit**

  `git add backend/go.mod backend/go.sum backend/internal/controller/http/mcp_handler.go backend/internal/controller/http/mcp_handler_test.go backend/internal/controller/http/router.go backend/cmd/api/main.go && git commit -m "feat: add authenticated MCP SDK transport"`

### Task 4: Prove JWT/API-key parity, revocation and owner isolation

**Files:**
- Create: `backend/internal/controller/http/mcp_auth_integration_test.go`
- Modify: `backend/internal/controller/http/mcp_handler.go`

**Interfaces:**
- Consumes: Task 3 transport and current auth middleware/API-key service.
- Produces: fail-closed auth behavior and negative proof for every dispatch.

- [ ] **Step 1: Write RED auth/owner tests**

  With two owners and disposable PostgreSQL/Redis, call the same operation
  using JWT and owner API key and compare authorized behavior. Test foreign
  wallet/transaction/category/budget/feedback IDs return existing 404 semantics.
  Test malformed key never falls back to JWT, revoked/expired keys return 401,
  and session expiry/logout is rechecked.

- [ ] **Step 2: Implement any missing principal propagation**

  Pass only the original bearer to the inner REST request; reject body/query
  owner selectors before dispatch. Do not add scopes or a second principal
  resolver. Ensure API-key logout revokes the current key exactly as REST.

- [ ] **Step 3: Add advisor lifecycle revalidation coverage**

  Start an async advisor MCP call, revoke the API key, and prove provider/tool
  work and answer persistence fail closed according to the existing validator.

- [ ] **Step 4: Run auth integration tests**

  Run: `rtk go test -race ./internal/controller/http ./internal/usecase ./internal/infrastructure/repository -run 'TestMCP.*(Auth|Owner|Revocation|Advisor)' -count=1`
  Expected: PASS with a disposable local database; production is forbidden.

- [ ] **Step 5: Commit**

  `git add backend/internal/controller/http/mcp_auth_integration_test.go backend/internal/controller/http/mcp_handler.go && git commit -m "test: prove MCP JWT and API-key owner parity"`

### Task 5: Verify ordinary write semantics, idempotency and CAS

**Files:**
- Create: `backend/internal/controller/http/mcp_write_integration_test.go`
- Test support: existing transaction mutation/category/wallet/budget fixtures.

**Interfaces:**
- Consumes: Task 3 descriptors and canonical REST handlers.
- Produces: parity proof for direct user writes without MCP-specific approval.

- [ ] **Step 1: Write RED write tests**

  Cover wallet/category/budget/jar/transaction/transfer/adjustment/bulk
  delete/credit/recurring/travel/portfolio/feedback writes. For each, assert
  state readback through MCP and REST. Include required idempotency-key replay,
  changed payload 409, transfer pair atomicity, and stale expected-version CAS.

- [ ] **Step 2: Implement descriptor encoders and header mapping**

  Map `_meta.idempotency_key` to `Idempotency-Key`, preserve other supported
  headers only where explicitly declared, and reject duplicate/body owner
  metadata. Do not auto-generate or retry keys.

- [ ] **Step 3: Run real database write proof**

  Run: `TEST_DATABASE_URL=<disposable-local-postgres> rtk go test -race ./internal/controller/http ./internal/infrastructure/repository -run 'TestMCP.*Write|TestMCP.*Idempotency|TestMCP.*CAS' -count=1`
  Expected: PASS with only synthetic rows; no production DB.

- [ ] **Step 4: Commit**

  `git add backend/internal/controller/http/mcp_write_integration_test.go && git commit -m "test: verify MCP direct user write parity"`

### Task 6: Implement AI, async, multipart and binary operation adapters

**Files:**
- Modify: `backend/internal/controller/http/mcp_operation_catalog.go`
- Modify: `backend/internal/controller/http/mcp_handler.go`
- Create: `backend/internal/controller/http/mcp_ai_integration_test.go`

**Interfaces:**
- Consumes: existing AI entry/advisor routes and descriptor encoders.
- Produces: safe MCP mappings for async handles, files, replay events and
  attachment resources.

- [ ] **Step 1: Write RED AI/file tests**

  Cover advisor sync/async/capabilities/history/reset/status/cancel/events,
  action confirmation provenance, AI-entry 20-file multipart, request recovery,
  proposal edit/approve/reject, attachment download resource link and no
  automatic resubmission.

- [ ] **Step 2: Implement multipart encoding**

  Accept `{filename,mime_type,base64}` entries, preserve order, enforce 20 files,
  5 MiB/file, 101 MiB total and detected MIME before building the existing
  multipart request. Never forward OCR/provider credentials.

- [ ] **Step 3: Implement async/event normalization**

  Return run/request handles for 202 responses. Map finite advisor SSE replay
  to bounded event data with `after_seq`; do not claim token streaming. Keep
  action confirmation inputs limited to action ID/conversation/generation.

- [ ] **Step 4: Implement attachment resource result**

  Follow the canonical short-lived 302 location into an MCP resource link,
  preserve expiry/owner enforcement, and redact the location from logs.

- [ ] **Step 5: Run focused AI tests**

  Run: `rtk go test -race ./internal/controller/http ./internal/infrastructure/ai ./internal/usecase -run 'TestMCP.*(AI|Advisor|Entry|Multipart|Attachment|Async)' -count=1`
  Expected: PASS with local HTTP/provider fixtures only.

- [ ] **Step 6: Commit**

  `git add backend/internal/controller/http/mcp_operation_catalog.go backend/internal/controller/http/mcp_handler.go backend/internal/controller/http/mcp_ai_integration_test.go && git commit -m "feat: map MCP AI async and file operations"`

### Task 7: Add MCP audit, privacy and operational limits

**Files:**
- Modify: `backend/internal/controller/http/mcp_handler.go`
- Create: `backend/internal/controller/http/mcp_audit_test.go`

**Interfaces:**
- Consumes: existing `repository.AuditSink`, Redis fast path and PostgreSQL
  audit outbox wiring.
- Produces: redacted `mcp.*` audit events and bounded operational behavior.

- [ ] **Step 1: Write RED privacy/audit tests**

  Assert events contain request/credential/operation/status/decision/latency
  only; assert they do not contain bearer secrets, arguments, note text,
  amounts, OCR/base64, signed URLs, prompts or provider usage. Test audit
  failure remains best-effort without changing the business response.

- [ ] **Step 2: Implement audit wrapper and limits**

  Record `mcp.<operation>` using the existing sink/outbox path. Bound request,
  response, argument depth, file bytes, per-call timeout and event count. Do
  not reuse advisor-prefix detection as a substitute for MCP audit.

- [ ] **Step 3: Run audit/race tests**

  Run: `rtk go test -race ./internal/controller/http ./internal/infrastructure/cache ./internal/infrastructure/repository -run 'TestMCP.*(Audit|Privacy|Limit)' -count=1`
  Expected: PASS.

- [ ] **Step 4: Commit**

  `git add backend/internal/controller/http/mcp_handler.go backend/internal/controller/http/mcp_audit_test.go && git commit -m "feat: audit and bound MCP operations"`

### Task 8: Route/OpenAPI and machine-readable documentation reconciliation

**Files:**
- Modify: `backend/docs/swagger.yaml`
- Modify: `backend/docs/swagger.json`
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/AUTH_API.md`
- Modify: `docs/guides/api-keys.md`
- Modify: `docs/architecture/AI_ADVISOR_TOOLS.md`
- Create/update: `docs/ai/mcp-api-parity.md`, `llms.txt` or equivalent public AI surface
- Create: `docs/decisions/ADR-025-mcp-api-parity.md`

**Interfaces:**
- Consumes: implemented descriptor catalog and route tests.
- Produces: human and AI-agent-readable MCP endpoint/tool schemas, parity
  exclusions, auth/key lifecycle, limits, errors and examples.

- [ ] **Step 1: Write docs contract tests**

  Assert `/api/v1/mcp` appears in generated OpenAPI, security is declared,
  descriptor names and schemas appear in the machine-readable catalog, and
  excluded login/service-token paths are documented.

- [ ] **Step 2: Reconcile docs after implementation**

  State that same-owner API keys and JWTs have identical user authority, no
  extra scopes exist, MCP does not bootstrap credentials, and Assistant
  confirmation remains operation-specific. Document multipart, async, replay,
  idempotency, CAS and signed-resource behavior.

- [ ] **Step 3: Generate/check OpenAPI**

  Run: `cd backend && rtk go generate ./cmd/api && rtk go test ./internal/controller/http -run 'TestMCP.*Docs|TestCandidateRouteContract' -count=1`
  Expected: generated documents contain the authenticated MCP route and no
  stale route/tool catalog.

- [ ] **Step 4: Commit**

  `git add backend/docs docs/architecture docs/guides docs/ai llms.txt && git commit -m "docs: publish MCP API parity contract"`

### Task 9: Aggregate verification and rollout gate

Final-review follow-up: the [correction wave](../../work/bugs/MCP-FINAL-REVIEW-20261020.md)
adds permanent regression proof for feedback PNG upload/redirect variants,
canonical audit stalls and the category receipt dependency. It verifies an
archive of the exact staged source separately from unrelated shared-worktree
edits; final counts are recorded in that note and the validation matrix.

Local execution checkpoint (2026-10-20): Tasks 1–9 locally verified against the assembled
candidate including the exact injection-proof correction `35d3c0e`.
PostgreSQL 17/Redis 7 are isolated loopback fixtures at clean migration 40.
Backend normal/full race: each 727 passing test/subtest results, zero failures,
11 explicit live provider/storage skips. Focused MCP/router/audit race: 178
passing results, zero failures/skips. Route/docs gate: 9 passing tests;
OpenAPI regeneration has no drift (91 paths, 101 tools, 19 exclusions).
Vet, module verification, gofmt/diff checks and 229 design-doc links pass.
Explicit official SDK initialize/list/read/write/revoked-key smoke passes 27
results with zero skips. Both owned test containers and their anonymous
volumes were removed after zero remaining owner rows/namespaced MCP keys.
Detailed evidence and remaining staging gates:
[`VALIDATION_MATRIX.md`](../../work/VALIDATION_MATRIX.md) and
[`ROLLOUT_READINESS.md`](../../operations/ROLLOUT_READINESS.md).
This is local verification; staging/provider/device/release acceptance remains
open and no production deployment or database operation occurred.

**Files:**
- Modify only the implementation/design verification sections as needed.

**Interfaces:**
- Consumes: all prior task outputs and disposable test fixtures.
- Produces: release-ready local evidence and explicit deployment residuals.

- [x] **Step 1: Run focused and race verification**

  Run:
  `rtk go test -race ./internal/controller/http ./internal/usecase ./internal/infrastructure/ai ./internal/infrastructure/repository`
  and `rtk go vet ./...`.
  Expected: PASS; any unrelated pre-existing failure is recorded by file and
  not hidden.

- [x] **Step 2: Run real disposable PostgreSQL/Redis parity suite**

  Migrate only a disposable local database, run the full owner isolation,
  idempotency, audit and async tests, then remove the fixture. Expected: no
  production URL, credentials or rows touched.

- [x] **Step 3: Run route and docs gates**

  Run `rtk git diff --check`, route parity tests, generated OpenAPI checks and
  all existing project validation commands relevant to backend contracts.

- [x] **Step 4: Record rollout prerequisites**

  Before deployment, require a staging owner-authenticated smoke using both a
  JWT and an API key, an MCP client initialize/list/call probe, revoked-key
  negative probe, write replay/CAS probe, AI-entry multipart probe and advisor
  async/cancel probe. No production data reset or migration is needed for the
  adapter unless implementation proves otherwise.

- [x] **Step 5: Commit verification notes**

  Commit only implementation-owned verification notes after actual commands;
  do not claim device/provider/live-production acceptance from local tests.

## Plan self-review

- REST coverage: all authenticated user route families are explicitly listed;
  Task 1 makes omission fail closed.
- API-key/JWT parity: Task 4 checks same-owner equivalence and revocation.
- Direct writes: Task 5 covers ordinary mutations, idempotency, CAS and pairs.
- AI/file/async: Task 6 covers multipart, polling, replay, cancellation and
  confirmation provenance.
- Privacy/audit: Task 7 covers secret/financial redaction and outbox behavior.
- Human/machine docs: Task 8 covers OpenAPI and agent-readable catalog.
- Runtime proof: Task 9 separates disposable/staging/device/provider gates.
- Known gaps are explicit: credential bootstrap routes, service-token routes,
  and live MCP streaming are not silently represented as available user tools.
