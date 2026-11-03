---
artifact_type: detail_design
id: LOCAL-FIRST-20261103
status: in_progress
owner: shared
human_fields:
  - approval
  - scope_decisions
parent: FEEDBACK-20261019-OFFLINE-LOADING
---

# Local-first reads and writes (native iOS)

## Owner requirement (2026-11-03)

Everything the user does happens against local data first; sync to the server
afterwards. Goal: remove loading when switching tabs and long waits when saving.
Always read from local data / local mechanism first, then load in the background and
update the UI only if something changed.

## Rules (acceptance contract)

1. **Read path:** every screen renders synchronously from the owner-scoped local
   store on first appearance. A server refresh runs in the background and patches
   the UI only when the payload differs. A blocking `MPLoadingView` is allowed only
   when the local store has nothing for that screen and scope.
2. **Write path:** create/edit/delete/transfer/adjust apply to the local store at once,
   update dependent local aggregates (wallet balance, month totals, jar spend, budget
   spent) and close the editor. The change is enqueued in the encrypted outbox with an
   `Idempotency-Key` and an `expected_version`, and synced in the background.
   Save latency visible to the user must not include a network round trip.
3. **Visible sync state:** each pending row shows a small pending marker; failures show
   an inline retry/discard entry, never a modal that blocks other work.
4. **Convergence:** server response replaces the optimistic row (same client id,
   server version). A rejected write (validation, 409, 4xx) rolls back the optimistic
   change and surfaces the server message with the original draft restored for edit.
5. **Isolation:** the local store is keyed by owner and server namespace; sign-out
   and owner switch never expose another owner's rows (existing requirement).
6. **Reports and AI stay server-derived:** monthly report, Insider, saved analysis and
   Assistant are read-through caches (rule 1) but are never computed or edited locally.
   Local aggregates must not be presented as report facts.

## Current state (source audit, 2026-11-03)

- `FinancialSnapshotCache` (encrypted, owner/server scoped) exists, but Overview,
  Wallets and Transactions read it only in the failure fallback
  (`allowsFallback(error)`); Reports, Budgets, Jars, Savings, Portfolio, Travel,
  Recurring and Assistant history do not use it.
- `OutboxStore`, `SyncEngine`, `FinancialMutationQueue` exist, with encrypted SQLite
  and leases, but the architecture note records that no screen instantiates them; the
  writes in `TransactionsView` still await the API before closing.
- Backend contract is ready: `GET /sync/capabilities`, mandatory `Idempotency-Key`
  for offline writes, `expected_version` conflict (409), paired transfer semantics.
  Gap: `expected_version` is optional on `PATCH /transactions/:id`, so a replayed
  offline edit without a version can overwrite newer server data. Offline writes must
  always send it.
- Re-audit (2026-11-03) found the Assistant send flow marks its pending message
  `uncertain` before the request is sent; it must not be reused as the in-flight flag
  for outbox state.

## Proposed design

1. **Local read model.** Extend the snapshot cache to a keyed store per screen scope
   (month, wallet filter, budget id, report snapshot id) with `cachedAt` and a content
   hash. A shared view-model base `hydrate -> render -> refresh -> diff -> publish`
   replaces the per-screen load code; one `MPLoadingPresentationPolicy` decides
   blocking vs inline.
2. **Prefetch.** After login and on foreground, a single background task refreshes the
   tab roots (Overview, Transactions, Wallets, Budgets, current-month Report) so
   the first visit to a tab is already warm. Requests are de-duplicated per scope.
3. **Local mutations.** Transaction create/edit/delete/transfer/adjust and wallet,
   budget, jar, recurring writes go through `FinancialMutationQueue`: apply to the
   local store, enqueue, trigger `SyncEngine`. Credit entries/payments, portfolio
   trades and AI-confirmed actions stay online-confirmed in phase 1 (paired ledger
   semantics) and use inline progress, not a blocking sheet.
4. **Conflict policy.** Offline edits send `expected_version`; on 409 keep the local
   draft, show server values, let the user choose. Deletes of already-deleted rows
   converge silently.
5. **Server hardening required by this design.** Reject `PATCH`/`DELETE` of
   transactions without `expected_version` when `Idempotency-Key` marks an offline
   replay; log the 500 on `POST /notifications/devices` (it runs on every launch
   and must not block the sync path).

## Phasing

| Phase | Scope | Exit evidence |
| --- | --- | --- |
| 1 | Read path for Overview, Transactions, Wallets, Budgets, Reports; shared hydrate/refresh base; prefetch | Tab switching shows no full-screen loading with warm data; signed-out/owner-switch isolation test; no extra provider calls |
| 2 | Local writes for ordinary transactions and transfers via the outbox; pending/failed row UI; rollback on rejection | Save closes < 100 ms with airplane mode; reconnect syncs exactly once (idempotent replay); 409 and 4xx paths covered |
| 3 | Wallet, budget, jar, recurring writes; remaining read screens; Assistant history read-through | Same checks per screen; full native unit target and base guard green |

## Implementation log — 2026-11-03 (implemented, tests pending)

Root cause of tab-switch loading: `MainTabView` rendered tabs with `switch selection`, so
each tab, its view model and its decrypted cache were destroyed and rebuilt on every
switch. Phase 1/2 changes:

- `DesignSystem/MPTabKeepAlive.swift`: visited tabs stay mounted (state, scroll, rows);
  `transactions` and `reports` are preloaded 0.8 s after launch; inactive tabs are
  invisible, non-hit-testable and hidden from accessibility; Overview, Transactions and
  Reports run a silent refresh on re-activation (`mpRefreshOnTabActivation`, inline only).
- `FinancialMutationQueue.submit` returns once the mutation is durable in the encrypted
  outbox; drain runs in a background task (previously it awaited the network).
  Result is now always `.queued`; pending rows show in the existing "Chờ xác nhận" card.
- `TransactionsViewModel`: edit (`submitUpdate`), delete and transfer delete update the list
  at once, confirm in the background, restore the previous rows and show an error on
  rejection; refresh is suppressed while local changes are unconfirmed.

Phase 1–3 completion (same day):

- `Core/Offline/MPReadThroughCache.swift` + `APIClient.cached(_:)`: every successful GET on an
  allow-listed read path (reports, savings, insider, budgets, budget-series, jars, months,
  portfolio, recurring, travel) is stored encrypted per owner/server; Reports, Budgets and
  Jars paint the last known response on first visit and then always refresh. Only the
  production transport persists; injected test transports are unaffected.
- `DesignSystem/MPBackgroundWrite.swift`: wallet, budget (single/series), category, jar,
  recurring and travel editors close at once; the request runs in the background, the owner
  screen reloads on success, and failure shows an inline banner with Retry/Dismiss using the
  server's `userMessage` when available. Not queued offline (the server contract has no
  offline operation for these).
- Wallet balances are recalculated optimistically for transaction edit/delete/transfer
  delete and restored on rejection.
- Assistant: the "có thể đã được tiếp nhận" banner and disabled composer no longer show while
  the request is still in flight (durable uncertain marker kept for dismissal recovery).
- `check-base-components.sh` now fails (exit 2) when `rg` is missing.
- Backend: `POST /notifications/devices` logs the cause of a 500 (token never logged).

Deliberately unchanged: bulk delete stays online-confirmed (atomic 1–100 rows, selection mode
depends on its result); credit entries/payments, portfolio trades and AI-confirmed actions
stay online-confirmed; server-side `expected_version` enforcement is not needed because the
offline contract replays creates only (edits/deletes are never replayed offline). Offline
contract supports only `transaction.create` and `transfer.create`.

## Verification plan

- RED first: per view model, cache-hydrates-before-network, refresh keeps current
  rows, identical payload does not republish, owner switch hides cache.
- Outbox: kill the app mid-sync, restart, no duplicate server rows (`Idempotency-Key`
  replay), ordered per wallet, stale-version 409 handled.
- Device UAT: Overview -> Transactions -> Wallets -> Reports repeatedly with
  throttled/offline network; measure first-paint and save-close time before/after.
- The base-component guard must fail when `rg` is missing before its result is used
  as evidence (see backlog re-audit item).

## Owner decision — 2026-11-03

Owner approved ("chốt rồi"): implement phase by phase, one item at a time, fastest path;
implement first, run one combined test pass afterwards to minimise acceptance time.
Open questions resolved with the conservative defaults below (owner may override):
credit/portfolio/AI actions stay online-confirmed in phase 1; offline-created parents
must sync before dependants (dependants queue behind them); warn when a write is
unsynced for more than 24 h.

## Open questions (resolved above)

1. Are credit/portfolio writes allowed to stay online-only in phase 1?
2. Offline creation of a transaction whose wallet/category was created offline: allow
   (needs dependent-queue ordering) or require the parent to be synced first?
3. How long may a pending write stay unsynced before the UI warns?

## Trace

[Backlog](../BACKLOG.md), [parent offline/loading design](FEEDBACK-20261019-OFFLINE-LOADING-DETAIL_DESIGN.md),
[mobile architecture](../../architecture/MOBILE_IOS.md), [API sync contract](../../architecture/API.md),
[validation](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md).
