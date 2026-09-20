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
updated: 2026-09-13
---

# Project Context

## Field Ownership

- Human owns project intent, priority overrides, and unresolved product questions.
- AI owns concise state refreshes after work: touched areas, recent decisions, next steps, and queue summary.
- Shared fields can be updated by either human or AI, but AI must not silently override human priority.

## Current Status

- 2026-09-20 latest: owner requested API-only screens. All mounted data screens now use real APIs, including new budget explicit-interval CRUD/progress. Savings catalog filtering enforced FE/BE; category names/icons resolved in goal history. API roundtrip through FE proxy passed with PostgreSQL, test fixtures cleaned. Migration 000009 applied. [API-SCREENS-01](work/tickets/API-SCREENS-01-DETAIL_DESIGN.md) and savings remain in review/partial product scope; recurring budgets, internal paired transfers, reports and notifications are not completed. Earlier mock-budget/counterpart-pending notes below are historical.

- Latest owner request (2026-09-13): finish wallet management and adding transactions, then verify wallet–transaction–category logic and UI. Core wallet and basic income/expense slices are in review with [wallet proof](work/tickets/TICKET-01-02-VERIFICATION.md), [ledger proof](work/tickets/TICKET-02-01-VERIFICATION.md), and a durable [screen contract](design/screens/transactions/README.md).

- Status: The root `app/` Vite React Tailwind app renders the Financial Clarity preview and is connected to the dev Gin API; Google OAuth and development CORS are enabled for local testing. PostgreSQL uses explicit versioned migrations; API startup does not mutate schema.
- Active backlog: `docs/work/BACKLOG.md`
- Current queue focus: wallet CRUD and basic income/expense ledger are implemented and verified for review. Account → Nhóm remains closed; receipt/jar and credit/adjustment ledgers remain separate follow-ups.
- Active phase: None.
- Active ticket: TICKET-01-02 and the approved basic slice of TICKET-02-01 are `in_review`.
- Active bug: None.

## Current Focus

Yêu cầu hiện tại: cập nhật docs cho Tài khoản → Quản lý nhóm → chọn nhóm → Sửa, với nhóm cha, “category” và ví áp dụng. Nguồn UI: `docs/design/system/DESIGN.md#account--quản-lý-nhóm`. Ý nghĩa “category” còn chờ trả lời; cập nhật docs không tự phê duyệt API/schema.

Review the MyPocket product/business specification and its BA ticket breakdown. The owner now requests large tickets with smaller child tickets, written briefly in business language. This supersedes the earlier request to avoid creating tickets. No implementation phase or technical plan is scheduled by this breakdown.

## Recently Touched Areas

- `backend/internal/controller/http/transaction_handler.go`, wallet ledger balance repository/model, handler tests and generated Swagger.
- `app/src/services/transactions.ts`, transaction/wallet rule tests, real Overview/Header/Transactions/Quick Add/Wallet consumers.
- `docs/design/screens/transactions/`, wallet/transaction verification, API/backlog/validation/changelog reconciliation.

- `docs/work/tickets/`: 11 parent tickets, 31 children and a business-oriented index; all tickets are `draft` pending review.
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

- Owner visual acceptance of API-only budget and savings screens; complete shared paired transfers as separately designed work. Browser inspected budget empty/editor and grouped transaction form without writing to owner's account. Automated API persistence test covers savings and budgets using isolated created fixtures. Do not confuse this with complete browser CRUD UAT or full budget roadmap acceptance.

- Owner resolved savings counterpart question: money may come from outside app; ordinary savings entries affect one wallet. Shared transfer-to-another-wallet is separate. Official Money Lover transfer guide checked and linked in savings design. Continue specialized savings categories without requiring counterpart wallet; internal-transfer ledger remains separate unfinished work.

- Latest execution: wallet list rebuilt as reference-based WalletSelectionList; create uses BaseSelect. Goal date API and SavingsSummary/SavingsWalletPanel implemented; selecting goal opens real progress/history. QuickAdd accepts wallet context and shared searchable category list. [TICKET-06-01 design/proof](work/tickets/TICKET-06-01-DETAIL_DESIGN.md). Automated tests/build pass; browser creation draft inspected/canceled. Pending: answer whether savings transfers require counterpart wallet, specialized categories, exact transaction-form fidelity, full persistence UAT. Notifications/reports remain unsupported. Prior “select/date missing” notes are historical.

- Latest: [wallet screen specification](design/screens/wallets/README.md) normalized using owner decisions and official Money Lover docs. BaseSelect during creation only; type immutable after save. Runtime picker replacement is pending. Per-type goal/credit detail screens and date/statement fields remain explicitly incomplete.

- Latest owner direction: normalize both newly supplied wallet references, tighten base enforcement, reimplement Add Wallet and wallet-type selection. [UI-WALLET-02](work/tickets/UI-WALLET-02-DETAIL_DESIGN.md) implementation and automated checks pass; browser/owner visual review pending. Existing-wallet selector reference is normalized but its runtime scope/filter flow remains unimplemented. This UI work takes priority over transfer/adjustment.

- Owner follow-up: removed the note field from wallet creation only; editing existing notes remains available. UI review pending.

- UI-EMPTY-01 follow-up: owner spotted the remaining “Chưa có ví” border. Both Overview and wallet management now opt into the same plain status variant; pending visual review.

- 2026-09-20: owner reports previously tested bugs are OK and requests continued implementation. Record this as owner-reported acceptance of the tested fixes, not blanket completion of remaining product scope. Transaction empty-state border correction is tracked in [UI-EMPTY-01](work/tickets/UI-EMPTY-01.md). Next product work remains transfer/adjustment detail design before implementation.
- Local OAuth launch: load Google credentials from the existing reference app environment without printing secrets; use `GOOGLE_REDIRECT_URL=http://localhost:4173/api/v1/auth/google/callback` through the frontend `/api` proxy. This session verified HTTP 302 with that redirect URI; Google Console must register the identical URI. No credentials were copied into tracked files.

- Review the verified core wallet and basic ledger slices. The next independent product choices are receipt/OCR and jar assignment for TICKET-02-01, or transfer/adjustment and credit statement/payment semantics; none are represented by mocks in the implemented ledger.

- For every next UI slice, read `docs/design/README.md`, list reused bases, create missing base contracts/components first, and record the screen's behavior/proof. Run `npm run check:design`, `npm run test:design` and `npm run build` in `app/`.

- Review the parent/child ticket list and resolve business questions in the affected children; prioritize implementation only when requested.
- Continue implementation from the Financial Clarity design and reference-app behavior; local runtime proof for auth/profile/home is now available, while product feature validation remains pending.
- When implementing receipt OCR, use `docs/architecture/OCR_API.md` as the provider contract and keep `OCR_API_KEY` server-side only.
- Product UAT and runtime proof remain pending; documentation review must not be presented as implemented behavior.

## Open Questions

- Deletion of linked objects and paired-wallet effects; Travel Mode for backdated/offline/delayed confirmation; recurring month-end/catch-up/timezone changes.
- Exact income/refund classification for jar allocation; credit overpayment/refund/statement allocation; portfolio funding and historical deletion behavior.
