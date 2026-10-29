# Travel Mode implementation plan

## Goal

Implement the approved Travel Mode slice as an event dimension, not a wallet
type. A user can create events, have at most one active event, link/unlink
ordinary transactions, and inspect event-linked rows without changing wallet
balances or report inclusion.

## Contract

- `travel_events` is owner-scoped and stores a name, optional context,
  optional date-only start/end, and active state.
- `POST /api/v1/travel/events` creates an event; `GET` lists owner events;
  `PATCH` edits metadata; `DELETE` removes the event and clears links without
  deleting transactions.
- `POST /api/v1/travel/events/:id/activate` atomically deactivates any other
  active event for the owner and activates the selected event.
- Ordinary income/expense creation uses the explicit event when supplied, or
  the active event when omitted. Transfer, adjustment, recurring and AI
  approval flows do not auto-attach an event.
- `PATCH /api/v1/transactions/:id/travel` links or clears one owner-scoped
  ordinary transaction. Credit rows and transfer pairs are rejected.

## Verification

- Handler tests cover owner scope, one-active-event invariant, malformed dates,
  link/unlink validation, and automatic attachment only on ordinary creates.
- PostgreSQL tests cover event deletion clearing links and atomic activation.
- Frontend design/type/build checks cover the Account Travel Mode panel and
  shared base controls.
- Public API, migration ledger, page contract, and release notes are updated
  with remaining browser/UAT gates called out explicitly.
