# Deployment

## API-only production mode — 2026-10-13

The production origin routes directly to the Go API. `mypocket-api` and
`mypocket-worker` remain 1/1; `mypocket-web` is intentionally 0/0. The mobile
base URL remains `https://money.dungxbuif.com/api/v1`; health is
`/api/v1/health` and Swagger is `/api/v1/docs/index.html`. No database,
schema, migration, credential, or financial data operation is part of this
topology change. Restore the web route/service only with explicit owner
approval, using the recorded rollback backup below.

## Release 1.2.1 rollout — 2026-10-20

The API and worker now run the immutable image
`registry.dungxbuif.com/mypocket-api:release-1.2.1-837bd505` at digest
`sha256:34226ef0700f1ea4ff05d2c3de4d0fc0488d28ef71f7b4c735b09172c3d1fd54`.
Both services converged to `1/1`; `mypocket-web` remains intentionally `0/0`.
The database advanced cleanly from migration 40 to 41. A verified backup was
created at `/tmp/mypocket-pre-1.2.1.dump`; pre/post wallet, transaction, budget
and category checksums were identical. Health returned HTTP 200 and the strict
readonly OpenAPI audit reported 91 live paths matching 91 local paths.

The native artifact `1.2.1 (4)` was codesigned with the owner's Personal Team,
installed on the paired iPhone, and launched. This records installation, not
complete physical interaction, Face ID, offline, APNs, or Mobile MCP/WDA
acceptance.

## Current production rollout — 2026-10-19

The approved release window deployed the immutable image
`registry.dungxbuif.com/mypocket-api:release-20261019-uat2` at digest
`sha256:17e3842326abb397d5184e6a5d407b51efb37ec05230171ab83128c7626cd6b7`
to both `mypocket-api` and `mypocket-worker`; each converged to `1/1`.
Migrations 035–037 advanced the database from version 34 to version 37 with
`dirty=false`. The validated pre-migration custom-format backup is retained at
`/tmp/mypocket-pre-release-20261019.dump` (SHA-256
`888bbacf594270512fde46889652d91e100335e2c21f7a99db2f903c4c7a2fc7`). No
financial rows were reset or deleted. At that release point the production
read-only route audit reported 88 live paths matching the then-deployed
88-path candidate; health and Swagger returned HTTP 200. The current checkout
has since added two portfolio trade-correction operations, so a fresh strict
probe is expected to report 88 live versus 90 local until the matching image
and migrations are rolled out. `mypocket-web` remains intentionally `0/0` in
API-only mode.

The isolated UAT owner passed the authenticated GET-only regression harness;
see [physical/API UAT evidence](IOS-PHYSICAL-UAT-20261019.md). Physical
iPhone signing/install, Face ID, APNs and offline-network acceptance remain
external gates and are not implied by this backend rollout.

## SQL diagnostics privacy — deployed 2026-10-19

The shared Go PostgreSQL adapter now keeps only statement class (SELECT/INSERT/
UPDATE/etc), elapsed time, rows and safe SQLSTATE, never full SQL text/parameters
or raw database error strings. Generic GORM message arguments are omitted.
This applies to API/worker/cleanup in the deployed release image. Real
PostgreSQL privacy tests and full Go race/vet pass locally; sampled production
worker logs contain only redacted statement classes and row counts.
Use request IDs, operation spans, source and SQLSTATE to debug. Do not enable raw
SQL/body/DSN logging to recover query values. Returned errors remain unchanged;
do not blindly print them elsewhere. Older logs may contain sensitive values;
retention/access review is separate, no logs were deleted in this work.
[Decision/evidence](../decisions/ADR-019-financial-sql-log-redaction.md).

## Current native-only auth rollout — 2026-10-14

> Historical record. This pre-rollout signing state was superseded by the
> 2026-10-19 Personal Team install and the current status documented above.

Owner approved additive backend deployment preserving existing data. Auth-only
commit `12572c87` is deployed as `mypocket-api:auth-12572c87` with digest
`sha256:1e1c686ce90bfd7c333dcc7f9206aff23117ed9f9d95839c8612a1a0fbc26fce`.
Validated pg_dump backup precedes clean forward migration29→31; financial
counts/checksums unchanged. API1/1/health200, worker1/1 on existing release,
web0/0 deliberately preserved. The API is independently reachable; do not
restart archived web merely to follow the older combined web/API process below.
No worker update is needed for auth routes. [Execution/backup/gates](../work/tickets/NATIVE-DEVICE-AUTH-ROLLOUT-20261014.md).

The current worktree has not produced a final signed iPhone build. An unsigned
arm64 device build (Debug and Release, signing disabled) compiles locally, but
the connected-device provisioning profile is absent, so no install or launch
against production is claimed. Owner hardware biometric/visual acceptance is
therefore still pending. This operational note is not an owner-approved release
announcement.

## Production target

| Surface | Target |
| --- | --- |
| API origin | `https://money.dungxbuif.com` |
| Edge | Traefik on the gateway host, routed through the homelab Swarm edge |
| Runtime node | Swarm node `vm100-prod` (`linux/amd64`) |
| Services | `mypocket-api` (1/1), `mypocket-worker` (1/1), `mypocket-web` (0/0) |
| Image registry | `registry.dungxbuif.com` |
| Database | PostgreSQL database `mypocket`, reached by the API through `pgbouncer` |

The API and worker images must be built for `linux/amd64` because production
services are constrained to `vm100-prod`. Production is API-only: the Traefik
host router forwards directly to `tasks.mypocket-api:8080`; the web source and
image are retained for rollback but the web service is scaled to zero. Use an
immutable tag containing the release commit; never deploy a mutable `latest`
tag.

The mobile app continues to use `https://money.dungxbuif.com/api/v1`. Health is
available at `/api/v1/health`; interactive API docs are at
`/api/v1/docs/index.html`. The origin root `/` intentionally does not serve a
web application.

### Local release-candidate preflight — 2026-10-15

The current worktree cross-compiles both backend binaries for the production
target (`GOOS=linux GOARCH=amd64 CGO_ENABLED=0`). These are local artifacts only
and have not been pushed or deployed:

| Artifact | SHA-256 |
| --- | --- |
| `/tmp/mypocket-api-linux-amd64-20261015` | `32a17424cbaf52a10b7e98443ea47e3e51cec42ecec223acf82484a3efa94849` |
| `/tmp/mypocket-worker-linux-amd64-20261015` | `78a27f40b10692119777ce45cf927bdac1ceba0e313b557258f1b05cd59dfb27` |

The binaries were verified as statically linked x86-64 ELF executables. A
registry image build, migration, service update and production smoke remain
separate owner-authorized gates.

Read-only production recheck before the jar hotfix on 2026-10-15 confirmed the
tasks still ran the previous images: API `auth-12572c87` at digest
`sha256:1e1c686ce90bfd7c333dcc7f9206aff23117ed9f9d95839c8612a1a0fbc26fce`
and worker `release-73c4c37` at digest
`sha256:d1f3f3b1bba9648192feccb814bd0823b9bc7df2c80518bf29a89cf77df691c2`.
The vm100 SSH account is a Swarm worker, so manager-level service inspection
and updates require the manager path; no such mutation was attempted.

### Historical v1.1.0 release rollout — 2026-10-15 (superseded)

The current release tag `v1.1.0` is deployed. API and worker both
run the immutable image
`registry.dungxbuif.com/mypocket-api:release-1.1.0-20261015-2212` at digest
`sha256:bef8040e7a372774234fd96d69e690215dd5d823f4520f901fd5cf1638fb37d2`.
The public health endpoint returned `200` and the public Swagger document
reports version `1.1.0`. No migration ran and the web service was deliberately
not redeployed for this native-focused release. Physical iPhone signing is
still blocked by the missing Xcode account/profile; the unsigned Release
iphoneos artifact is recorded in the release note.

### Historical async advisor and AI-entry context rollout — 2026-10-16 (superseded)

The API and worker were updated to immutable image
`registry.dungxbuif.com/mypocket-api:release-20261016-7846945` at digest
`sha256:71fdd4b4155972f3e6380ae02de846264e503347ab63b9693f65d2f0c713db43`.
The image includes the first-request async advisor cancellation endpoint and
the multi-image OCR context/order fix. No migration ran and no financial data
was modified. Post-rollout checks returned API health `200`, API/worker `1/1`,
and production Swagger exposes 78 paths including
`POST /api/v1/ai/advisor/messages/async` and
`GET /api/v1/sync/capabilities` (the latter remains authenticated).

The production API briefly remained at `0/1` while its health check warmed;
the task became healthy before the Swarm update converged. The web service
remains intentionally `0/0` for the native-focused rollout.

The OCR fix is context assembly only: validated image bytes still go through
the server-side OCR contract, and the model receives ordered, explicitly
delimited OCR documents. A configured OCR provider and owner-authenticated
AI-entry UAT are still required to certify recognition quality.

#### Advisor idle-reset migration catch-up — 2026-10-16

The first async advisor rollout exposed a worker-only schema error because the
image was updated before migrations 033/034 ran:
`column "last_user_message_at" does not exist`. Before changing the live
schema, a custom-format PostgreSQL backup was created and verified at
`/tmp/mypocket-pre-migration-20261016.dump` on `vm100`; its SHA-256 is
`60bd27c795f510854d9ca8c6e16a6285ab1a9ab2f46b3edde1b5a71e160d3a3d`.

The API image's `/app/migrate up` then completed successfully. The production
migration version is `34` with `dirty: false`; API and worker remain `1/1` and
the worker no longer emits the missing-column error on its next pass. The
migrations are additive: they add the idle timestamp/index and the reset
idempotency ledger; they do not delete financial data. For future advisor
rollouts, run the forward migration before relying on the worker expiry pass.

### AI jar proposal hotfix — 2026-10-15

The owner-requested feedback fix updated only `mypocket-api` to immutable image
`registry.dungxbuif.com/mypocket-api:ai-jar-fix-20261015-214833` at digest
`sha256:00a7a3533d345103f746f59afac136a3674b599d2d9b520dfbdd85c074e26759`.
No migration ran and no worker/web service changed. The API converged at `1/1`
on `vm100-prod`, and the public health endpoint returned `200`. The image
contains the non-mutating `propose_create_jar` tool and explicit confirmation
handbook. Keep feedback `824dea29-0a7b-45ca-976d-038616c12448` open until the
owner verifies the real model response and confirmation card.

The prompt-hardening follow-up deployed only `mypocket-api` as immutable image
`registry.dungxbuif.com/mypocket-api:ai-jar-capability-20261015-215712` at
digest `sha256:20fee27c5da418ecb23709484e9d8a844cbf1e9471592df74415ca0e831d6da0`.
It moves the supported capability statement before the generic read-only
boundary so OpenAI-compatible models are told they may prepare a jar proposal.
Health returned `200`; this image was superseded by the v1.1.0 API/worker
rollout above, with no migration or web update in that rollout.

## Release process

### Production rollout — 2026-10-20

The API-only production rollout completed from `main` commit `a2de756a3`.
The immutable image `registry.dungxbuif.com/mypocket-api:release-20261020-a2de756`
was applied to both `mypocket-api` and `mypocket-worker`; Swarm resolved it to
`sha256:5b54465e743f230f416172e87646548110d8b5e0470858c94167b5aaca1f3b45`.
`mypocket-web` remains intentionally `0/0`.

The owner explicitly approved skipping the usual backup. Migration advanced
from version `37 dirty=false` to `40 dirty=false`; no financial data was reset
or deleted. Public health returned `200`, and the strict read-only route audit
passed with `LIVE_PATHS=91` and `LOCAL_PATHS=91`.

A dedicated production UAT owner was created with credentials held out-of-band.
The authenticated GET-only regression passed across the populated core read
surfaces. MobileNext sees the physical iPhone online with MyPocket installed;
it remains on the app lock screen, so owner unlock is required before claiming
physical UAT/Face ID acceptance.

1. Run the local release gate: backend tests, frontend release gate, design
   checks, typecheck and production build.
2. Commit and push the release branch.
3. Build and push `mypocket-api:<release-tag>` and
   `mypocket-worker:<release-tag>` to the private registry on the amd64 build
   node. Run frontend design/typecheck/build gates locally when changing web
   source, but do not deploy the web service in API-only mode.
4. For a normal release, run the read-only [rollout preflight](ROLLOUT_READINESS.md),
   create and validate the protected backup, then run the forward migration
   command from the API image on `homelab-net`, `data-net` and `mypocket-net`
   before updating the API. The current candidate advances version 34 through
   migrations `035` → `036` → `037`; verify `dirty=false` and version `037`
   before enabling the matching worker.
5. Update the API and worker Swarm services with the immutable image tag and
   registry auth; the API uses `/app/api` and the worker uses
   `--entrypoint /app/worker`. Wait until API and worker each report `1/1`.
6. Verify `GET /api/v1/health` and `/api/v1/docs/index.html` return HTTP 200;
   verify `/` does not serve the web bundle.

Before opening the change window, run the GET-only route/version audit against
the intended origin:

```sh
AUDIT_BASE_URL=https://money.dungxbuif.com \
  ./scripts/mypocket-production-readonly-audit.sh
```

It reports live-versus-candidate OpenAPI path drift without authenticated
requests. During a staged rollout, candidate-only routes are expected. After
the API and worker converge, rerun with `REQUIRE_ROUTE_PARITY=1`; a non-zero
result is a release stop condition and must be reconciled before declaring the
candidate live.

The repository-only preflight is deterministic and never mutates production:

```sh
./scripts/mypocket-rollout-preflight.sh
```

It is not a substitute for the production backup, migration-version check,
service convergence or worker-log observation. Use the full
[`ROLLOUT_READINESS.md`](ROLLOUT_READINESS.md) checklist and retain the
read-only command evidence with the release record.

## API-only routing change — 2026-10-13

The public Traefik dynamic config routes `money.dungxbuif.com` directly to the
Go API. A timestamped rollback backup was created on the gateway host before
the route switch at
`/home/dungxbuif/edge/traefik/dynamic.yml.bak-mypocket-api-only-20261013`.
The API and worker remain 1/1; only `mypocket-web` was
scaled to 0. No database, schema, migration, credential, or financial data
operation was performed.

To restore the web UI, restore the saved gateway dynamic config (or set the
MoneyPocket upstream back to `http://tasks.mypocket-web:80`), then scale
`mypocket-web` to 1 replica on the Swarm manager. Verify the public root serves
the web app and API health remains 200.

### 2026-10-06 production rollout note

The API rollout on 2026-10-06 migrated the live database from version 27 to
version 29 after a validated `pg_dump -Fc` backup. The deployed candidate is
the immutable digest recorded in `docs/work/VALIDATION_MATRIX.md`; the public
health and OpenAPI endpoints were checked after the Swarm update.

The API and recurring worker now both run immutable image
`release-73c4c37`. A one-shot worker pass completed on `vm100-prod`, then the
long-running `mypocket-worker` service was scaled from `0/0` to `1/1` with the
production Redis overlay attached.

During the rollout, `redis-prod` resolved to its `data-net` address from the
API worker, but Redis was reachable on `homelab-net`. The production fix was
additive: an attachable overlay named `mypocket-redis-net` was created, the
existing `redis-prod` container was connected to it with the `redis-mypocket`
alias, and the API service was updated to use that alias. A `redis-cli ping`
probe scheduled on `vm100-prod` returned `PONG` before the API restart. Do not
replace this with a container IP. Preserve the overlay and alias when
recreating Redis, and run the same worker-local probe before changing the API.

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
docker service ps mypocket-worker
docker service ps mypocket-web
```

The production health endpoint is intentionally unauthenticated and only
returns service status. All financial endpoints remain protected by the
normal auth/API-key middleware. In API-only mode, `mypocket-web` is expected to
show `0/0` and `/` is expected not to return the web UI.

## Security follow-up — 2026-10-13

A diagnostic inspection during the rollout emitted production service
environment values into operator output. Treat affected production credentials
as exposed and rotate them in a controlled window, including database/cache,
session/signing, object storage, OAuth, model/OCR, feedback-agent and
push-notification credentials. No secret values are recorded in this
repository, and no credentials were changed during the API-only rollout.

### Web/API routing dependency (2026-10-13)

Do not scale `mypocket-web` to zero as a way to disable only the browser UI.
During the owner-requested native cutover check, `mypocket-api` remained at
`1/1`, but `https://money.dungxbuif.com/api/v1/health` returned HTTP 502 while
`mypocket-web` was at `0/0`. Restoring `mypocket-web` to `1/1` restored both the
web origin and API health endpoint. Keep the web service running until the
public API has an independently routed ingress; never stop the API or database
as part of an iOS client rollout.
