# Wallet Behavior Model Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace fake wallet channel types with the behavior enum `basic | goal | credit`, including migration, API/backend validation, frontend wallet creation, docs and verification.

**Architecture:** Keep the API field name `type` to avoid a wider client compatibility break, but redefine its allowed values as behavior values. PostgreSQL migration converts existing rows and enforces behavior-specific metadata constraints. Frontend and tests stop using old pseudo types.

**Tech Stack:** Go backend with PostgreSQL migrations/tests; React TypeScript frontend with Vitest and Playwright; Docusaurus public docs.

**Spec:** `docs/superpowers/specs/2026-09-12-wallet-behavior-model-design.md`

## Global Constraints

- Preserve user balances, transactions, categories, obligations, budgets and sync state.
- Do not remove debt/loan categories or obligation behavior; only remove wallet type `debt`.
- Use TDD: write failing tests before production changes.
- Keep docs synchronized with actual shipped code.
- Scope excludes UAT and deploy.

---

### Task 1: Backend wallet behavior validation and schema

**Files:**
- Modify: `backend/internal/finance/types.go`
- Modify: `backend/internal/finance/wallets.go`
- Modify: `backend/internal/finance/wallets_test.go`
- Modify: `backend/internal/finance/repository.go`
- Create: `backend/migrations/0020_wallet_behavior_model.sql`
- Modify: `backend/internal/platform/db/migrate_test.go`

**Interfaces:**
- Produces: `finance.WalletBasic`, `finance.WalletGoal`, `finance.WalletCredit`.
- Produces: `CreateWalletInput.GoalTargetVND`, `CreateWalletInput.GoalDeadlineOn`, `Wallet.GoalTargetVND`, `Wallet.GoalDeadlineOn`.

- [ ] Write failing backend tests for accepted behavior values, rejected legacy values, goal metadata rules and migration conversion.
- [ ] Run focused backend tests and confirm RED failures come from missing behavior model.
- [ ] Implement constants, validation and repository scan/insert support.
- [ ] Add migration converting old wallet rows and replacing constraints.
- [ ] Run focused backend tests and confirm GREEN.

### Task 2: HTTP/API and frontend wallet create flow

**Files:**
- Modify: `backend/internal/platform/httpapi/finance.go`
- Modify: `backend/internal/platform/httpapi/finance_test.go`
- Modify: `frontend/src/app/finance.ts`
- Modify: `frontend/src/offline/types.ts`
- Modify: `frontend/src/offline/outbox.ts`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/app/App.test.tsx`
- Modify: `frontend/e2e/*.spec.ts`

**Interfaces:**
- Consumes: wallet behavior constants and goal fields from Task 1.
- Produces: wallet create payload `{ name, type, goal_target_vnd?, goal_deadline_on? }`.

- [ ] Write failing HTTP/frontend tests for wallet behavior enum and goal metadata payload.
- [ ] Run focused tests and confirm RED.
- [ ] Update request/response structs, TypeScript types, offline queue and wallet create UI.
- [ ] Replace E2E/API fixtures using old wallet types with `basic`, except behavior-specific `credit`/`goal` cases.
- [ ] Run focused tests and confirm GREEN.

### Task 3: Docs/OpenAPI/skill reconciliation

**Files:**
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/ERD.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/releases/CHANGELOG.md`
- Modify: `frontend/docs/docs/api/wallets.mdx`
- Modify: `frontend/docs/docs/erd/overview.mdx`
- Modify: `frontend/docs/docs/erd/entities.mdx`
- Modify: `skills/mypocket-api/SKILL.md`
- Modify: `skills/mypocket-api/references/api-quick-reference.md`
- Modify: `docs/work/test-verification/BUDGET-RECURRING-AGENT-COMPLETION-2026-09-12.md`

**Interfaces:**
- Consumes: actual behavior implemented by Tasks 1 and 2.
- Produces: no current live docs claiming fake wallet types are supported.

- [ ] Update docs to describe only `basic | goal | credit`.
- [ ] Record migration mapping and verification evidence.
- [ ] Search current docs/API/skill surface for stale fake wallet type claims and correct live-contract docs.

### Task 4: Full verification

**Files:**
- No production edits expected; update verification docs if new evidence is produced.

**Commands:**
- `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./... -count=1` from `backend/`
- `rtk npm test -- --run` from `frontend/`
- `rtk npm run build` from `frontend/`
- `rtk npx playwright test e2e/business-core.spec.ts e2e/planning-automation.spec.ts e2e/accounting-correctness.spec.ts e2e/agent.spec.ts` from `frontend/`
- `rtk npm run build` from `frontend/docs`
- `rtk proxy git diff --check`

- [ ] Run all verification commands.
- [ ] Fix any regressions with RED/GREEN tests where behavior changes are needed.
- [ ] Update verification record with exact commands/results.
