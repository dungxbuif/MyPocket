# Advisor runtime fencing and atomic completion

Status: implemented in backend runtime slice; production rollout requires migration `000041_advisor_runtime`.

## Contract

`AdvisorRuntimeRepository` owns provider-worker fencing. A worker must claim a queued run with an owner, generation, and opaque token. Heartbeats can extend only the matching run lease and are bounded by the server-side deadline. A stale token, wrong owner, cancelled run, expired lease, or expired deadline is rejected with `ErrAdvisorLeaseLost` (or `ErrAdvisorRuntimeExpired` at claim time).

Completion is one PostgreSQL transaction. It inserts the assistant message, optional fact bundle, completion event, model usage, terminal run state, and compact conversation context together. A retry after an acknowledged/uncertain commit replays the existing assistant message and never inserts a duplicate. The partial unique assistant index is retained as a database backstop.

The reaper selects expired queued/running runs without holding locks, then rechecks each candidate under the shared conversation-then-run lock order. It marks deadline expiry as `failed/deadline_exceeded` and lost executor leases as `interrupted/executor_lost`; it never calls a provider or replays a request.

## Storage and rollout

Migration `000041_advisor_runtime` adds the non-null `advisor_runs.deadline_at` (backfilled from `created_at + 240 seconds`), an active-run reaper index, and a duplicate-assistant guard. Apply it after the existing advisor tables are available. The down migration is disposable-only because removing the deadline column would make old runtime workers unsafe.

## Verification

With `TEST_DATABASE_URL` pointing at a migrated PostgreSQL database, run:

```sh
go test ./internal/infrastructure/repository -run 'Advisor(Lease|Completion)' -count=1
go test ./migrations ./internal/entity
```

The repository tests cover single-claim fencing, cancelled/foreign completion rejection, and acknowledgement-loss replay. Full package tests remain gated by unrelated advisor resource changes until that slice compiles.
