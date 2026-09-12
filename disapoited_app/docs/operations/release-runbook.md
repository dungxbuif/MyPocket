# MyPocket release runbook

This runbook deploys the existing production stack that owns `money.dungxbuif.com`. Do not create a second stack or point a restore drill at production.

## 1. Fail-closed prerequisites

1. Use the exact Git commit to be built; source inputs must be clean and `git diff --check` must pass.
2. Complete every physical iPhone/Safari row in `docs/work/completion/PHYSICAL-DEVICE-UAT.md` and set its status to `PASS` with device/build evidence.
3. Load the dedicated verification database URL, AI provider settings, OCR Platform settings and production secrets without printing them.
4. Run `scripts/verify-beta.sh`. It requires migration 0016, full backend race tests, frontend tests/build/E2E, Docusaurus build, provider settings, and backup/restore evidence.
5. Run `scripts/smoke-platform.sh` against the production-shaped local stack.

## 2. Immutable rollback point and backup

Record the current API/web image digests and current migration version. Create a new mode-0700 backup directory and run `scripts/backup-production.sh`, then restore it only into a `mypocket_restore_*` database and `restore-drill/*` object prefix using `scripts/restore-drill.sh`. Do not continue unless checksum, row-count and object-byte comparison pass.

Tag the verified source commit only after those gates. Record commit, image tags/digests, migration version, backup directory identifier and restore evidence in the release evidence file; never record credential values.

## 3. Build and publish

Build the API image from the repository root with `backend/Dockerfile`; it contains migrate/API/worker entrypoints. Build the web image from `frontend/Dockerfile`; it contains the PWA and Docusaurus output. Use immutable release tags, push, and resolve registry digests before editing production Compose.

The production Compose must bind S3 and every `AI_*`/`OCR_*` setting to both API and worker where required. `AI_ENABLED=true` requires HTTPS base URL, API key and model. `OCR_ENABLED=true` additionally requires AI enabled, S3 configured, HTTPS OCR base URL and OCR API key.

## 4. Deploy in compatibility order

1. Update `/Users/dungxbuif/production/mypocket/docker-compose.yml` to the exact tested image tags.
2. Pull images without stopping the current stack.
3. Run the one-shot migration service and verify schema version 0016.
4. Recreate API and wait for live/readiness checks.
5. Recreate worker and confirm it starts without config/lease/provider errors.
6. Recreate web last, after API and worker are compatible.
7. Keep the prior image digests and database backup immutable throughout smoke.

## 5. Authenticated production smoke

Run `scripts/smoke-production.sh https://money.dungxbuif.com` with a dedicated smoke-account API key plus `MYPOCKET_SMOKE_RECEIPT_PATH` and its exact `MYPOCKET_SMOKE_RECEIPT_CONTENT_TYPE`. The script checks TLS/public PWA/docs/OpenAPI, owner-scoped finance writes, exact transfer report exclusion, child-key revocation despite cache, export/download, real receipt upload, OCR-assisted Agent completion and reviewable draft creation. It rejects the draft and archives created finance entities on exit. Audit, Agent run, data-job and receipt metadata remain intentionally retained in the dedicated smoke account as release evidence because the public API has no unsafe hard-delete shortcut.

## 6. Rollback drill

If smoke fails, restore the previous API/worker/web image digests. Migration 0015/0016 only add Agent tables/links, so application-only rollback is expected to be schema-compatible; verify old live/readiness and authenticated reads. Never down-migrate production automatically.

If data restore is required, first restore the release backup into an isolated target and compare row counts/checksums again. A production restore requires an explicit maintenance window and operator approval.

After the rollback drill passes, deploy the verified release again in the same order and repeat full authenticated smoke. Publish URLs only with the final commit/image/migration/provider evidence.
