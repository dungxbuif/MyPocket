# Native credit statement

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
