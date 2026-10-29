# MyPocket Money Lover Reference Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the complete native MyPocket presentation around the interaction and layout grammar observed in Money Lover while preserving MyPocket branding, data rules, security and extended functions.

**Architecture:** Shared SwiftUI tokens and bases own all reference-derived geometry and behavior. Feature screens migrate in independently testable waves; no feature screen introduces local visual constants when a shared variant is required. Existing APIs and financial models remain unchanged unless a separately approved contract says otherwise.

**Tech Stack:** SwiftUI, XCTest/XCUITest, existing MyPocket design-system scripts, signed iOS simulator and physical-device UAT.

**Spec:** `docs/superpowers/specs/2026-10-26-money-lover-reference-redesign.md`

## Global Constraints

- Do not copy Money Lover trademarks, icons, illustrations, wording, screenshots or sample values.
- Keep every MyPocket-only capability reachable.
- Keep API, schema and accounting semantics unchanged for visual parity work.
- Use SF Symbols/existing MyPocket assets and SF system typography.
- Keep 44pt minimum iOS touch targets, Dynamic Type, safe areas, privacy masking and predictable Back behavior.
- Write/update the owning screen/base docs before each implementation and reconcile them after tests.
- Do not mark a production feedback record resolved until its deployed owner acceptance passes.

## Review Focus

- Small-screen and large-text layouts must keep headers, rows, footer actions and navigation reachable without overlap; Task 1/2 UI tests own this.
- Keyboard presentation must not hide Save, Add, search clear/close or feedback actions; Tasks 2/3 tests own this.
- Empty, loading, error, stale-cache and offline-pending states must retain their meaning through visual migration; every feature wave has one state test.
- Privacy masking must hide every financial value the current screen hides, without leaking it through summaries or accessibility labels; Tasks 3–6 own screen-specific assertions.
- Reference parity must not change report totals, transfer exclusion, category ownership, portfolio arithmetic or AI confirmation; Task 7 runs integrated business regressions.

---

### Task 1: Semantic tokens and grouped native surfaces

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/Tokens.swift`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPSurfaces.swift`
- Modify: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`
- Modify: `docs/design/system/TOKENS.md`
- Modify: `docs/design/system/BASE_COMPONENTS.md`

**Interfaces:**
- Produces: reference surface/canvas/separator/row metrics as semantic `MPColor` and `MPMetrics` values; grouped surface and inset-divider variants consumed by all later tasks.

- [ ] Write failing token/base tests for light/dark semantic colors, 44pt row targets, inset divider alignment, grouped-surface radius and Dynamic Type expansion.
- [ ] Run focused design-system tests and verify failures identify missing variants.
- [ ] Add only semantic tokens and shared variants; do not add screen-specific raw colors or spacing.
- [ ] Run focused tests plus `mobile-ios/scripts/check-base-components.sh` and reconcile token/base docs.
- [ ] Commit the isolated foundation change.

### Task 2: Shell, headers, bottom navigation and editor scaffold

**Files:**
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPNavigation.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Shell/MainTabView.swift`
- Modify: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`
- Modify: `docs/design/platforms/ios-base-guideline.md`
- Modify: `docs/design/platforms/ios.md`

**Interfaces:**
- Consumes: Task 1 semantic surfaces/metrics.
- Produces: safe-area floating navigation capsule, selected-tab surface, elevated Add action, shared centered/cancel/save headers and fixed editor footer.

- [ ] Add failing iPhone SE/large-iPhone UI assertions for content viewport, nav capsule, raised Add, feedback clearance, keyboard hide/show and editor footer reachability.
- [ ] Run focused UI tests to prove current geometry fails.
- [ ] Implement shell/header/editor variants through shared bases and keep Assistant/Feedback reachable.
- [ ] Run focused UI tests on both devices plus base/project guards and reconcile shell docs.
- [ ] Commit the shell change independently.

### Task 3: Daily money loop

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Overview/OverviewView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionsView.swift`
- Modify: transaction editor/detail source under `mobile-ios/MyPocketIOS/Features/`
- Modify: `mobile-ios/MyPocketIOSTests/OverviewReferenceTests.swift`
- Modify: `mobile-ios/MyPocketIOSTests/FinanceModelTests.swift`
- Modify: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`
- Modify: `docs/design/pages/overview/README.md`
- Modify: `docs/design/pages/transactions/README.md`
- Modify: `docs/guides/transactions.md`

**Interfaces:**
- Consumes: Tasks 1–2 bases.
- Produces: wallet-scope ledger, period rail, income/expense/net summary, category/note search, day/category groups, reference-style transaction editor and detail.

- [ ] Add failing model/UI tests for wallet+period-scoped search, group totals, note/date rows, privacy masking, empty search and keyboard-safe editor.
- [ ] Integrate the in-flight REA search slice into the shared header/period/row design; remove no existing ledger capability.
- [ ] Migrate Overview and transaction create/edit/detail using shared grouped rows and footer actions.
- [ ] Run focused units/UI, offline pending/conflict tests and screen guards; reconcile user and agent-readable transaction docs.
- [ ] Commit the daily-money wave.

### Task 4: Wallet and organization flows

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Wallets/WalletsView.swift`
- Modify: relevant category/recurring/credit/savings views under `mobile-ios/MyPocketIOS/Features/`
- Modify: `mobile-ios/MyPocketIOSTests/CategoriesRecurringReferenceTests.swift`
- Modify: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`
- Modify: owning docs under `docs/design/pages/{wallets,categories,savings,jars,recurring,credit}/`

**Interfaces:**
- Consumes: shared scope selector, grouped forms, calendar/selectors and editor shell.
- Produces: coherent wallet list/type/editor/detail, category hierarchy/pickers, savings/jars, recurring/bills/debt and transfer flows.

- [ ] Add failing reference-layout contracts for each owned flow, including
  system-locked categories and transfer report exclusion. Unit tests may prove
  mounted draft state and persisted-cache eligibility; API write/readback and
  rendered SwiftUI hierarchy require their existing integration/UI fixtures and
  must remain open when those fixtures are outside a bounded feature slice.
- [ ] Migrate one bounded feature family at a time, running its focused tests before the next.
- [ ] Preserve existing validations, ownership, recurrence, paired-transfer and offline behavior.
- [ ] Run the organization slice and reconcile screen/user docs.
- [ ] Commit each feature family separately.

### Task 5: Planning, reports and chart hierarchy

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Reports/SavingsReportsView.swift`
- Modify: budget/report/month/credit views under `mobile-ios/MyPocketIOS/Features/`
- Modify: `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`
- Modify: relevant model/UI tests
- Modify: owning docs under `docs/design/pages/{budgets,reports,month-detail,credit}/`

**Interfaces:**
- Consumes: Tasks 1–2 metric/surface/selection bases.
- Produces: budget summaries/progress, report overview/drilldowns, category/trend/comparison chart families and month hierarchy.

- [ ] Write failing model/presentation tests that use existing calculated values
  and over-limit/empty/privacy states. Do not describe presentation-policy tests
  as rendered hierarchy proof; retain screenshot/UI acceptance as a separate
  gate when no bounded fixture mounts the screen.
- [ ] Add shared chart/summary variants before migrating feature views.
- [ ] Migrate reports and planning screens without changing arithmetic or API payloads.
- [ ] Run finance/report suites, screenshot checks and reconcile docs.
- [ ] Commit planning/report families independently.

### Task 6: MyPocket-only extensions under the same grammar

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/More/MoreView.swift`
- Modify: feature tests and owning docs for Assistant, AI entry, Portfolio, Travel, Notifications, Feedback and Account.

**Interfaces:**
- Consumes: all shared bases and migrated feature patterns.
- Produces: visually coherent extended flows without reducing AI confirmation, portfolio corrections, security/API-key controls, feedback or offline recovery.

- [ ] Add focused tests for proposal cards, attachment review, portfolio editors, feedback dock, Face ID entry, API-key masking and owner-only actions.
- [ ] Recompose existing feature logic with shared surfaces/navigation/forms; do not rewrite business logic for appearance.
- [ ] Run provider-independent tests and signed simulator visual checks; record provider/device gates separately.
- [ ] Reconcile human and agent-facing guides for changed navigation/presentation.
- [ ] Commit each extension independently.

### Task 7: Feedback reconciliation and integrated candidate

**Files:**
- Modify: `docs/work/audits/FEEDBACK-UNRESOLVED-UAT-20261020.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`

**Interfaces:**
- Consumes: Tasks 1–6 tested commits and the private 21-item inventory.
- Produces: one exact release candidate manifest mapping each feedback ID to deployed code, UAT, decision blocker or remaining work.

- [ ] Review every local feedback patch against its work note and correct unsupported “shipped” claims.
- [ ] Run full backend tests, native units/UI slices, base/project/design guards and production build.
- [ ] Build a signed Release candidate and record commit/image/build identifiers.
- [ ] Do not mark decision-blocked items complete; identify only owner actions that cannot be automated.
- [ ] Commit documentation reconciliation.

### Task 8: Controlled rollout and owner acceptance

**Files:**
- Modify: `docs/operations/DEPLOYMENT.md` only if the verified workflow changes.
- Modify: release/validation/feedback evidence documents.

**Interfaces:**
- Consumes: Task 7 release candidate.
- Produces: installed physical build, production smoke evidence and owner-resolved feedback statuses only where acceptance actually passes.

- [ ] Back up/version-check affected production services and verify rollback artifacts.
- [ ] Deploy only required backend/worker configuration and native build; never bundle unrelated financial mutations.
- [ ] Run authenticated smoke, signed physical-device UAT and reference screenshots on populated owner data.
- [ ] Resolve feedback records individually after owner acceptance; leave decision/provider/device-blocked records open with exact reason.
- [ ] Reconcile final docs and publish the release evidence.
