# Native overview transaction-title parity implementation plan

## Goal

Match the web Overview recent-transaction title rule: use the resolved
category name first, then fall back to `Khoản thu`/`Khoản chi` by transaction
type. Notes remain metadata and must not replace the semantic category title.

## Contract

- Load owner-scoped categories alongside the existing Overview data.
- Resolve recent-row titles with one pure helper shared by the view model:
  category name when present, otherwise income/expense fallback.
- Keep note, wallet, amount, masking and navigation behavior unchanged.
- Category failure must not block Overview; fallback titles remain safe.

## Verification

- [x] Add a pure unit regression for category, income and expense fallback
  titles.
- [x] Run the native unit suite, Release warnings-as-errors, project/base
  guards and diff checks.
- [x] Reconcile validation/audit/iOS docs and changelog; no backend/API/schema
  or migration change was made.

## Shipped evidence

- `OverviewViewModel` loads categories as optional supporting metadata and uses
  `transactionTitle(type:categoryName:)` for the web-equivalent category-first
  title rule; category failure still falls back safely.
- RED compile evidence is `/tmp/mypocket-overview-title-red`.
- Focused regression is 1/1 at `/tmp/mypocket-overview-title-unit/Logs/Test/`.
- Full native unit/contract suite is 96/96 at
  `/tmp/mypocket-overview-title-full/Logs/Test/`.
- Release warnings-as-errors build is `/tmp/mypocket-overview-title-release`;
  project/base-component and diff guards pass.
