# MyPocket Complete Personal Finance Design

**Status:** approved for implementation
**Date:** 2026-09-11  
**Scope:** complete a production-grade personal-finance product from the existing MyPocket brownfield codebase.

## Product Goal

MyPocket is a Vietnamese, VND-first, offline-capable personal-finance PWA for individuals. The milestone completes the core personal-finance lifecycle rather than reproducing Money Lover branding or commercial-only behavior. A capability counts as complete only when its UI, API-key route, business rules, offline behavior where applicable, public docs, automated proof, and required production acceptance agree.

The core value is trustworthy money state: wallet balances, transfers, budgets, debts, investments, and reports must remain mathematically correct across edits, archives, retries, conflicts, background jobs, and reconnects.

## Approved Boundaries

- Include wallets, categories, income, expense, transfer, balance adjustment, recurring drafts, budgets, events, debts/repayments, investments, reports, search, notifications, receipts, import/export, account reset/deletion, offline sync, public APIs, and production operations.
- Add an OpenAI-compatible backend layer for transaction drafting, classification, and financial analysis. Provider endpoint, model, and API key are configured through backend environment variables.
- AI output is untrusted structured input. It may only create reviewable drafts; user confirmation is required before accounting changes.
- Integrate the owner's existing [OCR Platform](https://github.com/dungxbuif/mac-ocr) as a third-party image-processing tool for the MyPocket agent. Receipt workflows may consume its results, but OCR remains outside the finance and receipt domains and is deployed, authenticated, and operated independently.
- Do not implement bank connection, bank sync, or bank webhooks in this milestone.
- Do not copy Money Lover trademarks, branding, proprietary assets, paywalls, or commercial-only flows.
- Voice input remains deferred.

## Delivery Strategy

Use vertical gap-first phases. Each phase delivers user-visible journeys through storage, domain rules, REST/API-key access, UI, offline reconciliation where relevant, docs, and tests. Existing green behavior is preserved; phases close verified gaps rather than rewriting working modules.

### Phase 1 — Baseline and Production-Safe Data

- Reconcile current dirty changes into reviewable commits without discarding user work.
- Apply and verify migration 0012 against a production-shaped copy.
- Define and exercise full client resync for historical change-feed omissions.
- Prove backup, restore, rollback, migration locking, health checks, and recovery from interrupted rollout.
- Retain PostgreSQL as authority; Redis remains a non-authoritative authentication/audit acceleration layer.

### Phase 2 — Complete Finance and Planning Journeys

- Finish missing mounted flows for transaction adjustment, category hierarchy, wallet/category settings, recurring drafts, budgets, events, obligations, repayments, and portfolio operations.
- Enforce checked integer money, atomic transfer effects, exact edit/archive reversal, optimistic versions, ownership, idempotency, and explicit conflicts at domain boundaries.
- Add E2E journeys for create/edit/archive and cross-feature effects, including report exclusions and transfer neutrality.

### Phase 3 — Analytics, Data Lifecycle, and Receipts

- Complete dashboard/report coverage, filters, drilldowns, cumulative/comparison views, privacy masking, loading/error/empty states, and mathematical cross-checks against known ledgers.
- Complete import/export contracts, account reset/deletion, owned-object cleanup, audit traceability, and safe background execution.
- Complete receipt capture, durable offline bytes, upload/download, retry, authorization, quota behavior, and physical Safari/PWA behavior.
- Allow a receipt workflow to reuse an agent image-processing result when the user chooses, without making the receipt domain responsible for OCR provider calls or provider lifecycle.

### Phase 4 — Public API and Agent Usability

- Expose every supported user-owned capability needed by third parties through versioned REST endpoints authenticated by API key.
- Apply the same ownership, authorization, validation, optimistic version, idempotency, rate-limit, audit, correlation-ID, and error-code rules as browser flows.
- Publish Docusaurus endpoint contracts, curl examples, agent image-input/tool behavior, lifecycle guides, sync semantics, error handling, and a repository-versioned agent skill.
- Machine-readable OpenAPI must match the implemented router and examples; contract tests prevent drift.

### Phase 5 — OpenAI-Compatible Assistance

- Add a provider-neutral client with backend env configuration for base URL, model, API key, timeout, and bounded retries.
- Support natural-language transaction entry, category suggestions, and analysis grounded only in the authenticated user's authorized data.
- Support authenticated image attachments by giving the agent an OCR tool backed by the separately deployed OCR Platform.
- Require structured schemas and deterministic validation. Reject hallucinated wallet/category identifiers, invalid money, unsupported transaction shapes, stale versions, and unsafe analysis requests.
- Persist drafts and provenance, never raw secrets. Redact prompts/provider responses from logs and audits unless a deliberately safe metadata subset is defined.
- Treat OCR output as untrusted agent context. For an optional receipt flow, the agent may propose merchant, date, amount, currency, category, note, and line-item metadata, but deterministic validation and user confirmation remain mandatory.

### Phase 6 — Shared UI, Device Acceptance, and Release

- Use existing or newly created base components for every repeated interaction; keep `App.tsx` as composition rather than growing feature logic.
- Complete responsive mobile/PWA interaction, keyboard/safe-area behavior, install guidance, accessibility, and consistent feedback without forced clicks or hidden errors.
- Remove legacy UI only after replacement routes and regression tests prove equivalent or improved behavior.
- Run full backend race/integration, frontend unit/type/build, complete multi-browser E2E, docs build, security/authorization audit, and physical-device UAT.
- Deploy only after migration/resync/backup/restore gates pass. Verify authenticated production web, API-key operations, worker jobs, object storage, public docs, and rollback readiness.

## Architecture

Keep the existing modular monolith:

- React PWA screens and base components call focused frontend application modules.
- User-scoped IndexedDB is an offline mirror/outbox, never server truth.
- Go HTTP adapters authenticate, authorize, validate transport shape, and delegate to domain/application services.
- Finance, planning, portfolio, analytics, identity, notification, audit, and future AI packages own their respective invariants behind narrow interfaces.
- Shared command transactions commit domain state, change feed, and mutation receipts atomically.
- PostgreSQL is authoritative. Redis, S3-compatible storage, OpenAI-compatible providers, OCR Platform, push, and pricing are adapters with explicit failure behavior.

New AI code belongs behind small model-provider and agent-tool interfaces plus an application service that reads authorized user context and emits typed drafts/analysis. It must not embed provider calls in HTTP handlers or mutate finance repositories directly.

## Third-Party OCR Tool Contract

- The planning baseline is OCR Platform commit `c7a43540be54614a8fb51cb43db75f57f526e2ff`. The implementation must also validate the deployed `/api/v1/openapi.json` and `/v1/ocr/capabilities` at integration/release time so provider drift fails visibly instead of silently changing behavior.
- MyPocket treats OCR Platform as a third-party agent tool, not as a MyPocket subservice or a receipt-domain component. The backend invokes it only for an authenticated agent image request; MyPocket does not re-expose a generic OCR proxy API.
- `OCR_BASE_URL`, `OCR_API_KEY`, request timeout, polling interval, and maximum processing duration are injected through backend environment variables. The upstream key is never sent to the browser, persisted in MyPocket, returned by a MyPocket API, or written to logs/audits.
- The agent tool calls OCR Platform's versioned `POST /v1/documents` and `GET /v1/documents/{documentId}` endpoints with `Authorization: Bearer ...`. It discovers limits through `/v1/ocr/capabilities` and treats `queued`, `processing`, `completed`, `failed`, `cancelled`, expired-result, quota, and dependency-unavailable states distinctly.
- Agent image inputs use MyPocket's private, user-owned object storage and the existing JPEG/PNG/WebP allowlist with a 15 MiB limit, below OCR Platform's 25 MiB decoded Base64 limit. Extend the object-store adapter with a size-bounded private read; the backend verifies stored size/checksum and submits Base64 without creating a public image URL or exposing storage credentials.
- Persist a user-owned agent-tool invocation with its image reference, optional conversation/receipt association, upstream document ID, state, bounded attempts, timestamps, result expiry, and redacted failure metadata. This is agent provenance, not a `receipt_ocr_job`; the receipt domain may reference a completed result but does not own its lifecycle.
- Versioned MyPocket agent endpoints accept an authorized image attachment and read the resulting agent response/tool state. They use existing cookie/API-key identity convergence, CSRF rules for cookie writes, correlation IDs, rate limiting, audit metadata, and stable errors. They are not a public substitute for OCR Platform's own API.
- MyPocket persists only OCR result data needed for authorized agent context/provenance before the upstream Redis-backed result TTL expires. Raw image bytes follow MyPocket's object-storage retention and authorization rules; OCR Platform is not MyPocket's system of record.
- OCR output is untrusted tool data. The model may analyze it or map it into a typed receipt/transaction draft, but any draft must reference valid user-owned wallets/categories, use checked integer money, expose provenance, and require explicit confirmation before a ledger mutation.
- Submission and polling run through the existing worker process, outside request-scoped database transactions. Stable tool-run identifiers, a database lease/claim, bounded retry with backoff, and terminal failure states prevent duplicate work or accounting effects across polls, workers, and repeated client requests.
- Production configuration fails fast when OCR is enabled but its URL/key or safe timeout limits are missing. A runtime OCR outage degrades only the OCR capability, leaves core finance available, produces retryable jobs, and is visible through capability/operational health metadata and redacted audit events. Release acceptance requires OCR enabled and its configured-provider smoke passing.
- Public MyPocket docs describe agent image input, optional receipt-result reuse, asynchronous states, supported formats, limits, expiry, retry behavior, and stable errors. They link to OCR Platform for its own generic OCR contract without copying credentials or presenting MyPocket as the OCR provider.

## API and Authorization Contract

- Browser-cookie and Bearer API-key identities converge on one authenticated user context.
- An Authorization header is authoritative; invalid Bearer credentials never fall back to cookies.
- Every user-owned read and write filters by authenticated user ID. Request payloads cannot choose ownership.
- PostgreSQL validates API-key state; Redis may cache hints but cannot preserve revoked access.
- State-changing operations use stable mutation/idempotency IDs and optimistic base versions when a record can be concurrently modified.
- Audit events record actor, action, target, result, correlation ID, and redacted metadata. Secrets, receipt bytes, raw prompts, and provider payloads are excluded.
- Documentation and contract tests cover both success and stable failure responses.

## Error and Offline Behavior

- Validation, authorization, conflict, retryable infrastructure failure, and permanent provider failure remain distinct.
- Confirmed writes are not rolled back in the UI because a subsequent refresh fails.
- Offline mutations preserve entity identity and original mutation IDs across retries.
- Conflicts retain both local intent and authoritative server state until an explicit user action.
- Optional provider failure never corrupts accounting. AI/OCR/pricing/push failures produce retryable or unavailable states with clear feedback.

## Verification Contract

Each requirement needs evidence at the same scope:

- Pure rules: table-driven unit and property/boundary tests.
- Persistence and authorization: PostgreSQL integration tests with two real users and rollback assertions.
- API/API-key parity: HTTP contract tests for cookie, valid key, revoked key, foreign ownership, malformed input, stale version, and idempotent replay.
- User journeys: real-browser E2E on mobile Chromium, desktop Chromium, and mobile WebKit without force-clicking or dismissing overlays to mask defects.
- Offline: reconnect, lost committed response, duplicate replay, conflict resolution, full resync, account switching, and durable receipt tests.
- AI: deterministic fake-provider contract tests, schema rejection, authorization isolation, prompt-injection resistance, timeout/retry limits, redaction, and confirmation-before-accounting.
- OCR tool: fake-provider submission/polling contract tests, agent-image ownership isolation, quota/expiry/failure mapping, duplicate retry safety, safe context injection, optional OCR-to-receipt-draft schema rejection, and confirmation-before-accounting; configured OCR Platform smoke in the release environment.
- Release: production-shaped migration, backup/restore, physical iOS PWA/receipt UAT, configured-provider smoke, docs build, authenticated production smoke, and rollback drill.

Automated green tests are necessary but do not substitute for physical-device or production acceptance explicitly listed above.

## Completion Definition

The milestone is complete only when:

1. Every in-scope requirement is mapped to exactly one phase and has current evidence.
2. No supported UI action is inert, unreachable, or backed only by prototype data.
3. Public APIs cover the supported third-party workflows and API-key authorization audit is green.
4. AI creates only validated, reviewable drafts or read-only analyses through the env-configured OpenAI-compatible layer.
5. The agent can process authorized images through the third-party OCR Platform with ownership and retry safety; receipt flows may reuse results but only validated, user-confirmed drafts can affect accounting. Bank and voice remain out of scope.
6. Full automated suites, docs, migration/resync, backup/restore, physical-device UAT, production smoke, and rollback checks pass.
7. Legacy UI targeted for replacement is removed only after verified replacement coverage.

## Requirement Sources

- Existing implementation map: `.planning/codebase/`
- Current requirements: `docs/requirements/REQUIREMENTS.md`
- Money Lover comparison: `docs/research/moneylover/PARITY-AUDIT-2026-09-08.md`
- Active gaps and proof: `docs/work/BACKLOG.md`, `docs/work/VALIDATION_MATRIX.md`
- Atomic sync decision: `docs/decisions/ADR-008-atomic-offline-sync-commit.md`
- OCR provider contract: [dungxbuif/mac-ocr](https://github.com/dungxbuif/mac-ocr) at commit `c7a43540be54614a8fb51cb43db75f57f526e2ff`, especially `README.md`, `docs/api/API_REFERENCE.md`, and `docs/api/OCR_RESPONSE.md`

This specification supersedes earlier milestone boundaries that deferred AI, OCR, export, account lifecycle, and production operations. It includes the owner's OCR Platform strictly as a third-party image-processing tool for the agent, permits optional receipt-result reuse, and explicitly removes bank ingestion from the active scope.
