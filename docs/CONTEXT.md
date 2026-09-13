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

- Status: The root `app/` Vite React Tailwind app renders the Financial Clarity preview and is connected to the dev Gin API; Google OAuth and development CORS are enabled for local testing.
- Active backlog: `docs/work/BACKLOG.md`
- Current queue focus: first wallet/category persistence slice is implemented for runtime review; CRUD API/UI management remains the next slice.
- Active phase: None.
- Active ticket: TICKET-01-02 (wallet schema/seed slice); management CRUD is not yet implemented.
- Active bug: None.

## Current Focus

Yêu cầu hiện tại: cập nhật docs cho Tài khoản → Quản lý nhóm → chọn nhóm → Sửa, với nhóm cha, “category” và ví áp dụng. Nguồn UI: `docs/design/system/DESIGN.md#account--quản-lý-nhóm`. Ý nghĩa “category” còn chờ trả lời; cập nhật docs không tự phê duyệt API/schema.

Review the MyPocket product/business specification and its BA ticket breakdown. The owner now requests large tickets with smaller child tickets, written briefly in business language. This supersedes the earlier request to avoid creating tickets. No implementation phase or technical plan is scheduled by this breakdown.

## Recently Touched Areas

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

- Owner chốt tên bảng tài khoản là `user`. ERD dùng tên đích này; GORM mapping và rename migration chưa triển khai, phải giữ dữ liệu/ID tài khoản hiện có.

- Wallet review: xoá thực sau xác nhận đã chốt, cho phép trùng tên. Money Lover adjustment tạo giao dịch mới bằng chênh lệch, mặc định loại khỏi báo cáo; thiết kế ledger MyPocket cần review theo kết quả này.
- Nhóm mặc định lấy từ `refereces/disappointed_app/backend/migrations/0011_phase002_category_catalog.sql` và seed tiền nhiệm; giữ tên/system key/cha-con. Không thay bằng nhóm tự nghĩ hoặc mock FE.
- Schema startup now renames legacy `app_users` to `user`, migrates wallets/categories/category-wallet assignments, and idempotently seeds stable system groups.
- Base `CategoryTreeCard` styling was reconciled with the two category component references; account UI for backend-unimplemented features is commented out.
- Added protected `GET /api/v1/categories` returning system and owner-visible groups; Swagger regenerated.
- Unified Gin responses with `{data, meta}`, RFC 9457-style errors, request IDs, and shared constants/helpers; profile route is `/api/v1/auth/profile`.
- Added `/account/groups` UI using the real categories API, with loading/error/empty states and reusable `SurfaceCard`/`IconBadge` atoms. Account links to wallets/groups are visible; unsupported sections remain commented.
- TICKET-01-03 is now ready with an implementation plan at `docs/superpowers/plans/2026-09-13-group-management.md`; CRUD group is scoped separately from wallet management.
- Vite dev proxy now forwards `/api` to Gin, allowing FE and BE to be tested through one browser origin.
- Base-cards export tại `docs/design/atoms/base-cards/` là nguồn tham chiếu mới; Overview đã dùng `SurfaceCard`, `IconBadge`, `WalletCard` và `TransactionItem` để khớp các biến thể card/list, không lặp markup card tại màn hình.
- Variant maps của surface, badge, button, wallet và transaction được tập trung trong `app/src/atomic/atoms/tokens.ts`; consumer chỉ tham chiếu token global.

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

- Review runtime schema and seeded groups, then implement the protected wallet/group CRUD slice.

- Review the parent/child ticket list and resolve business questions in the affected children; prioritize implementation only when requested.
- Continue implementation from the Financial Clarity design and reference-app behavior; local runtime proof for auth/profile/home is now available, while product feature validation remains pending.
- When implementing receipt OCR, use `docs/architecture/OCR_API.md` as the provider contract and keep `OCR_API_KEY` server-side only.
- Product UAT and runtime proof remain pending; documentation review must not be presented as implemented behavior.

## Open Questions

- Deletion of linked objects and paired-wallet effects; Travel Mode for backdated/offline/delayed confirmation; recurring month-end/catch-up/timezone changes.
- Exact income/refund classification for jar allocation; credit overpayment/refund/statement allocation; portfolio funding and historical deletion behavior.
