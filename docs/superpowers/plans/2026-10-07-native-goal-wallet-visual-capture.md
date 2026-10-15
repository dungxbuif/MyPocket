# Native goal-wallet visual capture

## Intent

Close the remaining goal-wallet visual/e2e evidence gap in the port audit by
rendering a disposable goal wallet through the real native wallet-detail route
and comparing it with the supplied Stitch savings reference.

## Contract

- Use the local API fixture only; no production data or credentials.
- Preserve the existing wallet/transaction API payloads and owner scope.
- Capture both the explicit wallet-detail route and the embedded goal-ledger
  route after the target, progress, date and transaction sections have loaded.
- Any visual correction must compose existing base components or add a shared
  base with a focused contract; no screen-local color/shape implementation.

## Verification

- Add a deterministic goal-wallet fixture and API-backed UI smokes for both
  entry points.
- Export and inspect the screenshot against the Stitch goal reference.
- Run focused UI, full native unit/contract, Release, project/base/diff guards
  and reconcile the audit/changelog/docs.

## Populated ledger proof — 2026-10-08

The current explicit-detail capture proves goal balance/progress from the
wallet's opening balance but leaves the real transaction list empty. Extend
that API-backed UI fixture with one uniquely noted, owner-scoped expense before
launch. Assert the native detail route renders the persisted transaction row
and amount as well as the progress card, then export and inspect the screenshot.
Cleanup must delete fixture transactions before deleting the fixture wallet;
keep all writes on local `mypocket_stage_v1` through API 8081 and verify zero
matching rows after teardown. No production app behavior or endpoint changes
are planned unless the new live fixture exposes a real product defect.

## Status

- [x] Fixture and UI capture added
- [x] Visual comparison reviewed and actionable mismatches fixed
- [x] Verification/docs reconciled

## Follow-up — embedded savings ledger anatomy

The first API-backed screenshot exposed that the native scoped goal route was
still showing the ordinary `Giao dịch` shell and generic goal card. The native
route now switches to a shared savings-ledger composition: Stitch-style wallet
scope header, search/overflow controls, centered balance/remaining target,
shared progress/advice, month-grouped transactions, report navigation and
add-transaction CTA.
The selector assertion matches the user-facing wallet label because SwiftUI
may flatten descendant accessibility identifiers into the outer card button.
The follow-up anatomy contract places `TẤT CẢ CÁC GIAO DỊCH` between the
centered balance and the target/progress block, matching the supplied Stitch
ledger ordering while keeping balance/progress as shared base components.

Evidence:

- `/tmp/mypocket-goal-focused-final-4/Logs/Test/Test-MyPocketIOS-2026.10.07_17-15-08-+0700.xcresult` — focused API-backed goal detail + scoped ledger, 2/2 passed.
- `/tmp/mypocket-goal-focused-final-4-attachments/10590C2D-C217-43EE-972A-2D4C82B3D657.png` — inspected scoped savings-ledger simulator capture.
- `/tmp/mypocket-goal-focused-final-4-attachments/08298850-047B-44F4-9584-B40DF5E57694.png` — inspected explicit goal-detail simulator capture.
- `/tmp/mypocket-goal-order-green4/Logs/Test/Test-MyPocketIOS-2026.10.07_17-31-18-+0700.xcresult` — focused order regression 1/1 passed.
- `/tmp/mypocket-goal-order-green4-attachments/3F89495A-7555-47CD-B986-00A49E1449E6.png` — inspected reordered ledger capture.

## Follow-up — shell CTA clearance

The first ordering capture exposed a second visual/hit-test defect: the goal
ledger's green `Thêm giao dịch` label was covered by the persistent shell add
affordance on the compact iPhone viewport. A red UI regression reproduced the
overlap (`/tmp/mypocket-goal-cta-red2/Logs/Test/`, CTA maxY 755.33 versus shell
add minY 730). The CTA now lives in a reserved nested safe-area region, uses
the shared `MPButton`, and applies the same canvas treatment plus a 104pt
bottom cushion and 56pt visual lift needed to clear the shell's raised button.

Evidence:

- `/tmp/mypocket-goal-cta-green3/Logs/Test/` — focused API-backed CTA/order
  contract passed 1/1.
- `/tmp/mypocket-goal-cta-green3-attachments/603309EC-18BF-451E-BAB3-FB1D3F7732B7.png` — inspected capture shows the full CTA label and hit region above both shell navigation and the raised add affordance.
- `/tmp/mypocket-goal-cta-full-unit/Logs/Test/` — full native unit/contract
  target passed 110/110 after the final CTA correction.
- `/tmp/mypocket-goal-cta-release` — Release warnings-as-errors build passed;
  project, base-component and diff guards also passed.

## Populated scoped-ledger regression — 2026-10-08

The newly populated API fixture exposed that the old CTA proof only covered an
empty ledger. With one persisted row, the scoped ledger CTA's accessibility
frame ends at y=780.67, below the shell add affordance (y=730) and navigation
(y=767), despite the old inset/offset. A cushion-only experiment was falsified:
increasing it from 104pt to 160pt left the button frame unchanged. The embedded
goal route is a `WalletDetailView` (which owns a `ScrollView`) inside the
transactions screen's outer `ScrollView`; the screenshot places the CTA at the
end of the rendered ledger rather than at the viewport edge. Revised
hypothesis: the nested parent scroller is positioning the embedded content, so
the CTA inset is anchored to the wrong layout container. Render the goal route
outside the transactions outer scroller and rerun the existing API-backed
geometry assertions before changing global shell layout.

The structural fix made the CTA viewport-anchored and both goal route captures
pass. Extend the scoped proof to scroll the real ledger row into the viewport,
verify the whole row is above the fixed CTA (not merely hittable), and capture
that populated state for visual review.
