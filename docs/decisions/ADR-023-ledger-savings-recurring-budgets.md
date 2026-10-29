# ADR-023 — Ledger savings reports and versioned recurring budgets

Status: accepted by delegated owner decision, implemented locally; rollout and
owner/device UAT pending. [Design](../superpowers/specs/2026-10-16-savings-budget-design.md),
[plan](../superpowers/plans/2026-10-16-savings-budget-implementation.md),
[reports](../requirements/REPORTS.md), [API](../architecture/API.md),
[ERD](../architecture/ERD.md), [validation](../work/VALIDATION_MATRIX.md).

## Context and decision

The owner selected both goal-wallet reporting and account-wide surplus and
delegated carryover details. Keep those two measures separate: a paired
transfer changes goal balances without becoming ordinary income/expense.
Savings reports and budget forecasts are recalculated from owned ledger rows,
with shared drilldown predicates, integer money and account calendar dates.

Recurring configuration lives in `budget_series` and dated
`budget_series_configs`; generated budgets retain a configuration snapshot.
Owner locking and unique series/period index serialize worker and API catch-up.
Use the existing mutation receipt ledger plus expected version for writes.
Configuration edits take effect next period; stop preserves generated periods.
Direct edits/deletes of generated budgets return a managed-series conflict.

Carryover is optional, off by default, positive-only and recalculated after
historical ledger edits. Scope changes reset carryover. Forecast is a labeled
linear daily projection, never an LLM estimate or persisted financial counter.
Migration035 is additive; scope-deletion triggers stop affected series instead
of broadening nullable scopes. Archived revisions retain scoped IDs as snapshots.

## Alternatives and consequences

Rejected overwriting a single budget at rollover: loses period configuration.
Rejected frozen carryover counters: stale after late receipts/corrections.
Rejected autonomous AI forecasting: harder to reproduce or reconcile money.

Queries must read prior generated periods to recalculate carry chains. Large
catch-up is batched and may need several worker passes. Historical goal targets
use current configuration because no target revision model exists. Reports are
live reads; paged results can change between requests if the ledger changes.
Push producers and predicted savings completion dates remain future work.
