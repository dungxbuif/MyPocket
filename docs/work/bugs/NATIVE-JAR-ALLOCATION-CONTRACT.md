---
artifact_type: bug
id: NATIVE-JAR-ALLOCATION-CONTRACT
status: regression_verified
owner: shared
---

# Native jar allocation contract

## Finding

The native jar editor validated the allocation mode and positive values, but it
did not mirror two server constraints: fixed VND allocations must stay within
the safe integer ceiling `9_007_199_254_740_991`, and percentage allocations
may contain at most two decimal places. Invalid drafts reached the API and
returned a generic error instead of being rejected locally.

## Intended behavior

- `none` carries no amount or percentage;
- `fixed` carries `1...9_007_199_254_740_991` VND;
- `percent` carries `0 < percent <= 100` with at most two decimal places;
- changing mode clears stale fields and the UI never submits a mismatched
  amount/percentage pair.

The backend remains authoritative; this validator only keeps the native form
consistent with `validateJarInput` before a request is made.

## Verification plan

- Add pure validator tests for safe-integer overflow and three-decimal percent.
- Run native unit tests, release build/guards and signed API-backed simulator UI
  suite; update the validation matrix and changelog with the evidence.

## Verification result — 2026-10-06

- Pure native XCTest: 44/44 passed, including safe-integer overflow,
  non-finite percentage and three-decimal percentage cases.
- Release build with Swift warnings treated as errors, base-component guard,
  project verification and `git diff --check`: passed.
- Signed iPhone 16 simulator API-backed UI suite: 10/10 passed in 199.811s.

The native draft now rejects the same unsafe fixed/percentage values before the
request is sent. The backend remains authoritative for final validation.
