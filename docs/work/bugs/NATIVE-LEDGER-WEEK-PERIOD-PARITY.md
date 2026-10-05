---
id: NATIVE-LEDGER-WEEK-PERIOD-PARITY
status: regression_verified
scope: mobile-ios
---

# Native ledger week-period parity

## Symptom

The native transaction ledger still exposes an all-time `Tất cả` period and
does not provide previous/next account-local week navigation. Its summary also
uses a two-column metric layout instead of the web contract's `Tiền vào` /
`Tiền ra` period cashflow card. This makes `Tổng cộng` and basic-wallet scopes
behave differently from the approved ledger contract.

## Intended behavior

- Aggregate and basic-wallet scopes start in the current account-local
  Monday–Sunday week.
- The only period modes are `Theo tuần` and `Tùy chọn`; a week can move back
  or forward, but never beyond the current week.
- Both period filtering and cashflow totals use the selected wallet scope and
  account timezone.
- Goal wallets keep their complete history/progress body; credit wallets keep
  their dedicated statement/payment body.

## Verification plan

1. Add pure tests for Monday–Sunday filtering, week offsets and reversed custom
   ranges.
2. Add a shared `MPWeekNavigator` base component and remove the all-time option
   from the native ledger UI.
3. Match the web cashflow card labels and rerun native unit/UI suites, release
   build and base-component guards.

## Resolution

The native ledger now starts aggregate/basic scopes in account-local
Monday–Sunday `Theo tuần`, exposes previous/next week navigation bounded at the
current week, and keeps `Tùy chọn` for custom ranges with a reversed-range
validation message. The all-time tab was removed. Period cashflow now uses the
shared surface with `Tiền vào` and `Tiền ra`; goal and credit scopes retain their
specialized bodies.

Regression proof: the new week/filter tests and API-backed UI assertions cover
the previous all-time/no-navigation failure mode. Green proof: native unit
tests pass 41/41 and the targeted API-backed ledger UI test passes on iPhone 16.
