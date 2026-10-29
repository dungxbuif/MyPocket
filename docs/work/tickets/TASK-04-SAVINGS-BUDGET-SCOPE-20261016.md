---
artifact_type: work_note
id: TASK-04-SAVINGS-BUDGET-SCOPE-20261016
status: verified
owner: shared
trace:
  plan: ../../superpowers/plans/2026-10-16-gap-closure.md
  requirements: ../../requirements/BUSINESS_RULES.md
  reports: ../../requirements/REPORTS.md
  validation: ../VALIDATION_MATRIX.md
---

# Task 4 — Savings and budget scope audit

## Decision before implementation

This slice is documentation and existing-contract test coverage only. It does
not add financial semantics, migrations, routes, or notification producers.

The current contracts are:

- Goal wallets persist a positive `target_amount` and optional calendar
  `target_date`; displayed progress is derived from the current ledger balance.
  A withdrawal can move a goal from reached back to incomplete.
- Savings deposit/withdrawal/interest rows are ordinary single-wallet entries.
  They may come from outside the app and do not fabricate a counterpart.
- Moving money between two managed wallets uses the existing paired transfer
  operation. Both rows share a transfer ID and are excluded from report
  inclusion; this is not inferred from a savings category name.
- Goal notification preference and the owner-scoped in-app inbox exist. The
  inbox has list/read routes, but no scheduled goal/budget producer, push/APNs
  delivery, or fan-out contract is claimed.
- Budgets use one explicit inclusive calendar-date interval (`start_date` and
  `end_date`, interpreted in the account timezone). Current API progress is
  ledger-derived and includes matching expense descendants without counting a
  transaction twice in the active summary.

The following requested product areas remain partial because no concrete
current API/entity contract exists: recurring budget generation, end-of-period
forecasting, producer-driven budget/goal notifications, and a dedicated
savings report/drill-down endpoint. The requirements remain useful as target
behavior, but must not be read as shipped API behavior.

## Verification plan

- Add pure budget boundary coverage for account-local calendar dates and the
  existing explicit interval semantics.
- Add handler coverage proving authenticated owner propagation for the existing
  notification list/read routes.
- Retain existing wallet goal target/date, savings category, and paired
  transfer tests as the contract evidence for those paths.
- Run focused Go packages, then the full Go suite and `git diff --check`.

## Verification result

- Focused packages: 142 passed, 33 skipped.
- Full suite: `rtk go test ./... -count=1` — 311 passed in 16 packages.
- Race suite: `rtk go test -race ./... -count=1` — 311 passed in 16 packages.
- Static checks: `rtk go vet ./...` and `rtk git diff --check` — pass.

No production, database, device, provider, offline, AI, or Reports code is
changed by this work note.

