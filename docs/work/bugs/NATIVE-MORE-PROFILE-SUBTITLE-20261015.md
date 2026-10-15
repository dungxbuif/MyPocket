# More profile card subtitle

## Intended behavior

The shared `MPProfileCard` must render its supplied `accountSubtitle` between
the identity line and the email line. The More screen contract exposes
`Hồ sơ tài khoản` as a stable accessible text label and must keep it visible
without changing navigation or session data.

## Technical decision

Keep the subtitle in the shared base component rather than duplicating it in
`MoreView`. The component continues to own typography, spacing and accessibility;
feature screens only provide the localized value.

## Verification

- Focused signed simulator UI test:
  `LaunchTests.testMoreProfileCardOpensAccountSettings`
- Full signed native suite after the focused fix.
- No API, schema, database, environment or production change.

## Reconciliation

`DesignSystemTests.testProfileCardUsesCompactIdentityLayout` passes after the
subtitle/email row was compacted into the shared base. The profile UI test
passes 1/1, and the final signed simulator suite passes 294 native unit tests
plus 100 UI tests with 64 documented skips and no failures:
`/tmp/mypocket-full-native-budget-update-profile-final-20261015/Logs/Test/Test-MyPocketIOS-2026.10.15_19-52-13-+0700.xcresult`.
