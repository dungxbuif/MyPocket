# R0 web and Docusaurus release — 2026-09-09

Status: in_review for authenticated docs/device acceptance. Owner authorization: “Deploy và cho tôi link docs”, then “deploy laij ddiu”. Redeployment, public HTTPS smoke checks and registry publication completed at 21:25 Vietnam on 2026-09-09 (see verification). This authorizes this bounded deployment, not full R0 acceptance or physical Safari/PWA sign-off.

Trace: [R0](R0-DETAIL_DESIGN.md), [TICKET-028](../tickets/TICKET-028-production-hardening-debug-audit-foundation.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md), [verification](../test-verification/R0-WEB-DOCS-RELEASE-2026-09-09.md).

## Scope and approach

Context: repository gateway/standards, current context/backlog, prior search/receipt proof, Docusaurus sources, frontend/backend Dockerfiles and exact production compose inspected. Publish the tested current frontend and bounded Docusaurus updates covering recent behavior and remaining limits. Correct canonical docs URL to the observed production URL `https://money.dungxbuif.com`.

Production compose: `/Users/dungxbuif/production/mypocket/docker-compose.yml`. Only web/API image references change. Build web normally. Build a docs-only API derivative using `backend/Dockerfile.docs-release` with the current API base digest `registry.dungxbuif.com/mypocket-api@sha256:869b518de8f61b357388f5163d952eb39e04c635a2d7d85145f76999652722fd`. Compare api/worker/migrate checksums before replacing the API container. Keep worker and migrate image tags at `prod-2026.09.08.4`; use compose `up -d --no-deps api web` to avoid running migration.

Rejected alternative: building the normal backend Dockerfile would also deploy unrelated dirty backend changes. A docs-only derivative preserves existing server/auth behavior. No secrets, environment values, volumes, ports, DB schema or data are changed. No new permanent architecture/auth policy; this is a bounded packaging recipe, not an ADR-level system redesign.

## Acceptance and rollback

Run frontend tests, web image build/typecheck, Docusaurus build; check generated pages/links. Verify API executable checksums, inherited user/entrypoint, image registry push, container health, web asset identity and docs authentication redirect/content availability. Public HTTPS probe failure must be reported separately from local container health; do not weaken TLS/auth to force a pass. Do not create test records or forge sessions in production.

Rollback: restore only web image `registry.dungxbuif.com/mypocket-web:prod-2026.09.08.6` and API image `registry.dungxbuif.com/mypocket-api:prod-2026.09.08.4`, then run the same no-deps compose command. Original images remain local. No DB rollback is needed because no migration is run.

## Reconciliation

Update Docusaurus intro, search/auth behavior and release limits; retain prior API audit edits. Requirements, API implementation, ERD, auth/security policies and ADRs unchanged. Record platform proof in the linked verification report and refresh context/backlog/validation/changelog. Physical iPhone receipt/PWA UAT remains open; this deployment must not mark it passed.
