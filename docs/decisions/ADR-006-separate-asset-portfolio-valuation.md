---
artifact_type: adr
id: ADR-006
status: accepted
owner: shared
human_fields:
  - decision_approval
  - final_status
ai_fields:
  - context
  - alternatives_considered
  - consequences
  - linked_work
shared_fields:
  - decision
  - trace
trace:
  requirements: [REQ-F-018, REQ-NF-001, REQ-NF-002]
  phase: PHASE-005
  tickets_or_bugs: [TICKET-027]
  detail_design: ../work/phases/PHASE-005-asset-portfolio-detail-design.md
  master_docs:
    - ../requirements/REQUIREMENTS.md
    - ../architecture/ARCHITECTURE.md
    - ../architecture/API.md
    - ../architecture/ERD.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-006: Separate Asset Portfolio Valuation From Wallet Accounting

## Status

- Status: accepted
- Date: 2026-08-31
- Decision approval: approved directly by owner on 2026-08-31

## Context

MyPocket wallets represent monetary account balances changed by confirmed finance transactions. Gold, stocks, crypto, and foreign currency instead have quantities, acquisition cost, market prices, and valuation changes that do not represent cash income or expense. Reusing wallets would make balance history and financial reports misleading.

## Decision

Create a separate user-owned portfolio domain with asset positions, ordered buy/sell trades, and append-only price history. Trade accounting uses moving weighted-average cost and records realized P&L; current prices produce unrealized P&L. Portfolio valuation is computed independently from wallet accounting. Analytics may read both domains to present wallet net worth, investment market value, and combined net worth as separate values, but portfolio commands cannot mutate wallets or confirmed transactions.

Use hybrid pricing per position: an online leased job refreshes automatic positions through provider adapters, manual positions are skipped, and users can always append a source-tagged manual snapshot. Manual portfolio mutations use the existing offline outbox and explicit conflict model.

Use decimal strings at API boundaries, PostgreSQL `numeric(30,12)` for quantities, and integer VND for prices and totals. Go owns all authoritative rounding and P&L formulas.

## Alternatives Considered

- Model each asset as a savings wallet: rejected because it loses quantity, buy/sell history, cost-basis replay, and price history and conflates cash movement with market movement.
- Store only a manually adjusted total asset value: rejected because it cannot explain cost basis or unrealized P&L.
- Make every market-price movement a finance transaction: rejected because unrealized market movement is neither income nor expense and would corrupt cash-flow reports.
- Build a standalone microservice: rejected because the modular Go monolith and shared PostgreSQL transaction/ownership conventions are sufficient for this scope.

## Consequences

- Positive: wallet accounting stays trustworthy while net worth can include market-valued assets explicitly.
- Positive: append-only prices preserve valuation history and support a future provider adapter.
- Negative: sync, offline storage, conflicts, API, and analytics must support additional user-owned entity types.
- Negative: portfolio totals may be stale or incomplete when current prices are missing; the UI must expose this state.
- Negative: correcting an earlier trade requires deterministic replay of every later trade for that position.
- Neutral: FIFO/tax-lot selection and provider vendor choice remain outside this decision.

## Linked Work

- Ticket: [TICKET-027](../work/tickets/TICKET-027-asset-portfolio-valuation.md)
- Detail design: [PHASE-005 asset portfolio detail design](../work/phases/PHASE-005-asset-portfolio-detail-design.md)
- Implementation plan: [TICKET-027 implementation plan](../superpowers/plans/2026-08-31-ticket-027-asset-portfolio-valuation.md)
- Requirements: [REQUIREMENTS.md](../requirements/REQUIREMENTS.md)
- Roadmap: [ROADMAP.md](../work/ROADMAP.md)
- Architecture: [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)
- API: [API.md](../architecture/API.md)
- ERD: [ERD.md](../architecture/ERD.md)
