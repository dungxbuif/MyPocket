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
updated: 2026-09-12
---

# Project Context

## Field Ownership

- Human owns project intent, priority overrides, and unresolved product questions.
- AI owns concise state refreshes after work: touched areas, recent decisions, next steps, and queue summary.
- Shared fields can be updated by either human or AI, but AI must not silently override human priority.

## Current Status

- Status: Financial Clarity design has been refactored into an app initialization spec that keeps Money Lover-style parity plus personal extensions.
- Active backlog: `docs/work/BACKLOG.md`
- Current queue focus: Start the first app implementation slice from `design/DESIGN.md` and the reference app.
- Active phase: None.
- Active ticket: None.
- Active bug: None.

## Current Focus

Prepare MyPocket / Financial Clarity for app implementation. The current product direction is a personal finance app inspired by Money Lover, with a reusable design system and fast transaction logging as the first implementation target.

## Recently Touched Areas

- `AGENTS.md`
- `docs/`
- `docs/templates/`
- `docs/standards/`
- `design/`
- `design/DESIGN.md`
- `refereces/disappointed_app/`

## Recent Decisions

- Keep `AGENTS.md` at the repository root for agent discovery.
- Keep shared state in `docs/CONTEXT.md`.
- Use markdown-only enforcement for v1.
- Use `docs/work/phases/` for multi-ticket work.
- Use pay-as-you-go documentation for brownfield projects.
- Treat `design/DESIGN.md` and `design/INDEX.md` as the current design/product inputs for the next app-start phase.
- Treat `refereces/disappointed_app/` as the existing reference implementation unless the human decides to move or rename it.
- `design/DESIGN.md` now defines Money Lover-style app parity, personal extension backlog, component inventory, data domains, launch phases, and acceptance checklist.
- Design artifact folders under `design/` now use normalized ASCII names such as `component-amount-keypad`, `component-category-tree-line`, and `system-base-cards`.

## Next Steps

- Create the first app-start phase or ticket from the design system and reference app.
- Decide whether the active app source should live at the repository root, under `refereces/disappointed_app/`, or under a renamed app directory.
- Reconcile product requirements and architecture docs once the app-start direction is chosen.

## Open Questions

- Should `refereces/disappointed_app/` become the active app source, or should implementation start in a new clean app directory?
- Should the first app slice prioritize UI foundation/component library or the fast transaction logging flow?
