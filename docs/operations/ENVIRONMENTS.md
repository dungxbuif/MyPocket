---
artifact_type: environments
id: ENVIRONMENTS
status: active
owner: shared
updated: 2026-09-01
---

# Environments

| Environment | Purpose | URL | Host | Port | Notes |
| --- | --- | --- | --- | --- | --- |
| production | Public production | `https://money.dungxbuif.com` | Mac mini (10.10.0.10) | 18082 (api) / 18083 (web) | Docker Compose, Pi5 Postgres/Redis/RustFS |

## Dependencies

| Service | Host | Port | Auth | Notes |
| --- | --- | --- | --- | --- |
| PostgreSQL | 10.10.0.5 (Pi5 via PgBouncer) | 5432 | mypocket:WczE233UdMxMb1NW24IGeSLVhEUXSGWp | DB: mypocket, schema public owner: mypocket |
| Redis | 10.10.0.5 (Pi5) | 16379 | SldPateMZwZm2QKpHnKT5kDEDkMcGBFH | DB index: 1 (dedicated) |
| S3 (RustFS) | storage.dungxbuif.com | 443 | Access Key: AW4ZVX9S23BIWXZTK0IG / Secret: d6CTf9ucl5+1eWZzImBoZU0U5zmgPWgKRMqw9tc+ | Bucket: mypocket, policy scoped to bucket only |
| Image Registry | registry.dungxbuif.com | 443 | User: dungxbuif | Private registry on Pi5 |

## Routing

```
Internet → Cloud VPS (Traefik SNI whitelist) → Rathole tunnel → Orange Pi (Caddy TLS) → Mac mini (Docker Compose)
```

| Domain | Caddy Proxy Target | Notes |
| --- | --- | --- |
| `money.dungxbuif.com` | `10.10.0.10:18083` | Web SPA (Nginx serves `/api/` → `api:8080` internally) |

## Stack Layout

```
~/production/mypocket/
├── docker-compose.yml    # 4 services: api, web, worker, migrate (one-shot)
├── .env                  # secrets (mode 600) — from homelab/local_vars.json
```

## Services

| Container | Service | Image | Host Port | Internal Port |
| --- | --- | --- | --- | --- |
| mypocket-api | API | `registry.dungxbuif.com/mypocket-api:prod-2026.08.31` | 18082 | 8080 |
| mypocket-web | Web (Nginx) | `registry.dungxbuif.com/mypocket-web:prod-2026.08.31` | 18083 | 80 |
| mypocket-worker | Background Worker | `registry.dungxbuif.com/mypocket-api:prod-2026.08.31` | — | — |
| mypocket-migrate | Migration (one-shot) | `registry.dungxbuif.com/mypocket-api:prod-2026.08.31` | — | — |

## Secrets

Store in `homelab/local_vars.json` → `MYPOCKET_SECRETS`. Values are never committed.

Reference: `~/production/mypocket/.env` on the Mac mini.