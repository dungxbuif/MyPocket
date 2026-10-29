# Native Transaction Basic Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add bounded local category/note search to the ordinary native transaction ledger without filtering its period summary or changing offline/API behavior.

**Architecture:** Keep `TransactionsViewModel.filteredTransactions` as the confirmed wallet+period source of truth. Add a pure category/note matcher and a grouping function that accepts the matched rows, while `TransactionsView` owns transient open/query state and composes existing design-system bases.

**Tech Stack:** Swift 6, SwiftUI, XCTest, Xcode iOS Simulator.

**Spec:** `docs/work/tickets/NATIVE-TRANSACTION-BASIC-SEARCH-20261026-DETAIL_DESIGN.md`

## Global Constraints

- Ordinary aggregate/basic ledger only; goal and credit surfaces are unchanged.
- Search already-loaded confirmed rows only; no API/schema/network search.
- Existing period summary remains unfiltered; match count and group totals may describe matches.
- Preserve offline queue, privacy masking, stale-cache state, navigation, grouping and bulk-selection eligibility.
- Do not deploy, install, resolve feedback, edit production data, or touch unrelated backend/portfolio/AI files.

## Review Focus

- A blank or whitespace-only query must show the open-search prompt, not silently become a match-all result list.
- Category matching must be case-insensitive and must not accidentally search category IDs.
- Note matching must handle `nil` and empty notes without errors or false matches.
- Wallet/period filtering must run before search so out-of-scope rows never appear.
- Missing cached category metadata must still allow note matches and must not crash stale-data presentation.

---

### Task 1: Pure search behavior

**Files:**
- Modify: `mobile-ios/MyPocketIOSTests/FinanceModelTests.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionsView.swift`

**Interfaces:**
- Consumes: wallet+period rows from `TransactionsViewModel.applyFilters` and `[CategoryModel]`.
- Produces: `TransactionsViewModel.search(_:query:categories:) -> [TransactionModel]` and grouping over an explicit row list.

- [x] Add focused tests for case-insensitive category/note matches, trimmed blank query, absent category metadata and wallet/period composition.
- [x] Run the focused tests and confirm RED because the search helper is missing.
- [x] Implement the pure helper and explicit-row grouping with no state or network side effects.
- [x] Run the focused tests and confirm GREEN.

### Task 2: Native search states

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Transactions/TransactionsView.swift`

**Interfaces:**
- Consumes: the Task 1 search helper, existing `MPScreenHeader`, `MPIconButton`, `MPTextField`, `MPText`, `MPEmptyView`, `MPTransactionRow` and summary/group presentation.
- Produces: search action, scope line, prompt, count, no-match state, clear and close behavior for ordinary ledgers.

- [x] Add local open/query state and derived searched rows without changing `filteredTransactions` totals.
- [x] Compose the focused header and result states from existing bases; close must clear the query.
- [x] Confirm wallet/period changes automatically recompute results and preserve grouping/navigation/masking.

### Task 3: Reconcile and verify

**Files:**
- Modify: `docs/work/tickets/NATIVE-TRANSACTION-BASIC-SEARCH-20261026-DETAIL_DESIGN.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/releases/CHANGELOG.md`

**Interfaces:**
- Consumes: final implementation and fresh command evidence.
- Produces: accurate implementation status, test evidence and explicit remaining UAT gate.

- [x] Run focused tests, full `MyPocketIOSTests`, warnings-as-errors simulator build, native base/project guards and `git diff --check`.
- [x] Reconcile docs with exact outcomes and keep populated small-phone/large-text plus physical owner UAT open unless actually performed.
- [x] Inspect the final scoped diff and confirm no unrelated file was changed.
