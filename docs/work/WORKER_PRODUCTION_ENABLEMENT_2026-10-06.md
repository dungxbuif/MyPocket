---
artifact_type: implementation_note
id: WORKER_PRODUCTION_ENABLEMENT_2026-10-06
status: in_progress
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

## Verification plan

1. Build `api`, `migrate` and `worker` for `linux/amd64` with warnings/tests
   passing, and verify all three binaries exist in the image.
2. Pull the immutable image on `vm100-prod` and run a one-shot worker service
   with the production environment and `WORKER_ONCE=true`.
3. Confirm the one-shot task exits 0 without duplicate occurrence rows, then
   update the long-running worker service to one replica with automatic
   rollback and verify stable logs.
4. Recheck API health, Redis alias connectivity and service task states.

The visible iOS Simulator is an independent host UI concern; this worker
change does not alter the iOS app or its simulator runtime.
