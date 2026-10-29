# Native credit statement

## Task 4 ownership boundary

The credit statement implementation remains in
`Features/Secondary/SecondaryFeatureViews.swift`. This organization slice owns
that composition, its focused tests, and this contract. Wallet routing may
continue to open `CreditView`, but its summary, rows, forms, colors, or spacing
must not be duplicated in `WalletsView.swift`; paired payment accounting and
report exclusion remain unchanged.

Entry: credit wallet scope/detail or More → Sổ tín dụng. Existing wallet selector
is owner-scoped; selected wallet determines statement. Add entry/payment opens
MPFormBottomSheet; category uses shared tree. Payment selects eligible source
wallet; save/reload uses existing Go accounting, not UI-computed balance writes.

Bases: shared credit summary, MPOptionSelector, MPAmountField, MPTextField,
MPDateField, MPCategoryTree, MPTransactionRow and status/error states. Debt is
shown as positive owed amount, payment/refund uses correct semantic direction;
mask privacy applies to all finance displays. Cancel/dismiss must preserve or
discard draft according to shared editor contract, not save implicitly.
List/empty/pending/error, due-date and statement labels are native contracts.
Shared row labels prioritize explicit `credit_kind=payment` over a generic
incoming-transfer label, while keeping neutral transfer amount/icon semantics.
Purchases use category/note/amount like all transaction lists, not a competing
screen-specific row. Unit RED showed Nhận chuyển ví instead of Thanh toán; local
API UI readback now verifies purchase12000, payment7500, linked pair ID and
`included_in_reports=false`. Screenshot inspected; current final189/189 gate
passes185 units and4 selected UI, including this credit entry/payment flow.
[Wallet behavior](../wallets/README.md), [registry](../../platforms/ios-screen-registry.md).

## Planning reference composition — 2026-10-26

Statement totals use `MPFinanceSummaryCard`, action controls remain outside the
summary, and transactions render in one grouped surface using `MPGroupedRow`
and `MPInsetDivider`. `CreditReferencePresentation` supplies section spacing
and accessibility keys consumed by the rendered summary/list; transaction anatomy remains
`MPTransactionRow`.

Statement loading, debt arithmetic, entry/refund/fee/interest creation, paired
payment creation, report exclusion, wallet/category eligibility, and privacy
masking are unchanged. Focused unit tests cover the four-metric masked statement
contract and the exact category/source-wallet selections that enable the
rendered save action. They do not exercise API persistence, paired-payment
creation, or SwiftUI hierarchy. Integrated visual/device acceptance remains a
later redesign gate.

Review remediation evidence (2026-10-27): the four owned reference test classes
passed 24/24 and the normally signed `MyPocketIOSTests` target passed 422/422 on
the iPhone 16 iOS 18.5 simulator. The Release warnings-as-errors build,
base-component guard, and project verification also passed.
