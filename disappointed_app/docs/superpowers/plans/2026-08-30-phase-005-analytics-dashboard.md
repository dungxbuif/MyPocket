# PHASE-005 Analytics and Dashboard Implementation Plan

**Goal:** Complete the mobile PWA navigation/search/dashboard/report experience with trustworthy server-side financial aggregates.

**Spec:** `docs/work/phases/PHASE-005-detail-design.md`

## Global Constraints

- Every shell command starts with `rtk`.
- Backend owns all financial formulas; frontend displays returned values.
- VND stays integer until formatting.
- Report filters use authenticated user ownership and Ho Chi Minh date boundaries.
- Cached analytics are read-only offline and must show stale timestamps.

## Tasks

- [ ] **Task 1: Shared Analytics Query Foundation**
  - Files: `backend/internal/analytics`, HTTP handlers, frontend finance/report clients.
  - Implement normalized filter parsing, user-owned query helpers, `generated_at`, timezone, and data version.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/analytics -run Filter -count=1`

- [ ] **Task 2: PWA Navigation, Search, and Wallet Views**
  - Files: `frontend/src/app`, `frontend/src/styles.css`, search/wallet APIs.
  - Implement five-destination shell completion, global search, wallet detail, offline stale states.
  - Verify with: `rtk npm test -- --run src/app`

- [ ] **Task 3: Overview and Net-Worth Dashboard**
  - Files: `backend/internal/analytics`, dashboard API, overview components.
  - Implement net worth, included wallets, recent transactions, planning summary, privacy masking.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/analytics -run Dashboard -count=1`

- [ ] **Task 4: Reports and Trends**
  - Files: analytics report service, report APIs, chart components.
  - Implement cash-flow, categories, daily average, comparison, cumulative trend, accessible labels.
  - Verify with: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test ./internal/analytics -run 'CashFlow|Category|Daily|Comparison|Cumulative' -count=1`

- [ ] **Task 5: Reconciliation and Release Proof**
  - Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' go test -p 1 ./... -count=1`
  - Run: `rtk npm test -- --run`
  - Run: `rtk npm run build`
  - Run: `rtk npm run test:e2e -- analytics-dashboard.spec.ts`
  - Update verification, tickets, phase, validation matrix, backlog, context, changelog, API, architecture.
