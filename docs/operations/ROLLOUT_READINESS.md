---
artifact_type: operational_runbook
id: ROLLOUT_READINESS
status: deployed-uat-2026-10-19
owner: shared
trace:
  database: ../architecture/DATABASE.md
  deployment: DEPLOYMENT.md
  api: ../architecture/API.md
  audit_contract: ../work/tickets/FEEDBACK-20261019-AUDIT-DELIVERY-DETAIL_DESIGN.md
  validation: ../work/VALIDATION_MATRIX.md
---

# Release rollout readiness

This runbook is for the current native/API candidate. The repository preflight
commands remain read-only, while the approved 2026-10-19 change window is
recorded in the current-status section below. No preflight command applies a
migration, updates a Swarm service, or writes to a database.

## Current schema path

### Release 1.2.1 — 2026-10-20

Production currently uses release-20261020-a2de756 and schema 40. The new image
requires additive migration 000041 before API/worker update. Its deadline
default preserves old API inserts during transition. Back up and compare
financial checksums. Keep runtime/grounding flags off for this patch; broader
quality acceptance remains open. The preflight script now requires 000041.
Older upgrade paths below are historical context.

### MCP local candidate — 2026-10-20

The MCP adapter itself introduces no migration: it reuses current owner,
receipt, CAS and audit-outbox infrastructure. Its aggregate verification uses
disposable loopback PostgreSQL 17/Redis 7 migrated through clean schema 40;
this is not evidence of a staging or production schema change. Current local
OpenAPI has 91 paths, including authenticated GET/POST/DELETE `/api/v1/mcp`;
the route/catalog gate covers 101 authenticated tools and 19 explicit excluded
method/path entries through `NewApplicationRouter`.

Before release, run owner-authenticated staging checks with both JWT and an
API key: official SDK initialize/list/read/write, revoked-key 401, stable-key
replay and stale-version 409, AI-entry multipart recovery, advisor async
poll/replay/cancel and redacted audit delivery. Confirm the selected immutable
API/worker image and required schema together with the existing release gates
below. No staging endpoint, production service, credentials or financial rows
were accessed for MCP local verification; no deployment is claimed.
[Evidence](../work/VALIDATION_MATRIX.md),
[integration contract](../ai/mcp-api-parity.md).

The candidate adds six ordered migrations after the deployed version-34
baseline:

```text
000035_budget_series  ->  000036_push_tokens  ->  000037_audit_outbox
  -> 000038_portfolio_trade_corrections -> 000039_push_deliveries
  -> 000040_row_versions
```

Run the forward migration command once from the release image; the migrate
ledger makes a second invocation a no-op. Never skip a version, rerun the
stage-v1 baseline, or use `force` to repair an ordinary rollout. If the target
reports `dirty=true`, stop and investigate the failed migration before any
retry. `000036` stores only owner/device registration metadata. `000037` stores
redacted audit payloads for bounded Redis delivery. `000038` preserves existing
portfolio trades while adding correction revisions, `000039` stores only an
owner-scoped APNs hand-off, and `000040` initializes transaction/category
versions to `1`; none of these migrations changes existing financial values.

## Local deterministic preflight

From the repository root:

```sh
./scripts/mypocket-rollout-preflight.sh
```

This checks migration naming/order, requires 035–040, verifies the API
image contains both `/app/migrate` and `/app/worker`, and prints (without
executing) the operator commands. It warns when no backup is supplied. To
validate an already-created custom-format archive without exposing its
contents:

```sh
BACKUP_FILE=/secure/path/mypocket-pre-release.dump \
  ./scripts/mypocket-rollout-preflight.sh
```

To read a target database's migration version, explicitly opt in and keep the
DSN out of shell history/log output:

```sh
DATABASE_URL="$DATABASE_URL" CHECK_DATABASE_VERSION=1 \
  ./scripts/mypocket-rollout-preflight.sh
```

The command runs only `migrate version`; it does not run `up`.

The migration-order regression is also part of the backend suite:

```sh
cd backend
go test ./migrations -count=1
```

It proves strict ordering, latest version 040, reversible disposable down
files for 036–040, required owner/uniqueness/CAS constraints, and the absence
of destructive statements in the additive up files.

## Native release-artifact preflight

After a Release build, inspect the resolved `.app` rather than relying only on
source settings or build-log output:

```sh
APP_PATH=/tmp/mypocket-release-artifact-preflight-20261019/Build/Products/Release-iphonesimulator/MyPocketIOS.app \
  mobile-ios/scripts/release-artifact-preflight.sh
```

This is read-only. It verifies bundle `com.mypocket.ios`, marketing version
`1.1.0`, build `2`, the production HTTPS API URL, portrait-only orientation,
and absence of obvious secret-like files inside the packaged app. It does not
sign, install, contact production, or prove a physical-device release. The
deterministic fixture checks are run with:

```sh
mobile-ios/scripts/release-artifact-preflight.test.sh
```

## Production change-window gates

An operator must record each result before proceeding. Do not print secrets,
DSNs, tokens, backup contents, or financial row data.

1. **Release identity:** record the immutable API/worker image tag and digest;
   confirm the image was built for `linux/amd64` and includes the same source
   and migration directory.
2. **Service state:** read `docker service ps mypocket-api --no-trunc` and
   `docker service ps mypocket-worker --no-trunc`; do not update either service
   while the backup is being created.
3. **Database preflight:** run the release image's `/app/migrate version`.
   Require the expected current version, `dirty=false`, and no unexpected
   database target. A version-34 target should advance through 035–040; a
   target already at 35–39 advances only the remaining suffix. Do not start
   clients that send `expected_version` until 040 is clean.
4. **Backup:** create a timestamped `pg_dump -Fc` of `mypocket`, record its
   SHA-256, and validate it with `pg_restore --list`. Retain it in the
   protected production backup directory. The backup is the recovery path;
   migration down files are for disposable fixtures only.
5. **Apply schema:** from the release API image, with the target networks and
   environment already attached, run `/app/migrate up`. Read the version again
   and require `dirty=false` and version 040. If it fails, stop; do not force
   the ledger or start a worker against a dirty schema.
6. **API rollout:** update the API service to the immutable digest, wait for
   `1/1`, then check `curl -fsS https://money.dungxbuif.com/api/v1/health` and
   `/api/v1/docs/index.html` (HTTP 200). Confirm the OpenAPI document contains
   the device registration routes and the current assistant routes.
7. **Worker rollout:** update the worker to the same digest and wait for
   `1/1`. Read at least two configured worker intervals of logs. Require no
   missing-table/column errors, no migration-dirty errors, and successful
   completion of recurring, budget, notification, advisor-expiry and audit
   passes. A worker has no HTTP health endpoint; service replica state plus
   bounded pass logs are the readiness signal.
8. **Audit outbox:** follow the [audit outbox runbook](AUDIT_OUTBOX_RUNBOOK.md)
   to inspect pending/delivered/failed counts without reading payloads. A Redis
   outage may leave pending rows, but rows must not be silently deleted.
   Confirm retry timestamps advance and delivered rows are marked only after
   Redis publish. Retention/dead-letter alerting remains an operational gate
   before treating this as a compliance archive.
9. **Native/API acceptance:** with an owner-authenticated test account, verify
   device registration/revocation and in-app notification read state. Verify
   owner isolation and do not claim APNs delivery until Apple entitlements,
   provider configuration and a signed-device delivery test pass.

## Read-only commands after rollout

Before any change window, compare the public API's health/OpenAPI surface with
the checked-out candidate. This is GET-only and does not require an auth token:

```sh
AUDIT_BASE_URL=https://money.dungxbuif.com \
  ./scripts/mypocket-production-readonly-audit.sh
```

The audit prints the live/candidate OpenAPI versions and path counts, then
lists route and HTTP-method drift without printing schemas, credentials, or
financial data. It is the release guard for the ten candidate-only route
contracts; it also detects a new operation missing from a path that already
exists (for example a candidate `PATCH` on a deployed asset path).
Use `REQUIRE_ROUTE_PARITY=1` only when the release is expected to be fully
deployed; it exits non-zero when production is missing a candidate route or
operation. A candidate-only route is expected during a staged rollout and is
not by itself evidence of an outage.

The default command is observational. Once the matching API image is expected
to be live, enable the strict gate:

```sh
AUDIT_BASE_URL=https://money.dungxbuif.com \
REQUIRE_ROUTE_PARITY=1 \
  ./scripts/mypocket-production-readonly-audit.sh
```

The strict gate performs only the public health/OpenAPI GETs; it does not
authenticate, migrate, write feedback, or alter financial data. Local route
registration, checked-in OpenAPI, and authenticated owner/API-key behavior for
the ten candidate paths are covered by
`backend/internal/controller/http/route_contract_parity_test.go` plus the
route-specific integration tests.

```sh
curl -fsS https://money.dungxbuif.com/api/v1/health
curl -fsS -o /dev/null -w '%{http_code}\n' \
  https://money.dungxbuif.com/api/v1/docs/index.html
docker service ps mypocket-api --no-trunc
docker service ps mypocket-worker --no-trunc
docker service logs --since 10m mypocket-worker
```

The worker can be run once only in a staging/approved window:

```sh
WORKER_ONCE=true /app/worker
```

This is not a passive health check: it may materialize due recurring/budget
rows, so it is intentionally not part of the default preflight script.

## Rollback and stop conditions

- If the backup cannot be validated, the migration is dirty, or API/worker
  replicas do not converge, stop the rollout.
- If the schema is healthy but the new worker has a runtime defect, roll back
  the API/worker images to the previous immutable digest while preserving the
  additive schema. Older code must not be assumed to understand new routes;
  verify health and logs after rollback.
- For a schema/data incident, restore the validated backup to a disposable
  database first, verify it, then schedule an owner-approved production
  restore. Do not run `DROP DATABASE`, `TRUNCATE`, or migration `force` as an
  automatic rollback. The 036/037 down files are only for disposable local
  fixtures and would discard their new tables.

## Current status

### Read-only rollout dry-run — 2026-10-20

The current candidate was built for `linux/amd64` and pushed as an immutable
registry image (the push does not update any Swarm service):

```text
registry.dungxbuif.com/mypocket-api:release-20261020-4e44d364946a
sha256:28e4d2450df20f205e38d82aa923fb6fa86df46dd25ff695a1f4579098983807
```

It contains `/app/api`, `/app/migrate`, `/app/worker` and migrations through
`000040`. Service update remains blocked until the validated production backup
and manager-path migration version readback are available.

The production-readiness checks were rerun without an authenticated database
connection, SSH manager access, service updates, migrations, or worker passes:

```text
scripts/mypocket-rollout-preflight.test.sh                         PASS
go test ./migrations ./cmd/worker ./internal/infrastructure/repository -count=1  PASS
go build ./cmd/api ./cmd/worker ./cmd/migrate                         PASS
docker build --platform linux/amd64 -f backend/Dockerfile            PASS
AUDIT_BASE_URL=https://money.dungxbuif.com REQUIRE_ROUTE_PARITY=1      STOP (88 live / 90 local; 2 candidate-only routes)
```

The locally built image contains executable `/app/api`, `/app/migrate` and
`/app/worker`, plus the ordered `000035`–`000040` migration files. The public
health check is green, but strict OpenAPI parity is stopped because production
does not yet expose the two candidate portfolio correction routes. The audit
cannot prove the target database version
or worker service state. The latest recorded production state remains version
37 (`dirty=false`) from the approved 2026-10-19 rollout; therefore migration
`000038`–`000040`, including transaction/category row versions, still require
the documented backup, manager-path migration check and post-rollout worker
observation before claiming production readiness. No production mutation was
attempted in this dry-run.

Production rollout completed on 2026-10-19 after an explicit approved change
window. A validated custom-format backup was created before migration:

```text
backup: /tmp/mypocket-pre-release-20261019.dump
sha256: 888bbacf594270512fde46889652d91e100335e2c21f7a99db2f903c4c7a2fc7
```

The immutable image `registry.dungxbuif.com/mypocket-api:release-20261019-uat2`
(`sha256:17e3842326abb397d5184e6a5d407b51efb37ec05230171ab83128c7626cd6b7`)
was applied to both `mypocket-api` and `mypocket-worker`; each converged to
`1/1`. The database had already advanced from version 34 to version 37 with
`dirty=false`. No financial data was reset or deleted. The public strict route
audit passed for the then-deployed 88-path candidate; the current checkout now
contains 90 paths because of two portfolio correction operations. A fresh
strict probe correctly stops at 88 live versus 90 local; health/docs still
return successfully.

After the provider/OCR diagnostic hardening was included in `uat2`, health,
strict route parity and the authenticated GET-only UAT regression were rerun
successfully. The rerun log is
`/tmp/mypocket-production-uat-regression-20261019-rerun.log`.

### Regression recheck — 2026-10-19

The read-only gates were rerun after the current worktree changes:

```text
AUDIT_BASE_URL=https://money.dungxbuif.com REQUIRE_ROUTE_PARITY=1 \
  scripts/mypocket-production-readonly-audit.sh
  PASS: health status=ok; LIVE_PATHS=88; LOCAL_PATHS=90
  CANDIDATE_ONLY_PATHS: portfolio trade-corrections and preview
  FAIL: production OpenAPI route/method parity is incomplete

UAT_BASE_URL=https://money.dungxbuif.com UAT_MONTH=2026-10 \
  scripts/mypocket-production-uat-regression.sh
  all attempted authenticated GET checks passed; no production mutation
```

The local provider/OCR/AI focused packages and the disposable-database race
suite also passed. The full race command completed successfully for every Go
package (`go test -race ./... -count=1`); credentials were read from the
disposable test container and were not printed or committed.

The production UAT owner is isolated from existing users and has only seeded
UAT wallets, one UAT category, one transaction and one internal transfer. Its
credentials are delivered out-of-band to the owner; they are intentionally not
stored in repository documentation.

An authenticated GET-only regression run for that owner passed on
2026-10-19. Use `scripts/mypocket-production-uat-regression.sh` with
`UAT_BASE_URL`, `UAT_EMAIL`, `UAT_PASSWORD` and (optionally) `UAT_MONTH`; it
does not create, update or delete financial data. The run covered all
populated core read surfaces (profile/home, ledger, reports/months, jars,
recurring, travel, portfolio, notifications, feedback/changelog, sync,
advisor/AI capabilities and budget-series). Parameterized credit, jar and
budget report reads run only when the UAT owner already has matching fixtures.
Conversation history requires an explicit `UAT_CONVERSATION_ID` and was not
claimed in the baseline run. Evidence was written to
`/tmp/mypocket-production-uat-regression-20261019.log`.

Remaining release gates are external: APNs delivery/provider configuration,
physical iPhone Release signing/profile, human Face ID interaction, and a
physical offline-network fault test. These are not represented as passed by
the API rollout.

The route-drift audit is implemented by
`scripts/mypocket-production-readonly-audit.sh`; it performs only GET requests
and is safe to run before a backup or deployment. A non-zero result from
`REQUIRE_ROUTE_PARITY=1` is a release stop condition, not permission to mutate
production.

### Historical pre-rollout audit — 2026-10-19 (superseded)

The following snapshot records the intentional drift observed before the
`release-20261019-uat2` change window. It must not be used as the current
production status; see **Current status** above for the post-rollout 88/88
parity result.

The public API was checked without authentication or mutation:

- `GET https://money.dungxbuif.com/api/v1/health` returned `200` with
  `{"data":{"status":"ok"}}`.
- Swagger UI returned `200` from `/docs/` and resolved to
  `/api/v1/docs/index.html`; the document at `/api/v1/docs/doc.json` returned
  `200`, version `1.1.0`, with 78 paths.
- The current source candidate exposes 88 paths. Production is missing the
  following candidate routes: advisor event replay, budget-series CRUD/stop,
  budget report/transaction views, iOS push-token register/revoke, and savings
  report/transaction views. This is evidence that the candidate API has not
  been deployed; it is not evidence of a public outage.

Read-only access to `vm100` confirmed the API and worker are healthy/running on
the previous immutable image `release-20261016-7846945` (digest retained in
the deployment record). `/app/migrate version` reported `34 dirty: false`.
The worker produced SQL health probes and no error/missing-schema lines in the
sampled window, but the old image cannot run migrations 035–040 or the new
worker behavior. Manager-level service inspection remains unavailable from the
worker node; the rollout therefore still requires the manager path, a
validated backup, migration through 037, matching API/worker image rollout,
and post-rollout worker observation.

### Historical release/device recheck — 2026-10-19 22:38 +0700 (superseded)

This recheck was read-only with respect to production and financial data:

- `GET https://money.dungxbuif.com/api/v1/health` returned `200`.
- Production Swagger remains version `1.1.0` with 78 paths. The local
  candidate remains version `1.1.0` with 88 paths; the ten candidate-only
  routes are the advisor event replay, budget-series/report, push-token and
  savings-report routes listed above. No API rollout was attempted.
- `scripts/mypocket-rollout-preflight.sh` passed migration ordering through
  `000037`; it warned because `BACKUP_FILE` was not supplied and skipped the
  target database version check. This is an intentional rollout stop condition,
  not a production failure.
- `mobile-ios/scripts/verify-project.sh`,
  `mobile-ios/scripts/check-base-components.sh`, and `git diff --check` passed.
- `mobile-ios/scripts/release-artifact-preflight.test.sh` passed, and the fresh
  warnings-as-errors Release simulator artifact passed
  `mobile-ios/scripts/release-artifact-preflight.sh` (bundle `com.mypocket.ios`,
  version `1.1.0`, build `2`, production HTTPS URL, portrait-only).
- A fresh unsigned Release simulator build with
  `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` passed at
  `/tmp/mypocket-release-artifact-preflight-20261019/Build/Products/Release-iphonesimulator/MyPocketIOS.app`.
- The connected iPhone is paired and available, but still has `MyPocket 1.1.0
  (2)`. A signed device build remains blocked by Xcode reporting no account for
  team `GYRA8D4M54` and no development profile for `com.mypocket.ios`. The
  existing app was not replaced.

Next release action is to sign in to the owning Apple account in Xcode (or
provide a valid development/ad-hoc profile), then rerun the physical-device
build/install. Do not treat the simulator artifact as a signed-device release.
