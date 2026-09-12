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
| Production | VM100 Docker Swarm worker (`10.10.0.31`) behind Orange Pi Traefik | `https://money.dungxbuif.com` |

## Secrets

Secrets are copied to VM100 `/opt/apps/mypocket/run/.env` (mode 600) and used
as the Swarm service env source. Do not commit `.env`.

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

## Build

```bash
cd ~/workspace/MyPocket

# Build API image (builds api + worker + migrate from one Dockerfile)
# Build context is the repo root (backend/ + migrations/ are inside)
docker build -f backend/Dockerfile -t homelab/mypocket-api:vm100-YYYYMMDD-<git-sha> .

# Build Web image (frontend only)
docker build -f frontend/Dockerfile -t homelab/mypocket-web:vm100-YYYYMMDD-<git-sha> frontend
```

## Deploy

```bash
# Run migrations (one-shot) on VM100
docker run --rm --env-file /opt/apps/mypocket/run/.env \
  --entrypoint /app/migrate \
  homelab/mypocket-api:vm100-YYYYMMDD-<git-sha>

# Start/update services from the Pi5 Swarm manager
docker service create/update mypocket-api
docker service create/update mypocket-worker
docker service create/update mypocket-web

# Verify
curl https://money.dungxbuif.com/api/v1/health/live   # -> 200
curl https://money.dungxbuif.com/api/v1/health/ready  # -> 200
curl https://money.dungxbuif.com/                     # -> 200
```

## Traefik Route (Orange Pi `10.10.0.2`)

Orange Pi Traefik uses the file provider at `/opt/edge/traefik/dynamic.yml`.
Route MyPocket to the Swarm DNS service:

```yaml
money:
  loadBalancer:
    passHostHeader: true
    servers:
      - url: http://mypocket-web:80
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
