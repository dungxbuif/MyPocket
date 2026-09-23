---
artifact_type: detail_design
id: UI-FORMS-03
status: in_review
approval: owner_requested_implementation
owner: shared
updated: 2026-09-21
trace:
  feedback: FB-004
  backlog: UI-FORMS-03
  validation: ../VALIDATION_MATRIX.md
  adr: [ADR-002, ADR-003]
---

# Transaction and budget reference fidelity

Owner explicitly requested implementation against five supplied Stitch references and an old-code checkpoint (179854e), then added Money Lover AI visual guidance and requested the transaction category selector use the group-management tree base in all related forms. Source feedback: [FB-004](../FEEDBACK_LOG.md). No active phase.

## Design and reference mapping

| Reference region / action | Shared base / behavior |
| --- | --- |
| Cancel + centered title + tall canvas | Existing BaseBottomSheet form; add footer slot to keep save accessible; respect nested dialogs/focus |
| Main grouped white form card | SurfaceCard with named form variant shared with Add Wallet; Divider separates fields |
| Expense/income/debt tabs | SegmentedControl inside TransactionFields card; debt disabled with accessible unavailable explanation (no debt ledger) |
| Wallet icon/name/chevron | FormSelectorRow opens shared list of eligible basic/goal wallets; credit reference appearance does not authorize credit ledger |
| VND + large amount | New AmountField uses BaseTextInput amount variant; custom calculator/keypad composed from BaseButton |
| Category icon/name/chevron | FormSelectorRow opens CategoryTreeSelector; BaseCategoryTree selection mode mirrors the root/child group-management UI and only applicable IDs are selectable |
| Note row | Inline BaseTextInput + leading icon |
| Date with arrows | Shared DateField; change one local calendar day, preserve time; central button opens BaseCalendar modal |
| Calendar month arrows, weekday grid, selected day, cancel | New BaseCalendar + BaseModal; Monday-first, keyboard arrows/PageUp/PageDown/Home/End, Escape/return focus, month/year picker |
| More details | BaseButton and BaseSwitch for report inclusion; time editing in details |
| Save and attachment | Footer BaseButton; attachment invokes AI upload when available, with accurate accessible action label; manual retained attachments remain incomplete |
| Keypad digits/operators/clear/backspace/done | AmountField calculator; no eval; VND integer-only validation; suggestions excluded without real history source |
| Budget category/amount/period/wallet | BudgetEditorForm composes same FormSelectorRow/AmountField/DateField; explicit inclusive start/end range |
| Budget repeat switch | BaseSwitch disabled/off: recurring API not implemented; explanatory copy |
| Budget name | Retain editable inline name because real API requires it; explicit deviation from image |
| Day-group transaction card | SurfaceCard form + big day, weekday/month, signed total, Divider, existing TransactionItem |
| Transaction action menu | BaseButton/Modal for date filter and category grouping. Batch delete/transfer/adjustment/sync remain disabled or absent with explicit unavailable state; no mock actions |

## Scope and verification

Reuse: BaseBottomSheet, SurfaceCard, BaseButton, IconBadge, Divider, BaseTextInput, BaseSelect, BaseSwitch, SegmentedControl, BaseCategoryTree through CategoryTreeSelector, TransactionItem. Add shared BaseModal/BaseCalendar, DateField, AmountField, TransactionFields and BudgetEditorForm before consuming them. AI proposal cards reuse TransactionFields; AssistantComposer/AssistantResultCard follow the supplied Money Lover reference while preserving explicit review. Backend batch persistence and attachment linking remain AI-ENTRY-02 scope.

Preserve transaction/budget API semantics, money validation, category-wallet rules, ended-budget restrictions and error retention. Do not create mock transactions or enable unsupported credit/debt/recurrence. No database/API change required for these forms.

Tests: calendar month/leap/year boundaries, calculator valid/invalid expressions, selectable BaseCategoryTree behavior, browser category search/selection in transaction and budget pickers, date cancel/select/preservation, manual save payload, budget date range, grouped-day display, screenshots at 390x844 and Money Lover AI composition comparison. Commands and outputs are recorded in [validation matrix](../VALIDATION_MATRIX.md). The screen has no API/schema changes; AI batch persistence/integrations remain separate AI-ENTRY-02 work.

## Reconciliation

Updated [transaction page](../../design/pages/transactions/README.md), [budget page](../../design/pages/budgets/README.md), [assistant page](../../design/pages/assistant/README.md), [base contract](../../design/system/BASE_COMPONENTS.md), [validation](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md), [backlog](../BACKLOG.md), and [release](../../releases/CHANGELOG.md). Existing ADR-002/003 govern visual enforcement; API/ERD unchanged. Browser and screenshot proof are complete; owner visual UAT remains pending, so status is `in_review`.
