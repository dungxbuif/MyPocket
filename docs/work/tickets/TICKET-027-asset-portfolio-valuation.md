---
artifact_type: ticket
id: TICKET-027
status: in_review
owner: human
priority: high
lane: high-risk
human_fields:
  - title
  - priority
  - acceptance_criteria
  - scope
  - approval
ai_fields:
  - impacted_areas
  - test_expectations
  - verification_results
  - docs_review
  - context_updates
shared_fields:
  - status
  - trace
  - small_task_exemption
trace:
  backlog_item: BL-005
  requirement: [REQ-F-018, REQ-NF-001, REQ-NF-002, REQ-NF-003, REQ-NF-005, REQ-NF-007]
  phase: PHASE-005
  detail_design: ../phases/PHASE-005-asset-portfolio-detail-design.md
  implementation_plan: ../../superpowers/plans/2026-08-31-ticket-027-asset-portfolio-valuation.md
  test_verification: ../test-verification/TICKET-027-asset-portfolio-valuation.md
  validation_matrix: ../VALIDATION_MATRIX.md
  docs_review: ../DOCS-REVIEW-TICKET-027.md
  adrs: [../../decisions/ADR-006-separate-asset-portfolio-valuation.md]
  release_notes: ../../releases/CHANGELOG.md
---

# Ticket: TICKET-027 Asset Portfolio and Market Valuation

## Status

- Status: in_review
- Type: feature
- Priority: high
- Phase: PHASE-005 extension, sequenced immediately after TICKET-017
- Approval: approved directly by owner on 2026-08-31

## Context

MyPocket currently represents cash, bank, credit, savings, and debt as VND wallet balances. It cannot preserve the purchase price, current market price, price history, or unrealized profit/loss of gold, stocks, crypto, or foreign currency.

This ticket adds a separate user-owned investment portfolio. Portfolio positions are not wallets and do not silently create or change finance transactions.

### In Scope

- Asset types: `gold`, `stock`, `crypto`, `foreign_currency`, and `other`.
- User-owned asset positions with symbol/name, unit, reporting currency, and inclusion preference.
- Ordered buy/sell trades recording quantity, VND unit price, fee, and occurrence time.
- Timestamped price snapshots with source metadata from both user-entered manual prices and scheduled provider adapters.
- A leased backend refresh job for eligible provider-mapped positions; unsupported/unconfigured assets remain manual-only.
- Current quantity, moving-average cost basis, market value, realized P&L, unrealized P&L, and unrealized P&L percentage.
- Portfolio list/detail, create asset, buy/sell, update current price, archive, privacy masking, empty/error/offline/conflict states.
- Read/write offline support through the existing IndexedDB outbox and explicit conflict model.
- Dashboard separation between wallet net worth, investment market value, and combined net worth.

### Out Of Scope

- Brokerage execution/custody, bank/broker account linking, tax accounting, FIFO/tax-lot selection, dividends, and staking.
- Currency conversion between arbitrary reporting currencies; all accepted prices and totals use integer VND.
- Hard-coding one market-data vendor into the portfolio domain; vendors must sit behind adapters.
- Price prediction, investment advice, recommendations, or AI-generated buy/sell signals.

## Acceptance Criteria

- [ ] A signed-in user can create an asset position for gold, stock, crypto, foreign currency, or other; another user cannot read or mutate it.
- [ ] A position supports ordered buy and sell trades with quantity, VND unit price, fee, and occurrence timestamp.
- [ ] The backend applies moving weighted-average cost, rejects overselling, and recalculates later derived values after a historical trade correction/archive.
- [ ] A user can append a manual price snapshot without overwriting earlier snapshots; the latest valid snapshot becomes current price.
- [ ] A leased backend job refreshes eligible provider-mapped assets, appends idempotent source-tagged snapshots, and preserves the last known price when a provider fails.
- [ ] Manual price entry remains available even when automatic refresh is configured and never exposes provider credentials to the client.
- [ ] Each position exposes `automatic` or `manual` pricing mode; the job skips manual positions, and the UI always shows the latest price source and observation time.
- [ ] The backend calculates quantity, cost basis, market value, realized P&L, absolute unrealized P&L, and percentage unrealized P&L using documented decimal and rounding rules.
- [ ] Zero quantity, zero cost basis, missing current price, invalid negative values, and stale-version writes return explicit non-misleading states or stable errors.
- [ ] Wallet balances and confirmed transaction history remain unchanged when assets, trades, or price snapshots are created or edited.
- [ ] The dashboard shows wallet net worth and investment market value separately; combined net worth includes only positions with `include_in_net_worth=true`.
- [ ] Portfolio screens are compact mobile-first PWA surfaces, preserve balance privacy, and work on desktop without oversized controls.
- [ ] Cached portfolio data remains readable offline; supported offline mutations queue once, display pending state, and use explicit conflict resolution after reconnect.
- [ ] Archiving a position hides it from active totals without deleting its trades or price history.
- [ ] UAT is required on mobile PWA and desktop web using at least one gold, one stock, and one crypto fixture with known expected totals.

## Small Task Exemption

- Small task exemption: no
- Reason: Adds a new financial domain, PostgreSQL schema, public API, sync entities, analytics formulas, and user-facing screens.
- Impact checked: API=yes, DB=yes, Security=yes, Runtime=yes, Standards=no

## Impacted Areas

- Code: new `backend/internal/portfolio`, portfolio HTTP handlers, analytics aggregation, sync entity handling, frontend portfolio client/components, IndexedDB/outbox/conflicts.
- Requirements docs: REQ-F-018 accepted after direct human approval.
- Architecture docs: add the portfolio module and wallet/portfolio accounting boundary.
- API docs: add asset, trade, price-history, and portfolio-summary contracts.
- ERD/data docs: add `asset_positions`, `asset_trades`, and `asset_price_history`.
- Decisions: ADR-006 records why market-valued assets remain separate from wallet accounting.

## Detail Design

- Required: yes
- Link: [PHASE-005 asset portfolio detail design](../phases/PHASE-005-asset-portfolio-detail-design.md)
- Approval: approved

## Test Expectations

- Unit: decimal quantity rules, ordered buy/sell replay, weighted cost basis, oversell rejection, fees, VND rounding, latest-price selection, realized/unrealized P&L, zero/missing states.
- Integration: migrations, ownership isolation, append-only price history, trade recomputation, archive retention, optimistic versions, finance-transaction non-interference.
- E2E: create position, buy/sell, update current price, verify realized/unrealized totals, offline queue/reconnect, privacy masking.
- UAT: required on iPhone-sized PWA and desktop web.
- Manual/platform: verify installed-PWA offline reload and direct Vite test runtime.
- Docs review: requirements, architecture, API, ERD, ADR, validation matrix, context, backlog, and changelog.

## Verification Results

- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio -count=1 -v`
- Result: pass; pure domain proof passed, repository integration cases skipped without `MYPOCKET_TEST_DATABASE_URL`
- Command: `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/portfolio -count=1 -v`
- Result: pass; real PostgreSQL proof covered schema, ownership, moving-average replay, manual append-only prices, oversell rejection, historical correction recomputation, archive retention, and wallet non-interference
- Command: `rtk env 'MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable' GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/db -count=1 -v`
- Result: pass; migration chain applies from an empty PostgreSQL schema
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./... -count=1`
- Result: pass; backend regression without integration DB env
- Command: `rtk git diff --check`
- Result: pass
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/config ./internal/platform/httpapi -run 'Load|OAuth|CurrentUser|CORS|Logout' -count=1`
- Result: pass; login whitelist config and callback rejection are covered
- Command: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/analytics ./internal/platform/httpapi ./cmd/api -count=1`
- Result: pass; portfolio route wiring and dashboard investment totals compile
- Command: `rtk npm test -- --run src/app`
- Result: pass; app component regression includes the added portfolio client/account surface
- Command: `rtk npm run build`
- Result: pass; frontend production build succeeds
- Notes: Schema/domain/repository, REST API, dashboard total fields, Account-tab portfolio UI, IndexedDB asset cache/outbox, sync asset replay, static provider price refresh worker, and login whitelist are implemented. Human UAT and optional browser visual proof remain pending.

## Fix/Test Attempt Log

- Same-path failure attempts: 0 / 3
- Total fix/test cycles: 1 / 5
- Blocked by loop guard: no
- Human/design input needed: none; D-01 through D-04 are approved.

## UAT

- Required: yes
- Expected behavior: asset quantity, weighted cost, current value, realized/unrealized P&L match known fixture calculations without changing wallet balances.
- Verified behavior: automated backend/frontend proof passed; human UAT pending
- Sign-off: pending

## Completion Checklist

- [x] Detail design approved
- [x] Implementation complete
- [x] Wave 1 schema/domain/repository tests run and recorded
- [x] Validation matrix updated with Wave 1 evidence
- [ ] UAT completed
- [ ] Master docs reconciled
- [ ] Docs review completed
- [ ] ADR accepted or amended
- [ ] `docs/CONTEXT.md` updated after implementation
- [ ] `docs/work/BACKLOG.md` status updated
- [ ] Release notes updated
