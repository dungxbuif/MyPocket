# Pending recovery in savings wallets

Status: local_policy_fix_pending_ui_gate. Existing approved offline Task6/full-native gap.

Owner review response2026-10-14: explicitly permits test-only selector correction
and rerunning both full scenarios. Preserve attempt history below; no loop reset,
product behavior/assertion weakening/data reset/Feedback placement change.
Helper accepts existing goal selector only when its label matches intended
fixture name, otherwise opens the currently visible aggregate/goal selector,
chooses target and verifies target label. Run after current business sweep ends,
sequentially on same SE; all native units and goal503/422 gates remain required.

Approved rerun26896 stopped at unrelated strengthened-credit-test compilation:
optional fixture balance arithmetic. Fixed only test by XCTUnwrap, no defaults
or product changes. Savings UI had not run. Fresh94298 current gate now running
with both goal cases+credit+allunits; preserve full earlier attempt history.

Terminal94298exit65:273PASS/1FAIL0skip (271units+credit+goal503 pass), no runtime
warnings. `/tmp/mypocket-goal-queue-approved-final-20261015.xcresult`.
Approved selector correction works: intended savings scope survives restart.
Goal503 now proves pending100000/remaining900000, same-key retry, one excluded
pair, API+rendered112500/remaining887500 without manual refresh and unchanged
monthly aggregates. Capture98B2E756 inspected. This is not full goal-wallet UX.
Goal422 now reaches menu but tap opens ordinary TransactionEditor (normal Add)
instead of exposing Sửa bản nháp; target note E7727592 is present underneath.
No receipt/edit assertions after that point reached. Suspect fixed savings CTA
intercepts menu target inside helper's shell-wide bounds; no confirmed root or
fourth product-placement guess. Owner approved only selector correction, not
this new hit-testing change. Stop for bounded design/root review under Rule10.
Capture98B2E756 shows fixed Add across middle of visible savings content and
ledger below it; keep protected-content/full visual acceptance open.

Follow-up diagnosis (2026-10-15): the savings detail's queue editor callback was
guarded by `detailsReady`, which is derived from the category response. A
rejected transfer only needs wallet data to open the transfer editor, so a
transient/empty category response incorrectly removed the `Sửa bản nháp` action
from the queue overflow menu. The bounded fix keeps transaction drafts gated on
category readiness, but permits transfer-draft editing whenever the wallet is
not a credit wallet. This does not alter the fixed CTA, queue state, API or
financial totals. Add a policy unit test and rerun the rejected-transfer
savings UI gate before changing any hit-testing/layout code.
Proxy20834 cwd/command verified then stopped, normal8081 diagnostic suite7487
running separately; no data/app reset or production/device mutation.

Before code: TransactionsView selects a goal wallet by replacing its scroller
with WalletDetailView.goalLedgerContent. Only aggregate/basic branch composes
MPFinancialQueueStatus; WalletDetailView also lacks receipt-revision subscription.
Thus queued intents become inaccessible in savings scope and its private wallet
State may not refresh when the parent receives an acknowledgement.

Bounded implementation: reuse MPFinancialQueueStatus in both goal embedded ledger
and non-credit wallet detail. Keep existing account-wide queue semantics, not a
new wallet filter/permission/source of truth. Reuse existing TransactionEditorView
and TransferEditorView for queued edit via shared presentation modifier; reset
queued/manual targets on save/cancel/AI transition. WalletDetailView reloads on
acknowledged queue revision. No pending amounts added to balance/progress/reports.
No custom styles/new base family/API/schema/auth/config/dependency changes.

TDD real UI/API: create source basic and destination goal, inject503 transfer,
select savings scope and verify pending note/menu there, balance still100000;
restart/reselect savings, retry same intent/key, exactlyone pair, destination
balance112500 in API AND rendered savings view without manual refresh; monthly
report unchanged. Use existing loopback fault helper and preserve orphan test
drafts/production data. Existing goal scope/layout, ordinary/transfer recovery
edit and all native units/guards/project/diff remain regression gates. Then
reconcile savings/wallet/transactions specs, public human+agent guide, context/
backlog/validation/ledger. No Swagger/ERD/newADR if contract unchanged; no release
publication/prod/iPhone install in this slice.

Links: [full goal](NATIVE-FULL-COMPLETION-20261014.md),
[plan](../../superpowers/plans/2026-10-06-offline-financial-sync.md),
[validation](../VALIDATION_MATRIX.md), [guide](../../guides/offline-sync-api.md).

RED94206: actual goal scope loses pending recovery note/card (expected behavioral
failure). First implementation gate11119 failed compilation at onReceive Task
generic inference; no UI/unit execution claimed. Make new reload callbacks
explicit Task<Void,Never>/MainActor, preserving nonthrowing reload behavior.
UI/UX Pro Max verified reusable-modifier guidance supports retaining shared
presentation/base families; no style overrides or new palette adopted.

Second compile88671 still failed type-checking the expanded body without a useful
diagnostic; no tests executed. Split existing editor construction into a computed
view and explicit typed optional callbacks/private Void save function. This is
constraint isolation, not new UI/architecture. If the third compiler gate still
fails, stop this path for review per loop guard rather than guessing further.

Typed-block compile58326exit0, no warnings/errors. Shared-base UI gate will now
exercise both503 retry and422 rejected transfer edit in savings scope: unchanged
confirmation refused, changed amount22500/newkey yields balance122500/remaining
877500, report aggregates unchanged. Retry12500 yields112500/remaining887500.
No changes to pending sums; original100000/remaining900000 asserted before send.

Loop-guard review checkpoint: gate27662 exited65,263 native units passed and both
goal UI cases failed. Bundle: `/tmp/mypocket-goal-queue-green3-20261014.xcresult`.
Both failures occur after restart in selectOfflineGoalScope: it requires
transaction-wallet-selector, but the actual hierarchy already contains
wallet-detail-goal-wallet-selector with the intended destination name. The app
retains its selected savings scope; the test incorrectly assumes the aggregate
scope. This is not evidence of failed receipt refresh or broken pending recovery.
Those assertions have not yet been reached after restart and remain unverified.

Attempt log: baseline94206 missingqueue RED; implementation11119 compilerTask
inference failure; callback typing88671 compilerbody failure; editor factoring
58326 buildPASS then27662 runtimegate failure. Stop here under Rule10 rather than
add another fix/test cycle. Proposed review decision: permit a test-only helper
correction that accepts the already-selected named goal, otherwise uses the
visible goal/aggregate selector, then reruns both real UI cases and normal wallet
regressions. Do not change product selection behavior, weaken financial checks,
clear app data, or modify Feedback placement. No new code change at checkpoint.

Public/master behavior docs are not promoted because the new goal behavior is
not verified. API/schema/security contracts unchanged; no Swagger/ERD update.
Release publication remains withheld pending owner acceptance.

Latest bounded rerun (2026-10-15) could not reach the savings UI assertions:
both selected `LOCAL_API_UI + OFFLINE_FAULT_UI` cases failed during the
fixture login exchange with `401 GOOGLE_MOBILE_LOGIN_INVALID`. Fixture mode was
enabled on the running API, but its disposable PostgreSQL dependency at
`127.0.0.1:55434` was not listening; the old handler collapsed that persistence
failure into the invalid-profile response. The local code now returns a
diagnosable `503 GOOGLE_USER_STORE_UNAVAILABLE` for this class of failure, but
the service was intentionally not restarted. No auth bypass, database,
production state, or environment mutation was made; rerun requires restoring
the disposable local database fixture.

Safe runtime handoff: normal8081 simulator build58532exit0, same bundle installed
without uninstall/data reset, launch PID82860. APIhealth200; owned loopback fault
proxy76027 terminated after checking cwd/command. git diff --check passed.
This build/config restoration is not a new goal-behavior test or acceptance.
