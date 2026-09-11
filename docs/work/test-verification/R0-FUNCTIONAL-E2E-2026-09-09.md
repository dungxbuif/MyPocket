# Local functional/accounting E2E — 2026-09-09

Status: in_review for the bounded wallet/transfer/report changes; **whole-app completion remains in_progress**. No deployment, commit, production data mutation, backend contract/schema/auth change or provider acceptance in this run.

Trace: [design](../phases/R0-FUNCTIONAL-E2E-DETAIL_DESIGN.md), [R0](../phases/R0-DETAIL_DESIGN.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [architecture](../../architecture/ARCHITECTURE.md), [changelog](../../releases/CHANGELOG.md), [previous production release](R0-WEB-DOCS-RELEASE-2026-09-09.md).

## Findings and corrections

- Deployed `index-Dz4PSmjw.js` stack `Sb` maps to `WalletDetailCard`: wallet detail returns `transactions: null` for an empty Go slice, then the view reads `.length`. Normalize the collection at the frontend API boundary; verify empty and nonempty detail through real App and browser tests. This correction is local, not in the deployed image.
- TransactionsScreen displayed zero for transfer/adjustment. Pass the actual API amount and use a neutral shared TransactionRow amount presentation; label adjustment as the resulting balance, not an income delta.
- Mounted quick-add now exposes transfer with shared Select for source/destination, positive amount and distinct-wallet validation. Existing edit/archive flow is exercised. No accounting formula is reimplemented in the frontend.
- Overview's inert report label is now a shared Button opening ReportsPanel. Shared Button/Input/Select/Card/OperationError support five existing report kinds, date/wallet filters, accessible numeric tables, explicit loading/error/retry, cancellation, account isolation and privacy. Overview no longer fabricates zero when report loading failed or exposes report totals/chart while masked.
- ReportsPanel preserves only its last loaded in-memory result while mounted offline, with timestamp/staleness warning. Persistent per-query cache, charts, drilldown and full Money Lover report parity are NOT implemented.

## Fresh proof

All tests use disposable PostgreSQL `mypocket-r0-20260909-pg`, loopback port 64739. Browser database `mypocket_r0_e2e`; Go integration database `mypocket_integration` is separate to avoid resets contaminating fixture tests. Browser web 4187/API 18173, fixture OAuth, no production credentials.

| Check | Result |
| --- | --- |
| Existing browser baseline, 18 scenarios × 3 projects | 52 passed, 2 known WebKit offline failures |
| Full frontend unit/component tests after changes | **120 passed / 15 files**, exit 0 |
| Full Go suite with separate PostgreSQL integration DB | `go test -p 1 ./... -count=1`, exit 0; provider skips are not real Redis/S3 acceptance |
| Typecheck + Vite build | Passed as fresh browser web-server startup prerequisite |
| Final browser suite, 21 scenarios × 3 projects | **61 passed / 2 failed**, exit 1; Chrome mobile 21/21, desktop 21/21, WebKit 19/21 |
| New accounting scenarios | **9/9 passed**, included in the 61 above, not additional independent passes |
| Whitespace check | `rtk proxy git diff --check`, exit 0 before reconciliation |

The final browser run is `frontend/test-results/r0-functional-final`. Retained failures:

1. `pwa-shell.spec.ts:17`: WebKit internal error on reload under `context.setOffline(true)`.
2. `receipt-controls.spec.ts:103`: WebKit `NotReadableError` reading the stored File under offline emulation.

Neither case was skipped or weakened. The separate HTTP-outage receipt test passes on all projects. Prior investigation/physical-device gates remain open; no further offline fix attempt was made.

### Known-answer ledger, not presence-only proof

`frontend/e2e/accounting-correctness.spec.ts` uses unique users and the real API/database. Transfer: source starts at 1,000,000 VND; create 250,000 gives balances 750,000/250,000; edit to 400,000 gives 600,000/400,000; archive restores 1,000,000/0. Same-wallet request is rejected without changing balances. Exactly one transfer is listed; overall income stays 1,000,000 and expense 0.

Report period Sep 1–2 in Asia/Ho_Chi_Minh includes income 1,000,000 and expenses 120,000 + 80,000; excludes a 70,000 flagged expense, archived 50,000 expense, transfer and adjustment. Expected net 800,000. An income one second before Sep 1 local midnight belongs to the prior period; Sep 3 local midnight is excluded. Five API summaries and five mounted report selectors match literals; destination-only wallet reports zero income/expense. Category share is 100%; daily cumulative net is 880,000 then 800,000; prior income 999,999; Insider expense 200,000/average 100,000; current live net worth 1,973,456 is explicitly distinct from selected-period flow.

### Attempt log

- Initial E2E fixture calls used relative `/api` and failed against preview's default 8080 proxy; corrected to the existing test API URL 18173. Income/expense fixtures initially lacked required categories; added actual system categories. These were fixture failures, not accounting regressions or backend outages.
- Corrected RED evidence: report API literals passed; empty-wallet render crashed and transfer action was absent. Unit RED also reproduced the exact null access after completing fixture endpoints.
- Transfer display and report panel component tests were RED before implementation, then GREEN. Parent fresh full suite independently confirms 120 pass.
- One browser startup ran while worker test typing was incomplete (TS2493); no browser tests ran. Worker corrected the mock typing, fresh typecheck passed.
- After wiring transfer, empty-wallet and transfer lifecycle passed. Report entry RED confirmed missing button before parent wiring.
- First integrated report run timed out because `getByLabel` exact matching included nested option text. Snapshot already showed the correctly named combobox. Changed only test selection to the accessible combobox role; final report tests all passed with unchanged numeric assertions. No financial expectation was adjusted to match implementation.

## Functional inventory: coverage is not completion

| Area | Actual state / remaining work |
| --- | --- |
| Auth/API keys | Fixture login/reload/logout, two-user cookie/bearer isolation, create/revoke/retry covered. Production OAuth and exhaustive bearer endpoint contract suite are separate. |
| Wallets | CRUD existing coverage; new empty/nonempty detail regression. Failed/racing detail loads still need explicit feedback. |
| Categories | Mounted manager creates/lists. Rename/archive/per-wallet activation/hierarchy UI remains incomplete despite API capabilities. |
| Transactions | Income/expense CRUD and new transfer create/edit/archive proven. Editor does not expose changing date/category/wallet; adjustment creation UI absent. API accepts only positive adjustment targets; do not silently change that contract. Month control remains incomplete. |
| Reports | Five kinds/date/wallet filters and numeric server tables now local. Charts, persistent query cache, drilldown, six-period Insider/projection absent. Overview seven-day preview currently selects last seven days of returned month, not necessarily through today. |
| Budgets | CRUD/progress tests pass. Aggregate hero sums potentially overlapping/mixed-period budgets while showing one period; not yet corrected or accepted numerically. |
| Planning | Event/debt links covered. Draft confirm/reject missing; recurring edit controls and create-then-link partial failure need further work. |
| Portfolio | Current Account UI supports creation/listing; trading/price/archive controls and privacy coverage not complete in mounted UI. |
| Search/notifications | Search failure/race tests pass; result detail navigation/cache freshness and mark-read error feedback incomplete. |
| PWA/receipts | Chrome offline tests pass; two retained WebKit failures above. Real iPhone installation, airplane mode and provider upload are unverified. |
| Deferred capabilities | AI/OCR, exports, bank integrations, reset/delete account are not promoted to implemented by this run. |

## Docs audit requested during this run

Docusaurus source has 17 pages covering intro/update, ERD and major API groups. It is primarily technical reference, not a complete feature/user guide. Sidebar has no workflow guide section; OpenAPI-labelled snippets are not a validated standalone OpenAPI document. Transactions reference omits a complete update/archive/receipt contract and uses abbreviated conceptual field names; API/UI/production reachability is not mapped per action. Existing intro and update already acknowledge incompleteness. No authenticated live docs contents were verified: web-tool access to the supplied URL failed. Do not equate reading local MDX with checking the deployed site.

This run has **not updated or deployed the public Docusaurus pages**. Next docs work: a source-verified feature matrix, end-user workflows with error/offline limits, complete API examples/schema/security semantics, numerical report/transfer examples, and release-labelled coverage. Do not describe these new local corrections as production features.

## UI QA scope and remaining risk

In-app Browser available and connected. It rendered the local MyPocket login screen (page identity/nonblank), but the test runner stopped its servers before interactive login; navigation then failed `ERR_CONNECTION_REFUSED` to local API 18173. No successful interactive report screenshot or console-health claim from this attempt. Existing project Playwright provided the real rendered interaction proof above; it was launched before the supplementary Browser check, not as an alleged Browser-unavailable fallback. Visual fidelity/overlap acceptance remains open.

## Commands

Frontend prefix: `rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin`.

- `npm test -- --run` (with DEBUG_PRINT_LIMIT=300).
- `npx playwright test --config=playwright.r0.config.ts --workers=1 --output=test-results/r0-functional-final`.
- Backend: `rtk proxy env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:64739/mypocket_integration?sslmode=disable' go test -p 1 ./... -count=1`.

## Reconciliation checklist

Cleanup: browser runner stopped its API/preview servers. Stopped the exact disposable tmpfs PostgreSQL container `mypocket-r0-20260909-pg`; test records are removed and can be regenerated by fixtures. Production untouched. Temporary Browser tab close was blocked by the browser tool's error-page URL policy; no successful close is claimed and the tab was not marked to persist.

- [x] Existing finance/report design used; no schema, auth, accounting semantics, API or production runtime change; no new ADR required.
- [x] Design, context, backlog, validation, architecture and changelog point to this bounded evidence and residual inventory.
- [x] Existing dirty checkout preserved; no stage/commit/deploy.
- [ ] Public docs reconciliation and authenticated docs rendering.
- [ ] Full functional completion, independent final review, responsive screenshot QA and human/physical-device/provider acceptance.
