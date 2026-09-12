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

- Status: A new root `app/` Vite React Tailwind app now renders a mock-only MyPocket FE preview following the new Financial Clarity design and Atomic Design structure.
- Active backlog: `docs/work/BACKLOG.md`
- Current queue focus: Review 11 parent and 31 child business tickets in `docs/work/tickets/README.md`, together with unresolved business decisions. The existing `app/` FE preview remains mock-only.
- Active phase: None.
- Active ticket: None.
- Active bug: None.

## Current Focus

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
- `design/DESIGN.md`
- `app/`
- `app/src/atomic/`
- `refereces/disappointed_app/`

## Recent Decisions

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
- Treat `design/DESIGN.md` and `design/INDEX.md` as the current design/product inputs for the next app-start phase.
- Treat `refereces/disappointed_app/` as the existing reference implementation unless the human decides to move or rename it.
- `design/DESIGN.md` now defines Money Lover-style app parity, personal extension backlog, component inventory, data domains, launch phases, and acceptance checklist.
- Design artifact folders under `design/` now use normalized ASCII names such as `component-amount-keypad`, `component-category-tree-line`, and `system-base-cards`.
- Frontend preview uses the old app as logic/data-shape reference only; UI is new and follows `docs/design/DESIGN.md`.
- New app source lives in root `app/`; `refereces/disappointed_app/` should remain the reference app for moving logic/API contracts later.
- The standalone FE preview now includes Overview, Transactions, Budgets, Reports, Account, Quick Add sheet, goals/funds, quick personal actions, and a nested category report mock.
- OCR Platform docs have been captured in `docs/architecture/OCR_API.md` for future receipt OCR implementation against `https://ocr.dungxbuif.com/`.

## Next Steps

- Review the parent/child ticket list and resolve business questions in the affected children; prioritize implementation only when requested.
- Reconcile the existing mock preview and reference architecture against the accepted contract when implementation is requested; no runtime changes have been made in this documentation update.
- When implementing receipt OCR, use `docs/architecture/OCR_API.md` as the provider contract and keep `OCR_API_KEY` server-side only.
- Product UAT and runtime proof remain pending; documentation review must not be presented as implemented behavior.

## Open Questions

- Deletion of linked objects and paired-wallet effects; Travel Mode for backdated/offline/delayed confirmation; recurring month-end/catch-up/timezone changes.
- Exact income/refund classification for jar allocation; credit overpayment/refund/statement allocation; portfolio funding and historical deletion behavior.
