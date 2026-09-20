---
artifact_type: detail_design
id: API-SCREENS-01
status: in_review
owner: shared
approval: owner_requested_all_screens_real_api_2026_09_20
---

# API-only mounted screens

Owner requests all displayed screens use persisted data, not mockFinance. Context: [backlog](../BACKLOG.md), [savings](TICKET-06-01-DETAIL_DESIGN.md), [budget setup](TICKET-03-01-thiet-lap-budget.md), [budget progress](TICKET-03-02-tien-do-budget.md), [rules](../../requirements/BUSINESS_RULES.md). Phase none.

Audit: auth/account, overview, groups, wallets and transactions already consume APIs. Savings needs specialized real catalog filtering/history. Budget is the only mounted mock consumer; Reports is unmounted. This slice delivers budget persistence/list/create/edit/delete and live progress for explicit date intervals. Full recurring/week/quarter/year generation and notifications remain later budget acceptance, not silently completed.

Budget fields: owner, name, positive VND limit, optional owner wallet/category (expense only), start_at inclusive/end_at exclusive UTC instants. Browser converts local selected dates to boundaries. Scope includes category children. Exact same scope overlapping intervals rejected atomically under owner row lock. Ended budgets read/delete only; history recalculates from expense rows included in reports. Combined summary counts matched transactions once; limits sum actual budget allocations. Daily allowance uses actual remaining days per budget, not a fixed mock divisor. No balance mutation. Deletes wallet/category detach budget via cascade (budget is only derived tracking); explain dependency in UI. Prefer keeping FK restriction? Rejected: orphan scope would silently widen a budget. API is authenticated owner-scoped; same-origin proxy unchanged.

Architecture: entity/repository interface + Postgres repository and pure progress calculation, HTTP handler and explicit routes, migration 000009. Frontend service and Budget editor compose existing BaseBottomSheet, FormField, BaseTextInput, BaseSelect, BaseButton, StatusMessage, SurfaceCard, MetricBox, BudgetGauge, BudgetProgressItem, Text. No dependency/base styling changes. Retain loading/empty/error/retry, draft on failure and confirmation on delete. No fabricated fallback.

Risks: overlap race (serialize writes per owner), timezone date boundaries (send explicit RFC3339), stale aggregates (fetch on shared refresh), category descendants/double counting (pure tests), owner isolation (validation + owner-scoped repository; real API negative test). Test first pure calculation and validation; run all Go tests, design tests, transaction tests/build; real local API roundtrip and browser UAT. No production deployment.

Trace/reconciliation: [API](../../architecture/API.md), [ERD](../../architecture/ERD.md), [ADR](../../decisions/ADR-004-budget-api-data.md), [validation](../VALIDATION_MATRIX.md), [release](../../releases/CHANGELOG.md), [context](../../CONTEXT.md), [screen](../../design/screens/budgets/README.md). Automated and runtime proof to be recorded here; owner visual acceptance pending. Design approved by explicit API integration instruction; no separate approval inferred for recurring scheduler or transfer pairs.

## Verification and docs review — 2026-09-20

- Pass: `rtk proxy go run ./cmd/migrate up` (dev migration9), `rtk proxy go generate ./cmd/api`, `rtk proxy go test ./...` from backend. All Go tests also rerun with TEST_DATABASE_URL pointing at migrated local dev PostgreSQL so optional integration test executes rather than skips.
- Pass in app: `rtk proxy npm run check:design`, `rtk proxy npm run test:design` (7 guard cases + shared SSR), `rtk proxy npm run test:transactions` (5), `rtk proxy npm run build`, `rtk proxy node scripts/api-roundtrip.mjs`. Only Node's existing experimental type-stripping warning remains; no failing tests.
- API roundtrip verifies real category IDs, target date and balance after savings deposit/withdrawal/interest/edit/delete, goal catalog rejection; budget CRUD, overlap409, descendants, excluded report rows, recalculation and unauthenticated401. Test-created wallets/transactions/budgets are cleaned by exact IDs. First run exposed test catalog assumption: migration5 makes salary/food personal, hence icon_key fallback for test lookup; no production catalog altered.
- Independent code review found GORM Save could resurrect a deleted row and metadata edits could shift original timestamps after timezone changes. Fixed with owner-scoped Updates + RowsAffected and matching delete lock; optional Postgres disappearing-row regression failed before fix then passed. Frontend now preserves original start/end instants independently for unchanged displayed dates. Intentional date changes follow current browser-local calendar; account timezone persistence is not implemented.
- Independent read-only re-check: no remaining Important issues in the reviewed scope; reviewer inspected test source but execution evidence belongs to the main-agent runs above. No commits or production deployment performed.
- Browser: reauthenticated Google via FE origin; `/budgets` shows actual empty state, editor loads real wallet/category choices; grouped QuickAdd visually inspected. Drafts canceled, owner data unchanged. Persistence proof is API-level, not full browser CRUD UAT. User visual acceptance pending, therefore in_review.
- Docs review complete: API, ERD, ADR004, screen contracts, context, backlog, validation matrix and changelog reconciled. Parent budget tickets are partial; weekly/quarterly/yearly recurring automation, forecast/notifications and internal paired transfers are not represented as complete. Reports source remains unmounted and outside production import graph; it is not a delivered real reporting screen.
