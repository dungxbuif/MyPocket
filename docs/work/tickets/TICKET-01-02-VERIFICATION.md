---
artifact_type: test_verification
id: VERIFY-01-02
status: verified
owner: ai
trace:
  ticket: TICKET-01-02-quan-ly-vi.md
  detail_design: TICKET-01-02-DETAIL_DESIGN.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Verification — Quản lý ví

Verified 2026-09-13 for the implemented core wallet slice.

## Proven behavior

- Owner-scoped create, list, edit and permanent delete for `basic`, `goal` and `credit` wallets.
- Duplicate names remain allowed; IDs own all relationships.
- Goal requires a positive target amount; credit requires a positive credit limit. Statement-cycle metadata remains in the later credit slice.
- `current_balance` is derived from opening balance plus income minus expense for the wallet ledger. Only wallets with `is_in_total=true` contribute to the header total.
- Editing wallet metadata cannot overwrite opening balance; balance adjustment remains a separate ledger slice.
- Wallet type is immutable after creation so an existing ordinary ledger cannot be reinterpreted as credit debt (or the reverse).
- Delete confirmation states the exact number of dependent transactions and report impact; database cascade removes those transaction rows.
- UI uses shared base fields, cards, statuses and buttons; no screen-local primitives or colors were introduced.

## Evidence

- `rtk go test ./...` in `backend/`: pass, including type validation and ledger-balance tests.
- Dev PostgreSQL migration: `rtk go run ./cmd/migrate up`: pass at existing migration version.
- Real HTTP UAT: created basic/goal/credit wallets; income and expense changed a basic wallet from 100,000 to 115,000, update changed it to 113,000, delete changed it to 93,000; deleting the wallet reduced its transaction count from 1 to 0.
- Real HTTP metadata UAT: patching a basic wallet with `type=credit` and `opening_balance=999999` kept `type=basic`, `opening_balance=100000`, `current_balance=100000`, and cleared the irrelevant credit limit.
- Chrome UAT through the Vite proxy: creating a 100,000 VND wallet updated Header and wallet card; a 25,000 expense updated both to 75,000. A detected stale wallet-card refresh was fixed and rechecked.
- `npm run test:transactions`, `npm run test:design`, `npm run build` in `app/`: pass.

## Residual scope

Target date, credit statement/payment fields, balance adjustment and the credit purchase/payment ledger remain separate design work. They are not claimed by this verification.
