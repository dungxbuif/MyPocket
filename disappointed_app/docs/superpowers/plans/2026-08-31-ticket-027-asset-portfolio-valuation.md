# TICKET-027 Asset Portfolio and Market Valuation Implementation Plan

**Status:** Ready; D-01 through D-04 approved by owner on 2026-08-31.

**Goal:** Deliver a mobile-first, offline-capable asset portfolio that preserves buy/sell history and market prices, calculates moving-average cost plus realized/unrealized P&L, and remains isolated from wallet accounting.

**Requirement:** REQ-F-018

**Ticket:** `docs/work/tickets/TICKET-027-asset-portfolio-valuation.md`

**ADR:** `docs/decisions/ADR-006-separate-asset-portfolio-valuation.md`

## Global Constraints

- D-01 through D-04 are approved: hybrid automatic/manual prices, moving-average buy/sell accounting, three separate dashboard totals, and offline manual mutations.
- Every persisted row and query is scoped by authenticated `user_id`.
- Quantities use decimal strings at the API boundary and PostgreSQL `numeric(30,12)`; never JavaScript or Go binary floating point for authoritative calculations.
- VND values remain integer `bigint`; Go owns cost basis, valuation, P&L, and rounding formulas.
- Portfolio writes never mutate wallets or confirmed finance transactions.
- Manual and provider price history is append-only and source-tagged; missing price is `null`, never zero.
- Existing offline outbox, idempotency, version, tombstone, and explicit-conflict behavior is reused.
- Preserve compact Tailwind mobile styling and liquid-glass iPhone navigation/sheet behavior.

## Wave 1: Domain And Persistence

- [x] **Task 1: Add schema and portfolio domain tests first**
  - Files: next backend migration, `backend/internal/portfolio/*_test.go`.
  - Define tables, indexes, checks, same-user ownership constraints, archive fields, optimistic versions, ordered trade replay, append-only price rules, and provider quote uniqueness.
  - Add failing tests for quantity precision, type/unit validation, moving-average buy/sell cost, realized/unrealized P&L, oversell rejection, historical correction recomputation, fees, latest price selection, VND rounding, missing price, zero cost basis, archive behavior, and cross-user isolation.
  - Verify: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio -count=1` initially fails for the intended missing implementation.

- [x] **Task 2: Implement portfolio domain and PostgreSQL repository**
  - Files: `backend/internal/portfolio`, `backend/internal/platform/db`, migration registration.
  - Implement position/trade/price commands and queries, authoritative ordered-ledger summaries, transactional writes, idempotent retry hooks, archive/recomputation semantics, and bounded history.
  - Verify: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio -count=1`.
  - PostgreSQL proof: run the repository suite from an empty migrated test database and verify no wallet/transaction row changes.

## Wave 2: API And Sync

- [ ] **Task 3: Add authenticated portfolio HTTP contracts**
  - Depends on: Task 2.
  - Files: portfolio HTTP handler, router, API contract tests.
  - Implement position CRUD/archive, buy/sell create/correct/archive, manual price append/history, and portfolio summary routes.
  - Enforce CSRF, idempotency keys, decimal-string parsing, user ownership, stable error codes, pagination, and correlation IDs.
  - Verify: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'Asset|Portfolio' -count=1`.

- [ ] **Task 4: Extend sync and IndexedDB portfolio storage**
  - Depends on: Task 3.
  - Files: backend sync entity handling, `frontend/src/offline`, frontend portfolio API/types.
  - Add position/trade/manual-price entity snapshots, client UUIDs, outbox mutations, tombstones, bounded price cache, reconnect drain, conflict resolution, and logout cleanup.
  - Preserve existing wallet/category/transaction stores and migrations.
  - Verify backend: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/sync -count=1`.
  - Verify frontend: `rtk npm test -- --run src/offline` from `frontend/`.

## Wave 3: Mobile PWA And Dashboard

- [ ] **Task 5: Implement provider registry and leased price-refresh job**
  - Depends on: Tasks 2 and 3.
  - Files: `backend/internal/portfolio` provider interface/adapters, `backend/internal/worker`, configuration, job tests.
  - Resolve provider mappings for positions in `pricing_mode=automatic`, fetch and normalize quotes, append idempotent snapshots, apply bounded retry/backoff, preserve last known values, and emit safe stale/error status.
  - Skip `pricing_mode=manual` positions and expose explicit mode switching without leaking provider credentials.
  - Keep manual price entry available for every position regardless of provider configuration.
  - Verify: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio ./internal/worker -run 'Price|Portfolio' -count=1`.

- [ ] **Task 6: Implement compact portfolio UI**
  - Depends on: Tasks 3 and 4.
  - Files: `frontend/src/app/portfolio.ts`, portfolio list/detail/editor components, app routing/state, Tailwind utility classes, component tests.
  - Add the Account `Tài sản` section, overview investment summary, position detail, buy/sell editor, realized/unrealized summaries, manual-price editor, price history, missing/stale/offline/pending/conflict states, and privacy masking.
  - Keep wallets and assets in separate sections and add no sixth bottom-navigation tab.
  - Verify: `rtk npm test -- --run src/app` and test at 390x844 plus desktop viewport.

- [ ] **Task 7: Integrate authoritative combined net-worth summary**
  - Depends on: Tasks 2 and 5.
  - Files: backend analytics/dashboard response, frontend overview summary, analytics regression tests.
  - Return and display `wallet_net_worth_vnd`, `investment_market_value_vnd`, and `combined_net_worth_vnd` separately; exclude missing-price and opted-out positions with explicit counts.
  - Verify: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/analytics -run 'Portfolio|NetWorth' -count=1`.

## Wave 4: End-To-End Proof And Reconciliation

- [ ] **Task 8: Add asset portfolio E2E/UAT fixtures**
  - Depends on: Tasks 1 through 6.
  - Files: frontend Playwright asset spec, backend fixture helpers, verification artifact.
  - Cover gold with multiple buys and a partial sell, stock, crypto, manual/provider price history, expected realized/unrealized P&L totals, privacy mode, offline mutation replay, explicit conflict, archive retention, and wallet non-interference.
  - Verify direct runtime: start API and `rtk npm run dev -- --host 127.0.0.1 --port 5174`; do not require a Docker frontend build.
  - Run: `rtk npm run test:e2e -- asset-portfolio.spec.ts` from `frontend/`.

- [ ] **Task 9: Full regression and documentation reconciliation**
  - Run backend: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1`.
  - Run frontend: `rtk npm test -- --run` and `rtk npm run build`.
  - Run: `rtk git diff --check`.
  - Update API, architecture, ERD, SDD, DESIGN, validation matrix, verification artifact, ADR status, phase/ticket/backlog/context, and changelog.
  - Complete mobile/desktop UAT and record human sign-off before marking TICKET-027 verified.

## Verification Gate

The ticket cannot move to `verified` unless all of the following are true:

- Known fixture calculations match backend responses and visible UI values.
- Cross-user negative tests pass for positions, trades, prices, and summary.
- Duplicate mutation and stale-version tests pass.
- A portfolio operation is proven not to mutate wallet balances or finance transactions.
- Offline create/manual-price replay happens once and conflict resolution remains explicit.
- Mobile PWA and desktop web UAT pass with no overflow, oversized controls, or privacy leaks.
- Master docs and release evidence are reconciled.
