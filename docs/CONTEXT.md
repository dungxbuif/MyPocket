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
updated: 2026-09-22
---

# Project Context

## Field Ownership

- Human owns project intent, priority overrides, and unresolved product questions.
- AI owns concise state refreshes after work: touched areas, recent decisions, next steps, and queue summary.
- Shared fields can be updated by either human or AI, but AI must not silently override human priority.

## Current Status

- 2026-09-22 CORE-03 implementation checkpoint: account IANA timezone/date-only migrations, jar/month APIs and screens, optional transaction jar selection, live month totals/note are implemented through migrations 000013–000015. Local DB is version 15 clean; backend PostgreSQL suite and frontend design/calendar/jar/build checks pass. Owner UAT remains pending; see [validation evidence](work/VALIDATION_MATRIX.md#core-03-timezone-jars-and-live-monthly-summary--2026-09-22). Monthly completion is derived, with no cron close/snapshot.
- Dedicated shared-code scheduler request is separately captured as high-risk draft [WORKER-01](work/tickets/WORKER-01-cron-service.md), with recurring transactions and month-end reporting/close as candidate jobs. No worker/runtime/schema implementation is approved; report artifact semantics and recurring edge rules need resolution first.
- 2026-09-21 implementation follow-up: [AI-ENTRY-01](work/tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) implements normal Add/manual vs hold/AI entry, API-backed proposals, editing/approve/reject and atomic replay-safe ledger writes. Migration 000010 applied; real DB HTTP/concurrency tests pass. The 2026-09-22 strict JSON Schema follow-up now passes a synthetic five-case live extraction benchmark; owner UAT remains pending.
- 2026-09-22 AI-ENTRY-02 follow-up: one-shot multipart `POST /api/v1/ai/entry/process`, read-only recovery and independent proposal APIs are implemented. Backend stores originals in environment-qualified private S3, OCRs via a short-lived signed URL before LLM, and sends OCR text only. Migrations 000011–000012 add attachment/OCR metadata, approval links and cleanup claims; approval links are atomic and authenticated download checks owner/link. Local dev DB is migration 12; real PostgreSQL tests pass. Explicit cleanup command is implemented but not run against bucket contents; browser download UAT remains pending. Do not retry the user's prior PDF automatically.
- 2026-09-22 LLM resolution: JSON-object mode previously scored 0/15; strict OpenAI-compatible JSON Schema mode now passes 5/5 synthetic cases and 24/24 scored fields. Live S3 PUT/readback 83 ms, OCR 4.034 s, LLM p50 19.769 s/p95 25.655 s, OCR+receipt-case pipeline 29.689 s; transfer-safety passes. Provider calls use the existing standard `net/http` adapter, with no additional SDK. Saved wallet descriptions now enter the bounded, untrusted wallet-matching context (2 KiB each). No transactions were approved or written. Broader model evaluation and owner UAT remain pending. Future multi-provider selection and sourced price transparency are captured in draft [AI-ENTRY-03](work/tickets/AI-ENTRY-03-PROVIDER-SELECTION.md); no additional provider was enabled.
- Local test servers restarted with latest code: frontend `http://localhost:4173/` returns 200 and backend `http://localhost:8080/api/v1/health` returns 200. Keep them running for owner UAT.

- 2026-09-21: owner requested feasibility assessment and a reviewable file for the two AI flows. [DESIGN-09-AI](work/tickets/TICKET-09-DETAIL_DESIGN.md) is `in_review`, approval pending: text/OCR proposals, internal-transfer reconciliation, read-only financial Q&A, prompt/eval contracts and T0–T7 delivery plan. No AI/transfer/jar runtime implemented in this documentation task.

- 2026-09-20 latest: owner requested API-only screens. All mounted data screens now use real APIs, including new budget explicit-interval CRUD/progress. Savings catalog filtering enforced FE/BE; category names/icons resolved in goal history. API roundtrip through FE proxy passed with PostgreSQL, test fixtures cleaned. Migration 000009 applied. [API-SCREENS-01](work/tickets/API-SCREENS-01-DETAIL_DESIGN.md) and savings remain in review/partial product scope; recurring budgets, internal paired transfers, reports and notifications are not completed. Earlier mock-budget/counterpart-pending notes below are historical.

- Latest owner request (2026-09-13): finish wallet management and adding transactions, then verify wallet–transaction–category logic and UI. Core wallet and basic income/expense slices are in review with [wallet proof](work/tickets/TICKET-01-02-VERIFICATION.md), [ledger proof](work/tickets/TICKET-02-01-VERIFICATION.md), and a durable [screen contract](design/screens/transactions/README.md).

- Status: The root `app/` Vite React Tailwind app renders the Financial Clarity preview and is connected to the dev Gin API; Google OAuth and development CORS are enabled for local testing. PostgreSQL uses explicit versioned migrations; API startup does not mutate schema.
- Active backlog: `docs/work/BACKLOG.md`
- Current queue focus: finish CORE-03 reconciliation and owner UAT; WORKER-01 remains draft pending report and recurrence semantics. AI-ENTRY-01/02 UAT, AI-ENTRY-03 owner design review, UI-FORMS-03 visual review, transfer/adjustment and credit ledger remain separate open work.
- Active phase: None.
- Active ticket: [CORE-03](work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) is in progress through docs reconciliation and owner UAT. [AI-ENTRY-02](work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md) and [UI-FORMS-03](work/tickets/UI-FORMS-03-DETAIL_DESIGN.md) remain in progress/review for their independent UAT.
- Verified bug: [BUG-001](work/bugs/BUG-001-s3-presigned-get-signature.md), fixed with AWS SDK Go v2; [ADR-007](decisions/ADR-007-aws-s3-presigning.md) records the choice.

## Current Focus

Owner requested continuing the approved timezone/jars/month goal. [CORE-03](work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) now has real API-backed screens, migrations 13–15 and automated verification; next is owner UAT and reconciliation of remaining acceptance gaps. Monthly reports are live/recalculable, notes are independent, and automatic completion does not freeze ledger data. The new request for a cron worker is tracked separately in [WORKER-01](work/tickets/WORKER-01-cron-service.md); no implementation starts until month-end artifact semantics and recurrence rules are reviewed.

The independent UI-FORMS-03/AI flows still use the base-first contracts: `CategorySelectionList` is removed and transaction/budget/savings use `CategoryTreeSelector`; Money Lover guides AI input/result layout while MyPocket retains review/approve. Their visual/provider/browser UAT remains separate. AI entry is one-shot, stores no conversation, OCRs files before LLM, retains private S3 evidence and links it on approval.

## Recently Touched Areas

- `app/src/services/{accountTime,months,jars}.ts`, timezone context, jar/month API-backed screens and transaction/AI jar selectors; backend migrations 000013–000015 and owner-scoped calendar/jar/month repositories.
- `docs/work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md`, `WORKER-01-cron-service.md`, API/ERD/ADR-008, report/business rules, validation matrix, backlog and changelog.
- `app/src/atomic/atoms/BaseCategoryTree.tsx`, `molecules/CategoryTreeSelector.tsx`, AI composer/result-card bases, transaction/budget form composition and browser fixture tests.
- `docs/design/system/BASE_COMPONENTS.md`, category-tree contract, assistant/transaction/budget screen specs, UI-FORMS-03.

- `backend/internal/controller/http/transaction_handler.go`, wallet ledger balance repository/model, handler tests and generated Swagger; `backend/internal/infrastructure/ai/` now enforces strict JSON Schema output and includes bounded wallet descriptions for model disambiguation.
- `app/src/services/transactions.ts`, transaction/wallet rule tests, real Overview/Header/Transactions/Quick Add/Wallet consumers.
- `docs/design/screens/transactions/`, wallet/transaction verification, API/backlog/validation/changelog reconciliation.

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
- `refereces/disappointed_app/`

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
- Database schema and system category data are now owned by ordered SQL migrations in `backend/migrations/`, executed with `go run ./cmd/migrate up` in dev and prod. `AutoMigrate`/startup system seed were removed; dev DB reached migration version 8, clean. Migration `000008` assigns the system icon catalog for deployment. Wallet deletion cascades its transaction rows by the approved permanent-delete policy.
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
- The following older bullets describe the existing preview/history, not authority over the current product specification.
- Keep `AGENTS.md` at the repository root for agent discovery.
- Keep shared state in `docs/CONTEXT.md`.
- Use markdown-only enforcement for v1.
- Use `docs/work/phases/` for multi-ticket work.
- Use pay-as-you-go documentation for brownfield projects.
- Treat `design/system/DESIGN.md` and `design/INDEX.md` as the current design/product inputs for the next app-start phase.
- Treat `refereces/disappointed_app/` as the existing reference implementation unless the human decides to move or rename it.
- `design/system/DESIGN.md` now defines Money Lover-style app parity, personal extension backlog, component inventory, data domains, launch phases, and acceptance checklist.
- Design artifact folders under `design/` are exported source references; base implementation follows the component inventory in `design/INDEX.md`.
- Frontend preview uses the old app as logic/data-shape reference only; UI is new and follows `docs/design/system/DESIGN.md`.
- New app source lives in root `app/`; `refereces/disappointed_app/` should remain the reference app for moving logic/API contracts later.
- The standalone FE preview now includes Overview, Transactions, Budgets, Reports, Account, Quick Add sheet, goals/funds, quick personal actions, and a nested category report mock.
- OCR Platform docs have been captured in `docs/architecture/OCR_API.md` for future receipt OCR implementation against `https://ocr.dungxbuif.com/`.

## Next Steps

- Finish CORE-03 owner UAT at the local frontend (`/account` timezone setting, `/jars`, current/past `/months/YYYY-MM`, transaction and AI proposal jar assignment). Confirm data behavior visually and with an account the owner is comfortable testing; automated evidence is in [VALIDATION_MATRIX](work/VALIDATION_MATRIX.md).
- Clarify [WORKER-01](work/tickets/WORKER-01-cron-service.md): whether “chốt sổ/tạo report” means a persisted month-end artifact or only a scheduled trigger/cache over recalculable figures, plus recurring catch-up/idempotency/timezone behavior. Then prepare detail design for review before code/runtime changes.
- Keep AI-ENTRY-01/02 live/provider/attachment UAT and UI-FORMS-03 visual review as separate backlog work; broader AI model evaluation remains required beyond the five synthetic cases.
- Preserve base-first screen rules and run `npm run check:design`, `npm run test:design`, and `npm run build` for frontend changes. API/DB/runtime changes require PostgreSQL tests and migration evidence.

## Open Questions

- WORKER-01 month-end semantics: CORE-03 says live/recalculable report and derived completion; the new request names a cron job that closes/generates a report. Decide whether any stored artifact is authoritative or merely rebuildable/cache output. Recurring month-end/catch-up/duplicate prevention/timezone changes also need rules.
- Deletion of linked objects and paired-wallet effects; Travel Mode for backdated/offline/delayed confirmation.
- Exact income/refund classification for jar allocation; credit overpayment/refund/statement allocation; portfolio funding and historical deletion behavior.
