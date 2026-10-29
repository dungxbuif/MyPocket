---
artifact_type: bug
status: in_progress
owner: AI
---

# Feedback 2d0811c0: native portfolio money input

Source: owner feedback `2d0811c0-1894-404d-8385-e3d665fb058e`:
"Các ô nhập tiền đang ko đô gf nhất cách format tiền. Ví dụ thêm tài sản. Card phản hồi padding ngang nawgns quá viền ô chi tiết mất hết".

## Scope and current evidence

- This pass covers native portfolio trade, correction and manual-price VND inputs only. The add-asset editor itself contains code/name, not a money input. Feedback-card horizontal padding/clipped detail border is a separate screen/design issue and remains unresolved here.
- `PortfolioView` in `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift` already composes `MPAmountField` for unit price, fee and manual price; decimal quantity correctly uses `MPNumberField`.
- `MPAmountField` standard presentation currently displays raw integer digits. Its transaction presentation displays Vietnamese dot grouping while retaining raw digits in the binding. This is a visible inconsistency for portfolio amounts. `MPFormat.parseNonNegativeInteger` intentionally rejects separators in the bound draft.
- Existing `testMoneyGroupingNeverChangesTheBoundIntegerDraft` covers transaction presentation only. Portfolio editor/API tests exercise money submission but not grouped display.

## Decision and proof

- Reuse the shared amount control and add opt-in grouping to its standard visual form; enable it only for portfolio VND fields. Preserve labels, currency suffix, calculator, layout and raw ASCII integer binding. Invalid pasted drafts must stay invalid, rather than being silently converted to money.
- Add a failing regression for portfolio-style grouped display plus unchanged bound value; verify editing/deletion and malformed paste. Run focused native tests with a dedicated simulator and derived-data path, then the relevant surrounding test target. Do not run API-backed UI fixtures against shared services.
- No API, schema, finance calculation, permissions, production deployment or feedback status changes. Shared `MPFields.swift` has other consumers; opt-in presentation limits the effect. No ADR is needed for a display-only option.
- Reconcile this note and the portfolio iOS guide after implementation. The guide's existing AI-agent reference is the fetchable Markdown surface. Shared context/backlog/validation files are under other workers' ownership and are not edited in this pass.

## Result and verification

- `MPAmountField(groupsDigits: true)` now uses an opt-in grouped standard control for portfolio unit price, fee, correction and manual price. Only canonical nonnegative integer drafts display dot grouping; the binding stays raw digits. Quantity remains an ungrouped decimal field. Existing standard amount consumers and the transaction presentation retain their prior setting.
- Red proof: focused `xcodebuild test` on cloned simulator `F67C6E23-A2DE-43ED-B3D7-D7C8FB6A5CD7` failed to compile before the option existed. A second red run caught leading-zero collapse (`01` displayed as `1`); the formatter now leaves that draft intact.
- Green proof: `xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=F67C6E23-A2DE-43ED-B3D7-D7C8FB6A5CD7' -derivedDataPath /tmp/mypocket-money-2d0811c0-derived -resultBundlePath /tmp/mypocket-money-2d0811c0-final-suite.xcresult -parallel-testing-enabled NO -only-testing:MyPocketIOSTests/DesignSystemTests -only-testing:MyPocketIOSTests/FinanceModelTests -quiet` passed 193 tests with no failures or skips. Regression covers grouped display, raw binding, digit editing/deletion, invalid grouped paste and leading zeros.
- No API-backed UI test, physical-device visual sign-off or Feedback-sheet clipping proof is claimed. The Feedback-card portion of the original report remains a separate residual.
- Docs review: portfolio screen contract and user/agent guide now describe display-only grouping. API, schema, calculations, architecture and release notes need no change because their contracts did not change; shared context/backlog/validation and production feedback status were intentionally left to their owners.

## References

- [Portfolio native design](../../design/pages/portfolio/README.md)
- [Shared base components](../../design/system/BASE_COMPONENTS.md)
- [Portfolio user and agent guide](../../guides/portfolio-ios.md)
- [Native UI audit](../tickets/NATIVE-UI-SYSTEM-AUDIT-20261014.md)
