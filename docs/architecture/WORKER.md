---
artifact_type: operational_contract
id: WORKER-RECURRING
status: implemented
---

# Recurring worker

`backend/cmd/worker` is a separate process sharing the API's PostgreSQL repositories
and account-timezone logic. Each pass handles six independent operations:

1. Materialize due recurring transactions for owners with active schedules.
2. Ensure current periods for active recurring budget series, using the existing
   revision, overlap and idempotency rules (requires migration `000035`).
3. Expire idle advisor conversations in batches of at most 100, using the
   server-authoritative 30-minute idle boundary (migrations `000033`/`000034`).
4. Materialize owner-scoped in-app budget notifications at 80% and 100% of the
   effective limit. The unique owner/dedupe key makes retries safe.
5. When APNs is explicitly enabled and configured, hand off inbox notifications
   to `push_deliveries`, lease bounded batches, call Apple's token-authenticated
   transport, and fence sent/retry/revoked/failed updates by lease.
6. Poll enabled market products once per provider and ten-minute UTC slot. The
   BTMC and BTMH adapters follow the official-source contracts used by the n8n
   Hanoi gold workflow, store only validated buyback quotes in VND per chỉ, and
   fence overlapping worker instances with `market_poll_claims`.

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
there. Market polling requires migrations `000044`, `000045`, and `000049`; the
catalog migration enables four stable BTMC/BTMH products. Source endpoint
overrides are `GOLD_BTMC_ENDPOINT` and `GOLD_BTMH_ENDPOINT`; they are deployment
configuration, not user input. The advisor idle worker is recorded as deployed on 2026-10-16 in
[deployment](../operations/DEPLOYMENT.md). Budget inbox rows require migration
`000029`; APNs device registration is available through the authenticated
notification-device routes after migration `000036`. When APNs is enabled, the
worker hands off inbox notifications to `push_deliveries`, leases bounded
batches, uses ES256 JWT against the configured sandbox/production endpoint, and
fences sent/retry/revoked/failed updates by lease. APNs is disabled by default;
valid `APNS_TEAM_ID`, `APNS_KEY_ID`, `APNS_TOPIC`, `APNS_PRIVATE_KEY` and
environment are required, and invalid enabled configuration fails worker
startup. Apple entitlements and signed-device acceptance remain external gates.
