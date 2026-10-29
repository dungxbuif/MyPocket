---
artifact_type: test_verification
id: TICKET-027-asset-portfolio-valuation
status: in_review
owner: shared
trace:
  backlog_item: BL-005
  requirement: [REQ-F-018, REQ-NF-001, REQ-NF-002, REQ-NF-003, REQ-NF-005, REQ-NF-007]
  phase: PHASE-005
  ticket_or_bug: [TICKET-027]
  detail_design: ../phases/PHASE-005-asset-portfolio-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-31-ticket-027-asset-portfolio-valuation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: ../DOCS-REVIEW-TICKET-027.md
  release_notes: ../../releases/CHANGELOG.md
---

# TICKET-027 Asset Portfolio Verification

## Status

- Status: in_review
- Owner: shared
- Implementation evidence: schema, domain formulas, PostgreSQL repository, migration proof, REST routes, dashboard totals, sync entity replay, IndexedDB asset cache/outbox, static provider refresh worker, frontend component tests, and production build passed on 2026-08-31. Human UAT remains pending.

## Planned Commands

| Command | Expected Coverage |
| --- | --- |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio -count=1` | Domain formulas, validation, persistence, ownership, archive, and non-interference. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'Asset|Portfolio' -count=1` | Authenticated API, errors, idempotency, CSRF, and ownership. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/sync -count=1` | Asset entity replay, change feed, versions, and conflict behavior. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/analytics -run 'Portfolio|NetWorth' -count=1` | Separate wallet/investment/combined totals. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/worker -run 'Portfolio|Price' -count=1` | Leased provider refresh, automatic/manual mode selection, idempotency, stale state, and retry behavior. |
| `rtk npm test -- --run src/offline src/app` | IndexedDB/outbox and component/accessibility states. |
| `rtk npm run test:e2e -- asset-portfolio.spec.ts` | Mobile/desktop create, buy/sell, hybrid prices, realized/unrealized P&L, offline replay, privacy, archive. |
| `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1` | Full backend regression. |
| `rtk npm test -- --run && rtk npm run build` | Full frontend regression and production build. |

## Required Fixture

- Gold position with two buys, one partial sell, and one manual current price.
- FPT/HOSE stock position with one buy and two historical price snapshots.
- BTC position using fractional quantity precision.
- One position excluded from combined net worth.
- One position missing a current price.
- One automatic position with a provider fixture and one manual position skipped by the refresh job.
- One rejected oversell and one historical trade correction that proves subsequent ledger recomputation.
- A second user attempting cross-user read and mutation.

## Verification Results

| Date | Command | Result | Coverage |
| --- | --- | --- | --- |
| 2026-08-31 | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio -count=1 -v` | pass | Decimal normalization, moving-average buy/sell replay, fee handling, VND rounding, oversell rejection, missing-price state, zero-cost not-comparable state; PostgreSQL tests skipped in this run because `MYPOCKET_TEST_DATABASE_URL` was not set. |
| 2026-08-31 | `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio -count=1 -v` | pass | Real PostgreSQL proof for `asset_positions`, `asset_trades`, `asset_price_history`, user isolation, moving-average persisted ledger, manual append-only price history, oversell rejection, historical correction recomputation, archive retention, active summary exclusion, and wallet balance/version non-interference. |
| 2026-08-31 | `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/db -count=1 -v` | pass | Empty PostgreSQL schema reset and full migration chain including `0008_phase005_asset_portfolio.sql`. |
| 2026-08-31 | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1` | pass | Backend package regression compile and non-DB tests. |
| 2026-08-31 | `rtk git diff --check` | pass | Whitespace sanity check. |
| 2026-08-31 | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/config ./internal/platform/httpapi -run 'Load|OAuth|CurrentUser|CORS|Logout' -count=1` | pass | Login whitelist config parsing and OAuth callback rejection before user provisioning. |
| 2026-08-31 | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/analytics ./internal/platform/httpapi ./cmd/api -count=1` | pass | Portfolio API route compile/wiring and dashboard investment-total fields. |
| 2026-08-31 | `rtk npm test -- --run src/app` | pass | App component regression with portfolio client/account section. |
| 2026-08-31 | `rtk npm run build` | pass | Frontend production build after asset UI changes. |
| 2026-08-31 | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/sync ./internal/portfolio ./internal/worker ./internal/platform/httpapi ./cmd/api ./cmd/worker -count=1 -v` | pass | Asset sync mutation replay unit proof, portfolio formulas/repository compile, static provider price refresh worker, and API wiring. |
| 2026-08-31 | `rtk npm test -- --run src/app frontend/src/offline` | pass | Frontend app/outbox regression after IndexedDB asset cache, asset tombstones, conflict full-resync, and offline mutation queue support. |
| 2026-08-31 | `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1` | pass | Full backend regression. |
| 2026-08-31 | `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test -p 1 ./internal/portfolio ./internal/sync ./internal/platform/db -count=1 -v` | pass | Real PostgreSQL proof for portfolio tables, sync resync/mutation ledger, and migration chain; `-p 1` avoids concurrent schema reset races across packages. |
| 2026-08-31 | `rtk npm run build` | pass | Final frontend production build after offline asset and pricing-mode UI changes. |
| 2026-08-31 | Node REPL Playwright smoke against `http://127.0.0.1:5174` | blocked by environment | Playwright-managed Chromium was missing and system Chrome headless aborted before page load; automated frontend tests and production build still passed. |

## Remaining Proof

- Mobile/desktop human UAT with one gold, one stock, one crypto fixture and known expected totals.
- Optional browser visual proof once local Playwright/Chrome headless is available.
