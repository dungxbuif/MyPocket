# Coding Conventions

**Analysis Date:** 2026-09-10

## Naming Patterns

**Files:**
- Use lowercase Go filenames with underscores for multiword or specialized files: `backend/internal/finance/transactions.go`, `backend/internal/finance/repository_tx_test.go`.
- Keep Go tests beside their package implementation and suffix them `_test.go`: `backend/internal/sync/service.go`, `backend/internal/sync/service_test.go`.
- Use PascalCase filenames for React components and screens: `frontend/src/components/finance/TransactionRow.tsx`, `frontend/src/screens/ReportsPanel.tsx`.
- Use camelCase filenames for TypeScript domain helpers: `frontend/src/app/transactionInput.ts`, `frontend/src/app/userDataCache.ts`.
- Co-locate Vitest files with the implementation and suffix them `.test.ts` or `.test.tsx`: `frontend/src/app/transactionInput.test.ts`, `frontend/src/screens/ReportsPanel.test.tsx`.
- Put browser journeys in `frontend/e2e/` and suffix them `.spec.ts`: `frontend/e2e/business-core.spec.ts`, `frontend/e2e/user-isolation.spec.ts`.

**Functions:**
- Exported Go functions and methods use PascalCase; unexported helpers use camelCase: `ApplyAccountingEffect` and `checkedAdd` in `backend/internal/finance/transactions.go`.
- Go constructors use `New<Type>` and accept dependencies explicitly: `finance.NewRepository(conn)` in `backend/internal/finance/repository_test.go`.
- TypeScript functions use camelCase; React components use PascalCase: `calendarDateInHoChiMinh` in `frontend/src/app/transactionInput.ts` and `ReportsPanel` in `frontend/src/screens/ReportsPanel.tsx`.
- Test helpers describe the action or fixture they create: `createFinanceUser`, `assertWalletBalance`, and `migratedFinancePostgres` in `backend/internal/finance/repository_test.go`; `loginViaFixture` and `createWallet` in `frontend/e2e/business-core.spec.ts`.

**Variables:**
- Use Go initialisms in uppercase inside identifiers (`ID`, `VND`, `URL`, `SHA256`), as in `SourceWalletID`, `AmountVND`, and `ChecksumSHA256` in `backend/internal/finance/transactions.go`.
- Use camelCase for TypeScript variables and parameters (`sourceWalletID`, `occurredOn`) while preserving API wire names in snake_case (`source_wallet_id`, `occurred_at`) at request boundaries in `frontend/src/app/transactionInput.ts`.
- Use uppercase snake case for TypeScript module constants: `DB_NAME`, `DB_VERSION`, and `META_CURSOR` in `frontend/src/offline/db.ts`.
- Use short conventional receiver/local names in Go (`repo`, `conn`, `ctx`, `err`) and descriptive names for domain values (`sourceBalance`, `destinationBalance`) in `backend/internal/finance/transactions.go`.

**Types:**
- Go domain types use PascalCase and domain suffixes such as `Input`, `Effect`, `Repository`, and `Error`: `AccountingInput`, `AccountingEffect`, and `CreateTransactionInput` in `backend/internal/finance/types.go`.
- TypeScript props interfaces use `<Component>Props`: `ReportsPanelProps` in `frontend/src/screens/ReportsPanel.tsx` and `WalletRowProps` in `frontend/src/components/finance/WalletRow.tsx`.
- TypeScript unions model finite state and operation vocabularies: `OfflineMutationState` and `OfflineOperation` in `frontend/src/offline/types.ts`.
- Keep API response envelope types local to the consumer when they are narrow, as with `ApiEnvelope<T>` in `frontend/e2e/business-core.spec.ts`.

## Code Style

**Formatting:**
- Format Go with standard `gofmt`; source uses tabs, grouped imports, and conventional Go brace/layout rules throughout `backend/internal/`.
- TypeScript is formatted with two-space indentation, semicolons, trailing commas in multiline structures, and quoted module paths, as shown in `frontend/src/screens/ReportsPanel.tsx`.
- Quote style is not mechanically uniform: newer files commonly use double quotes (`frontend/src/app/transactionInput.ts`), while some component and E2E files use single quotes (`frontend/src/screens/AccountScreen.test.tsx`, `frontend/playwright.config.ts`). Match the file being edited.
- No Prettier, ESLint, or Biome configuration is present. Treat `frontend/tsconfig.json` strict typechecking and existing local formatting as the enforceable frontend baseline.

**Linting:**
- Run `go vet ./...` for backend static analysis; the release ladder codifies it in `scripts/verify-beta.sh`.
- Run `npm run typecheck` for TypeScript diagnostics; it executes `tsc --noEmit` from `frontend/package.json` with `strict: true` in `frontend/tsconfig.json`.
- Run `git diff --check` for whitespace errors as the final repository-level check in `scripts/verify-beta.sh`.
- Do not add a major linting or formatting dependency without detail design, an ADR, documentation, and test evidence, per `docs/standards/CODE.md`.

## Import Organization

**Order:**
1. Standard-library imports first in Go, grouped in parentheses (`context`, `errors`, `testing`) as in `backend/internal/finance/repository_test.go`.
2. Third-party Go packages next, separated by a blank line, such as `github.com/jackc/pgx/v5/stdlib` in `backend/internal/finance/repository_test.go`.
3. Project-local Go packages last, separated by a blank line and rooted at module `mypocket`, such as `mypocket/internal/finance` and `mypocket/internal/platform/db`.
4. External TypeScript packages first, then a blank line, then relative project imports, as in `frontend/src/screens/ReportsPanel.test.tsx` and `frontend/src/offline/db.test.ts`.
5. Use `import type` for type-only TypeScript imports when no runtime binding is needed, as in `frontend/src/app/transactionInput.ts` and `frontend/e2e/business-core.spec.ts`.

**Path Aliases:**
- No TypeScript path aliases are configured in `frontend/tsconfig.json`; use relative imports within `frontend/src/`.
- Go imports use the module path declared by `backend/go.mod`: `mypocket/internal/...`.
- Alias imports only to resolve ambiguity or reserved/common names, such as `mysync "mypocket/internal/sync"` in `backend/internal/sync/atomic_test.go`.

## Error Handling

**Patterns:**
- Define stable sentinel domain errors and wrap them with `%w` plus safe context so callers can use `errors.Is`; examples are validation paths in `backend/internal/finance/transactions.go` and assertions in `backend/internal/finance/transactions_test.go`.
- Return zero values with errors from validators and constructors; do not return partially normalized domain input after validation failure, as in `ValidateCreateReceiptObject` in `backend/internal/finance/transactions.go`.
- Check database and transaction errors immediately and add operation context. Repository tests expect exact rollback behavior in `backend/internal/sync/atomic_test.go`.
- Translate internal failures to stable, sanitized HTTP envelopes with correlation IDs; `backend/internal/platform/httpapi/errors_test.go` proves unsafe database details are not exposed.
- Frontend async operations preserve safe UI/input state, convert unknown failures to a user-safe `OperationFailure`, and retain correlation IDs through `frontend/src/components/feedback/OperationError.tsx`.
- Do not swallow promise rejections or stale responses. Track request identity/lifecycle where results can race user, filter, or authentication changes; tests in `frontend/src/screens/ReportsPanel.test.tsx` assert obsolete and cross-user responses are ignored.

## Logging

**Framework:** Go standard structured logging and HTTP correlation metadata; no standalone frontend logging framework is detected.

**Patterns:**
- Keep user-facing errors sanitized and expose only stable codes plus correlation IDs; use `backend/internal/platform/httpapi/errors.go` and `backend/internal/platform/httpapi/errors_test.go` as the reference boundary.
- Log operational context at backend HTTP/worker boundaries rather than embedding debug output in domain functions. Relevant boundary code lives under `backend/internal/platform/httpapi/` and `backend/internal/worker/`.
- Never put secrets, bearer keys, raw provider credentials, or unsafe internal error text in logs or responses. The negative contract is tested in `backend/internal/platform/httpapi/errors_test.go`.
- Tests should fail with descriptive assertion messages via `t.Fatalf` or Testing Library assertions rather than printing ad hoc diagnostics, as in `backend/internal/finance/transactions_test.go` and `frontend/src/screens/ReportsPanel.test.tsx`.

## Comments

**When to Comment:**
- Comment the reason for a non-obvious constraint, not a restatement of code. `frontend/vite.config.ts` explains same-origin proxying and why Vitest files run serially.
- Document safety boundaries next to automation. `scripts/verify-beta.sh` explains why integration tests require a dedicated database before schema reset.
- Keep business invariants in durable docs and prove them in tests; repository-specific rules live in `.agents/skills/auditing-mypocket-business-logic/references/invariants.md`.
- Avoid speculative TODO comments. Approved work and known gaps belong in `docs/work/BACKLOG.md` and verification artifacts under `docs/work/test-verification/`.

**JSDoc/TSDoc:**
- Not routinely used. Types, props interfaces, narrow helpers, and test names carry the contract in files such as `frontend/src/screens/ReportsPanel.tsx`.
- Add API documentation only when behavior cannot be made clear through types and naming; do not create redundant per-function prose.

## Function Design

**Size:**
- Keep pure business rules in focused functions that can be table-tested, such as `ApplyAccountingEffect`, `checkedAdd`, and `checkedSub` in `backend/internal/finance/transactions.go`.
- Keep orchestration and persistence in repositories/services; do not move ledger, ownership, idempotency, or optimistic-version rules into UI-only guards. The expected boundary is demonstrated by `backend/internal/finance/repository.go` and `backend/internal/sync/service.go`.
- Extract frontend transformations and date semantics into pure exported helpers when they need direct proof, as with `buildTransactionInput` and `calendarDateInHoChiMinh` in `frontend/src/app/transactionInput.ts`.

**Parameters:**
- Use input structs for multi-field Go commands, and normalize/validate at the domain boundary: `CreateTransactionInput` in `backend/internal/finance/types.go` and `ValidateCreateTransaction` in `backend/internal/finance/transactions.go`.
- Pass `context.Context` first to Go repository/service methods, followed by the authenticated owner and command input, as exercised in `backend/internal/finance/repository_test.go`.
- Use typed object parameters for TypeScript helpers with several related values, as in `buildTransactionInput` in `frontend/src/app/transactionInput.ts`.
- Inject infrastructure only where it improves determinism or portability, such as optional `IDBFactory` parameters in `frontend/src/offline/db.ts`.

**Return Values:**
- Return `(value, error)` in Go and use zero-value results on failure. Preserve sentinel identity through wrapping.
- Return explicit frontend state rather than ambiguous falsy placeholders; offline initialization returns an `OfflineSnapshot` with a mode in `frontend/src/offline/db.ts`.
- Use `null` deliberately for absent loaded domain data and distinguish it from empty arrays when loading state matters, as tested in `frontend/src/screens/BudgetsScreen.test.tsx`.

## Module Design

**Exports:**
- Export only public domain/service/component entry points; keep calculation and storage primitives private unless tests or callers require a stable contract. Examples: exported `ApplyAccountingEffect` and private `checkedAdd` in `backend/internal/finance/transactions.go`.
- Prefer named exports for frontend components, helpers, types, and interfaces. Examples include `ReportsPanel`, `ReportsPanelProps`, and `buildTransactionInput`.
- Keep types near the domain they describe: backend types in package-level `types.go` files such as `backend/internal/finance/types.go`; offline frontend types in `frontend/src/offline/types.ts`.

**Barrel Files:**
- Barrel exports are not an established pattern. Import directly from the owning module, such as `./transactionInput` and `./ReportsPanel` in their co-located tests.
- Add new code to the nearest domain module rather than creating a broad shared index. Shared visual primitives belong in `frontend/src/components/`, domain clients/helpers in `frontend/src/app/`, offline persistence/sync in `frontend/src/offline/`, and backend logic in the matching `backend/internal/<domain>/` package.

---

*Convention analysis: 2026-09-10*
