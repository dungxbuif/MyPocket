# AI budget update action

## Goal

Extend the Finance Assistant's prepared-action boundary with an explicit,
owner-scoped proposal for replacing an existing budget. The model may prepare
the change, but only the user-confirmed action may call the existing budget
PATCH validator. No model tool writes directly, and no new migration or
provider/runtime change is introduced.

## Contract

- Read-only tool: `propose_update_budget`.
- Required arguments: `budget_id`, `name`, `limit_amount`, `start_date`,
  `end_date`.
- Optional arguments: `wallet_id`, `category_id` (UUID or null).
- Validate positive safe-integer VND limit, calendar dates, end >= start,
  non-empty name <= 200 characters, and UUID fields before returning a
  proposal.
- Structured view: `budget_update_proposal`, normalized to the native
  `budget_confirmation` card with action `update_budget`.
- Confirmation payload contains only action ID, conversation ID and
  generation. The persisted proposal remains the source of financial fields.
- Dispatcher calls the existing owner-scoped `PATCH /api/v1/budgets/:id`
  handler with a stable idempotency key derived from the prepared action;
  budget updates retain the existing replace semantics and do not create a
  second ledger row.

## Verification

- Red/green Go tests for tool schema/validation, structured-card mapping and
  confirmation dispatch with persisted proposal binding.
- Native decoder/card tests for valid, malformed and reversed-date proposals.
- Full Go test/vet/race, native DesignSystem and base-component guards.
- Reconcile API, assistant, iOS and validation docs after implementation.

## Non-goals

Budget deletion, transaction mutation tools, production rollout, migration,
live provider calls, device installation and UI redesign remain separate work.

## Reconciliation

Implemented locally: the tool, structured view, owner/generation-bound
dispatcher, native shared confirmation card and focused tests/docs are now in
the worktree. Full local gates are being rerun; production/provider/device
acceptance remains intentionally open.
