# Local functional and accounting verification

Status: in_progress. Owner requested all functions be run locally with E2E and accurate data, prioritizing reports/statistics and wallet transfers after a deployed wallet-detail crash. This approves local verification and corrections within existing finance/report designs, not new production releases, provider credentials, schema/auth changes or automatic promotion of deferred capabilities.

Execution update: bounded wallet/transfer/report slice is in_review; the full requested functional audit/remediation remains in_progress. [Evidence](../test-verification/R0-FUNCTIONAL-E2E-2026-09-09.md): 120 frontend pass, Go/PostgreSQL pass, full browsers 61/63 with two retained WebKit offline failures. Existing report design enabled a mounted five-kind/date/wallet selector with accessible server-number tables and local-only stale-result warning; full charts/cache/drilldown remain out of this slice. Public docs completeness is not accepted and these local changes are not deployed.

Trace: [R0](R0-DETAIL_DESIGN.md), [finance design](PHASE-002-detail-design.md), [report design](PHASE-005-detail-design.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [verification](../test-verification/R0-FUNCTIONAL-E2E-2026-09-09.md), [changelog](../../releases/CHANGELOG.md).

## Investigation and proposed work

Confirmed deployed stack `Sb` maps to `WalletDetailCard`: API wallet detail returns a nil Go transaction slice as JSON null, while frontend assumes an array and accesses length. Other detail-list payload fields must be examined, not merely hidden with a render guard. Independently, TransactionsScreen hardcodes amount zero for transfer/adjustment, concealing actual ledger amounts. Reports require independent known-answer checks; existing E2E mostly verifies presence/CRUD and does not establish all financial totals.

Use existing dirty checkout, preserve earlier edits, no commits. Keep current UI and shared base components. Reproduce bugs with failing tests before changing code. Normalize compatible empty collections at the API-client boundary for the current deployed API, and verify nonempty wallet-detail wire mapping. Correct neutral transfer presentation without treating it as income/expense. Compare reports to approved literal expected amounts, selected wallet/timezone/date ranges, exclusion flags, edits and archival. Add bounded UI wiring only where approved existing design defines behavior; record larger/deferred gaps explicitly.

## Local environment and proof

Disposable PostgreSQL `mypocket-r0-20260909-pg`, loopback 64739, database `mypocket_r0_e2e`, fixture-auth API 18173, preview web 4187. Existing R0 config with no server reuse; run one browser runner at a time. Use unique fixture identities where needed so suites cannot contaminate aggregate expectations. Never use production credentials/data, no public-site financial mutations, no deploy. Backend integration suite may use a separate disposable test database to avoid fixture races.

Run existing full browser baseline on Chrome mobile/desktop and WebKit; retain known WebKit offline failures rather than skip. Add empty-wallet browser regression and transfer lifecycle checks: create, edit, archive, both balances, single transaction, unchanged global income/expense; test same-wallet rejection. Report proof uses known income/expense/transfer/adjustment/excluded transactions across periods with independent expected values. Re-run unit/typecheck/build and browser regression after fixes. Record failures, unsupported flows, provider/device limitations and coverage per function; no claim of all functions complete if any unverified or missing.

## Boundaries, alternatives and reconciliation

Rejected: mock-only evidence, hardcoded/sample financial values, disabling failed tests, entire UI redesign, bundling dirty backend into production. No changes to accounting semantics, DB/schema/auth/runtime services or dependencies without a specific design decision. Refresh API/docs only for actual corrected contracts, and context/backlog/validation/changelog with results. No new ADR unless a durable architecture/contract decision is required. Human physical-device/provider acceptance remains separate.
