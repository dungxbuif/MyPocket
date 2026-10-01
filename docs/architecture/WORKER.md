---
artifact_type: operational_contract
id: WORKER-RECURRING
status: implemented
---

# Recurring worker

`backend/cmd/worker` is a separate process for materializing due recurring
transactions. It shares the API's PostgreSQL repository and account-timezone
logic; it does not duplicate schedule calculations or write report snapshots.

## Run modes

```sh
# long-running process (default: one pass every 60 seconds)
go run ./backend/cmd/worker

# one pass for a container job or external cron
WORKER_ONCE=true go run ./backend/cmd/worker
```

`WORKER_INTERVAL_SECONDS` changes the loop interval. Each pass discovers owners
with active schedules and calls the bounded, idempotent `RunDue` operation. The
database unique `(schedule_id, due_at)` occurrence marker prevents duplicate
rows after retries or overlapping worker instances. `DATABASE_URL` and the
usual local environment are loaded through the shared backend config.

The worker emits only owner/count-independent operational errors; financial
content and prompts are not logged. Month-close snapshots, bank jobs and
notifications remain separate features.
