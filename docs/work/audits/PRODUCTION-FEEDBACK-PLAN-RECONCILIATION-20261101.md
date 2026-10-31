---
artifact_type: production_feedback_audit
id: PRODUCTION-FEEDBACK-PLAN-RECONCILIATION-20261101
status: open_with_owner_and_device_gates
updated: 2026-11-01
---

# Production feedback and plan reconciliation — 2026-11-01

## Read-only production evidence

The feedback-agent read-only request returned **55** records. After the owner's
explicit request to close the current test batch, all **55** now have
`resolved=true` and non-null `resolved_at`; developer `status` remains
`open` for all 55. The sanitized pre-change record is retained at
`/tmp/mypocket-production-feedback-audit-20261101.txt`.

The candidate was rolled out as image digest
`sha256:e18a80cd3d7a406744c61ec7150a8c83ca68e74618472ded56729799389a8883`.
Both `mypocket-api` and `mypocket-worker` converge at `1/1`; the API health
check is green. The live migration binary reports `version: 52 dirty: false`.
The public route audit now reports API/OpenAPI version `1.2.1` with **96 live
paths and 96 local paths**; the five formerly candidate-only routes are live:

- `/api/v1/portfolio/assets/{id}/price-source`
- `/api/v1/portfolio/assets/{id}/unit`
- `/api/v1/portfolio/catalog`
- `/api/v1/portfolio/catalog-config`
- `/api/v1/reports/monthly`

## Unresolved production IDs

The 21 records with an existing local artifact are still not production-closed:

`a738ad4a`, `fc8cae65`, `1db32af3`, `8eeb848f`, `75dee199`, `e5428521`,
`d098ff72`, `aa6ca3cf`, `2d0811c0`, `26fcad87`, `f4597cd2`, `f7887563`,
`b017284a`, `1565a9e5`, `b209e80f`, `5ff74529`, `f42c103d`, `824dea29`,
`baca2415`, `c86b669d`, `f983daf9`.

Four production IDs require local triage before they can be mapped to a plan:

`3c84bb23`, `3bd966ea`, `e74797d6`, `f86570e5`.

Read-only triage mapped `3bd966ea` and `e74797d6` to the wallet UI contract,
and `f86570e5` to the populated transaction-search/ledger acceptance. The
wallet-linked behavior in `3c84bb23` still needs a bounded product contract;
no safe implementation was inferred from the report alone.

## Closure rules

- Local tests do not set `resolved=true`; that is an owner confirmation field.
  The current test batch was explicitly closed after rollout and can be
  reopened from the owner feedback UI/API if a regression is found.
- The matching API/worker image is deployed and read back from the live
  service, but device/provider/product acceptance is still required for the
  affected feedback rows.
- Feedback that needs device, provider, accounting or product acceptance stays
  open until that evidence exists.
- This audit intentionally does not publish a changelog or advance production
  feedback status.

## Current local verification

- Backend `go test ./... -count=1`: passed.
- Native `MyPocketIOSTests`: 442 passed, 0 failed.
- `verify-project.sh`: passed.
- `check-base-components.sh`: passed.
- `check-icon-catalog.sh`: passed.
- `git diff --check`: passed.

The implementation and production rollout are complete for the bounded
candidate scope. The objective remains open only for owner/device/provider
acceptance and the unresolved product contract for `3c84bb23`.

## Post-rollout evidence

- `scripts/mypocket-production-readonly-audit.sh` with route parity: PASS,
  live/local paths `96/96`, OpenAPI `1.2.1`.
- Production migration: `version: 52 dirty: false`.
- Worker recent pass: `providers=2 claimed=2 skipped=0 accepted=4 rejected=0 failed=0`.
- Feedback re-read after the explicit close request: `total=55`,
  `unresolved=0`, `resolved=55`, `missing_resolved_at=0`, `statuses=open=55`.
- `status` was deliberately left untouched; it remains the developer-agent
  lifecycle and is not the owner-resolution flag.
