---
artifact_type: bug
id: NATIVE-TRAVEL-ADJUSTMENT-SCOPE
status: regression_verified
created: 2026-10-06
updated: 2026-10-06
---

# Native Travel Mode and balance-adjustment scope

## Finding

The transaction detail UI represented Travel Mode choices by event name, so
two events with the same display name could update the wrong event. The native
balance-adjustment editor also serialized `occurred_at` with the device
timezone instead of the account timezone used by the rest of the ledger.

## Intended behavior

- Travel Mode selectors use stable event IDs while displaying event names.
- Saving or adjusting a transaction uses the account IANA timezone, preserving
  the user's local calendar day across devices.

## Implementation and verification result

- Transaction detail now uses `MPOptionSelector` with stable Travel Event IDs;
  duplicate event names no longer select an arbitrary record.
- Adjustment writes now normalize to account-local noon through the same
  timezone contract used by transfer and portfolio date payloads.
- Native unit tests pass 48/48, including a cross-timezone adjustment-date
  assertion.
- The signed iPhone 16 Simulator UI suite passes 14/14 after the change.

## Verification plan

1. Add ID-based selector state and an account-timezone timestamp helper.
2. Add unit coverage for the adjustment timestamp and rerun native unit/UI
   suites on the iPhone 16 Simulator.
3. Reconcile this note and the validation matrix after the full suite passes.

No production data or provider credentials are used for this regression.
