# Wallet scope documentation reconciliation

## Intent

Reconcile the wallet-selector reference, wallet page contract, and active backlog
with the wallet scope behavior already present in the web and native clients.
This is a documentation-only change: no API, schema, migration, or runtime
behavior is introduced.

## Current behavior to document

- The wallet management list groups included/excluded wallets, supports the
  aggregate `Tổng cộng` selection, and exposes add/edit actions.
- Selecting an existing basic wallet opens a wallet-scoped transaction ledger;
  the ledger selector supports `Tổng cộng` and individual wallets.
- Basic wallets use the same week/custom ledger layout as the aggregate,
  filtered by wallet. Goal and credit wallets route to their specialized detail
  surfaces and retain their specialized ledger semantics.
- Linked-service integration and authenticated visual/device UAT remain open;
  they are not reasons to describe the selector/filter implementation as
  missing.

## Affected references

- `docs/design/references/wallet-selector/README.md`
- `docs/design/pages/wallets/README.md`
- `docs/work/BACKLOG.md`

## Verification

- Search the updated documents for stale “local selection only” and “planned
  existing-wallet selection/filter” statements.
- Run `git diff --check`.
- Re-run the existing native/web base-component and design guards after any
  subsequent code change; this reconciliation itself does not alter runtime.
