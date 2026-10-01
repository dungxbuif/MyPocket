# Deployment

## Production target

| Surface | Target |
| --- | --- |
| Web/API | `https://money.dungxbuif.com` |
| Edge | Traefik on the gateway host, routed through the homelab Swarm edge |
| Runtime node | Swarm node `vm100-prod` (`linux/amd64`) |
| Services | `mypocket-api`, `mypocket-web` |
| Image registry | `registry.dungxbuif.com` |
| Database | PostgreSQL database `mypocket`, reached by the API through `pgbouncer` |

The API and Web images must be built for `linux/amd64` because the production
service is constrained to `vm100-prod`. Use an immutable tag containing the
release commit; never deploy a mutable `latest` tag.

## Release process

1. Run the local release gate: backend tests, frontend release gate, design
   checks, typecheck and production build.
2. Commit and push the release branch.
3. Build and push `mypocket-api:<release-tag>` and
   `mypocket-web:<release-tag>` to the private registry on the amd64 build
   node. The frontend image build runs its own design/typecheck/build gate.
4. For a normal release, run the forward migration command from the API image
   on `homelab-net`, `data-net` and `mypocket-net` before updating the API.
5. Update both Swarm services with the immutable image tag and registry auth;
   wait until each service reports `1/1` running.
6. Verify `GET /api/v1/health` returns HTTP 200 and smoke-test the public web
   origin.

## Database reset (explicit owner approval only)

A production reset is destructive and is not part of a normal release. If an
owner explicitly requests a stage-v1 reset:

1. Scale `mypocket-api` and `mypocket-worker` to zero.
2. Create and validate a `pg_dump -Fc` backup of database `mypocket`.
3. Drop and recreate only the `mypocket` database, preserving the PostgreSQL
   cluster and all other databases.
4. Run `/app/migrate up` from the release API image and verify the migration
   version and table count.
5. Update the services and repeat the health checks above.

The stage-v1 production reset was performed on 2026-10-01 from commit
`d5c6ba8c`; the validated pre-reset backup is retained under the production
backup directory. Do not delete or overwrite that backup during cleanup.

## Rollback

- Roll back the Swarm service images to the previous immutable digests.
- For a schema/data incident, restore the validated backup to a disposable
  database first, verify it, then perform an owner-approved restore window.
- Never use `DROP DATABASE`, `TRUNCATE`, or a force migration as an automatic
  rollback.

## Verification

```sh
curl -fsS https://money.dungxbuif.com/api/v1/health
docker service ls
docker service ps mypocket-api
docker service ps mypocket-web
```

The production health endpoint is intentionally unauthenticated and only
returns service status. All financial endpoints remain protected by the
normal auth/API-key middleware.
