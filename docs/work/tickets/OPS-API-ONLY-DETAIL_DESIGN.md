---
artifact_type: detail_design
id: OPS-API-ONLY
status: complete
owner: shared
updated: 2026-10-13
---

# Production API-only runtime

## Owner decision

The owner asked to turn off the deployed web UI while keeping the production
backend available for the native iOS app, and selected the previously presented
API-only route plan on 2026-10-13. The web source remains archived in the
repository; this change only affects production routing and the web service
replica count.

## Intended behavior

- `https://money.dungxbuif.com/api/v1/*` continues to reach the Go API.
- `GET /api/v1/health` remains healthy and `/api/v1/docs/index.html` remains
  available.
- The production root URL no longer serves the web application.
- `mypocket-web` is scaled to zero after the API route is verified.
- `mypocket-api`, `mypocket-worker`, PostgreSQL, Redis, object storage, and
  existing user data are not modified.

## Technical decision and scope

The current public Traefik router sends the MoneyPocket host to
`tasks.mypocket-web:80`; the web Nginx container then proxies `/api/` to the
backend. Change the Traefik service target directly to
`http://tasks.mypocket-api:8080` before scaling the web service to zero.
Keep API path prefixes unchanged. Do not add redirects or a replacement web
landing page in this slice.

This is a production runtime/configuration change. It does not change API
contracts, app source behavior, auth, database schema, migrations, secrets, or
the release version. The API docs surface is served by the Go API at
`/api/v1/docs/*any`.

## Rollout and rollback

1. Save a timestamped copy of the live Traefik dynamic config.
2. Point the MoneyPocket Traefik service at `tasks.mypocket-api:8080` and
   verify Traefik accepts the reload.
3. Verify public API health, OpenAPI/Swagger docs, login/auth reachability, and
   that `/` no longer returns the web app.
4. Scale only `mypocket-web` to zero.
5. Verify API and worker remain running and repeat public API checks.

Rollback: restore the timestamped Traefik config (which routes through
`mypocket-web`) and scale `mypocket-web` back to one replica. Do not touch the
database or other services.

## Verification evidence required

- Before/after Swarm replica state for API, web, and worker.
- Public health endpoint returns HTTP 200 after each routing step.
- Public API docs path `/api/v1/docs/index.html` returns HTTP 200.
- Root URL no longer serves the web application after web scale-down.
- No production database/schema/data operation was performed.

## Documentation reconciliation

The operations runbook, API access reference, backlog, validation matrix,
changelog and project context now reflect the live rollout. The Homelab source
route was updated narrowly while preserving unrelated working-tree changes.

## Completion evidence — 2026-10-13

- Live edge config now routes to `http://tasks.mypocket-api:8080`; backup:
  `/home/dungxbuif/edge/traefik/dynamic.yml.bak-mypocket-api-only-20261013`.
- Public health and `/api/v1/docs/index.html` return HTTP 200; `/` returns
  HTTP 404 instead of the web UI.
- Swarm state: API 1/1, worker 1/1, web 0/0.
- The corresponding Homelab source route was updated narrowly; unrelated
  existing Homelab worktree edits were preserved.
- Production data and schema were not touched.

## Security follow-up

An unfiltered production service inspection emitted environment values into
the operator task output. The values were not copied into repository files and
were not rotated during this rollout. Treat affected credentials as exposed;
credential rotation is a separate operational change requiring a controlled
window to prevent API/provider outages. See the security follow-up in the
deployment runbook.
