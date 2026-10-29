# Native reports Stitch parity implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Recompose native Reports around the Stitch month-filter, summary, detail/trend and category surfaces while preserving real report API data.

**Architecture:** Keep `ReportsView.load()` and the existing insider/month endpoints. Add only presentation state for the detail/trend segmented control and compose the existing `MP*` bases into reference cards. The month text field and reload identifier stay for compatibility.

**Tech Stack:** SwiftUI, existing MyPocket iOS design system, XCTest/XCUITest.

**Spec:** Stitch report states `img_7697`–`img_7706` and `financial_clarity_system/DESIGN.md`.

## Global Constraints

- All amounts and categories come from the owner-scoped `insider`/`monthSummary` responses.
- Keep `reports-month`, `reports-load-button`, route title and back action stable.
- No new report endpoint, aggregation or timezone behavior.
- Use reference cards, segmented control, chart, metric and transaction base components.

## Review Focus

- Invalid or empty months keep the current error/empty state.
- Switching detail/trend does not trigger a second provider/API call.
- Missing income must not render a fake ratio.
- Masking is not applicable to this route; amounts remain real report values.
- Reload remains keyboard-safe and hittable.

### Task 1: Contract

**Files:** `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`

- [x] Add and run a failing Reports section-order contract.

### Task 2: Recompose Reports

**Files:** `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`

- [x] Add detail/trend presentation state and stable section IDs.
- [x] Replace the first summary block with a Stitch-style reference card while preserving the existing category/activity/detail sections and API load behavior.

### Task 3: Verify/docs

**Files:** `docs/design/references/STITCH-MY-POCKET.md`, `docs/work/VALIDATION_MATRIX.md`, `docs/work/PORT-100-AUDIT.md`, `docs/releases/CHANGELOG.md`

- [x] Run targeted/full native unit checks, build, guards and diff check; record exact results and keep authenticated visual UAT open.

## Follow-up: Money Insider detail parity

The web `ReportsPanel` renders the complete `fetchMoneyInsider` result after
the month summary: period change, average daily expense, spending/income ratio,
top categories and top expenses. The native route already fetched the same
owner-scoped `insider` response, but only exposed a partial category/expense
summary and omitted the Money Insider explanation card. This follow-up keeps
the existing endpoint and adds a shared base card so the native report exposes
the same information without screen-local styling or new aggregation logic.

### Intended behavior

- Render a shared Money Insider detail card after the month activity section.
- Use the existing `InsiderSummaryModel` values exactly as returned by the API.
- Preserve privacy masking for every monetary value; counts, labels and
  percentage context remain visible just as on the web.
- Show safe fallback copy when the previous period or income ratio is absent.
- Keep the existing report filter/reload contract and no new API or migration.

### Implementation and verification

- [x] Add the shared base card and mount it in `ReportsView`.
- [x] Add a unit contract for change/ratio fallback and masking.
- [x] Run the focused/full native tests, reports UI smoke, Release build,
  project/base-component/diff guards, and reconcile the validation docs.

Evidence for this follow-up:

- `/tmp/mypocket-reports-insider-unit/Logs/Test/` — 93/93 native unit/
  contract tests passed.
- `/tmp/mypocket-reports-insider-ui/Logs/Test/` — API-backed Reports UI
  smoke passed 1/1 and found `reports-section-insider`.
- `/tmp/mypocket-reports-insider-release` — Release build with warnings as
  errors passed.
- `verify-project.sh`, `check-base-components.test.sh` and `git diff --check`
  passed.
