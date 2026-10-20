---
artifact_type: operational_contract
id: WORKER-RECURRING
status: implemented
---

# Recurring worker

`backend/cmd/worker` is a separate process sharing the API's PostgreSQL repositories
and account-timezone logic. Each pass handles four independent operations:

1. Materialize due recurring transactions for owners with active schedules.
2. Ensure current periods for active recurring budget series, using the existing
   revision, overlap and idempotency rules (requires migration `000035`).
3. Expire idle advisor conversations in batches of at most 100, using the
   server-authoritative 30-minute idle boundary (migrations `000033`/`000034`).
4. Materialize owner-scoped in-app budget notifications at 80% and 100% of the
   effective limit. The unique owner/dedupe key makes retries safe; this is not
   APNs delivery.

An error in one operation is logged and does not skip the other operations.
The worker does not duplicate schedule calculations or write month-report snapshots.
Budget notifications use each account's timezone and exclude transfers,
excluded transactions, future/ended budgets, and zero-limit budgets.

## Run modes

```sh
# long-running process (default: one pass every 60 seconds)
cd backend
go run ./cmd/worker

# one pass for a container job or external cron, from backend/
WORKER_ONCE=true go run ./cmd/worker
```

`WORKER_INTERVAL_SECONDS` changes the loop interval. Each pass discovers owners
with active schedules and calls the bounded, idempotent `RunDue` operation. The
database unique `(schedule_id, due_at)` occurrence marker prevents duplicate
rows after retries or overlapping worker instances. `DATABASE_URL` and the
usual local environment are loaded through the shared backend config.

The worker emits operational errors, not financial content or prompts. The
budget-series worker is present in the local candidate; the production rollout
record must confirm migration `000035` before claiming that operation is active
there. The advisor idle worker is recorded as deployed on 2026-10-16 in
[deployment](../operations/DEPLOYMENT.md). Budget inbox rows require migration
`000029`; APNs device registration is available through the authenticated
notification-device routes after migration `000036`, but provider delivery,
entitlements, credentials and retry/dead-letter handling remain separate and
are not implied by this worker pass.
