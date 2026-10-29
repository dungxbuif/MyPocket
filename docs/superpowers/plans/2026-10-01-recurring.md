# Recurring transactions implementation plan

1. Add schedule/occurrence schema and owner-scoped repository with a bounded,
   atomic due-run operation.
2. Add HTTP contracts for list/create/update/delete and `run-due`, with the
   same auth/ownership/error conventions as wallet and transaction APIs.
3. Add a base-component schedule manager reachable from Account; keep generated
   rows in the existing transaction ledger.
4. Reconcile public docs and Swagger, then run backend/frontend verification.

Out of scope: bank sync, credit ledger, Travel Mode linkage, and automatic
deployment-specific cron wiring.
