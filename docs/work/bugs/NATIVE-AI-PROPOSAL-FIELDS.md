# Native AI proposal field parity

## Finding

The native AI proposal editor had the shared jar/report controls, but it used a
date-only control and silently cleared an existing inactive hũ when the month
summary loaded. It also allowed the transfer-out category to display a hũ
selector, unlike the web proposal row contract.

## Intended behavior

- Preserve the proposal's account-local date and time when editing.
- Offer active account-month hũ options for ordinary expense rows and keep an
  existing inactive assignment visible with an explicit validation error.
- Never offer or persist a hũ for income, credit wallets, or the
  `expense_transfer_out` category.
- Keep the report-inclusion choice in the proposal patch.

## Verification

Add pure policy coverage and extend the deterministic proposal UI assertion;
run native unit/UI suites, Release warnings-as-errors, base/project guards and
reconcile the public validation docs.

## Resolution — 2026-10-06

The proposal editor now uses account-local date/time, keeps inactive existing
hũ assignments visible for explicit resolution, rejects stale hũ values on
save, and excludes the transfer-out category from hũ selection. Report
inclusion remains editable and is sent in the versioned proposal patch. The
policy regression and deterministic simulator proposal review pass; the full
unit/UI, Release and guard results are recorded in the validation matrix.
