---
id: NATIVE-GOAL-TARGET-DATE-TIMEZONE
status: regression_verified
scope: mobile-ios
---

# Native goal target-date timezone parity

## Symptom

The goal-wallet editor displays a date picker in the account timezone but
serializes the selected date with a UTC-only formatter. For account zones east
or west of UTC, a date selected at local midnight can be persisted as the
previous or next calendar day.

## Intended behavior

Goal `target_date` is a calendar date, not an instant. Loading and saving the
editor must interpret and serialize it in the account IANA timezone without
depending on the simulator/device timezone.

## Verification plan

1. Add timezone-aware date-only parse/format helpers and a regression test.
2. Pass the loaded account timezone into the wallet editor.
3. Run native unit/UI suites, release build and base-component guards.

## Resolution

The wallet editor now receives the loaded account timezone, uses a timezone-aware
date-only parser for existing goal targets, and serializes new target dates in
that same IANA timezone. The date remains a calendar value; no UTC/device-day
conversion is used. The regression test passes on the iPhone 16 simulator.
