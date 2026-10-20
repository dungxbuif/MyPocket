---
artifact_type: verification_note
id: API-PARITY-AUTH-AUDIT-20261020
status: in_progress
owner: shared
updated: 2026-10-20
---

# Production API route and auth parity audit

## Scope

Verify that the production OpenAPI document and the checked-in candidate agree
on path, HTTP operation, and authentication metadata. This is read-only: it
does not send an authenticated request, mutate feedback, or deploy anything.

## Decision

Extend `scripts/mypocket-production-readonly-audit.sh` so strict parity mode
fails not only for missing paths or methods, but also when a shared operation
changes from authenticated to public (or the reverse). A route can be present
in both documents while its authorization contract is still unsafe to ship.

## Verification plan

```text
AUDIT_BASE_URL=https://money.dungxbuif.com \
REQUIRE_ROUTE_PARITY=1 scripts/mypocket-production-readonly-audit.sh
```

Expected evidence for a fully rolled-out candidate: health `200`, identical
path/operation sets, and no `SECURITY_DRIFT` rows. The script remains
read-only and reports live/local OpenAPI versions and counts. A non-zero
strict result is a release stop condition.

## Reconciliation

- `docs/architecture/API.md`: record the stronger route/auth parity gate and
  the 2026-10-20 production result.
- `docs/work/VALIDATION_MATRIX.md`: record the command and result.
- No API, database, authentication, or runtime behavior changes are made by
  this audit.

## Current read-only result — 2026-10-20

The production GET-only probe returned health `200`, live/local OpenAPI
version `1.1.0`, and `88` live paths versus `90` local candidate paths.
Strict parity failed as designed. The two candidate-only operations are:

```text
/api/v1/portfolio/assets/{id}/trade-corrections
/api/v1/portfolio/assets/{id}/trade-corrections/preview
```

No `SECURITY_DRIFT` rows were reported for shared operations. This is staged
release drift until the matching API/worker image and migrations `000038`–
`000040` are explicitly rolled out. No production mutation, database
connection, SSH session, or authenticated request was made.

