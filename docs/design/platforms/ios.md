# iOS renderer contract

This document defines how the existing MyPocket screen contracts render in
SwiftUI. It does not redefine API or finance behavior.

## Renderer rules

- Feature screens live under `mobile-ios/MyPocketIOS/Features`.
- User-facing controls and visual behavior live under
  `mobile-ios/MyPocketIOS/DesignSystem`.
- Feature views compose named bases; raw controls are allowed only inside the
  base that owns their accessibility, focus, validation and styling.
- Safe-area, keyboard, Dynamic Type, VoiceOver and 44pt hit-target behavior are
  owned by the base components and verified in previews/UI tests.
- Web and iOS use the same semantic token roles and business rules, but do not
  share DOM, React Native or CSS implementation files.

## Screen mapping

| Semantic screen | SwiftUI feature | Required base composition |
| --- | --- | --- |
| Overview | `Features/Overview` | `MPScreenHeader`, `MPMetricCard`, `MPSurfaceCard`, async states |
| Transactions | `Features/Transactions` | `MPWalletSelector`, `MPSegmentedControl`, `MPTransactionRow`, `MPBottomSheet` |
| Wallets | `Features/Wallets` | `MPWalletRow`, `MPAmountField`, `MPDateField`, `MPConfirmDialog` |
| Budgets/reports | `Features/Budgets`, `Features/Reports` | `MPMetricCard`, `MPProgressBar`, `MPChart`, `MPDateRangePicker` |
| Jars/months | `Features/Jars`, `Features/Months` | `MPProgressBar`, `MPTextField`, `MPBottomSheet` |
| AI entry | `Features/AIEntry` | `MPAttachmentPicker`, `MPAttachmentThumb`, `MPActionCard`, async states |
| Assistant | `Features/Assistant` | `MPAssistantCard`, `MPActionCard`, `MPTextField` |
| Feedback | `Features/Feedback` | `MPFeedbackBubble`, `MPBottomSheet`, `MPAttachmentThumb`, `MPStatusView` |
| Account | `Features/Account` | `MPScreenHeader`, `MPButton`, `MPConfirmDialog`, `MPStatusView` |

## States and accessibility

Every feature exposes loading, ready, empty and error states. Write forms also
expose dirty, submitting, success and conflict states. Async state bases must
provide a visible label and accessibility value; masked money must not leak the
unmasked amount through VoiceOver. UI tests use stable accessibility IDs rather
than screen text alone.
