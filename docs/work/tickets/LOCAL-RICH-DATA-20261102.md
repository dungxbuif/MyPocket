---
artifact_type: ticket
id: LOCAL-RICH-DATA-20261102
status: in_review
owner: shared
---

# Rich local data for owner product review

Owner requested additive seed data across MyPocket business domains, from the
start of 2026 through November 2, including products, gold and jars. Preparation
is complete; owner acceptance of the app itself remains open. No API/schema or
production changes are part of this fixture.

Trace: [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md),
[context](../../CONTEXT.md), [release notes](../../releases/CHANGELOG.md),
[native phase](../phases/PHASE-002-native-ios-swiftui.md),
[parallel transaction change](TRANSACTION-DRAFT-SAFETY-20261102-DETAIL_DESIGN.md).

## Target and execution

- Simulator: `MyPocket-iPhone-16`, iOS 18.5,
  `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`.
- API: `http://127.0.0.1:8081/api/v1`.
- Database: `mypocket_native_ui_20261101140940-91507` on loopback PostgreSQL.
- Existing account: `native-fixture@example.test`, owner
  `3da302f3-e9ae-447d-8316-3a7952ca9d18`, Asia/Ho_Chi_Minh.
- Private pre-seed dump: `/tmp/mypocket-before-rich-seed-20261102.dump`.
- Resumable journal: `/tmp/mypocket-rich-seed-20261102.json` (mode 0600).
- Script: `mobile-ios/scripts/seed-rich-local-data.py`. Requires a token file,
  expected owner, local fixture DB and journal; refuses remote API endpoints.

Recovery note: the host reboot at 21:22 on November 2 cleared the temporary dump
and journal paths above. The seeded database survived and was verified with 1,295
transactions. Do not rerun this additive seed against that account with a fresh
journal; reconcile existing sample objects first. Paths above document the
original seed execution, not currently available backup artifacts.

The script completed **1,338 journaled API writes**. API-created sample objects
use actual domain validation and stable idempotency keys where supported.
For pre-existing materialized jar months, PUT only updates an existing config;
there is no public route to add a stable jar to such a month. Bounded local SQL
therefore inserts missing configurations only for the six new, API-created jar
IDs after verifying they belong to the same owner/database. No old row is replaced.

Two initial rejected operations were reconciled before resuming: absent jar
month configuration (404), and historical recurring budget start (400). The
script now backfills only its jar configs, uses historical standalone budgets,
and starts recurring budget series in the current canonical period.

## Available scenarios

| Domain | Readback / scenarios |
| --- | --- |
| Ledger | 1,295 total transactions, including preserved original data. Monthly counts January–November: 126, 110, 124, 122, 116, 114, 124, 137, 144, 157, 21. No seeded actual transaction is dated after November 2. |
| Wallets | 14 total; bank/cash/shopping, internal transfers, balance adjustments, cash account for gold funding, included/excluded wallet totals. |
| Goals | Ongoing purchase goal, exactly achieved laptop goal, overdue underfunded travel goal; contributions, withdrawal and interest. |
| Jars | 10 total; six new jars with monthly history, assigned and unassigned spending, fixed/percent/no-allocation modes, historical Tet jar absent from current month. Original jars retained. |
| Budgets | 27 total; historical monthly budgets, current over-limit shopping, monthly carryover-enabled series and weekly wallet-scoped series. Historical carryover chains are not synthesized. |
| Recurring | Six schedules; active monthly/weekly, paused, and two daily occurrences materialized through the real run-due operation with an ended schedule. |
| Credit | Existing and new cards, purchases, refunds, fees, interest, partial repayments, outstanding debt and fully paid card. |
| Travel | Four total events; completed historical trips, existing future trip, active current Hanoi trip with explicit linked expenses. |
| Gold/assets | 10 total assets and 37 trades. Four configured automatic catalog products; manual chỉ/lượng/phân; gain/loss, partial sale, closed position and missing valuation. Quotes from existing market providers are preserved, including stale status. |
| Gold cash reconciliation | All 36 new sample trades have separate excluded-from-spending cashflows: purchases plus fees 297,815,000 VND; net sales 76,691,600 VND. A 300,000,000 VND opening funding wallet retains 78,876,600 VND. Portfolio APIs themselves do not move wallet money. |
| Categories | Parent/child, wallet-specific, income/debt and archived historical sample categories. Generic borrowing/lending is not fabricated: ordinary transaction validation rejects debt-category transactions. |
| Notifications/feedback | Actual budget worker generated 80%/100% inbox notifications; one marked read, one unread. Three clearly labeled local sample feedback items, including resolved. |
| Reports | All 11 months have ledger data. API monthly October/November readbacks include aggregates/comparisons. Existing user month notes preserved; empty months receive sample notes. |
| Assistant | Actual async request completed and persisted a pending `jar_confirmation` for “Quà cuối năm mẫu”, 2,000,000 VND, November. It has not been approved. Existing conversation retained. |
| Saved analyses | October and November analysis jobs completed. Current November result `01f42926-3f9c-4a73-a856-7c4ca3305d86` restores through latest endpoint; owner identified low content quality, recorded in the follow-up audit. |

Synthetic purchase/manual valuation amounts are test data, not historical
market-price claims. Current catalog supports gold only; no stock/crypto or
term-deposit business is invented. Device-local offline queues, biometric
enrollment and unsent drafts are not backend fixtures. No fake attachment
bindings, APNs tokens, provider results or audit receipts are inserted.

## App handoff and verification

Signed Debug build succeeded with explicit local API 8081; installed without
uninstalling or resetting the container. Device Hub showed authenticated Overview
with seeded recent transactions, November income 48,000,000 VND and expenses
12,367,000 VND. The transaction draft-safety changes are included.

An initial compile-only unsigned install failed Keychain access (`-34018`,
missing application identifier/access group). Rebuilding with project signing
restored the existing authenticated session. Do not install unsigned compile
artifacts for owner login UAT. The signed build has the harmless AppIntents
metadata-extraction warning; no automated suites were executed by owner preference.

API/DB readback and launch are preparation evidence, not full UI acceptance.
Manual review should cover Reports month selection/drill-down, each gold unit,
jar month/history, pending Assistant approval and draft keep/discard behavior.
Docs review: fixture scope, preserved data, limits and runtime signing correction
recorded; no new architectural ADR is required.
