# Testing Patterns

**Analysis Date:** 2026-09-10

## Test Framework

**Runner:**
- Go `testing` package with Go 1.23 for backend unit, repository, HTTP, worker, migration, and adapter tests.
- Vitest 3.2 for frontend unit and component tests.
- Config: `frontend/vite.config.ts` (jsdom, setup file, E2E/docs exclusion, serial test files).
- Playwright 1.62 for browser/API end-to-end tests.
- Config: `frontend/playwright.config.ts`; opt-in isolated R0 harness: `frontend/playwright.r0.config.ts`.

**Assertion Library:**
- Go standard `testing` assertions expressed with `t.Fatal`, `t.Fatalf`, and `errors.Is`.
- Vitest `expect` plus `@testing-library/jest-dom` matchers configured by `frontend/src/test/setup.ts`.
- React Testing Library and `@testing-library/user-event` for component interaction.
- Playwright `expect` for page, API response, and polling assertions.

**Run Commands:**
```bash
cd backend && go test ./...                                      # Run backend tests; DB-backed tests skip without test configuration
cd backend && MYPOCKET_TEST_DATABASE_URL='<dedicated URL>' go test -race -p 1 ./... -count=1  # Full backend proof
cd frontend && npm test -- --run                                # Run all Vitest tests once
cd frontend && npm test                                         # Vitest watch mode
cd frontend && npm run test:e2e                                 # Run all Playwright projects
cd frontend && npx playwright test e2e/business-core.spec.ts --workers=1  # Focus one browser journey
MYPOCKET_TEST_DATABASE_URL='<dedicated mypocket_verify_* URL>' ./scripts/verify-beta.sh  # Release ladder
```

No line/branch coverage script or enforced numeric threshold is configured in `frontend/package.json`, `frontend/vite.config.ts`, or `backend/go.mod`.

## Test File Organization

**Location:**
- Go tests are co-located with implementation packages under `backend/internal/`: `backend/internal/finance/transactions_test.go`, `backend/internal/platform/httpapi/auth_test.go`.
- TypeScript unit/component tests are co-located under `frontend/src/`: `frontend/src/offline/db.test.ts`, `frontend/src/screens/ReportsPanel.test.tsx`.
- Shared frontend setup and reusable component-contract tests live under `frontend/src/test/`: `frontend/src/test/setup.ts`, `frontend/src/test/baseComponents.test.tsx`.
- Browser journeys are separate under `frontend/e2e/`: `frontend/e2e/business-core.spec.ts`, `frontend/e2e/offline-sync.spec.ts`.
- Runtime proof and exact command results are recorded separately in `docs/work/test-verification/` and `docs/work/VALIDATION_MATRIX.md`; tests alone are not completion evidence.

**Naming:**
- Go: `*_test.go`, with exported behavior named `Test<Behavior>` and subtests named as readable scenarios.
- Vitest: `<module>.test.ts` or `<Component>.test.tsx`, with `describe` naming the subject and `it` naming observable behavior.
- Playwright: `<workflow>.spec.ts`, with full-sentence `test(...)` names describing the user or system journey.

**Structure:**
```
backend/internal/<domain>/
├── <subject>.go
├── <subject>_test.go
└── repository_test.go

frontend/src/<area>/
├── <subject>.ts[x]
└── <subject>.test.ts[x]

frontend/e2e/
└── <workflow>.spec.ts
```

## Test Structure

**Suite Organization:**
```go
func TestApplyAccountingEffect(t *testing.T) {
	tests := []struct {
		name  string
		input finance.AccountingInput
		want  int64
	}{
		{name: "income increases source", input: /* ... */, want: 1_100_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := finance.ApplyAccountingEffect(tt.input)
			if err != nil { t.Fatalf("expected accounting effect, got %v", err) }
			if got.SourceBalanceVND != tt.want { t.Fatalf("source balance = %d, want %d", got.SourceBalanceVND, tt.want) }
		})
	}
}
```
Pattern source: `backend/internal/finance/transactions_test.go`.

```typescript
describe("calendarDateInHoChiMinh", () => {
  it("uses the Vietnam calendar day instead of UTC or device timezone", () => {
    expect(calendarDateInHoChiMinh(new Date("2026-09-09T18:30:00.000Z"))).toBe("2026-09-10");
  });
});
```
Pattern source: `frontend/src/app/transactionInput.test.ts`.

**Patterns:**
- Use table-driven Go subtests for multiple inputs with the same invariant, especially validation and arithmetic boundaries: `backend/internal/finance/transactions_test.go`.
- Use external Go test packages (`finance_test`, `sync_test`, `httpapi_test`) when proving public behavior; use internal-package tests only when a narrow unexported boundary requires it, as in `backend/internal/planning/budgets_internal_test.go`.
- Assert both result and error semantics. Use `errors.Is` for wrapped domain sentinels rather than comparing strings.
- For repository changes, assert persisted state, versions, row counts, change-feed records, and rollback—not only returned values—as in `backend/internal/sync/atomic_test.go`.
- For React tests, render the real component, operate through accessible labels/roles, and assert visible state. `frontend/src/screens/ReportsPanel.test.tsx` is the primary pattern.
- For browser tests, combine UI actions with API/database-observable outcomes and use `expect.poll` for eventual consistency, as in `frontend/e2e/business-core.spec.ts`.

## Mocking

**Framework:**
- Vitest mocks/spies (`vi.fn`, `vi.mock`, `vi.stubGlobal`, fake timers).
- Go hand-written fakes/test doubles and explicit failure injection through test database triggers.
- Playwright request interception and fixture OAuth mode for controlled browser flows.

**Patterns:**
```typescript
const fetcher = vi.fn(async () => new Response(JSON.stringify({ report }), { status: 200 }));
vi.stubGlobal("fetch", fetcher);
render(<ReportsPanel {...props} />);
await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
expect(await screen.findByRole("table", { name: "Tổng hợp báo cáo" })).toBeVisible();
```
Pattern source: `frontend/src/screens/ReportsPanel.test.tsx`.

```go
_, err := conn.ExecContext(ctx, `CREATE FUNCTION fail_receipt() RETURNS trigger ...`)
if err != nil { t.Fatal(err) }
if _, err := service.ApplyMutations(ctx, owner, []mysync.Mutation{mutation}); err == nil {
	t.Fatal("expected injected failure")
}
assertSyncWalletBalance(t, conn, wallet.ID, 0)
```
Pattern source: `backend/internal/sync/atomic_test.go`.

**What to Mock:**
- Mock browser primitives and network boundaries in focused Vitest tests: `fetch`, IndexedDB, online state, timers, match media, and service-worker events. Global setup provides a fresh fake IndexedDB in `frontend/src/test/setup.ts`.
- Mock provider/error boundaries when proving safe behavior without live infrastructure, while retaining separate adapter smoke tests under `backend/internal/platform/objectstore/objectstore_test.go` and `backend/internal/platform/authcache/cache_test.go`.
- Inject deterministic clocks, fixed dates, IDs, or failure points where the production API already permits it.

**What NOT to Mock:**
- Do not mock domain accounting, ownership, version, idempotency, rollback, or change-feed behavior when the claim concerns persisted correctness. Use a migrated PostgreSQL database and query state directly, following `backend/internal/finance/repository_test.go` and `backend/internal/sync/atomic_test.go`.
- Do not treat a handler-only test as proof of a ledger invariant. Exercise the repository and persisted balances.
- Do not mock the entire application in user-facing E2E proof. `frontend/playwright.config.ts` starts a real Go API, migrates a dedicated database, builds the frontend, and runs against the preview server.
- Do not replace browser accessibility behavior with implementation selectors when a role, label, or visible name exists.

## Fixtures and Factories

**Test Data:**
```go
func createFinanceUser(t *testing.T, conn *sql.DB, email string) string {
	t.Helper()
	var userID string
	err := conn.QueryRowContext(context.Background(), `INSERT INTO users (...) VALUES (...) RETURNING id::text`, "subject-"+email, email).Scan(&userID)
	if err != nil { t.Fatalf("create finance user: %v", err) }
	return userID
}
```
Pattern source: `backend/internal/finance/repository_test.go`.

```typescript
function randomName(prefix: string) {
  return `${prefix} ${Date.now().toString().slice(-6)}-${crypto.randomUUID().slice(0, 4)}`;
}
```
Pattern source: `frontend/e2e/business-core.spec.ts`.

**Location:**
- Go fixture helpers live at the bottom of the owning test file and call `t.Helper()`: `backend/internal/finance/repository_test.go`, `backend/internal/planning/budgets_test.go`.
- Frontend test builders stay local to the suite when narrowly used: `summary`, `report`, `props`, and `filters` in `frontend/src/screens/ReportsPanel.test.tsx`.
- E2E action helpers live near the top of each spec: login, wallet, and transaction helpers in `frontend/e2e/business-core.spec.ts`.
- There is no central cross-suite fixture factory directory. Prefer local helpers until reuse across several suites justifies a shared test module.

## Coverage

**Requirements:** None enforced numerically. Behavioral coverage is required by `docs/standards/TESTING.md` and `docs/standards/VALIDATION.md`: features need relevant unit/integration proof, bugs need reproduction/regression proof, API work needs success/failure contract proof, schema work needs migration-sensitive proof, and user-visible work needs UAT or a recorded reason.

**View Coverage:**
```bash
# No repository coverage command is configured.
cd backend && go test -cover ./...          # Ad hoc Go package coverage if needed
cd frontend && npm test -- --coverage       # Requires adding/configuring a compatible Vitest coverage provider first
```

Do not report Go test-event counts or passing test counts as line/branch coverage. `docs/work/test-verification/MONEYLOVER-PARITY-2026-09-08.md` explicitly distinguishes these.

## Test Types

**Unit Tests:**
- Pure Go domain validation, arithmetic, parsing, retry, redaction, and state rules under packages such as `backend/internal/finance/transactions_test.go`, `backend/internal/audit/redact_test.go`, and `backend/internal/notification/retry_test.go`.
- Pure TypeScript data conversion, cache, offline reducer, migration, and outbox rules under `frontend/src/app/*.test.ts` and `frontend/src/offline/*.test.ts`.
- Use fixed boundary values for VND arithmetic, time-zone dates, zero baselines, overflow, stale versions, and malformed input.

**Integration Tests:**
- PostgreSQL repository and migration tests reset and migrate a dedicated schema, then assert persisted behavior: `backend/internal/finance/repository_test.go`, `backend/internal/platform/db/migrate_test.go`.
- HTTP tests exercise routers/auth/error contracts through real request/response boundaries: `backend/internal/platform/httpapi/*_test.go`.
- Redis and S3 adapter tests run only when their test environment variables are configured and otherwise skip explicitly: `backend/internal/platform/authcache/cache_test.go`, `backend/internal/platform/objectstore/objectstore_test.go`.
- Because multiple packages reset the same public schema, run database-backed package tests with `-p 1`. Parallel package execution against one database is invalid evidence; this constraint is codified in `.agents/skills/auditing-mypocket-business-logic/SKILL.md` and `scripts/verify-beta.sh`.

**E2E Tests:**
- Playwright covers Chromium mobile (`Pixel 5`), desktop Chrome, and mobile WebKit (`iPhone 13`) via `frontend/playwright.config.ts`.
- The harness starts the migrated API in OAuth fixture mode and a production frontend build; do not point it at production or reuse a shared database/server.
- Browser coverage includes auth/PWA, finance CRUD, accounting correctness, offline sync, planning, operation/search feedback, receipts, layout, and cross-user isolation under `frontend/e2e/`.
- Use the isolated `frontend/playwright.r0.config.ts` when a second local harness must avoid the default web/database ports; it derives from the base config and changes output, web port, and database URL.

## Common Patterns

**Async Testing:**
```typescript
await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
const table = await screen.findByRole("table", { name: "Tổng hợp báo cáo" });
expect(table).toBeVisible();
```
Use `findBy*` for state that appears after promises resolve and `act` when manually resolving a deferred response, following `frontend/src/screens/ReportsPanel.test.tsx`.

```typescript
await expect.poll(async () => Promise.all([
  walletBalance(page, sourceName),
  walletBalance(page, destinationName),
])).toEqual([750_000, 250_000]);
```
Use `expect.poll` for server-observable eventual state in `frontend/e2e/business-core.spec.ts`.

**Error Testing:**
```go
_, err := finance.ApplyAccountingEffect(input)
if !errors.Is(err, finance.ErrValidation) {
	t.Fatalf("expected validation error, got %v", err)
}
```
Pattern source: `backend/internal/finance/transactions_test.go`.

```typescript
expect(await screen.findByRole("alert")).toHaveTextContent("req_report");
expect(screen.queryByText("SECRET_TRACE")).not.toBeInTheDocument();
expect(screen.getByLabelText("Ví báo cáo")).toHaveValue("cash");
```
Test both safe feedback and preserved user state, and prove sensitive server text is hidden, following `frontend/src/screens/ReportsPanel.test.tsx`.

**Database Setup:**
- Read `MYPOCKET_TEST_DATABASE_URL`; if absent, skip repository integration proof with an explicit reason.
- Open with pgx, register `t.Cleanup`, drop/recreate the public schema, and run migrations from `backend/migrations/`, following `migratedFinancePostgres` in `backend/internal/finance/repository_test.go`.
- Never use production. `scripts/verify-beta.sh` refuses schema-reset tests unless the database name begins with `mypocket_verify_`.

**Isolation and Cleanup:**
- `frontend/src/test/setup.ts` installs a fresh `fake-indexeddb` factory before every test and cleans React plus localStorage afterward.
- Suites that stub globals or timers restore them in `afterEach`, as in `frontend/src/screens/ReportsPanel.test.tsx`.
- Vitest files run serially (`fileParallelism: false`) because suites replace process-wide browser primitives; preserve this constraint in `frontend/vite.config.ts`.
- Playwright tests create unique users/names and dispose request contexts in `finally`, as in `frontend/e2e/user-isolation.spec.ts`.

**Verification Ladder:**
1. Run the smallest focused test and record expected RED for a defect or new behavior.
2. Run the owning package or frontend suite after GREEN.
3. Run neighboring invariant/workflow tests, especially create/edit/archive/replay/conflict for accounting changes.
4. Run `go vet`, backend race tests with serial database packages, frontend typecheck/Vitest/build, and relevant browser projects.
5. Record exact commands and pass/fail/skipped outcomes in the active artifact and `docs/work/VALIDATION_MATRIX.md`; use `scripts/verify-beta.sh` for release-level proof.

---

*Testing analysis: 2026-09-10*
