# Budget, Recurring, and Agent Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete budget transaction assignment, recurring schedule controls, and durable Agent chat/action cards with full business verification.

**Architecture:** Add schema support at the lowest shared boundaries, enforce ownership/type invariants in backend repositories, then expose compatible HTTP and frontend flows. Keep default behavior review-first; auto-post and Agent actions are explicit and idempotent. Reconcile public/internal docs only after tests prove actual behavior.

**Tech Stack:** Go backend with PostgreSQL migrations and repository integration tests, React/Vitest frontend, Playwright E2E, Docusaurus docs.

**Spec:** `docs/superpowers/specs/2026-09-12-budget-recurring-agent-completion-design.md`

## Global Constraints

- Wallet behavior model A is deferred.
- No bank webhook/bank integration.
- No UAT or production deploy in this scope.
- Use TDD: write failing test, verify RED, implement minimal code, verify GREEN.
- Preserve existing unrelated dirty changes, especially `AGENTS.md`.
- Every user-owned lookup must be scoped by authenticated `user_id`.
- API-key callers must get the same business behavior as cookie callers where the endpoint is documented as API-key capable.

---

### Task 1: Budget transaction-level assignment backend

**Files:**
- Create: `backend/migrations/0018_budget_recurring_agent_completion.sql`
- Modify: `backend/internal/finance/types.go`
- Modify: `backend/internal/finance/transactions.go`
- Modify: `backend/internal/finance/repository.go`
- Modify: `backend/internal/platform/httpapi/finance.go`
- Modify: `backend/internal/planning/types.go`
- Modify: `backend/internal/planning/repository.go`
- Test: `backend/internal/planning/budget_assignment_test.go`
- Test: `backend/internal/finance/repository_test.go`
- Test: `backend/internal/platform/httpapi/finance_test.go`

**Interfaces:**
- Produces: `budget_id` on finance transactions and draft confirmation path.
- Consumes: existing budget rows from planning schema.

- [ ] **Step 1: Write RED integration tests**

Add tests proving:

```go
// Break caught: budget progress still uses category scope instead of explicit transaction assignment.
func TestBudgetProgressCountsOnlyAssignedExpenseTransactions(t *testing.T) { ... }

// Break caught: non-expense transactions can be assigned to budgets.
func TestBudgetIDRejectedForIncomeTransferAndAdjustment(t *testing.T) { ... }

// Break caught: cross-user budget IDs can be used on another user's transaction.
func TestTransactionBudgetIDMustBelongToOwner(t *testing.T) { ... }
```

- [ ] **Step 2: Run RED**

Run:

```bash
cd backend
MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/planning ./internal/finance ./internal/platform/httpapi -run 'Budget|Transaction' -count=1
```

Expected: fail because `budget_id` field/column and transaction-level progress do not exist.

- [ ] **Step 3: Implement schema and finance/planning support**

Add `budget_id` columns and indexes, wire request structs, domain structs, validation, insert/update/select scan, budget ownership checks, and progress query using `transactions.budget_id`.

- [ ] **Step 4: Run GREEN focused tests**

Run the same command. Expected: pass.

### Task 2: Budget assignment frontend

**Files:**
- Modify: `frontend/src/app/finance.ts`
- Modify: `frontend/src/app/planning.ts`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/screens/BudgetsScreen.tsx`
- Test: `frontend/src/app/App.test.tsx`
- Test: `frontend/e2e/business-core.spec.ts`

**Interfaces:**
- Consumes: `Transaction.budget_id`, `BudgetProgress`.
- Produces: transaction create/edit payloads with optional `budget_id`.

- [ ] **Step 1: Write RED frontend tests**

Add Vitest and E2E expectations for selecting a budget on expense create/edit and seeing budget progress change only for assigned expenses.

- [ ] **Step 2: Run RED**

```bash
cd frontend
npm test -- --run src/app/App.test.tsx
npx playwright test e2e/business-core.spec.ts --project=desktop --workers=1
```

Expected: fail because UI/client has no `budget_id` controls.

- [ ] **Step 3: Implement UI/client support**

Expose budget selection for expense create/edit, hide it for income/transfer/adjustment, and keep base/shared components.

- [ ] **Step 4: Run GREEN**

Run the same frontend tests. Expected: pass.

### Task 3: Recurring edit/pause/end/auto-post backend

**Files:**
- Modify: `backend/migrations/0018_budget_recurring_agent_completion.sql`
- Modify: `backend/internal/planning/types.go`
- Modify: `backend/internal/planning/repository.go`
- Modify: `backend/internal/platform/httpapi/planning.go`
- Modify: `backend/internal/platform/httpapi/router.go`
- Test: `backend/internal/planning/recurring_test.go`
- Test: `backend/internal/platform/httpapi/planning_test.go`

**Interfaces:**
- Produces: `UpdateRecurringSchedule`, `PauseRecurringSchedule`, `ResumeRecurringSchedule`.
- Produces routes: `PATCH /recurring-schedules/{id}`, `POST /pause`, `POST /resume`.

- [ ] **Step 1: Write RED tests**

Add tests for versioned edit, pause skip, resume, `ends_at`, explicit `posting_mode=auto_post`, and idempotent auto-post occurrence.

- [ ] **Step 2: Run RED**

```bash
cd backend
MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/planning ./internal/platform/httpapi -run 'Recurring' -count=1
```

Expected: fail because update/pause/resume/auto-post are missing.

- [ ] **Step 3: Implement recurring controls**

Wire schema fields, validation, repository methods, HTTP routes, and worker branching for draft vs auto-post.

- [ ] **Step 4: Run GREEN**

Run the same backend recurring tests. Expected: pass.

### Task 4: Recurring frontend controls

**Files:**
- Modify: `frontend/src/app/planning.ts`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/screens/BudgetsScreen.tsx`
- Test: `frontend/src/app/App.test.tsx`
- Test: `frontend/e2e/business-core.spec.ts`

**Interfaces:**
- Consumes recurring `posting_mode`, `paused_at`, `ends_at`, `budget_id`.
- Produces update/pause/resume calls.

- [ ] **Step 1: Write RED tests**

Add tests proving the schedule sheet can edit, pause/resume, end, choose draft/auto-post, and preserve version conflict behavior.

- [ ] **Step 2: Run RED**

```bash
cd frontend
npm test -- --run src/app/App.test.tsx
npx playwright test e2e/business-core.spec.ts --project=desktop --workers=1
```

- [ ] **Step 3: Implement UI/client support**

Add controls using existing base components; keep archive available.

- [ ] **Step 4: Run GREEN**

Run the same frontend tests. Expected: pass.

### Task 5: Agent durable sessions/history/action cards backend

**Files:**
- Modify: `backend/migrations/0018_budget_recurring_agent_completion.sql`
- Modify: `backend/internal/agent/types.go`
- Modify: `backend/internal/agent/repository.go`
- Modify: `backend/internal/agent/validation.go`
- Modify: `backend/internal/worker/agent.go`
- Modify: `backend/internal/platform/httpapi/agent.go`
- Modify: `backend/internal/platform/httpapi/router.go`
- Test: `backend/internal/agent/repository_test.go`
- Test: `backend/internal/platform/httpapi/agent_test.go`

**Interfaces:**
- Produces: `agent.Session`, `agent.Message`, `SubmitMessage`, `GetSession`.
- Produces action payload shape: `{ "type": "transaction_drafts", "draft_ids": [...] }`.

- [ ] **Step 1: Write RED tests**

Add tests proving singleton intake/advisor sessions, durable messages, assistant action-card message on draft completion, advisor creates no drafts, and cross-user isolation.

- [ ] **Step 2: Run RED**

```bash
cd backend
MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/agent ./internal/worker ./internal/platform/httpapi -run 'Agent' -count=1
```

Expected: fail because session/message tables and API payloads are missing.

- [ ] **Step 3: Implement durable Agent chat**

Add tables, repository methods, service/http wrappers, bounded context loading, and completion-time assistant messages/action cards.

- [ ] **Step 4: Run GREEN**

Run the same backend agent tests. Expected: pass.

### Task 6: Agent frontend chat history/action cards

**Files:**
- Modify: `frontend/src/app/agent.ts`
- Modify: `frontend/src/screens/AgentScreen.tsx`
- Modify: `frontend/src/components/agent/AgentMessage.tsx`
- Test: `frontend/src/app/agent.test.ts`
- Test: `frontend/src/screens/AgentScreen.test.tsx`
- Test: `frontend/e2e/agent.spec.ts`

**Interfaces:**
- Consumes session/message/action payloads.
- Produces user-visible two chat modes with durable history and draft action cards.

- [ ] **Step 1: Write RED tests**

Add tests proving loaded history renders, user messages append, assistant action cards open drafts, and advisor cards never show draft actions.

- [ ] **Step 2: Run RED**

```bash
cd frontend
npm test -- --run src/app/agent.test.ts src/screens/AgentScreen.test.tsx
npx playwright test e2e/agent.spec.ts --project=desktop --workers=1
```

- [ ] **Step 3: Implement UI/client support**

Use existing `Card`, `ActionButton`, and message components; do not add one-off legacy UI.

- [ ] **Step 4: Run GREEN**

Run the same tests. Expected: pass.

### Task 7: Docs/OpenAPI/reusable skill reconciliation

**Files:**
- Modify: `backend/internal/platform/httpapi/openapi.go`
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/ERD.md`
- Modify: `frontend/docs/docs/api/budgets.mdx`
- Modify: `frontend/docs/docs/api/planning.mdx`
- Modify: `frontend/docs/docs/api/transactions.mdx`
- Modify: `frontend/docs/docs/api/agent.mdx`
- Modify: `frontend/docs/docs/erd/overview.mdx`
- Modify: `frontend/docs/docs/erd/relationships.mdx`
- Modify: `frontend/docs/docs/skills/mypocket-api.mdx`
- Modify: `skills/mypocket-api/SKILL.md`
- Modify: `skills/mypocket-api/references/api-quick-reference.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/releases/CHANGELOG.md`
- Create: `docs/work/test-verification/BUDGET-RECURRING-AGENT-COMPLETION-2026-09-12.md`

**Interfaces:**
- Consumes implemented route/schema behavior.
- Produces human and AI-agent-readable current docs.

- [ ] **Step 1: Reconcile docs from code**

Update route tables, schemas, guides, skill references and evidence files. Remove or qualify stale “planned” statements where code is implemented, and mark truly deferred bank/UAT/deploy items clearly.

- [ ] **Step 2: Verify docs build**

```bash
cd frontend/docs
npm run build
```

Expected: pass.

### Task 8: Full business verification

**Files:**
- Test: backend full suite
- Test: frontend full suite/build
- Test: Playwright business suites
- Modify: `docs/work/test-verification/BUDGET-RECURRING-AGENT-COMPLETION-2026-09-12.md`

**Interfaces:**
- Produces completion evidence for handoff.

- [ ] **Step 1: Run backend full integration**

```bash
cd backend
MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket_verify_business_audit?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./... -count=1
```

- [ ] **Step 2: Run frontend unit/build/docs**

```bash
cd frontend
npm test -- --run
npm run build
cd docs
npm run build
```

- [ ] **Step 3: Run important business E2E**

```bash
cd frontend
npx playwright test e2e/business-core.spec.ts e2e/accounting-correctness.spec.ts e2e/agent.spec.ts --project=desktop --project=mobile --project=webkit-mobile --workers=1
```

- [ ] **Step 4: Record evidence and final gaps**

Update the verification document with exact pass/fail outputs and remaining out-of-scope UAT/deploy statements.
