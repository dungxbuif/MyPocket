---
artifact_type: verification
id: UI-BASE-01-VERIFICATION
status: verified
owner: shared
---

# UI base normalization — verification

[Design](UI-BASE-01-DETAIL_DESIGN.md) · [Backlog](../BACKLOG.md) · [Matrix](../VALIDATION_MATRIX.md) · [ADR](../../decisions/ADR-002-design-contract-enforcement.md) · [Design gateway](../../design/README.md) · [Release](../../releases/CHANGELOG.md).

## Acceptance

- [x] Seven source images read; observed text/design and HTML behavior extracted into seven component specs plus shared source register.
- [x] Conflicting palette/radius/behavior claims resolved or explicitly marked pending product decisions. Inventory does not claim all 28 targets are complete.
- [x] Removed exactly 7 PNG and 8 HTML exports after extraction; original versions recoverable in Git commit 1bc013d. Only Markdown remains in docs/design.
- [x] Canonical theme imported; duplicate token registry replaced with compatibility re-export. Domain badge mapping uses tone props.
- [x] Native controls, shared typography, card surfaces, feedback, progress/gauge, sheet, navigation and chart presentation use shared bases. Existing screens/routes and API/auth/data contracts preserved.
- [x] Guardrails run in production build; reject common base bypasses and check design-spec links/layout. Root AGENTS.md and standards route future UI work through the design gateway.
- [x] Screen Markdown records current behavior; source-only keypad/report and mock financial data remain explicitly partial.

## Automated evidence (2026-09-13)

Commands from repository root unless noted:

| Command | Result | Proof |
| --- | --- | --- |
| rtk proxy node app/scripts/check-design.mjs, before refactor | fail / expected red | Native controls outside atoms, literal colors and missing theme import reproduced |
| rtk proxy npm run test:design, cwd app | pass | 6 guardrail regression fixtures; SSR contracts for loading/disabled, one elevation, progress bounds, gauge danger, inline focus, wallet saving controls and empty donut |
| rtk proxy npm run build, cwd app | pass | Source guard, 17 Markdown files / 57 valid local design links, TypeScript and Vite production build |
| rtk git diff --check | pass | No whitespace errors |

The guard was strengthened after initial green to cover visual overrides, raw card/type recreation, aliases/local class constants and screen inline visuals. These intentional negative fixtures are not failed fix attempts. One React title-children warning was fixed by using a single string; subsequent contract test output is clean. No repeated failed fix/test path or loop-guard threshold reached.

## Browser / shared UI UAT

Fixture: app/tests/design.html served through Vite at localhost:4180. Chrome actual DOM/computed styles and screenshot inspected:

- Primary Lưu background rgb(76,175,80), matching brand; min-height 48px. Loading CTA disabled and labeled Đang xử lý.
- ArrowRight changed Khoản chi → Khoản thu and moved focus to selected tab.
- Mở chọn biểu tượng opened labeled dialog and focused Đóng.
- Shift+Tab from first control wrapped to Chọn.
- Escape dismissed dialog and restored focus to Mở chọn biểu tượng.
- Screenshot reviewed card/input/text styling, tree rows, 180° gauge and danger progress.

Shared component acceptance completed by agent browser check. No new owner product sign-off is inferred. Existing TICKET-01-03 owner closure remains unchanged. Real wallet/category CRUD, OAuth, ledger and budget UAT are not re-run here: no backend or domain contracts changed; their existing tickets own that evidence. Fixture does not contact production APIs.

## Docs review

- [x] Requirements checked: current product semantics retained; no update required. Static screenshot/demo values do not override SPEC.
- [x] API/ERD/security/runtime: no contract or deployment edits; no reconciliation required.
- [x] Architecture updated for theme/base ownership and build guard. SDD absent; architecture is the applicable master doc.
- [x] ADR-002 records standards/authority tradeoff; docs standards now require Markdown specs rather than duplicated PNG/HTML.
- [x] Source extraction, shared tokens/behavior, component inventory and current screen specs reconciled.
- [x] Context, feedback, backlog, validation and changelog updated; trace links checked.

## Remaining boundaries

Static guardrails catch known syntax patterns, not all possible semantic misuse or runtime-generated styles; component review remains necessary. No per-screen bypass was added. Full 28-base delivery, real calculator, date picker, tree collapse, time-aware budget warning and chart interactions remain planned/partial in the inventory. They are future feature work rather than hidden completion claims for this normalization.
