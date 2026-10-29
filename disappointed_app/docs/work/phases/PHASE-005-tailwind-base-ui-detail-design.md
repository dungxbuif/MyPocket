---
artifact_type: detail_design
id: PHASE-005-TAILWIND-BASE-UI-DETAIL-DESIGN
status: approved
owner: shared
approval: approved_by_current_user_instruction
approved_on: 2026-08-31
trace:
  backlog_item: BL-005
  phase: PHASE-005
  tickets: [TICKET-015, TICKET-016, TICKET-017]
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
  master_docs_touched: []
---

# DETAIL DESIGN: Tailwind Base UI Migration

## Problem

The current PWA UI is driven by a large handwritten `frontend/src/styles.css`. Recent mobile UAT shows repeated spacing, typography, rounded-corner, scrollbar, and liquid-glass issues because core surfaces are not governed by reusable base components or consistent tokens.

## Scope

- Add Tailwind CSS to the Vite frontend using the official Vite plugin path.
- Establish app-level Tailwind tokens for MyPocket colors, compact mobile sizing, liquid-glass surfaces, safe-area layout, typography, and scrollbars.
- Create reusable frontend base components for common primitives: app surface, card, icon button, pill button, bottom nav, sheet, section title, list rows, segmented controls, and primary actions.
- Migrate the visible app shell and core screens through TICKET-017 to these base components first: header, nav, overview wallet card, wallet manager sheet/create-wallet flow, notification/search popovers, transaction rows, budget/planning rows, and account surface.
- Keep existing API contracts, backend behavior, database schema, auth, offline sync, and PWA service worker behavior unchanged.

## Out Of Scope

- TICKET-018 and later AI/OCR/bank ingestion UI.
- New banking service linking behavior.
- New charting library or report formula changes.
- A full component package extraction outside `frontend/src/app` unless the migration proves it is needed.

## Approach

1. Install `tailwindcss` and `@tailwindcss/vite`, add the Vite plugin, and import Tailwind in `frontend/src/styles.css`.
2. Convert `styles.css` into a Tailwind-first base layer:
   - `@theme` variables for accent, ink, muted, surface, background, danger, warning, and app max width.
   - `@layer base` for body, buttons, scrollbars, and safe mobile defaults.
   - `@layer components` for stable class names currently used by `App.tsx`, implemented with `@apply` where practical.
3. Introduce small React base components in `frontend/src/app/components.tsx` only where they reduce duplication without changing behavior.
4. Migrate screen markup incrementally while preserving accessible labels used by tests.
5. Run `npm test -- --run`, `npm run build`, `git diff --check`, then rebuild the web container.

## Risks

- Tailwind v4 compile errors can surface if `@apply` uses unsupported arbitrary values.
- Replacing class names too aggressively can break existing component tests and user flows.
- Liquid-glass effects need `-webkit-backdrop-filter` fallback for iPhone Safari.
- A wholesale rewrite could hide behavior regressions, so the first pass keeps stable public class names and test selectors.

## Verification

- Component/unit: `rtk npm test -- --run` from `frontend/`.
- Build: `rtk npm run build` from `frontend/`.
- Diff hygiene: `rtk git diff --check`.
- Runtime: `rtk docker compose up -d --build web`.
- UAT focus: mobile iPhone-sized viewport for bottom nav, wallet sheet, create-wallet flow, notification popover, and scrollbars.

## Reconciliation

- Update `docs/work/VALIDATION_MATRIX.md` for REQ-F-016 after proof.
- Update `docs/releases/CHANGELOG.md` with Tailwind/base UI migration.
- Update `docs/CONTEXT.md` recently touched areas and next steps.
- No ADR expected unless Tailwind changes module boundaries or build/runtime assumptions beyond Vite CSS compilation.
