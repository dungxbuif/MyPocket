---
artifact_type: deployment_work_note
status: in_progress
approval: owner explicitly approved additive production backend deployment, preserving data
---

# Remembered-device auth rollout

Links: [design](NATIVE-FEEDBACK-CONSISTENCY-AUTH-DETAIL_DESIGN.md),
[auth contract](../../architecture/AUTH_API.md),
[Face ID guide](../../guides/face-id.md), [validation](../VALIDATION_MATRIX.md).

## Scope and safety

Create a reproducible auth-only candidate from deployed source baseline plus
device auth routes, repository, entity, session-family logic and their tests.
Do not include unrelated uncommitted finance/web/UI changes. An isolated Git
index/candidate branch preserves the active checkout, index and branch state.
Regenerate candidate Swagger from that candidate, not the dirty full checkout.
No changelog/release note publication before owner review.

Migration 30 only adds category is_active default true; migration 31 adds
device_credentials. Include 30 to keep the forward migration chain complete;
neither resets or modifies financial history. Before migration, create and
validate a custom-format pg_dump backup, record migration version/dirty flag
and counts of financial tables. Apply forward migrations only, compare counts,
then update only the API to an immutable image digest. Existing worker/web
service state remains unchanged. Never force, drop, truncate or reset prod.

Preflight: API health 200, API/worker 1/1, web 0/0. Current public API is reachable
without the web service; preserve this observed state rather than restarting
web from an older runbook assumption.

## Verification

- Independent auth review found legacy-family and delayed-issuance revocation
  races. Reproduced by deterministic TTL-aware tests; fixed with stable legacy
  migration groups and post-issuance revocation check. Full disposable-DB Go
  race suite passes. First logout of a legacy session can sign out the owner's
  other pre-upgrade sessions once; fresh password/device families are separate.
- Test candidate itself against disposable PostgreSQL before image build.
- Validate pg_dump catalog; retain backup, do not print/store secrets in docs.
- Migration dirty=false, unchanged financial counts, API/worker health, new
  public OpenAPI auth paths and invalid-device rejection after rollout.
- Build signed native app for owner's paired iPhone using current Personal Team
  provisioning, production API URL and no test launch flags. Verify install and
  launch through devicectl. Owner performs physical Face ID acceptance.
- Keep production feedback status unchanged until owner accepts fixes/release.

## Execution record

Pending candidate, backup, migration, image rollout and physical installation.
