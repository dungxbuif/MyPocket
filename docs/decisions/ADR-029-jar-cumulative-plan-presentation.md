---
artifact_type: adr
id: ADR-029
status: accepted_for_implementation
updated: 2026-11-03
---

# ADR-029: Jars presented like wallets with a cumulative plan balance and a wallet reconciliation line

## Context

The owner needs to classify money by purpose while continuing to spend from ordinary wallets, without moving money between wallets. Jars already attach optionally to ordinary expenses, carry a stable identity across months and support a monthly allocation, but JAR-07/JAR-11 forbid carrying leftovers forward and the UI buries the cumulative view. The owner wants jars shown like wallets, with unassigned spending visible, monthly targets kept, and leftovers from earlier months still visible later. Showing a wallet-like number that is not tied to real money would mislead users.

## Decision

- Keep one concept, **Hũ**: an optional dimension on ordinary expenses; expenses stay in wallets; no schema change.
- Each jar card shows **"Còn có thể chi"** = cumulative plan remaining from the jar's first configured month to the viewed month (allocation minus assigned spending; months without an allocation still subtract spending). It is a plan figure and is never labelled "số dư".
- The jar screen always shows a **reconciliation line** comparing total plan remaining with the total balance of wallets that count toward the total (`is_in_total = true`), warning when the plan exceeds real money.
- An **"Chưa phân bổ"** card shows unassigned ordinary expenses (month primary, cumulative secondary) so per-jar totals always reconcile with monthly spending.
- Monthly allocation, soft warnings and %-of-actual-income allocation are unchanged. Income is not assigned to jars in this slice.
- The Assistant may **propose** jar assignments from transaction notes, categories and similar past transactions through a confirmation card (`propose_assign_jar`); it never writes directly and never assigns income, transfers or credit rows.

## Alternatives

- Delete jars and use wallet-description hashtags: rejected; hashtags cannot carry a plan balance or an "unassigned" figure.
- Full envelope model (assign income to jars, jars sum to wallets): deferred; it changes money rules (TX-07, JAR-07) and is only justified after the owner has used the plan view.
- Separate "Quỹ" entity beside jars: rejected; a fourth overlapping concept (category, budget, jar, goal).

## Consequences

- JAR-12/13/14 and AI-JAR-01 are added to the business rules; JAR-07/JAR-11 stay in force (no money carries over; the displayed figure is a plan).
- `GET /jars?month=` grows additive cumulative fields; `GET /transactions` gains a `jar_id` filter; the Assistant gains one proposal tool and one confirmation card type.
- Screens: jar list (wallet-style), new jar detail, transaction filter/chip/editor hint, overview and month lines, report label change.

## Trace

[Detail design](../work/tickets/JARS-AS-WALLET-CARDS-20261103-DETAIL_DESIGN.md) · [Business rules](../requirements/BUSINESS_RULES.md) · [Jar screen spec](../design/pages/jars/README.md) · [API](../architecture/API.md) · [Backlog](../work/BACKLOG.md)
