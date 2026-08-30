---
artifact_type: test_verification
id: PHASE-002-finance-core
status: ready
owner: shared
human_fields:
  - uat_sign_off
  - manual_acceptance_notes
ai_fields:
  - commands
  - automated_tests
  - manual_checks
  - failure_summary
  - evidence_notes
  - attempt_log
shared_fields:
  - status
  - trace
  - uat
trace:
  backlog_item: BL-002
  requirement:
    - REQ-F-002
    - REQ-F-003
    - REQ-F-011
    - REQ-NF-002
    - REQ-NF-005
  phase: PHASE-002
  ticket_or_bug:
    - TICKET-005
    - TICKET-006
    - TICKET-007
  detail_design: ../phases/PHASE-002-detail-design.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: per-ticket_completion_checklist
  release_notes: ../../releases/CHANGELOG.md
---

# PHASE-002 Finance Core Verification

## Status

- ID: PHASE-002-finance-core
- Status: ready
- Owner: shared

## Scope

- Ticket/Bug/Phase: PHASE-002, TICKET-005, TICKET-006, TICKET-007
- Tested behavior: wallet/category domain, transaction accounting, idempotency, Vietnamese seed data, receipt metadata foundation, and mobile finance workflows.

## Trace Links

- Backlog item: [BL-002](../BACKLOG.md)
- Requirement: [REQ-F-002, REQ-F-003, REQ-F-011, REQ-NF-002, REQ-NF-005](../../requirements/REQUIREMENTS.md)
- Phase: [PHASE-002](../phases/PHASE-002-finance-core.md)
- Tickets: [TICKET-005](../tickets/TICKET-005-wallet-category-domain.md), [TICKET-006](../tickets/TICKET-006-transaction-accounting-engine.md), [TICKET-007](../tickets/TICKET-007-vietnamese-seeds-receipt-metadata.md)
- Detail design: [PHASE-002-detail-design.md](../phases/PHASE-002-detail-design.md)
- Validation matrix: [VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md)
- Release notes: [CHANGELOG.md](../../releases/CHANGELOG.md)

## Commands

| Command | Result | Notes |
| --- | --- | --- |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -run TestPhase002FinanceTablesAndSeeds -count=1` | pass | RED first failed because `wallets` did not exist; GREEN passed after adding `0002_phase002_finance_core.sql`. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -count=1` | pass | Full migration package regression after PHASE-002 migration. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/finance -count=1` | pass | Unit-only finance validation proof; 4 tests passed when PostgreSQL URL was absent. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1` | pass | Wallet/category unit and PostgreSQL integration proof for TICKET-005 backend domain/repository behavior. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Backend regression after finance repository; API, worker, finance, identity, config, db, httpapi, objectstore packages passed. `-p 1` avoids package-level races because integration helpers reset the same local test schema. |
| `rtk npm test -- --run` | pending | Frontend component proof. |
| `rtk npm run test:e2e` | pending | Mobile browser finance workflow proof. |

## Fix/Test Attempt Log

| Attempt | Change Made | Command | Result | Failure Summary |
| --- | --- | --- | --- | --- |
| 1 | Added PHASE-002 migration, then changed `categories_system_key_unique` from partial to full unique index so `ON CONFLICT (system_key)` is valid. | `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -run TestPhase002FinanceTablesAndSeeds -count=1` | pass | Initial GREEN attempt failed with SQLSTATE 42P10 because PostgreSQL cannot use a partial unique index for that conflict target. |
| 2 | Added finance validation and repository methods after RED tests exposed missing package/methods; adjusted `UpdateCategoryInput` to match the test-facing API. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance -count=1` | pass | Initial finance GREEN compile failed because `UpdateCategoryInput.Name` was a pointer while tests and plan used a string field. |
| 3 | Reran backend regression sequentially after package-parallel integration tests raced on shared schema reset. | `rtk env GOCACHE=/private/tmp/mypocket-go-cache 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1` | pass | Plain package-parallel `go test ./...` can reset the shared local test schema while another package is using it. |

Loop guard:

- Same-path failure attempts: 1 / 3
- Total fix/test cycles: 3 / 5
- Blocked: no
- Human/design input needed: none before starting approved PHASE-002 plan.

## Automated Tests

- Passed: PHASE-002 migration/schema/seed test, full `internal/platform/db` package, wallet/category finance package tests, and full backend regression.
- Failed: none remaining for migration and backend wallet/category domain slices.
- Skipped: finance HTTP API, frontend, and E2E checks pending later PHASE-002 tasks.

## Manual Checks

- Pending mobile UI implementation.

## UAT

- Required: yes for wallet/category and transaction workflows; not required for receipt metadata foundation until PHASE-006.
- Reason if not required: partial exception applies only to non-user-facing receipt metadata foundation.
- Expected behavior: finance workflows use Vietnamese copy, exact VND integer formatting, user-owned data, and correct wallet balances.
- Verified behavior: pending implementation.
- Sign-off: pending.

## Failures And Follow-Up

- Migration GREEN attempt 1 exposed a partial-index conflict-target issue; fixed in the migration before committing.
- Finance validation GREEN attempt 1 exposed a test-facing type mismatch; fixed before repository implementation continued.
- Package-parallel backend regression with a shared DB URL exposed test harness schema-reset contention; rerun with `-p 1` passed.

## Evidence Notes

- PHASE-002 migration test was written and observed failing before production migration code existed: `wallets table missing id column: map[string]bool{}`.
- After adding migration `0002_phase002_finance_core.sql` and fixing the category unique index, `internal/platform/db` tests passed against the local Compose PostgreSQL test URL.
- Wallet/category repository tests were written and observed failing on missing methods before repository code existed, then passed against local Compose PostgreSQL after implementation.
