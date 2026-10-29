# Savings and Recurring Budget Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Deliver savings reports, recurring budget configuration and deterministic forecast in Go and native iOS.
**Architecture:** Owner-scoped ledger reads produce reports; versioned budget series create unique interval instances; native renders server results through existing bases.
**Tech Stack:** Go/Gin/GORM/PostgreSQL, SwiftUI/XCTest.
**Spec:** [Approved delegated design](../specs/2026-10-16-savings-budget-design.md).

## Global Constraints

- Keep existing explicit budget CRUD and base limit semantics.
- Preserve ledger data; use migration 035 for additive series schema only.
- All financial routes support existing owner JWT/API key auth.
- Integer money, account calendar dates, same predicates in totals/drilldown.
- Carryover disabled by default, positive-only when enabled; forecasts labeled.

## Review Focus

- Transfers between selected goals net to zero, distinct from surplus.
- No duplicate generated periods under concurrent worker/API execution.
- Backdated ledger edits update carry chain and estimates.
- Archived/foreign/deleted references never widen scope.
- Native old-server responses and failed requests never fabricate zero results.

### Task 1: Savings ledger reports (backend owner: savings agent)

Files: new entity/savings_report.go, usecase/savings_report.go,
controller/http/savings_report_handler.go and matching tests. Public guide
docs/guides/savings-budget-api.md savings section. Parent owns cmd/api wiring.

- [ ] Write pure report tests for 20M income, 12M expense, 5M goal transfer:
  surplus=8M, goal transfers=5M; internal goal transfer aggregate=0.
- [ ] Add missing-date/calendar bounds/zero-income/adjustment/future-row tests.
- [ ] Implement exact report and paged drilldown schema from spec using owned
  wallet/transaction/user repositories; validate predicates and filter ownership.
- [ ] Add handler owner/validation tests; run focused Go tests and update docs.

### Task 2: Recurrence (backend owner: recurrence agent)

Files: budget.go calculation additions, new budget_series.go entity/repository,
budget_series_postgres.go, series handlers/routes, budget_handler.go and
budget_postgres.go integration; migration 035; cmd/worker wiring. Parent owns
cmd/api wiring, forecast files and final generated Swagger.

- [ ] Add calendar/carryover/overflow tests, then implement pure functions.
- [ ] Add series/version persistence and mutation idempotency tests, then
  implement additive schema, owned locks and unique generation.
- [ ] Wire series ensure-on-list and worker catch-up, preserve snapshots,
  reject direct mutation of generated rows, stop dependent deleted references.
- [ ] Add JWT/API-key and concurrent-generation integration tests, then run
  focused real-fixture suite and reconcile recurrence guide/docs.

### Task 3: Forecast/drilldown (parent)

Files: new entity/budget_report.go + tests; controller/http/budget_report_handler.go
+ tests; cmd/api registration, generated Swagger.

- [ ] Test 3 days/300k spent/30 total -> 3M estimate; ended actual; future null.
- [ ] Implement integer/overflow-safe projection and zero-day daily breakdown.
- [ ] Use recurrence effective limits and common matching expense predicates.
- [ ] Add owned report/drilldown handlers and API-key negative tests.

### Task 4: Native feature composition (native agent)

Files: Core/API/FinanceAPI.swift, Core/Models/FinanceModels.swift additive models,
new Features/Reports/SavingsReportsView.swift, Features/Budgets/BudgetReportView.swift;
modify SecondaryFeatureViews.swift budget editor/report entry. Parent handles
project file registration only if needed after new file inventory is known.

- [ ] Add decoding/presentation tests with spec examples and old-server payloads.
- [ ] Add shared-base recurring controls, series management, savings sections,
  forecast and drilldown, with owner guard and busy/error/empty states.
- [ ] Run signed focused XCTest/build and record UI acceptance still outstanding.

### Task 5: Integration and delivery (parent)

- [ ] Review all agent changes without shared staging; parent creates commits.
- [ ] Run Go full/race/vet, real schema integration and signed native suite.
- [ ] Generate Swagger; update API/ERD/ADR/requirements/agent docs.
- [ ] Update backlog/context/validation/changelog with actual results.
- [ ] Build release artifacts; deployment/device acceptance is recorded separately.
