---
artifact_type: deployment_guide
id: DEPLOYMENT
status: active
owner: shared
updated: 2026-09-01
---

# Deployment

## Deployment Targets

| Environment | Target | URL |
| --- | --- | --- |
| Production | Mac mini Docker Compose (`10.10.0.10`) | `https://money.dungxbuif.com` |

## Secrets

Secrets are stored in `homelab/local_vars.json` under `MYPOCKET_SECRETS` and copied to `~/production/mypocket/.env` (mode 600). Do not commit `.env`.

## Pi5 Dependencies Setup

### PostgreSQL

```sql
CREATE USER mypocket WITH PASSWORD '<password>';
CREATE DATABASE mypocket OWNER mypocket ENCODING 'UTF8';
ALTER SCHEMA public OWNER TO mypocket;
```

Run via `ssh pi docker exec postgres psql -U admin -d postgres -c "..."`

### S3 (RustFS)

Create bucket + dedicated user scoped to `mypocket` bucket:

```bash
mc alias set rustfs-admin https://storage.dungxbuif.com admin <RUSTFS_ADMIN_SECRET>
mc mb rustfs-admin/mypocket

# Create policy for mypocket bucket only
mc admin policy create rustfs-admin mypocket-policy /path/to/policy.json
mc admin user add rustfs-admin mypocket <password>
mc admin policy attach rustfs-admin mypocket-policy --user mypocket
mc admin user svcacct add rustfs-admin mypocket  # generates access/secret key
```

Policy JSON (scoped to `mypocket` bucket only):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetBucketLocation", "s3:ListBucket", "s3:ListBucketMultipartUploads"],
      "Resource": ["arn:aws:s3:::mypocket"]
    },
    {
      "Effect": "Allow",
      "Action": ["s3:AbortMultipartUpload", "s3:DeleteObject", "s3:GetObject", "s3:ListMultipartUploadParts", "s3:PutObject"],
      "Resource": ["arn:aws:s3:::mypocket/*"]
    }
  ]
}
```

## Build & Push

```bash
cd ~/workspace/mypocket

# Tag and push API image (builds api + worker + migrate from one Dockerfile)
# Build context is the repo root (backend/ + migrations/ are inside)
docker build -f backend/Dockerfile -t registry.dungxbuif.com/mypocket-api:prod-YYYY.MM.DD .
docker push registry.dungxbuif.com/mypocket-api:prod-YYYY.MM.DD

# Tag and push Web image (frontend only, includes Docusaurus docs)
docker build -f frontend/Dockerfile -t registry.dungxbuif.com/mypocket-web:prod-YYYY.MM.DD ~/workspace/mypocket/frontend
docker push registry.dungxbuif.com/mypocket-web:prod-YYYY.MM.DD
```

## Deploy

```bash
cd ~/production/mypocket

# Pull latest images
docker compose pull

# Run migrations (one-shot)
docker compose run --rm migrate

# Start services
docker compose up -d

# Verify
curl http://localhost:18082/api/v1/health/live    # → 200
curl http://localhost:18082/api/v1/health/ready   # → 200
curl http://localhost:18083/api/v1/health/live    # → 200 (web proxy)
```

## Caddy Route (Orange Pi `10.10.0.2`)

Add to `/opt/edge/caddy/Caddyfile` (NOT `/ssd-data/infra/Caddyfile` on Pi5 — the running Caddy is on orange-pi via Swarm):

```
@money host money.dungxbuif.com
handle @money {
    reverse_proxy 10.10.0.10:18083
}
```

Then reload:

```bash
CADDY_ID=$(docker ps --format '{{.ID}}' -f name=edge_caddy)
docker exec $CADDY_ID caddy reload --config /etc/caddy/Caddyfile
```

## VPS Traefik Whitelist

The VPS (`103.82.21.202`) uses **Traefik**, not Nginx. Add to `/root/gateway/dynamic.yml` in the `HostSNI` rule:

```yml
rule: "HostSNI(`...`) || HostSNI(`money.dungxbuif.com`)"
```

No restart needed — Traefik watches for file changes automatically.

## Google OAuth

Add `https://money.dungxbuif.com/api/v1/auth/google/callback` to Authorized redirect URIs in Google Cloud Console.

## Rollback

```bash
cd ~/production/mypocket
docker compose down

# Remove Caddy route from /opt/edge/caddy/Caddyfile, reload
# Remove money.dungxbuif.com from VPS /root/gateway/dynamic.yml
```