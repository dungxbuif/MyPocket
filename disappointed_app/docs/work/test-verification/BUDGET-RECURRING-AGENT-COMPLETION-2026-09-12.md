---
artifact_type: test_verification
status: in_progress
owner: ai
scope: wallet behavior model, budget transaction assignment, recurring controls, agent history/action cards
release_state: local-only
---

# Wallet, Budget, Recurring, and Agent Completion Verification — 2026-09-12

## Scope

Tracks the owner-approved A/B/C/D slice:

- Wallet behavior model cleanup: `basic`/`goal`/`credit`, remove pseudo wallet types.
- Budget transaction-level assignment.
- Recurring edit/pause/resume/end/auto-post.
- Agent durable chat history/action cards.
- Docs/OpenAPI reconciliation.

Out of scope for this record: bank integration, UAT, production deploy.

## Current evidence

| Command | Result | Notes |
| --- | --- | --- |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/finance ./internal/platform/httpapi ./internal/platform/db -run 'TestCreateWallet\|TestWalletsAPI\|TestMigration0020' -count=1` | pass | RED first failed because `WalletBasic`/`WalletGoal` and goal metadata fields did not exist and migration left old wallet types unchanged; GREEN proves behavior enum validation, HTTP goal metadata, and migration mapping from legacy pseudo types. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/planning -run TestBudgetProgressCountsOnlyExplicitlyAssignedExpenseTransactions -count=1` | pass | RED first failed because `BudgetID` did not exist; GREEN proves explicit expense assignment drives budget progress and invalid/cross-user budget assignment is rejected. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/planning -run TestConfirmTransactionDraftCarriesBudgetAssignmentToConfirmedExpense -count=1` | pass | RED first failed because draft confirmation dropped budget assignment; GREEN confirms draft `budget_id` reaches the confirmed transaction and progress. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/planning -run 'TestRecurringScheduleCanBeUpdatedPausedResumedAndEnded\|TestRecurringScheduleAutoPostCreatesConfirmedTransactionOnce' -count=1` | pass | RED first failed because recurring update/pause/resume/end/auto-post APIs did not exist; GREEN covers backend repository behavior. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/planning ./internal/finance ./internal/platform/httpapi -run 'TestRecurring\|TestBudget\|Test.*Transaction\|TestOpenAPI' -count=1` | pass | Focused affected backend packages pass after HTTP route/stub updates. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./... -count=1` from `backend/` | pass | Full backend suite passed after budget/recurring schema and repository changes. |
| `rtk npm test -- --run src/app/agent.test.ts src/app/App.test.tsx src/screens/BudgetsScreen.test.tsx` from `frontend/` | pass | Focused frontend tests still pass after API type additions. |
| `rtk npm test -- --run src/app/App.test.tsx -t "assigns a selected budget\|edits and pauses"` from `frontend/` | pass | RED first failed because transaction sheet lacked `Ngân sách giao dịch` and schedule edit controls were disabled; GREEN proves create payload sends `budget_id`, recurring PATCH sends `posting_mode`, `ends_at`, `budget_id`, and pause calls `/pause`. |
| `rtk npm test -- --run src/app/App.test.tsx` from `frontend/` | pass | App shell regression: 41 tests pass after transaction/recurring sheet changes. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/agent -run TestAgentIntakePersistsSessionMessagesAndDraftActionCard -count=1` | pass | RED first failed because session/message/action-card types and repository methods were absent; GREEN proves durable intake session messages and `drafts_created` action payload. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/platform/httpapi -run 'TestAgentIntakeEndpointQueuesIntakeKind\|TestAgentIntakeHistoryEndpointReturnsMessages\|TestAgentAdvisorEndpointQueuesAdvisorKind\|TestAgentMessagesQueuesReviewFirstRun' -count=1` | pass | RED first failed because HTTP Agent endpoints returned run-only; GREEN proves session envelope and intake history/action-card response while preserving legacy run route. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/agent ./internal/platform/httpapi -run 'TestAgent\|TestOpenAPI\|TestService' -count=1` | pass | Affected Agent/HTTP tests pass after service/interface updates. |
| `rtk npm test -- --run src/app/agent.test.ts src/app/App.test.tsx` from `frontend/` | pass | Agent client envelope/history tests and App shell tests pass after frontend chat/UI updates. |
| `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./... -count=1` from `backend/` | pass | Final full backend rerun after Agent sessions/messages, wallet behavior migration 0020, OpenAPI descriptors and docs-linked changes. |
| `rtk npm test -- --run` from `frontend/` | pass | Full frontend unit/component suite: 23 files, 184 tests. |
| `rtk npm run build` from `frontend/` | pass | TypeScript `tsc --noEmit` and Vite production build pass. |
| `rtk npx playwright test e2e/business-core.spec.ts e2e/planning-automation.spec.ts e2e/accounting-correctness.spec.ts e2e/agent.spec.ts` from `frontend/` | pass | Business E2E: 24/24 across configured mobile, desktop and webkit-mobile projects. Covers wallet create/edit/archive, transaction create/edit/archive, transfer wallet effects and report exclusion, budget transaction-level assignment, event/debt links, recurring create/edit/pause/auto-post option UI, accounting correctness, and Agent review-first draft action-card flow. |
| `rtk npx playwright test e2e/account-lifecycle.spec.ts e2e/agent-image.spec.ts e2e/financial-feedback.spec.ts e2e/offline-sync.spec.ts` from `frontend/` | pass | Rerun after wallet type and transaction-level budget fixes: 21/21. Proves account lifecycle wallet fixture uses `basic`, Agent image intake uses `/agent/intakes`, budget UI reads explicit `budget_id`, and prior WebKit offline-sync timeout did not reproduce in the focused rerun. |
| `rtk npx playwright test` from `frontend/` | pass | Full browser E2E suite: 105/105 across configured mobile, desktop and webkit-mobile projects after all wallet/budget/Agent E2E fixture updates. |
| `rtk npm run build` from `frontend/docs` | pass | Docusaurus production build passes after public docs/ERD/skill surface changes. |
| `rtk proxy git diff --check` | pass | No whitespace errors. |

## Implemented locally so far

- Added migration `0020_wallet_behavior_model.sql`.
- Replaced pseudo wallet types with behavior values `basic`, `goal`, and `credit`; legacy `cash`, `bank`, `e_wallet`, `debt` migrate to `basic`, `savings` migrates to `goal`, and `credit` remains `credit`.
- Added wallet goal metadata `goal_target_vnd` and `goal_deadline_on`, with backend/DB validation that goal metadata only applies to `goal` wallets and credit metadata only applies to `credit` wallets.
- Frontend wallet create UI now offers only `Cơ bản`, `Mục tiêu`, and `Tín dụng`; selecting `Mục tiêu` can send optional goal target/deadline.
- Added migration `0018_budget_recurring_agent_completion.sql`.
- Added `budget_id` to confirmed transactions, drafts, and recurring schedules.
- Budget progress now counts explicit assigned, report-included expense transactions only.
- Draft confirmation carries `budget_id` to confirmed expense transactions.
- Recurring schedules now have backend support for update, pause, resume, optional `ends_at`, explicit `posting_mode=draft|auto_post`, and budget carry-through in drafts/auto-posted transactions.
- HTTP recurring routes now expose `PATCH /recurring-schedules/{id}`, `POST /pause`, and `POST /resume`.
- Frontend API client types now include `budget_id`, recurring posting mode/end/pause fields, and update/pause/resume helpers.
- Frontend transaction create/edit sheets expose `Ngân sách giao dịch` for expenses and send `budget_id`.
- Frontend recurring sheet now edits existing schedules, supports end date, `draft` vs `auto_post`, budget assignment, pause/resume, archive and save.
- Added migration `0019_agent_sessions_messages.sql`.
- Added durable Agent `agent_sessions`, `agent_messages`, `agent_runs.session_id`, session-run creation, intake history fetch and assistant `drafts_created` action cards.
- Agent HTTP intake/advisor submit now returns `session`, persisted user `message`, and `run`; `GET /agent/intakes/{session_id}` returns durable message history.
- Frontend Agent client understands session envelopes and intake history/action cards; Agent screen uses the two approved chat modes (`intake`, `advisor`) rather than legacy ambiguous modes.
- Public/internal docs updated for wallet behavior model, API route table, ERD, Docusaurus Wallet/Agent/Budget/Planning pages and `skills/mypocket-api` API guidance.

## Remaining required work before handoff

- Scope exclusions still stand: no UAT and no production deployment in this verification record.
