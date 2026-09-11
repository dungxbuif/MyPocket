# Portfolio checked-money audit

Status: in_review (local). Scope authorized by the owner's continuing business-correctness audit and instruction to fix confirmed defects without asking again. This preserves ADR-006 accounting and existing validation errors; no new API fields, schema, permission, runtime or architecture decisions.

Problem/hypothesis: portfolio converts arbitrary-precision quantity products with unchecked `Int64()` and adds money with unchecked integer arithmetic. Values outside signed 64-bit range can silently wrap in trade replay, realized profit and market summaries.

Plan: reproduce with literal boundary fixtures; check rounded products and use exact intermediate arithmetic for ledger calculations before conversion. Propagate existing `ErrValidation` for out-of-range results through repository reads/commands; preserve transaction rollback/feed behavior. Keep half-away-from-zero rounding and moving-average accounting unchanged. Reject clamping or float conversion because they conceal wrong financial values.

Verification: unit overflow and boundary cases, real PostgreSQL failed-command rollback and aggregate overflow, neighboring portfolio/sync/analytics packages, full release script. Physical-device/provider gates remain separate. No production writes.

Docs reconciliation: API assets and audit update page, internal API, audit evidence, context/backlog/validation/changelog. Requirements and ADR-006 accounting remain unchanged; ERD and runtime unchanged. Tests serve as automated acceptance for the internal arithmetic correction; physical UAT is not substituted.

Links: [audit](../test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [ADR-006](../../decisions/ADR-006-separate-asset-portfolio-valuation.md), [API](../../architecture/API.md), [changelog](../../releases/CHANGELOG.md).

Verification 2026-09-11: five ledger overflow subcases observed RED; focused portfolio and neighboring analytics/sync race suites GREEN. Full backend `go test -race -p 1 ./... -count=1` with the dedicated `MYPOCKET_TEST_DATABASE_URL` passed after final changes; `go vet ./...` passed. Regression tests prove max valid boundary acceptance, trade/price rollback including version/feed, and aggregate overflow rejection. Public assets/API docs and audit work records reconciled; Docusaurus build and whitespace check pass. No deployment, no schema changes.
