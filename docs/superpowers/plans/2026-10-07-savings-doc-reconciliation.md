# Savings/goal wallet documentation reconciliation

## Intent

Remove stale “transfer unfinished” statements from the savings-wallet design
contract after the shared transfer flow was implemented in the web and native
transaction routes.

## Contract

- External savings deposit/withdrawal/interest entries remain single-wallet
  transaction CRUD and do not fabricate a counterpart wallet.
- A managed-wallet transfer is a separate source/destination flow, excluded
  from reports by the existing ledger contract.
- Both web and native expose the shared transfer editor from the transaction
  flow; this doc must distinguish it from savings-specific external entries.
- Remaining visual/device UAT and notification/reporting gaps stay explicit.

## Verification

Reconcile the page doc against `QuickAddSheet`/`TransactionsPanel` and native
`TransferEditorView`, then run docs link/design checks and `git diff --check`.

Also reconcile the shared API/backlog references so they do not describe the
already-mounted Reports route or the implemented internal transfer editor as
future work.

## Result

The stale unfinished-transfer wording was replaced with the implemented
source/destination contract. External savings entries remain explicitly
single-wallet, while visual/device UAT and notification/reporting gaps are
still called out. API navigation and backlog entries were also reconciled so
mounted Reports and the transfer editor are not described as unimplemented.
The active savings ticket was reconciled as well; its historical mock-budget
and unmounted-Reports notes are explicitly marked superseded.
