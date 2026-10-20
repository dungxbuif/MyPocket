# MyPocket 1.2.1 — Advisor reliability corrections

Status: released on 2026-10-20. Production API/worker and signed native
artifact were verified after rollout.
Native metadata: 1.2.1 (4). API-only production topology remains in use.

## Changes

- Validate proposal references and current versions inside the real tool loop;
  unread/stale/system/archived selections cannot become confirmation cards.
- Restore owner/generation-scoped conversation hints after revalidating entity
  references. Keep 12 text messages/24 KiB and 96 KiB tool transcript budgets.
- Render evidence labels and handbook answers from server data. Preserve
  terminal finish_answer after eight finance calls, repair protocol with
  matched tool results, and avoid duplicate-evidence hangs.
- Recheck live leases under the reaper lock. Migration 000041 provides a
  database default for deadlines so the previous API can keep writing during
  the rolling update. This migration has not previously reached production.
- Advertise actual search limits/enums and return repairable date/range/amount
  errors to the model without accepting invalid inputs into finance reads.
- Correct API documentation version, previously still reported as 1.1.0.

## Verification and rollout

See [review ticket](../work/tickets/AI-ADVISOR-REVIEW-FIX-20261020.md),
[validation matrix](../work/VALIDATION_MATRIX.md),
[spec](../superpowers/specs/2026-10-20-advisor-reliability-design.md), and
[deployment](../operations/DEPLOYMENT.md).

Provider smoke, normal/race suites, real disposable PostgreSQL migration and
native build evidence are recorded after final verification. The 31-prompt
dataset schema test is not a live quality score. Runtime/grounding flags remain
off by default; a patch rollout does not silently enable unaccepted features.

No financial reset, primary-account mutation, API-key policy change or external
banking integration is included. Production migrated 40→41 with a verified
pre-migration backup and unchanged wallet/transaction/budget/category
checksums. API and worker run immutable image
`registry.dungxbuif.com/mypocket-api:release-1.2.1-837bd505` at digest
`sha256:34226ef0700f1ea4ff05d2c3de4d0fc0488d28ef71f7b4c735b09172c3d1fd54`;
health returned 200 and the live OpenAPI surface matched all 91 local routes.
Signed iPhone artifact 1.2.1 (4) was installed and launched. This does not
claim full physical UAT, Face ID, offline, or Mobile MCP/WDA acceptance.
