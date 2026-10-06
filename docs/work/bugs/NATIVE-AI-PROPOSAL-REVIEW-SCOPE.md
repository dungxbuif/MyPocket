# Native AI proposal review scope

Status: implemented and simulator-verified (2026-10-06)

## Problem

The native AI transaction-entry review currently renders the model's proposal
as either income or expense without allowing the user to correct an unknown
type. Jar assignment and report inclusion are carried from the model but are
not reviewable in the editor. This violates the product contract that AI is
only a suggestion and every proposal must be verified before approval.

## Intended behavior

- A proposal with `type=unknown` (or any unsupported value) must show
  `Chưa xác định — chọn loại` and cannot be saved until the user selects
  `expense` or `income`.
- Category options are filtered against the selected transaction type and
  wallet. Changing type or wallet clears an invalid category.
- Expense proposals may select a jar from the account-local month of the
  proposal date; income proposals must not send a jar assignment.
- The user can explicitly choose whether the transaction is included in
  reports. The chosen value is included in the existing proposal PATCH body.
- No API or wire-value changes are introduced. The backend remains the source
  of truth and approval stays an explicit user action.

## Technical decision

Keep the behavior in the native base-component layer: use `MPOptionSelector`
for type/category/jar and a shared base control for the report inclusion flag.
Add pure review helpers so the contract is testable without a live backend.
Load jar identities independently from wallets/categories so a temporary jars
failure does not hide the rest of the review data.

## Verification plan

1. Add failing unit tests for unknown-type display/validation, expense-only
   jar eligibility, and report flag preservation.
2. Implement the smallest editor/data-loading change and run the native unit
   suite.
3. Build Release for the iPhone simulator with `LOCAL_API_UI`, run the
   deterministic unknown-type review fixture plus the AI-entry capability path
   on the simulator, then run base-component/project guards and `git diff --check`.
4. Reconcile this note, validation matrix, and release changelog with the
   shipped behavior.

## Evidence

- `MyPocketIOSTests`: 58/58 passed.
- Signed iPhone 16 Simulator review fixture:
  `testAIEntryProposalReviewRequiresExplicitType` passed; the editor showed the
  explicit type error, then exposed the active hũ selector and report toggle
  after choosing `Chi tiêu`.
- Full signed `LOCAL_API_UI` simulator suite: 15/15 passed.
- Release build with warnings-as-errors and base-component/project guards passed.
