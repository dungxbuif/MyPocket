# Offline overview/report scope audit

Status: historical audit; bounded implementation completed 2026-10-15. The
original observations and approach alternatives below are retained as evidence;
current behavior is recorded in `NATIVE-OFFLINE-REPORT-CACHE-20261015.md` and
the reconciliation matrix.

## Authoritative observations at audit time (before the 2026-10-15 fix)

- FinancialSnapshotCache persisted only transactions/wallets/categories/timezone,
  owner+server encrypted namespace; no MonthSummary/Insider report document.
- MainTabView injected financialCache and ownerID into TransactionsView, not
  either OverviewView construction. No overview/report offline fallback existed.
- OverviewViewModel.load publishes wallets before awaiting transactions, swallows
  category/profile failures, and continues month/insider loads after core failure.
  loadMonthSummary records an error without clearing a previous summary. This
  permits partially updated/stale sections without explicit cached-at semantics.
- ReportsView holds month input separately from summary but renders summary action
  and trend using the mutable input. load awaits profile, may reset month when
  summary==nil, publishes insider first and swallows monthDetail/transactions
  failures. Catch sets error but retains previous summary; body checks summary
  before error. Risk: failed month change leaves older totals under newer scope,
  or missing transactions appears as an empty trend rather than unavailable data.
- MonthsView similarly retains old summary/note on failed changed-month load.
  saveNote uses mutable month. Potential old-note/new-month writes require a
  runtime regression before declaring an observed data corruption incident.
- FinanceAPI.monthSummary and insider do not take expectedOwnerID, unlike existing
  wallets/transactions/categories/profile. A cached read integration must use the
  ADR022 owner-bound transport before storing results under an owner namespace.

Sources: mobile-ios/MyPocketIOS/Core/Offline/FinancialSnapshotCache.swift;
Features/Overview/OverviewView.swift; Features/Secondary/SecondaryFeatureViews.swift
(ReportsView/MonthsView); Features/Shell/MainTabView.swift; Core/API/FinanceAPI.swift.
These are inspected-code risks, not reproduced UI failures or production findings.

## New manual runtime reproduction — supersedes the Reports-only uncertainty

Device Hub provides simulator display/control on this Xcode installation; there
is no standalone Simulator.app at the legacy Developer/Applications path.
Selected exact SE182D41A2 simulator, launched existing DEBUG local API fixture
(-UITestAuthenticatedAPI,PID92464), not owner's iPhone or production account.
Via native UI: Account→Reports, loaded2026-10, cleared month, entered2026-99,
closed keyboard and tapped View report. Both local GETs returned400 at23:02:55:
`/api/v1/months/2026-99` and `/api/v1/reports/insider?month=2026-99`.
Screen continues to display previous summary/zero totals and trend while invalid
month remains in field; failure is not visibly surfaced. Capture inspected:
`/tmp/mypocket-report-invalid-month-repro-20261014.png`.
This confirms hidden error/stale report retention after invalid scope, NOT
nonzero cross-month amount corruption, offline behavior or Months note-write risk.
No automated regression/new product fix yet; approach approval remains pending.
Returned fixture UI to Overview via tab without saving notes/financial records.
No production/DB/schema/build/install/reset action. New generated DEBUG session
is synthetic; existing durable outbox/cache/app data preserved.

## Approaches presented for review

1. Recommended: typed, query-scoped encrypted snapshots using the existing cache
   protection/owner-bound transport. Preserve server aggregates; atomically publish
   a complete matching read document, distinguish input scope from loaded scope,
   show stale timestamp/retry on transient fallback. Prior ledger cache namespace
   and ciphertext remain compatible; no destructive cache migration.
2. Generic HTTP response cache: less feature wiring, but individual responses have
   different freshness and can form a misleading mixed report. Not recommended
   for financial overview/report coherence.
3. Recompute reports from cached ledger: duplicates server financial policy and
   assumes full rows/catalog/settings. Not recommended without a separately
   proven local finance engine; would not satisfy source-of-truth fidelity here.

## Verification required after approved spec/plan

- Reproduce changed-month failure and partial-load cases with actual native UI and
  independently read API state. No fabricated green test from source searches.
- Save/load/restart/encryption/no plaintext, owner/server/query isolation;
  legacy ledger cache survives, corruption/lost key refuses overwrite.
- Auth/403/malformed/version/owner mismatch never fall back; transient cases do.
- Captured scope/generation rejects late data after month/account changes;
  failed read never enables writes against an unverified month.
- Month/timezone rollover never relabels old data as current; no-cache means
  unavailable, not a zero chart/zero activity count.
- Correct server totals/exclusions and chart data; pending outbox remains separate.
- Foreground recovery/receipt refetch replaces stale state once, no synthetic
  ledger/report mutation or unsupported claim of fresh data.
- Existing base status/button/card/chart families, small-phone UI/AX/keyboard
  reachability, full native units and critical API UI regressions.

## Workflow boundary

Brainstorming classified this as architectural read-coherence extension: existing
cache is reused, but multi-response snapshots and publication ownership need a
written spec + reviewed plan after approach approval. Do not implement a smaller
wallet-only cache and call overview/reports complete. No new production schema,
API owner override, new crypto/dependencies, signing or rollout proposed.
Public behavior docs will change with implementation, not present this draft as
shipped. [Reconciled implementation](../bugs/NATIVE-OFFLINE-REPORT-CACHE-20261015.md)
now covers the typed snapshot path; no release publication. Goalqueue/Feedback
loop-review blocks unchanged.

Links: [full goal](NATIVE-FULL-COMPLETION-20261014.md),
[cache foundation](OFFLINE-FINANCIAL-CACHE-20261014.md),
[ADR022](../../decisions/ADR-022-owner-bound-native-refresh.md),
[offline contract](../../superpowers/specs/2026-10-06-offline-financial-sync-contract.md),
[validation](../VALIDATION_MATRIX.md).
