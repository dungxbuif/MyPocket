---
artifact_type: implementation_note
id: WORKER_PRODUCTION_ENABLEMENT_2026-10-06
status: complete
owner: shared
---

# Enable the recurring worker in production

## Intended behavior

Build the existing `backend/cmd/worker` binary into the same immutable API
image, keep the Swarm worker on the existing PostgreSQL/Redis production
networks, run one bounded `WORKER_ONCE=true` pass as a production smoke check,
then scale `mypocket-worker` to one replica. The worker must remain separate
from the HTTP API and must not create report snapshots or notification fan-out.

## Technical decision

The API and worker share one image digest so migrations, configuration parsing,
repository code and recurring idempotency rules cannot drift. The worker keeps
the current `DATABASE_URL` and stable `mypocket-redis-net` network attachment;
no production data reset or Redis restart is allowed.

## Verification result

1. PASS: built `api`, `migrate` and `worker` for `linux/amd64`; all three
   binaries exist in `release-73c4c37`.
2. PASS: a one-shot service on `vm100-prod` ran with `WORKER_ONCE=true`, used
   the production database and completed the active-schedule query cleanly.
3. PASS: `mypocket-worker` now runs one replica with automatic rollback and
   the stable `/app/worker` entrypoint.
4. PASS: API health remained HTTP 200 and Redis alias connectivity remained
   verified after the worker rollout.

The visible iOS Simulator is an independent host UI concern; this worker
change does not alter the iOS app or its simulator runtime.
