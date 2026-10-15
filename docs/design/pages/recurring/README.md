# Native recurring transactions

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
save clears it; leaving/restarting does not promise durable persistence. Same-item
cancel/reopen passed in final188/188 regression gate, including interval2 persistence.
[Registry](../../platforms/ios-screen-registry.md).
