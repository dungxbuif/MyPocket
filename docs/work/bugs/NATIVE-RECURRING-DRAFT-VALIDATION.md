---
artifact_type: bug
id: NATIVE-RECURRING-DRAFT-VALIDATION
status: regression_verified
owner: shared
---

# Native recurring draft validation

## Finding

`RecurringView.save()` parsed the amount and interval from the number-field
text but only checked `interval > 0`. A zero amount could therefore pass the
native guard and rely on a backend rejection, while type/frequency constraints
were also not represented in the draft validation contract.

## Intended behavior

- name is non-blank after trimming;
- amount and interval are positive integer values after the same digit-only
  normalization used by the amount fields;
- type is `expense` or `income` and frequency is one of `daily`, `weekly`,
  `monthly`, `yearly`;
- an enabled end date is not earlier than the next-run date;
- invalid drafts never issue a recurring create/update request.

The server remains authoritative for owner wallet/category scope and all
recurring persistence invariants. This is a client-side parity guard, not a
replacement for server validation.

## Verification plan

- Add pure `RecurringDraftValidator` tests for zero/invalid amount, interval,
  type/frequency, blank name and reversed end date.
- Run the full native unit suite, release build with warnings-as-errors, base
  component/project guards, and the signed `LOCAL_API_UI` simulator suite.

## Verification result — 2026-10-06

`RecurringDraftValidator` now rejects zero/overflow/empty amount or interval,
unsupported type/frequency, blank names and reversed enabled end dates before
the API call. XCTest passes 44/44, including the new recurring validator cases.
Release and signed simulator UI verification are recorded in the active
validation matrix after the implementation pass.
