# PHASE-002 Finance Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build authenticated wallet/category management and exact VND transaction accounting for income, expense, transfer, adjustment, edit, archive, and search.

**Architecture:** The Go backend owns finance validation, ownership checks, idempotency, and atomic wallet balance effects in `backend/internal/finance`. HTTP handlers in `backend/internal/platform/httpapi` expose `/api/v1` finance routes and derive `user_id` from PHASE-001 auth only. The React PWA in `frontend/` renders mobile-first finance screens against these APIs; PHASE-003 will later add IndexedDB sync on top of the version fields created here.

**Tech Stack:** Go 1.24, PostgreSQL/pgx, React 19, TypeScript, Vite, Vitest, Playwright, Docker Compose.

**Spec:** `docs/work/phases/PHASE-002-detail-design.md`

## Global Constraints

- Use `backend/` for all Go commands and `frontend/` for npm commands.
- Every shell command in docs and execution starts with `rtk`.
- No finance endpoint accepts `user_id` from request payloads.
- VND amounts are positive integers; transaction type determines balance signs.
- State-changing finance requests require `X-CSRF-Token` and `Idempotency-Key` when retryable.
- Every user-owned query and mutation is scoped by authenticated user ID.
- System categories are locked; user categories are archiveable, not hard-deleted.
- Transfer source and destination wallets must differ and belong to the authenticated user.
- Do not implement offline sync, analytics aggregation, OCR, AI, or recurring automation in PHASE-002.

---

### Task 1: Finance Migration And Seed Baseline

**Files:**
- Create: `backend/migrations/0002_phase002_finance_core.sql`
- Modify: `backend/internal/platform/db/migrate_test.go`
- Modify: `docs/architecture/ERD.md` during reconciliation

**Interfaces:**
- Produces tables: `wallets`, `categories`, `wallet_category_settings`, `transactions`, `finance_idempotency_keys`, `receipt_objects`.
- Produces stable system category keys consumed by Task 2 and frontend fixtures.

- [x] **Step 1: Write the failing migration tests**

Add tests in `backend/internal/platform/db/migrate_test.go`:

```go
func TestPhase002FinanceTablesAndSeeds(t *testing.T) {
    conn := openTestDB(t)
    runMigrations(t, conn)

    assertTableColumns(t, conn, "wallets", []string{"id", "user_id", "name", "type", "balance_vnd", "include_in_total", "is_default_ai", "archived_at", "version"})
    assertTableColumns(t, conn, "categories", []string{"id", "user_id", "parent_id", "kind", "name", "system_key", "is_system", "archived_at"})
    assertTableColumns(t, conn, "transactions", []string{"id", "user_id", "type", "source_wallet_id", "destination_wallet_id", "category_id", "amount_vnd", "occurred_at", "excluded_from_reports", "archived_at", "version"})

    var count int
    err := conn.QueryRow(`SELECT count(*) FROM categories WHERE is_system AND system_key IN ('expense_food', 'income_salary', 'debt_loan')`).Scan(&count)
    if err != nil {
        t.Fatal(err)
    }
    if count != 3 {
        t.Fatalf("expected required Vietnamese system categories, got %d", count)
    }
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db -run TestPhase002FinanceTablesAndSeeds`

Expected: FAIL because PHASE-002 tables do not exist.

- [x] **Step 3: Add migration**

Create `0002_phase002_finance_core.sql` with:

```sql
CREATE TABLE wallets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name text NOT NULL,
  type text NOT NULL CHECK (type IN ('cash', 'bank', 'credit', 'e_wallet', 'savings', 'debt')),
  balance_vnd bigint NOT NULL DEFAULT 0,
  include_in_total boolean NOT NULL DEFAULT true,
  is_default_ai boolean NOT NULL DEFAULT false,
  credit_limit_vnd bigint,
  statement_day integer CHECK (statement_day BETWEEN 1 AND 31),
  payment_due_day integer CHECK (payment_due_day BETWEEN 1 AND 31),
  archived_at timestamptz,
  version bigint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX wallets_one_default_ai_per_user ON wallets(user_id) WHERE is_default_ai AND archived_at IS NULL;
```

Include equivalent DDL for categories, wallet category settings, transactions, idempotency keys, and receipt objects per the detail design.

- [x] **Step 4: Seed Vietnamese system categories**

Insert stable `system_key` rows including:

```sql
INSERT INTO categories (kind, name, system_key, is_system)
VALUES
  ('expense', 'Ăn uống', 'expense_food', true),
  ('expense', 'Mua sắm', 'expense_shopping', true),
  ('expense', 'Di chuyển', 'expense_transport', true),
  ('income', 'Lương', 'income_salary', true),
  ('income', 'Thưởng', 'income_bonus', true),
  ('debt', 'Vay nợ', 'debt_loan', true)
ON CONFLICT (system_key) DO UPDATE SET name = EXCLUDED.name;
```

- [x] **Step 5: Run migration tests**

Run: `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/platform/db`

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
rtk git add backend/migrations backend/internal/platform/db docs/architecture/ERD.md
rtk git commit -m "feat: add finance core migrations"
```

---

### Task 2: Wallet And Category Service

**Files:**
- Create: `backend/internal/finance/types.go`
- Create: `backend/internal/finance/repository.go`
- Create: `backend/internal/finance/wallets.go`
- Create: `backend/internal/finance/categories.go`
- Create: `backend/internal/finance/wallets_test.go`
- Create: `backend/internal/finance/categories_test.go`

**Interfaces:**
- Produces `finance.Service` with `ListWallets`, `CreateWallet`, `UpdateWallet`, `ListCategories`, `CreateCategory`, `UpdateCategory`, `SetWalletCategoryActive`.
- Consumes authenticated user IDs from HTTP handlers in Task 4.

- [x] **Step 1: Write failing unit tests**

Test wallet validation:

```go
func TestCreateWalletRejectsCreditMetadataForCashWallet(t *testing.T) {
    _, err := finance.ValidateCreateWallet(finance.CreateWalletInput{
        Name: "Tiền mặt", Type: finance.WalletCash, CreditLimitVND: ptrMoney(10_000_000),
    })
    if !errors.Is(err, finance.ErrValidation) {
        t.Fatalf("expected validation error, got %v", err)
    }
}
```

Test category depth/lock behavior:

```go
func TestSystemCategoryCannotBeRenamed(t *testing.T) {
    err := finance.ValidateCategoryUpdate(finance.Category{IsSystem: true}, finance.UpdateCategoryInput{Name: "Tên mới"})
    if !errors.Is(err, finance.ErrSystemCategoryLocked) {
        t.Fatalf("expected system lock, got %v", err)
    }
}
```

- [x] **Step 2: Run RED**

Run: `rtk go test ./internal/finance`

Expected: FAIL because `backend/internal/finance` does not exist.

- [x] **Step 3: Implement validation and repository methods**

Create focused types and methods:

```go
type WalletType string
const (
    WalletCash WalletType = "cash"
    WalletBank WalletType = "bank"
    WalletCredit WalletType = "credit"
)

type Service struct { repo Repository }

func ValidateCreateWallet(input CreateWalletInput) (CreateWalletInput, error)
func ValidateCategoryUpdate(category Category, input UpdateCategoryInput) error
```

- [x] **Step 4: Add PostgreSQL integration tests**

Cover:
- user A cannot list/update user B wallet
- at most one active default AI wallet per user
- category depth above two is rejected
- wallet/category activation is unique per wallet/category

- [x] **Step 5: Run tests**

Run: `rtk go test ./internal/finance`

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
rtk git add backend/internal/finance
rtk git commit -m "feat: add wallet and category domain"
```

---

### Task 3: Transaction Accounting And Idempotency

**Files:**
- Create: `backend/internal/finance/accounting.go`
- Create: `backend/internal/finance/transactions.go`
- Create: `backend/internal/finance/transactions_test.go`
- Modify: `backend/internal/finance/repository.go`

**Interfaces:**
- Produces `CreateTransaction`, `UpdateTransaction`, `ArchiveTransaction`, and `ListTransactions`.
- Produces idempotency response storage consumed by HTTP handlers.

- [x] **Step 1: Write failing accounting effect tests**

Add table-driven tests:

```go
func TestAccountingEffects(t *testing.T) {
    tests := []struct {
        name string
        txType finance.TransactionType
        amount int64
        sourceBefore int64
        destinationBefore int64
        wantSource int64
        wantDestination int64
    }{
        {"income increases source", finance.TransactionIncome, 100_000, 1_000_000, 0, 1_100_000, 0},
        {"expense decreases source", finance.TransactionExpense, 40_000, 1_000_000, 0, 960_000, 0},
        {"transfer moves money", finance.TransactionTransfer, 250_000, 1_000_000, 100_000, 750_000, 350_000},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := finance.ApplyEffect(tt.txType, tt.amount, tt.sourceBefore, tt.destinationBefore)
            if got.SourceBalanceVND != tt.wantSource || got.DestinationBalanceVND != tt.wantDestination {
                t.Fatalf("wrong effect: %#v", got)
            }
        })
    }
}
```

- [x] **Step 2: Run RED**

Run: `rtk go test ./internal/finance -run TestAccountingEffects`

Expected: FAIL because accounting functions are missing.

- [x] **Step 3: Implement accounting and transaction service**

Accounting effects, create transaction persistence, edit/archive reversal, and repository search are implemented and verified. HTTP API and mobile transaction workflows remain pending in later tasks.

Rules:
- amount must be positive
- income requires source wallet and income category
- expense requires source wallet and expense category
- transfer requires source and destination wallets, no same-wallet transfer
- adjustment stores target balance as amount and computes delta internally
- edit/archive reverses old effect before applying new state

- [x] **Step 4: Add integration tests**

Integration tests now cover income, expense, transfer, adjustment, duplicate idempotent replay, idempotency conflict rejection, inactive category rejection, cross-user wallet rejection, wallet balance/version updates, edit reversal/reapply, archive reversal once, and search filters.

Cover:
- create income/expense/transfer/adjustment updates balances in one DB transaction
- duplicate `Idempotency-Key` returns same stored response and does not double-change balances
- edit reverses old effect and applies new effect
- archive reverses effect once and sets `archived_at`
- search filters remain user-scoped

- [x] **Step 5: Run tests**

Run: `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/finance`

Expected: PASS.

- [x] **Step 6: Commit**

Run:

```bash
rtk git add backend/internal/finance
rtk git commit -m "feat: add transaction accounting engine"
```

---

### Task 4: Finance HTTP API

**Files:**
- Create: `backend/internal/platform/httpapi/finance.go`
- Create: `backend/internal/platform/httpapi/finance_test.go`
- Modify: `backend/internal/platform/httpapi/router.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `docs/architecture/API.md`

**Interfaces:**
- Consumes `finance.Service`.
- Produces `/api/v1/wallets`, `/api/v1/categories`, `/api/v1/wallets/{wallet_id}/categories/{category_id}`, and `/api/v1/transactions`.

- [x] **Step 1: Write failing route tests**

Add tests:
- unauthenticated `GET /api/v1/wallets` returns `AUTH_REQUIRED`
- authenticated `POST /api/v1/wallets` without CSRF returns `CSRF_REQUIRED`
- authenticated create wallet returns stable JSON with wallet ID and balance
- duplicate transaction idempotency key returns same JSON and no duplicate balance effect

- [x] **Step 2: Run RED**

Run: `rtk go test ./internal/platform/httpapi -run 'TestFinance'`

Expected: FAIL because finance routes are missing.

- [x] **Step 3: Implement handlers**

Handlers parse JSON, call finance service, map finance errors to:
- `VALIDATION_FAILED`
- `AUTH_REQUIRED`
- `FORBIDDEN`
- `NOT_FOUND`
- `IDEMPOTENT_REPLAY`
- `INTERNAL_RETRYABLE`

- [x] **Step 4: Wire API main**

Instantiate `finance.NewService(finance.NewRepository(conn))` and add it to `httpapi.Dependencies`.

- [x] **Step 5: Run backend tests**

Run: `rtk go test ./...`

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
rtk git add backend/internal/platform/httpapi backend/cmd/api/main.go docs/architecture/API.md
rtk git commit -m "feat: expose finance API"
```

---

### Task 5: Mobile Wallet, Category, And Transaction UI

**Files:**
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/app/apiClient.ts`
- Create: `frontend/src/app/finance.ts`
- Modify: `frontend/src/styles.css`
- Modify: `frontend/src/app/App.test.tsx`
- Create: `frontend/e2e/finance-core.spec.ts`

**Interfaces:**
- Consumes finance API from Task 4.
- Produces mobile workflows matching `design/DESIGN.md`.

- [x] **Step 1: Write failing frontend tests**

Add Vitest checks:
- wallet totals render from API fixture data
- category activation control is visible in wallet detail
- quick add submits expense payload with VND integer amount
- transaction search filters rendered list

- [x] **Step 2: Run RED**

Run: `rtk npm test -- --run`

Expected: FAIL because finance client/UI does not exist.

- [x] **Step 3: Implement finance client and UI**

Create typed calls:

```ts
export async function listWallets(): Promise<Wallet[]>
export async function listCategories(): Promise<Category[]>
export async function listTransactions(filters: TransactionFilters): Promise<Transaction[]>
export async function createTransaction(input: CreateTransactionInput): Promise<Transaction>
```

Wire screens:
- overview wallet card uses API wallets
- transaction tab loads/searches transactions
- add sheet submits income/expense/transfer/adjustment
- budget tab remains placeholder until PHASE-004/005

- [ ] **Step 4: Add Playwright E2E**

Test fixture login, create wallet/category if needed, add expense, verify wallet balance changes, search transaction, edit/archive transaction.

- [x] **Step 5: Run frontend verification**

Run:

```bash
rtk npm test -- --run
rtk npm run build
rtk npm run test:e2e
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
rtk git add frontend
rtk git commit -m "feat: add mobile finance workflows"
```

---

### Task 6: PHASE-002 Reconciliation And Verification

**Files:**
- Create: `docs/work/test-verification/PHASE-002-finance-core.md`
- Modify: `docs/work/tickets/TICKET-005-wallet-category-domain.md`
- Modify: `docs/work/tickets/TICKET-006-transaction-accounting-engine.md`
- Modify: `docs/work/tickets/TICKET-007-vietnamese-seeds-receipt-metadata.md`
- Modify: `docs/work/phases/PHASE-002-finance-core.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/releases/CHANGELOG.md`

**Interfaces:**
- Consumes all code and test results from Tasks 1-5.
- Produces phase-level verification evidence and next queue state for PHASE-003.

- [ ] **Step 1: Run full verification**

Run:

```bash
rtk go test ./...
rtk npm test -- --run
rtk npm run build
rtk npm run test:e2e
rtk docker compose config
rtk docker compose up -d --build
rtk bash scripts/smoke-platform.sh
```

Expected: all commands pass.

- [ ] **Step 2: Record evidence**

Create `docs/work/test-verification/PHASE-002-finance-core.md` with each command, result, and focused notes.

- [ ] **Step 3: Reconcile docs**

Update:
- API contracts for concrete finance payloads and errors
- ERD implemented schema columns and constraints
- Validation rows for REQ-F-002, REQ-F-003, REQ-F-011, REQ-NF-002, REQ-NF-005
- Ticket completion checklists
- Backlog next artifact to PHASE-003 offline sync
- Context current focus to PHASE-003 after review
- Changelog user-facing finance entries

- [ ] **Step 4: Commit**

Run:

```bash
rtk git add docs
rtk git commit -m "docs: verify finance core phase"
```

## Self-Review

- Spec coverage: TICKET-005 covers wallets/categories; TICKET-006 covers transaction accounting, idempotency, search, edit, archive; TICKET-007 covers Vietnamese seeds and receipt metadata; Task 6 covers docs and validation.
- Placeholder scan: no `TBD`, `TODO`, or `implement later` placeholders are used.
- Type consistency: service names and route families are consistent across tasks.
- Known gap: PHASE-003 offline sync is intentionally excluded; PHASE-002 only creates version/idempotency foundations that PHASE-003 will consume.
