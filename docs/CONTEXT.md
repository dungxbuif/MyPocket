---
artifact_type: project_context
id: CONTEXT
status: active
owner: shared
human_fields:
  - current_focus
  - open_questions
  - priority_override
ai_fields:
  - recently_touched_areas
  - recent_decisions
  - next_steps
  - queue_summary
shared_fields:
  - current_status
  - active_backlog
  - current_queue_focus
  - active_phase
  - active_ticket
  - active_bug
updated: 2026-10-06
---

# Project Context

2026-10-06 native simulator auth fixture correction: the API-backed UI harness
now uses a DEBUG-only in-memory session store when launched with
`-UITestAuthenticatedAPI`. The previous unsigned `CODE_SIGNING_ALLOWED=NO`
run reached the disposable OAuth exchange with HTTP 200 but stayed on the login
screen because Keychain returned `errSecMissingEntitlement (-34018)`. Production
and normal app builds continue to use `KeychainSessionStore`; no token values
are logged. The full iPhone 16 `LOCAL_API_UI` suite passes 15/15 after the
correction.

2026-10-06 native transaction editor parity: ordinary expense editing now
exposes the account-month hũ selector, report-inclusion toggle, and account-
timezone date/time field. Active hũ options are limited to eligible basic/goal
expense wallets; an existing inactive hũ remains visible with an explicit save
error instead of being silently remapped. Type, wallet and category changes
clear incompatible hũ assignments. Native XCTest passes 61/61, the signed
iPhone 16 `LOCAL_API_UI` suite passes 15/15, the Release warnings-as-errors
build and base/project guards pass.

2026-10-06 native AI proposal parity: the proposal editor now preserves the
account-local date/time, keeps an inactive existing hũ visible for explicit
resolution, rejects stale hũ values instead of silently clearing them, and
excludes the transfer-out category from hũ choices. Report inclusion remains
part of the versioned proposal patch; deterministic proposal review and policy
coverage pass.

## Latest AI entry correction — 2026-09-23

[Review-first bug](work/bugs/AI-ENTRY-REVIEW-FIRST-DETAIL_DESIGN.md) is in_review pending owner UAT. Replaced conflicting clarification-first prompt rules with partial draft extraction, within-batch duplicate merging and inferred/default supplied wallets. Real stored nine-image OCR replay changed from zero drafts to 30 with owned wallets after raising the truncated 2048-token response budget; the current provider envelope is 32768 tokens. The application no longer caps proposal count; only provider response/transport envelopes remain. Held-out duplicate/distinct-date/missing-date fixture, 31-draft parser test, full Go tests, focused race tests and backend build pass. Local API restarted; backend health and frontend :4173 both HTTP 200. No ledger writes from replay, no production deployment; owner still reviews amounts/duplicates and existing requests retain cached outcomes.

2026-09-23 diagnostics follow-up: AI entry and Finance Assistant provider calls now emit privacy-safe stage/latency/status/count diagnostics, and `AI_STORE_USAGE=true` persists normalized request/finish/latency/token metadata internally without exposing it in public responses. User text is sent as a separate `user_instruction` field, so explicit filters such as “chỉ lấy giao dịch tháng 9” guide selection while OCR/source text remains untrusted evidence. User-facing repository reads were audited for owner predicates; the only owner-independent operation is the internal attachment cleanup sweeper, which does not return financial data. Streaming remains deferred; the synchronous JSON contract and bounded 180s model/210s extraction timeouts are the current behavior.
2026-09-23 runtime regression fix: initializing empty `model_usage` maps on new AI entry/advisor rows prevents the non-null JSONB columns from receiving SQL `NULL`, which had surfaced as a generic 500 before provider work. The database default remains as a safety net; creation and integration tests now cover the path.
2026-09-30 production feedback fix: AI entry normalizes recognized date-only model output (`yyyy-MM-dd`, `dd/MM/yyyy`, `dd-MM-yyyy`) in the account timezone and explicitly flags the missing time for review. Exact duplicate drafts from one OCR/text submission are merged while preserving duplicate-row questions. Backend package tests pass (176 tests); production API image `feedback-api-ai-date-20260930-0011` is healthy.
2026-09-30 auth feedback fix: login and Google OAuth now issue short-lived JWT access tokens plus rotating opaque refresh tokens stored hashed in Redis; `POST /api/v1/auth/refresh` consumes each refresh token once. The browser retries one expired session request, never API-key requests. Production API `feedback-api-refresh-20260930-0041` and web `feedback-ui-refresh-20260930-0032` are healthy.
2026-09-23 wallet/ledger UI follow-up: wallet selection now opens a shared detail panel for basic, goal and credit wallets; the transaction ledger has a top wallet scope select with `Tổng cộng` or an owner wallet. Credit detail is read-only until its dedicated debt ledger exists.
2026-09-24 ledger correction: `/transactions` has one wallet selector. Basic and `Tổng cộng` now share Week/Custom controls, period cashflow card and grouped rows; basic adds only a wallet-ID predicate. Goal alone shows progress and full monthly history; provisional credit keeps the shared period layout read-only. No shared All-time tab. Live Chrome verified basic/aggregate parity and previous-week empty state after the correction; owner mobile sign-off and deployment are not claimed. Credit statement/payment/detail contract is not approved; its visible rows cannot open the ordinary editor. Travel Mode is now an implemented event linkage, not a wallet type; the `/account/travel` panel manages one active owner-scoped event and ordinary transaction auto-linking, while report narrative remains a follow-up. See [ticket](work/tickets/UI-LEDGER-WEEK-DETAIL_DESIGN.md) and [Travel Mode contract](design/pages/travel/README.md).
2026-09-30 wallet UI reconciliation: the transaction ledger scope control now uses a compact `WalletScopeSelector` trigger and the shared `WalletSelectionList` inside `BaseBottomSheet`, matching the supplied wallet-selector reference while preserving aggregate/basic/goal/credit filtering semantics. The selector can route to wallet management for add/edit; no API or schema change.
2026-10-05 simulator/integration audit: iPhone 16 native unit/UI smoke (19 + 3) passed, the Go backend race suite passed against disposable PostgreSQL + Redis, transfer integration/browser probes now cover idempotent pair update/delete, and feedback was verified through JWT, scoped API key, service-token agent lifecycle and Redis audit events. Local fixtures were cleaned. Provider-backed AI/OCR, real Google iOS OAuth, physical-device and authenticated staging/production proof remain open release gates.
2026-10-06 AI browser regression: a fixed September/October fixture date had drifted outside the current account-local week and caused a false browser failure. The fixture now derives today/yesterday in `Asia/Ho_Chi_Minh`; the complete AI contract + Chromium flow is green again.
2026-10-06 quality gate: backend AI/feedback test fixtures were normalized with `gofmt`; `go vet ./...`, targeted Go tests and a Swift Debug simulator build with warnings-as-errors pass.
2026-10-06 native parity fix: session restore now refreshes expired access tokens before Face ID and clears invalid sessions safely. Category editing now uses a shared multi-select base and preserves every selected wallet ID instead of collapsing a multi-wallet category to its first scope. Refresh rejection during an authenticated request now invalidates the AuthViewModel instead of leaving a stale shell. Shared native buttons now use their full rendered hit shape, and the simulator route sweep opened all 13 “Thêm” destinations. Malformed ledger timestamps are excluded from period filters, wallet detail renders the complete API result, and assistant polling keeps a timed-out run cancellable. iPhone 16 Simulator unit tests pass 25/25 and UI smoke passes 4/4.
2026-10-06 native parity fix: session restore now refreshes expired access tokens before Face ID and clears invalid sessions safely. Category editing now uses a shared multi-select base and preserves every selected wallet ID instead of collapsing a multi-wallet category to its first scope. Refresh rejection during an authenticated request now invalidates the AuthViewModel instead of leaving a stale shell. Shared native buttons now use their full rendered hit shape, and the simulator route sweep opened all 13 “Thêm” destinations. Malformed ledger timestamps are excluded from period filters, wallet detail renders the complete API result, and assistant polling keeps a timed-out run cancellable. iPhone 16 Simulator unit tests pass 24/24 and UI smoke passes 4/4.
2026-10-06 native simulator fixture: Debug launch argument `-UITestAuthenticatedAPI` now exchanges against the disposable local OAuth fixture, stores the real JWT in Keychain and renders the overview through live URLSession/API requests. Manual iPhone 16 screenshot `/tmp/mypocket-ios-simulator-api-auth.png` shows the authenticated empty-owner state; this is local validation only and does not certify production OAuth.
2026-10-06 native transaction parity: bulk-delete now matches the web contract (base selection, no transfer/credit/adjustment rows, max 100 IDs, atomic confirmation and retained error state). The typed API client now handles backend `204 No Content` responses for all empty mutations; native XCTest count is 28/28 and UI smoke remains 4/4.
2026-10-06 native Assistant parity: Finance Assistant now restores one owner-keyed opaque conversation ID, reloads the latest 50 messages and read-only monthly overview on entry, pages older messages, and keeps message text out of UserDefaults. The iPhone 16 API-backed UI smoke passed against the disposable local API with the capability-disabled/overview-200 contract; native XCTest count is 29/29 and standard UI smoke remains 4/4.
2026-10-06 native credit parity: the credit ledger entry/payment form now preserves entry kind, applicable category, source wallet, amount and optional note through the native base selectors/inputs. Category kind/wallet filtering is covered by XCTest; native XCTest count is 30/30.
2026-10-06 native wallet-detail parity: basic/goal wallet detail now loads categories, renders account-timezone monthly groups with signed amounts and labels, shows income/expense totals, and opens the shared transaction editor for add/edit. The API-backed simulator smoke seeded and removed a disposable wallet after opening the editor.
2026-10-06 native transaction-ledger parity: the shared overflow menu now owns period/grouping/transfer/adjustment/bulk actions; rows include category and wallet context with signed amounts, and goal/credit wallet scopes suppress ordinary period controls. Standard simulator UI smoke now passes 5/5.
2026-10-06 native selector/jar parity: API enum values now render localized labels through the shared selector base, jar mode changes clear stale drafts, and zero-percent allocations are rejected like the web editor. The clean iPhone 16 simulator run passes 35/35 native unit tests and 5/5 standard UI tests (3 API-backed tests also pass with disposable local fixtures; all fixtures were deleted with HTTP 204).
2026-10-06 native transaction boundary parity: ordinary and AI proposal editors now exclude credit wallets, invalid categories clear when wallet/type changes, and aggregate credit rows remain read-only outside the dedicated credit ledger.
2026-10-06 native scope parity: budget categories are expense-only and budget/adjustment wallet selectors exclude credit wallets, matching the ordinary ledger boundary.
2026-10-06 native budget parity: ended budgets are view/delete-only and no longer expose an edit/save path after their end date.
2026-10-06 native portfolio date parity: trade date selection now follows the account timezone and serializes at local noon to avoid device-timezone drift.
2026-10-06 native category scope parity: system categories now expose a lock/read-only metadata state, hide delete, and update only applicable wallet IDs through the dedicated wallet-scope route; personal categories retain full editing. Native XCTest is 49/49 and the signed simulator More-route sweep is 1/1.
2026-10-06 native jars month/cumulative parity: Hũ chi tiêu now uses a bounded account-local month selector, actual/allocated/unassigned metrics, advisory over-allocation warnings, active/history action gating, and selected-jar cumulative reports that clear stale state on month changes. Native XCTest is 50/50, Release build passes with warnings-as-errors, and the signed More-route sweep is 1/1.
2026-10-06 native savings-summary parity: goal-wallet detail now shows reached/remaining amount, account-local target-day status, bounded progress and target amount through shared base components. Native XCTest is 52/52 and the signed More-route sweep is 1/1 in 124.194s.
2026-10-06 native portfolio draft parity: asset symbol/name drafts are trimmed and whitespace-only values are rejected before API submission; the portfolio header now exposes the shared refresh action. Native XCTest is 53/53 and the signed More-route sweep is 1/1 in 124.144s.
2026-10-06 native feedback draft parity: title/description are normalized and feedback type is restricted before JSON or screenshot multipart submission. Native XCTest is 54/54 and the signed feedback-entry UI test is 1/1 in 9.821s.
2026-10-06 native month-summary parity: Tháng now decodes jar usage, uses bounded account-local month selection, renders period/status/category/jar details, protects dirty notes and links to the selected month’s Hũ view. Native XCTest remains 54/54 and the signed More-route sweep is 1/1 in 122.295s.
2026-10-06 native simulator release check: the Release simulator build passed with warnings treated as errors, project/base-component guards passed, and `com.mypocket.ios` launched on iPhone 16. The post-commit cross-surface rerun passed 48/48 native unit tests, 14/14 signed API-backed UI tests, backend race tests, and the web/API/Chromium/PWA gates; the disposable fixture was empty afterward. Provider-backed AI/OCR, real Google OAuth, authenticated staging/production and physical-device validation remain open release gates; no production OAuth claim is made.
2026-10-06 native AI proposal review follow-up: review now requires an explicit income/expense selection for unknown model output, exposes active account-month hũ choices only for ordinary expense rows, filters/clears incompatible categories, and lets the user choose report inclusion through shared base selectors. The red-to-green contract passes; the deterministic iPhone 16 review fixture passes and the full signed `LOCAL_API_UI` suite is 15/15. Provider-backed extraction remains an external release gate.
2026-10-06 native bug audit: wallet detail month headers now use validated calendar interpolation instead of literal placeholder text; Travel Mode trims required names and rejects malformed/reversed optional date ranges before API submission. The red-to-green regression tests and iPhone 16 run now pass 37/37 native unit tests and 5/5 standard UI tests.
2026-10-06 native UI test isolation: `testAppLaunches` now starts with the DEBUG-only `-UITestUnauthenticated` fixture, clearing only simulator Keychain/biometric state so a prior local login cannot change the assertion. Targeted red-to-green proof and the full iPhone 16 default UI suite pass 9/9 (5 passed, 4 explicit API-fixture skips).
2026-10-06 native API UI fixture isolation: `LOCAL_API_UI` now creates/deletes its own basic and credit wallets through local `:8080` after fixture OAuth. Wallet detail, credit ledger, Assistant and AI-entry capability tests pass, the full opt-in UI suite is 9/9, and a post-run owner check finds zero leftover `Fixture*` wallets.
2026-10-06 native AI-entry capability gate: the base document picker is disabled until `files_configured=true` and while processing; text-only entry remains usable when OCR/private storage is unavailable. Native unit tests are 38/38 and the local API-backed capability UI assertion passes.
2026-10-06 native ledger week-period parity: aggregate/basic scopes now use account-local Monday–Sunday `Theo tuần` plus bounded previous/next navigation and `Tùy chọn`; the incorrect all-time tab was removed and the summary uses `Tiền vào`/`Tiền ra`. Native unit tests are 41/41 and the API-backed ledger UI assertion passes.
2026-10-06 native goal target-date parity: wallet editor date-only parsing and serialization now use the account IANA timezone instead of UTC/device time. Native unit tests are 42/42.

## Field Ownership

- Human owns project intent, priority overrides, and unresolved product questions.
- AI owns concise state refreshes after work: touched areas, recent decisions, next steps, and queue summary.
- Shared fields can be updated by either human or AI, but AI must not silently override human priority.

## Current Status

- 2026-09-22 Stage v1 reset: the historical incremental SQL files were squashed into `backend/migrations/000001_stage_v1.up.sql` with a non-destructive down file. A fresh disposable database applies cleanly at version 1. Existing databases that still report an older historical version must be recreated/reset in the staging environment before using this baseline; production data is never dropped automatically. The former legacy reference tree was removed; this repository is now the canonical implementation and documentation source.

- 2026-09-22 planning follow-up (before implementation): owner requested detailed writing-plan and technical instructions. Added [technical guide](superpowers/plans/2026-09-22-finance-assistant-technical-guide.md) and expanded V1 into ordered subtask gates: concrete Go ports/SQL/API/SSE/UI contracts, isolated fixtures and operational checks. Code inspection found `JarRepository.ListMonth` writes month config; the advisor runtime now uses a read-only jar path. The later implementation checkpoint below is the current proof source.
- 2026-09-22 Finance Assistant implementation has started from the approved plan: semantic query foundation, advisor conversation/run and user API-key schema, read-only tool registry/provider loop, JWT/API-key endpoints, best-effort Redis advisor audit and `/assistant` UI are present locally. Replayable SSE, durable audit delivery/alerting, configured-provider proof, public deployment and merchant normalization remain release gaps; V2 confirmed actions and later intelligence remain roadmap.
- Feedback → Fix → Changelog is wired locally: persistence, service-token middleware/audit, owner-scoped HTTP API, JWT/API-key user auth (`feedback:read`/`feedback:write`), lifecycle locking, Swagger and Account → Phản hồi UI are implemented and verified by backend/frontend suites. The latest local agent lifecycle and Redis audit probe passed; owner visual sign-off and production deployment remain pending. The service token remains separate from user API keys and has no finance access. See [Feedback API](architecture/FEEDBACK_API.md) and [validation evidence](work/VALIDATION_MATRIX.md).
- 2026-09-22 transfer implementation checkpoint: the Stage v1 baseline includes nullable `transactions.transfer_id`; `POST /api/v1/transactions/transfer` atomically creates paired source-expense/destination-income rows with transfer categories and report exclusion. Quick Add has a shared-base transfer mode. The 2026-10-05 local harness now verifies pair update/delete with idempotency keys; owner visual UAT and external deployment proof remain pending.
- 2026-09-22 CORE-03 implementation checkpoint: the Stage v1 baseline includes account IANA timezone/date-only fields, jar/month APIs and screens, optional transaction jar selection, and live month totals/notes. Owner UAT remains pending; see [validation evidence](work/VALIDATION_MATRIX.md#core-03-timezone-jars-and-live-monthly-summary--2026-09-22). Monthly completion is derived, with no cron close/snapshot.
- 2026-09-22 AI usage policy follow-up: owner clarified there should be no user usage limit right now. The hardcoded AI-entry request cap was removed; idempotent replay still prevents repeating provider work for the same request key. Future provider pricing/cost consent/usage budgets remain draft [AI-ENTRY-03](work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md), and the current no-limit decision is captured in [AI-USAGE-01](work/tickets/AI-USAGE-01-DETAIL_DESIGN.md).
- Dedicated shared-code scheduler request is separately captured as high-risk draft [WORKER-01](work/tickets/WORKER-01-cron-service.md), with recurring transactions and month-end reporting/close as candidate jobs. No worker/runtime/schema implementation is approved; report artifact semantics and recurring edge rules need resolution first.
- 2026-09-21 implementation follow-up: [AI-ENTRY-01](work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) implements normal Add/manual vs hold/AI entry, API-backed proposals, editing/approve/reject and atomic replay-safe ledger writes. Migration 000010 applied; real DB HTTP/concurrency tests pass. The 2026-09-22 strict JSON Schema follow-up now passes a synthetic five-case live extraction benchmark; owner UAT remains pending.
- 2026-09-22 AI-ENTRY-02 follow-up: one-shot multipart `POST /api/v1/ai/entry/process`, read-only recovery and independent proposal APIs are implemented. Backend stores originals in environment-qualified private S3, OCRs via a short-lived signed URL before LLM, and sends OCR text only. The Stage v1 baseline owns attachment/OCR metadata, approval links and cleanup claims; approval links are atomic and authenticated download checks owner/link. Explicit cleanup command is implemented but not run against bucket contents; browser download UAT remains pending. Do not retry the user's prior PDF automatically.
- 2026-09-22 LLM resolution: JSON-object mode previously scored 0/15; strict OpenAI-compatible JSON Schema mode now passes 5/5 synthetic cases and 24/24 scored fields. Live S3 PUT/readback 83 ms, OCR 4.034 s, LLM p50 19.769 s/p95 25.655 s, OCR+receipt-case pipeline 29.689 s; transfer-safety passes. Provider calls use the existing standard `net/http` adapter, with no additional SDK. Saved wallet descriptions now enter the bounded, untrusted wallet-matching context (2 KiB each). No transactions were approved or written. Broader model evaluation and owner UAT remain pending. Future multi-provider selection and sourced price transparency are captured in draft [AI-ENTRY-03](work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md); no additional provider was enabled.
- Local test servers restarted with latest code: frontend `http://localhost:4173/` returns 200 and backend `http://localhost:8080/api/v1/health` returns 200. Keep them running for owner UAT.

- 2026-09-21: owner requested feasibility assessment and a reviewable file for the two AI flows. [DESIGN-09-AI](work/tickets/TICKET-09-DETAIL_DESIGN.md) is `in_review`, approval pending: text/OCR proposals, internal-transfer reconciliation, read-only financial Q&A, prompt/eval contracts and T0–T7 delivery plan. No AI/transfer/jar runtime implemented in this documentation task.

- 2026-09-20 latest: owner requested API-only screens. All mounted data screens now use real APIs, including new budget explicit-interval CRUD/progress. Savings catalog filtering enforced FE/BE; category names/icons resolved in goal history. API roundtrip through FE proxy passed with PostgreSQL, test fixtures cleaned. Migration 000009 applied. [API-SCREENS-01](work/tickets/API-SCREENS-01-DETAIL_DESIGN.md) and savings remain in review/partial product scope; recurring budgets, internal paired transfers, reports and notifications are not completed. Earlier mock-budget/counterpart-pending notes below are historical.

- Latest owner request (2026-09-13): finish wallet management and adding transactions, then verify wallet–transaction–category logic and UI. Core wallet and basic income/expense slices are in review with [wallet proof](work/tickets/TICKET-01-02-VERIFICATION.md), [ledger proof](work/tickets/TICKET-02-01-VERIFICATION.md), and a durable [page contract](design/pages/transactions/README.md).

- Status: The root `app/` Vite React Tailwind app renders the Financial Clarity preview and is connected to the dev Gin API; Google OAuth and development CORS are enabled for local testing. PostgreSQL uses the explicit Stage v1 baseline migration; API startup does not mutate schema.
- Active backlog: `docs/work/BACKLOG.md`
- Current queue focus: transfer owner UAT and CORE-03 reconciliation; WORKER-01 remains draft pending report and recurrence semantics. AI-ENTRY-01/02 UAT, AI-ENTRY-03 owner design review, UI-FORMS-03 visual review, transfer pair edit/delete/idempotency and credit ledger remain separate open work.
- Active phase: None.
- Active ticket: [CORE-03](work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) is in progress through docs reconciliation and owner UAT. [AI-ENTRY-02](work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md) and [UI-FORMS-03](work/tickets/UI-FORMS-03-DETAIL_DESIGN.md) remain in progress/review for their independent UAT.
- Verified bug: [BUG-001](work/bugs/BUG-001-s3-presigned-get-signature.md), fixed with AWS SDK Go v2; [ADR-007](decisions/ADR-007-aws-s3-presigning.md) records the choice.

## Current Focus

Latest-turn override: implementation/reconciliation of [AI-ADVISOR-01](work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md). Earlier implementation queues below remain open; the local advisor slice is not a public deployment claim.

Owner requested continuing the approved timezone/jars/month goal. [CORE-03](work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) now has real API-backed screens and automated verification; its schema is included in the Stage v1 baseline. Next is owner UAT and reconciliation of remaining acceptance gaps. Monthly reports are live/recalculable, notes are independent, and automatic completion does not freeze ledger data. The new request for a cron worker is tracked separately in [WORKER-01](work/tickets/WORKER-01-cron-service.md); no implementation starts until month-end artifact semantics and recurrence rules are reviewed.

The independent UI-FORMS-03/AI flows still use the base-first contracts: `CategorySelectionList` is removed and transaction/budget/savings use `CategoryTreeSelector`; Money Lover guides AI input/result layout while MyPocket retains review/approve. Their visual/provider/browser UAT remains separate. AI entry is one-shot, stores no conversation, OCRs files before LLM, retains private S3 evidence and links it on approval.

## Recently Touched Areas

- `app/src/services/{accountTime,months,jars}.ts`, timezone context, jar/month API-backed screens and transaction/AI jar selectors; the Stage v1 baseline and owner-scoped calendar/jar/month repositories.
- `docs/work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md`, `WORKER-01-cron-service.md`, API/ERD/ADR-008, report/business rules, validation matrix, backlog and changelog.
- `app/src/atomic/atoms/BaseCategoryTree.tsx`, `molecules/CategoryTreeSelector.tsx`, AI composer/result-card bases, transaction/budget form composition and browser fixture tests.
- `docs/design/system/BASE_COMPONENTS.md`, category-tree contract, assistant/transaction/budget screen specs, UI-FORMS-03.

- `backend/internal/controller/http/transaction_handler.go`, wallet ledger balance repository/model, handler tests and generated Swagger; `backend/internal/infrastructure/ai/` now enforces strict JSON Schema output and includes bounded wallet descriptions for model disambiguation.
- `app/src/services/transactions.ts`, transaction/wallet rule tests, real Overview/Header/Transactions/Quick Add/Wallet consumers.
- `docs/design/pages/transactions/`, wallet/transaction verification, API/backlog/validation/changelog reconciliation.

- `docs/work/tickets/`: 11 parent tickets, 32 children and a business-oriented index; newly captured provider-choice draft remains pending owner review.
- `docs/requirements/SPEC.md`, `BUSINESS_RULES.md`, `REPORTS.md`, `REQUIREMENTS.md`, `USER_STORIES.md`
- `docs/work/VALIDATION_MATRIX.md`, `docs/work/ROADMAP.md`, `docs/releases/CHANGELOG.md`
- Standards relocated unchanged from `docs/templates/standards/` to the mandatory `docs/standards/` path; unused Harness CLI phase example removed.
- `AGENTS.md`
- `docs/`
- `docs/templates/`
- `docs/standards/`
- `design/`
- `design/system/DESIGN.md`
- `app/`
- `app/src/atomic/`

## Recent Decisions

- CORE-03 month completion is account-local and derived; monthly figures stay recalculable and user notes remain independent. A cron close/snapshot would change this contract, so WORKER-01 is intake only until the owner chooses report artifact semantics.
- Basic income/expense entries may use `basic` and `goal` wallets. `credit` is intentionally rejected until CRD-01…04 purchase/refund/payment semantics have an approved ledger; applying ordinary expense direction would corrupt debt meaning.
- Category selection obeys both transaction kind and applicable-wallet scope at UI and API. Empty `wallet_ids` means all owner wallets.
- Wallet API list exposes ledger-derived `current_balance`; `opening_balance` remains unchanged by transaction CRUD. Header total uses current balances only for `is_in_total` wallets.
- Global add and transaction edit/delete now use real APIs. Header, Overview, Transactions and Wallet management share refresh propagation; browser UAT found and fixed two stale-balance branches.

- `docs/design` now retains Markdown specifications only; 7 PNG/8 HTML exports were extracted and removed by owner request, recoverable at Git commit `1bc013d`. Shared behavior and per-component constraints are in `docs/design/system/` and component READMEs. Write screen-specific behavior as each screen is implemented.
- UI primitives use the imported `app/src/ui/theme.css`; component variants use `ui/variants.ts`, with atom tokens only re-exporting. Build checks base ownership, literal colors and common visual overrides. See ADR-002. CTA uses brand, action uses emerald; current default card/control radius remains 12px.
- Shared text, controls, cards, progress/gauge, feedback, navigation and chart compositions were normalized. Keyboard segmented selection and bottom-sheet focus/close/scroll behavior are shared. Prototype data and unmounted incomplete keypad/report flows remain explicitly partial.

- Owner chốt tên bảng tài khoản là `user`. ERD dùng tên đích này; GORM mapping và rename migration chưa triển khai, phải giữ dữ liệu/ID tài khoản hiện có.

- Wallet review: xoá thực sau xác nhận đã chốt, cho phép trùng tên. Money Lover adjustment tạo giao dịch mới bằng chênh lệch, mặc định loại khỏi báo cáo; thiết kế ledger MyPocket cần review theo kết quả này.
- Nhóm mặc định theo catalog owner chốt ngày 2026-09-13; migration `000004` là nguồn runtime cho tên, system key và quan hệ cha–con.
- Schema startup now renames legacy `app_users` to `user`, migrates wallets/categories/category-wallet assignments, and idempotently seeds stable system groups.
- The two category-tree exports are consolidated under `docs/design/molecules/category-tree/`; `BaseCategoryTree` is the canonical atom-composed molecule, with `nested` and `line` named variants.
- Category APIs return `wallet_ids`; create/update validates every selected wallet belongs to the authenticated owner. Swagger is regenerated from handler annotations.
- Unified Gin responses with `{data, meta}`, RFC 9457-style errors, request IDs, and shared constants/helpers; profile route is `/api/v1/auth/profile`.
- Database schema and system category data are now owned by the single Stage v1 SQL baseline in `backend/migrations/000001_stage_v1.up.sql`, executed with `go run ./cmd/migrate up` in dev and prod. `AutoMigrate`/startup system seed were removed; a fresh disposable Stage v1 database applies at version 1 clean. Wallet deletion cascades its transaction rows by the approved permanent-delete policy.
- Gin route registration is split by public-auth, account, category, wallet and transaction domain. Swagger is code-first: annotations are beside Go handlers and `go generate ./cmd/api` regenerates only implemented endpoints.
- Group management uses owner scoping, valid kind/parent rules, max two levels, child re-parenting on personal-parent deletion and owner-wallet validation. System metadata is read-only, but an owner may choose its applicable wallets through a dedicated endpoint. `/account/groups` has real list/create/edit/delete states, not mock data.
- Added `/account/groups` UI using the real categories/wallets APIs and reusable bases (`BaseCategoryTree`, `Heading`, `FormField`, `BaseCheckbox`, `SurfaceCard`, `IconBadge`, `BaseBottomSheet`). The canonical tree follows `docs/design/molecules/category-tree/nested`; personal-category edit exposes confirmed deletion. Account links to wallets/groups are visible. Reports and quick-add are retained source but deliberately not mounted until their APIs exist.
- TICKET-01-03 (group management) is complete; CRUD group remains separately scoped from wallet management.
- The bottom-navigation create button is vertically aligned inside the shared navigation frame; its center slot remains intentionally unlabeled.
- Vite dev proxy now forwards `/api` to Gin, allowing FE and BE to be tested through one browser origin.
- Base-cards export tại `docs/design/atoms/base-cards/` là nguồn tham chiếu mới; Overview đã dùng `SurfaceCard`, `IconBadge`, `WalletCard` và `TransactionItem` để khớp các biến thể card/list, không lặp markup card tại màn hình.
- Variant maps của surface, badge, button, wallet và transaction được tập trung trong `app/src/atomic/atoms/tokens.ts`; consumer chỉ tham chiếu token global.
- `BaseButton` owns inline-flex icon-plus-label alignment; screen callers may use `gap` but must not recreate this layout locally.
- Base card/form/row corners are reduced one token step; circular icon, progress and pill CTA geometry remain semantic exceptions. In `BaseCategoryTree`, the trunk begins at the first child row rather than crossing the parent row.

- Divide the full business contract into parent/child tickets with concise scope and acceptance criteria. Keep unresolved decisions in the affected child; do not silently resolve them during breakdown.
- Current product contract is `docs/requirements/SPEC.md`, with `BUSINESS_RULES.md` and `REPORTS.md`; older design inputs do not override these decisions.
- Budget and Jar are separate. Jars are optional expense-group tracking with monthly copied configuration, advisory allocation/warnings, no balance carryover, and a cumulative reporting view.
- Monthly reports remain recalculable after month-end. User notes are independent; related context and AI summaries are generated separately and cannot overwrite manual notes.
- Timestamps use UTC; account timezone drives query boundaries; calendar-only dates/months keep their explicit labels. Only VND is seeded; multi-currency has no designed contract.
- Travel Mode automatically links eligible new transactions; recurring-generated transactions do not inherit it. Recurring creates ordinary editable transactions with a default note.
- Native recurring numeric drafts now preserve valid thousands grouping while rejecting malformed characters/separators before create/update; see [NATIVE-RECURRING-NUMERIC-STRICTNESS](work/bugs/NATIVE-RECURRING-NUMERIC-STRICTNESS.md).
- The following older bullets describe the existing preview/history, not authority over the current product specification.
- Keep `AGENTS.md` at the repository root for agent discovery.
- Keep shared state in `docs/CONTEXT.md`.
- Use markdown-only enforcement for v1.
- Use `docs/work/phases/` for multi-ticket work.
- Use pay-as-you-go documentation for brownfield projects.
- Treat `design/system/DESIGN.md` and `design/INDEX.md` as the current design/product inputs for the next app-start phase.
- `design/system/DESIGN.md` now defines Money Lover-style app parity, personal extension backlog, component inventory, data domains, launch phases, and acceptance checklist.
- Design artifact folders under `design/` are exported source references; base implementation follows the component inventory in `design/INDEX.md`.
- Frontend preview uses the old app as logic/data-shape reference only; UI is new and follows `docs/design/system/DESIGN.md`.
- New app source lives in root `app/`; runtime behavior and API contracts are documented in this repository.
- The standalone FE preview now includes Overview, Transactions, Budgets, Reports, Account, Quick Add sheet, goals/funds, quick personal actions, and a nested category report mock.
- OCR Platform docs have been captured in `docs/architecture/OCR_API.md` for future receipt OCR implementation against `https://ocr.dungxbuif.com/`.

## Next Steps

Native Travel Mode parity was restored on 2026-10-06: active context/date
scope, safe deletion confirmation, and the ordinary-income/expense-only
auto-link warning are now covered by the shared SwiftUI components. Native
XCTest 54/54, Release warnings-as-errors, base/project guards, and the signed
authenticated More-route sweep 1/1 in 122.031s pass for this correction.

Native recurring parity was then restored: account-timezone scheduling copy,
next-run/wallet/note/type details, and safe delete confirmation now match the
web contract. Native XCTest passes 55/55, the Release warnings-as-errors
build and guards pass, and the signed More-route sweep passes 1/1 in 124.207s.

Native Reports now also loads the optional month-summary breakdown, showing
category/type/count rows and the ledger activity exclusions beside Insider
metrics. Native XCTest remains 55/55, Release/guards pass, and the signed
More-route sweep passes 1/1 in 124.287s.

Native credit statement parity was then restored: current debt is shown as a
positive amount, payment status and credit entry kinds are localized, due-date
context is visible when configured, and row notes take precedence over dates.
The existing credit accounting/API contract is unchanged. Native XCTest
remains 55/55, the warnings-as-errors simulator build passes, and the signed
More-route sweep passes 1/1 in 125.111s.

Native portfolio detail parity was then restored: each asset now shows weighted
average cost, realized/unrealized P/L, current value or explicit unknown
pricing, and newest-first account-timezone buy/sell history with quantity,
price, fee and note context. Native XCTest passes 56/56, Release/guards pass,
and the signed More-route sweep passes 1/1 in 121.948s.

The complete signed `LOCAL_API_UI` simulator regression was rerun after the
credit/portfolio batch: 14/14 passed in 293.119s, including wallet, transfer,
credit entry/payment, Assistant, AI-entry capability, feedback and secondary
route coverage. Credit assertions now intentionally verify localized display
labels rather than raw API enum strings.

The following enum-label parity pass centralized the remaining iOS display
labels for wallet/category/credit/feedback/Assistant statuses. Native XCTest
passes 57/57, Release/guards pass, and the signed credit-wallet plus
feedback/more targeted UI checks pass 1/1 each. Wire enum values are unchanged.

- Review the [dedicated advisor design](work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md) and [V1 task sequence](superpowers/plans/2026-09-22-finance-assistant-v1.md). Feedback's browser round-trip/owner UAT and deployment checks remain separate. Finance Assistant has a local JWT/API-key synchronous JSON vertical slice with best-effort Redis access audit; next gates are typed fact cards, SSE/recovery, durable audit delivery/alerting and configured-provider browser proof.
- Finance Assistant contract/lifecycle fixes are now locally verified: typed schemas, OpenAI-compatible tool history, bounded final-answer loop, credential revalidation, cancelled-run commit guard, and wallet/budget/goal/jar cards. Remaining gates are configured-provider browser proof, SSE/reconnect, durable audit delivery/alerting and owner UAT.
- Finance Assistant cancellation now interrupts provider work handled by the same backend process; the database lease guard remains the source of truth. Do not describe cross-process immediate cancellation as shipped until durable pub/sub or worker ownership is implemented.
- Finish CORE-03 owner UAT at the local frontend (`/account` timezone setting, `/jars`, current/past `/months/YYYY-MM`, transaction and AI proposal jar assignment). Confirm data behavior visually and with an account the owner is comfortable testing; automated evidence is in [VALIDATION_MATRIX](work/VALIDATION_MATRIX.md).
- Clarify [WORKER-01](work/tickets/WORKER-01-cron-service.md): whether “chốt sổ/tạo report” means a persisted month-end artifact or only a scheduled trigger/cache over recalculable figures, plus recurring catch-up/idempotency/timezone behavior. Then prepare detail design for review before code/runtime changes.
- Keep AI-ENTRY-01/02 live/provider/attachment UAT and UI-FORMS-03 visual review as separate backlog work; broader AI model evaluation remains required beyond the five synthetic cases.
- Preserve base-first screen rules and run `npm run check:design`, `npm run test:design`, and `npm run build` for frontend changes. API/DB/runtime changes require PostgreSQL tests and migration evidence.

## Open Questions

- WORKER-01 month-end semantics: CORE-03 says live/recalculable report and derived completion; the new request names a cron job that closes/generates a report. Decide whether any stored artifact is authoritative or merely rebuildable/cache output. Recurring month-end/catch-up/duplicate prevention/timezone changes also need rules.
- Deletion of linked objects and paired-wallet effects; Travel Mode for backdated/offline/delayed confirmation.
- Exact income/refund classification for jar allocation; credit overpayment/refund/statement allocation; portfolio funding and historical deletion behavior.
