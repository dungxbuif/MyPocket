# Native recurring transactions

## Task 4 ownership boundary

The native implementation remains in
`Features/Secondary/SecondaryFeatureViews.swift`. This organization slice owns
that composition, its focused tests, and this contract; grouped rows and editor
scaffolds must not be duplicated elsewhere.

Count input now uses `MPNumberField(.integer)`, not currency/keypad. Interval
validation/save accepts strict positive ASCII integer only, rejecting `1.000`
rather than silently treating it as1000. Amount remains `MPAmountField` with its
money parser. Existing API-backed create/update/delete verifies interval2 after
editing. [Number input guide](../../../guides/numeric-input-ios.md).

Entry: Tài khoản → Giao dịch định kỳ. Add/edit replaces the list with the shared
`MPEditorScaffold`, one scroll and pinned Save. Existing delete confirmation preserves
already-created transactions. Save/cancel return to list; category is scoped to
type/wallet and uses shared searchable category sheet.

Bases: MPAmountField, MPTextField, MPOptionSelector/MPWalletSelector,
MPDateField, MPCategoryTree, MPFormBottomSheet, MPConfirmDialog. Amount is money;
recurrence interval is integer count and must not use currency semantics in the
new field family. Render next run/type/wallet/note and active state with shared
row/meta typography. Pending disables edits/cancel/re-submit; failure keeps the
draft. Hủy does not write; reopening the same schedule in the mounted screen
resumes fields/count, selecting another schedule starts its own input. Successful
save clears it; leaving/restarting does not promise durable persistence. The
owned unit contract covers same-item mounted cancel/reopen and recurrence-label
behavior, not API-backed interval persistence.
[Registry](../../platforms/ios-screen-registry.md).

## Organization reference composition — 2026-10-26

The route summary/actions use `MPFinanceCard` and populated schedules use one
grouped surface. Each schedule is inset through `MPGroupedRow`, with
`MPInsetDivider` between rows; density comes from
`RecurringReferencePresentation`. The existing editor remains
`MPEditorScaffold` with its category sheet and pinned save action.

Recurrence validation, due execution, owner-scoped wallet/category selection,
draft retention, deletion, and generated-transaction behavior are unchanged.
Focused unit tests cover cadence labels, strict-integer validation, and the
same-item cancel/reopen session contract consumed by this editor. They do
not prove API persistence, due execution, or rendered hierarchy. Recurring row
geometry uses shared metrics; integrated visual/device acceptance remains a
later redesign gate.

Review remediation evidence (2026-10-27): the four owned reference test classes
passed 24/24 and the normally signed `MyPocketIOSTests` target passed 422/422 on
the iPhone 16 iOS 18.5 simulator. The Release warnings-as-errors build,
base-component guard, and project verification also passed.
