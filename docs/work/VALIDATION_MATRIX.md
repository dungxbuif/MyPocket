---
artifact_type: validation_matrix
id: VALIDATION_MATRIX
status: active
owner: shared
human_fields:
  - proof_override
  - acceptance_signoff
ai_fields:
  - proof_recommendation
  - evidence_links
  - status_updates
shared_fields:
  - matrix_rows
  - validation_status
updated: 2026-09-22
---

# Validation Matrix

## Field Ownership

- Human owns proof overrides and acceptance sign-off.
- AI recommends proof types, links evidence, and updates status from verification.

This file maps accepted behavior and work items to proof.

## Full implemented-scope regression audit — 2026-09-22

- Backend: `TEST_DATABASE_URL=<local dev PostgreSQL> go test -race ./... -count=1` PASS — 110 tests in 15 packages; migration CLI reports `version: 16 dirty: false`.
- Frontend: `npm run check:design`, `npm run test:design` (7), `node --test scripts/form-logic.test.ts` (2), `npm run test:transactions` (5), `npm run test:calendar` (9), `npm run test:transaction-jars`, `npm run test:ai`, `npm run test:api:full`, `npm run test:transfer:integrated`, `npm run test:transfer:e2e`, and `npm run build` PASS. Production build transformed 1,764 Vite modules. AI/transfer browser suites were run sequentially; a deliberately parallel attempt produced a fixture timeout from shared Vite/Chrome ports and passed on the clean serial rerun.
- API harness: `TEST_API_ORIGIN=http://127.0.0.1:4173 TEST_DATABASE_URL=<local dev PostgreSQL> npm run test:api:full` PASS through the real FE proxy. It verifies authenticated profile/home, timezone validation, category CRUD and wallet scope, jar create/list/update/report/remove, month summary and note save/delete, AI capabilities, Swagger, and unauthenticated guards. It removes only its generated category, jar/configuration, and note; no ledger rows are created. The seeded account timezone value is unchanged (the valid same-timezone call may confirm an unconfirmed account, as the endpoint contract requires).
- Transfer harnesses: integrated FE-proxy/PostgreSQL roundtrip and real Chromium 390×844 flow both PASS; paired rows, balances, report/jar exclusion, auth guard, rendered wallet options, submit, refresh and persisted state were asserted. Only generated wallets/transactions are removed.
- Fixture correction: `app/scripts/api-roundtrip.mjs` now uses the documented budget contract `start_date`/`end_date` (the stale `start_at`/`end_at` fixture was the only initial red).
- Residual human/platform proof remains explicit: owner visual sign-off, mobile/PWA install, Google OAuth with real credentials, live S3/OCR/LLM, and deployment verification are outside this local regression run. AI browser/provider tests intentionally do not claim live OCR/LLM quality or proposal approval.

## TICKET-02-02 atomic wallet transfer — 2026-09-22

- Scope and trace: [detail design](tickets/TICKET-02-02-TRANSFER-DETAIL_DESIGN.md), [transaction screen](../design/screens/transactions/README.md), [API contract](../architecture/API.md), [ERD](../architecture/ERD.md).
- Schema proof: migration `000016_transaction_transfer_links` applied locally; `go run ./cmd/migrate version` reports `version: 16 dirty: false`.
- Backend proof: `TEST_DATABASE_URL=<local dev database> go test ./... -count=1` PASS. Handler tests cover valid paired creation and same-wallet rejection. PostgreSQL integration proves exactly two linked rows, explicit `included_in_reports=false`, owner/credit guards and atomic persistence.
- Frontend proof: `npm run check:design` PASS; `npm run test:design` PASS (7 guardrails plus base contracts); `npm run build` PASS (1,764 Vite modules). TransferFields composes `BaseSelect`, `AmountField`, `DateField`, `BaseTextInput` and `SurfaceCard`; no native controls or literal colors were added outside bases.
- Integrated proof: `npm run test:transfer:integrated` PASS through Vite proxy → Gin → PostgreSQL; verifies persisted pair, balances, report/jar exclusion and unauthenticated 401, then cleans only test IDs. `npm run test:transfer:e2e` PASS in real Chromium at 390×844; opens the actual transaction page, selects both rendered wallet options, submits the real sheet, waits for the refreshed UI note, re-reads API state and saves `/tmp/mypocket-transfer-e2e.png`.
- Residual UAT: owner visual sign-off remains human-owned. Pair edit/delete and idempotency-key replay are intentionally follow-up scope.

## CORE-03 timezone, jars, and live monthly summary — 2026-09-22

- Scope and trace: [CORE-03 design](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), [execution plan](../superpowers/plans/2026-09-22-timezone-jars-month.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md), [account settings](../design/screens/current-ui.md), [jar screen](../design/screens/jars/README.md), [month detail](../design/screens/month-detail/README.md), [overview](../design/screens/overview/README.md), [transaction form](../design/screens/transactions/README.md).
- Persistence/API proof: additive migrations `000013`–`000015` are applied locally; `go run ./cmd/migrate version` reports version 15, `dirty=false`. PostgreSQL repository tests cover month-note owner scope/clear, stable jar month initialization under concurrent calls and owner isolation, empty summary collections, and account-local calendar behavior. Backend full suite `TEST_DATABASE_URL=<local dev database> go test ./... -count=1` PASS; database URL intentionally omitted from this record.
- Frontend proof: `npm run check:design` PASS (105 local design links and base/color/native-control checks); `npm run test:design` PASS (7 design guardrails plus base contracts); `npm run test:transactions` PASS (5); `npm run test:ai` PASS; `npm run test:calendar` PASS (9); `npm run test:transaction-jars` PASS; `npm run build` PASS (1,763 Vite modules). The Vite-backed AI and jar tests were run sequentially to avoid shared dev-server port contention.
- Integration/runtime proof: backend health and frontend root returned HTTP 200 on the already-running local servers. No owner ledger was changed. Migration backfill, live transaction CRUD with jars, month note editing, timezone switching and visual acceptance still require owner UAT; do not mark the feature complete on automated tests alone.
- Reporting semantics: month figures remain recalculable from live ledger; completion is derived by account-local month, with no close snapshot or cron job. The separately requested worker is draft [WORKER-01](tickets/WORKER-01-cron-service.md); its requested month-end report/close behavior must be reconciled with this contract before detail design.

## UI-FORMS-03 reference follow-up — 2026-09-21

- Scope and trace: [UI-FORMS-03](tickets/UI-FORMS-03-DETAIL_DESIGN.md), [transaction screen](../design/screens/transactions/README.md), [budget screen](../design/screens/budgets/README.md), [assistant screen](../design/screens/assistant/README.md), [BaseCategoryTree](../design/molecules/category-tree/README.md).
- Owner decisions: Money Lover guides the AI composer/result presentation; MyPocket keeps per-item edit/remove/save and save-all confirmation. Group pickers use the same BaseCategoryTree as Account → Manage Groups; no screen-local category list.
- Frontend fixture proof: `rtk proxy npm run test:ai` in `app/` PASS after checking AI textarea/attachment/result card, transaction and budget tree pickers, parent/child selection, and transaction category ID payload. `rtk proxy node scripts/base-contract.test.mjs` PASS for tree selection and selected mark. Earlier first red was a test fixture omission (no `onSelect` callback), then isolated rerun passed; no code fix was needed for that assertion.
- `rtk proxy npm run check:design` in `app/`: PASS (theme/native-control guard and 96 local links); `rtk proxy npm run test:design`: PASS (7 guardrail fixtures + base contracts); `rtk proxy npm run test:transactions`: PASS (5 cases); `rtk proxy node --test scripts/form-logic.test.ts`: PASS (2 cases); `rtk proxy npm run build`: PASS (TypeScript and 1,756 Vite modules); `rtk proxy git diff --check`: PASS.
- Visual proof: browser fixture at 390×844 compared Money Lover composer/result screenshots to `/tmp/mypocket-ui-pIE4DC/ai-entry-final.png`; layout/structure and MyPocket theme pass. AI browser fixture passed after the final UI adjustment. Owner visual UAT remains pending. AI-ENTRY-02 still owns replacing current session/message persistence with one-shot batch APIs and private receipt linkage.

## AI-ENTRY-02 stateless/OCR-first follow-up — 2026-09-22

- RED→GREEN: backend OCR adapter initially rejected `application/pdf`; a regression proved zero OCR/model calls. PDF validation and OCR Platform's private presigned-upload flow now pass. Captured LLM request tests prove OCR text is present and PDF bytes/base64/source URL/image input are absent. Invalid file validation now returns a client error before request/provider claim.
- Backend route contract test proves `POST /api/v1/ai/entry/process` and read-only `GET /api/v1/ai/entry/requests/:request_id` exist and no public `/sessions` or `/messages` routes remain. Process response uses `process_id`; repository writes no conversation messages and LLM extraction type no longer carries history.
- Frontend AI service test proves a raw PDF `File` is sent as multipart without a base64 field or manually set Content-Type. Browser fixture passed real Chrome one-shot submit and asserts no session/message calls. `npm run typecheck`, AI service/browser tests, `npm run check:design`, `npm run test:design`, and `npm run build` pass.
- `go test ./... -count=1` in `backend/`: PASS; rerun with the local dev `TEST_DATABASE_URL`: PASS, including real PostgreSQL approval/link/download authorization, rejection of unlinked downloads, expired-unlinked cleanup claims and stable OCR-text-to-file mapping. Migration CLI: `go run ./cmd/migrate up && go run ./cmd/migrate version` → version 12, `dirty=false`. Fake-store/use-case proof confirms original bytes are retained and only the signed URL reaches OCR; OCR adapter proof confirms only recognized text reaches the LLM and storage failure stops before extraction. The explicit cleanup command exists but was not run against bucket contents. Live S3/OCR and browser download UAT remain pending. Do not auto-retry the user's prior PDF.
- Startup config: `go test ./cmd/api -run TestAttachmentStorageIsOptionalOnlyInDevelopment -count=1` PASS; empty storage is allowed only in development and complete config initializes the adapter. S3 boundary regressions reject control characters/traversal in environment, prefix, bucket, region and key IDs.
- Local runtime restarted after implementation: `GET http://localhost:4173/` → 200 and `GET http://localhost:8080/api/v1/health` → 200. No live OCR/S3 request was made.

## AI-ENTRY-02 live provider evaluation — 2026-09-22

- Scope: synthetic receipt image only; no user PDF, no proposal approval, and no ledger write. Harness: `backend/internal/infrastructure/ai/live_eval_test.go`.
- Command: `rtk proxy env AI_EVAL_LIVE=1 AI_EVAL_IMAGE=/tmp/mypocket-ai-eval-receipt.png go test ./internal/infrastructure/ai -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v`.
- Regression history: before the fix, five live runs reproduced S3 PUT success + signed GET `403 SignatureDoesNotMatch`; alternate `us-east-1` signing also failed.
- Fix/evidence: [ADR-007](../decisions/ADR-007-aws-s3-presigning.md) replaces handwritten signed GET with AWS SDK for Go v2 S3 presigning. `rtk proxy go test ./internal/infrastructure/storage -count=1` and `rtk proxy go test ./... -count=1` in `backend/`: PASS. Live private PUT+GET readback PASS (1,027,882 bytes, 62 ms); the test deferred object cleanup.
- OCR: PASS against the configured live provider, 3.815 s; synthetic mock total of 45,000 VND recognized.
- LLM: all five cases failed strict draft-schema validation; field score 0/15 (0%). LLM latency p50 15.089 s, p95 21.7 s. Transfer-safety gate failed; OCR + receipt-case pipeline measured 25.516 s. AI-ENTRY-01 schema compatibility is not verified and is not a successful intelligence score.
- Safe diagnostic log follow-up: `rtk proxy go test ./internal/infrastructure/ai -run '^(TestSchemaMismatchLogReportsStructureWithoutLoggingModelContent|TestSchemaDiagnosticShowsArrayElementShapeWithoutValues|TestSchemaValidation)$' -count=1 -v` PASS; full `rtk proxy go test ./... -count=1` in `backend/` PASS; `rtk proxy git diff --check` PASS. A live run showed root arrays with wallet-like item fields (`currency,id,name,type`) and objects with input-context keys (`categories,now,timezone,untrusted_source_text,wallets`); one of five calls timed out. Five-case metrics: LLM p50 21.138 s, p95 30.001 s; OCR 3.805 s; OCR + receipt pipeline 33.806 s; 0/15 fields. Latest text-only live call logged `finish_reason=stop`, root array/item shape with input-context keys, 363 response bytes, and 14.706 s latency. Logs retain only allowlisted schema/catalog keys, redact arbitrary key names and exclude values/full response bodies. See [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md).
- Cleanup regression: `TestAttachmentDeleteStatusRequiresClaimedDeletingState` red→green; stale or unclaimed attachment cleanup finalization now returns conflict instead of silently reporting success.
- Synthetic fixture only; no user PDF, proposal approval, or ledger write. See [BUG-001](bugs/BUG-001-s3-presigned-get-signature.md).

## AI-ENTRY-01 OpenAI-compatible JSON Schema resolution — 2026-09-22

- Comparison: prior `response_format: json_object` production path scored 0/15 and returned input-context-shaped roots. A forced tool/function-calling probe reached response streaming but did not complete within the 30 s HTTP-client timeout; tool calling is not used because extraction must never execute a tool. Strict `response_format: {type: json_schema, json_schema: {strict: true, ...}}` is supported by the configured OpenAI-compatible endpoint/model.
- Implementation: production adapter now uses strict JSON Schema and retains the strict application parser, human review, and no-ledger-before-approval boundary. No SDK/dependency was needed: standard OpenAI-compatible `net/http` transport successfully exercised the provider contract. Optional owner wallet descriptions are included as bounded (2 KiB each), explicitly untrusted disambiguation context; no DB/API change.
- Live command: `rtk proxy env AI_EVAL_LIVE=1 AI_EVAL_IMAGE=/tmp/mypocket-ai-eval-receipt.png go test ./internal/infrastructure/ai -run '^TestLiveQwenImageAndTransactionEval$' -count=1 -v` in `backend/`.
- Live result: private S3 PUT/readback PASS (1,027,882 bytes, 83 ms); OCR PASS (4.034 s, synthetic total recognized); LLM synthetic evaluation PASS 5/5 cases, 24/24 scored fields (100%), transfer-safety PASS. LLM p50 19.769 s, p95 25.655 s; OCR + receipt-case pipeline 29.689 s. Cases: single expense, two expenses/relative date, income/large VND, internal transfer, OCR receipt.
- Tests: `rtk proxy go test -race ./internal/infrastructure/ai -count=1` PASS; `rtk proxy go test ./... -count=1` PASS in `backend/`; `rtk proxy git diff --check` PASS. Synthetic-only; no proposal approval or ledger write. A five-case benchmark is not broad model certification; owner review and broader edge-case UAT remain required.
- Wallet-description and future provider/pricing intent: [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [AI-ENTRY-03](tickets/AI-ENTRY-03-PROVIDER-SELECTION.md) and its draft design.

## AI plan review — 2026-09-21

### AI-ENTRY-01 implementation follow-up

- [Approved slice and execution proof](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [UI contract/proof](../design/screens/assistant/README.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md). Entry scope is implemented for review; this does not complete all REQ-17, financial Q&A, transfer matching or retained receipts.
- Backend red→green tests: missing validator/repository/service/config functionality; actual regressions caught hidden category-template acceptance and too-large history blocking follow-up. PostgreSQL fixtures prove zero writes before approval, eight concurrent approvals producing one row, stale version conflict, cross-owner rejection, terminal reject, deleted-ledger replay safety, explicit report=false preservation and the current no-user-usage-limit policy for AI entry. HTTP fixture test exercises real provider adapter → service → DB → edit/approve/reject and reload.
- Independent review remediation: shared category wallet scopes no longer disclose/overwrite other owners' assignments; OCR source survives model failure; post-claim history includes the previous completion. All three regressions observed red then green; new fixture category/wallets removed after verification.
- Final post-review run: full `go test -race ./... -count=1` with local `TEST_DATABASE_URL` passed; frontend `test:ai`, `check:design`, `test:design`, `test:transactions`, `build` passed. Local documentation scan 184 links/22 Markdown files: pass. Git-visible files contain none of the supplied credentials; ignored local env mode is 0600. Runtime restarted with final source.
- `rtk go run ./cmd/migrate up` pass, schema 000010; `rtk go generate ./cmd/api` pass. Go suite/race with `TEST_DATABASE_URL` runs real DB tests; fixture records are cleaned, owner data untouched.
- Frontend `test:ai`, `check:design`, `test:design`, `test:transactions`, `build` pass; Chrome test fixtures prove hold/cancel/release suppression and review flow. Live Chrome through localhost:4173 verifies normal Add/manual form, AI entry alternative, configured OCR/unconfigured AI display. No artificial proposals were inserted into the owner's account.
- Live proxy auth guard returns 401, Swagger 200 with eight entry paths. OCR credentials are configured but public capability/auth-format probes do not prove actual OCR extraction. The Qwen extraction test previously failed strict schema validation; live AI/OCR quality, physical touch and owner acceptance remain pending. Private S3 configuration passed put/presign smoke checks, while download roundtrip and application attachment-link implementation remain pending.

- [DESIGN-09-AI](tickets/TICKET-09-DETAIL_DESIGN.md): docs-only feasibility/design/plan delivered for review. Existing two AI tickets remain draft; product approval, provider quality, ledger integration and UAT are pending.
- Read-only evidence: transaction entity/handler/repository confirm current income/expense-only ledger and missing AI confirmation path. OCR public OpenAPI and capabilities returned HTTP 200; current contract does not list scan endpoints. Authenticated OCR/model calls were not run.
- Docs validation: `rtk proxy node --input-type=module -e '…'` checked 117 local Markdown links across the 10 touched documents: pass, zero missing targets. `rtk git diff --check`: pass. Placeholder scan in the new plan found no TODO/TBD. App/Go tests skipped because no runtime code changed. Planned unit, PostgreSQL concurrency, provider contract, model eval and browser UAT gates are in T0–T7; none are claimed passed by this planning work.

## Runtime Verification — 2026-09-13

### Follow-up 2026-09-20

- Budget review follow-up: real-Postgres disappearing-row update regression red (nil error/resurrection) → green (not found/no insert fallback); metadata-date preservation regression passes with browser timezone UTC. `go test ./...` with TEST_DATABASE_URL executes optional integration test. Production API restarted and full `api-roundtrip.mjs` rerun passes after fixes. `npm run build` passes; `git diff --check` clean. Independent review evidence in [API-SCREENS-01](tickets/API-SCREENS-01-DETAIL_DESIGN.md).

- API-SCREENS-01: `rtk proxy go run ./cmd/migrate up` pass (000009); `rtk proxy node scripts/api-roundtrip.mjs` in app pass through localhost:4173 to PostgreSQL: goal date/deposit/withdrawal/interest/edit/delete/current_balance, invalid goal catalog rejection; budget CRUD/exact overlap409/child category/report exclusion/recalculation and unauthenticated401. Only test-created IDs cleaned, no owner data removed. `go test ./...` pass; design checks/SSR/build pass. New tests fail before implementation for goal filter, category history, budget derivation/validation, masked amounts and problem+json parsing. Browser Google re-login via FE proxy, budget empty/editor data and grouped transaction form inspected; drafts canceled. Owner acceptance and full browser persistence UAT pending. [Evidence and limitations](tickets/API-SCREENS-01-DETAIL_DESIGN.md).

- Savings slice: target-date red/green regression, update omit/clear checks, `go test ./...`, Swagger generation pass. Frontend design, shared SSR (goal progress/selection/form), ledger tests and build pass. Browser confirmed grouped selector, goal form and Save/cancel states without saving test data. Full goal create/reload/transaction UAT pending. Transfer/report/notifications are not verified features. [Work evidence](tickets/TICKET-06-01-DETAIL_DESIGN.md).

- Wallet screen spec normalized with source-backed Money Lover references and owner select/immutable-type decisions. Docs review complete; UAT not required for this docs-only update. Runtime select change and per-type detail UI are pending; earlier tests do not cover them.

- UI-WALLET-02: check:design, test:design (7 guard regressions plus SSR form/switch/type contracts), test:transactions (4), build all pass. Add Wallet uses shared sheet/card/switch/type picker; normalized both supplied references. Browser visual and new-form API UAT pending. [Evidence](tickets/UI-WALLET-02-DETAIL_DESIGN.md).

- Add Wallet note removal: `rtk proxy npm run check:design`, `rtk proxy npm run test:design`, `rtk proxy npm run build` passed in `app/`. SSR confirms creation omits “Ghi chú” and editing retains it. Owner visual UAT pending. Docs review: wallet design, context, backlog and changelog reconciled; API/schema unchanged, no ADR required.

- Wallet empty-state follow-up: Overview and WalletManagementPanel now use the tested `StatusMessage` plain variant as requested; design checks/tests/build rerun, visual UAT pending.

- Owner reports previously tested bugs are OK; this is user-reported acceptance of tested fixes, not exhaustive product verification.
- [UI-EMPTY-01](tickets/UI-EMPTY-01.md): `npm run check:design`, `npm run test:design`, and `npm run build` pass in `app/`. SSR regression checks cover plain status without card decoration and preserved danger feedback. New empty-state visual UAT pending owner review.
- Local Google OAuth start endpoint verified HTTP 302 with `redirect_uri=http://localhost:4173/api/v1/auth/google/callback` via FE proxy; session-only launch configuration, no tracked credentials.

### UI-BASE-01 normalization

- Status: implemented / verified for shared UI foundation and standards. [Exact evidence and docs review](tickets/UI-BASE-01-VERIFICATION.md).
- `rtk proxy npm run test:design` in `app/`: pass (6 guardrail regression fixtures and shared SSR contracts).
- `rtk proxy npm run build` in `app/`: pass (source guard, docs guard, typecheck, Vite). Docs guard: 17 Markdown specs, 57 valid links, no PNG/HTML exports.
- Chrome fixture: theme primary color/min touch height, selected keyboard tab, disabled/loading CTA, dialog open/focus trap/Escape/return-focus and visual card/tree/gauge checks passed.
- Root cause reproduced before migration: unimported theme, duplicate literal token values, native controls outside atoms. Owner authorized specs + guardrail + refactor; this supersedes prior UI-BASE-01 removal proposal.
- Product-level CRUD/OAuth/ledger acceptance is unchanged; fixture is shared UI proof, not a new business UAT sign-off.

| Area | Proof | Result | Evidence |
| --- | --- | --- | --- |
| Base component styling | TypeScript + production build and authenticated browser inspection | pass | `npm run typecheck && npm run build` in `app/`; canonical `BaseCategoryTree` now powers both Group management and the Reports adapter. Browser proof at `/account/groups`: `nested` source geometry with global colorful icon catalog, 40px parent / 32px child icons, `pl-10` children, a 2px trunk starting at the first child row and curved branches. Regression proof: `BaseButton` now owns inline flex alignment; Quay lại and Nhóm mới both render icon plus label in one centered row. Base card/form/row radii were reduced one token step; semantic circles and pills remain intact. Browser check at `/`: the center create button is vertically contained in the bottom-navigation frame. |
| Database migrations | Dev PostgreSQL migration CLI | pass | `go run ./cmd/migrate up && go run ./cmd/migrate version` in `backend/`; version `8`, `dirty=false`; SQL migrations own schema and the system icon catalog. |
| Default category catalog | Dev PostgreSQL migration, Go test and browser list | pass | `go run ./cmd/migrate up && go run ./cmd/migrate version && go test ./...` in `backend/`; version `8`, `dirty=false`; migration `000008` assigns catalog icon keys for deployment as well as development. |
| API routing and Swagger generation | Go test + code generation | pass | `go generate ./cmd/api && go test ./...` in `backend/`; generated Swagger includes the implemented applicable-wallet endpoint. |
| Group management CRUD | Use-case tests, FE build and owner closure | pass | `go test ./...`, `go generate ./cmd/api`, `npm run typecheck && npm run build` pass. Use case proves owner scope, level/cycle validation and re-parenting children after a personal-parent deletion. System metadata `PATCH` is refused while the dedicated applicable-wallet endpoint remains permitted. Owner requested ticket closure after UI review. |

| Area | Proof | Result | Evidence |
| --- | --- | --- | --- |
| Frontend routing | TypeScript compile check | pass | `npm run typecheck` in `app/` |
| Schema and system catalog | Backend unit test | pass | `go test ./...` in `backend/` |
| Categories API auth guard | HTTP manual check | pass | `GET /api/v1/categories` without bearer returned `401` |
| Unified response contract | Go unit tests | pass | `TestOKUsesDataEnvelope`, `TestFailUsesProblemDetails`; `go test ./...` |
| Group management UI | Frontend production build | pass | `npm run typecheck && npm run build` in `app/`; runtime requires authenticated browser session |

- CORS preflight: `OPTIONS /api/v1/profile` with `Origin: http://localhost:4173` returned `204` with allow-origin, credentials, methods and headers.
- API health: `GET /api/v1/health` returned `{"status":"ok"}`.
- Google OAuth start: `GET /api/v1/auth/google` returned `302` to Google with redirect URI `http://localhost:8080/api/v1/auth/google/callback`.
- Swagger UI: `GET /api/v1/docs/index.html` and `GET /api/v1/docs/doc.json` returned `200`; generated paths cover only the implemented auth/profile/home/health endpoints.

Policy lives in `docs/standards/VALIDATION.md`. This matrix is runtime project state and should change as work is planned, implemented, changed, or retired.

### Wallet and basic ledger runtime verification

- Status: core wallet management and the approved basic income/expense ledger are implemented and verified for review. [Wallet evidence](tickets/TICKET-01-02-VERIFICATION.md) · [Ledger evidence](tickets/TICKET-02-01-VERIFICATION.md).
- `rtk go test ./...` in `backend/`: pass, including owner-scoped handler rules, wallet type validation, category-wallet applicability, credit-ledger exclusion and derived balances.
- `rtk go run ./cmd/migrate up` and real HTTP UAT against dev PostgreSQL: pass. Restricted category returned `400`; create/update/delete balance sequence was 115,000 → 113,000 → 93,000; deleting the wallet cascaded its final transaction.
- `rtk go generate ./cmd/api`: pass; Swagger includes wallet `current_balance` and current transaction endpoints.
- `npm run test:transactions`, `npm run test:design`, and `npm run build` in `app/`: pass.
- Chrome UAT through the Vite proxy: create wallet, create expense, list, edit, synchronized Header/Wallet balance, and cleanup all passed. Two stale cross-panel refresh gaps were found during UAT, fixed, and rechecked.
- UAT boundary: receipt/OCR, jar assignment, transfer/adjustment and credit ledger are not required for the approved basic slice and remain open BA/follow-up scope. Human acceptance sign-off remains human-owned.

## Status Values

| Status | Meaning |
| --- | --- |
| planned | Accepted as intended behavior, not implemented |
| in_progress | Actively being built or verified |
| implemented | Implemented and evidence exists |
| changed | Contract or expected proof changed after earlier implementation |
| retired | No longer part of the accepted project contract |

## Matrix

| Requirement | Phase | Ticket/Bug | Contract/Behavior | Unit | Integration | E2E | UAT | Platform/Manual | Docs Review | Status | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| REQ-02 | not_applicable | [TICKET-01-02](tickets/TICKET-01-02-quan-ly-vi.md) | Core wallet CRUD, type fields, permanent-delete impact, derived/current total balance | yes | yes | yes | yes | not_required | yes | implemented | [Wallet verification](tickets/TICKET-01-02-VERIFICATION.md); human review remains open for later statement/adjustment scope |
| REQ-04 | not_applicable | [TICKET-02-01](tickets/TICKET-02-01-ghi-thu-chi.md) | Basic income/expense CRUD with wallet/category compatibility and synchronized balances | yes | yes | yes | yes | not_required | yes | implemented | [Ledger verification](tickets/TICKET-02-01-VERIFICATION.md); receipt/jar acceptance remains outside verified slice |
| REQ-01, NFR-04 | not_applicable | [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) | UTC/date-only/account timezone and fixed VND | yes | yes | yes | yes | not_required | yes | in_progress | Automated calendar/API/DB proof and migration 15 recorded above; owner timezone-switch and visual UAT pending |
| REQ-02–05, REQ-09–11 | not_applicable | [BA tickets](tickets/README.md) | Wallet, transaction, budget, savings/credit/debt accounting | yes | yes | yes | yes | not_required | yes | planned | Business rules documented; open linked-deletion/credit scenarios remain |
| REQ-06 | not_applicable | [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) | Optional jar grouping, soft warnings, monthly configuration and cumulative view | yes | yes | yes | yes | not_required | yes | in_progress | PostgreSQL initialization/ownership tests, expense-only selector test and build pass; real owner CRUD/history UAT pending |
| REQ-07–08 | not_applicable | [BA tickets](tickets/README.md) | Ordinary recurring and automatic Travel Mode excluding recurring | yes | yes | yes | yes | not_required | yes | planned | REC/TRV rules; backdating/catch-up decisions still open |
| REQ-12–13 | not_applicable | [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), [TICKET-07-04](tickets/TICKET-07-04-tong-ket-thang-ai.md) | Recalculable monthly reports, independent notes, context/AI and Insider | yes | yes | yes | yes | not_required | yes | in_progress | Live numeric month API/note and automatic local completion implemented; UAT pending; AI narrative/context and Insider remain unimplemented |
| REQ-14 | not_applicable | [BA tickets](tickets/README.md) | Moving weighted-average portfolio | yes | yes | yes | yes | not_required | yes | planned | AST formulas documented; funding/historical deletion open |
| REQ-15–18, NFR-01–03, NFR-05–07 | not_applicable | [BA tickets](tickets/README.md) | Keys/API/AI/offline/data lifecycle | yes | yes | yes | yes | yes | yes | planned | Product contract consolidated; no runtime implementation evidence added |

Proof columns above describe required product validation, not test execution. The current [BA ticket index](tickets/README.md) maps these rules to draft parent/child tickets. Phases remain not applicable; ticket creation does not count as implementation proof. Owner document review is pending; no product UAT has been marked passed.

## Documentation Review — 2026-09-13

- Scope: canonical product/business/report docs, requirements/stories, docs guide, context/queue, roadmap cleanup and standards relocation.
- Docs review: passed for captured decisions, trace targets, obsolete product wording and unchanged relocated standards. Open business questions are explicitly retained in [BUSINESS_RULES.md](../requirements/BUSINESS_RULES.md).
- `rtk git diff --check`: pass, exit 0.
- Read-only Node check: 17 docs inspected, 49 Markdown/frontmatter file targets checked, zero missing targets; six product docs have expected review metadata (README excepted).
- Read-only comparison with Git HEAD: all nine relocated standards are byte-identical to their former tracked files.
- Runtime tests/UAT: not run; this update changes documentation only. Product validation remains planned and owner review remains pending.
- API/ERD/runtime docs: no runtime or schema implementation changed; existing unrelated architecture edits were preserved. Technical ADRs remain a prerequisite where future implementation introduces durable architecture/schema decisions, not evidence supplied by this review.
- Reconciliation: requirements, report contract, stories, context, queue note, roadmap and changelog updated. No implementation ticket/phase created.
- Removed material: two superseded product drafts and one inherited Harness CLI phase example. Canonical product content is in `docs/requirements/`; tracked originals can be recovered from Git history. Standards were moved, not discarded.

### Reproduce Link And Trace Check

```sh
rtk proxy node -e 'const fs=require("fs"),path=require("path");let files=["docs/README.md","docs/work/ROADMAP.md",...fs.readdirSync("docs/standards").map(n=>"docs/standards/"+n),...fs.readdirSync("docs/requirements").map(n=>"docs/requirements/"+n)];let failures=[],count=0;for(const file of files){let s=fs.readFileSync(file,"utf8");for(const m of s.matchAll(/\[[^\]]*\]\(([^)]+)\)/g)){const ref=m[1].split("#")[0];if(ref&&!/^[a-z]+:/i.test(ref)){count++;if(!fs.existsSync(path.resolve(path.dirname(file),ref)))failures.push(file+": "+ref);}}let fm=s.match(/^---\n([\s\S]*?)\n---/);if(fm)for(const m of fm[1].matchAll(/^\s+[a-z_]+:\s+(\S+\.md)\s*$/gm)){count++;if(!fs.existsSync(path.resolve(path.dirname(file),m[1])))failures.push(file+": trace "+m[1]);}}console.log(JSON.stringify({files:files.length,linksAndTraces:count,failures},null,2));if(failures.length)process.exit(1);'
```

## BA ticket review

- Owner request: parent tickets with smaller children, concise BA scope and acceptance criteria.
- Result: 11 parents and 31 children, all `draft`; REQ-01 through REQ-18 covered. Common quality requirements are referenced by the shared ticket guide.
- Check: read-only ticket inventory verified unique IDs, draft status, reciprocal parent/child links, at least three acceptance criteria per child, no placeholders/trailing whitespace, and 106 valid Markdown file links. Pass, zero failures.
- `rtk git diff --check`: pass, exit 0.
- Docs review: scope and acceptance use business language; open questions remain in affected tickets; Reports/AI responsibilities avoid separate conflicting formulas.
- UAT: pending future implementation and owner acceptance. Runtime tests not required for this documentation-only breakdown.
- Reconciliation: ticket index, backlog, context, product docs entry point, roadmap note and changelog updated. No runtime, API/schema or architecture change; no technical ADR needed for the breakdown.

## Rules

### Category UI and ASCII ERD docs review

- Naming review: owner chọn `user`; sơ đồ và FK dùng tên đích, hiện trạng mapping cũ chỉ ghi trong mục migration. Kiểm tra `rtk git diff --check`; chưa chạy migration hoặc runtime tests cho thay đổi tên trong docs.

- Phạm vi: Tài khoản → Quản lý nhóm → chọn nhóm → Sửa; nhóm cha, category và ví áp dụng. [UI design](../design/system/DESIGN.md#account--quản-lý-nhóm) liên kết [ASCII ERD](../architecture/ERD.md#ascii--mô-hình-ví-và-nhóm-để-review) và [ticket](tickets/TICKET-01-03-tong-vi-danh-muc.md).
- Docs review: đã phân biệt user model hiện có và quan hệ nhóm/ví đề xuất; giữ catalog cũ; không tự chốt ý nghĩa category, scope rỗng hoặc chiến lược seed. Base component reuse bắt buộc.
- Kiểm tra: `rtk git diff --check` và kiểm tra file đích của Markdown links bằng Node. Runtime/UAT không chạy vì thay đổi chỉ ở docs; UI/API/migration nhóm chưa implement. Review mô hình dữ liệu còn pending; chưa có quyết định kiến trúc được áp dụng cần ADR implementation.

### Wallet research and docs review — 2026-09-13

- Evidence: official Money Lover support for balance adjustment, wallet management, goal/credit wallets; exact URLs and limits recorded in [DESIGN-01-02](tickets/TICKET-01-02-DETAIL_DESIGN.md).
- Reviewed old category catalog `0011_phase002_category_catalog.sql` and research sections 17–18; pinned reuse in wallet/category tickets.
- Corrected deletion/duplicate-name decisions across design, backlog, ERD, architecture and context. Research is not app UAT; no wallet API/schema/seed executed. Full technical design still needs review.
- Documentation-only verification: diff whitespace and local Markdown link checks; product tests/UAT not run for this research update.

### Finance Assistant planning and implementation — 2026-09-22

- [AI-ADVISOR-01](tickets/AI-ADVISOR-01-DETAIL_DESIGN.md) and [V1 task plan](../superpowers/plans/2026-09-22-finance-assistant-v1.md): design remains in review; advisor implementation is **partially shipped locally** as a JWT/API-key synchronous JSON pilot.
- Docs proof: `rtk git diff --check` passes; Node local-link check reports 8 valid relative targets across the new design/plan. Reviewed current code evidence and official chat-library docs; no dependency installed.
- Runtime/provider/E2E/UAT: backend race/DB and frontend type/design/build are proven for the local slice; local user API-key unit, PostgreSQL, middleware and Redis stream audit proof is now present. Configured provider, typed cards, SSE/recovery, durable audit delivery/alerting, browser state and public deployment remain unproven. Detailed gates are specified in the design/plan.
- API/schema/runtime changed for the local advisor slice (Stage v1 baseline plus JWT routes/UI). Existing transfer work remains independent; Feedback implementation and deployment checks remain separate.

Technical-plan follow-up:

- [Implementation guide](../superpowers/plans/2026-09-22-finance-assistant-technical-guide.md) added; [task plan](../superpowers/plans/2026-09-22-finance-assistant-v1.md) expanded to 8 tasks / 42 ordered subtask gates. Covers SQL/Go ports, tool defaults, recovery/drilldown APIs, SSE/React, exact accounting fixtures and operational limits.
- Code evidence from plan review: jar `ListMonth` can initialize config, API helper buffers response text, composer copy is entry-specific, navigation had only five existing slots, and migrate CLI has no down command. The runtime slice now addresses the jar path, composer/navigation wiring, query/overview boundary, local API-key auth and best-effort Redis access audit; SSE/recovery/durable audit remain open.
- Docs validation: local Node checker over design/plan/guide verified 12 relative link targets, 4 JSON examples and 27 paired fenced blocks; no failures. `rtk git diff --check` passes. Placeholder scan of plan/guide returned no matches.
- Runtime/code snippets in the guide remain instructional; the actual local implementation is verified in the vertical-slice checkpoint below. Public docs now include the implemented tool catalog, routes and API-key contract; proposed SSE/recovery contracts remain marked as unreleased.

Feedback implementation completion checkpoint — 2026-09-22:

- PASS: backend `rtk go test -race ./...` after wiring `main.go`, lifecycle locking, audit context, handler validation and Swagger annotations.
- PASS: frontend `rtk npm run typecheck`, `rtk npm run check:design`, `rtk npm run build` after Account → Feedback route/panel/service.
- PASS: `rtk go generate ./cmd/api`; generated Swagger includes feedback/changelog models and routes.
- PASS: public/agent API contract documented in [FEEDBACK_API](../architecture/FEEDBACK_API.md). Redis audit remains best-effort after commit; no finance access is granted to the service token.
- NOT RUN: browser feedback E2E and real API roundtrip; no dedicated script currently exists in the dirty worktree. Residual risk is recorded; do not claim production deployment or full user UAT.

Finance Assistant Task 1 foundation — 2026-09-22:

- PASS: entity normalization tests cover invalid/reversed/overlong calendar ranges, DST next-local-midnight conversion, JavaScript-safe money bounds, filter defaults/deduplication and zero-baseline percentage semantics.
- PASS: PostgreSQL fixture through `TEST_DATABASE_URL=postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable go test ./internal/infrastructure/repository -run TestFinanceQueryPostgresSummary -count=1 -v` proves owner scope, 1000/150/850 report totals, 55 full-scope rows versus a 50-row page, cursor replay/scope rejection, category subtree filtering, baseline-zero comparison and no jar/config row mutation.
- PASS: shared `FinanceQueryService` delegates validated owner-scoped bundles; full backend suite remains green. Advisor persistence/provider/HTTP/UI proof is recorded in the vertical-slice checkpoint below; SSE is still open.

Finance Assistant local vertical slice — 2026-09-22:

- PASS: migration `000018_advisor_conversations` applied locally (`version: 18 dirty: false`); PostgreSQL proof covers one conversation per owner, same-request replay, changed-payload conflict, active-run busy guard and cancellation.
- PASS: tool registry/provider/orchestrator/parts tests cover eight read-only tools, `get_transaction` owner isolation, unknown/owner override rejection, provider tool-call round trip, bounded loop, server-built view parts and no write/SQL/confirmation tools.
- PASS: JWT advisor routes now include a model-free `/overview` summary, real Account navigation slot, read-only composer mode, owner-scoped localStorage, masked assistant values, typecheck/design/build and full backend race/DB tests. The local provider-backed chat remains disabled when AI credentials are absent, while overview/history remain available.
- PASS: local API smoke with fixture login against `127.0.0.1:18080` returned `enabled=false` without AI credentials and a real account-local September overview (`income=0`, `expense=0`, `net=0`, `count=0`) without a model call; the route was then stopped. No provider secret was logged or committed.
- PASS: local user API-key creation/list/revoke, digest-only persistence, scope implication/denial, expiry/revocation and advisor middleware owner-isolation tests; the Stage v1 baseline applied to a fresh disposable database (`version: 1 dirty: false`).
- PASS: advisor auth audit writes redacted access metadata to the dedicated Redis stream `mypocket:audit:advisor`; integration proof ran against local Redis.
- NOT RUN: real browser chat round-trip through a configured provider, SSE/reconnect transport, durable audit delivery/alerting and production deployment. These are release gates, not claimed by this local slice.

- Add or update a row when a requirement, ticket, bug, public contract, or accepted behavior is created or changed.
- Mark proof columns `yes`, `no`, or `not_required`.
- Link evidence to `docs/templates/TEST_VERIFICATION.md`, ticket verification sections, UAT, docs review, release notes, or command output summaries.
- Do not set `implemented` until required proof has evidence.
- If a proof type is `not_required`, record the reason in the linked ticket, bug, or verification artifact.

### AI Advisor bug-fix verification — 2026-09-22

| Area | Evidence | Result |
| --- | --- | --- |
| Tool contract | `go test ./internal/usecase -run TestAdvisorToolDefinitionsExposeModelParameters` | PASS: all eight read tools expose model-visible properties; range/month required fields are explicit |
| Provider wire contract | `go test ./internal/infrastructure/ai -run TestAdvisorClientEncodesOpenAICompatibleToolCalls` | PASS: assistant calls use `type=function` and nested function name/arguments; tool result keeps `tool_call_id` |
| Orchestration cap | `go test ./internal/usecase -run TestAdvisorOrchestratorAllowsFinalAnswerAfterFourToolCalls` | PASS: four read calls may be followed by one final answer; further calls remain rejected |
| Credential lifecycle | `go test ./internal/usecase -run 'TestAdvisorOrchestratorRevalidatesPrincipalBeforeProviderAndTool|TestUserAPIKeyValidatePrincipalRejectsRevokedKey'` | PASS: provider/tool boundaries re-read principal; revoked key is rejected |
| Cancellation commit guard | `TEST_DATABASE_URL=... go test ./internal/infrastructure/repository -run TestAdvisorPostgresStartRunIsIdempotentAndBusyScoped` | PASS: cancelled run rejects late assistant message with lease-lost error |
| In-process cancellation | `go test -race ./internal/usecase -run TestAdvisorServiceCancelInterruptsInFlightProvider` | PASS: cancel endpoint path cancels the provider context and submit returns an error |
| History pagination | `npm run test:ai-history` | PASS: older pages merge without duplicate boundary messages and remain chronological |
| Context bound | `go test ./internal/usecase -run TestBuildAdvisorContextKeepsOnlyTwelveRecentMessages` | PASS: provider context keeps the newest 12 messages in chronological order |
| Assistant cards | `npm run typecheck && npm run check:design && npm run build` | PASS: normalized summary/comparison/search/budget plus wallet/goal/jar cards compile and satisfy base/design guards |

The above is local regression proof. Configured-provider browser chat, SSE/reconnect, durable Redis audit delivery/alerting and production deployment remain unverified release gates.
