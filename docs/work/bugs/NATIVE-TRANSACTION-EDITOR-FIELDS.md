# Native transaction editor field parity

## Finding

The SwiftUI ordinary transaction editor already sends `jar_id` and
`included_in_reports` in its API model, but the form did not expose either
field and always sent `jar_id = nil` and `included_in_reports = true`. This
made manual transactions from the ledger or wallet detail diverge from the
web Quick Add contract and from the jar/report business rules.

## Intended behavior

- Expense transactions can select an active hũ for the account-local month;
  income and transfer-out categories clear the hũ.
- An existing inactive hũ remains visible only when it is already assigned,
  and saving it after a date/month change is rejected until the user clears or
  replaces it.
- The user can explicitly choose whether the transaction is included in
  reports.
- The selected date and time are serialized in the account timezone.

## Verification

Add pure policy coverage for jar eligibility/options and run the native unit
tests, manual transaction UI path, full local API UI suite, Release build and
base/project guards.

## Resolution — 2026-10-06

The editor now exposes active account-month hũ choices for eligible ordinary
expense wallets, preserves an already-assigned inactive hũ so the user can
resolve it explicitly, sends the report-inclusion flag, and serializes the
account-local date and time. Type, wallet and category changes clear
incompatible hũ assignments. Policy tests cover wallet/category boundaries and
inactive selections; the manual editor UI assertion and full signed
`LOCAL_API_UI` suite pass. Native XCTest is 61/61, the Release simulator build
passes with warnings-as-errors, and base/project guards pass.
