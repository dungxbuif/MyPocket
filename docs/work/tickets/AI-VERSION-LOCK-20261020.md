---
status: in_review
owner: shared
---

# AI transaction/category optimistic locking

## Intent

Close the existing `expected_version` gap for Finance Assistant proposals. A
confirmed update, recategorization, or delete must not mutate a row that
changed after the proposal was generated.

## Contract

- `transactions.version` and `categories.version` are positive integer row
  versions, initialized to `1`.
- Every successful transaction/category mutation increments the version.
- AI confirmation carries `expected_version`; a stale value returns HTTP `409`
  and performs no write.
- AI proposal schemas require a positive `expected_version` for transaction
  update/delete/recategorize and category update. Persisted cards missing the
  snapshot are rejected as `409 advisor_action_stale` before dispatch.
- Owner scoping remains mandatory. Legacy non-AI callers may omit the field;
  the handler binds the version read with the row so the repository still uses
  an atomic compare-and-swap.

## Verification

- Red/green HTTP tests for stale update/delete/recategorize and category update.
- PostgreSQL repository test for atomic version increment and conflict.
- Full Go test/race/vet/build gates.
- Native transaction/report visual/accessibility reference suites and a fresh
  Mobile MCP/WDA probe; if WDA is unavailable, record the exact environment
  blocker rather than claiming automated touch coverage.

## Implementation evidence — 2026-10-20

- `000040_row_versions` adds positive, non-null `version` columns and checks to
  `transactions` and `categories`.
- PostgreSQL repositories use compare-and-swap updates and increment versions;
  category wallet replacement is in the same transaction as the category edit.
  Transfer-pair edits also increment both row versions.
- Focused proof: `go test ./internal/controller/http ./internal/usecase ./internal/infrastructure/repository ./migrations -count=1`
  passes. Database-backed version assertions run when `TEST_DATABASE_URL` is
  configured; they skip explicitly when no migrated PostgreSQL is available.
- Mobile MCP/WDA is still unavailable in this environment (`mobilecli` not
  installed/WDA probe unavailable), so automated touch regression is not
  claimed. Physical device acceptance remains a separate gate.

## Migration safety follow-up — 2026-10-20

The migration test now asserts the `000040` SQL contract directly: both version
columns are `NOT NULL DEFAULT 1` and both tables reject values below `1`. This
prevents a release from passing only the filename/latest-version check while
silently weakening the CAS invariant. The test is fixture-only and does not
connect to or mutate production.
