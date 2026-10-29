# Current native critical-business regression

Status: in_review (isolated category280PASS; broader291-test regression291PASS/0FAIL/0skip; test-only helper/fixture corrections verified). Existing approved UI Task6/full-native verification, not a
new feature/design. Previous goal turn yielded fresh backend AI audit evidence;
confirmation-summary design is pending and is not implemented by this gate.

## Approved reveal helper correction — 2026-10-15

Broad gate11171 exit0:291PASS/0FAIL/0skip/no runtime warnings,
`/tmp/mypocket-reconcile-full-green-20261015.xcresult`. The prior category,
credit and recurring failures were corrected only in the XCTest helper/fixture:
keyboard dismissal before category icon reveal, unique smallest visible nested
ScrollView selection, and explicit recurring wallet selection/readback. Full
exact-data assertions remain. This is local simulator/API proof only; it does
not close physical-device, production-backend compatibility, or release/UAT.

Owner approved test-only semantic scrolling. Draft-only RED54467 exit65,
`/tmp/mypocket-name-scroll-probe-20261015.xcresult`: literal name matches after
typing, fails immediately after reveal and before balance focus. Activity trace
shows twenty swipes on ScrollView (0,407,375,44), the keyboard suggestion bar;
the actual editor ScrollView is (0,0,375,139). No Save/API write in this probe.
This establishes a helper-induced mutation, not product autocorrection failure.
The separate approved literal-name requirement does not authorize an unsupported
base fix; notes/chat, keyboard settings, palette/font/API remain unchanged.

Design: existing targets identify their containing ScrollView through stable ID
or label. Use that content viewport, intersected with the app window, rather than
global firstMatch (which includes QuickType). Missing lazy rows may scroll only
the identified shell content container. Missing/ambiguous containers fail with
geometry instead of blind swipes. Preserve all exact name/data assertions.
Rejected: disable autocorrection to hide test gestures; arbitrary lastMatch;
relax persisted-name assertions. Scope: LaunchTests helper only, no product,
schema, auth, config, runtime or ADR change. Risk: nested/unnamed containers
must fail explicitly rather than silently select a different surface.

Verification: rerun the no-save focus probe, wallet/category CRUD and wallet
transaction CRUD with all native units, LOCAL_API_UI on SE against isolated8081.
Then run broader critical-business regression; retain old RED history and stop
for review if the bounded correction exposes a different root. Reconcile this
ticket, context/backlog, validation and plan ledger. Public human/agent API docs
need no change: test infrastructure only, no shipped contract change.

Approved pass1/probe46637 exit65: all four literal-name boundaries now PASS;
helper alone fails its new bounded-visibility check. Geometry target y−130.5/h70,
content h139; full swipes overshoot in the compressed editor. No Save, no product
patch. Pass2 stays within approved helper scope: drag40% of the identified content
height rather than a full-page swipe. Keep visibility/data assertions unchanged.

Pass2/probe34635 exit0:1/1PASS0skip, no runtime warnings,
`/tmp/mypocket-reveal-bounded-probe-20261015.xcresult`. Four literal boundaries
and reveal visibility pass; captures36071F8C/9B59DDC3 inspected. Form remains
keyboard-compressed and feedback overlays upper content after dismissal; this is
not full wallet visual acceptance. All-unit+three CRUD regression80693 now runs
with unchanged API/persisted-data assertions. Base guards/guard tests/project
verification/diff pass. Horizontal search finds other firstMatch scrolling in
More, savings, chat/photo tests; not blindly replaced without scoped evidence.
No release entry/publication: owner acceptance still required and this fix is
test-only. No production/device/app reset or broad fixture cleanup performed.

## Regression result / next bounded review

Final reviewed helper pass84393exit0:280/280PASS0skip/no runtime warnings
(279units+category CRUD/archive/restore/delete UI). Exact singleton wallet scope
and literal name API readback, row reveal and kind-filter transitions nowPASS.
Bounded20drags/21visibility checks closes final-drag false timeout; no product
category/keyboard/API change. Evidence `/tmp/mypocket-reconciled-category3-20261015.xcresult`.
Next broaden unchanged12business gate and strengthen final category-delete proof:
GET category collection must succeed and exclude the deleted ID, not merely show
an empty filtered list. Assert wallet scope survives archive/restore too. This is
additional real-state proof, not a relaxed assertion or new product fix.

Approved refinement pass1,53977exit65:281PASS/1categoryFAIL0skip, all278units
and wallet probe/create/update/transaction CRUDPASS. Category has5assertion
failures, not just the first subtitle. Trace confirms kind transitions complete;
search focus leaves307pt content, category-tree begins383 and lazy target is
absent. Waiting for existence without scrolling cannot materialize this row.
Scope subtitle was read without first revealing created row. Next pass remains
test-only: reveal created row before reading value, dismiss search keyboard and
reveal exact-ID lazy row before existence/edit. Fail fast with actual label/value
if scope still fails; preserve literal names/API scope/archive/data assertions.
No product/list/keyboard setting change. Attempt2 within renewed approval.

Pass2,27187exit65 fails fast on scope with actual AX captured: created category
has “Hoạt động trong3ví”, because test deselected onlyB from all existing wallets.
Prior API assertions only checked includesA/excludesB; they did not prove singleton.
Also current AX places value on wrapper Other, while identified Button combines
title+subtitle in label. Prior source-only value hypothesis was incorrect for
the queried element. Pass3 (last under loop guard): deselect every other wallet
through existing checkbox UI, leaveA selected, assert exact API wallet-ID set[A],
then original combined-button label singleton/archived semantics. Keep new lazy
reveal and transition waits. No product fix or relaxed scope expectation.

Pass3,77512exit65:1FAIL0skip at3579, waiting for newly saved name before fetching
its API row. Editor closed, actual AX shows list at top with long lazy tree;
created name not materialized. Earlier singleton checkbox and literal name
assertions pass; cannot claim persistence/singleton/archive from this stopped
run because API readback happens afterward. No fourth helper patch. Review sent:
after Save wait for editor dismissal, read exact-name API fixture, reveal its
exact-ID row, then assert UI title/scope and continue unchanged CRUD/archive/
restore/delete assertions. Hypothesis is lazy existence ordering, not proven
product create failure or autocorrection. Exported AX75265AAF inspected; running
screenshot blank was captured near terminal transition, not proof of app crash.

Owner now explicitly requests reconciliation/completion of the listed remaining
items. Proceed with the previously presented bounded test-only pass: wait for
editor dismissal, exact-name API readback, exact singleton scope, reveal row by
persisted ID, then unchanged title/scope/kind/edit/archive/restore/delete checks.
This is the reviewed next approach after the prior loop guard, not a hidden
fourth guess or an approval to alter product keyboard/category behavior.

Current pass25223exit65:279unitsPASS/1UIFAIL. Actual failure is reveal of
category-icon-selector, not category persistence: last geometry targety233.99/h44
inside editor viewport0..307 while helper reports timeout. Loop performs20drags
but never rechecks the frame after its final drag. The recorded last drag made
the control visible; test then continued and reached subsequent steps. Correct
the bounded helper to perform21visibility checks with at most20drags, preserving
same containers/bounds and every assertion. No extra blind swipe, product or
keyboard change. First13913 was compile-only missing async XCTest wait; corrected
to await fulfillment before25223, not a behavioral RED or product issue.

80693 terminal65:273PASS/2FAIL0skip, no runtime warnings,
`/tmp/mypocket-reveal-crud-regression-20261015.xcresult`. All271unit tests,
literal-name probe and wallet create/update PASS. Category now creates its exact
name and API wallet scope correctly, but UI test reads subtitle from label while
MPCategoryTreeCard deliberately exposes title as accessibilityLabel and subtitle
as accessibilityValue (MPFinanceCards.swift705–706). Subsequent immediate kind
filter and edit assertions fail; do not assume all are timing-only without proof.
Wallet transaction create/update data assertions run, but new fail-closed helper
reports two pinned-save controls without scroll ancestors: frame(16,601,275,44)
and(16,607,343,44), shell viewport ends611, app window ends667. This is not proof
the actual saves failed. Do not weaken financial assertions or call suite green.

Two correction passes completed under approved helper design; regression exposed
different pinned-footer/AX cases. Stop further fix attempts per loop guard and
request bounded review: visible pinned control must fit actual available window,
be hittable and not keyboard-covered, never scroll unrelated content; category
scope reads accessibilityValue and filter assertions wait for actual state
transition, failing if absent. Approval pending. Keep prior attempt history.
Captures94298DDD/geometry06291559/2CE8F6B6 inspected. No live tests at checkpoint.

Exact regression command (from repository root):
```sh
rtk proxy xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=182D41A2-4FC6-445B-A592-70FDCB5C662E' -derivedDataPath /tmp/mypocket-reveal-approved-build-20261015 -resultBundlePath /tmp/mypocket-reveal-crud-regression-20261015.xcresult -only-testing:MyPocketIOSTests -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedWalletNameDraftStaysLiteralAcrossFocusWithoutSaving -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedWalletCreateAndUpdateFromNativeEditor -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedCategoryCreateUpdateArchiveRestoreDeleteFromNativeUI -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedWalletDetailCreatesAndUpdatesTransaction -parallel-testing-enabled NO -collect-test-diagnostics never 'SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI' MP_API_BASE_URL=http://127.0.0.1:8081/api/v1 SWIFT_TREAT_WARNINGS_AS_ERRORS=YES -quiet
```
Probe uses the same build flags and only the literal-name test, with unique
`/tmp/mypocket-reveal-bounded-probe-20261015.xcresult` (34635exit0,1/1PASS).
Docs review: no requirements/architecture/API/schema/security/ADR/public human
or agent surface changes needed for this test-only correction. Internal ticket,
context/backlog/matrix/ledger reconciled; no release publication or resolution.

2026-10-15 next root probe (no product/base patch): manual DeviceHub draft-only
input inconclusive because native control drops fast keystrokes, paste delivered
no text, per-character batch also dropped input. Cancelled form; no Save used,
list returned unchanged. Do not call this autocorrection reproduction.
Use one deterministic XCTest draft-only probe: assert exact wallet name directly
after typing, after focusing balance, after Xong; capture each boundary and cancel
without POST. This isolates input/focus from backend persistence and avoids new
malformed fixtures. Literal-name design still pending; no global keyboard config,
UIKit correction toggle or product change before evidence/approval. Prior goal
turn was progress (selector/test correction and new actual financial evidence).

Latest diagnostic7487exit65:272PASS/2FAIL0skip (271units+correct-wallet transaction
CRUD PASS), no runtime warnings. `/tmp/mypocket-business-diagnostic-20261015.xcresult`.
Category root semantic query works. Both exact-name assertions immediately after
typing PASS; create/list exact-name checks still FAIL after focus changes/save.
New AX BD6CAF77/BD47573A shows appended prose persisted, so early typed value is
not enough. Wallet/category input-completion hypothesis remains unconfirmed.
Base MPTextField uses MPUikitTextField for both named/unnamed inputs; it currently
does not configure autocorrection. No base/product patch applied. Proposed
literal-name-only contract (notes/chat unchanged) sent for owner design decision;
test any keyboard hypothesis before adoption, do not disable global settings.
Known owned malformed fixture IDs from evidence require exact validation before
cleanup; strict-name teardown cannot remove names it did not expect. No broad
prefix purge or financial data reset authorized/performed.
Normal8081 simulator binary default-launched25086; no18081listener, APIhealth
previous200; guards/guardtests/project/diff pass. No tests running now, no deploy/
physical install/release. Full business suite remains RED, not accepted.

Sweep86110 terminal65:280PASS/3FAIL0skip (271units+9UI pass), no runtime warnings.
`/tmp/mypocket-current-business-sweep-20261014.xcresult`.
Failures: categoryCRUD (root identity asserted as Other but actual ScrollView;
created name has unexpected appended prose), walletCRUD (same appended-name
phenomenon), wallet transactionCRUD (firstMatch Chi tiết selects the wrong wallet,
so expected fixture has no matching transaction). These are not dismissed as
test-only without followup; input mutation root cause remains unknown.
Actual AX F4E694F0/386661FD show stored category/wallet names with long appended
sentences absent from test literals. Wallet detail video4456BE40 shows test opened
that wrong wallet; capture570D77A0 inspected. Own wallet-create videoF1A8455E
frames13/18 show name changing during keyboard entry; no provider call is involved.
Do not silently change app auto-correction, cleanup other user data, hide input
failures, or call full business suite green. Strengthened credit test and approved
savings helper will run separately. Initial combined gate26896 failed compile
only because FixtureWallet.currentBalance is optional; replaced arithmetic with
XCTUnwrap, not zero fallback. No test execution from that build failure claimed.

Current local API8081 health200; branch codex/native-ios-port, preserve dirty
work. Run all native units and existing actual API/native UI scenarios: wallet
create/update, transaction create/update, atomic transfer create/edit/delete,
category CRUD/archive/restore, budget, jar, recurring, travel, portfolio trade,
credit entry/payment, persisted report trend, all More routes. Tests own synthetic
fixtures and API readbacks; no production, owner credentials, app/database reset,
new permissions, provider calls or schema changes. Sequential SE simulator,
DEBUG LOCAL_API_UI, base8081, warnings-as-errors, fresh result bundle.

Inspect individual failures and captures, not only aggregate pass. A passing
critical subset does not prove every screen/function/offline/action/RAG/APNs or
physical FaceID/install goal complete. If failure occurs, reproduce/root-cause
and TDD under existing contract; do not guess around blocked Feedback placement
or savings selector. Reconcile validation/context/backlog and human/agent docs
if actual stale behavior is exposed. Existing loop guards remain per work item.

Test-source audit found credit scenario creates source fixture but never selects
it; product defaults first non-credit wallet. Strengthen test only: unique fixture
names, explicitly choose source through existing form-row selector, assert same
pair IDs/two wallet directions/exclusion, source and credit balance deltas and
unchanged report after payment. Teardown removes only pair IDs linked to owned
test wallets before wallet cleanup. Current sweep already compiled old test;
rerun updated credit scenario separately after sweep, do not conflate results.
This is fixture isolation/proof correction, not product behavior or API change.

Strengthened credit gate94298: actual native source selection and two-sided
pair/balance/report assertions PASS (combined suite273PASS/1goal422FAIL0skip).
Credit source100000→92500, credit balance88000→95500; exactly2excluded pair rows,
same transfer/credit payment identity, monthly aggregates unchanged after payment.
Capture76F1A2EF inspected; schema/API/product unchanged. Visual caution: positive
credit balance rendered as Dư nợ and statement-config unavailable despite fixture
inputs warrants separate credit-label/config readback audit; not accepted here.

Three earlier failures diagnostic correction: root category screen queried by
semantic identifier across element types (observed ScrollView), transaction opens
wallet-detail-<fixtureID> instead of first Chi tiết, and exact name assertions
stop before save if keyboard entry changed the literal. No product keyboard
configuration or relaxed persisted-data assertions. Fresh normal8081 gate7487
running (271units+3UI); no green claim until terminal. Proxy stopped beforehand.

Links: [goal](NATIVE-FULL-COMPLETION-20261014.md),
[UI plan](../../superpowers/plans/2026-10-14-native-ui-system-consolidation.md),
[validation](../VALIDATION_MATRIX.md),
[registry](../../design/platforms/ios-screen-registry.md).
