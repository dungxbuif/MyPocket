# Credit wallet ledger implementation plan

1. Add `credit_kind`/payment linkage fields and repository operations that use
   the existing transaction ledger with atomic card-payment pairs.
2. Add owner-scoped credit entry/payment/statement routes and immutable ordinary
   transaction guards; regenerate Swagger.
3. Add a shared-base credit statement panel with purchase/refund/fee/interest
   and payment actions, then wire it into the wallet selector.
4. Reconcile API/design docs and run backend/frontend verification.

Out of scope: bank synchronization, installment amortization, multi-currency,
and automatic statement-close jobs.
