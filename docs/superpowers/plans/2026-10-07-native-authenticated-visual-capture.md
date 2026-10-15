# Native authenticated visual capture

## Intent

Turn the existing authenticated UI smoke tests into inspectable visual
evidence for the remaining port-parity gates, starting with More/account and
the add-transaction editor.

## Contract

- Screenshots are test attachments only; no production runtime behavior or API
  contract changes.
- Launches use the existing deterministic simulator fixture and never mutate
  production data.
- Capture points occur after the existing route/element assertions so an image
  cannot be mistaken for a route-only proof.

## Verification

- Add stable screenshot attachments to existing UI tests.
- Run the focused tests and export/inspect attachments from xcresult.
- Update visual audit docs with exact evidence and any follow-up implementation
  plan; rerun native/web guards after test-only changes.

## Status

- [x] More/account and transaction screenshots captured
- [x] Visual mismatches reviewed and addressed where actionable
- [x] Docs and verification reconciled
- [x] Capture each authenticated secondary destination during the existing route
  sweep so screen-level visual review covers more than route reachability.
  In `LOCAL_API_UI` builds the launch must use the isolated stage API so account,
  feedback, wallet, and finance screens do not capture artificial network-error
  states. These simulator images do not replace populated multi-record fixtures,
  Stitch/web comparison, or owner/device acceptance.

## API-backed secondary-route sweep — 2026-10-08

- The route test now uses `-UITestAuthenticatedAPI` under `LOCAL_API_UI` and
  attaches one screenshot after assertions for each destination.
- iPhone 16 Simulator test passes 1/1 (15 destinations) at
  `/tmp/mypocket-secondary-visual-sweep-api/Logs/Test/Test-MyPocketIOS-2026.10.08_04-28-14-+0700.xcresult`.
- All 15 screens loaded from isolated local API stage (`mypocket_stage_v1`,
  API `127.0.0.1:8081`); no network-error placeholder appeared. Evidence is
  empty-account visual capture, not populated-data or pixel-parity acceptance.
- Inspection found the global feedback bubble obscured the trailing part of
  the Face ID explanation on the Account screen; the isolated API-backed UI
  test reproduced the collision before the fix.
- Follow-up contract: keep the floating feedback action available on Account,
  but its tappable frame must not intersect the Face ID explanation; prove this
  with an API-backed simulator geometry assertion and inspect a fresh screenshot.
- Implementation choice: reserve the floating bubble's width on that explanatory
  text only, allowing it to wrap without moving the global feedback action or
  changing the shared shell/navigation geometry.
- RED/GREEN: the `LOCAL_API_UI` iPhone 16 test failed on the original one-line
  explanation and passes after reserving the bubble width. Screenshot:
  `/tmp/mypocket-feedback-collision-api-green-attachments/D93BF6B9-DB9F-4F60-829C-24BE51B9497D.png`.
- Remaining gates: populated multi-record fixtures, side-by-side Stitch/web
  comparison and owner/device acceptance.

## Wallet-selector header pixel parity — 2026-10-08

- Visual comparison of Stitch `wallet-selector/screen.png` with the API-backed
  native management screenshot showed the title left-aligned and both actions
  uncontained, unlike Stitch's centered `Chọn Ví` and capsule actions.
- Preserve the shell route and semantics, but use a reusable centered selection
  header: visible `Đóng` returns to the previous screen; `Sửa`/`Xong` toggles
  edit mode. Preserve the stable `screen-back` and `wallet-edit-mode` IDs.
- RED/GREEN proof: assert the real title geometry is centered and each action
  occupies a capsule-sized hit target, then inspect an API-backed screenshot.
- No API, persistence, wallet-scope or business-rule changes; update the wallet
  screen/reference contracts and validation/changelog after verification.

## Wallet-selector row pixel parity — 2026-10-08

1. Add an API-backed UI regression for the reference row anatomy: stacked
   title/balance, included-wallet total and selected check.
2. Refactor the shared `MPWalletSelectionList` row base to match web and Stitch
   hierarchy without removing native detail/edit/goal/credit actions.
3. Capture the populated management list; verify selection state and narrow
   device readability; place the native-only Chi tiết extension in the same
   compact row to avoid adding a second-height action row; run native Release
   and base-component guards.

- RED: the populated API-backed contract initially failed because native did
  not expose title/balance geometry IDs or a selected aggregate mark.
- GREEN: `testAPIBackedWalletManagementMatchesGroupedWebContract` passes 1/1
  and asserts both rows stack their labels and balances with aligned leading
  edges plus the selected aggregate check. Screenshot:
  `/tmp/mypocket-wallet-rows-final-attachments/4F92374E-3F0E-49FF-A097-675913C3EB34.png`.
- The native selection row is now a shared DesignSystem base; native-only
  Chi tiết is a compact trailing action on the same row; goal progress and
  credit availability remain present below it. A second RED/GREEN geometry
  assertion verifies the Chi tiết action shares the row centerline. Final
  screenshot: `/tmp/mypocket-wallet-detail-row-attachments/1DD156EC-692D-4DC9-8CCE-B541B2E74EEF.png`.
  No API, database or wallet calculation contract changed.

## Wallet aggregate glyph parity — 2026-10-08

The reference/web contract calls for a globe badge in `categoryTeal` on the
aggregate row. Native reused the basic-wallet orange glyph. An API-backed UI
assertion for the aggregate glyph's stable semantic identity failed before
implementation and passes after `MPWalletAggregateGlyph` replaced only that
icon. Inspected populated screenshot:
`/tmp/mypocket-total-glyph-attachments/3D11766A-DDD7-459D-BC11-C16FA65867C9.png`.
The icon uses the shared categoryTeal foreground/soft-surface/border and globe
symbol. Selection, amount, filters, wallet-specific glyphs and public API remain
unchanged.

## Evidence and implementation notes — 2026-10-07

- More/account screenshot exported from
  `/tmp/mypocket-more-visual-capture/Logs/Test/Test-MyPocketIOS-2026.10.07_16-12-40-+0700.xcresult`
  and inspected at `/tmp/mypocket-more-visual-attachments/8AD2A57B-93FC-4A57-9136-B1A71ADB0D62.png`.
- The transaction editor was captured first with the deterministic unauthenticated
  launch and then with a disposable local API wallet/category fixture. The
  fixture-backed visual evidence is
  `/tmp/mypocket-transaction-api-ai-final-attachments/1C3FFCC1-E76B-4E97-A570-8E384B43E280.png`.
- The actionable mismatch was that the editor rendered below the shell bottom
  navigation. Transaction editors now use the shared full-screen modal
  presentation from both the transaction ledger and wallet detail, so the save
  footer is not occluded. `MPReferenceScreenHeader` now owns the Stitch close
  capsule, and the shared `MPActionIconButton` exposes the existing AI-entry
  route from new-transaction footers.
- Focused unauthenticated editor smoke: 1/1 at
  `/tmp/mypocket-transaction-editor-guarded/Logs/Test/Test-MyPocketIOS-2026.10.07_16-30-49-+0700.xcresult`.
- Focused API-backed wallet editor smoke: 1/1 at
  `/tmp/mypocket-transaction-api-guarded/Logs/Test/Test-MyPocketIOS-2026.10.07_16-31-24-+0700.xcresult`.
- The full native unit/contract target remains 110/110 at
  `/tmp/mypocket-visual-full-tests-2/Logs/Test/Test-MyPocketIOS-2026.10.07_16-30-20-+0700.xcresult`;
  Release warnings-as-errors, project/base-component guards, diff guard and
  web build/design checks are green.
- This is inspectable simulator evidence only; owner pixel approval on the
  physical device and live provider/OCR/Google/production feedback gates remain
  open.

## Follow-up decision — More header alignment

The inspected Stitch More reference centers `Mở rộng` independently of the
right-side `Hỗ trợ` pill, while the generic native header left-aligns it. Add a
shared `MPCenteredScreenHeader` base and use it only on More so the title is
centered without changing secondary-screen headers or route IDs. Verify with
the existing More route smoke and base-component guard.

Implemented and verified: `testMoreProfileCardOpensAccountSettings` passes 1/1
at `/tmp/mypocket-more-centered-header/Logs/Test/Test-MyPocketIOS-2026.10.07_16-35-18-+0700.xcresult`,
with the inspected attachment at
`/tmp/mypocket-more-centered-header-attachments/3B1118C1-AE96-4736-B481-71CAD1C76490.png`.
