# Credit wallet ledger — implementation contract

## Scope

Add a dedicated credit operation API for purchases, refunds, fees/interest and
payments. Credit operations remain owner-scoped and are represented in the
existing transaction ledger so list/search/report drill-down can use one source
of truth.

## Decisions

- `credit_kind` is `purchase`, `refund`, `fee`, `interest`, or `payment` and is
  non-null only for credit-domain rows. Purchases/fees/interest are expense
  rows; refunds/payments are income rows. Amounts are always positive.
- Credit wallet balance is signed as current available debt: purchases/fees/
  interest subtract, refunds/payments add. `available_credit = credit_limit +
  current_balance`; a positive balance means overpayment.
- Purchases, refunds, fees and interest are report-included once according to
  their ledger direction. Payments create a paired excluded source-wallet
  expense and credit-wallet income row, so paying a card never becomes a
  second expense. The pair shares a payment ID and is atomic.
- Ordinary transaction create/edit/delete and balance adjustments continue to
  reject credit wallets. Credit rows are immutable through those routes; use a
  compensating credit operation instead.
- A credit operation may use an applicable visible category matching its
  derived income/expense direction. Payment categories are system transfer
  categories and are excluded from reports.

## Verification

Prove owner isolation, credit-wallet-only validation, signed balance/available
credit, report inclusion rules, atomic payment pairs, immutable ordinary-route
behavior and UI statement operations. Run backend/frontend full verification.
