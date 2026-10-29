---
artifact_type: detail_design
id: DESIGN-06-01
status: in_progress
owner: shared
approval: owner_requested_savings_implementation_from_supplied_references
---

# Savings UI and goal-date slice

## Approved continuation — API-only screens (2026-09-20)

Owner agreed to complete savings and requested API integration for all screens, without mock data. First implementation slice completes savings using existing endpoints/catalog. `income_transfer_in`, `income_interest`, `expense_transfer_out` already exist in migration 000004; no duplicate seeds or migration. Goal category picker restricts to these real system rows plus existing kind/wallet applicability. Ordinary and credit entries now require an applicable category; the UI defaults the first valid group and the backend enforces the invariant. Basic wallets retain their catalog. Existing historical rows remain visible and deletable; editing an unsupported old category requires clearing/reselecting. External entries retain current explicit report flag/default true; no counterpart or automatic transfer pair.

History resolves category names/icons from API, never from sample amounts. Shared form groups wallet/amount/category/note/date with SurfaceCard, FormField, BaseTextInput, BaseSelect, Text, SegmentedControl and CategoryTreeSelector backed by the same BaseCategoryTree selection mode as group management. No new palette or screen-local group selector. Errors preserve drafts; reload failures are surfaced rather than displaying success with stale goal numbers.

Historical 2026-09-20 mounted-screen audit (superseded): Overview, Transactions, Account, Wallets and Groups used APIs while budgets/reports were still being integrated. Current runtime has API-backed budget CRUD/progress and a mounted Reports route; no fixture totals are rendered. Regression coverage includes goal category filtering/backend rejection, API create/edit/delete roundtrip with derived balance/date, shared component tests and production build. Alternatives remain rejected: synthetic category objects, inferring internal transfer from names, or pretending an unavailable API returned an empty list.

## Owner clarification — external money vs internal transfer

Owner clarified savings inflow/outflow may involve money outside the app. Do not require a counterpart wallet for these single-wallet entries. Moving money to another tracked wallet is a separate shared transfer flow. This resolves the counterpart question below for ordinary savings entries.

Official Money Lover [transfer guide](https://moneylover.zendesk.com/hc/en-us/articles/900000358523-Transfer-money-between-wallets) confirms a dedicated action with source wallet, destination wallet, amount, note/date; it creates outgoing/incoming entries, excludes transfers from reports by default and records fees separately. The guide does not establish atomic database behavior or edit/delete pairing. Those remain MyPocket implementation requirements. No claim that every category named “Tiền chuyển đến/đi” requires another tracked wallet.

Next savings work: external deposit/withdrawal and manually recorded interest categories remain the single-wallet savings contract. The shared internal transfer editor now keeps both sides consistent with idempotent pair mutation and avoids double-counting in income/expense reports. External-entry report defaults remain explicit and must not inherit internal-transfer exclusion automatically.

Owner requests implementation using shared components and supplied five screenshots. Reuse current wallet/transaction APIs; expose target_date on wallet input using existing database column. Accept optional YYYY-MM-DD, parse strictly, store UTC midnight, preserve date on metadata updates when omitted, explicit empty string clears. Non-goal input cannot retain a goal date. No schema migration.

UI: BaseSelect for immutable-on-save type; grouped goal form; date input; real goal detail and history using current_balance. Shared QuickAddSheet accepts initial wallet context. No automatic bank interest or fabricated transfers. Managed-wallet transfer uses the separate source/destination editor; savings deposit/withdrawal entries do not require a counterpart wallet. Notifications have no push infrastructure, so show unavailable rather than a nonfunctional enabled toggle.

Alternatives: copied reference HTML rejected; synthetic categories/transfer entries rejected because they corrupt semantics. Change only wallet date contract, goal UI and direct consumers. Tests: invalid/valid dates, update omission/clear, goal calculations, shared SSR, design checks, ledger helpers, backend Go tests/build. Regenerate Swagger.

Trace: [ticket](TICKET-06-01-muc-tieu-tiet-kiem.md), [page](../../design/pages/savings/README.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [release](../../releases/CHANGELOG.md), [context](../../CONTEXT.md), [API](../../architecture/API.md). Phase none. ADR: additive date exposure uses existing schema and date-label semantics, no new architecture. Docs review and UAT recorded after verification.

## Verification and remaining work

Continuation proof: goal filter FE+BE tests failed on unrestricted regular categories, then passed; SSR history failed on generic title then passed with API category names. `app/scripts/api-roundtrip.mjs` verifies persisted goal date and deposit/withdrawal/interest/edit/delete balances through FE proxy; cleanup only IDs it created. Browser inspected grouped QuickAdd with real wallet options; owner/full browser persistence acceptance remains pending. External counterpart question resolved. Specialized categories and shared internal transfer pairs are implemented; Reports is mounted, while narrative reporting, push notification delivery and owner UAT remain separate. [Cross-screen API work](API-SCREENS-01-DETAIL_DESIGN.md) replaces the historical mounted-budget mock. Docs review: API/screens/context/backlog/matrix/changelog reconciled; no goal migration required.

Date regression first failed (date not persisted), then passed. `rtk proxy go test ./...` and `rtk proxy go generate ./cmd/api` pass. `rtk proxy npm run check:design`, `rtk proxy npm run test:design`, `rtk proxy npm run test:transactions`, `rtk proxy npm run build` pass in app. Tests cover date validity, omit/clear updates, progress at/above/below goal, excluded wallet grouping and select form.

Browser: selector layout, goal select, name/target/balance input and Save state, cancel returning focus inspected. Test draft canceled; no user data persisted. Backend restarted with same OAuth config. Full browser persistence, transfer and goal transaction UAT pending. Docs reconciled: API, screen/base contracts, backlog/context/validation/release. Remains in_progress for owner acceptance and closer transaction-form fidelity; push/fan-out notifications and report narrative are separate product gaps. Do not mark whole ticket complete.
