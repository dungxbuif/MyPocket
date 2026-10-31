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
