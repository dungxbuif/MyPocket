---
artifact_type: detail_design
id: CORE-03-TIME-JARS-MONTH
status: in_progress
owner: shared
approval: owner_approved_implementation_in_conversation
trace:
  backlog: ../BACKLOG.md
  timezone_ticket: TICKET-01-01-thiet-lap-ca-nhan.md
  jars_ticket: TICKET-04-hu-chi-tieu.md
  month_ticket: TICKET-07-04-tong-ket-thang-ai.md
  requirements: ../../requirements/BUSINESS_RULES.md
  reports: ../../requirements/REPORTS.md
  erd: ../../architecture/ERD.md
  api: ../../architecture/API.md
  validation: ../VALIDATION_MATRIX.md
  release: ../../releases/CHANGELOG.md
  adr: ../../decisions/ADR-008-account-timezone-and-calendar-dates.md
---

# Account timezone, jars, and automatic monthly summary

## Outcome and approved decisions

Deliver the three unfinished user-facing capabilities as one coherent ledger slice: account-wide IANA timezone, optional jar tracking with per-month configuration and cumulative reports, and a live monthly overview that becomes automatically marked complete when the account enters a later month. This implements ACC-03, TIME-01–05, JAR-01–11, and the non-AI parts of MONTH-01–03. The user has already approved implementation and settled the material product decisions in the conversation.

- Existing accounts start in `Asia/Ho_Chi_Minh`; a new account adopts the browser IANA zone once at onboarding and may change it in Account settings.
- Instants remain UTC `timestamptz`; calendar dates/month labels are date-only values. Month intervals are half-open and derived in the saved account zone.
- Changing timezone changes timestamp-based grouping only. It does not rewrite UTC instants, date-only labels, month-note keys, or stored month configuration.
- Jar IDs are stable. Names and optional fixed-VND/actual-income-percent allocations belong to an account/month snapshot. A new month copies the nearest prior initialized month's configuration exactly once. Removing a jar from one month removes only that configuration row; prior transaction links and other month snapshots remain.
- A transaction may have no jar or one jar. Only ordinary expense rows may be assigned. Unassigned expenses remain valid. Historical spend is derived from the live ledger, so edits and deletes recalculate naturally.
- Month completion is derived automatically from the current account-local month; it never locks data or requires a close action. A month note is independently stored by `(owner, YYYY-MM)` and never generated or overwritten by calculations.
- The Overview shows the current month. A month detail view supports past/current month navigation, totals, category/jar context where available, and editing the independent note. AI-generated narrative, recurring budgets, paired transfer/adjustment, and credit/debt ledger semantics are separate work.

## Current-state evidence and scope

Hydration read `docs/CONTEXT.md`, `docs/work/BACKLOG.md`, `docs/standards/{README,QUALITY_BAR,VALIDATION}.md`, jar children TICKET-04-01–03, account settings TICKET-01-01, monthly summary TICKET-07-04, BUSINESS_RULES, REPORTS, ERD, and the design contracts. No active phase exists. Current code has no jar/month-note tables or APIs and no account timezone field. Budget date bounds and wallet target date are currently `timestamptz`; transaction and audit timestamps are instants. Preserve existing user data and unrelated dirty worktree changes.

## Design

### Persistence and time rules

Add an account `timezone` IANA name and `timezone_confirmed` flag. The additive migration gives existing accounts `Asia/Ho_Chi_Minh` and marks them confirmed; new accounts receive that zone provisionally and set it to the browser zone once only. Explicit settings changes mark it confirmed. Validate every value with Go's IANA location database. Invalidate cached profile data after a change.

Keep `transactions.occurred_at`, audit timestamps, AI-processing timestamps, and attachment expiry timestamps as UTC instants in PostgreSQL `timestamptz`. Convert `wallets.target_date`, `budgets.start_at/end_at` calendar selections to SQL `date` and API `YYYY-MM-DD`; preserve existing visible labels during migration using the existing-account default timezone for budgets and the stored UTC date for wallet targets. Derive budget and report instants at query time from the owner's current timezone. No server-local timezone participates in calculations.

Introduce shared Go month parsing and `[start,next)` UTC boundary helpers. Validate `YYYY-MM`, use `time.LoadLocation(account.Timezone)`, and compute current month and completion from `now.In(location)`. Frontend date/time formatting and conversion use a shared account-zone module based on `Intl`; reject nonexistent local times and deterministically choose the earlier instant for ambiguous repeated wall times. If an unchanged edit retains its original UTC instant, do not round-trip it through local text.

### Jars and transaction link

Add stable owner-scoped `jars`, one-time initialized `jar_months`, and `jar_month_configs` keyed by owner/month/jar with name, allocation mode (`none`, `fixed`, `percent`), fixed VND amount or percentage basis points. A transactional owner-row lock serializes month initialization and config writes. Initialize a month from its nearest strictly earlier initialized month once; an empty initialization is still recorded. Reports calculate each selected month using its own saved config and live transactions.

Add nullable `transactions.jar_id` with same-owner foreign-key protection. Handler/use case validates selected jar config exists for the transaction's account-local month and rejects income, transfer-category expenses, and out-of-scope types. Manual transaction editing and AI proposal review can select/clear one configured jar; extraction itself does not guess a jar. Month spend and allocation are derived, never cached as mutable balances. Actual-income percentages use included, ordinary income and exclude internal transfer categories. Warnings for over-allocation or overspend never block saves.

Expose authenticated owner-scoped endpoints to list/create jars, replace/remove a month config, retrieve monthly jar summaries, and retrieve a stable-jar cumulative breakdown over a month range. Delete removes the current month's config only; it does not delete the stable identity, linked transactions, or historical configurations.

### Monthly overview and note

Add `month_notes` keyed by `(owner_id, month DATE)` with note and audit timestamps. Add an authenticated month-report API accepting only `YYYY-MM`; the server resolves account timezone, range, and current/complete state, and aggregates included ordinary income/expense from persisted transactions. Internal transfers/adjustments do not inflate those totals. It returns the timezone, month label, `[start,next)` instants, totals, category breakdown, jar totals/coverage when available, note, and an `is_complete` flag computed against current local month. A note endpoint upserts or deletes only the selected month's note; reads and recalculations never mutate it.

There is no scheduled close job: the month changes state at the account-local month boundary from the same persisted ledger/report query, including after downtime. This avoids duplicate close jobs and stale frozen totals. AI prose and external calendar context stay outside this implementation.

### Frontend composition

Account timezone settings reuse `FormField`, `BaseSelect`, `BaseButton`, `StatusMessage`, and shared profile presentation. Onboarding initializes once from `Intl.DateTimeFormat().resolvedOptions().timeZone`; failure leaves the existing valid account zone intact.

Overview and new month detail screens use shared `SurfaceCard`, `Text`/`Heading`, `BaseButton`, `BaseSelect`, `BaseTextArea`, and existing chart/progress bases only. Add a documented `/jars` screen for month configuration and cumulative view using the same bases. Transaction and AI result editors expose a single optional jar selector sourced from that month's real API configuration. Screen classes remain layout-only; no new colors, custom controls, or mock finance rows.

## Files and contracts expected to change

- Backend migrations/entities/repositories: `backend/migrations/000013*` onward; `internal/entity/{user,wallet,budget,jar,month}.go`; user, budget, transaction and new jar/month repositories.
- Backend API/use cases: account profile handler/interactor/routes; transaction and AI proposal validation/approval; new jar/month handlers/routes; Swagger generated from annotations.
- Frontend API/state: `app/src/services/{auth,transactions,budgetDates,formDates}.ts` plus timezone/jar/month services and account timezone provider.
- Frontend consumers: `AccountPanel`, `FinancePrototypePage`, `OverviewPanel`, new month/jar panels, `QuickAddSheet`, `TransactionsPanel`, and AI proposal cards.
- Contracts/docs: account, transaction, budget, Overview, month and jar screen specs; BASE_COMPONENTS only if a required shared behavior is missing; ERD/API/REPORTS/BUSINESS_RULES, validation matrix, backlog, context and changelog.

Reuse bases: `FormField`, `BaseSelect`, `BaseTextArea`, `BaseButton`, `BaseSwitch`, `SurfaceCard`, `Text`, `Heading`, `Progress`, `StatusMessage`, `BaseBottomSheet`, `DateField`, `AmountField`, `BaseCategoryTree`, `CategoryTreeSelector`, `AssistantResultCard`, `TransactionItem`, and `BudgetProgressItem`. Do not add a shared base unless these contracts prove insufficient.

## Alternatives considered

- Persisting a close job/snapshot was rejected: it creates catch-up, retry, duplicate-close and stale-report cases, while requirements explicitly keep closed months recalculable.
- Storing date-only fields as UTC midnight was rejected: it shifts calendar labels across timezones.
- Hard-deleting jars or transaction links was rejected: it destroys historical attribution and cumulative reporting.
- Treating jar allocation as wallet money or auto-carrying variance was rejected by JAR-07/JAR-11.
- Letting the LLM infer/commit jar assignment was rejected: the user can select a jar while reviewing the editable proposal; ledger writes still require explicit approval.

## Impacts and risks

Code/API/data/auth-owner scoping and UI all change. Migrations must be additive/backfilled and rollback must preserve recoverability. Main risks are date-label preservation for existing budget rows, DST gaps/folds in local datetime entry, cross-owner jar references, concurrent month initialization, historical config copy order, duplicate aggregation through category trees, and stale cached timezone. Owner scopes remain enforced in every repository/API. No provider, auth model, or external service is added.

## Verification and reconciliation

Required evidence is recorded in `docs/work/VALIDATION_MATRIX.md`: migration preservation/rollback review; timezone boundary and DST behavior; owner-isolation and concurrent initialization; ordinary-vs-transfer aggregation; jar month-copy/archive/history semantics; note independence and automatic completion; FE design gates/build; real API/browser UAT. The requested implementation is in progress; no completion claim is made until that proof and docs review exist.

### Execution checkpoint — 2026-09-22

Migrations `000013`–`000015`, account timezone/date-only APIs, owner-scoped jar/month APIs, optional manual/AI transaction jar assignment, and `/jars`, `/months/YYYY-MM`, and Overview integration are implemented. The local migration ledger is version 15 and clean. Full backend PostgreSQL tests and frontend design, transaction, AI, calendar, jar-selector and build checks pass; evidence and owner-UAT gaps are recorded in [VALIDATION_MATRIX](../VALIDATION_MATRIX.md#core-03-timezone-jars-and-live-monthly-summary--2026-09-22). Runtime routes return HTTP 200, but owner data was not modified and visual/functional owner UAT remains pending. Therefore CORE-03 and its backlog row stay `in_progress`.

The API/ERD, business/report rules, screen specs, backlog, validation matrix, context and changelog are reconciled to the current live/recalculable report behavior. [ADR-008](../../decisions/ADR-008-account-timezone-and-calendar-dates.md) records the implemented date/time contract as proposed pending human decision status. The separate [WORKER-01](WORKER-01-cron-service.md) request includes month-end report/close generation; clarify that conflict before changing CORE-03 or designing a persisted close artifact.

Ticket links: [TICKET-01-01](TICKET-01-01-thiet-lap-ca-nhan.md), [TICKET-04](TICKET-04-hu-chi-tieu.md), [TICKET-07-04](TICKET-07-04-tong-ket-thang-ai.md).
