# Production login recovery — 2026-10-31

## Intended behavior

The production API must remain reachable at `https://money.dungxbuif.com` and
must authenticate the existing owner account without changing wallets,
transactions, budgets, jars, or other financial rows.

## Root-cause evidence

- After the API restart, the Swarm task exited before serving HTTP because
  `redis-prod` did not resolve from `mypocket-redis-net`.
- The legacy Redis endpoint on the documented Pi5 host port responds to an
  authenticated `PING`, while the overlay hostname does not resolve.
- The password reset must therefore be verified only after the API is serving
  and can reach both PostgreSQL and Redis.

## Recovery decision

Restore API availability with the existing immutable API image and the
documented legacy Redis endpoint as a bounded emergency runtime workaround.
Do not reset or migrate the database. Reconcile the Swarm Redis network/alias
from the manager path before making this workaround permanent.

## Verification plan

1. Confirm API health returns HTTP 200.
2. Confirm the exact owner email can log in with the current known password
   (or a newly generated password only if explicitly reset).
3. Confirm the login response is HTTP 200 and does not mutate financial data.
4. Record the final runtime state and remove any temporary container if the
   manager-side Swarm repair is applied.

## Execution result

- The documented Pi5 Redis host endpoint responded to an authenticated `PING`.
- The overlay hostname did not resolve from the VM100 task namespace, so the
  Swarm API task exited before serving HTTP.
- A bounded emergency API container using the existing immutable image and the
  legacy Pi5 Redis endpoint restored the public health route (`200`).
- The existing owner account password was replaced only after the live API was
  reachable; a real production login probe returned `200`.
- No financial tables were reset, dropped, truncated, migrated, or rewritten.
- Permanent remediation remains manager-side: attach the existing Redis
  container to `mypocket-redis-net` with the `redis-prod` alias, then update
  `mypocket-api`/`mypocket-worker` through Swarm and remove the emergency
  container after health/login verification.

## AI image-processing follow-up

The first authenticated image-processing request on 2026-10-31 reached
MyPocket successfully but failed at OCR submission with provider HTTP `503`.
The OCR proxy's `/readyz` reports `storage backend unavailable` because its
`redis-prod` hostname is missing from `macocr-redis-net`; the proxy's
`/healthz` and public OpenAPI endpoints remain available, which can otherwise
hide the dependency failure.

The temporary TCP forwarder created during diagnosis is not the desired fix and
must not be treated as production topology. The durable repair is to attach the
existing `redis-prod` container on the Pi5 Docker manager to both overlay
networks with the `redis-prod` alias, then converge the affected Swarm services.

The current workstation can reach the manager's private Semaphore API and has
the read-only Pi5 observer path, but it has no approved mutation key for the
manager. The observer identity is intentionally forced to `host-audit` and the
existing release identities are scoped to unrelated jobs. Therefore the
manager-side `docker network connect` cannot be safely executed from this
workspace until the approved manager/release key is restored or the operator
runs the fixed commands on Pi5.

Verification after the durable repair: authenticated Redis `PING` from both
overlay namespaces, MyPocket `/api/v1/health`, OCR `/readyz`, one authenticated
image-processing smoke request, then removal of the emergency API and TCP
forwarder only after the native Swarm tasks are healthy.

## Recovery result, 2026-10-31

After the PVE/VM100 network link was restored, SSH confirmed the worker was
online. The Pi5 data containers were recovered with their existing Compose
definitions and bind-mounted data: PostgreSQL reached `healthy`, PgBouncer was
running, and Redis was recreated by its existing Compose project without
changing its data directory. The existing Redis container was then attached to
`mypocket-redis-net` and `macocr-redis-net` with the `redis-prod` alias.

The native Swarm tasks recovered without changing application images:
NetBird server, MyPocket API/worker, and macOCR proxy are running. Verified
routes returned NetBird OIDC `200`, MyPocket health `200`, OCR health/ready
`200`, and KubeKosh `200`; the full edge scan showed no remaining gateway 502s
except the separate LLM tunnel timeout. Expected `401`, `403`, and `404`
responses on protected or path-specific probes were not classified as outages.

The emergency MyPocket API and Redis forwarder are stopped and no longer serve
traffic. They remain as stopped diagnostic containers for now; no financial
database reset, drop, truncate, or migration was performed.
