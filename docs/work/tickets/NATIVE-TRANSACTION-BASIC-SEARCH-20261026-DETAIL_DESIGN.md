---
artifact_type: detail_design
id: NATIVE-TRANSACTION-BASIC-SEARCH-20261026
status: in_review
owner: native-transactions
created: 2026-10-26
updated: 2026-10-26
---

# Native transaction basic search

## Approval and scope

The owner approved implementation of the bounded basic-search proposal from
[REA and Money Lover UI research](../../design/reference-survey/2026-10-26-rea-money-lover-ui-research.md).
This work owns only the ordinary aggregate/basic-wallet transaction ledger,
focused native tests, and the directly required transaction docs. It does not
change goal/credit surfaces, APIs, schemas, backend code, portfolio, Assistant,
production data, feedback resolution, installation, or deployment.

## Intended behavior

- The normal ordinary-ledger header adds a native search action. Opening it
  reveals `MPTextField`, a Close icon action, and a scope line naming the
  selected wallet and week/custom period.
- A blank query shows a prompt to enter a category name or note. Clear returns
  to that prompt without leaving search. Close clears the query and restores
  the normal unfiltered ledger.
- A nonblank query searches only the already-loaded confirmed rows that passed
  the current wallet and period predicates. Category name and note use
  case-insensitive, diacritic-preserving substring matching after trimming the
  query. Pending queue rows are not search inputs.
- Match count, category/day groups, row count and group totals describe the
  matches. The existing income/expense/net summary continues to describe every
  confirmed row in the selected wallet and period.
- No-match state uses `MPEmptyView`. Changing wallet, period, custom dates or
  week offset reruns the same local query. Search does not fetch data.
- Privacy masking, stale-cache messaging/reload, row navigation, bulk-delete
  eligibility and grouping mode keep their existing behavior.

## State and base mapping

| State/action | Native composition | Contract |
| --- | --- | --- |
| Search action | `MPIconButton` in `MPScreenHeader` | Available only on ordinary aggregate/basic ledgers. |
| Open / type | `MPTextField`, `MPText`, close `MPIconButton` | Search mode is local view state; the scope label updates with wallet/period. |
| Clear | Existing `MPTextField` clear action | Query becomes blank and the search prompt returns. |
| Matches | Existing cashflow summary plus match count and `MPTransactionRow` groups | Summary remains period-wide; grouped rows and group totals use matches. |
| No matches | `MPEmptyView` | Explains that no category or note matched in the selected scope. |
| Close | `MPIconButton` | Clears search state and restores the unfiltered list. |

No new shared base, visual token, API/config/schema, architecture decision, or
financial write is required.

## Test-first verification

1. Add focused `FinanceModelTests` for trimmed blank queries, case-insensitive
   category and note matches, missing category metadata, and scope-before-search
   composition. Run them before production code and record the expected missing
   search helper compile failure.
2. Implement one pure search helper and let grouping accept the searched rows;
   keep the existing wallet/period filtered rows as summary input.
3. Run the focused tests, then the complete `MyPocketIOSTests` target and a
   warnings-as-errors simulator build. Run native base/project guards and
   `git diff --check`.
4. Remaining UAT: populated small-phone and large-text simulator interaction
   for action/open/type/clear/no-match/close, followed by owner physical-device
   acceptance. This task does not deploy or install a build.

## Documentation reconciliation

The native transaction screen spec and transaction guide are updated before
code. The guide's Markdown AI-agent table remains the public machine-readable
surface; no separate `llms.txt`, OpenAPI, API/config, migration, operations, or
ADR update is needed because search is local presentation over an unchanged
payload. After implementation, reconcile this note, project context, backlog,
validation matrix and changelog with actual evidence.

## Implementation and evidence

`TransactionsViewModel.search` now performs the bounded category-name/note
match over `filteredTransactions`, while `groupedTransactions(_:)` groups an
explicit displayed row set. `TransactionsView` owns transient search state and
keeps its existing summary on the full wallet+period rows. The header retains
wallet/period access during search; existing queue, stale-cache, masking,
navigation, grouping and bulk-selection paths were not changed.

TDD evidence on iPhone 16 Simulator (iOS 18.5): the first focused build failed
because `TransactionsViewModel.search` was absent; the presentation test then
failed because `TransactionsPresentation.searchState` was absent. The final
focused result passed 6/6 with zero failures, skips or runtime warnings at
`/tmp/mypocket-transaction-search-focused-final-20261026/Logs/Test/`. The
complete signed native unit target passed 392/392 with zero failures, skips or
runtime warnings at `/tmp/mypocket-transaction-search-review-full-20261026/Logs/Test/`.
The final self-review found that search could hide already-selected bulk-delete
rows; a failing availability test preceded the fix, and entering bulk selection
now closes/clears search while selection mode hides the search action.

The first full-target command forced `CODE_SIGNING_ALLOWED=NO` and four existing
Security.framework tests failed with `errSecMissingEntitlement (-34018)`; an
isolated rerun reproduced that command error. No Keychain/outbox code changed.
The correct normally signed full-target command passed. Release simulator
warnings-as-errors build, native base/project guards, `git diff --check`, web
design checks (7/7) and web production build also passed; the web build emitted
only its existing chunk-size advisory.

No API/schema/config/runtime, backend, portfolio, Assistant, production data,
feedback status, installation or deployment changed. Remaining UAT is the
populated small-phone and large-text simulator interaction/capture for
open/type/clear/no-match/close, followed by owner physical-device acceptance.

## Trace

[Backlog](../BACKLOG.md) · [Phase](../phases/PHASE-002-native-ios-swiftui.md) ·
[Plan](../../superpowers/plans/2026-10-26-native-transaction-basic-search.md) ·
[Validation](../VALIDATION_MATRIX.md) ·
[Screen spec](../../design/pages/transactions/README.md) ·
[Guide](../../guides/transactions.md)
