# Web trend chart parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Mount the same account-timezone daily income/expense trend on the web Reports surface so the Stitch trend reference is not native-only.

**Architecture:** Keep the backend unchanged. A pure TypeScript helper groups the existing owner-scoped transaction response by account-local day; `BaseTrendChart` owns the SVG geometry, semantics, empty state and accessibility label; `ReportsPanel` loads transactions in parallel with its existing summary calls.

**Tech Stack:** React 19, TypeScript, Vite SSR design tests, existing atomic base components and account-timezone helpers.

**Spec:** `docs/design/references/STITCH-MY-POCKET.md`, `docs/design/INDEX.md` trend chart row and the native counterpart in `2026-10-07-native-trend-chart-parity.md`.

## Global Constraints

- No API, schema, migration, provider, or dependency change.
- Filter `included_in_reports === true` and only `income`/`expense` types.
- Group by `accountMonthKey`/`dateKeyAt` in the profile timezone, not the browser timezone.
- Use design tokens/base components; no literal color or feature-local chart CSS.
- Empty months render a plain status message and never fabricated points.

## Tasks

### Task 1: Pure web trend aggregation

**Files:**
- Create: `app/src/services/reportTrend.ts`
- Create: `app/scripts/report-trend.test.mjs`

- [ ] Write the Vite SSR test first for timezone boundary, filtering, independent totals, sorted days and empty month.
- [ ] Run the focused test red because the helper is absent.
- [ ] Implement `buildReportTrend(transactions, month, timezone)` with a stable `ReportTrendPoint` type.
- [ ] Run the focused test green.

### Task 2: Shared web chart base and Reports integration

**Files:**
- Modify: `app/src/atomic/molecules/BaseCharts.tsx`
- Modify: `app/src/atomic/organisms/ReportsPanel.tsx`
- Modify: `app/scripts/wallet-scope-design.test.mjs` or a focused Reports SSR test

- [ ] Add a base chart contract test for the trend accessible role/label and empty copy.
- [ ] Run it red before mounting the component.
- [ ] Implement `BaseTrendChart` with token CSS variables, separate action/danger lines, bounded SVG geometry and no displayed money values.
- [ ] Load `fetchTransactions()` in Reports and mount a “Xu hướng chi tiêu” card using the helper and base; keep existing error handling and summary cards.
- [ ] Run focused SSR/build tests green.

### Task 3: Docs and gates

**Files:**
- Modify: `docs/design/INDEX.md`
- Modify: `docs/design/system/BASE_COMPONENTS.md`
- Modify: `docs/work/PORT-100-AUDIT.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`

- [x] Mark the trend base mounted on web and native.
- [x] Run web design tests, typecheck/build, native guards and `git diff --check`.

## Result

- `buildReportTrend` and `BaseTrendChart` are mounted by `ReportsPanel` and
  consume `fetchTransactions()` without changing the backend.
- RED/GREEN SSR cycles passed for timezone/filter/empty behavior and the chart
  accessibility contract.
- `npm run test:design` and `npm run build` pass; the existing Vite chunk-size
  advisory remains informational.
- Native 126/126, native API-backed Reports 1/1, Release and shared guards were
  rerun for the combined parity checkpoint.
